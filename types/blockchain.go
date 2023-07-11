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

	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/internal/merkle"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/minio/sha256-simd"
)

const (
	// DefaultMaxBlockSize is the maximum block size (currently 1MiB)
	DefaultMaxBlockSize = 1 << 20
	DefaultEvictPeriod  = 12

	// DefaultMaxHdrSize is the max size of the proto-buf serialized non-body
	// fields. For the estimation detail, check 'TestBlockHeaderLimit' in
	// 'blockchain_test.go.' Caution: Be sure to adjust the value below if the
	// structure of the header is changed.
	DefaultMaxHdrSize = 400
	lastFieldOfBH     = "Consensus"
)

type TxHash = []byte
type AvgTime struct {
	val  atomic.Value
	mavg *MovingAverage
}

var (
	DefaultVerifierCnt = int(math.Max(float64(runtime.NumCPU()/2), float64(1)))
	DefaultAvgTimeSize = 60 * 60 * 24
	AvgTxVerifyTime    = NewAvgTime(DefaultAvgTimeSize)
	// MaxAER is maximum value of aergo
	MaxAER *big.Int
	// StakingMinimum is minimum amount for staking
	StakingMinimum *big.Int
	// ProposalPrice is default value of creating proposal
	ProposalPrice *big.Int
	lastIndexOfBH int
)

func init() {
	MaxAER = NewAmount(5*1e8, Aergo)       // 500,000,000 aergo
	StakingMinimum = NewAmount(1e4, Aergo) // 10,000 aergo
	ProposalPrice = NewZeroAmount()        // 0 aergo
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

//go:generate stringer -type=SystemValue
type SystemValue int

const (
	StakingTotal SystemValue = 0 + iota
	StakingMin
	GasPrice
	NamePrice
	TotalVotingPower
	VotingReward
)

/*
func (s SystemValue) String() string {
	return [...]string{"StakingTotal", "StakingMin", "GasPrice", "NamePrice"}[s]
}
*/

// ChainAccessor is an interface for a another actor module to get info of chain
type ChainAccessor interface {
	GetGenesisInfo() *Genesis
	GetConsensusInfo() string
	GetBestBlock() (*Block, error)
	// GetBlock return block of blockHash. It return nil and error if not found block of that hash or there is a problem in db store
	GetBlock(blockHash []byte) (*Block, error)
	// GetHashByNo returns hash of block. It return nil and error if not found block of that number or there is a problem in db store
	GetHashByNo(blockNo BlockNo) ([]byte, error)
	GetChainStats() string
	GetSystemValue(key SystemValue) (*big.Int, error)

	// GetEnterpriseConfig always return non-nil object if there is no error, but it can return EnterpriseConfig with empty values
	GetEnterpriseConfig(key string) (*EnterpriseConfig, error)
	ChainID(bno BlockNo) *ChainID
}

type SyncContext struct {
	Seq uint64

	PeerID PeerID

	BestNo   BlockNo
	TargetNo BlockNo //sync target blockno

	CommonAncestor *Block

	TotalCnt   uint64
	RemainCnt  uint64
	LastAnchor BlockNo

	NotifyC chan error
}

func NewSyncCtx(seq uint64, peerID PeerID, targetNo uint64, bestNo uint64, notifyC chan error) *SyncContext {
	return &SyncContext{Seq: seq, PeerID: peerID, TargetNo: targetNo, BestNo: bestNo, LastAnchor: 0, NotifyC: notifyC}
}

func (ctx *SyncContext) SetAncestor(ancestor *Block) {
	ctx.CommonAncestor = ancestor
	ctx.TotalCnt = ctx.TargetNo - ctx.CommonAncestor.BlockNo()
	ctx.RemainCnt = ctx.TotalCnt
}

// BlockInfo is used for actor message to send block info
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

func Uint64ToBytes(num uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, num)
	return buf
}

func BytesToUint64(data []byte) uint64 {
	buf := binary.LittleEndian.Uint64(data)
	return buf
}

