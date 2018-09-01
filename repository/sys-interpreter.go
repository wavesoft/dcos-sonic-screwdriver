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
    return IsPythonInterpreterValid(interpreter)
  }
  if (interpreter.ShellInterpreter != nil) {
    return IsShellInterpreterValid(interpreter)
  }

  return false
}

/**
 * Return a wrapper script contents that runs the specified file
 */
func GetInterpreterWrapperContents(sandboxPath string,
    toolDir string,
    entrypoint string,
    interpreter *registry.ExecutableInterpreter) ([]byte, error) {
  if (interpreter.PythonInterpreter != nil) {
    return PythonCreateWrapper(sandboxPath, toolDir, entrypoint, interpreter)
  }
  if (interpreter.ShellInterpreter != nil) {
    return ShellCreateWrapper(sandboxPath, toolDir, entrypoint, interpreter)
  }

  return nil, fmt.Errorf("unsupported interpreter")
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
