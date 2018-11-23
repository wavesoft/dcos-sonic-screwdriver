package main

import (
  "flag"
  "fmt"
  "io/ioutil"
  "crypto/rsa"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "os"
  . "github.com/logrusorgru/aurora"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * The current (hard-coded) version
 */
func GetVersion() VersionTriplet {
  return VersionTriplet{0,1,0}
}

/**
 * Display banner
 */
func banner() {
  fmt.Printf("%s%s %s %s - A multi-tool for everything DC/OS\n",
    Bold(Magenta("Mesos")),
    Magenta("phere"),
    Bold(Gray("Sonic Screwdriver Registry Tool")),
    GetVersion().ToString() )
  fmt.Println("")
}

/**
 * Display help message
 */
func help() {
  banner()
  fmt.Println("Typical usage:")
  fmt.Println("  registry-tool -f [regsitry.json] -k [private.pem] sign")
  fmt.Println("  registry-tool -f [regsitry.json] -k [private.pem] -d [tools/dir] update")
  fmt.Println("")
  os.Exit(2)
}

/**
 * Exit with error
 */
func die(msg string) {
  fmt.Printf("%s %s\n", Red("Error:"), msg)
  os.Exit(1)
}

/**
 * Exit with a success message
 */
func complete(msg string) {
  fmt.Printf("👨🏻‍🚀  %s\n", msg)
  os.Exit(0)
}

/**
 * Get the registry path ether from environment or from the argument
 */
func getRegistryPath(fRegistry string) string {
  var regPath string = "registry.json"

  envKey := os.Getenv("SS_REGISTRY_FILE")
  if envKey != "" {
    regPath = envKey
  }

  if fRegistry != "" {
    regPath = fRegistry
  }

  return regPath
}

/**
 * Get private key either from environment or from the arguments and load it
 */
func loadPrivateKey(fKey string) *rsa.PrivateKey {
  var keyPath string = "private.pem"

  envKey := os.Getenv("SS_REGISTRY_PRIVATE_KEY")
  if envKey != "" {
    keyPath = envKey
  }

  if fKey != "" {
    keyPath = fKey
  }

  key, err := LoadPrivateKey(keyPath)
  if err != nil {
    die(err.Error())
  }

  return key
}

/**
 * Save and sign registry
 */
func saveAndSign(reg *registry.Registry, registryPath string, fKey string) {
  // Load private key
  key := loadPrivateKey(fKey)

  // Stringify registry
  registryContents, err := registry.RegistryToBytes(reg)
  if err != nil {
    die(err.Error())
  }

  // Calculate signature
  signature, err := SignPayload(registryContents, key)
  if err != nil {
    die(err.Error())
  }

  // Save signature
  err = ioutil.WriteFile(registryPath + ".sig", signature, 0644)
  if err != nil {
    die(err.Error())
  }

  // Backup current registry file
  if _, err := os.Stat(registryPath); err == nil {
    if _, err := os.Stat(registryPath + ".bak"); err == nil {
      err = os.Remove(registryPath + ".bak")
      if err != nil {
        die(err.Error())
      }
    }
    err = os.Rename(registryPath, registryPath + ".bak")
    if err != nil {
      die(err.Error())
    }
  }

  // Replace registry file
  err = ioutil.WriteFile(registryPath, registryContents, 0644)
  if err != nil {
    die(err.Error())
  }
}

/**
 * Entry point
 */
func main() {
  fRegistry := flag.String("f", "registry.json", "Path to the registry file")
  fRegistryFolder := flag.String("d", "registry", "Path to the tools directory")
  fKey := flag.String("k", "private.pem", "Path to the private key file")
  flag.Parse()
  if flag.NArg() < 1 {
    help()
  }


  // Check actions
  switch flag.Arg(0) {

    //
    // Update the registry by reading the tools folder
    //
    case "u", "update":

      // Load tools registry
      reg, err := LoadRegistryFromFolder(*fRegistryFolder)
      if err != nil {
        die(err.Error())
      }

      // Configure settings
      reg.Version = 1
      reg.ToolVersion = VersionTriplet{0,1,3}

      // Save and sign
      registryPath := getRegistryPath(*fRegistry)
      saveAndSign(reg, registryPath, *fKey)

      return

    //
    // Sign the existing registry
    //
    case "s", "sign":

      // Load JSON registry
      registryPath := getRegistryPath(*fRegistry)
      reg, err := registry.RegistryFromDisk(registryPath)
      if err != nil {
        die(err.Error())
      }

      // Save and sign
      saveAndSign(reg, registryPath, *fKey)
      complete("Signature saved on " + registryPath + ".sig")

    ///
    /// Show the version
    ///
    case "v", "version":
      fmt.Println(GetVersion().ToString())

    default:
      fmt.Printf("Unknown action '%s'\n", flag.Arg(0))
      help()
  }
}
