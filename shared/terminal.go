package shared

import (
  "fmt"
  "syscall"
  "golang.org/x/crypto/ssh/terminal"
)

/**
 * Read a password from the terminal
 */
func PasswordPrompt(prompt string) ([]byte, error) {

  // Read password without echoing the characters
  fmt.Print(prompt)
  bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
  if err != nil {
    return nil, err
  }

  return bytePassword, nil
}
