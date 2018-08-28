package registry

import (
  "net/http"
  "os/user"
  "time"
)

/**
 * A customized HTTP client
 */
func RegistryHttpClient(disableCompression bool) *http.Client {
  tr := &http.Transport{
    MaxIdleConns:       10,
    IdleConnTimeout:    30 * time.Second,
    DisableCompression: disableCompression,
  }
  return &http.Client{Transport: tr}
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
 * Get the location of the on-line registry
 */
func GetRegistryURL() string {
  return "https://raw.githubusercontent.com/wavesoft/dcos-sonic-screwdriver/master/pub/registry.json"
}

/**
 * Get the path where to put the symbolic links
 */
func GetSymlinkBinPath() string {
  return "/usr/local/bin"
}

/**
 * Checks if the interpreter name is legit
 */
func IsValidInterpreter(name string) bool {
  switch name {
    case "python", "python2", "python3", "ruby", "perl", "bash", "java":
      return true
    default:
      return false
  }
}

/**
 * The default entrypoint on archives
 */
func GetDefaultEntrypoint() string {
  return "/run-tool"
}
