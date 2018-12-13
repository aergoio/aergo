/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"math/big"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/internal/merkle"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/minio/sha256-simd"
)

const (
	// DefaultMaxBlockSize is the maximum block size (currently 1MiB)
	DefaultMaxBlockSize = 1 << 20
	DefaultCoinbaseFee  = 1
	lastFieldOfBH       = "Sign"
	DefaultTxVerifyTime = time.Microsecond * 200
)

type AvgTime struct {
	val  atomic.Value
	mavg *MovingAverage
}

var (
	DefaultVerifierCnt          = int(math.Max(float64(runtime.NumCPU()/2), float64(1)))
	DefaultAvgTimeSize          = 60 * 60 * 24
	AvgTxVerifyTime    *AvgTime = NewAvgTime(DefaultAvgTimeSize)
)

//MaxAER is maximum value of aergo
var MaxAER *big.Int

//StakingMinimum is minimum amount for staking
var StakingMinimum *big.Int

var lastIndexOfBH int

func init() {
	MaxAER = new(big.Int)
	MaxAER.SetString("500000000000000000000000000", 10)
	StakingMinimum = new(big.Int)
	StakingMinimum.SetString("1000000000000000000", 10)
	lastIndexOfBH = getLastIndexOfBH()
}

func NewAvgTime(sizeMavg int) *AvgTime {
	avgTime := &AvgTime{}
	avgTime.mavg = NewMovingAverage(sizeMavg)
	avgTime.val.Store(time.Duration(0))
	return avgTime
}

func (avgTime *AvgTime) Get() time.Duration {
	var avg time.Duration
	aopv := avgTime.val.Load()
	if aopv != nil {
		avg = aopv.(time.Duration)
	} else {
		panic("AvgTxSignTime is not set")
	}
	return avg
}

func (avgTime *AvgTime) UpdateAverage(cur time.Duration) time.Duration {
	newAvg := time.Duration(avgTime.mavg.Add(int64(cur)))
	avgTime.set(newAvg)

	return newAvg
}

func (avgTime *AvgTime) set(val time.Duration) {
	avgTime.val.Store(val)
}

func getLastIndexOfBH() (lastIndex int) {
	v := reflect.ValueOf(BlockHeader{})

	nField := v.NumField()
	var i int
	for i = 0; i < nField; i++ {
		name := v.Type().Field(i).Name
		if name == lastFieldOfBH {
			lastIndex = i
			break
		}
	}

	return i
}

// ChainAccessor is an interface for a another actor module to get info of chain
type ChainAccessor interface {
	GetBestBlock() (*Block, error)
	// GetBlock return block of blockHash. It return nil and error if not found block of that hash or there is a problem in db store
	GetBlock(blockHash []byte) (*Block, error)
	// GetHashByNo returns hash of block. It return nil and error if not found block of that number or there is a problem in db store
	GetHashByNo(blockNo BlockNo) ([]byte, error)
}

type SyncContext struct {
	PeerID peer.ID

	BestNo   BlockNo
	TargetNo BlockNo //sync target blockno

	CommonAncestor *Block

	TotalCnt   uint64
	RemainCnt  uint64
	LastAnchor BlockNo
}

func NewSyncCtx(peerID peer.ID, targetNo uint64, bestNo uint64) *SyncContext {
	return &SyncContext{PeerID: peerID, TargetNo: targetNo, BestNo: bestNo, LastAnchor: 0}
}

func (ctx *SyncContext) SetAncestor(ancestor *Block) {
	ctx.CommonAncestor = ancestor
	ctx.TotalCnt = ctx.TargetNo - ctx.CommonAncestor.BlockNo()
	ctx.RemainCnt = ctx.TotalCnt
}

// NodeInfo is used for actor message to send block info
type BlockInfo struct {
	Hash []byte
	No   BlockNo
}

func (bi *BlockInfo) Equal(target *BlockInfo) bool {
	if target == nil {
		return false
	}

	if bi.No == target.No && bytes.Equal(bi.Hash, target.Hash) {
		return true
	} else {
		return false
	}
}

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
func NewBlock(prevBlock *Block, blockRoot []byte, receipts Receipts, txs []*Tx, coinbaseAcc []byte, ts int64) *Block {
	var prevBlockHash []byte
	var blockNo BlockNo

	if prevBlock != nil {
		prevBlockHash = prevBlock.BlockHash()
		blockNo = prevBlock.Header.BlockNo + 1
	}

	body := BlockBody{
		Txs: txs,
	}
	header := BlockHeader{
		PrevBlockHash:   prevBlockHash,
		BlockNo:         blockNo,
		Timestamp:       ts,
		BlocksRootHash:  blockRoot,
		CoinbaseAccount: coinbaseAcc,
	}
	block := Block{
		Header: &header,
		Body:   &body,
	}

	block.Header.TxsRootHash = CalculateTxsRootHash(body.Txs)
	block.Header.ReceiptsRootHash = receipts.MerkleRoot()

	return &block
}

