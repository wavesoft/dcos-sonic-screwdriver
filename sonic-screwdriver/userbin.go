package main

import (
  "os"
  "fmt"
  . "github.com/logrusorgru/aurora"
)

/**
 * Create a symbolic link on the user bin directory
 */
 func CreateBinSymlink(config *ScrewdriverConfig, fullPath string, tool string) error {
  fmt.Printf("%s %s %s\n", Bold(Blue("==> ")), Bold(Gray("Link")), Bold(Blue(tool)))
  linkTarget := config.UserBinDir + "/" + tool
  if _, err := os.Stat(linkTarget); err == nil {
    os.Remove(linkTarget)
  }
  return os.Symlink(fullPath, linkTarget)
 }

/**
 * Remove a symbolic link to the user bin directory
 */
func RemoveBinSymlink(config *ScrewdriverConfig, tool string) error {
  fmt.Printf("%s %s %s\n", Bold(Blue("==> ")), Bold(Gray("Unlink")), Bold(Blue(tool)))
  linkTarget := config.UserBinDir + "/" + tool
  return os.Remove(linkTarget)
}

/**
 * Check if a symbolic link already exists in the user bin directory
 */
func HasBinSymlink(config *ScrewdriverConfig, tool string) bool {
  linkTarget := config.UserBinDir + "/" + tool
  if _, err := os.Stat(linkTarget); err == nil {
    return true
  }
  return false
}

/**
 * Get the target of the bin symlink
 */
 func ReadBinSymlink(config *ScrewdriverConfig, tool string) (string, error) {
  linkTarget := config.UserBinDir + "/" + tool
  if _, err := os.Stat(linkTarget); err != nil {
    return "", nil
  }
  return os.Readlink(linkTarget)
}
