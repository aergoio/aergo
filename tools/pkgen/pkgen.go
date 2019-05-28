/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package main

import (
	"fmt"
	"github.com/aergoio/aergo/types"
	"os"

	"github.com/libp2p/go-libp2p-core/crypto"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if 0 == len(argsWithoutProg) {
		panic("Usage: pkgen <fileprefix>")
	}
	prefix := argsWithoutProg[0]
	priv, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	pid, _ := types.IDFromPublicKey(pub)

	pkFile := prefix + ".key"
	pubFile := prefix + ".pub"
	idFile := prefix + ".id"

	pkf, err := os.Create(pkFile)
	if err != nil {
		panic("fail to create key file: " + pkFile)
	}

	pkBytes, err := priv.Bytes()
	if err != nil {
		panic("wrong key <fileprefix>")
	}
	pkf.Write(pkBytes)
	pkf.Sync()

	pubf, err := os.Create(pubFile)
	if err != nil {
		panic("fail to create pub file: " + pubFile)
	}

	pubBytes, err := pub.Bytes()
	if err != nil {
		panic("wrong key <fileprefix>")
	}
	pubf.Write(pubBytes)
	pubf.Sync()

	idf, err := os.Create(idFile)
	idBytes := []byte(types.IDB58Encode(pid))
	if err != nil {
		panic("wrong key <fileprefix>")
	}
	idf.Write(idBytes)
	idf.Sync()

	fmt.Println("Done!")
}