// calculateBlockHash computes sha256 hash of block header.
func (block *Block) calculateBlockHash() []byte {
	digest := sha256.New()
	serializeBH(digest, block.Header)

	return digest.Sum(nil)
}

func serializeStruct(w io.Writer, s interface{}, stopIndex int) error {
	v := reflect.Indirect(reflect.ValueOf(s))

	var i int
	for i = 0; i <= stopIndex; i++ {
		if err := binary.Write(w, binary.LittleEndian, v.Field(i).Interface()); err != nil {
			return err
		}
	}

	return nil
}

func serializeBH(w io.Writer, bh *BlockHeader) error {
	return serializeStruct(w, bh, lastIndexOfBH)
}

func serializeBhForDigest(w io.Writer, bh *BlockHeader) error {
	return serializeStruct(w, bh, lastIndexOfBH-1)
}

func writeBlockHeaderOld(w io.Writer, bh *BlockHeader) error {
	for _, f := range []interface{}{
		bh.PrevBlockHash,
		bh.BlockNo,
		bh.Timestamp,
		bh.BlocksRootHash,
		bh.TxsRootHash,
		bh.ReceiptsRootHash,
		bh.Confirms,
		bh.PubKey,
		bh.Sign,
	} {
		if err := binary.Write(w, binary.LittleEndian, f); err != nil {
			return err
		}
	}

	return nil
}

// BlockHash returns block hash. It returns a calculated value if the hash is nil.
func (block *Block) BlockHash() []byte {
	hash := block.GetHash()
	if len(hash) == 0 {
		block.Hash = block.calculateBlockHash()
	}

	return block.GetHash()
}

// BlockID converts block hash ([]byte) to BlockID.
func (block *Block) BlockID() BlockID {
	return ToBlockID(block.BlockHash())
}

// PrevBlockID converts parent block hash ([]byte) to BlockID.
func (block *Block) PrevBlockID() BlockID {
	return ToBlockID(block.GetHeader().GetPrevBlockHash())
}

// Confirms returns block.Header.Confirms which indicates how many block is confirmed
// by block.
func (block *Block) Confirms() BlockNo {
	return block.GetHeader().GetConfirms()
}

// SetConfirms sets block.Header.Confirms to confirms.
func (block *Block) SetConfirms(confirms BlockNo) {
	block.Header.Confirms = confirms
}

// BlockNo returns the block number of block.
func (block *Block) BlockNo() BlockNo {
	return block.GetHeader().GetBlockNo()
}

// Sign adds a pubkey and a block signature to block.
func (block *Block) Sign(privKey crypto.PrivKey) error {
	var err error

	if err = block.setPubKey(privKey.GetPublic()); err != nil {
		return err
	}

	var msg []byte
	if msg, err = block.Header.bytesForDigest(); err != nil {
		return err
	}

	var sig []byte
	if sig, err = privKey.Sign(msg); err != nil {
		return err
	}
	block.Header.Sign = sig

	return nil
}

