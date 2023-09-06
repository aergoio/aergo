package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "mpdumpdiag",
	}
	printCmd = &cobra.Command{
		Use:  "print <path to mempool dump>",
		Args: cobra.MinimumNArgs(1),
		Run:  runPrintCmd,
	}
	genCmd = &cobra.Command{
		Use:  "gen <file which has json formatted tx array> <dump path to be generated>",
		Args: cobra.MinimumNArgs(2),
		Run:  runGenCmd,
	}
)

func init() {
	rootCmd.SetOutput(os.Stdout)
	rootCmd.AddCommand(printCmd)
	rootCmd.AddCommand(genCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runPrintCmd(cmd *cobra.Command, args []string) {
	filename := args[0]

	file, err := os.Open(filename)
	if err != nil {
		cmd.Printf("error: failed to open file %s\n", filename)
		return
	}
	reader := bufio.NewReader(file)

	var count int
	var out []*util.InOutTx
	for {
		buf := types.Tx{}
		byteInt := make([]byte, 4)
		_, err := io.ReadFull(reader, byteInt)
		if err != nil {
			if err != io.EOF {
				cmd.Println("error: on read file for getting record length", err.Error())
			}
			break
		}

		reclen := binary.LittleEndian.Uint32(byteInt)
		buffer := make([]byte, int(reclen))
		_, err = io.ReadFull(reader, buffer)
		if err != nil {
			if err != io.EOF {
				cmd.Println("error: on read file during loading", err.Error())
			}
			break
		}

		err = proto.Unmarshal(buffer, &buf)
		if err != nil {
			cmd.Println("error: unmarshall tx err, continue", err.Error())
			continue
		}
		count++
		//mp.put(types.NewTransaction(&buf)) // nolint: errcheck

		out = append(out, util.ConvTx(types.NewTransaction(&buf).GetTx()))
	}
	b, e := json.MarshalIndent(out, "", " ")
	if e == nil {
		cmd.Printf("%s\n", b)
	} else {
		cmd.Println("error: convert to json ")
	}
	//fmt.Println("total ", count, "txs")
}

func runGenCmd(cmd *cobra.Command, args []string) {
	file, err := os.Create(args[1])
	if err != nil {
		cmd.Println("error: failed to create target file", err.Error())
		return
	}
	defer file.Close() // nolint: errcheck

	writer := bufio.NewWriter(file)
	defer writer.Flush() //nolint: errcheck

	b, err := ioutil.ReadFile(args[0])
	if err != nil {
		cmd.Println("error: failed to read source file", err.Error())
	}
	txlist, err := util.ParseBase58Tx(b)
	for _, v := range txlist {
		var total_data []byte
		data, err := proto.Marshal(v)
		if err != nil {
			cmd.Println("error: marshal failed", err.Error())
			continue
		}

		byteInt := make([]byte, 4)
		binary.LittleEndian.PutUint32(byteInt, uint32(len(data)))
		total_data = append(total_data, byteInt...)
		total_data = append(total_data, data...)
		length := len(total_data)
		for {
			size, err := writer.Write(total_data)
			if err != nil {
				cmd.Println("error: writing encoded tx fail", err.Error())
				break
			}
			if length != size {
				total_data = total_data[size:]
				length -= size
			} else {
				break
			}
		}
	}
}
