package registry

// import (
//   "bufio"
//   "crypto/sha256"
//   "encoding/hex"
//   "encoding/json"
//   "errors"
//   "fmt"
//   "gopkg.in/cheggaaa/pb.v1"
//   "gopkg.in/src-d/go-git.v4"
//   "io"
//   "io/ioutil"
//   "os"
//   "strconv"
//   . "github.com/logrusorgru/aurora"
// )

// type InstalledVersions []InstalledVersion

// type InstalledVersion struct {
//   Version     VersionTriplet
//   Artifact    ToolArtifact
//   Entrypoint  string
// }


// /**
//  * Return the base directory to the tool
//  */
// func GetToolDir(tool string) (string, error) {
//   registryPath, err := GetRegistryPath()
//   if err != nil {
//     return "", err
//   }

//   return fmt.Sprintf(
//     "%s/tools/%s",
//     registryPath,
//     tool), nil
// }
// /**
//  * Compose a name for the given too/version/artifact combination
//  */
// func GetArchiveDir(tool string, version *VersionTriplet, artifact *ToolArtifact) (string, error) {
//   toolDir, err := GetToolDir(tool)
//   if err != nil {
//     return "", err
//   }

//   return fmt.Sprintf(
//     "%s/%d.%d.%d",
//     toolDir,
//     uint32(version[0]),
//     uint32(version[1]),
//     uint32(version[2])), nil
// }

// /**
//  * Return a list of the installed versions of the given tool
//  */
// func GetInstalledVersions(tool string) (InstalledVersions, error) {
//   versions := InstalledVersions{}
//   toolDir, err := GetToolDir(tool)
//   if err != nil {
//     return versions, err
//   }

//   // Check if the tool folder is missing
//   if _, err := os.Stat(toolDir); err != nil {
//     return versions, nil
//   }

//   // List versions
//   files, err := ioutil.ReadDir(toolDir)
//   if err != nil {
//     return versions, err
//   }
//   for _, f := range files {
//     verInfo, err := VersionFromString(f.Name())
//     if err != nil {
//       continue
//     }

//     // Load artifact from the state
//     artifactDir := toolDir + "/" + f.Name()
//     artifact, err := ReadStateFlag(artifactDir + "/.state")
//     if err != nil {
//       continue
//     }

//     // Get the entrypoint
//     entrypoint, err := GetArtifactEntrypoint(artifactDir, artifact)
//     if err != nil {
//       continue
//     }

//     versions = append(versions, InstalledVersion{ *verInfo, *artifact, entrypoint })
//   }

//   return versions, nil
// }

// /**
//  * Resolve to the artifact entrypoint, according to it's type
//  */
// func GetArtifactEntrypoint(archiveDir string, artifact *ToolArtifact) (string, error) {
//   entrypoint := ""
//   if (artifact.DockerToolArtifact != nil) {
//     entrypoint = "/run.sh"
//   } else {
//     if artifact.Extract {
//       entrypoint := artifact.Entrypoint
//       if entrypoint == "" {
//         entrypoint = GetDefaultEntrypoint()
//       }
//     } else {
//       if artifact.Interpreter != "" {
//         entrypoint = "/run.sh"
//       } else {
//         entrypoint = "/tool"
//       }
//     }
//   }

//   // Make sure entrypoint exists
//   entrypointFile := archiveDir + "/" + entrypoint
//   if _, err := os.Stat(entrypointFile); os.IsNotExist(err) {
//     return "", errors.New(fmt.Sprintf(
//       "missing tool entrypoint: %s", entrypoint))
//   }

//   return entrypointFile, nil
// }

// /**
//  * Checks if a tool with this name exists
//  */
// func IsToolInstalled(tool string) bool {
//   toolDir, err := GetToolDir(tool)
//   if err != nil {
//     return false
//   }

//   if _, err := os.Stat(toolDir); err == nil {
//     return true
//   }
//   return false
// }

// /**
//  * Remove the tool from the registry
//  */
// func RemoveTool(tool string) error {
//   toolDir, err := GetToolDir(tool)
//   if err != nil {
//     return err
//   }

//   return os.RemoveAll(toolDir)
// }

// /**
//  * Read the artifact metadata from the ready flag
//  */
// func ReadStateFlag(path string) (*ToolArtifact, error) {
//   byt, err := ioutil.ReadFile(path)
//   if (err != nil) {
//     return nil, err
//   }

//   // Parse the JSON document
//   artifact := new(ToolArtifact)
//   if err := json.Unmarshal(byt, artifact); err != nil {
//     return nil, err
//   }

