package repository

import (
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "fmt"
)

/**
 * Check if a particular requirement is satisfied
 */
func RequirementSatisfied(req registry.ArtifactRequirement) bool {
  if (req.CommandRequirement != nil) {
    return SysHasCommand(req.Command)
  }
  if (req.ExecRequirement != nil) {
    return SysCommandExitsWithZero(req.Exec)
  }

  return false
}

/**
 * Check if all requirements for this artifact are satisfied
 */
func CheckAllRequirements(artifact *registry.ToolArtifact) bool {
  if (artifact.ExecutableToolArtifact == nil) {
    return true
  }
  if len(artifact.Require) == 0 {
    return true
  }

  for _, req := range artifact.Require {
    if RequirementSatisfied(req) {
      return true
    }
  }
  return false
}

/**
 * Get an artifact that matches our system
 */
func FindFirstRunableArtifact(artifacts registry.ToolArtifacts) (*registry.ToolArtifact, error) {
  for _, artifact := range artifacts {
    switch artifact.Type {

      // Check for docker containers
      case registry.Docker:
        if DockerIsAvailable() {
          if CheckAllRequirements(&artifact) {
            return &artifact, nil
          }
        }

      // Check for executables
      case registry.Executable:

        // If the executable is interpreted, we don't care about the
        // CPU architecture. Just make sure we have the interpreter
        if artifact.Interpreter != nil && InterpreterIsValid(artifact.Interpreter) {
          if CheckAllRequirements(&artifact) {
            return &artifact, nil
          }
        }

        // Otherwise, we need to match platform/cpu
        if (artifact.Arch == "*" || artifact.Arch == SysGetArch()) && (artifact.Platform == "*" || artifact.Platform == SysGetPlatform()) {
          if CheckAllRequirements(&artifact) {
            return &artifact, nil
          }
        }

    }
  }

  return nil, fmt.Errorf("No runnable artifact found for your system")
}
