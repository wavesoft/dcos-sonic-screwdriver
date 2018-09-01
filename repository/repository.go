package repository

import (
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "io/ioutil"
  "path/filepath"
  "path"
  "fmt"
  "strings"
  "os"
  . "github.com/logrusorgru/aurora"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)


/**
 * Load a tool description based on what we will find in the `toolDir`
 */
func LoadRepositoryTool(toolDir string,
    pkgDir string,
    sources *map[string]*InstalledArtifact) (*InstalledTool, error) {

  tool := new(InstalledTool)
  tool.Folder = toolDir
  tool.Name = path.Base(toolDir)

  files, err := ioutil.ReadDir(toolDir)
  if err != nil {
    return tool, err
  }
  for _, f := range files {

    // Extract version/source pair
    parts := strings.Split(f.Name(), "-")
    if len(parts) == 1 {

      // TODO: Version 0.1.1 of the tool was not separating sources from tools.
      // Therefore we have to create a virtual source in the same directory
      // with the tool.

    } else if (len(parts) == 2) {

      // Parse version
      versionDir := toolDir + "/" + f.Name()
      version, err := VersionFromString(parts[0])
      if err != nil {
        return tool, fmt.Errorf("Could not parse tool '%s' version '%s': %s",
          filepath.Base(toolDir), f.Name(), err.Error())
      }

      // Get source folder
      sourceId := parts[1]
      sourceDir := pkgDir + "/" + sourceId
      if _, err := os.Stat(sourceDir); err != nil {
        return tool, fmt.Errorf("Source package '%s' was not found", sourceDir)
      }

      // Make sure we have a source in the sources
      var foundSource *InstalledArtifact
      var ok bool
      if foundSource, ok = (*sources)[sourceId]; !ok {
        foundSource = &InstalledArtifact{sourceId, sourceDir, 1}
        (*sources)[sourceId] = foundSource
      } else {
        foundSource.References += 1
      }

      // Keep track of this version
      tool.Versions = append(tool.Versions,
        InstalledVersion{ *version, foundSource, versionDir })

    }
  }

  return tool, nil
}

/**
 * Reads the repository directory and populates the repository structure
 */
func LoadRepository(baseDir string) (*Repository, error) {
  repository := new(Repository)
  repository.BaseDir = baseDir
  repository.Tools = make(map[string]*InstalledTool)
  repository.Sources = make(map[string]*InstalledArtifact)

  // Check if the tool folder is missing
  if _, err := os.Stat(baseDir); err != nil {
    return repository, nil
  }

  // Compute the directory names
  toolDir := baseDir + "/tools"
  pkgDir := baseDir + "/pkg"

  // List versions
  files, err := ioutil.ReadDir(toolDir)
  if err != nil {
    return repository, err
  }
  for _, f := range files {
    tool, err := LoadRepositoryTool(toolDir + "/" + f.Name(), pkgDir, &repository.Sources)
    if err != nil {
      return nil, err
    }

    repository.Tools[f.Name()] = tool
  }

  return repository, nil
}

/**
 * Find the version of the specified tool
 */
func (repo *Repository) FindToolVersion(tool string,
  version VersionTriplet) *InstalledVersion {

  if tool, ok := repo.Tools[tool]; ok {
    return tool.FindVersion(version)
  }

  return nil
}

/**
 * Scan the list of versions and find the given version
 */
func (tool *InstalledTool) FindVersion(version VersionTriplet) *InstalledVersion {
  for _, ver := range tool.Versions {
    if ver.Version.Equals(version) {
      return &ver
    }
  }
  return nil
}

/**
 * Find the installed version by matching the link of every tool in the registry
 */
func (repo *Repository) FindToolFromLink(linkPath string) (*InstalledTool, *InstalledVersion) {
  for _, tool := range repo.Tools {
    for _, ver := range tool.Versions {
      toolPath := ver.GetExecutablePath()
      if toolPath == linkPath {
        return tool, &ver
      }
    }
  }

  return nil, nil
}

/**
 * Find the installed artifact
 */
func (repo *Repository) FindArtifact(artifact *registry.ToolArtifact) *InstalledArtifact {
  aid := ArtifactID(artifact)
  if artifact, ok := repo.Sources[aid]; ok {
    return artifact
  }

  return nil
}

