package registry

import (
  "encoding/json"
  "errors"
  "fmt"
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

type ToolVersion struct {
  Version     [3]float64              `json:"version"`
  Artifacts   ToolArtifacts           `json:"artifacts"`
}

/**
 * Various versions of the tool
 */
type ToolVersions []ToolVersion

/**
 * The registry entry point
 */
type Registry struct {
  Tools     map[string] ToolVersions  `json:"tools"`
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
func (v ToolVersion) VersionString() string {
  return fmt.Sprintf("%d.%d.%d",
    uint32(v.Version[0]),
    uint32(v.Version[1]),
    uint32(v.Version[2]))
}

/**
 * Get the latest version of a tool
 */
func (v ToolVersions) Latest() ToolVersion {
  var biggest float64 = v[0].Version[0] * 1000000 + v[0].Version[1] * 1000 + v[0].Version[2]
  var found = v[0]

  for _, ver := range v {
    versionNum := ver.Version[0] * 1000000 + ver.Version[1] * 1000 + ver.Version[2]
    if versionNum > biggest {
      found = ver
    }
  }

  return found
}