func (bh *BlockHeader) bytesForDigest() ([]byte, error) {
	var buf bytes.Buffer

	if err := serializeBhForDigest(&buf, bh); err != nil {
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
	if msg, err = block.Header.bytesForDigest(); err != nil {
		return false, err
	}

	if valid, err = pubKey.Verify(msg, block.Header.Sign); err != nil {
		return
	}

	return valid, nil
}

// BPID returns its Block Producer's ID from block.
func (block *Block) BPID() (id peer.ID, err error) {
	var pubKey crypto.PubKey
	if pubKey, err = crypto.UnmarshalPublicKey(block.Header.PubKey); err != nil {
		return peer.ID(""), err
	}

	if id, err = peer.IDFromPublicKey(pubKey); err != nil {
		return peer.ID(""), err
	}

	return
}

// BpID2Str returns its Block Producer's ID in base64 format.
func (block *Block) BPID2Str() string {
	id, err := block.BPID()
	if err != nil {
		return ""
	}

	return enc.ToString([]byte(id))
}

// ID returns the base64 encoded formated ID (hash) of block.
func (block *Block) ID() string {
	hash := block.BlockHash()
	if hash != nil {
		return enc.ToString(hash)
	}

	return ""

}

// PrevID returns the base64 encoded formated ID (hash) of the parent block.
func (block *Block) PrevID() string {
	hash := block.GetHeader().GetPrevBlockHash()
	if hash != nil {
		return enc.ToString(hash)
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

func (block *Block) SetBlocksRootHash(blockRootHash []byte) {
	block.GetHeader().BlocksRootHash = blockRootHash
}

// CalculateTxsRootHash generates merkle tree of transactions and returns root hash.
func CalculateTxsRootHash(txs []*Tx) []byte {
	mes := make([]merkle.MerkleEntry, len(txs))
	for i, tx := range txs {
		mes[i] = tx
	}
	return merkle.CalculateMerkleRoot(mes)
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
	digest.Write(txBody.Amount)
	digest.Write(txBody.Payload)
	binary.Write(digest, binary.LittleEndian, txBody.Limit)
	digest.Write(txBody.Price)
	binary.Write(digest, binary.LittleEndian, txBody.Type)
	digest.Write(txBody.Sign)
	return digest.Sum(nil)
}

func (tx *Tx) Validate() error {
	account := tx.GetBody().GetAccount()
	if account == nil {
		return ErrTxFormatInvalid
	}

	if !bytes.Equal(tx.Hash, tx.CalculateTxHash()) {
		return ErrTxHasInvalidHash
	}
	amount := tx.GetBody().GetAmountBigInt()
	if amount.Cmp(MaxAER) > 0 {
		return ErrInsufficientBalance
	}
	/*
		MaxAER is bigger than max of uint64
		if tx.GetBody().GetLimit() > MaxAER {
			return ErrInsufficientBalance
		}
	*/

	price := tx.GetBody().GetPriceBigInt()
	if price.Cmp(MaxAER) > 0 {
		return ErrInsufficientBalance
	}

	switch tx.Body.Type {
	case TxType_NORMAL:
		if tx.GetBody().GetRecipient() == nil && len(tx.GetBody().GetPayload()) == 0 {
			//contract deploy
			return ErrTxInvalidRecipient
		}
	case TxType_GOVERNANCE:
		if len(tx.Body.Payload) <= 0 {
			return ErrTxFormatInvalid
		}
		if (tx.GetBody().GetPayload()[0] == 's' || tx.GetBody().GetPayload()[0] == 'u') &&
			amount.Cmp(StakingMinimum) < 0 {
			return ErrTooSmallAmount
		}
	default:
		return ErrTxInvalidType
	}
	return nil
}

func (tx *Tx) ValidateWithSenderState(senderState *State, coinbaseFee uint64) error {
	if (senderState.GetNonce() + 1) > tx.GetBody().GetNonce() {
		return ErrTxNonceTooLow
	}
	amount := tx.GetBody().GetAmountBigInt()
	balance := senderState.GetBalanceBigInt()
	switch tx.GetBody().GetType() {
	case TxType_NORMAL:
		fee := new(big.Int).SetUint64(coinbaseFee)
		spending := new(big.Int).Add(amount, fee)
		if spending.Cmp(balance) > 0 {
			return ErrInsufficientBalance
		}
	case TxType_GOVERNANCE:
		if string(tx.GetBody().GetRecipient()) == AergoSystem {
			if tx.GetBody().GetPayload()[0] == 's' &&
				amount.Cmp(balance) > 0 {
				return ErrInsufficientBalance
			}
		} else {
			return ErrTxInvalidRecipient
		}
	}
	if (senderState.GetNonce() + 1) < tx.GetBody().GetNonce() {
		return ErrTxNonceToohigh
	}
	return nil
}

//TODO : refoctor after ContractState move to types
func (tx *Tx) ValidateWithContractState(contractState *State) error {
	//in system.ValidateSystemTx
	return nil
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

func (b *TxBody) GetAmountBigInt() *big.Int {
	return new(big.Int).SetBytes(b.GetAmount())
}

func (b *TxBody) GetPriceBigInt() *big.Int {
	return new(big.Int).SetBytes(b.GetPrice())
}

type MovingAverage struct {
	values []int64
	size   int
	count  int

	sum        int64
	removedVal int64
	curPos     int
}

func NewMovingAverage(size int) *MovingAverage {
	return &MovingAverage{
		values:     make([]int64, size),
		size:       size,
		removedVal: 0,
		sum:        0,
		curPos:     -1,
		count:      0,
	}
}

func (ma *MovingAverage) Add(val int64) int64 {
	ma.curPos = (ma.curPos + 1) % ma.size
	ma.removedVal = ma.values[ma.curPos]
	ma.values[ma.curPos] = val

	if ma.count != ma.size {
		ma.count++
	}

	return ma.calculateAvg()
}

func (ma *MovingAverage) calculateAvg() int64 {
	//values is empty
	if ma.count == 0 {
		return 0
	}

	ma.sum = ma.sum - ma.removedVal + ma.values[ma.curPos]

	// Finalize average and return
	avg := ma.sum / int64(ma.count)
	return avg
}