/**
 * Remove a version from the tool list
 */
func (tool *InstalledTool) RemoveVersion(ver *InstalledVersion) {
  newVersions := tool.Versions[:0]
  for _, x := range tool.Versions {
    if !x.Version.Equals(ver.Version) {
      newVersions = append(newVersions, x)
    }
  }
  tool.Versions = newVersions
}

/**
 * Return the full path to the executable file of the tool
 */
func (ver *InstalledVersion) GetExecutablePath() string {
  return ver.Folder + "/run"
}

/**
 * Check if the tool exists in the repository
 */
func (repo *Repository) IsToolInstalled(tool string) bool {
  if _, ok := repo.Tools[tool]; ok {
    return true
  }
  return false
}

/**
 * Check if there is at least one version in the tool
 */
func (repo *InstalledTool) HasInstalledVersions() bool {
  return len(repo.Versions) > 0
}

/**
 * Check if a specific tool version exists
 */
func (repo *Repository) IsToolVersionInstalled(tool string, version *VersionTriplet) bool {
  if toolRef, ok := repo.Tools[tool]; ok {
    for _, x := range toolRef.Versions {
      if x.Version.Equals(*version) {
        return true
      }
    }
  }
  return false
}

/**
 * Install the specified tool in the repository
 */
func (repo *Repository) InstallToolVersion(tool string,
  version *registry.ToolVersion,
  artifact *registry.ToolArtifact) (*InstalledVersion, error) {
  var err error

  fmt.Printf("%s %s %s\n",
    Bold(Green("==> ")),
    Bold(Gray("Add")),
    Bold(Green(tool + "/" + version.Version.ToString())))

  // Find or install the related tool artifact
  installedArtifact := repo.FindArtifact(artifact)
  if installedArtifact == nil {
    installedArtifact, err = InstallArtifact(repo.BaseDir + "/pkg", artifact)
    if err != nil {
      return nil, err
    }

    // Keep track of the installed artifact
    repo.Sources[ArtifactID(artifact)] = installedArtifact
  }

  // Install the tool
  toolDir := repo.BaseDir + "/tools/" + SanitizedToolName(tool)
  installedVersion, err := InstallToolVersion(
    toolDir,
    version,
    artifact,
    installedArtifact,
  )
  if err != nil {
    return nil, err
  }

  // Put the version in the registry
  if toolRef, ok := repo.Tools[tool]; ok {
    toolRef.Versions = append(toolRef.Versions, *installedVersion)
  } else {
    toolRef := new(InstalledTool)
    toolRef.Folder = toolDir
    toolRef.Versions = append(toolRef.Versions, *installedVersion)
    toolRef.Name = tool
  }

  // Return the new pointers
  return installedVersion, nil
}

/**
 * Install the specified tool version from the repository
 */
func (repo *Repository) UninstallToolVersion(tool *InstalledTool, version *InstalledVersion) error {

  fmt.Printf("%s %s %s\n",
    Bold(Red("==> ")),
    Bold(Gray("Remove")),
    Bold(Red(tool.Name + "/" + version.Version.ToString())))

  // Decrement usages of the artifact and clean-up if we are not using it anywhere else
  artifact := version.Artifact
  artifact.References -= 1
  if (artifact.References == 0) {

    // Uninstall the artifact
    err := UninstallArtifact(artifact)
    if err != nil {
      return err
    }

    // Remove from the repository
    delete(repo.Sources, artifact.ID)
  }

  // Uninstall the tool
  err := UninstallToolVersion(version)
  if err != nil {
    return err
  }

  // Remove version from the tool
  tool.RemoveVersion(version)
  return nil
}

/**
 * Install the specified tool from the repository
 */
func (repo *Repository) UninstallTool(tool *InstalledTool) error {
  err := UninstallTool(tool)
  if err != nil {
    return err
  }

  // Remove the tool from the repository
  delete(repo.Tools, tool.Name)
  return nil
}

/**
 * Calculate the size of the installed version of a tool
 */
func (ver *InstalledVersion) Size() (uint64, error) {
  toolSize, err := DirSize(ver.Folder)
  if err != nil {
    return 0, err
  }

  pkgSize, err := DirSize(ver.Artifact.Folder)
  if err != nil {
    return 0, err
  }

  return pkgSize + toolSize, nil
}
