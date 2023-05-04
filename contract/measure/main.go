//go:build measure
// +build measure

package main

import (
	"io/ioutil"
	"log"

	"github.com/aergoio/aergo/contract"
)

func main() {
	bc, err := contract.LoadDummyChain()
	if err != nil {
		log.Printf("failed to create test database: %v\n", err)
	}

	err = bc.ConnectBlock(
		contract.NewLuaTxAccount("ktlee", 100),
	)
	if err != nil {
		log.Println(err)
	}

	runner := func(luaFileName, mainFuncName string) {
		src, err := ioutil.ReadFile(luaFileName)
		if err != nil {
			log.Printf("can not read `%s`: %s", luaFileName, err.Error())
		}

		err = bc.ConnectBlock(
			contract.NewLuaTxDef("ktlee", luaFileName, 0, string(src)),
		)
		if err != nil {
			log.Println(err)
		}

		err = bc.ConnectBlock(
			contract.NewLuaTxCall("ktlee", luaFileName, 0, `{"Name": "`+mainFuncName+`"}`),
		)
		if err != nil {
			log.Println(err)
		}
	}

	runner("./bf.lua", "basic_fns")
	runner("./bf.lua", "string_fns")
	runner("./bf.lua", "table_fns1")
	runner("./bf.lua", "table_fns2")
	runner("./bf.lua", "table_fns3")
	runner("./bf.lua", "table_fns4")
	runner("./bf.lua", "math_fns")
	runner("./bf.lua", "bit_fns")
	runner("./aef.lua", "run_test")
}
