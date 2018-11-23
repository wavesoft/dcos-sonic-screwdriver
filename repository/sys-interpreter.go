package repository

import (
  "fmt"
  "regexp"
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
 * Replaces templates with the full paths
 */
func ReplacePathTemplates(expr string, pkgDir string, toolDir string) string {
  r := regexp.MustCompile(`%\w+%`)
  return r.ReplaceAllStringFunc(expr, func(m string) string {
    switch strings.ToLower(m[1:len(m)-1]) {
      case "artifact":
        return pkgDir
      case "tool":
        return toolDir
      default:
        return ""
    }
  })
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
    builder.WriteString(fmt.Sprintf("cd \"%s\"\n", strings.Replace(
      ReplacePathTemplates(workDir, pkgDir, toolDir),
      `"`, `"'"'"`, -1,
    )))
  }

  // Expose environment variables
  for name, value := range envVars {
    builder.WriteString(
      fmt.Sprintf("export %s=\"%s\"\n",
        name,
        strings.Replace(
          ReplacePathTemplates(value, pkgDir, toolDir),
          `"`, `"'"'"`, -1,
        ),
      ),
    )
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
