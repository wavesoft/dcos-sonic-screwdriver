package main

import (
  "flag"
  "fmt"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "github.com/pkg/browser"
  "os"
  "sort"
  "strings"
  . "github.com/logrusorgru/aurora"
)

/**
 * The current (hard-coded) version
 */
func GetVersion() registry.VersionInfo {
  return registry.VersionInfo{0,1,1}
}

/**
 * Display banner
 */
func banner() {
  fmt.Printf("%s%s %s %s - A multi-tool for everything DC/OS\n",
    Bold(Magenta("Mesos")),
    Magenta("phere"),
    Bold(Gray("Sonic Screwdriver")),
    GetVersion().ToString() )
  fmt.Println("")
}

/**
 * Display help message
 */
func help() {
  banner()
  fmt.Println("Typical usage:")
  fmt.Println("  ss [-v VERSION] add [TOOL]")
  fmt.Println("  ss rm [TOOL]")
  fmt.Println("  ss link [TOOL]")
  fmt.Println("  ss unlink [TOOL]")
  fmt.Println("")
  fmt.Println("Discovery:")
  fmt.Println("  ss ls [TOPIC | NAME | REGEX]")
  fmt.Println("  ss help [TOOL]")
  fmt.Println("  ss info [TOOL]")
  fmt.Println("")
  fmt.Println("Management commands:")
  fmt.Println("  ss update")
  fmt.Println("  ss upgrade")
  fmt.Println("  ss version")
  fmt.Println("")
  os.Exit(2)
}

/**
 * Exit with error
 */
func die(msg string) {
  fmt.Printf("%s %s\n", Red("Error:"), msg)
  os.Exit(1)
}

/**
 * Exit with a success message
 */
func complete(msg string) {
  fmt.Printf("👨🏻‍🚀  %s\n", msg)
  os.Exit(0)
}

/**
 * Pad message with the remaining tabs until we reach 32 characters-wide
 */
func wideTab(msg string) string {
  tabs := (32 - len(msg)) / 8
  return msg + strings.Repeat("\t", tabs)
}

/**
 * Search the given version into the list of installed versions and return
 * the matched installed version
 */
func findInstalledVersion(installedVersions registry.InstalledVersions,
    searchVer registry.VersionInfo) *registry.InstalledVersion {
  if len(installedVersions) == 0 {
    return nil
  }

  // If the version we want is already cached, just switch the link there
  for _, isntVer := range installedVersions {
    if isntVer.Version.Equals(searchVer) {
      return &isntVer
    }
  }

  return nil
}

/**
 * Entry point
 */
