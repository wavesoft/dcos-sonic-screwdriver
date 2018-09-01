package shared

import (
  "fmt"
  "io"
  "os"
  "os/exec"
  "syscall"
)

/**
 * Run the given command and pipe stdout/stderr
 */
func ExecuteAndPassthrough(binary string, args ...string) (int, error) {
  cmd := exec.Command(binary, args...)
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return 0, fmt.Errorf("Unable to open StdOut Pipe: %s", err.Error())
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    return 0, fmt.Errorf("Unable to open StdErr Pipe: %s", err.Error())
  }
  if err := cmd.Start(); err != nil {
    return 0, err
  }

  // Async readers of the Stdout/Err
  go func() {
      _, _ = io.Copy(os.Stdout, stdout)
  }()
  go func() {
      _, _ = io.Copy(os.Stderr, stderr)
  }()

  if err := cmd.Wait(); err != nil {
    // Get exit code on non-zero exits
    if exiterr, ok := err.(*exec.ExitError); ok {
      if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
        return status.ExitStatus(), nil
      }
    } else {
      return 0, err
    }
  }

  return 0, nil
}

/**
 * Change directory and run the given command and pipe stdout/stderr
 */
func ExecuteInFolderAndPassthrough(workDir string, binary string, args ...string) (int, error) {
  cmd := exec.Command(binary, args...)
  cmd.Dir = workDir

  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return 0, fmt.Errorf("Unable to open StdOut Pipe: %s", err.Error())
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    return 0, fmt.Errorf("Unable to open StdErr Pipe: %s", err.Error())
  }
  if err := cmd.Start(); err != nil {
    return 0, err
  }

  // Async readers of the Stdout/Err
  go func() {
      _, _ = io.Copy(os.Stdout, stdout)
  }()
  go func() {
      _, _ = io.Copy(os.Stderr, stderr)
  }()

  if err := cmd.Wait(); err != nil {
    // Get exit code on non-zero exits
    if exiterr, ok := err.(*exec.ExitError); ok {
      if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
        return status.ExitStatus(), nil
      }
    } else {
      return 0, err
    }
  }

  return 0, nil
}


/**
 * Change directory and run the given command and pipe stdout/stderr
 */
func ShellExecuteInFolderAndPassthrough(workDir string, cmdline string) (int, error) {
  cmd := exec.Command("sh", "-c", cmdline)
  cmd.Dir = workDir

  stdout, err := cmd.StdoutPipe()
  if err != nil {
    return 0, fmt.Errorf("Unable to open StdOut Pipe: %s", err.Error())
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    return 0, fmt.Errorf("Unable to open StdErr Pipe: %s", err.Error())
  }
  if err := cmd.Start(); err != nil {
    return 0, err
  }

  // Async readers of the Stdout/Err
  go func() {
      _, _ = io.Copy(os.Stdout, stdout)
  }()
  go func() {
      _, _ = io.Copy(os.Stderr, stderr)
  }()

  if err := cmd.Wait(); err != nil {
    // Get exit code on non-zero exits
    if exiterr, ok := err.(*exec.ExitError); ok {
      if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
        return status.ExitStatus(), nil
      }
    } else {
      return 0, err
    }
  }

  return 0, nil
}

/**
 * Execute silently and return exit code
 */
func ExecuteSilently(binary string, args ...string) (int, error) {
  cmd := exec.Command(binary, args...)
  if err := cmd.Start(); err != nil {
    return 0, err
  }

  if err := cmd.Wait(); err != nil {
    // Get exit code on non-zero exits
    if exiterr, ok := err.(*exec.ExitError); ok {
      if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
        return status.ExitStatus(), nil
      }
    } else {
      return 0, err
    }
  }

  return 0, nil
}

/**
 * Execute silently on a shell terminal return exit code
 */
func ShellExecuteSilently(cmdline string) (int, error) {
  cmd := exec.Command("sh", "-c", cmdline)
  if err := cmd.Start(); err != nil {
    return 0, err
  }

  if err := cmd.Wait(); err != nil {
    // Get exit code on non-zero exits
    if exiterr, ok := err.(*exec.ExitError); ok {
      if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
        return status.ExitStatus(), nil
      }
    } else {
      return 0, err
    }
  }

  return 0, nil
}
