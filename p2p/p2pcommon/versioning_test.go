/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/coreos/go-semver/semver"
	"os/exec"
	"strings"
	"testing"
)

func TestParseAergoVersion(t *testing.T) {

	tests := []struct {
		name    string
		args    string
		wantM int
		wantm int
		wantp int
		wantA string
		wantErr bool
	}{
		{"TR1","v0.0.1", 0,0, 1, "",false},
		{"TR1","v1.2.3", 1,2, 3, "",false},
		{"TR1","0.0.1", 0,0, 1, "",false},
		{"TR1","1.2.3", 1,2, 3, "",false},
		{"TDev1","v1.2.2-20-g8905410d", 1,2, 2, "20-g8905410d",false},
		{"TDev2","vd1.2.2-20-g8905410d", 0,0, 0, "",true},
		{"TDev3","vd1.2.2.55", 0,0, 0, "",true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAergoVersion(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAergoVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != int64(tt.wantM) || got.Minor != int64(tt.wantm) ||got.Patch != int64(tt.wantp) {
					t.Errorf("ParseAergoVersion() got = %v, want %v.%v.%v-%v", got.String(), tt.wantM, tt.wantm, tt.wantp, tt.wantA)
				} else if len(tt.wantA) > 0 && got.PreRelease != semver.PreRelease(tt.wantA) {
					t.Errorf("ParseAergoVersion() got = prerelase %v, want %v", got.PreRelease, tt.wantA)
				}
			}
		})
	}
}

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		name string
		arg string
		want bool
	}{
		{"TOld", "0.0.1", false},
		{"TOld2", "v1.2.0", false},
		{"TOld2", "v1.2.99", false},
		{"TOld2", "v1.99.99", false},
		{"TPre", "2.0.0-30-g8905410d", false},
		{"TMin", "2.0.0", true},
		{"TMin3", "v2.0.0", true},
		{"TMax", "2.99.99", true},
		{"TMax3", "v2.99.99", true},
		{"TSomewhatUnclear", "3.0.0-30-g8905410d", true},
		{"TNew", "3.0.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckVersion(tt.arg); got != tt.want {
				t.Errorf("CheckVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrentVersion(t *testing.T) {
	t.Skip("This test is not worked well,")
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "git"
	cmdArgs := []string{"describe", "--tags"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		t.Skipf("Skip test since not testable environment: %s", err.Error())
	}
	tagVer := strings.TrimSpace(string(cmdOut))
	_, err = ParseAergoVersion(tagVer)
	if err != nil {
		t.Skipf("Skip test since not testable environment: %s", err.Error())
	}

	if !CheckVersion(tagVer) {
		t.Fatalf("invalid version filtering setting. build version %v , want between %v (inclusive) and %v (exclusive)",
			tagVer, MinimumAergoVersion, MaximumAergoVersion)
	}
	t.Logf("build version is %v ", tagVer)

}