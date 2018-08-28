package registry

// hat tip https://gist.github.com/indraniel/1a91458984179ab4cf80

import (
  "archive/tar"
  "bufio"
  "compress/gzip"
  "errors"
  "fmt"
  "github.com/russross/blackfriday"
  "io"
  "io/ioutil"
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

/**
 * Extract the file stream to the designated file object
 */
func DownloadFileGz(gzipStream io.Reader, fileStream io.Writer) error {
  bReader := bufio.NewReader(gzipStream)

  testBytes, err := bReader.Peek(2)
  if err != nil {
    return errors.New(
      fmt.Sprintf("DownloadFileGz: could not detect stream type: %s", err.Error()))
  }

  // We have a GZip stream
  if testBytes[0] == 31 && testBytes[1] == 139 {
    uncompressedStream, err := gzip.NewReader(bReader)
    if err != nil {
      return errors.New(
        fmt.Sprintf("DownloadFileGz: could not open GZip stream: %s", err.Error()))
    }
    if _, err := io.Copy(fileStream, uncompressedStream); err != nil {
      return errors.New(
        fmt.Sprintf("DownloadFileGz: Copy() failed: %s", err.Error()))
    }

  // Otherwise we have plaintext stream
  } else {
    if _, err := io.Copy(fileStream, bReader); err != nil {
      return errors.New(
        fmt.Sprintf("DownloadFileGz: Copy() failed: %s", err.Error()))
    }
  }

  return nil
}

/**
 * Download the help text
 */
func DownloadHelpText(s string) ([]byte, error) {
  client := RegistryHttpClient(false)
  resp, err := client.Get(s)
  if err != nil {
    return []byte{}, errors.New("Error updating registry: " + err.Error())
  }

  // Parse contents as JSON in memory
  defer resp.Body.Close()
  return ioutil.ReadAll(resp.Body)
}

/**
 * Prints a markdown text to the console
 */
func PrintMarkdownText(input []byte) {
  renderer := &Console{}
  extensions := 0 |
    blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
    blackfriday.EXTENSION_FENCED_CODE |
    blackfriday.EXTENSION_AUTOLINK |
    blackfriday.EXTENSION_STRIKETHROUGH |
    blackfriday.EXTENSION_SPACE_HEADERS |
    blackfriday.EXTENSION_HEADER_IDS |
    blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
    blackfriday.EXTENSION_DEFINITION_LISTS

  output := blackfriday.Markdown(input, renderer, extensions)
  os.Stdout.Write(output)
}
