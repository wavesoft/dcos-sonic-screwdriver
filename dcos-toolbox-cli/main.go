package main

import (
  . "github.com/logrusorgru/aurora"
  "flag"
  "fmt"
  "os"
  "sort"
  "github.com/mesosphere/dcos-toolbox-cli/registry"
)

func help() {
  fmt.Println("Typical usage:")
  fmt.Println("  dcos-toolbox add [TOOL] [-v VERSION] [-dcos VERSION]")
  fmt.Println("  dcos-toolbox rm [TOOL]")
  fmt.Println("  dcos-toolbox link [TOOL]")
  fmt.Println("  dcos-toolbox unlink [TOOL]")
  fmt.Println("  dcos-toolbox ls [REGEX]")
  fmt.Println("")
  fmt.Println("Management commands:")
  fmt.Println("  dcos-toolbox update")
  fmt.Println("  dcos-toolbox uninstall")
  fmt.Println("")
  os.Exit(2)
}

func die(msg string) {
  fmt.Printf("%s %s\n", Red("Error:"), msg)
  os.Exit(1)
}

func main() {
  fVersion := flag.String("v", "", "The package version to use")
  fDcos := flag.String("dcos", "", "The DC/OS cluster version you are targeting")
  flag.Parse()
  if flag.NArg() < 1 {
    help()
  }

  fmt.Printf("ver=%s\n", fVersion)
  fmt.Printf("dcos=%s\n", fDcos)

  // Check actions
  switch flag.Arg(0) {

    ///
    /// Install a new tool
    ///
    case "add":
      if flag.NArg() < 2 {
        fmt.Println("Missing package name")
        help()
      }

      // Load registry
      reg, err := registry.GetRegistry()
      if err != nil {
        die(err.Error())
      }

      // Lookup package
      tool := flag.Arg(1)
      if reg.Tools[tool] == nil {
        die(fmt.Sprintf("🥔  Could not find tool '%s', here is a potato...", tool))
      }

      // Lookup version

      version := reg.Tools[tool].Latest()

      // Lookup matching artifact
      artifact, err := registry.FindFirstRunableArtifact(version)
      if err != nil {
        die(err.Error())
      }

      // Fetch the archive
      fmt.Printf("%s %s %s\n", Bold(Green("==> ")), Bold(Gray("Add")), Bold(Green(tool)))
      toolPath, err := registry.FetchArchive(tool, &version, artifact)
      if err != nil {
        die(err.Error())
      }

      // Install symbolic link
      err = registry.CreateBinSymlink(toolPath, tool)
      if err != nil {
        die(err.Error())
      }

      fmt.Printf("👨🏻‍🚀  %s/%s has landed!\n", tool, version.VersionString())

    ///
    /// Remove an existing tool
    ///
    case "rm", "remove":
      if flag.NArg() < 2 {
        fmt.Println("Missing package name")
        help()
      }

      fmt.Printf("Getting %s...\n", flag.Arg(1))

    ///
    /// List the available tools in the repository
    ///
    case "ls":

      // Load registry
      reg, err := registry.GetRegistry()
      if err != nil {
        die(err.Error())
      }

      // Sort names
      var keys []string
      for tool, _ := range reg.Tools {
        keys = append(keys, tool)
      }
      sort.Strings(keys)

      // Print
      for _, tool := range keys {
        fmt.Println(tool)
      }

    default:
      fmt.Printf("Unknown action '%s'\n", flag.Arg(0))
      help()
  }
}
