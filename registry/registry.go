package registry

import (
  "encoding/json"
  "errors"
  "fmt"
  "io/ioutil"
  "os"
  "time"
)

/**
 * Get or refresh registry file
 */
func GetRegistry() (*Registry, error) {
  var info os.FileInfo
  var err error
  registryPath, err := GetRegistryPath()
  if err != nil {
    return nil, err
  }

  // Prepare package dir
  if _, err := os.Stat(registryPath); os.IsNotExist(err) {
    err = os.MkdirAll(registryPath, 0755)
    if err != nil {
      return nil, err
    }
  }

  // First try to load the file from disk, and if it failed, try web
  registryFile := fmt.Sprintf("%s/registry.json", registryPath)
  info, err = os.Stat(registryFile)
  if err != nil {
    return refreshRegistry(registryFile)
  }

  registryAge := time.Since(info.ModTime())
  if registryAge > time.Hour {
    return refreshRegistry(registryFile)
  }

  return RegistryFromDisk(registryFile)
}

/**
 * Download a fresh registry
 */
func refreshRegistry(registryFile string) (*Registry, error) {
  reg, err := RegistryFromURL(GetRegistryURL())
  if err != nil {
    return nil, errors.New("Unable to fetch registry: " + err.Error())
  }

  err = RegistryToDisk(reg, registryFile)
  if err != nil{
    return nil, errors.New("Unable to save the new registry: " + err.Error())
  }

  return reg, nil
}

/**
 * Parse a JSON buffer into a registry structure
 */
func ParseRegistry(byt []byte) (*Registry, error) {
  var reg *Registry = new(Registry)

  // Parse the JSON document
  if err := json.Unmarshal(byt, reg); err != nil {
    return nil, err
  }

  return reg, nil
}

/**
 * Load registry from the disk
 */
func RegistryFromDisk(s string) (*Registry, error) {
  byt, err := ioutil.ReadFile(s)
  if (err != nil) {
    return nil, errors.New("Error loading registry: " + err.Error())
  }

  return ParseRegistry(byt)
}

/**
 * Save registry to disk
 */
func RegistryToDisk(reg *Registry, s string) error {
    bytes, err := json.Marshal(reg)
    if err != nil {
      return errors.New("Error saving registry: " + err.Error())
    }

    return ioutil.WriteFile(s, bytes, 0644)
}

/**
 * Download the registry from URL
 */
func RegistryFromURL(s string) (*Registry, error) {
  client := RegistryHttpClient(false)
  resp, err := client.Get(s)
  if err != nil {
    return nil, errors.New("Error updating registry: " + err.Error())
  }

  // Parse contents as JSON in memory
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, errors.New("Error updating registry: " + err.Error())
  }

  return ParseRegistry(body)
}
