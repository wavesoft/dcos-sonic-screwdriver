package registry

import (
  . "github.com/logrusorgru/aurora"
  "gopkg.in/src-d/go-git.v4"
  "crypto/sha256"
  "encoding/hex"
  "errors"
  "fmt"
  "gopkg.in/cheggaaa/pb.v1"
  "io"
  "io/ioutil"
  "os"
  "strconv"
)

/**
 * Compose a name for the given too/version/artifact combination
 */
func PkgDir(tool string, version *ToolVersion, artifact *ToolArtifact) (string, error) {
  registryPath, err := GetRegistryPath()
  if err != nil {
    return "", err
  }

  return fmt.Sprintf(
    "%s/%s-%d.%d.%d",
    registryPath,
    tool,
    uint32(version.Version[0]),
    uint32(version.Version[1]),
    uint32(version.Version[2])), nil
}

/**
 * Download package and validate
 */
func FetchArchive(tool string, version *ToolVersion, artifact *ToolArtifact) (string, error) {
  // Install artifact
  if (artifact.DockerToolArtifact != nil) {
    return CreateDockerWrapper(tool, version, artifact)
  } else if (artifact.ExecutableToolArtifact != nil) {
    if artifact.Source.GitSource != nil {
      return FetchGitArchive(tool, version, artifact)
    } else if artifact.Source.HttpSource != nil {
      return FetchHttpArchive(tool, version, artifact)
    } else {
      return "", errors.New(fmt.Sprintf(
        "No package sources are available for %s/%s", tool, version.VersionString()))
    }
  }

  return "", errors.New(fmt.Sprintf(
    "No known installable artifacts found for %s/%s", tool, version.VersionString()))
}

/**
 * Create a shell script that wraps docker
 */
func CreateDockerWrapper(tool string, version *ToolVersion, artifact *ToolArtifact) (string, error) {
  // Prepare package dir
  dir, err := PkgDir(tool, version, artifact)
  if err != nil {
    return "", err
  }

  if _, err := os.Stat(dir); err == nil {
    os.RemoveAll(dir)
  }
  err = os.MkdirAll(dir, 0755)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not create package dir: %s", err.Error()))
  }

  // First try to pull the image
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Pulling"), Bold(Gray(artifact.Image+":"+artifact.Tag)))
  err = PullDockerImage(artifact.Image, artifact.Tag)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not pull the docker image %s", err.Error()))
  }

  // Create wrapper script
  dat := []byte(fmt.Sprintf("#!/bin/bash\ndocker run -it --rm %s:%s $*", artifact.Image, artifact.Tag))
  err = ioutil.WriteFile(dir + "/docker-run.sh", dat, 0755)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not create wrapper: %s", err.Error()))
  }

  // Create a flag that indicates that the downloaded archive is ready for use
  err = ioutil.WriteFile(dir + "/.ready", []byte("OK"), 0644)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not set ready flag: %s", err.Error()))
  }

  // Return the entrypoint
  return dir + "/docker-run.sh", nil
}

/**
 * Download git archive and validate
 */
func FetchGitArchive(tool string, version *ToolVersion, artifact *ToolArtifact) (string, error) {
  // Prepare package dir
  dir, err := PkgDir(tool, version, artifact)
  if err != nil {
    return "", err
  }
  if _, err := os.Stat(dir); err == nil {
    os.RemoveAll(dir)
  }
  err = os.MkdirAll(dir, 0755)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not create package dir: %s", err.Error()))
  }

  // Git clone
  _, err = git.PlainClone(dir, false, &git.CloneOptions{
      URL:      "https://github.com/src-d/go-git",
      Progress: os.Stdout,
  })
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not clone git repository: %s", err.Error()))
  }

  // Get entrypoint
  entrypoint := artifact.Entrypoint
  if entrypoint == "" {
    entrypoint = GetDefaultEntrypoint()
  }
  entrypointFile := dir + "/" + entrypoint

  // Make sure entrypoint exists
  if _, err := os.Stat(entrypointFile); os.IsNotExist(err) {
    return "", errors.New(fmt.Sprintf(
      "Missing tool entrypoint: %s", entrypoint))
  }

  // Create a flag that indicates that the downloaded archive is ready for use
  err = ioutil.WriteFile(dir + "/.ready", []byte("OK"), 0644)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not set ready flag: %s", err.Error()))
  }

  return entrypointFile, nil
}

/**
 * Download web archive and validate
 */
func FetchHttpArchive(tool string, version *ToolVersion, artifact *ToolArtifact) (string, error) {
  client := RegistryHttpClient(true)
  srcUrl := artifact.Source.URL
  fmt.Printf("%s %s\n", Blue("==> "), Bold(Gray(srcUrl)))
  resp, err := client.Get(srcUrl)
  if err != nil {
    return "", errors.New(
      fmt.Sprintf("Could not request %s: %s", srcUrl, err.Error()))
  }
  defer resp.Body.Close()

  // Parse Content-Length header
  contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not parse Content-Length header: %s", err.Error()))
  }

  // Prepare package dir
  dir, err := PkgDir(tool, version, artifact)
  if err != nil {
    return "", err
  }
  if _, err := os.Stat(dir); err == nil {
    os.RemoveAll(dir)
  }
  err = os.MkdirAll(dir, 0755)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not create package dir: %s", err.Error()))
  }

  // Split streams, so we can calculate the checksum AND extract
  // while at the same time downloading the file.
  hasher := sha256.New()
  data := io.TeeReader(resp.Body, hasher)

  // Create progress bar
  bar := pb.New(contentLength).SetUnits(pb.U_BYTES)
  bar.Start()
  defer bar.Finish()

  // Extract
  err = ExtractTarGz(bar.NewProxyReader(data), dir + "/")
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could extract archive: %s", err.Error()))
  }

  // Now validate checksum
  csum := hex.EncodeToString(hasher.Sum(nil))
  if csum != artifact.Source.Checksum {
    os.RemoveAll(dir)
    return "", errors.New("Checksum validation failed")
  }

  // Get entrypoint
  entrypoint := artifact.Entrypoint
  if entrypoint == "" {
    entrypoint = GetDefaultEntrypoint()
  }
  entrypointFile := dir + "/" + entrypoint

  // Make sure entrypoint exists
  if _, err := os.Stat(entrypointFile); os.IsNotExist(err) {
    return "", errors.New(fmt.Sprintf(
      "Missing tool entrypoint: %s", entrypoint))
  }

  // Create a flag that indicates that the downloaded archive is ready for use
  err = ioutil.WriteFile(dir + "/.ready", []byte("OK"), 0644)
  if err != nil {
    return "", errors.New(fmt.Sprintf(
      "Could not set ready flag: %s", err.Error()))
  }

  return entrypointFile, nil
}
