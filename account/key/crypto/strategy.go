/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type PrivateKey = secp256k1.PrivateKey

type KeyCryptoStrategy interface {
	Encrypt(key *PrivateKey, passphrase string) ([]byte, error)
	Decrypt(encrypted []byte, passphrase string) (*PrivateKey, error)
}
