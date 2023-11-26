package types

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/minio/sha256-simd"
	"github.com/stretchr/testify/assert"
)

func TestValidateInitValue(t *testing.T) {
	// Max Aergo value ( 500,000,000 aergo )
	assert.Equalf(t, "500000000000000000000000000", MaxAER.String(), "MaxAER is not valid. check types/blockchain.go")
	// Staking minimum amount ( 10,000 aergo )
	assert.Equalf(t, "10000000000000000000000", StakingMinimum.String(), "StakingMinimum is not valid. check types/blockchain.go")
	// Proposal price ( 0 aergo )
	assert.Equalf(t, "0", ProposalPrice.String(), "ProposalPrice is not valid. check types/blockchain.go")
}

func TestBlockHash(t *testing.T) {
	blockHash := func(block *Block) []byte {
		header := block.Header
		digest := sha256.New()
		writeBlockHeader(digest, header)
		return digest.Sum(nil)
	}

	txIn := make([]*Tx, 0)
	block := NewBlock(EmptyBlockHeaderInfo, nil, nil, txIn, nil, nil)

	h1 := blockHash(block)
	h2 := block.calculateBlockHash()

	assert.Equal(t, h1, h2)
}

func genKeyPair(assert *assert.Assertions) (crypto.PrivKey, crypto.PubKey) {
	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	assert.Nil(err)

	return privKey, pubKey
}

func genSign(assert *assert.Assertions, block *Block, privKey crypto.PrivKey) []byte {
	msg, err := block.Header.bytesForDigest()
	assert.Nil(err)

	sig, err := privKey.Sign(msg)
	assert.Nil(err)

	return sig
}

func TestBlockSignBasic(t *testing.T) {
	signAssert := assert.New(t)

	message := func(block *Block) []byte {
		msg, err := block.Header.bytesForDigest()
		signAssert.Nil(err)

		return msg
	}

	sign := func(block *Block, privKey crypto.PrivKey) []byte {
		sig, err := privKey.Sign(message(block))
		signAssert.Nil(err)

		return sig
	}

	verify := func(block *Block, sig []byte, pubKey crypto.PubKey) bool {
		valid, err := pubKey.Verify(message(block), sig)
		signAssert.Nil(err)

		return valid
	}

	block := NewBlock(EmptyBlockHeaderInfo, nil, nil, make([]*Tx, 0), nil, nil)

	privKey, pubKey := genKeyPair(signAssert)
	sig := sign(block, privKey)
	valid := verify(block, sig, pubKey)
	signAssert.True(valid)
}

func TestBlockSign(t *testing.T) {
	signAssert := assert.New(t)
	block := NewBlock(EmptyBlockHeaderInfo, nil, nil, make([]*Tx, 0), nil, nil)

	privKey, _ := genKeyPair(signAssert)
	signAssert.Nil(block.Sign(privKey))

	sig := genSign(signAssert, block, privKey)
	signAssert.NotNil(sig)

	signAssert.Equal(sig, block.Header.Sign)

	valid, err := block.VerifySign()
	signAssert.Nil(err)
	signAssert.True(valid)
}

func TestMovingAverage(t *testing.T) {
	size := int64(10)
	mv := NewMovingAverage(int(size))

	assert.Equal(t, int64(0), mv.calculateAvg())

	var sum = int64(0)
	var expAvg int64
	for i := int64(1); i <= size; i++ {
		avg := mv.Add(int64(i))
		assert.Equal(t, i, int64(mv.count))

		sum += i
		expAvg = sum / i
		assert.Equal(t, expAvg, avg)
	}

	for i := int64(size + 1); i <= (size + 10); i++ {
		avg := mv.Add(int64(i))
		assert.Equal(t, size, int64(mv.count))

		sum = sum + i - (i - size)
		expAvg = sum / size
		assert.Equal(t, expAvg, avg)
	}
}

