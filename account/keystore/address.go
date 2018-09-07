package keystore

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
)

type Address = []byte

func generateAddress(pubkey *ecdsa.PublicKey) []byte {
	addr := new(bytes.Buffer)
	binary.Write(addr, binary.LittleEndian, pubkey.X.Bytes())
	binary.Write(addr, binary.LittleEndian, pubkey.Y.Bytes())
	return addr.Bytes()[:20] //TODO: ADDRESSLENGTH ?
}
