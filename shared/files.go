package shared

import (
  "path/filepath"
  "os"
)
/**
 * Calculate the size of the designated directory
 */
func DirSize(path string) (uint64, error) {
  var size uint64
  err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
    if !info.IsDir() {
      size += uint64(info.Size())
    }
    return err
  })
  return size, err
}
