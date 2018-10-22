package key

import (
	sha256 "github.com/minio/sha256-simd"

	"encoding/binary"

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
	sign, err := key.Sign(hash)
	if err != nil {
		return nil, err
	}
	return sign.Serialize(), nil
}

func SignTx(tx *types.Tx, key *aergokey) error {
	hash := CalculateHashWithoutSign(tx.Body)
	sign, err := key.Sign(hash)
	if err != nil {
		return err
	}
	tx.Body.Sign = sign.Serialize()
	tx.Hash = tx.CalculateTxHash()
	return nil
}

//SignTx return transaction which signed with unlocked key
func (ks *Store) SignTx(tx *types.Tx) error {
	addr := tx.Body.Account
	key, exist := ks.unlocked[types.EncodeAddress(addr)]
	if !exist {
		return types.ErrShouldUnlockAccount
	}
	return SignTx(tx, key)
}

//VerifyTx return result to varify sign
func VerifyTx(tx *types.Tx) error {
	txBody := tx.Body
	hash := CalculateHashWithoutSign(txBody)
	sign, err := btcec.ParseSignature(txBody.Sign, btcec.S256())
	if err != nil {
		return err
	}
	account := tx.Body.Account
	pubkey, err := btcec.ParsePubKey(account, btcec.S256())
	if err != nil {
		return err
	}
	if !sign.Verify(hash, pubkey) {
		return types.ErrSignNotMatch
	}
	return nil
}

//VerifyTx return result to varify sign
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
