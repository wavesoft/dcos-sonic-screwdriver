package repository

import (
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "crypto/sha256"
  "encoding/hex"
  "fmt"
)

/**
 * An installed source package in the repository
 */
type InstalledArtifact struct {
  ID          string
  Folder      string
  References  int
}

/**
 * A version of an installed tool in the repository
 */
type InstalledVersion struct {
  Version     registry.VersionTriplet
  Artifact    *InstalledArtifact
  Folder      string
}

/**
 * A tool in the repository
 */
type InstalledTool struct {
  Versions    []InstalledVersion
  Name        string
  Folder      string
}

/**
 * The local repository
 */
type Repository struct {
  Sources     map[string]*InstalledArtifact
  Tools       map[string]*InstalledTool
  BaseDir     string
}

/**
 * Calculate a unique ID based on the given artifact
 */
func ArtifactID(a *registry.ToolArtifact) string {
  if (a.ExecutableToolArtifact != nil) {
    s := a.Source
    if (s.VCSGitSource != nil) {
      sum := sha256.Sum256([]byte(fmt.Sprintf("git:%s:%s", s.GitURL, s.GitBranch)))
      return hex.EncodeToString(sum[:])
    }
    if (s.WebFileSource != nil) {
      sum := sha256.Sum256([]byte(fmt.Sprintf("file:%s:%s", s.FileURL, s.FileChecksum)))
      return hex.EncodeToString(sum[:])
    }
    if (s.WebArchiveTarSource != nil) {
      sum := sha256.Sum256([]byte(fmt.Sprintf("tar:%s:%s", s.TarURL, s.TarChecksum)))
      return hex.EncodeToString(sum[:])
    }
  } else if (a.DockerToolArtifact != nil) {
    sum := sha256.Sum256([]byte(fmt.Sprintf("docker:%s:%s:%s", a.Image, a.Tag, a.DockerArgs)))
    return hex.EncodeToString(sum[:])
  }

  return ""
}

/**
 * Sanitize the tool name in order to resolve it to the folder name
 */
func SanitizedToolName(name string) string {
  return name
}
