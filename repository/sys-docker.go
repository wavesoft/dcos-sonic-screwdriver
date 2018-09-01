package repository

import (
  "fmt"
  . "github.com/mesosphere/dcos-sonic-screwdriver/shared"
)

/**
 * Check if docker found in the system
 */
func DockerIsAvailable() bool {
  return SysHasCommand("docker")
}

/**
 * Pull docker image, while echoing progress on terminal
 */
func DockerPullImage(image string, tag string) error {
  exitcode, err := ExecuteAndPassthrough("docker", "pull", image + ":" + tag)
  if err != nil {
    return err
  }
  if exitcode != 0 {
    return fmt.Errorf("Unable to pull the docker image")
  }
  return nil
}

/**
 * Remove docker image, while echoing progress on terminal
 */
func DockerRemoveImage(image string, tag string) error {
  exitcode, err := ExecuteAndPassthrough("docker", "rmi", image + ":" + tag)
  if err != nil {
    return err
  }
  if exitcode != 0 {
    return fmt.Errorf("Unable to pull the docker image")
  }
  return nil
}
