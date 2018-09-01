package repository

import (
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
)

/**
 * Check if the given interpreter is valid
 */
func InterpreterIsValid(interpreter *registry.ExecutableInterpreter) bool {
  if (interpreter.PythonInterpreter != nil) {
    return SysHasCommand(interpreter.Python)
  }

  return false
}

/**
 * Perform the required actions to prepare a sandbox for the interpreter
 */
func PrepareInterpreterSandbox(sandboxPath string, interpreter *registry.ExecutableInterpreter) error {
  if (interpreter.PythonInterpreter != nil) {
    return PythonPrepareSandbox(sandboxPath, interpreter)
  }

  return fmt.Errorf("unknown interpreter specified")
}

/**
 * Return a wrapper script contents that runs the specified file
 */
func GetInterpreterWrapperContents(sandboxPath string, entrypoint string, interpreter *registry.ExecutableInterpreter) []byte {
  if (interpreter.PythonInterpreter != nil) {
    return PythonCreateWrapper(sandboxPath, entrypoint, interpreter)
  }

  return nil
}

/**
 *  Return the interpreter name
 */
func InterpreterName(interpreter *registry.ExecutableInterpreter) string {
  if (interpreter.PythonInterpreter != nil) {
    return interpreter.Python
  }
  if (interpreter.ShellInterpreter != nil) {
    return interpreter.Shell
  }

  return "unknown"
}
