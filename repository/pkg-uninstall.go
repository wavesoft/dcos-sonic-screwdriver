package repository

import (
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "os"
  "fmt"
  "io/ioutil"
  . "github.com/logrusorgru/aurora"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * Return the contents of the uninstall script that will be placed on the
 * uninstall directory and will be used before removing the tool.
 */
func CreateUninstallScript(pkgDir string, artifact *registry.ToolArtifact) error {
  if artifact.ExecutableToolArtifact == nil {
    return nil
  }
  if artifact.UninstallScript == "" {
    return nil
  }

  // Create a shell script wrapper for the uninstall script
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Creating"), Bold(Gray("uninstall script")))
  dat := fmt.Sprintf("#!/bin/sh\n%s\n", artifact.UninstallScript)

  // Write down the uninstall script
  filePath := pkgDir + "/.uninstall"
  err := ioutil.WriteFile(filePath, []byte(dat), 0755)
  if err != nil {
    return fmt.Errorf("could not create uninstall script: %s", err.Error())
  }

  return nil
}

/**
 * Run preparation script from within the package dir
 */
func RunUninstallScript(pkgDir string) error {
  filePath := pkgDir + "/.uninstall"
  if _, err := os.Stat(filePath); err != nil {
    return nil
  }

  // Run install script
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Running"), Bold(Gray("uninstall script")))
  exitcode, err := ExecuteInFolderAndPassthrough(pkgDir, "sh", "-c", filePath)
  if err != nil {
    return err
  }
  if exitcode != 0 {
    return fmt.Errorf("uninstall script failed")
  }

  return nil
}

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

  // If there is an uninstall script, run it now
  err = RunUninstallScript(dstDir)
  if err != nil {
    return err
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
