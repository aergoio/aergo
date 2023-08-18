//go:generate go run ./hardfork_gen/main.go hardfork.json hardfork_gen.go

/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package config

import (
	"fmt"
	"strconv"

	"github.com/aergoio/aergo/v2/types"
)

type forkError struct {
	version           string
	latest, node, cdb uint64
}

func newForkError(version string, latest, node, cdb uint64) *forkError {
	return &forkError{version, latest, node, cdb}
}

func (e *forkError) Error() string {
	return fmt.Sprintf(
		"the fork %q is incompatible: latest block(%d), node(%d), and chain(%d)",
		e.version, e.latest, e.node, e.cdb,
	)
}

func isFork(forkBlkNo, currBlkNo types.BlockNo) bool {
	return forkBlkNo <= currBlkNo
}

func checkOlderNode(maxVer uint64, latest types.BlockNo, dbCfg HardforkDbConfig) error {
	for k, bno := range dbCfg {
		ver, err := strconv.ParseUint(k[1:], 10, 64)
		if err != nil {
			return err
		}
		if ver > maxVer {
			if isFork(bno, latest) {
				return newForkError(k, latest, 0, bno)
			}
		}
	}
	return nil
}
