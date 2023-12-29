/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"strconv"

	"github.com/aergoio/aergo/v2/types"
	sha256 "github.com/minio/sha256-simd"
)

// GenerateAddress calculates the raw (not-encoded) address for a private key.
// Address is the raw, internally used identifier of an account
// In Aergo, the address is synonymous with the compressed public key (33 bytes).
func GenerateAddress(pubkey *ecdsa.PublicKey) []byte {
	if pubkey == nil {
		return nil
	}
	addr := new(bytes.Buffer)
	// Calculate compressed pubkey
	binary.Write(addr, binary.LittleEndian, uint8(0x2+pubkey.Y.Bit(0))) // 0x2 for even, 0x3 for odd Y
	keyLength := len(pubkey.X.Bytes())
	if keyLength < 32 { // Add padding
		for i := 1; i < types.AddressLength-keyLength; i++ {
			binary.Write(addr, binary.LittleEndian, uint8(0))
		}
	}
	binary.Write(addr, binary.LittleEndian, pubkey.X.Bytes())
	return addr.Bytes() // 33 bytes
}

func CreateContractID(account []byte, nonce uint64) []byte {
	h := sha256.New()
	h.Write(account)
	h.Write([]byte(strconv.FormatUint(nonce, 10)))
	recipientHash := h.Sum(nil)                   // byte array with length 32
	return append([]byte{0x0C}, recipientHash...) // prepend 0x0C to make it same length as account addresses
}
