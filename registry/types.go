package registry

import (
  "encoding/json"
  "fmt"
  "strconv"
  "strings"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
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
type JavaInterpreter struct {
  Java            string        `json:"java"`
  JavaArgs        string        `json:"javaArgs,omitempty"`
}

type ExecutableInterpreter struct {
  *JavaInterpreter
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
  Tag         string                  `json:"tag,omitempty"`
  DockerArgs  string                  `json:"dockerArgs,omitempty"`
}

type ExecutableToolArtifact struct {
  Source          WebSource             `json:"source"`
  Require         ArtifactRequirements  `json:"require,omitempty"`
  Entrypoint      string                `json:"entrypoint,omitempty"`
  Arch            string                `json:"arch,omitempty"`
  Platform        string                `json:"platform,omitempty"`
  Interpreter     *ExecutableInterpreter `json:"interpreter,omitempty"`
  InstallScript   string                `json:"installScript,omitempty"`
  UninstallScript string                `json:"uninstallScript,omitempty"`
  Workdir         string                `json:"workdir",omitempty`
  Environment     map[string]string     `json:"env",omitempty`
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

type ToolVersion struct {
  Version     VersionTriplet          `json:"version"`
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
  Inline    bool                      `json:"inline,omitempty"`
  Markdown  bool                      `json:"md,omitempty"`
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
func (v ToolVersion) ToString() string {
  return v.Version.ToString()
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
