package main

import (
  "os/user"
)

type ScrewdriverConfig struct {
  UserBinDir          string
  DataDir             string
  RegistryURL         string
}

/**
 * Get the location of the registry
 */
func GetRegistryPath() (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }

  return usr.HomeDir + "/.mesosphere/toolbox", nil
}

/**
 * Get the location of the user binary path
 */
func GetUserBinPath() (string, error) {
  return "/usr/local/bin", nil
}

/**
 * Return the default configuration
 */
func GetDefaultConfig() (*ScrewdriverConfig, error) {
  regPath, err := GetRegistryPath()
  if err != nil {
    return nil, err
  }

  binPath, err := GetUserBinPath()
  if err != nil {
    return nil, err
  }

  return &ScrewdriverConfig{
    binPath,
    regPath,
    "https://raw.githubusercontent.com/wavesoft/dcos-sonic-screwdriver/master/pub/registry.json",
  }, nil
}
