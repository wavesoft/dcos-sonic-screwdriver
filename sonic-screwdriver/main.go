package main

import (
  "errors"
  "flag"
  "fmt"
  "github.com/briandowns/spinner"
  "github.com/dustin/go-humanize"
  "github.com/mesosphere/dcos-sonic-screwdriver/registry"
  "github.com/mesosphere/dcos-sonic-screwdriver/repository"
  "github.com/pkg/browser"
  "os"
  "sort"
  "strings"
  "time"
  . "github.com/logrusorgru/aurora"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

var VERSION VersionTriplet = VersionTriplet{0,1,2}
var AlreadyUpgraded = errors.New("You already run the latest version")

/**
 */

/**
 * Display banner
 */
func banner() {
  fmt.Printf("%s%s %s %s - A multi-tool for everything DC/OS\n",
    Bold(Magenta("Mesos")),
    Magenta("phere"),
    Bold(Gray("Sonic Screwdriver")),
    VERSION.ToString() )
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
  tabs := (40 - len(msg)) / 8
  return msg + strings.Repeat("\t", tabs)
}

/**
 * Load the registry  pair and exit on errors
 */
func getRegistry(config *ScrewdriverConfig) *registry.Registry {
  // Load registry (could be slower)
  spinner := spinner.New(spinner.CharSets[13], 100*time.Millisecond)
  spinner.Start()
  reg, err := registry.GetRegistry(
    config.DataDir,
    config.RegistryURL,
    config.RegistryPubKey)
  if err != nil {
    spinner.Stop()
    die(err.Error())
  }
  spinner.Stop()
  checkMinToolVersion(reg)

  return reg
}

/**
 * Load the registry + repository pair and exit on errors
 */
func getRegistryRepository(config *ScrewdriverConfig) (*registry.Registry, *repository.Repository) {
  // Load repository (should be fast)
  repo, err := repository.LoadRepository(config.DataDir)
  if err != nil {
    die(err.Error())
  }

  // Load registry (could be slower)
  reg := getRegistry(config)

  // Return tuple
  return reg, repo
}

/**
 * Perform a fully automated tool upgrade
 */
func upgradeTool() error {
  spinner := spinner.New(spinner.CharSets[13], 100*time.Millisecond)
  spinner.Start()
  lastVersion, err := GetLatestVersion()
  if err != nil {
    spinner.Stop()
    return err
  }
  spinner.Stop()

  // Check if there is a new version to upgrade
  if lastVersion.Version.GraterThan(&VERSION) {
    fmt.Printf("%s %s from %s -> to %s\n",
      Magenta("==>"),
      Bold(Gray("Upgrading")),
      VERSION.ToString(),
      lastVersion.Version.ToString())

    // Perform upgrade and check for errors
    err := PerformUpgrade(lastVersion)
    if err != nil {
      return err
    } else {
      complete("Upgraded to version " + lastVersion.Version.ToString())
      return nil
    }
  }

  return AlreadyUpgraded
}

/**
 * Check if the registry is targeting a newer tool version
 */
func checkMinToolVersion(reg *registry.Registry) {
  if reg.ToolVersion.GraterThan(&VERSION) {
    die("👴🏻  Your tool is outdated, try `ss upgrade` to get the latest version.")
  }
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

  config, err := GetDefaultConfig()
  if err != nil {
    die(err.Error())
  }

  // Check actions
  switch flag.Arg(0) {

    case "test":
      repository, err := repository.LoadRepository(config.DataDir)
      if err != nil {
        die(err.Error())
      }
      fmt.Printf("repo=%s\n", repository)

      artifact, err := repository.Tools["marathon-storage-tool"].Versions[0].Artifact.GetRegistryArtifact()
      if err != nil {
        die(err.Error())
      }
      fmt.Printf("artifact=%s\n", artifact)

      // Download the archive
      err = Download("https://github.com/wavesoft/dot-dom/archive/0.2.2.tar.gz", WithDefaults).
            AndShowProgress("Downloading").
            AndValidateChecksum("1ad0ee9ef8debb1bdccda28073453834d995434f3d211b51bc0e02054a428ad7").
            AndDecompressIfCompressed().
            EventuallyUntarTo("/Users/icharala/Develop/test/docs/xxx/", 1)
      if err != nil {
        die(err.Error())
      }

      // reg, err := registry.GetRegistry(config.DataDir)
      // if err != nil {
      //   die(err.Error())
      // }
      // fmt.Printf("tools=%s\n", reg.Tools["dcos-import-aws-cred"].Name)


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

      // Load registry and repository
      reg, repo := getRegistryRepository(config)

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
      artifact, err := repository.FindFirstRunableArtifact(version.Artifacts)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      // Check if there is already a symlink for this tool
      symlinkTarget, err := ReadBinSymlink(config, tool)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      // Check if we have a tool already installed on this symlink
      if symlinkTarget != "" {
        symlinkedTool, symlinkedVersion := repo.FindToolFromLink(symlinkTarget)

        // If we have a symlink, but we don't have a tool installed on this
        // symlink target, we are most probably going to touch something that
        // does not belong to us... warn the user
        if symlinkedTool == nil {
          if HasBinSymlink(config, tool) && !*fForce {
            die("There is already a tool with the same name in your path. Not installing.")
          }
        } else {

          // If the linked version is desired version, we are good
          if symlinkedVersion.Version.Equals(version.Version) {
            complete(fmt.Sprintf("%s/%s is already there!", tool, version.ToString()))
          }

          // Check if we have the target version already installed, and if
          // we have it, switch link target to the installed version
          targetVersion := repo.FindToolVersion(tool, version.Version)
          if targetVersion != nil {
            err = CreateBinSymlink(config, targetVersion.GetExecutablePath(), tool)
            if err != nil {
              die(fmt.Sprintf("%s: %s", tool, err.Error()))
            }
            complete(fmt.Sprintf("switched %s to %s!", tool, version.ToString()))
          }
        }
      }

      // Download the archive
      installedVer, err := repo.InstallToolVersion(tool, version, artifact)
      if err != nil {
        die(fmt.Sprintf("%s: %s", tool, err.Error()))
      }

      // Install symbolic link
      err = CreateBinSymlink(config, installedVer.GetExecutablePath(), tool)
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

      // Load repository (should be fast)
      repo, err := repository.LoadRepository(config.DataDir)
      if err != nil {
        die(err.Error())
      }

      // Check if we have neither a symlink, nor a tool
      tool := flag.Arg(1)
      if !repo.IsToolInstalled(tool) && !HasBinSymlink(config, tool) {
        complete(fmt.Sprintf("%s is not installed", tool))
      }

      // If the user has not requested the removal of a specific version,
      // remove everything
      if *fVersion == "" {

        // If we have a tray symlink, remove it
        if HasBinSymlink(config, tool) {
          RemoveBinSymlink(config, tool)
        }

        // Remove all tool versions
        if toolRef, ok := repo.Tools[tool]; ok {
          for _, versionRef := range toolRef.Versions {
            repo.UninstallToolVersion(toolRef, &versionRef)
          }

          // Remove the tool itself
          err := repo.UninstallTool(toolRef)
          if err != nil {
            die(err.Error())
          }
        }

        complete(fmt.Sprintf("%s has left the rocket ship!", tool))


      // Otherwise remove the specific version
      } else {

        // Parse version
        verTriplet, err := VersionFromString(*fVersion)
        if err != nil {
          die(fmt.Sprintf("invalid version: %s", *fVersion))
        }

        // Check if there is already a symlink for this tool
        symlinkTarget, err := ReadBinSymlink(config, tool)
        if err != nil {
          die(fmt.Sprintf("%s: %s", tool, err.Error()))
        }

        // Walk versions and find the matching one
        if toolRef, ok := repo.Tools[tool]; ok {
          for _, versionRef := range toolRef.Versions {
            if versionRef.Version.Equals(*verTriplet) {
              repo.UninstallToolVersion(toolRef, &versionRef)

              // If this version is the linked one, remove it
              if symlinkTarget == versionRef.GetExecutablePath() {
                RemoveBinSymlink(config, tool)
              }

              // Check if this was the last version
              if !toolRef.HasInstalledVersions() {
                // Remove the tool itself
                err := repo.UninstallTool(toolRef)
                if err != nil {
                  die(err.Error())
                }
              }

              complete(fmt.Sprintf("%s has left the rocket ship!", tool))
            }
          }
        }

        die(fmt.Sprintf("Unable to find version %s/%s", tool, *fVersion))
      }


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
      if HasBinSymlink(config, tool) {
        RemoveBinSymlink(config, tool)
      } else {
        complete(fmt.Sprintf("%s is not linked (or installed)", tool))
      }

    ///
    /// List the available tools in the repository
    ///
    case "l", "ls", "list":

      // Load registry and repository
      reg, repo := getRegistryRepository(config)

      // Sort names
      var keys []string
      for tool, _ := range reg.Tools {
        keys = append(keys, tool)
      }
      sort.Strings(keys)

      // Print
      fmt.Println("Available tools in the registry:")
      for _, tool := range keys {

        suffix := ""
        if repo.IsToolInstalled(tool) {
          suffix = " *"
        }

        fmt.Printf("%s%s\n", Bold(Gray(wideTab(" "+tool+suffix))), reg.Tools[tool].Desc)
      }

    ///
    /// List the versions of a tool
    ///
    case "i", "info":
      if flag.NArg() < 2 {
        fmt.Println("Missing tool name")
        help()
      }

      // Load registry and repository
      reg, repo := getRegistryRepository(config)

      // Lookup tool
      tool := flag.Arg(1)
      if toolInfo, ok = reg.Tools[tool]; !ok {
        die(fmt.Sprintf("🥔  Could not find tool '%s', here is a potato...", tool))
      }

      // List versions
      fmt.Printf("Available versions for '%s':\n", tool)
      for _, ver := range toolInfo.Versions {

        suffix := ""
        installedTool := repo.FindToolVersion(tool, ver.Version)
        if installedTool != nil {
          suffix = " (installed)"
        }
        fmt.Printf("  * %s %s\n", Bold(Gray(ver.ToString())), suffix)

        if installedTool != nil {
          verSize, err := installedTool.Size()
          if err != nil {
            fmt.Printf("    - size        : error: %s\n", err.Error())
          } else {
            fmt.Printf("    - size        : %s\n", humanize.Bytes(verSize))
          }
        }

        for _, artifact := range ver.Artifacts {
          if artifact.DockerToolArtifact != nil {
            fmt.Printf("    - platform    : docker\n")
            fmt.Printf("      image       : %s:%s\n", artifact.Image, artifact.Tag)
          }
          if artifact.ExecutableToolArtifact != nil {
            if artifact.Interpreter != nil {
              fmt.Printf("    - platform    : interpreter\n")
              fmt.Printf("      interpreter : %s\n", repository.InterpreterName(artifact.Interpreter))
            } else {
              fmt.Printf("    - platform    : %s\n", artifact.Platform)
              fmt.Printf("      CPU arch    : %s\n", artifact.Arch)
            }

            if artifact.Source.WebFileSource != nil {
              fmt.Printf("    - source file : %s\n", artifact.Source.FileURL)
            }
            if artifact.Source.WebArchiveTarSource != nil {
              fmt.Printf("    - source tar  : %s\n", artifact.Source.TarURL)
            }
            if artifact.Source.VCSGitSource != nil {
              fmt.Printf("    - source git  : %s\n", artifact.Source.GitURL)
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

      // Load registry and repository
      reg := getRegistry(config)

      // Lookup tool
      tool := flag.Arg(1)
      if toolInfo, ok = reg.Tools[tool]; !ok {
        die(fmt.Sprintf("🥔  Could not find tool '%s', here is a potato...", tool))
      }

      // Handle help
      if toolInfo.Help.ToolHelpText != nil {
        fmt.Printf("--=[ %s ]=--\n", Bold(Gray(tool)))
        fmt.Println("")
        if toolInfo.Help.Markdown {
          PrintMarkdownText([]byte(toolInfo.Help.Text))
        } else {
          fmt.Println(toolInfo.Help.Text)
        }
        return
      } else if toolInfo.Help.ToolHelpURL != nil {
        if toolInfo.Help.Inline {
          fmt.Printf("--=[ %s ]=--\n", Bold(Gray(tool)))
          contents, err := DownloadHelpText(toolInfo.Help.URL)
          if err != nil {
            die(err.Error())
          }

          if toolInfo.Help.Markdown {
            PrintMarkdownText(contents)
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
      err := upgradeTool()
      if err == AlreadyUpgraded {
        complete(err.Error())
      } else if err != nil {
        die(err.Error())
      }

    ///
    /// Update the database
    ///
    case "update":
      fmt.Printf("Updating registry...\n")
      _, err := registry.UpdateRegistry(
        config.DataDir,
        config.RegistryURL,
        config.RegistryPubKey)
      if err != nil {
        die(err.Error())
      }
      complete("Registry is updated")

    ///
    /// Show the version
    ///
    case "v", "version":
      fmt.Println(VERSION.ToString())

    default:
      fmt.Printf("Unknown action '%s'\n", flag.Arg(0))
      help()
  }
}
