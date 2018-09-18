package key

import (
	"bytes"

	sha256 "github.com/minio/sha256-simd"

	"encoding/binary"

	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
)

//Sign return sign with key in the store
func (ks *Store) Sign(addr Address, pass string, hash []byte) ([]byte, error) {
	k, err := ks.getKey(addr, pass)
	if k == nil {
		return nil, err
	}
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), k)
	return btcec.SignCompact(btcec.S256(), key, hash, true)
}

func SignTx(tx *types.Tx, key *aergokey) error {
	hash := CalculateHashWithoutSign(tx.Body)
	sign, err := btcec.SignCompact(btcec.S256(), key, hash, true)
	if err != nil {
		return err
	}
	tx.Body.Sign = sign
	tx.Hash = tx.CalculateTxHash()
	return nil
}

//SignTx return transaction which signed with unlocked key
func (ks *Store) SignTx(tx *types.Tx) error {
	addr := tx.Body.Account
	key, exist := ks.unlocked[types.EncodeAddress(addr)]
	if !exist {
		return message.ErrShouldUnlockAccount
	}
	return SignTx(tx, key)
}

//VerifyTx return result to varify sign
func VerifyTx(tx *types.Tx) error {
	txBody := tx.Body
	hash := CalculateHashWithoutSign(txBody)
	pubkey, _, err := btcec.RecoverCompact(btcec.S256(), txBody.Sign, hash)
	if err != nil {
		return message.ErrCouldNotRecoverPubKey
	}
	address := GenerateAddress(pubkey.ToECDSA())
	if !bytes.Equal(address, txBody.Account) {
		return message.ErrSignNotMatch
	}
	return nil
}
func (ks *Store) VerifyTx(tx *types.Tx) error {
	return VerifyTx(tx)
}

//CalculateHashWithoutSign return hash of tx without sign field
func CalculateHashWithoutSign(txBody *types.TxBody) []byte {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, txBody.Nonce)
	h.Write(txBody.Account)
	h.Write(txBody.Recipient)
	binary.Write(h, binary.LittleEndian, txBody.Amount)
	h.Write(txBody.Payload)
	binary.Write(h, binary.LittleEndian, txBody.Limit)
	binary.Write(h, binary.LittleEndian, txBody.Price)
	binary.Write(h, binary.LittleEndian, txBody.Type)
	return h.Sum(nil)
}
