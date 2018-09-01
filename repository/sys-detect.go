package repository

import (
  "os/exec"
  "runtime"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * Returns the name of the system
 */
func SysGetPlatform() string {
  return runtime.GOOS
}

/**
 * Return the system CPU architecture
 */
func SysGetArch() string {
  return runtime.GOARCH
}

/**
 * Checks if we can run the given command
 */
func SysHasCommand(name string) bool {
  _, err := exec.LookPath(name)
  if err != nil {
    return false
  }
  return true
}

/**
 * Run command and check that exit code is zero
 */
func SysCommandExitsWithZero(shellExpr string) bool {
  exitcode, err := ExecuteSilently("sh", "-c", shellExpr)
  if err != nil {
    return false
  }
  return exitcode == 0
}
