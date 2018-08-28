package main

import (
  . "github.com/logrusorgru/aurora"
  "flag"
  "fmt"
  "os"
  "sort"
  "strings"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
)


func banner() {
  VERSION := "0.1"

  fmt.Printf("%s%s %s %s - A multi-tool for everything DC/OS\n",
    Bold(Magenta("Mesos")),
    Magenta("phere"),
    Bold(Gray("Sonic Screwdriver")),
    VERSION )
  fmt.Println("")
}

func help() {
  banner()
  fmt.Println("Typical usage:")
  fmt.Println("  ss [-v VERSION] [-dcos VERSION] add [TOOL]")
  fmt.Println("  ss rm [TOOL]")
  fmt.Println("  ss link [TOOL]")
  fmt.Println("  ss unlink [TOOL]")
  fmt.Println("")
  fmt.Println("Discovery:")
  fmt.Println("  ss topic [NAME]")
  fmt.Println("  ss ver [TOOL]")
  fmt.Println("  ss ls [REGEX]")
  fmt.Println("")
  fmt.Println("Management commands:")
  fmt.Println("  ss update")
  fmt.Println("  ss uninstall")
  fmt.Println("")
  os.Exit(2)
}

func die(msg string) {
  fmt.Printf("%s %s\n", Red("Error:"), msg)
  os.Exit(1)
}

func complete(msg string) {
  fmt.Printf("👨🏻‍🚀  %s\n", msg)
  os.Exit(0)
}

func wideTab(msg string) string {
  tabs := (32 - len(msg)) / 8
  return msg + strings.Repeat("\t", tabs)
}

func main() {
  var toolInfo registry.ToolInfo
  var ok bool

  fVersion := flag.String("v", "", "The tool version to use")
  // fDcos := flag.String("dcos", "", "The DC/OS cluster version you are targeting")
  flag.Parse()
  if flag.NArg() < 1 {
    help()
  }

  // Check actions
  switch flag.Arg(0) {

    ///
    /// Install a new tool
    ///
    case "add":
      if flag.NArg() < 2 {
        fmt.Println("Missing tool name")
        help()
      }

      // Load registry
      reg, err := registry.GetRegistry()
      if err != nil {
        die(err.Error())
      }

      // Lookup tool
      tool := flag.Arg(1)
      if toolInfo, ok = reg.Tools[tool]; !ok {
        die(fmt.Sprintf("🥔  Could not find tool '%s', here is a potato...", tool))
      }

      // Lookup version
      var version *registry.ToolVersion = nil
      if *fVersion != "" {
        version, err = toolInfo.Versions.Find(*fVersion)
        if err != nil {
          die(fmt.Sprintf("%s: %s (use `ver` to list available versions)", tool, err.Error()))
        }
      } else {
        version = toolInfo.Versions.Latest()
      }

      // Find the first artifact that can be executed on our current system
      // configuration (CPU architecture, installed interpreters or docker)
      artifact, err := registry.FindFirstRunableArtifact(version)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      // Get installed versions
      installedVersions, err := registry.GetInstalledVersions(tool)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      // Handle cases where there are already installed artifacts
      if len(installedVersions) > 0 {
        // If the version we want is already cached, just switch the link there
        for _, isntVer := range installedVersions {
          if isntVer.Version.Equals(version.Version) {
            currentEntrypoint, err := registry.ReadBinSymlink(tool)
            if err != nil || currentEntrypoint != isntVer.Entrypoint {
              registry.RemoveBinSymlink(tool)
              err = registry.CreateBinSymlink(isntVer.Entrypoint, tool)
              if err != nil {
                die(fmt.Sprintf("%s: %s", tool, err.Error()))
              }
              complete(fmt.Sprintf("switched %s to %s!", tool, version.ToString()))
            } else {
              complete(fmt.Sprintf("%s/%s is already there!", tool, version.ToString()))
            }

            return;
          }
        }
      }

      // Download the archive
      fmt.Printf("%s %s %s\n", Bold(Green("==> ")), Bold(Gray("Add")), Bold(Green(tool)))
      toolPath, err := registry.FetchArchive(tool, version, artifact)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      // Install symbolic link
      registry.RemoveBinSymlink(tool)
      err = registry.CreateBinSymlink(toolPath, tool)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      complete(fmt.Sprintf("%s/%s has landed!", tool, version.ToString()))

    ///
    /// Remove an existing tool
    ///
    case "rm", "remove":
      if flag.NArg() < 2 {
        fmt.Println("Missing tool name")
        help()
      }

      // Remove symbolic link
      tool := flag.Arg(1)
      if registry.IsToolInstalled(tool) || registry.HasBinSymlink(tool) {
        fmt.Printf("%s %s %s\n", Bold(Red("==> ")), Bold(Gray("Remove")), Bold(Red(tool)))
      } else {
        complete(fmt.Sprintf("%s is not installed", tool))
      }

      // Remove files and links
      if registry.HasBinSymlink(tool) {
        registry.RemoveBinSymlink(tool)
      }
      if registry.IsToolInstalled(tool) {
        registry.RemoveTool(tool)
      }

      complete(fmt.Sprintf("%s has left the rocket ship!", tool))

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
      fmt.Println("Available tools in the registry:")
      for _, tool := range keys {
        fmt.Printf("%s%s\n", Bold(Gray(wideTab(" "+tool))), reg.Tools[tool].Desc)
      }

    ///
    /// List the versions of a tool
    ///
    case "ver":
      if flag.NArg() < 2 {
        fmt.Println("Missing tool name")
        help()
      }

      // Load registry
      reg, err := registry.GetRegistry()
      if err != nil {
        die(err.Error())
      }

      // Lookup tool
      tool := flag.Arg(1)
      if toolInfo, ok = reg.Tools[tool]; !ok {
        die(fmt.Sprintf("🥔  Could not find tool '%s', here is a potato...", tool))
      }

      // List versions
      fmt.Printf("Avilable versions for '%s':\n", tool)
      for _, ver := range toolInfo.Versions {
        fmt.Printf("  - %s\n", ver.ToString())
      }


    default:
      fmt.Printf("Unknown action '%s'\n", flag.Arg(0))
      help()
  }
}
