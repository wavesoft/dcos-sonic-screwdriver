package registry

import (
  "errors"
  "path/filepath"
  "os/exec"
  "runtime"
  "io"
  "os"
)

/**
 * Returns the name of the system
 */
func getSystmePlatform() string {
  return runtime.GOOS
}

/**
 * Return the system CPU architecture
 */
func getSystemArch() string {
  return runtime.GOARCH
}

/**
 * Checks if we can run the given command
 */
func isCommandAvailable(name string) bool {
  _, err := exec.LookPath(name)
  if err != nil {
    return false
  }
  return true
}

/**
 * Check if docker found in the system
 */
func HasDocker() bool {
  return isCommandAvailable("docker")
}

/**
 * Pull docker image, while echoing progress on terminal
 */
func PullDockerImage(image string, tag string) error {
  cmd := exec.Command("docker", "pull", image+":"+tag)
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return errors.New("Unable to open StdOut Pipe: " + err.Error())
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    return errors.New("Unable to open StdErr Pipe: " + err.Error())
  }
  if err := cmd.Start(); err != nil {
    return err
  }

  go func() {
      _, _ = io.Copy(os.Stdout, stdout)
  }()

  go func() {
      _, _ = io.Copy(os.Stderr, stderr)
  }()

  if err := cmd.Wait(); err != nil {
    return err
  }
  return nil
}

/**
 * Create a symbolic link on the user bin directory
 */
 func CreateBinSymlink(fullPath string, name string) error {
  linkTarget := GetSymlinkBinPath() + "/" + name
  return os.Symlink(fullPath, linkTarget)
 }

/**
 * Remove a symbolic link to the user bin directory
 */
func RemoveBinSymlink(name string) error {
  linkTarget := GetSymlinkBinPath() + "/" + name
  return os.Remove(linkTarget)
}

/**
 * Check if a symbolic link already exists in the user bin directory
 */
func HasBinSymlink(name string) bool {
  linkTarget := GetSymlinkBinPath() + "/" + name
  if _, err := os.Stat(linkTarget); err == nil {
    return true
  }
  return false
}

/**
 * Calculate the size of the designated directory
 */
func DirSize(path string) (int64, error) {
  var size int64
  err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
    if !info.IsDir() {
      size += info.Size()
    }
    return err
  })
  return size, err
}

/**
 * Get an artifact that matches our system
 */
func FindFirstRunableArtifact(version ToolVersion) (*ToolArtifact, error) {
  for _, artifact := range version.Artifacts {
    switch artifact.Type {

      // Check for docker containers
      case Docker:
        if HasDocker() {
          return &artifact, nil
        }

      // Check for executables
      case Executable:

        // If the executable is interpreted, we don't care about the
        // CPU architecture. Just make sure we have the interpreter
        if artifact.Interpreter != "" && IsValidInterpreter(artifact.Interpreter) && isCommandAvailable(artifact.Interpreter) {
          return &artifact, nil
        }

        // Otherwise, we need to match platform/cpu
        if (artifact.Arch == "*" || artifact.Arch == getSystemArch()) && (artifact.Platform == "*" || artifact.Platform == getSystmePlatform()) {
          return &artifact, nil
        }

    }
  }

  return nil, errors.New("No runnable artifact found for your system")
}
