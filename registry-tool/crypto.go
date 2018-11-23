package main

import (
  "fmt"
  "io/ioutil"
  "crypto"
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)


func LoadPrivateKey(file string) (*rsa.PrivateKey, error) {
  var keyBytes []byte

  // Load file
  byt, err := ioutil.ReadFile(file)
  if (err != nil) {
    return nil, fmt.Errorf("unable to load private key: %s", err.Error())
  }

  // Parse PEM
  block, _ := pem.Decode(byt)
  if block == nil {
    return nil, fmt.Errorf("failed to parse PEM block of private key")
  }

  // Read password
  if x509.IsEncryptedPEMBlock(block) {
    passwd, err := PasswordPrompt("Key Password: ")
    if err != nil {
      return nil, fmt.Errorf("unable to read private key password: %s", err.Error())
    }

    // Decrypt PEM block
    keyBytes, err = x509.DecryptPEMBlock(block, passwd)
    if err != nil {
      return nil, fmt.Errorf("unable to decrypt private key: %s", err.Error())
    }
  } else {
    keyBytes = block.Bytes
  }

  // Parse private key
  key, err := x509.ParsePKCS1PrivateKey(keyBytes)
  if err != nil {
    panic("failed to parse DER encoded private key: " + err.Error())
  }

  // Key is ready
  return key, nil
}

func SignPayload(byt []byte, key *rsa.PrivateKey) ([]byte, error) {
  var opts rsa.PSSOptions
  opts.SaltLength = rsa.PSSSaltLengthAuto

  hashAlgo := crypto.SHA256

  pssh := hashAlgo.New()
  pssh.Write(byt)
  hashed := pssh.Sum(nil)

  signature, err := rsa.SignPSS(rand.Reader, key, hashAlgo, hashed, &opts)
  if err != nil {
    return nil, err
  }

  return signature, nil
}
