package registry

// hat tip https://gist.github.com/indraniel/1a91458984179ab4cf80

import (
  "archive/tar"
  "compress/gzip"
  "errors"
  "fmt"
  "io"
  "os"
  "strings"
)

/**
 * Extract the given tar+gz stream to the given prefix
 *
 * Note that this function strips the first path component from the archive
 */
func ExtractTarGz(gzipStream io.Reader, prefix string) error {
  uncompressedStream, err := gzip.NewReader(gzipStream)
  if err != nil {
    return errors.New(
      fmt.Sprintf("ExtractTarGz: NewReader failed: %s", err.Error()))
  }

  tarReader := tar.NewReader(uncompressedStream)

  for true {
    header, err := tarReader.Next()

    if err == io.EOF {
      break
    }

    if err != nil {
      return errors.New(
        fmt.Sprintf("ExtractTarGz: Next() failed: %s", err.Error()))
    }

    switch header.Typeflag {
    case tar.TypeDir:
      fName := strings.SplitN(header.Name, "/", 2)
      if len(fName) > 1 && fName[1] != "" {
        if err := os.Mkdir(prefix + fName[1], 0755); err != nil {
          return errors.New(
            fmt.Sprintf("ExtractTarGz: Mkdir() failed: %s", err.Error()))
        }
      }
    case tar.TypeReg:
      fName := strings.SplitN(header.Name, "/", 2)
      if len(fName) > 1 && fName[1] != "" {
        outFile, err := os.Create(prefix + fName[1])
        if err != nil {
          return errors.New(
            fmt.Sprintf("ExtractTarGz: Create() failed: %s", err.Error()))
        }
        defer outFile.Close()
        if _, err := io.Copy(outFile, tarReader); err != nil {
          return errors.New(
            fmt.Sprintf("ExtractTarGz: Copy() failed: %s", err.Error()))
        }
      }
    default:
      // return errors.New(
      //   fmt.Sprintf("ExtractTarGz: uknown type: %s in %s",
      //     header.Typeflag,
      //     header.Name))
    }
  }

  return nil
}
