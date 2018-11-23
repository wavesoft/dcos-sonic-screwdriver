package repository

import (
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "gopkg.in/src-d/go-git.v4"
  "gopkg.in/src-d/go-git.v4/plumbing"
  "io/ioutil"
  "os"
  . "github.com/logrusorgru/aurora"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * Run preparation script from within the package dir
 */
func RunInstallScript(pkgDir string, artifact *registry.ToolArtifact) error {
  if artifact.ExecutableToolArtifact == nil {
    return nil
  }
  if artifact.InstallScript == "" {
    return nil
  }

  // Run install script
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Running"), Bold(Gray("install script")))
  exitcode, err := ExecuteInFolderAndPassthrough(pkgDir, "sh", "-c", artifact.InstallScript)
  if err != nil {
    return err
  }
  if exitcode != 0 {
    return fmt.Errorf("install script failed")
  }

  return nil
}

/**
 * Download and install the source package from the given artifact
 */
func InstallArtifact(pkgDir string, toolDir string, artifact *registry.ToolArtifact) (*InstalledArtifact, error) {
  // If target directory already exists, wipe it and re-create it
  // NOTE: If this function is called, it means that the directory is NOT known
  //       to the system and therefore it should be considered wrong.
  artifactId := ArtifactID(artifact)
  dstDir := pkgDir + "/" + artifactId
  if _, err := os.Stat(dstDir); err == nil {
    os.RemoveAll(dstDir)
  }
  err := os.MkdirAll(dstDir, 0755)
  if err != nil {
    return nil, fmt.Errorf("could not create package dir: %s", err.Error())
  }

  // Download artifact contents
  if (artifact.DockerToolArtifact != nil) {
    err := InstallDockerArtifact(dstDir, artifact)
    if err != nil {
      os.RemoveAll(dstDir)
      return nil, err
    }
  } else if (artifact.ExecutableToolArtifact != nil) {
    if artifact.Source.WebArchiveTarSource != nil {
      err := InstallWebArchiveTarSource(dstDir, artifact)
      if err != nil {
        os.RemoveAll(dstDir)
        return nil, err
      }
    } else if artifact.Source.WebFileSource != nil {
      err := InstallWebFileSource(dstDir, artifact)
      if err != nil {
        os.RemoveAll(dstDir)
        return nil, err
      }
    } else if artifact.Source.VCSGitSource != nil {
      err := InstallVcsGitSource(dstDir, artifact)
      if err != nil {
        os.RemoveAll(dstDir)
        return nil, err
      }
    } else {
      os.RemoveAll(dstDir)
      return nil, fmt.Errorf("unknown web artifact source type")
    }
  } else {
    os.RemoveAll(dstDir)
    return nil, fmt.Errorf("unknown artifact type")
  }

  // If there is an install script, run it now
  err = RunInstallScript(dstDir, artifact)
  if err != nil {
    return nil, err
  }

  // If there is an uninstall script, create an uninstall wrapper now
  err = CreateUninstallScript(dstDir, artifact)
  if err != nil {
    return nil, err
  }

  // If we reached this point, the operation was successful, so
  // dump the artifact state in the directory
  instArtifact := &InstalledArtifact{
    artifactId,
    dstDir,
    0,
  }
  instArtifact.SetRegistryArtifact(artifact)

  // Return the installed artifact
  return instArtifact, nil
}

/**
 * Download & Install a docker image
 */
func InstallDockerArtifact(dstDir string, artifact *registry.ToolArtifact) error {
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Pulling"), Bold(Gray(artifact.Image+":"+artifact.Tag)))
  return DockerPullImage(artifact.Image, artifact.Tag)
}

/**
 * Download & Install a Tar Archive
 */
func InstallWebFileSource(dstDir string, artifact *registry.ToolArtifact) error {
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Downloading"), Bold(Gray(artifact.Source.FileURL)))
  return Download(artifact.Source.FileURL, WithoutCompression).
         AndShowProgress("").
         AndValidateChecksum(artifact.Source.FileChecksum).
         AndDecompressIfCompressed().
         EventuallyWriteTo(dstDir + "/run")
}

/**
 * Download & Install a Tar Archive
 */
func InstallWebArchiveTarSource(dstDir string, artifact *registry.ToolArtifact) error {
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Downloading"), Bold(Gray(artifact.Source.TarURL)))
  return Download(artifact.Source.TarURL, WithoutCompression).
         AndShowProgress("").
         AndValidateChecksum(artifact.Source.TarChecksum).
         AndDecompressIfCompressed().
         EventuallyUntarTo(dstDir, 1)
}

/**
 * Download & Install a Git repository
 */
func InstallVcsGitSource(pkgDir string, artifact *registry.ToolArtifact) error {
  fmt.Printf("%s %s %s\n", Blue("==> "), Gray("Cloning"), Bold(Gray(artifact.Source.GitURL)))

  // Find the branch to fetch
  remoteBranch := artifact.Source.GitBranch
  if remoteBranch == "" {
    remoteBranch = "refs/heads/master"
  }

  // Clone the repository
  _, err := git.PlainClone(pkgDir, false, &git.CloneOptions{
      URL:            artifact.Source.GitURL,
      SingleBranch:   true,
      ReferenceName:  plumbing.ReferenceName(remoteBranch),
      Progress:       os.Stdout,
  })
  if err != nil {
    return fmt.Errorf("git clone error: %s", err.Error())
  }

  // Success
  return nil
}

/**
 * Install a specific tool version
 */
func InstallToolVersion(toolDir string,
  version *registry.ToolVersion,
  artifact *registry.ToolArtifact,
  installedArtifact *InstalledArtifact) (*InstalledVersion, error) {

  // Prepare tool directories
  toolVerDir := toolDir + "/" + version.Version.ToString() + "-" + installedArtifact.ID
  err := os.MkdirAll(toolVerDir, 0755)
  if err != nil {
    return nil, fmt.Errorf("unable to create the tool directory: %s", err.Error())
  }

  // Create wrappers
  if (artifact.DockerToolArtifact != nil) {
    err := CreateDockerWrapper(toolVerDir, artifact)
    if err != nil {
      os.RemoveAll(toolVerDir)
      return nil, err
    }
  } else if (artifact.ExecutableToolArtifact != nil) {
    if artifact.Interpreter != nil {
      err := CreateInterpreterWrapper(toolVerDir, artifact, installedArtifact)
      if err != nil {
        os.RemoveAll(toolVerDir)
        return nil, err
      }
    } else {
      err := CreateBinaryWrapper(toolVerDir, artifact, installedArtifact)
      if err != nil {
        os.RemoveAll(toolVerDir)
        return nil, err
      }
    }
  } else {
    os.RemoveAll(toolVerDir)
    return nil, fmt.Errorf("unknown artifact type")
  }

  // increment artifact references
  installedArtifact.References += 1

  return &InstalledVersion{
    version.Version,
    installedArtifact,
    toolVerDir,
  }, nil
}

/**
 * Create a docker wrapper script
 */
func CreateDockerWrapper(toolDir string,
  artifact *registry.ToolArtifact) error {
  execPath := toolDir + "/run"

  // Create wrapper script
  dat := []byte(fmt.Sprintf("#!/bin/sh\ndocker run -it --rm %s %s:%s $*\n",
    artifact.DockerArgs, artifact.Image, artifact.Tag))
  err := ioutil.WriteFile(execPath, dat, 0755)
  if err != nil {
    return fmt.Errorf("could not create wrapper: %s", err.Error())
  }

  return nil
}

/**
 * Create a script that executes the artifact entrypoint
 */
func CreateInterpreterWrapper(toolDir string,
  artifact *registry.ToolArtifact,
  installedArtifact *InstalledArtifact) error {
  execPath := toolDir + "/run"

  // Find entrypoint
  pkgDir := installedArtifact.Folder
  entryPoint := artifact.Entrypoint
  if entryPoint == "" {
    entryPoint = "run"
  }

  // Create wrapper script
  dat, err := GetInterpreterWrapperContents(
    pkgDir,
    toolDir,
    entryPoint,
    artifact.Interpreter,
    artifact.Workdir,
    artifact.Environment,
  )
  if err != nil {
    return fmt.Errorf("could not create wrapper: %s", err.Error())
  }

  err = ioutil.WriteFile(execPath, dat, 0755)
  if err != nil {
    return fmt.Errorf("could not create wrapper: %s", err.Error())
  }

  return nil
}

/**
 * Return the path to the entrypoint in the package
 */
func CreateBinaryWrapper(toolDir string,
  artifact *registry.ToolArtifact,
  installedArtifact *InstalledArtifact) error {

  pkgDir := installedArtifact.Folder
  entryPoint := artifact.Entrypoint
  if entryPoint == "" {
    entryPoint = "run"
  }

  execPath := pkgDir + "/" + entryPoint
  if _, err := os.Stat(execPath); err != nil {
    return fmt.Errorf("unable to find entrypoint: %s", entryPoint)
  }

  // Make sure it's executable
  os.Chmod(execPath, 0755)

  // Create symlink into the tool folder
  binPath := toolDir + "/run"
  os.Symlink(execPath, binPath)

  return nil
}
