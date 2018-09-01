package registry

import (
  "encoding/json"
  "fmt"
  "crypto/rsa"
  "io/ioutil"
  "os"
  "time"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * Get or refresh registry file
 */
func GetRegistry(cachePath string, registryUrl string, pub *rsa.PublicKey) (*Registry, error) {
  var info os.FileInfo
  var err error

  // Prepare package dir
  if _, err := os.Stat(cachePath); os.IsNotExist(err) {
    err = os.MkdirAll(cachePath, 0755)
    if err != nil {
      return nil, err
    }
  }

  // First try to load the file from disk, and if it failed, try web
  registryFile := fmt.Sprintf("%s/registry.json", cachePath)
  info, err = os.Stat(registryFile)
  if err != nil {
    return RefreshRegistry(registryFile, registryUrl, pub)
  }

  registryAge := time.Since(info.ModTime())
  if registryAge > time.Hour {
    return RefreshRegistry(registryFile, registryUrl, pub)
  }

  return RegistryFromDisk(registryFile)
}

/**
 * Get or refresh registry file
 */
func UpdateRegistry(cachePath string, registryUrl string, pub *rsa.PublicKey) (*Registry, error) {
  // Prepare package dir
  if _, err := os.Stat(cachePath); os.IsNotExist(err) {
    err = os.MkdirAll(cachePath, 0755)
    if err != nil {
      return nil, err
    }
  }

  // First try to load the file from disk, and if it failed, try web
  registryFile := fmt.Sprintf("%s/registry.json", cachePath)
  return RefreshRegistry(registryFile, registryUrl, pub)
}

/**
 * Download a fresh registry
 */
func RefreshRegistry(registryFile string, registryUrl string, pub *rsa.PublicKey) (*Registry, error) {
  // Download the latest registry
  reg, err := RegistryFromURL(registryUrl, pub)
  if err != nil {
    return nil, fmt.Errorf("Unable to fetch registry: %s", err.Error())
  }

  // Write cached version of the registry
  err = RegistryToDisk(reg, registryFile)
  if err != nil{
    return nil, fmt.Errorf("Unable to save the new registry: %s", err.Error())
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

  // Make sure version is 1
  if reg.Version != 1 {
    return nil, fmt.Errorf("unsupported registry version %d", reg.Version)
  }

  return reg, nil
}

/**
 * Load registry from the disk
 */
func RegistryFromDisk(s string) (*Registry, error) {
  byt, err := ioutil.ReadFile(s)
  if (err != nil) {
    return nil, fmt.Errorf("Error loading registry: %s", err.Error())
  }

  return ParseRegistry(byt)
}

/**
 * Return the byte stream of the registry
 */
func RegistryToBytes(reg *Registry) ([]byte, error) {
  return json.Marshal(reg)
}

/**
 * Save registry to disk
 */
func RegistryToDisk(reg *Registry, s string) error {
  bytes, err := RegistryToBytes(reg)
  if err != nil {
    return fmt.Errorf("Error saving registry: %s", err.Error())
  }

  return ioutil.WriteFile(s, bytes, 0644)
}

/**
 * Download the registry from URL
 */
func RegistryFromURL(s string, pub *rsa.PublicKey) (*Registry, error) {

  // Download signature
  byt, err := Download(s + ".sig", WithDefaults).
              EventuallyReadAll()
  if err != nil {
    return nil, fmt.Errorf("signature error: %s", err.Error())
  }

  // Download latest version
  byt, err = Download(s, WithDefaults).
             AndDecompressIfCompressed().
             AndValidatePSSSignature(byt, pub).
             EventuallyReadAll()
  if err != nil {
    return nil, fmt.Errorf("download error: %s", err.Error())
  }

  return ParseRegistry(byt)
}
