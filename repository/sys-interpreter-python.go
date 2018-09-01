package repository

import (
  "fmt"
  "strings"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

func IsPythonInterpreterValid(interpreter *registry.ExecutableInterpreter) bool {
  if (interpreter.PythonInterpreter != nil) {
    return SysHasCommand(interpreter.Python)
  }
  return true
}

/**
 * Perform the required actions to prepare a sandbox for the interpreter
 */
func PythonPrepareSandbox(sandboxPath string,
  interpreter *registry.ExecutableInterpreter) error {

  if strings.HasPrefix(interpreter.Python, "python2") {
    if !SysHasCommand("virtualenv") {
      return fmt.Errorf("Python2 packages require `virtualenv` to be installed")
    }

    // Create virtualenv
    exitcode, err := ExecuteSilently("virtualenv", "-p", interpreter.Python, sandboxPath + "/.venv")
    if err != nil {
      return fmt.Errorf("cannot create python2 sandbox: %s", err.Error())
    }
    if exitcode != 0 {
      return fmt.Errorf("cannot create python2 sandbox: process exited with %d", exitcode)
    }

  } else if strings.HasPrefix(interpreter.Python, "python3") {

    // Create virtualenv
    exitcode, err := ExecuteSilently("python3", "-m", "venv", sandboxPath + "/.venv")
    if err != nil {
      return fmt.Errorf("cannot create python3 sandbox: %s", err.Error())
    }
    if exitcode != 0 {
      return fmt.Errorf("cannot create python3 sandbox: process exited with %d", exitcode)
    }

  } else {
    return fmt.Errorf("unknown python version: `%s`", interpreter.Python)
  }

  // Install requirements
  if interpreter.InstReq != "" {
    exitcode, err := ShellExecuteInFolderAndPassthrough(sandboxPath, fmt.Sprintf(
      "(source .venv/bin/activate; pip install -r %s)",
      interpreter.InstReq))
    if err != nil {
      return fmt.Errorf("cannot install requirements: %s", err.Error())
    }
    if exitcode != 0 {
      return fmt.Errorf("cannot install requirements: process exited with %d", exitcode)
    }
  }
  if interpreter.InstPip != "" {
    exitcode, err := ShellExecuteInFolderAndPassthrough(sandboxPath, fmt.Sprintf(
      "(source .venv/bin/activate; pip install %s)",
      interpreter.InstPip))
    if err != nil {
      return fmt.Errorf("cannot install requirements: %s", err.Error())
    }
    if exitcode != 0 {
      return fmt.Errorf("cannot install requirements: process exited with %d", exitcode)
    }
  }

  return nil
}

/**
 * Return the wrapper contents for running `entrypoint` from within the sandbox
 */
func PythonCreateWrapper(sandboxPath string, entrypoint string,
  interpreter *registry.ExecutableInterpreter) []byte {

  expr := fmt.Sprintf("#!/bin/sh\nsource %s/.venv/bin/activate\n%s/.venv/bin/python %s/%s $*\n",
    sandboxPath, sandboxPath, sandboxPath, entrypoint)
  return []byte(expr)
}
