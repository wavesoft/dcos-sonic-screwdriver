package shared

import (
  "fmt"
  "strings"
  "strconv"
)

/**
 * Version representation
 */
type VersionTriplet   [3]float64

/**
 * Get the version as string
 */
func (v VersionTriplet) ToString() string {
  return fmt.Sprintf("%d.%d.%d",
    uint32(v[0]),
    uint32(v[1]),
    uint32(v[2]))
}

/**
 * Parse a version string to a version info structure
 */
func VersionFromString(version string) (*VersionTriplet, error) {
  verInfo := new(VersionTriplet)
  verFrag := strings.SplitN(version, ".", 3)

  for idx, fragStr := range verFrag {
    fragInt, err := strconv.Atoi(fragStr)
    if err != nil {
      return nil, err
    }
    verInfo[idx] = float64(fragInt)
  }

  return verInfo, nil
}

/**
 * Compare two versions
 */
func (v VersionTriplet) Equals(n VersionTriplet) bool {
  return v[0] == n[0] &&
         v[1] == n[1] &&
         v[2] == n[2]
}

/**
 * Compare two versions
 */
func (v VersionTriplet) LessThan(n * VersionTriplet) bool {
  var left float64 = v[0] * 1000000 + v[1] * 1000 + v[2]
  var right float64 = n[0] * 1000000 + n[1] * 1000 + n[2]
  return left < right
}
func (v VersionTriplet) GraterThan(n * VersionTriplet) bool {
  var left float64 = v[0] * 1000000 + v[1] * 1000 + v[2]
  var right float64 = n[0] * 1000000 + n[1] * 1000 + n[2]
  return left > right
}
