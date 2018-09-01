package registry

import (
  "encoding/json"
  "fmt"
  "strconv"
  "strings"
)

type ArtifactType int

const (
    Docker     ArtifactType = iota
    Executable
)

/**
 * Interpreters for a `ExecutableToolArtifact`
 */

type ShellInterpreter struct {
  Shell           string        `json:"shell"`
}
type PythonInterpreter struct {
  Python          string        `json:"python"`
  InstReq         string        `json:"installRequirements,omitempty"`
  InstPip         string        `json:"installPip,omitempty"`
}

type ExecutableInterpreter struct {
  *PythonInterpreter
  *ShellInterpreter
}

/**
 * Sources for a `ExecutableToolArtifact`
 */

type WebFileSource struct {
  FileURL         string
  FileChecksum    string
}
type WebArchiveTarSource struct {
  TarURL          string
  TarChecksum     string
}
type VCSGitSource struct {
  GitURL          string
  GitBranch       string
}

type WebSource struct {
  *WebFileSource
  *WebArchiveTarSource
  *VCSGitSource
}

type MarshalledWebSource struct {
  Type        string                  `json:"type"`
  URL         string                  `json:"url,omitempty"`
  Checksum    string                  `json:"checksum,omitempty"`
  Branch      string                  `json:"branch,omitempty"`
}

/**
 * Types of artifact  requirements
 */

type CommandRequirement struct {
  Command     string                  `json:"cmd"`
}
type ExecRequirement struct {
  Exec        string                  `json:"exec"`
}

type ArtifactRequirement struct {
  *CommandRequirement
  *ExecRequirement
}

type ArtifactRequirements []ArtifactRequirement

/**
 * Types of a `ToolArtifact`
 */
type DockerToolArtifact struct {
  Image       string                  `json:"image"`
  Tag         string                  `json:"tag"`
  DockerArgs  string                  `json:"dockerArgs"`
}

type ExecutableToolArtifact struct {
  Source        WebSource             `json:"source"`
  Require       ArtifactRequirements  `json:"require"`
  Entrypoint    string                `json:"entrypoint"`
  Arch          string                `json:"arch"`
  Platform      string                `json:"platform"`
  Interpreter  *ExecutableInterpreter `json:"interpreter,omitempty"`
  InstallScript string                `json:"installScript"`
}

type ToolArtifact struct {
  Type        ArtifactType            `json:"type"`

  *DockerToolArtifact
  *ExecutableToolArtifact
}

/**
 * Tool details
 */

type ToolArtifacts []ToolArtifact
type VersionTriplet   [3]float64

type ToolVersion struct {
  Version     VersionTriplet             `json:"version"`
  Artifacts   ToolArtifacts           `json:"artifacts"`
}

/**
 * Various versions of the tool
 */
type ToolVersions []ToolVersion

/**
 * Help text or link
 */
type ToolHelpText struct {
  Text      string                    `json:"text"`
}
type ToolHelpURL struct {
  URL       string                    `json:"url"`
  Inline    bool                      `json:"inline"`
  Markdown  bool                      `json:"md"`
}
type ToolHelp struct {
  *ToolHelpText
  *ToolHelpURL
}

/**
 * Tool details
 */
type ToolInfo struct {
  Versions  ToolVersions              `json:"versions"`
  Help      ToolHelp                  `json:"help"`
  Desc      string                    `json:"desc"`
  Topics    []string                  `json:"topics"`
}

/**
 * The registry entry point
 */
type Registry struct {
  Tools       map[string] ToolInfo    `json:"tools"`
  Version     float64                 `json:"version"`
  ToolVersion VersionTriplet          `json:"toolVersion"`
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
        return fmt.Errorf("Invalid ArtifactType value")
    }
    *e = value
    return nil
}
func (e *ArtifactType) MarshalJSON() ([]byte, error) {
    value, ok := map[ArtifactType]string{Docker: "docker", Executable: "executable"}[*e]
    if !ok {
        return nil, fmt.Errorf("Invalid ArtifactType value")
    }
    return json.Marshal(value)
}

/**
 * Source Enum
 */
func (e *WebSource) UnmarshalJSON(data []byte) error {
  var s MarshalledWebSource
  err := json.Unmarshal(data, &s)
  if err != nil {
      return err
  }

  switch s.Type {
    case "file":
      *e = WebSource{
        WebFileSource: &WebFileSource{
          s.URL,
          s.Checksum,
        },
      }
      return nil

    case "archive/tar":
      *e = WebSource{
        WebArchiveTarSource: &WebArchiveTarSource{
          s.URL,
          s.Checksum,
        },
      }
      return nil

    case "vcs/git":
      *e = WebSource{
        VCSGitSource: &VCSGitSource{
          s.URL,
          s.Branch,
        },
      }
      return nil
  }

  return fmt.Errorf("unknown source type `%s`", s.Type)
}

func (e *WebSource) MarshalJSON() ([]byte, error) {
  var value MarshalledWebSource

  if e.WebFileSource != nil {
    value.Type = "file"
    value.URL = e.WebFileSource.FileURL
    value.Checksum = e.WebFileSource.FileChecksum

  } else if e.WebArchiveTarSource != nil {
    value.Type = "archive/tar"
    value.URL = e.WebArchiveTarSource.TarURL
    value.Checksum = e.WebArchiveTarSource.TarChecksum

  } else if e.VCSGitSource != nil {
    value.Type = "vcs/git"
    value.URL = e.VCSGitSource.GitURL
    value.Branch = e.VCSGitSource.GitBranch

  } else {
    return nil, fmt.Errorf("unexpected source type")
  }

  return json.Marshal(value)
}


/**
 * Get the version as string
 */
func (v VersionTriplet) ToString() string {
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
func VersionFromString(version string) (*VersionTriplet, error) {
  verInfo := new(VersionTriplet)
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
func (v VersionTriplet) Equals(n VersionTriplet) bool {
  return v[0] == n[0] &&
         v[1] == n[1] &&
         v[2] == n[2]
}

/**
 * Compare two versions
 */
func (v VersionTriplet) LessThan(n * VersionTriplet) bool {
  var left float64 = v[0] * 1000000 + v[1] * 1000 + v[2]
  var right float64 = n[0] * 1000000 + n[1] * 1000 + n[2]
  return left < right
}
func (v VersionTriplet) GraterThan(n * VersionTriplet) bool {
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
      return nil, fmt.Errorf("Cannot parse %d-ht component: %s", idx + 1, err.Error())
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
  return nil, fmt.Errorf("version %s not found", version)
}
