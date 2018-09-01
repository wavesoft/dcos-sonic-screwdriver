package repository

import (
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "strings"
  . "github.com/logrusorgru/aurora"
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
  toolDir string,
  interpreter *registry.ExecutableInterpreter) error {

  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Preparing"), Bold(Gray("python sandbox")))

  venvPath := toolDir + "/python-venv"

  if strings.HasPrefix(interpreter.Python, "python2") {
    if !SysHasCommand("virtualenv") {
      return fmt.Errorf("Python2 packages require `virtualenv` to be installed")
    }

    // Create virtualenv
    exitcode, err := ExecuteSilently("virtualenv", "-p", interpreter.Python, venvPath)
    if err != nil {
      return fmt.Errorf("cannot create python2 sandbox: %s", err.Error())
    }
    if exitcode != 0 {
      return fmt.Errorf("cannot create python2 sandbox: process exited with %d", exitcode)
    }

  } else if strings.HasPrefix(interpreter.Python, "python3") {

    // Create virtualenv
    exitcode, err := ExecuteSilently("python3", "-m", "venv", venvPath)
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
      "(source %s/bin/activate; pip install -r %s)",
      venvPath,
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
      "(source %s/bin/activate; pip install %s)",
      venvPath,
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
func PythonCreateWrapper(sandboxPath string,
  toolDir string,
  entrypoint string,
  interpreter *registry.ExecutableInterpreter) ([]byte, error) {

  // Create sandbox
  err := PythonPrepareSandbox(sandboxPath, toolDir, interpreter)
  if err != nil {
    return nil, err
  }

  venvPath := toolDir + "/python-venv"

  // Create a wrapper to run the script from within the sandbox
  expr := fmt.Sprintf("#!/bin/sh\nsource %s/bin/activate\n%s/bin/python %s/%s $*\n",
    venvPath, venvPath, sandboxPath, entrypoint)
  return []byte(expr), nil
}
