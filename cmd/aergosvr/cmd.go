package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/types"
	"github.com/spf13/cobra"
)

var dataDir string

func init() {
	initGenesis.Flags().StringVar(&dataDir, "dir", "", "Data directory")
	rootCmd.AddCommand(initGenesis)
	rootCmd.AddCommand(versionCmd)
}

var initGenesis = &cobra.Command{
	Use:   "init",
	Short: "Create genesis block based on input json file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "Usage: aergosvr init {genesis.json} --dir {target directory}")
			return
		}
		jsonpath := args[0]

		file, err := os.Open(jsonpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to open %s \n", jsonpath)
			return
		}
		defer file.Close()

		if dataDir == "" {
			dataDir = cfg.DataDir
		}

		// if initpath is feeded, gaurantee initpath is accessible directory
		fi, err := os.Stat(dataDir)
		if err == nil && !fi.IsDir() {
			fmt.Fprintf(os.Stderr, "%s is not a directory\n", dataDir)
			return
		}
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "cannot access %s(error:%s)\n", dataDir, err)
				return
			}

			err := os.MkdirAll(dataDir, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "fail to create %s (error:%s)\n", dataDir, err)
				return
			}
		}
		// use default config's DataDir if empty
		genesis := new(types.Genesis)
		if err := json.NewDecoder(file).Decode(genesis); err != nil {
			fmt.Fprintf(os.Stderr, "fail to deserialize %s (error:%s)\n", jsonpath, err)
			return
		}

		if err := genesis.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, " %s (error:%s)\n", jsonpath, err)
			return
		}
		core, err := chain.NewCore(cfg.DbType, dataDir, false, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to init a blockchain core (error:%s)\n", err)
		}

		err = core.InitGenesisBlock(genesis)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to init genesis block data (error:%s)\n", err)
		}
		fmt.Fprintf(os.Stderr, "genesis block is created in (%s)\n", dataDir)

		core.Close()
	},
}

var githash = "No git hash provided"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Aergosvr",
	Long:  `All software has versions. This is Aergo's`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Aergosvr %s\n", githash)
	},
}
