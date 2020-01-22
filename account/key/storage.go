package key

import (
	"github.com/btcsuite/btcd/btcec"
)

type Identity = []byte
type PrivateKey = btcec.PrivateKey

type Storage interface {
	Save(identity Identity, password string, key *PrivateKey) (Identity, error)
	Load(identity Identity, password string) (*PrivateKey, error)
	List() ([]Identity, error)
	Close()
}
