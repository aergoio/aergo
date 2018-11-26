package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var initpath string

func init() {
	initGenesis.Flags().StringVar(&initpath, "dir", "", "Data directory")
	rootCmd.AddCommand(initGenesis)
}

var initGenesis = &cobra.Command{
	Use:   "init",
	Short: "Create genesis block based on input json file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: aergosvr init {genesis.json} --data {target directory}")
			return
		}
		jsonpath := args[0]

		file, err := os.Open(jsonpath)
		if err != nil {
			fmt.Printf("fail to open %s \n", jsonpath)
			return
		}
		defer file.Close()

		if initpath == "" {
			initpath = cfg.DataDir
		}

		// if initpath is feeded, gaurantee initpath is accessible directory
		fi, err := os.Stat(initpath)
		if err == nil && !fi.IsDir() {
			fmt.Printf("%s is not a directory\n", initpath)
			return
		}
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("cannot access %s(err:%s)\n", initpath, err)
				return
			} else {
				err := os.MkdirAll(initpath, 0755)
				if err != nil {
					fmt.Printf("fail to create %s (err:%s)\n", initpath, err)
					return
				}
			}
		}
		// use default config's DataDir if empty
		genesis := new(types.Genesis)
		if err := json.NewDecoder(file).Decode(genesis); err != nil {
			fmt.Printf("fail to deserialize %s(error:%s)\n", jsonpath, err)
			return
		}

		chainsvc := chain.NewChainService(cfg)
		err = chainsvc.InitGenesisBlock(genesis, cfg.DbType, initpath)
		if err != nil {
			fmt.Printf("fail to init genesis block data (error:%s)\n", err)
		}
		fmt.Printf("genesis block is created in (%s)\n", initpath)

		chainsvc.CloseDB()
	},
}
