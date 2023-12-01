/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"github.com/btcsuite/btcd/btcec/v2"
)

// Identity is a raw, i.e. decoded address generated from a public key
type Identity = []byte

// PrivateKey is a raw, decrypted private key
type PrivateKey = btcec.PrivateKey

// Storage defines an interfaces for persistent storage of identities
type Storage interface {
	Save(identity Identity, passphrase string, key *PrivateKey) (Identity, error)
	Load(identity Identity, passphrase string) (*PrivateKey, error)
	List() ([]Identity, error)
	Close()
}
