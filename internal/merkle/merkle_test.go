package merkle

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	maxAccount   = 2
	maxRecipient = 2
)

var (
	accs      [maxAccount][]byte
	sign      [maxAccount]*btcec.PrivateKey
	recipient [maxRecipient][]byte
	txs       []*types.Tx
)

func TestTXs(t *testing.T) {
}

func _itobU32(argv uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, argv)
	return bs
}

func getAccount(tx *types.Tx) string {
	return hex.EncodeToString(tx.GetBody().GetAccount())
}

func beforeTest(txCount int) error {
	for i := 0; i < maxAccount; i++ {
		privkey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			return err
		}
		//gen new address
		accs[i] = key.GenerateAddress(&privkey.PublicKey)
		sign[i] = privkey
		recipient[i] = _itobU32(uint32(i))
	}

	txCountPerAcc := txCount / maxAccount
	txs = make([]*types.Tx, 0, txCount)

	// gen Tx
	nonce := make([]uint64, txCountPerAcc)
	for i := 0; i < txCountPerAcc; i++ {
		nonce[i] = uint64(i + 1)
	}
	for i := 0; i < maxAccount; i++ {
		for j := 0; j < txCountPerAcc; j++ {
			tmp := genTx(i, j%maxAccount, nonce[j], uint64(i+1))
			txs = append(txs, tmp)
		}
	}

	return nil
}

func afterTest() {

}

func genTx(acc int, rec int, nonce uint64, amount uint64) *types.Tx {
	tx := types.Tx{
		Body: &types.TxBody{
			Nonce:     nonce,
			Account:   accs[acc],
			Recipient: recipient[rec],
			Amount:    amount,
		},
	}
	//tx.Hash = tx.CalculateTxHash()
	key.SignTx(&tx, sign[acc])
	return &tx
}

func TestMerkle0Tx(t *testing.T) {
	t.Log("TestMerkle1Tx")

	testTxs := make([]*types.Tx, 0)

	merkles := GetMerkleTree(testTxs)

	nilHash := make([]byte, 32)

	assert.Equal(t, len(merkles), 1)
	assert.Equal(t, len(merkles[0]), 32)

	assert.Equal(t, bytes.Compare(merkles[0], nilHash), 0)
}

func TestMerkle1Tx(t *testing.T) {
	t.Log("TestMerkle1Tx")
	beforeTest(2)

	testTxs := txs[:1]
	t.Logf("lentxs %d", len(testTxs))
	merkles := GetMerkleTree(testTxs)

	assert.Equal(t, len(merkles), 1)
	assert.Equal(t, 0, bytes.Compare(merkles[0], testTxs[0].GetHash()))
}

func TestMerkle2Tx(t *testing.T) {
	t.Log("TestMerkle1Tx")
	beforeTest(2)

	testTxs := txs[:2]

	merkles := GetMerkleTree(testTxs)

	totalCount := 3
	assert.Equal(t, len(merkles), totalCount)
	for i, tx := range testTxs {
		assert.True(t, bytes.Equal(merkles[i], tx.GetHash()))
	}

	assert.NotNil(t, merkles[2])
	assert.NotEqual(t, merkles[0], merkles[2])

	for i, merkle := range merkles {
		assert.Equal(t, len(merkle), 32)
		t.Logf("%d:%v", i, types.EncodeB64(merkle))
	}
}

func TestMerkle3Tx(t *testing.T) {
	t.Log("TestMerkle1Tx")
	beforeTest(4)

	testTxs := txs[:3]
	t.Logf("lentxs %d", len(testTxs))
	merkles := GetMerkleTree(testTxs)

	totalCount := 7
	assert.Equal(t, len(merkles), totalCount)

	for i, tx := range testTxs {
		assert.True(t, bytes.Equal(merkles[i], tx.GetHash()))
	}

	assert.True(t, bytes.Equal(merkles[2], merkles[3]))

	for i, merkle := range merkles {
		assert.NotNil(t, merkle, "nil=%d", i)
		assert.Equal(t, len(merkle), 32)
		t.Logf("%d:%v", i, types.EncodeB64(merkle))
	}
}

func TestMerkle32Tx(t *testing.T) {
	t.Log("TestMerkle1Tx")
	beforeTest(32)

	testTxs := txs[:10]
	t.Logf("lentxs %d", len(testTxs))
	merkles := GetMerkleTree(testTxs)

	//totalCount := 10240
	totalCount := 31
	assert.Equal(t, len(merkles), totalCount)

	for i, tx := range testTxs {
		assert.True(t, bytes.Equal(merkles[i], tx.GetHash()))
	}

	//copy lc to rc
	// 21, 27
	assert.True(t, bytes.Equal(merkles[20], merkles[21]))
	assert.True(t, bytes.Equal(merkles[26], merkles[27]))

	for i := 10; i <= 15; i++ {
		assert.Nil(t, merkles[i])
	}
	assert.Nil(t, merkles[22])
	assert.Nil(t, merkles[23])

	for i, merkle := range merkles {
		if merkle == nil {
			t.Logf("node:%d is nil", i)
			continue
		}
		assert.Equal(t, len(merkle), 32)
		//t.Logf("%d:%v", i, types.EncodeB64(merkle))
	}

	merkleRoot := merkles[len(merkles)-1]
	assert.NotNil(t, merkleRoot)
}

func BenchmarkMerkle32Tx(b *testing.B) {
	b.Log("BenchmarkMerkle32Tx")
	beforeTest(10000)

	b.ResetTimer()

	testTxs := txs[:10000]

	for i := 0; i < b.N; i++ {
		GetMerkleTree(testTxs)
	}
}
