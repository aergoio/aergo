package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/consensus/impl"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var (
	testNet     bool
	jsonGenesis string
)

func init() {
	initGenesis.Flags().BoolVar(&testNet, "testnet", false, "create genesis block for Aergo TestNet")
	initGenesis.Flags().StringVar(&jsonGenesis, "genesis", "", "genesis json file for private net")

	rootCmd.AddCommand(initGenesis)
}

var initGenesis = &cobra.Command{
	Use:   "init",
	Short: "Create genesis block",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		var genesis *types.Genesis

		core := getCore(cfg.DataDir)
		if core != nil {
			exist := core.GetGenesisInfo()
			if exist != nil {
				fmt.Printf("genesis block(%s) is already initialized\n", enc.ToString(exist.Block().GetHash()))
				core.Close()
				return
			}
		}

		if jsonGenesis != "" {
			fmt.Println("create genesis block for PrivateNet")
			genesis = getGenesis(jsonGenesis)
			if genesis == nil {
				fmt.Printf("failed to obtain GenesisInfo\n")
				return
			}

			if err := impl.ValidateGenesis(genesis); err != nil {
				fmt.Printf(" %s (error:%s)\n", jsonGenesis, err)
				return
			}
		}

		if genesis == nil {
			if testNet == false {
				fmt.Println("create genesis block for Aergo Mainnet")
			} else {
				fmt.Println("create genesis block for Aergo Testnet")
			}
		}

		if core != nil {
			err := core.InitGenesisBlock(genesis, !testNet)
			if err != nil {
				fmt.Printf("fail to init genesis block data (error:%s)\n", err)
			}

			g := core.GetGenesisInfo()
			fmt.Printf("genesis block[%s] is created in (%s)\n", enc.ToString(g.Block().GetHash()), cfg.DataDir)
		}
	},
}

func getGenesis(path string) *types.Genesis {
	file, err := os.Open(jsonGenesis)
	if err != nil {
		fmt.Printf("fail to open %s \n", jsonGenesis)
		return nil
	}
	defer file.Close()
	genesis := new(types.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		fmt.Printf("fail to deserialize %s (error:%s)\n", jsonGenesis, err)
		return nil
	}
	return genesis
}

func getCore(dataDir string) *chain.Core {
	// if initpath is feeded, gaurantee initpath is accessible directory
	fi, err := os.Stat(dataDir)
	if err == nil && !fi.IsDir() {
		fmt.Printf("%s is not a directory\n", dataDir)
		return nil
	}
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("cannot access %s(error:%s)\n", dataDir, err)
			return nil
		}

		err := os.MkdirAll(dataDir, 0755)
		if err != nil {
			fmt.Printf("fail to create %s (error:%s)\n", dataDir, err)
			return nil
		}
	}

	core, err := chain.NewCore(cfg.DbType, dataDir, false, 0)
	if err != nil {
		fmt.Printf("fail to init a blockchain core (error:%s)\n", err)
		return nil
	}

	return core
}
