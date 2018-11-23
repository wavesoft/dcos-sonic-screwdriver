package repository

import (
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
)

func IsShellInterpreterValid(interpreter *registry.ExecutableInterpreter) bool {
  if (interpreter.ShellInterpreter != nil) {
    return SysHasCommand(interpreter.Shell)
  }
  return true
}

/**
 * Return the wrapper contents for running `entrypoint` from within the sandbox
 */
func ShellCreateWrapper(sandboxPath string,
  toolDir string,
  entrypoint string,
  interpreter *registry.ExecutableInterpreter,
  envPreparation string) ([]byte, error) {

  // Create a wrapper to run the script from within the sandbox
  expr := fmt.Sprintf("#!/bin/sh\n%s\n%s %s/%s $*\n",
    envPreparation, interpreter.Shell, sandboxPath, entrypoint)
  return []byte(expr), nil
}
