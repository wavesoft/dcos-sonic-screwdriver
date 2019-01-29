package shared

import (
  "os"
  "unicode"
  "regexp"
  "strings"
)

func UcFirst(str string) string {
  for i, v := range str {
    return string(unicode.ToUpper(v)) + str[i+1:]
  }
  return ""
}

/**
 * Replaces templates with the full paths
 */
func ReplacePathTemplates(expr string, pkgDir string, toolDir string, additionalMapping map[string]string) string {
  r := regexp.MustCompile(`%\w+(:\w+)?%`)
  return r.ReplaceAllStringFunc(expr, func(m string) string {
    parts := strings.Split(m[1:len(m)-1], ":")
    key := strings.ToLower(parts[0])

    switch key {

      // "%artifact%"
      case "artifact":
        return pkgDir

      // "%tool%"
      case "tool":
        return toolDir

      // %env:VAR_NAME%
      case "env":
        if len(parts) == 1 {
          return ""
        }
        return os.Getenv(parts[1])

      // %pwd%
      case "pwd":
        str, err := os.Getwd();
        if err != nil {
          return ""
        }
        return str

      // Unknown
      default:
        if additionalMapping != nil {
          if value, ok := additionalMapping[key]; ok {
            return value
          }
        }
        return ""
    }
  })
}