func main() {
  var toolInfo registry.ToolInfo
  var ok bool

  fVersion := flag.String("v", "", "The tool version to use")
  fForce := flag.Bool("f", false, "Force overwriting symlinks not created by us")
  flag.Parse()
  if flag.NArg() < 1 {
    help()
  }

  // Check actions
  switch flag.Arg(0) {

    ///
    /// Complete upgrade (Called by the upgrade tool)
    ///
    case "complete-upgrade":
      CompleteUpgrade(flag.Arg(1))

    ///
    /// Install a new tool
    ///
    case "a", "add", "install":
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

      installedVersion := findInstalledVersion(installedVersions, version.Version)
      if installedVersion != nil {
        currentEntrypoint, err := registry.ReadBinSymlink(tool)
        if err != nil || currentEntrypoint != installedVersion.Entrypoint {
          registry.RemoveBinSymlink(tool)
          err = registry.CreateBinSymlink(installedVersion.Entrypoint, tool)
          if err != nil {
            die(fmt.Sprintf("%s: %s", tool, err.Error()))
          }
          complete(fmt.Sprintf("switched %s to %s!", tool, version.ToString()))
        } else {
          complete(fmt.Sprintf("%s/%s is already there!", tool, version.ToString()))
        }

        return;
      }

      // If we have nothing installed, but there is a symlink, bail because
      // this could overwrite something that we didn't put there
      if len(installedVersions) == 0 {
        if registry.HasBinSymlink(tool) && !*fForce {
          die("There is already a tool with the same name in your path. Not installing.")
        }
      }

      // Download the archive
      fmt.Printf("%s %s %s\n", Bold(Green("==> ")), Bold(Gray("Add")), Bold(Green(tool)))
      toolPath, err := registry.FetchArchive(tool, &version.Version, artifact)
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
    case "r", "rm", "remove":
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

        // Remove all artifacts individually
        installedVersions, err := registry.GetInstalledVersions(tool)
        if err != nil {
          die(fmt.Sprintf("%s: %s", tool, err.Error()))
        }
        if len(installedVersions) > 0 {
          for _, isntVer := range installedVersions {
            registry.RemoveArtifact(tool, &isntVer.Version, &isntVer.Artifact)
          }
        }

        // Cleanup any residue
        registry.RemoveTool(tool)
      }

      complete(fmt.Sprintf("%s has left the rocket ship!", tool))

    ///
    /// Unlink without removing
    ///
    case "u", "unlink":
      if flag.NArg() < 2 {
        fmt.Println("Missing tool name")
        help()
      }

      // Remove symbolic link
      tool := flag.Arg(1)
      if registry.HasBinSymlink(tool) {
        fmt.Printf("%s %s %s\n", Bold(Red("==> ")), Bold(Gray("Unlink")), Bold(Red(tool)))
        registry.RemoveBinSymlink(tool)
      } else {
        complete(fmt.Sprintf("%s is not linked (or installed)", tool))
      }

    ///
    /// List the available tools in the repository
    ///
    case "l", "ls", "list":

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
    case "i", "info":
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

      // Get installed versions
      installedVersions, err := registry.GetInstalledVersions(tool)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      // List versions
      fmt.Printf("Available versions for '%s':\n", tool)
      for _, ver := range toolInfo.Versions {

        suffix := ""
        if findInstalledVersion(installedVersions, ver.Version) != nil {
          suffix = " (installed)"
        }

        fmt.Printf("  * %s %s\n", Bold(Gray(ver.ToString())), suffix)

        for _, artifact := range ver.Artifacts {
          if artifact.DockerToolArtifact != nil {
            fmt.Printf("    - platform    : docker\n")
            fmt.Printf("      image       : %s:%s\n", artifact.Image, artifact.Tag)
          }
          if artifact.ExecutableToolArtifact != nil {
            if artifact.Interpreter != "" {
              fmt.Printf("    - platform    : interpreter\n")
              fmt.Printf("      interpreter : %s\n", artifact.Interpreter)
            } else {
              fmt.Printf("    - platform    : %s\n", artifact.Platform)
              fmt.Printf("      CPU arch    : %s\n", artifact.Arch)
            }
          }
        }
      }

    ///
    /// Get help for a tool
    ///
    case "h", "help":
      if flag.NArg() < 2 {
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

      // Handle help
      if toolInfo.Help.ToolHelpText != nil {
        fmt.Printf("--=[ %s ]=--\n", Bold(Gray(tool)))
        fmt.Println("")
        fmt.Println(toolInfo.Help.Text)
        return
      } else if toolInfo.Help.ToolHelpURL != nil {
        if toolInfo.Help.Inline {
          fmt.Printf("--=[ %s ]=--\n", Bold(Gray(tool)))
          contents, err := registry.DownloadHelpText(toolInfo.Help.URL)
          if err != nil {
            die(err.Error())
          }

          if toolInfo.Help.Markdown {
            registry.PrintMarkdownText(contents)
          } else {
            fmt.Printf("%s\n", contents)
          }
        } else {
          browser.OpenURL(toolInfo.Help.URL)
        }
        return
      } else {
        fmt.Println("No help available for this tool")
      }

    ///
    /// Update the tool
    ///
    case "upgrade":
      lastVersion, err := GetLatestVersion()
      if err != nil {
        die(err.Error())
      }

      myVersion := GetVersion()

      // Check if there is a new version to upgrade
      if lastVersion.Version.GraterThan(&myVersion) {
        fmt.Printf("%s %s from %s -> to %s\n",
          Magenta("==>"),
          Bold(Gray("Upgrading")),
          myVersion.ToString(),
          lastVersion.Version.ToString())
        PerformUpgrade(lastVersion)
      } else {
        complete("You already run the latest version")
      }

    ///
    /// Update the database
    ///
    case "update":
      fmt.Printf("Updating registry...\n")
      _, err := registry.UpdateRegistry()
      if err != nil {
        die(err.Error())
      }
      complete("Registry is updated")

    ///
    /// Show the version
    ///
    case "v", "version":
      fmt.Printf(GetVersion().ToString())

    default:
      fmt.Printf("Unknown action '%s'\n", flag.Arg(0))
      help()
  }
}