//   return artifact, nil
// }

// /**
//  * Create a flag with the tool artifact details
//  */
// func CreateStateFlag(path string, artifact *ToolArtifact) error {
//   bytes, err := json.Marshal(artifact)
//   if err != nil {
//     return err
//   }

//   return ioutil.WriteFile(path, bytes, 0644)
// }

// /**
//  * Remove archive
//  */
// func RemoveArtifact(tool string, version *VersionTriplet, artifact *ToolArtifact) error {
//   if (artifact.DockerToolArtifact != nil) {
//     fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Removing"), Bold(Gray(artifact.Image+":"+artifact.Tag)))
//     err := RemoveDockerImage(artifact.Image, artifact.Tag)
//     if err != nil {
//       return errors.New("unable to remove docker image: " + err.Error())
//     }

//   } else if (artifact.ExecutableToolArtifact != nil) {
//     fmt.Printf("%s %s %s/%s\n", Blue("==> "), Gray("Removing"), Bold(Gray(tool)), Gray(version.ToString()))
//     dir, err := GetArchiveDir(tool, version, artifact)
//     if err != nil {
//       return errors.New("unable to locate archive dir: " + err.Error())
//     }
//     if _, err := os.Stat(dir); err == nil {
//       err = os.RemoveAll(dir)
//       if err != nil {
//         return errors.New("unable to remove archive folder: " + err.Error())
//       }
//     }

//   }

//   return nil
// }

// /**
//  * Download package and validate
//  */
// func FetchArchive(tool string, version *VersionTriplet, artifact *ToolArtifact) (string, error) {
//   // Install artifact
//   if (artifact.DockerToolArtifact != nil) {
//     return CreateDockerWrapper(tool, version, artifact)
//   } else if (artifact.ExecutableToolArtifact != nil) {
//     if artifact.Source.VCSGitSource != nil {
//       return FetchGitArchive(tool, version, artifact)
//     } else if artifact.Source.HttpSource != nil {
//       return FetchHttpArchive(tool, version, artifact)
//     } else {
//       return "", errors.New(fmt.Sprintf(
//         "no package sources are available for %s/%s", tool, version.ToString()))
//     }
//   }

//   return "", errors.New(fmt.Sprintf(
//     "no known installable artifacts found for %s/%s", tool, version.ToString()))
// }

// /**
//  * Create a shell script that wraps docker
//  */
// func CreateDockerWrapper(tool string, version *VersionTriplet, artifact *ToolArtifact) (string, error) {
//   // Prepare package dir
//   dir, err := GetArchiveDir(tool, version, artifact)
//   if err != nil {
//     return "", err
//   }

//   if _, err := os.Stat(dir); err == nil {
//     os.RemoveAll(dir)
//   }
//   err = os.MkdirAll(dir, 0755)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not create package dir: %s", err.Error()))
//   }

//   // First try to pull the image
//   fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Pulling"), Bold(Gray(artifact.Image+":"+artifact.Tag)))
//   err = PullDockerImage(artifact.Image, artifact.Tag)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not pull the docker image %s", err.Error()))
//   }

//   // Create wrapper script
//   dat := []byte(fmt.Sprintf("#!/bin/bash\ndocker run -it --rm %s %s:%s $*\n",
//     artifact.DockerArgs, artifact.Image, artifact.Tag))
//   err = ioutil.WriteFile(dir + "/run.sh", dat, 0755)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not create wrapper: %s", err.Error()))
//   }

//   // Create a flag that indicates that the downloaded archive is ready for use
//   err = CreateStateFlag(dir + "/.state", artifact)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not set ready flag: %s", err.Error()))
//   }

//   // Return the entrypoint
//   return dir + "/run.sh", nil
// }

// /**
//  * Download git archive and validate
//  */
// func FetchGitArchive(tool string, version *VersionTriplet, artifact *ToolArtifact) (string, error) {
//   // Prepare package dir
//   dir, err := GetArchiveDir(tool, version, artifact)
//   if err != nil {
//     return "", err
//   }
//   if _, err := os.Stat(dir); err == nil {
//     os.RemoveAll(dir)
//   }
//   err = os.MkdirAll(dir, 0755)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not create package dir: %s", err.Error()))
//   }

//   // Git clone
//   _, err = git.PlainClone(dir, false, &git.CloneOptions{
//       URL:      "https://github.com/src-d/go-git",
//       Progress: os.Stdout,
//   })
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not clone git repository: %s", err.Error()))
//   }

//   // Get entrypoint
//   entrypoint := artifact.Entrypoint
//   if entrypoint == "" {
//     entrypoint = GetDefaultEntrypoint()
//   }
//   entrypointFile := dir + "/" + entrypoint

