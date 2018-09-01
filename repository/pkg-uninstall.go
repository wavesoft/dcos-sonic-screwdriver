package repository

import (
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "os"
  "fmt"
  . "github.com/logrusorgru/aurora"
)

/**
 * Uninstall an artifact
 */
func UninstallArtifact(artifact *InstalledArtifact) error {
  dstDir := artifact.Folder
  if _, err := os.Stat(dstDir); err != nil {
    return fmt.Errorf("the artifact does not exist")
  }

  // Load the registry artifact
  registryArtifact, err := artifact.GetRegistryArtifact()
  if err != nil {
    return fmt.Errorf("could not read state: %s", err.Error())
  }

  // Handle removal of docker artifacts
  if (registryArtifact.DockerToolArtifact != nil) {
    err := UnininstallDockerArtifact(dstDir, registryArtifact)
    if err != nil {
      return err
    }
  } else {
    err := UninstallArchiveFolder(dstDir, registryArtifact)
    if err != nil {
      return err
    }
  }

  // Remove directory
  return os.RemoveAll(dstDir)
}

/**
 * Uninstall a docker artifact
 */
func UnininstallDockerArtifact(dstDir string, artifact *registry.ToolArtifact) error {
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Removing"), Bold(Gray(artifact.Image+":"+artifact.Tag)))
  return DockerRemoveImage(artifact.Image, artifact.Tag)
}

/**
 * Uninstall an archive folder
 */
func UninstallArchiveFolder(dstDir string, artifact *registry.ToolArtifact) error {
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Deleting"), Bold(Gray("local files")))
  return nil
}

/**
 * Install a specific tool version
 */
func UninstallToolVersion(version *InstalledVersion) error {
  return os.RemoveAll(version.Folder)
}

/**
 * Unistall the tool folder
 */
func UninstallTool(tool *InstalledTool) error {
  return os.Remove(tool.Folder)
}
