package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aergoio/aergo/v2/types"
)

func getGenesis(path string) *types.Genesis {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("fail to open %s \n", path)
		return nil
	}
	defer file.Close()
	genesis := new(types.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		fmt.Printf("fail to deserialize %s (error:%s)\n", path, err)
		return nil
	}
	return genesis
}

func main() {

	argsWithoutProg := os.Args[1:]
	if 0 == len(argsWithoutProg) {
		panic("Usage: dumpblk <genesis file path>")
	}

	g := getGenesis(argsWithoutProg[0])
	if g == nil || g.Validate() != nil {
		panic("fail to read or validate")
	}

	bs, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}

	str := "\"" + hex.EncodeToString(bs) + "\""
	fmt.Println(str)
}