// BlockNoToBytes represents to serialize block no to bytes
func BlockNoToBytes(bn BlockNo) []byte {
	return Uint64ToBytes(bn)
}

// BlockNoFromBytes represents to deserialize bytes to block no
func BlockNoFromBytes(raw []byte) BlockNo {
	buf := BytesToUint64(raw)
	return BlockNo(buf)
}

type BlockVersionner interface {
	Version(no BlockNo) int32
	IsV2Fork(BlockNo) bool
}

type DummyBlockVersionner int32

func (v DummyBlockVersionner) Version(BlockNo) int32 {
	return int32(v)
}

func (v DummyBlockVersionner) IsV2Fork(BlockNo) bool {
	return (v >= 2)
}

// NewBlock represents to create a block to store transactions.
func NewBlock(bi *BlockHeaderInfo, blockRoot []byte, receipts *Receipts, txs []*Tx, coinbaseAcc []byte, consensus []byte) *Block {
	return &Block{
		Header: &BlockHeader{
			ChainID:          bi.ChainId,
			PrevBlockHash:    bi.PrevBlockHash,
			BlockNo:          bi.No,
			Timestamp:        bi.Ts,
			BlocksRootHash:   blockRoot,
			TxsRootHash:      CalculateTxsRootHash(txs),
			ReceiptsRootHash: receipts.MerkleRoot(),
			CoinbaseAccount:  coinbaseAcc,
			Consensus:        consensus,
		},
		Body: &BlockBody{
			Txs: txs,
		},
	}
}

// Localtime retrurns a time.Time object, which is coverted from block
// timestamp.
func (block *Block) Localtime() time.Time {
	return time.Unix(0, block.GetHeader().GetTimestamp())
}

// calculateBlockHash computes sha256 hash of block header.
func (block *Block) calculateBlockHash() []byte {
	digest := sha256.New()
	serializeBH(digest, block.Header)

	return digest.Sum(nil)
}

