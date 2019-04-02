package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"github.com/gogo/protobuf/proto"
)

func main() {

	argsWithoutProg := os.Args[1:]
	if 0 == len(argsWithoutProg) {
		panic("Usage: mpdiag /path/to/mempool.dump>")
	}

	filename := argsWithoutProg[0]

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("error: failed to open file")
		return
	}
	reader := csv.NewReader(bufio.NewReader(file))
	var count int
	for {
		buf := types.Tx{}
		rc, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Println("error: read file err")
			}
			break
		}
		count++
		dataBuf, err := enc.ToBytes(rc[0])
		if err != nil {
			fmt.Println("error: decode tx err, continue")
			continue
		}
		err = proto.Unmarshal(dataBuf, &buf)
		if err != nil {
			fmt.Println("error: unmarshall tx err, continue")
			continue
		}
		//mp.put(types.NewTransaction(&buf)) // nolint: errcheck
		b, e := json.MarshalIndent(util.ConvTx(types.NewTransaction(&buf).GetTx()), "", " ")
		if e == nil {
			fmt.Printf("%s\n", b)
		} else {
			fmt.Println("error: convert to json ")
		}
	}

	//fmt.Println("total ", count, "txs")

}
