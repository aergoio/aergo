/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"errors"
	"github.com/coreos/go-semver/semver"
	"regexp"
)

// AergoVersion follows sementic versioning https://semver.org/
type AergoVersion = semver.Version

// const verPattern = `v([0-9]+)\.([0-9]+)\..([0-9]+)(-(.+))?`
const verPattern = `v[0-9].+`
var checker, _ = regexp.Compile(verPattern)

// ParseAergoVersion parse version string to sementic version with slightly different manner. This function allows not only standard sementic version but also version strings containing v in prefixes.
func ParseAergoVersion(verStr string) (AergoVersion, error) {
	if checker.MatchString(verStr) {
		verStr = verStr[1:]
	}
	ver, err := semver.NewVersion(verStr)
	if err != nil {
		return AergoVersion{}, errors.New("invalid version "+verStr)
	}
	return AergoVersion(*ver), nil
}

// Supported Aergo version. polaris will register aergosvr within the version range. This version range should be modified when new release is born.
const (
	MinimumAergoVersion = "v1.2.1"
	MaximumAergoVersion = "v2.0.0"
)
var (
	minAergoVersion AergoVersion // inclusive
	maxAergoVersion AergoVersion // exclusive
)

func init() {
	var err error
	minAergoVersion, err = ParseAergoVersion(MinimumAergoVersion)
	if err != nil {
		panic("Invalid minimum version "+MinimumAergoVersion)
	}
	maxAergoVersion, err = ParseAergoVersion(MaximumAergoVersion)
	if err != nil {
		panic("Invalid maximum version "+MaximumAergoVersion)
	}
}


func CheckVersion(version string) bool {
	ver, err := ParseAergoVersion(version)
	if err != nil {
		return false
	}

	return ver.Compare(minAergoVersion) >= 0 && ver.Compare(maxAergoVersion) < 0
}