//   // Make sure entrypoint exists
//   if _, err := os.Stat(entrypointFile); os.IsNotExist(err) {
//     return "", errors.New(fmt.Sprintf(
//       "missing tool entrypoint: %s", entrypoint))
//   }

//   // Create a flag that indicates that the downloaded archive is ready for use
//   err = CreateStateFlag(dir + "/.state", artifact)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not set ready flag: %s", err.Error()))
//   }

//   return entrypointFile, nil
// }

// /**
//  * Download web archive and validate
//  */
// func FetchHttpArchive(tool string, version *VersionTriplet, artifact *ToolArtifact) (string, error) {
//   client := RegistryHttpClient(true)
//   srcUrl := artifact.Source.URL
//   fmt.Printf("%s %s\n", Blue("==> "), Bold(Gray(srcUrl)))
//   resp, err := client.Get(srcUrl)
//   if err != nil {
//     return "", errors.New(
//       fmt.Sprintf("could not request %s: %s", srcUrl, err.Error()))
//   }
//   defer resp.Body.Close()

//   // Parse Content-Length header
//   contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not parse Content-Length header: %s", err.Error()))
//   }

//   // Prepare package dir
//   dir, err := GetArchiveDir(tool, version, artifact)
//   if err != nil {
//     return "", err
//   }
//   if _, err := os.Stat(dir); err == nil {
//     os.RemoveAll(dir)
//   }
//   err = os.MkdirAll(dir, 0755)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not create package dir: %s", err.Error()))
//   }

//   // Split streams, so we can calculate the checksum AND extract
//   // while at the same time downloading the file.
//   hasher := sha256.New()
//   data := io.TeeReader(resp.Body, hasher)

//   // Create progress bar
//   bar := pb.New(contentLength).SetUnits(pb.U_BYTES)
//   bar.Start()
//   defer bar.Finish()
//   barStream := bar.NewProxyReader(data)

//   // If this is an archive, extract it
//   entrypointFile := ""
//   if artifact.Extract {
//     err = ExtractTarGz(barStream, dir + "/")
//     if err != nil {
//       return "", errors.New(fmt.Sprintf(
//         "could extract archive: %s", err.Error()))
//     }

//     // Now validate checksum
//     csum := hex.EncodeToString(hasher.Sum(nil))
//     if csum != artifact.Source.Checksum {
//       os.RemoveAll(dir)
//       return "", errors.New("checksum validation failed")
//     }

//     // Get entrypoint
//     entrypoint := artifact.Entrypoint
//     if entrypoint == "" {
//       entrypoint = GetDefaultEntrypoint()
//     }
//     entrypointFile := dir + "/" + entrypoint

//     // Make sure entrypoint exists
//     if _, err := os.Stat(entrypointFile); os.IsNotExist(err) {
//       return "", errors.New(fmt.Sprintf(
//         "missing tool entrypoint: %s", entrypoint))
//     }

//   // Otherwise, just download the file
//   } else {

//     // Download contents to /tool
//     entrypointFile = dir + "/tool"
//     f, err := os.Create(entrypointFile)
//     if err != nil {
//       return "", errors.New(fmt.Sprintf(
//         "could create destination file: %s", err.Error()))
//     }
//     defer f.Close()

//     // Try GZip and fall-back to plaintext
//     writer := bufio.NewWriter(f)
//     err = DownloadFileGz(barStream, writer)
//     if err != nil {
//       return "", errors.New(fmt.Sprintf(
//         "could not process file stream: %s", err.Error()))
//     }
//     writer.Flush()

//     // If we are using an interpreter, create a wrapper file
//     if artifact.Interpreter != "" {
//       dat := []byte(fmt.Sprintf("#!/bin/bash\n%s %s $*\n",
//         artifact.Interpreter, entrypointFile))
//       err = ioutil.WriteFile(dir + "/run.sh", dat, 0755)
//       if err != nil {
//         return "", errors.New(fmt.Sprintf(
//           "could not create wrapper: %s", err.Error()))
//       }
//       entrypointFile = dir + "/run.sh"

//     // Otherwise make it executable
//     } else {
//       f.Chmod(0755)
//     }

//   }

//   // Create a flag that indicates that the downloaded archive is ready for use
//   err = CreateStateFlag(dir + "/.state", artifact)
//   if err != nil {
//     return "", errors.New(fmt.Sprintf(
//       "could not set ready flag: %s", err.Error()))
//   }

//   return entrypointFile, nil
// }
