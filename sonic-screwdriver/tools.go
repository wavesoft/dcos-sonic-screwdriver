package main

import (
  "github.com/russross/blackfriday"
  "os"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * Download the help text
 */
func DownloadHelpText(s string) ([]byte, error) {
  return Download(s, WithDefaults).
         EventuallyReadAll()
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
