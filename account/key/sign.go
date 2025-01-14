/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"encoding/binary"

	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sha256 "github.com/minio/sha256-simd"
)

// Sign return signature using stored key
func (ks *Store) Sign(addr Identity, pass string, hash []byte) ([]byte, error) {
	key, err := ks.getKey(addr, pass)
	if err != nil {
		return nil, err
	}
	sign := ecdsa.Sign(key, hash)
	return sign.Serialize(), nil
}

// SignTx return tx signature using stored key
func SignTx(tx *types.Tx, key *aergokey) error {
	hash := CalculateHashWithoutSign(tx.Body)
	sign := ecdsa.Sign(key, hash)
	tx.Body.Sign = sign.Serialize()
	tx.Hash = tx.CalculateTxHash()
	return nil
}

// SignTx return transaction which signed with unlocked key. if requester is nil, requester is assumed to tx.Account
func (ks *Store) SignTx(tx *types.Tx, requester []byte) error {
	addr := tx.Body.Account
	if requester != nil {
		addr = requester
	}
	keyPair, exist := ks.unlocked[types.EncodeAddress(addr)]
	if !exist {
		return types.ErrShouldUnlockAccount
	}
	return SignTx(tx, keyPair.key)
}

// VerifyTx return result to verify sign
func VerifyTx(tx *types.Tx) error {
	return VerifyTxWithAddress(tx, tx.Body.Account)
}

func VerifyTxWithAddress(tx *types.Tx, address []byte) error {
	txBody := tx.Body
	hash := CalculateHashWithoutSign(txBody)
	sign, err := ecdsa.ParseSignature(txBody.Sign)
	if err != nil {
		return err
	}
	pubkey, err := btcec.ParsePubKey(address)
	if err != nil {
		return err
	}
	if !sign.Verify(hash, pubkey) {
		return types.ErrSignNotMatch
	}
	return nil
}

// VerifyTx return result to varify sign
func (ks *Store) VerifyTx(tx *types.Tx) error {
	return VerifyTx(tx)
}

// CalculateHashWithoutSign return hash of tx without sign field
func CalculateHashWithoutSign(txBody *types.TxBody) []byte {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, txBody.Nonce)
	h.Write(txBody.Account)
	h.Write(txBody.Recipient)
	h.Write(txBody.Amount)
	h.Write(txBody.Payload)
	binary.Write(h, binary.LittleEndian, txBody.GasLimit)
	h.Write(txBody.GasPrice)
	binary.Write(h, binary.LittleEndian, txBody.Type)
	h.Write(txBody.ChainIdHash)
	return h.Sum(nil)
}