func TestUpdateAvgVerifyTime(t *testing.T) {
	var size = int64(10)
	avgTime := NewAvgTime(int(size))

	val := avgTime.Get()

	assert.Equal(t, int64(0), int64(val))

	var sum = int64(0)
	var expAvg int64

	for i := int64(1); i <= size; i++ {
		avgTime.UpdateAverage(time.Duration(i))
		newAvg := avgTime.Get()

		sum += i
		expAvg = sum / i
		assert.Equal(t, expAvg, int64(newAvg))
	}

	for i := int64(size + 1); i <= (size + 10); i++ {
		avgTime.UpdateAverage(time.Duration(i))
		newAvg := avgTime.Get()

		sum = sum + i - (i - size)
		expAvg = sum / size
		assert.Equal(t, expAvg, int64(newAvg))
	}
}

func TestBlockLimit(t *testing.T) {
	a := assert.New(t)

	chainID := "0123456789012345678901234567890123456789"
	dig := sha256.New()
	dig.Write([]byte(chainID))
	h := dig.Sum(nil)

	addr, err := DecodeAddress("AmLquXjQSiDdR8FTDK78LJ16Ycrq3uNL6NQuiwqXRCGw9Riq2DE4")
	a.Nil(err, "invalid coinbase account")

	_, pub := genKeyPair(a)

	block := &Block{
		Header: &BlockHeader{
			ChainID:          []byte(chainID), // len <= 40
			PrevBlockHash:    h,               // 32byte
			BlockNo:          math.MaxUint64,
			Timestamp:        math.MinInt64,
			BlocksRootHash:   h, // 32byte. Currenly not used but for the future use.
			TxsRootHash:      h, // 32byte
			ReceiptsRootHash: h, // 32byte
			Confirms:         math.MaxUint64,
			CoinbaseAccount:  addr,
		},
		Body: &BlockBody{Txs: make([]*Tx, 0)},
	}
	err = block.setPubKey(pub)
	a.Nil(err, "PubKey set failed")

	// The signature's max length is 72 (strict DER format). But usually it is
	// 70 or 71. For header length measurement, generate a byte array of length
	// 72 from a hash (sha256) value xrather than a real signature.
	block.Header.Sign = h
	block.Header.Sign = append(block.Header.Sign, h...)
	block.Header.Sign = append(block.Header.Sign, h[:8]...)

	// A block can include its sha256 hash value. Thus the size of blcok hash
	// value should be considered when the block header limit estimated.
	fmt.Println("block id:", block.ID())

	hdrSize := proto.Size(block)
	a.True(hdrSize <= DefaultMaxHdrSize, "too large header (> %v): %v", DefaultMaxHdrSize, hdrSize)

	fmt.Println("hdr size", hdrSize)

	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	account, _ := DecodeAddress(testSender)
	amount, ok := new(big.Int).SetString("999999999999999999", 10)
	a.True(ok, "amount conversion failed")

	//blk := &Block{}
	tx := &Tx{
		Body: &TxBody{
			Nonce:     math.MaxUint64,
			Account:   account,
			Recipient: account,
			Amount:    amount.Bytes(),
			Payload:   []byte(`{"Name":"v1unstake"}`),
			GasLimit:  math.MaxUint64,
			GasPrice:  amount.Bytes(),
			Sign:      block.GetHeader().GetSign(),
			Type:      TxType_GOVERNANCE,
		},
	}
	tx.Hash = tx.CalculateTxHash()

	txSize := proto.Size(tx)

	nTx := 10000
	var i int

	hdrLimit := 400
	bodyLimit := 1000000
	limit := hdrLimit + bodyLimit
	s := 0
	for i = 1; i <= nTx; i++ {
		if s += proto.Size(tx); s > bodyLimit {
			break
		}
		block.Body.Txs = append(block.Body.Txs, tx)
	}

	fmt.Println("estimate #1: ", block.Size(), "esimate #2: ", txSize*i+hdrSize, "actual: ", proto.Size(block))
	a.True(block.Size() <= proto.Size(block), "block size violation")
	a.True(block.Size() <= txSize*i+hdrSize, "block size violation")
	a.True(block.Size() <= limit, "block size violation")
}
