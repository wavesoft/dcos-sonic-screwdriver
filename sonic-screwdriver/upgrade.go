package main

import (
  "bufio"
  "encoding/json"
  "errors"
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "gopkg.in/cheggaaa/pb.v1"
  "io/ioutil"
  "net/http"
  "time"
  "os"
  "os/exec"
  "strconv"
  "strings"
)

type LatestVersion struct {
  Version     registry.VersionInfo
  URL         string
}

/**
 * Get the latest released version
 */
func GetLatestVersion() (LatestVersion, error) {
  res := LatestVersion{}
  resp, err := http.Get("http://api.github.com/repos/wavesoft/dcos-sonic-screwdriver/releases/latest")
  if err != nil {
    return res, errors.New("error getting latest version info: " + err.Error())
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return res, errors.New("error reading latest version info: " + err.Error())
  }

  // Parse contents
  var dat map[string]interface{}
  if err := json.Unmarshal(body, &dat); err != nil {
    return res, errors.New("error parsing version info: " + err.Error())
  }

  // Extract version
  var ok bool
  var tagName string
  if tagName, ok = dat["tag_name"].(string); !ok {
    return res, errors.New("invalid version info: missing `tag_name`")
  }
  if !strings.HasPrefix(tagName, "v") {
    return res, errors.New("invalid tag name")
  }

  // Scan assets
  var downloadUrl string = ""
  var assets []interface{}
  if assets, ok = dat["assets"].([]interface{}); !ok {
    return res, errors.New("invalid version info: missing `assets`")
  }
  for _, asset := range assets {
    var mapInst map[string]interface{}
    if mapInst, ok = asset.(map[string]interface{}); !ok {
      return res, errors.New("invalid version info: invalid field `assets`")
    }
    if url, ok := mapInst["browser_download_url"].(string); ok {
      if strings.HasSuffix(url, ".darwin") {
        downloadUrl = url
        break
      }
    }
  }
  if downloadUrl == "" {
    return res, errors.New("could not find a download URL")
  }

  ver, err := registry.VersionFromString(tagName[1:])
  if err != nil {
    return res, err
  }

  res.Version = *ver
  res.URL = downloadUrl

  return res, nil
}

/**
 * Perform upgrade
 */
func PerformUpgrade(newVersion LatestVersion) error {
  resp, err := http.Get(newVersion.URL)
  if err != nil {
    return errors.New(
      fmt.Sprintf("could not request %s: %s", newVersion.URL, err.Error()))
  }
  defer resp.Body.Close()

  // Parse Content-Length header
  contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
  if err != nil {
    return errors.New(fmt.Sprintf(
      "could not parse Content-Length header: %s", err.Error()))
  }

  // Prepare package dir
  destDir, err := registry.GetRegistryPath()
  if err != nil {
    return errors.New(fmt.Sprintf(
      "could not detect registry path: %s", err.Error()))
  }
  if _, err := os.Stat(destDir); err != nil {
    err = os.MkdirAll(destDir, 0755)
    if err != nil {
      return errors.New(fmt.Sprintf(
        "could not create package dir: %s", err.Error()))
    }
  }

  // Create progress bar
  bar := pb.New(contentLength).SetUnits(pb.U_BYTES)
  bar.Start()
  defer bar.Finish()
  barStream := bar.NewProxyReader(resp.Body)

  // Find the path of our current executable
  replaceTarget, err := os.Executable()
  if err != nil {
    return errors.New(fmt.Sprintf(
      "could not find the location of the tool: %s", err.Error()))
  }

  // Rename the current executable
  bakTarget := replaceTarget + ".bak"
  err = os.Rename(replaceTarget, bakTarget)
  if err != nil {
    return errors.New(fmt.Sprintf(
      "could not rename old version: %s", err.Error()))
  }

  // Download contents to /tool
  f, err := os.Create(replaceTarget)
  if err != nil {
    os.Rename(bakTarget, replaceTarget)
    return errors.New(fmt.Sprintf(
      "could create destination file: %s", err.Error()))
  }

  // Try GZip and fall-back to plaintext
  writer := bufio.NewWriter(f)
  err = registry.DownloadFileGz(barStream, writer)
  if err != nil {
    f.Close()
    os.Remove(replaceTarget)
    os.Rename(bakTarget, replaceTarget)
    return errors.New(fmt.Sprintf(
      "could not process file stream: %s", err.Error()))
  }
  writer.Flush()

  // Make it executable and close it
  f.Chmod(0755)
  f.Close()

  // Run the new version that is going to remove the backup file
  cmd := exec.Command(replaceTarget, "complete-upgrade", bakTarget)
  err = cmd.Start()
  if err != nil {
    os.Remove(replaceTarget)
    os.Rename(bakTarget, replaceTarget)
    return errors.New(fmt.Sprintf(
      "could run the new version: %s", err.Error()))
  }

  // We are 99% sure that the upgrade was successful
  complete("Upgraded to version " + newVersion.Version.ToString())

  // We should now exit
  os.Exit(0)
  return nil
}

/**
 * Helper function to complete an upgrade process
 */
func CompleteUpgrade(bakTarget string) {
  // Wait for the other process to exit
  time.Sleep(500 * time.Millisecond)

  // Remove target
  err := os.Remove(bakTarget)
  if err != nil {
    die(fmt.Sprintf("could not remove old version: %s", err.Error()))
  }

  // Just exit
  os.Exit(0)
}
