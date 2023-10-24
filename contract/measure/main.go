//go:build measure
// +build measure

package main

import (
	"io/ioutil"
	"log"

	"github.com/aergoio/aergo/v2/contract/vm_dummy"
	"github.com/aergoio/aergo/v2/types"
)

func main() {
	bc, err := vm_dummy.LoadDummyChain()
	if err != nil {
		log.Printf("failed to create test database: %v\n", err)
	}

	err = bc.ConnectBlock(
		vm_dummy.NewLuaTxAccount("user1", 100, types.Aergo),
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
			vm_dummy.NewLuaTxDeploy("user1", luaFileName, 0, string(src)),
		)
		if err != nil {
			log.Println(err)
		}

		err = bc.ConnectBlock(
			vm_dummy.NewLuaTxCall("user1", luaFileName, 0, `{"Name": "`+mainFuncName+`"}`),
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