func serializeStructOmit(w io.Writer, s interface{}, stopIndex int, omit string) error {
	v := reflect.Indirect(reflect.ValueOf(s))

	var i int
	for i = 0; i <= stopIndex; i++ {
		if v.Type().Field(i).Name == omit {
			continue
		}
		if err := binary.Write(w, binary.LittleEndian, v.Field(i).Interface()); err != nil {
			return err
		}
	}

	return nil
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
	return serializeStructOmit(w, bh, lastIndexOfBH, "Sign")
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
		bh.Consensus,
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

// SetChainID sets id to block.ChainID
func (block *Block) SetChainID(id []byte) {
	block.Header.ChainID = id
}

// ValidChildOf reports whether block is a varid child of parent.
func (block *Block) ValidChildOf(parent *Block) bool {
	parChainID := parent.GetHeader().GetChainID()
	curChainID := block.GetHeader().GetChainID()

	// empty chain id case: an older verion of block has no chain id in its
	// block header.
	if len(parChainID) == 0 && len(curChainID) == 0 {
		return true
	}

	return ChainIdEqualWithoutVersion(parChainID, curChainID)
}

// Size returns a block size where the tx size is individually calculated. A
// similar method is used to limit the block size by the block factory.
//
// THE REASON WHY THE BLOCK FACTORY DOESN'T USE THE EXACT SIZE OF A MARSHALED
// BLOCK: The actual size of a marshaled block is larger than this because it
// includes an additional data associated with the marshaling of the
// transations array in the block body. It is ineffective that the (DPoS) block
// factory measures the exact size of the additional probuf data when it
// produces a block. Thus we use the slightly(?) different and less expensive
// estimation of the block size.
func (block *Block) Size() int {
	size := proto.Size(block.GetHeader()) + len(block.GetHash())
	for _, tx := range block.GetBody().GetTxs() {
		size += proto.Size(tx)
	}
	return size
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
func (block *Block) BPID() (id PeerID, err error) {
	var pubKey crypto.PubKey
	if pubKey, err = crypto.UnmarshalPublicKey(block.Header.PubKey); err != nil {
		return PeerID(""), err
	}

	if id, err = IDFromPublicKey(pubKey); err != nil {
		return PeerID(""), err
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

// GetMetadata generates Metadata object for block
func (block *Block) GetMetadata() *BlockMetadata {
	return &BlockMetadata{
		Hash:    block.BlockHash(),
		Header:  block.GetHeader(),
		Txcount: int32(len(block.GetBody().GetTxs())),
		Size:    int64(proto.Size(block)),
	}
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
	binary.Write(digest, binary.LittleEndian, txBody.GasLimit)
	digest.Write(txBody.GasPrice)
	binary.Write(digest, binary.LittleEndian, txBody.Type)
	digest.Write(txBody.ChainIdHash)
	digest.Write(txBody.Sign)
	return digest.Sum(nil)
}

func (tx *Tx) NeedNameVerify() bool {
	return tx.HasNameAccount()
}

func (tx *Tx) HasNameAccount() bool {
	return len(tx.Body.Account) <= NameLength
}

func (tx *Tx) HasNameRecipient() bool {
	return tx.Body.Recipient != nil && len(tx.Body.Recipient) <= NameLength
}

func (tx *Tx) Clone() *Tx {
	if tx == nil {
		return nil
	}
	if tx.Body == nil {
		return &Tx{}
	}
	body := &TxBody{
		Nonce:       tx.Body.Nonce,
		Account:     Clone(tx.Body.Account).([]byte),
		Recipient:   Clone(tx.Body.Recipient).([]byte),
		Amount:      Clone(tx.Body.Amount).([]byte),
		Payload:     Clone(tx.Body.Payload).([]byte),
		GasLimit:    tx.Body.GasLimit,
		GasPrice:    Clone(tx.Body.GasPrice).([]byte),
		Type:        tx.Body.Type,
		ChainIdHash: Clone(tx.Body.ChainIdHash).([]byte),
		Sign:        Clone(tx.Body.Sign).([]byte),
	}
	res := &Tx{
		Body: body,
	}
	res.Hash = res.CalculateTxHash()
	return res
}

func (b *TxBody) GetAmountBigInt() *big.Int {
	return new(big.Int).SetBytes(b.GetAmount())
}

func (b *TxBody) GetGasPriceBigInt() *big.Int {
	return new(big.Int).SetBytes(b.GetGasPrice())
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

type BlockHeaderInfo struct {
	No            BlockNo
	Ts            int64
	PrevBlockHash []byte
	ChainId       []byte
	ForkVersion   int32
}

var EmptyBlockHeaderInfo = &BlockHeaderInfo{}

func NewBlockHeaderInfo(b *Block) *BlockHeaderInfo {
	cid := b.GetHeader().GetChainID()
	v := DecodeChainIdVersion(cid)
	return &BlockHeaderInfo{
		b.BlockNo(),
		b.GetHeader().GetTimestamp(),
		b.GetHeader().GetPrevBlockHash(),
		cid,
		v,
	}
}

func MakeChainId(cid []byte, v int32) []byte {
	nv := ChainIdVersion(v)
	if bytes.Equal(cid[:4], nv) {
		return cid
	}
	newCid := make([]byte, len(cid))
	copy(newCid, nv)
	copy(newCid[4:], cid[4:])
	return newCid
}

func NewBlockHeaderInfoFromPrevBlock(prev *Block, ts int64, bv BlockVersionner) *BlockHeaderInfo {
	no := prev.GetHeader().GetBlockNo() + 1
	cid := prev.GetHeader().GetChainID()
	v := bv.Version(no)
	return &BlockHeaderInfo{
		no,
		ts,
		prev.GetHash(),
		MakeChainId(cid, v),
		v,
	}
}

func (b *BlockHeaderInfo) ChainIdHash() []byte {
	return common.Hasher(b.ChainId)
}

// HasFunction returns if a function with the given name exists in the ABI definition
func (abi *ABI) HasFunction(name string) bool {
	for _, fn := range abi.Functions {
		if fn.GetName() == name {
			return true
		}
	}
	return false
}
