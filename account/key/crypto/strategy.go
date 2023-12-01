/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"github.com/btcsuite/btcd/btcec/v2"
)

type PrivateKey = btcec.PrivateKey

type KeyCryptoStrategy interface {
	Encrypt(key *PrivateKey, passphrase string) ([]byte, error)
	Decrypt(encrypted []byte, passphrase string) (*PrivateKey, error)
}
