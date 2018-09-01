package repository

import (
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
)

func IsJavaInterpreterValid(interpreter *registry.ExecutableInterpreter) bool {
  if (interpreter.JavaInterpreter != nil) {
    return SysHasCommand("java")
  }
  return true
}

/**
 * Return the wrapper contents for running `entrypoint` from within the sandbox
 */
func JavaCreateWrapper(sandboxPath string,
  toolDir string,
  entrypoint string,
  interpreter *registry.ExecutableInterpreter) ([]byte, error) {

  // Create a wrapper to run the script from within the sandbox
  expr := fmt.Sprintf("#!/bin/sh\njava %s -jar %s/%s $*\n",
    interpreter.JavaArgs, sandboxPath, entrypoint)
  return []byte(expr), nil
}
