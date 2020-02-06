/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"github.com/btcsuite/btcd/btcec"
)

type Identity = []byte
type PrivateKey = btcec.PrivateKey

type Storage interface {
	Save(identity Identity, passphrase string, key *PrivateKey) (Identity, error)
	Load(identity Identity, passphrase string) (*PrivateKey, error)
	List() ([]Identity, error)
	Close()
}
