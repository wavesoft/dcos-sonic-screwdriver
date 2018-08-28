package registry

import (
  "encoding/json"
  "errors"
  "fmt"
  "strconv"
  "strings"
)

type ArtifactType int

const (
    Docker     ArtifactType = iota
    Executable
)

type DockerToolArtifact struct {
  Image       string                  `json:"image"`
  Tag         string                  `json:"tag"`
}

type HttpSource struct {
  URL         string                  `json:"url"`
  Checksum    string                  `json:"checksum"`
}

type GitSource struct {
  GitURL      string                  `json:"gitUrl"`
}

type Source struct {
  *HttpSource
  *GitSource
}

type ExecutableToolArtifact struct {
  Source      Source                  `json:"source"`
  Entrypoint  string                  `json:"entrypoint"`
  Arch        string                  `json:"arch"`
  Platform    string                  `json:"platform"`
  Interpreter string                  `json:"interpreter"`
}

type ToolArtifact struct {
  Type        ArtifactType            `json:"type"`
  *ExecutableToolArtifact
  *DockerToolArtifact
}

type ToolArtifacts []ToolArtifact
type VersionInfo   [3]float64

type ToolVersion struct {
  Version     VersionInfo             `json:"version"`
  Artifacts   ToolArtifacts           `json:"artifacts"`
}

/**
 * Various versions of the tool
 */
type ToolVersions []ToolVersion

/**
 * Tool details
 */
type ToolInfo struct {
  Versions  ToolVersions              `json:"versions"`
  Desc      string                    `json:"desc"`
  Topics    []string                  `json:"topics"`
}

/**
 * The registry entry point
 */
type Registry struct {
  Tools     map[string] ToolInfo      `json:"tools"`
  Version   float64                   `json:"version"`
}

/**
 * Type Enum Marshalling
 */
func (e *ArtifactType) UnmarshalJSON(data []byte) error {
    var s string
    err := json.Unmarshal(data, &s)
    if err != nil {
        return err
    }

    value, ok := map[string]ArtifactType{"docker": Docker, "executable": Executable}[s]
    if !ok {
        return errors.New("Invalid ArtifactType value")
    }
    *e = value
    return nil
}
func (e *ArtifactType) MarshalJSON() ([]byte, error) {
    value, ok := map[ArtifactType]string{Docker: "docker", Executable: "executable"}[*e]
    if !ok {
        return nil, errors.New("Invalid ArtifactType value")
    }
    return json.Marshal(value)
}

/**
 * Get the version as string
 */
func (v VersionInfo) ToString() string {
  return fmt.Sprintf("%d.%d.%d",
    uint32(v[0]),
    uint32(v[1]),
    uint32(v[2]))
}

/**
 * Get the version as string
 */
func (v ToolVersion) ToString() string {
  return v.Version.ToString()
}

/**
 * Parse a version string to a version info structure
 */
func VersionFromString(version string) (*VersionInfo, error) {
  verInfo := new(VersionInfo)
  verFrag := strings.SplitN(version, ".", 3)

  for idx, fragStr := range verFrag {
    fragInt, err := strconv.Atoi(fragStr)
    if err != nil {
      return nil, err
    }
    verInfo[idx] = float64(fragInt)
  }

  return verInfo, nil
}

/**
 * Get the latest version of a tool
 */
func (v ToolVersions) Latest() *ToolVersion {
  var biggest float64 = v[0].Version[0] * 1000000 + v[0].Version[1] * 1000 + v[0].Version[2]
  var found = v[0]

  for _, ver := range v {
    versionNum := ver.Version[0] * 1000000 + ver.Version[1] * 1000 + ver.Version[2]
    if versionNum > biggest {
      found = ver
    }
  }

  return &found
}

/**
 * Compare two versions
 */
func (v VersionInfo) Equals(n VersionInfo) bool {
  return v[0] == n[0] &&
         v[1] == n[1] &&
         v[2] == n[2]
}

/**
 * Compare two versions
 */
func (v VersionInfo) LessThan(n * VersionInfo) bool {
  var left float64 = v[0] * 1000000 + v[1] * 1000 + v[2]
  var right float64 = n[0] * 1000000 + n[1] * 1000 + n[2]
  return left < right
}
func (v VersionInfo) GreaterThan(n * VersionInfo) bool {
  var left float64 = v[0] * 1000000 + v[1] * 1000 + v[2]
  var right float64 = n[0] * 1000000 + n[1] * 1000 + n[2]
  return left > right
}

/**
 * Find a particular version
 */
func (v ToolVersions) Find(version string) (*ToolVersion, error) {
  verFragInt := []int{}
  verFrag := strings.Split(version, ".")

  for idx, fragStr := range verFrag {
    fragInt, err := strconv.Atoi(fragStr)
    if err != nil {
      return nil, errors.New(fmt.Sprintf(
        "Cannot parse %d-ht component: %s", idx + 1, err.Error()))
    }
    verFragInt = append(verFragInt, fragInt)
  }

  for _, ver := range v {
    found := true

    // Match version components
    for idx, fragInt := range verFragInt {
      if int(ver.Version[idx]) != fragInt {
        found = false
        break
      }
    }

    // If found, return it
    if found {
      return &ver, nil
    }
  }

  // Not found
  return nil, errors.New(fmt.Sprintf("version %s not found", version))
}
