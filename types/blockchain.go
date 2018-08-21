/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// BlockNo is the height of a block, which starts from 0 (genesis block).
type BlockNo = uint64

// BlockNoToBytes represents to serialize block no to bytes
func BlockNoToBytes(bn BlockNo) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, bn)
	return buf
}

// BlockNoFromBytes represents to deserialize bytes to block no
func BlockNoFromBytes(raw []byte) BlockNo {
	buf := binary.LittleEndian.Uint64(raw)
	return BlockNo(buf)
}

// NewBlock represents to create a block to store transactions.
func NewBlock(prevBlock *Block, txs []*Tx, ts int64) *Block {
	var prevBlockHash []byte
	var blockNo BlockNo

	if prevBlock != nil {
		prevBlockHash = prevBlock.Hash
		blockNo = prevBlock.Header.BlockNo + 1
	}

	body := BlockBody{
		Txs: txs,
	}
	header := BlockHeader{
		PrevBlockHash: prevBlockHash,
		BlockNo:       blockNo,
		Timestamp:     ts,
		//BlockRootHash: nil,
		//StateRootHash: nil,
	}
	block := Block{
		Header: &header,
		Body:   &body,
	}
	block.Header.TxsRootHash = CalculateTxsRootHash(body.Txs)
	block.Hash = block.CalculateBlockHash()
	return &block
}

// CalculateBlockHash computes sha256 hash of block header.
func (block *Block) CalculateBlockHash() []byte {
	digest := sha256.New()
	writeBlockHeader(digest, block.Header)

	return digest.Sum(nil)
}

// Sign adds a pubkey and a block signature to block.
func (block *Block) Sign(privKey crypto.PrivKey) error {
	var err error

	if err = block.setPubKey(privKey.GetPublic()); err != nil {
		return err
	}

	var msg []byte
	if msg, err = block.Header.Bytes(); err != nil {
		return err
	}

	var sig []byte
	if sig, err = privKey.Sign(msg); err != nil {
		return err
	}
	block.Header.Sign = sig

	//block hash must be recomputed
	block.Hash = block.CalculateBlockHash()
	return nil
}

// Bytes returns a binary representation of bh.
func (bh *BlockHeader) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := writeBlockHeader(&buf, bh); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// VerifySign verifies the signature of block.
func (block *Block) VerifySign() (valid bool, err error) {
	var pubKey crypto.PubKey
	if pubKey, err = crypto.UnmarshalPublicKey(block.Header.PubKey); err != nil {
		return false, err
	}

	var msg []byte
	if msg, err = block.Header.Bytes(); err != nil {
		return false, err
	}

	if valid, err = pubKey.Verify(msg, block.Header.Sign); err != nil {
		return
	}

	return valid, nil
}

// BpID returns its Block Producer's ID from block.
func (block *Block) BpID() (id peer.ID, err error) {
	var pubKey crypto.PubKey
	if pubKey, err = crypto.UnmarshalPublicKey(block.Header.PubKey); err != nil {
		return peer.ID(""), err
	}

	if id, err = peer.IDFromPublicKey(pubKey); err != nil {
		return peer.ID(""), err
	}

	return
}

// ID returns the base64 encoded formated ID (hash) of block.
func (block *Block) ID() string {
	hash := block.GetHash()
	if hash != nil {
		return base64.StdEncoding.EncodeToString(hash)
	}

	return ""

}

// PrevID returns the base64 encoded formated ID (hash) of the parent block.
func (block *Block) PrevID() string {
	hash := block.GetHeader().GetPrevBlockHash()
	if hash != nil {
		return base64.StdEncoding.EncodeToString(hash)
	}

	return ""

}

// SetPubKey sets block.Header.PubKey to pubkey.
func (block *Block) setPubKey(pubKey crypto.PubKey) error {
	var pk []byte
	var err error
	if pk, err = pubKey.Bytes(); err != nil {
		return err
	}
	block.Header.PubKey = pk

	return nil
}

func (block *Block) Clone() *Block {
	if block == nil {
		return nil
	}
	res := Block{
		Header: block.Header.Clone(),
		Body:   block.Body.Clone(),
	}
	if res.Header != nil {
		if res.Body != nil {
			res.Header.TxsRootHash = CalculateTxsRootHash(res.Body.Txs)
		}
		res.Hash = res.CalculateBlockHash()
	}
	return &res
}

func (header *BlockHeader) Clone() *BlockHeader {
	if header == nil {
		return nil
	}
	// XXX Very risky code. This part below must be rewritten whenever a new
	// XXX struct member is addded. Need to rewritten in a more robust way.
	return &BlockHeader{
		PrevBlockHash: Clone(header.PrevBlockHash).([]byte),
		BlockNo:       header.BlockNo,
		Timestamp:     header.Timestamp,
		PubKey:        Clone(header.PubKey).([]byte),
		Sign:          Clone(header.Sign).([]byte),
	}
}
func (body *BlockBody) Clone() *BlockBody {
	if body == nil {
		return nil
	}
	txs := make([]*Tx, len(body.Txs))
	for i, v := range body.Txs {
		txs[i] = v.Clone()
	}
	return &BlockBody{
		Txs: txs,
	}
}

func writeBlockHeader(w io.Writer, bh *BlockHeader) error {
	for _, f := range []interface{}{
		bh.PrevBlockHash,
		bh.BlockNo,
		bh.Timestamp,
		bh.TxsRootHash,
		bh.PubKey,
	} {
		if err := binary.Write(w, binary.LittleEndian, f); err != nil {
			return err
		}
	}
	return nil
}

// CalculateBlocksRootHash generates merkle tree of block headers and returns root hash.
func CalculateBlocksRootHash(blocks []*Block) []byte {
	return nil
}

// CalculateTxsRootHash generates merkle tree of transactions and returns root hash.
func CalculateTxsRootHash(txs []*Tx) []byte {
	return nil
}

func NewTx() *Tx {
	tx := &Tx{
		Body: &TxBody{
			Nonce: uint64(1),
		},
	}
	return tx
}

func (tx *Tx) CalculateTxHash() []byte {
	txBody := tx.Body
	digest := sha256.New()
	binary.Write(digest, binary.LittleEndian, txBody.Nonce)
	digest.Write(txBody.Account)
	digest.Write(txBody.Recipient)
	binary.Write(digest, binary.LittleEndian, txBody.Amount)
	digest.Write(txBody.Payload)
	binary.Write(digest, binary.LittleEndian, txBody.Limit)
	binary.Write(digest, binary.LittleEndian, txBody.Price)
	digest.Write(txBody.Sign)
	binary.Write(digest, binary.LittleEndian, txBody.Type)
	return digest.Sum(nil)
}

func (tx *Tx) Clone() *Tx {
	if tx == nil {
		return nil
	}
	if tx.Body == nil {
		return &Tx{}
	}
	body := &TxBody{
		Nonce:     tx.Body.Nonce,
		Account:   Clone(tx.Body.Account).([]byte),
		Recipient: Clone(tx.Body.Recipient).([]byte),
		Amount:    tx.Body.Amount,
		Payload:   Clone(tx.Body.Payload).([]byte),
		Limit:     tx.Body.Limit,
		Price:     tx.Body.Price,
		Sign:      Clone(tx.Body.Sign).([]byte),
		Type:      tx.Body.Type,
	}
	res := &Tx{
		Body: body,
	}
	res.Hash = tx.CalculateTxHash()
	return res
}
