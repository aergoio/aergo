/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aergoio/aergo/v2/types"
)

// readGenesis is based on code in cmd/aergosvr/cmd.go .
func readGenesis(filePath string) (ret *types.Genesis, rerr error) {
	jsonpath := filePath

	file, err := os.Open(jsonpath)
	if err != nil {
		rerr = fmt.Errorf("fail to open %s \n", jsonpath)
		return
	}
	defer file.Close()

	// use default config's DataDir if empty
	genesis := new(types.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		rerr = fmt.Errorf("fail to deserialize %s (error:%s)\n", jsonpath, err)
		return
	}

	if err := genesis.Validate(); err != nil {
		rerr = fmt.Errorf(" %s (error:%s)\n", jsonpath, err)
		return
	}
	ret = genesis
	return
}
