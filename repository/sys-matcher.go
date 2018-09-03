package repository

import (
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "fmt"
)

/**
 * Return the failed requirement error message
 */
func CollectRequirementErrors(req registry.ArtifactRequirement) []string {
  var errors []string = nil

  if (req.CommandRequirement != nil) {
    if !SysHasCommand(req.Command) {
      errors = append(errors, fmt.Sprintf("required command '%s' does not exist", req.Command))
    }
  }
  if (req.ExecRequirement != nil) {
    if !SysCommandExitsWithZero(req.Exec) {
      errors = append(errors, fmt.Sprintf("pilot command '%s' exited with error", req.Exec))
    }
  }

  return errors
}

/**
 * Collect all the failed requirements as a list of error messages
 */
func CollectAllRequirementErrors(artifact *registry.ToolArtifact) []string {
  var errors []string = nil

  if (artifact.ExecutableToolArtifact == nil) {
    return errors
  }
  if len(artifact.Require) == 0 {
    return errors
  }

  for _, req := range artifact.Require {
    errors = append(errors, CollectRequirementErrors(req)...)
  }

  return errors
}

/**
 * Collect all incompatibilities as a list of error messages
 */
func CollectIncompatibilities(artifact *registry.ToolArtifact) []string {
  var errors []string = nil

  switch artifact.Type {
    case registry.Docker:

      // Check missing docker
      if !DockerIsAvailable() {
        return []string{ "`docker` command is not available" }
      }

    case registry.Executable:

      // Check interpreter incompatibilities
      if artifact.Interpreter != nil {
        return CollectInterpreterIncompatibilities(artifact.Interpreter)
      }

      // Check binary incompatibilities
      if (artifact.Arch != "*" && artifact.Arch != SysGetArch()) {
        errors = append(errors, fmt.Sprintf(
          "architecture '%s' is incompatible with your system", artifact.Arch))
      }
      if (artifact.Platform != "*" && artifact.Platform != SysGetPlatform()) {
        errors = append(errors, fmt.Sprintf(
          "platform '%s' is incompatible with your system", artifact.Platform))
      }
      if errors != nil {
        return errors
      }

  }

  // Check requirements
  return CollectAllRequirementErrors(artifact)
}

/**
 * Get an artifact that matches our system
 */
func FindFirstRunableArtifact(artifacts registry.ToolArtifacts) (*registry.ToolArtifact, [][]string) {
  var errors [][]string = nil

  for _, artifact := range artifacts {

    // Collect errors for this artifact
    artifactErrs := CollectIncompatibilities(&artifact)
    if artifactErrs == nil {
      return &artifact, nil
    }

    // Collect errors
    errors = append(errors, artifactErrs)

  }

  return nil, errors
}
