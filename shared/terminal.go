package shared

import (
  "fmt"
  "bufio"
  "syscall"
  "os"
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

  fmt.Println()
  return bytePassword, nil
}

/**
 * Read an arbitrary input from the user
 */
func InputPrompt(prompt string) string {
  reader := bufio.NewReader(os.Stdin)

  fmt.Print("Enter Username: ")
  input, _ := reader.ReadString('\n')
  return input
}
