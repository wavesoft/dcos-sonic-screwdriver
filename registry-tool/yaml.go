package main

import (
  "fmt"
  "io/ioutil"
  "strings"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "github.com/ghodss/yaml"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * Load a version from the YAML file given
 */
func LoadVersionYAML(file string) (*registry.ToolVersion, error) {

  // Read file
  byt, err := ioutil.ReadFile(file)
  if (err != nil) {
    return nil, fmt.Errorf("cannot read %s: %s", file, err.Error())
  }

  // Parse contents
  var dat registry.ToolVersion
  err = yaml.Unmarshal(byt, &dat)
  if err != nil {
    return nil, fmt.Errorf("cannot parse %s: %s", file, err.Error())
  }

  return &dat, nil
}

/**
 * Load a tool info from the given YAML given
 */
func LoadToolInfoYAML(file string) (*registry.ToolInfo, error) {

  // Read file
  byt, err := ioutil.ReadFile(file)
  if (err != nil) {
    return nil, fmt.Errorf("cannot read %s: %s", file, err.Error())
  }

  // Parse contents
  var dat registry.ToolInfo
  err = yaml.Unmarshal(byt, &dat)
  if err != nil {
    return nil, fmt.Errorf("cannot parse %s: %s", file, err.Error())
  }

  return &dat, nil
}

/**
 * Load a tool from the tool folder
 */
func LoadToolFromFolder(folder string) (*registry.ToolInfo, error) {
  var toolInfo *registry.ToolInfo = nil

  files, err := ioutil.ReadDir(folder)
  if err != nil {
    return nil, err
  }

  // First find and load package
  for _, f := range files {
    fileName := f.Name()
    if fileName == "package.yml" {
      toolInfo, err = LoadToolInfoYAML(folder + "/" + fileName)
      if err != nil {
        return nil, err
      }
      break
    }
  }
  if toolInfo == nil {
    return nil, fmt.Errorf("could not find 'package.yml' in %s", folder)
  }

  // Then load versions
  for _, f := range files {
    fileName := f.Name()
    if fileName == "package.yml" {
      continue
    }
    if !strings.HasSuffix(fileName, ".yml") {
      return nil, fmt.Errorf("unexpected file '%s' in %s", fileName, folder)
    }

    verStr := fileName[:len(fileName) - 4]
    version, err := VersionFromString(verStr)
    if err != nil {
      return nil, fmt.Errorf("unexpected file '%s' in %s: %s", fileName, folder, err.Error())
    }

    toolVersion, err := LoadVersionYAML(folder + "/" + fileName)
    if err != nil {
      return nil, err
    }

    toolVersion.Version = *version
    toolInfo.Versions = append(toolInfo.Versions, *toolVersion)
  }

  return toolInfo, err
}

/**
 * Load a registry from the tools folder
 */
func LoadRegistryFromFolder(folder string) (*registry.Registry, error) {
  var toolsRegistry registry.Registry
  toolsRegistry.Tools = make(map[string]registry.ToolInfo)

  files, err := ioutil.ReadDir(folder)
  if err != nil {
    return nil, err
  }

  // Load every tool in the folder
  for _, f := range files {
    tool, err := LoadToolFromFolder(folder + "/" + f.Name())
    if err != nil {
      return nil, err
    }

    toolsRegistry.Tools[f.Name()] = *tool
  }

  return &toolsRegistry, nil
}
