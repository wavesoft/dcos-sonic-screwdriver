package repository

import (
  "encoding/json"
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "io/ioutil"
  "os"
)

/**
 * Read artifact details from the specified package
 */
func (pk InstalledArtifact) GetRegistryArtifact() (*registry.ToolArtifact, error) {
  stateFile := pk.Folder + "/.state"
  if _, err := os.Stat(stateFile); err != nil {
    return nil, fmt.Errorf("Could not find state file: %s", stateFile)
  }

  return ReadArtifactStateFile(stateFile)
}

/**
 * Write artifact details on the specified package
 */
func (pk InstalledArtifact) SetRegistryArtifact(artifact *registry.ToolArtifact) error {
  stateFile := pk.Folder + "/.state"
  return WriteArtifactStateFile(stateFile, artifact)
}

/**
 * Read the artifact metadata from the ready flag
 */
func ReadArtifactStateFile(path string) (*registry.ToolArtifact, error) {
  byt, err := ioutil.ReadFile(path)
  if (err != nil) {
    return nil, err
  }

  // Parse the JSON document
  artifact := new(registry.ToolArtifact)
  if err := json.Unmarshal(byt, artifact); err != nil {
    return nil, err
  }

  return artifact, nil
}

/**
 * Create a flag with the tool artifact details
 */
func WriteArtifactStateFile(path string, artifact *registry.ToolArtifact) error {
  bytes, err := json.Marshal(artifact)
  if err != nil {
    return err
  }

  return ioutil.WriteFile(path, bytes, 0644)
}
