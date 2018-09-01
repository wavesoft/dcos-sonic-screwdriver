package main

import (
  "os/user"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
)

type ScrewdriverConfig struct {
  UserBinDir          string
  DataDir             string
  RegistryURL         string
  RegistryPubKey      *rsa.PublicKey
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
 * Get the pre-shared public key
 */
func GetHardCodedPublicKey() *rsa.PublicKey {
  pubPEM := `
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAvjNE44H0+Y2PRrrZ7lgu
UG0jDJ65qtBwDKRahf8skGoDOrbsavfL7Vojn9wnv2ZD35bmeyyY+eszPTgiyCeU
1AlcFRlHrbWtoiEvDOuRKausNXeNYLmvZIkx8BkKGCr6U+iu2wme6Oio9GTETexq
mijmWMab5EFZydgv5VYdpRA3ADnoxGU1lIVEeTncuKaouhJxde4P/3cye5r1pxgS
V6A9F/oVGaq5DWc9cOulaLcT+l448CDvIlAqdxgqcuxh7Gh2qK1PUiWeOFuhmtJv
2UUbrrjD3Z3qJc5+eRcxPyDV3duIXNYdwQZmy9/2wNlXvV0zniyHRj+NDyi2Yx/M
vUXEuFytI4w4H/VmjrGnUriqWMhvPeFAt1wZG3G9VkDdJegUe1+cF89oWsRclf+v
CHEHfR1UvsxdLko4Kr1S6WLJaKheqBbz9i3Sw9eWF32t7rPMwYF0fMmlg9IpWNcS
h4TBkKV3yKpxRkLt+TmbpJNWERqWFvOWe/40eh4nCugM+cCwWHhg5e6l2FecJucA
PBvyayhgcRXduYiU/uscLa0Ff1g+OAuZx2aOqhRSo9obFmuYufjW9gI/K+iZGkTf
LXtsTgKvr9KCmRO9VV3+UoVyD9Q0S3b/r5BKxBDh0MfBlahHK5pHxmmUQK9RqNoV
EJjLk3BgQ1/uRRnduijwSb0CAwEAAQ==
-----END PUBLIC KEY-----`

  // NOTE: This is a critical part of the program, so it's safer to panic
  //       rather than returning the error to be handled

  block, _ := pem.Decode([]byte(pubPEM))
  if block == nil {
    panic("failed to parse PEM block containing the public key")
  }
  pub, err := x509.ParsePKIXPublicKey(block.Bytes)
  if err != nil {
    panic("failed to parse DER encoded public key: " + err.Error())
  }

  return pub.(*rsa.PublicKey)
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
    "https://raw.githubusercontent.com/wavesoft/dcos-sonic-screwdriver-registry/master/registry.json",
    GetHardCodedPublicKey(),
  }, nil
}
