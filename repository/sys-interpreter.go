package repository

import (
  "fmt"
  "strings"
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
  if (interpreter.JavaInterpreter != nil) {
    return IsJavaInterpreterValid(interpreter)
  }

  return false
}

/**
 * Return the shell script contents that prepare the environment according to the args given
 */
func GetEnvironmentPreparationContents(pkgDir string,
  toolDir string,
  workDir string,
  envVars map[string]string) string {

  var builder strings.Builder

  // Change to the appropriate directory, as configured
  if workDir != "" {
    // Replace templates
    workDir = strings.Replace(workDir, "$ARTIFACT", pkgDir, -1)
    workDir = strings.Replace(workDir, "$TOOL", toolDir, -1)

    // Compose the cd command, escaping stray single quotes
    builder.WriteString(fmt.Sprintf("cd '%s'\n", strings.Replace(workDir, `'`, `'"'"'`, -1)))
  }

  // Expose environment variables
  for name, value := range envVars {
    builder.WriteString(fmt.Sprintf("export %s='%s'\n", name, strings.Replace(value, `'`, `'"'"'`, -1)))
  }

  // Return the composed string
  return builder.String()
}

/**
 * Return a wrapper script contents that runs the specified file
 */
func GetInterpreterWrapperContents(pkgDir string,
    toolDir string,
    entrypoint string,
    interpreter *registry.ExecutableInterpreter,
    workDir string,
    envVars map[string]string) ([]byte, error) {

  // Get the environment preparation shell contents
  envPreparation := GetEnvironmentPreparationContents(pkgDir, toolDir, workDir, envVars)

  // Return the interpreter wrapper
  if (interpreter.PythonInterpreter != nil) {
    return PythonCreateWrapper(pkgDir, toolDir, entrypoint, interpreter, envPreparation)
  }
  if (interpreter.ShellInterpreter != nil) {
    return ShellCreateWrapper(pkgDir, toolDir, entrypoint, interpreter, envPreparation)
  }
  if (interpreter.JavaInterpreter != nil) {
    return JavaCreateWrapper(pkgDir, toolDir, entrypoint, interpreter, envPreparation)
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
  if (interpreter.JavaInterpreter != nil) {
    return "java"
  }

  return "unknown"
}

/**
 * Check for interpreter incompatibilities
 */
func CollectInterpreterIncompatibilities(interpreter *registry.ExecutableInterpreter) []string {
  if (interpreter.PythonInterpreter != nil) {
    return CollectPythonIncompatibilities(interpreter)
  }

  return nil
}
