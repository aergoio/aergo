package merkle

import (
	"bytes"

	"github.com/minio/sha256-simd"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/assert"

	"encoding/base64"
	"encoding/binary"
	"hash"
	"testing"
)

var (
	tms []MerkleEntry
)

func EncodeB64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}

func EncodeB58(bs []byte) string {
	return base58.Encode(bs)
}

func beforeTest(count int) error {
	tms = make([]MerkleEntry, count)

	h := sha256.New()
	for i := 0; i < count; i++ {
		tms[i] = genMerkleEntry(h, i)
	}

	return nil
}

func afterTest() {

}

type testME struct {
	hash []byte
}

func (me *testME) GetHash() []byte {
	return me.hash
}

func genMerkleEntry(h hash.Hash, i int) MerkleEntry {
	tm := testME{}

	h.Reset()
	binary.Write(h, binary.LittleEndian, i)
	tm.hash = h.Sum(nil)

	return &tm
}

func TestMerkle0Tx(t *testing.T) {
	t.Log("TestMerkle1Tx")

	beforeTest(0)

	merkles := CalculateMerkleTree(tms)

	nilHash := make([]byte, 32)

	assert.Equal(t, len(merkles), 1)
	assert.Equal(t, len(merkles[0]), 32)

	assert.Equal(t, bytes.Compare(merkles[0], nilHash), 0)
}

func TestMerkle1Tx(t *testing.T) {
	t.Log("TestMerkle1Tx")
	beforeTest(1)

	t.Logf("lentxs %d", len(tms))
	merkles := CalculateMerkleTree(tms)

	assert.Equal(t, len(merkles), 1)
	assert.Equal(t, 0, bytes.Compare(merkles[0], tms[0].GetHash()))
}

func TestMerkle2Tx(t *testing.T) {
	t.Log("TestMerkle2Tx")
	beforeTest(2)

	merkles := CalculateMerkleTree(tms)

	totalCount := 3
	assert.Equal(t, len(merkles), totalCount)
	for i, tm := range tms {
		assert.True(t, bytes.Equal(merkles[i], tm.GetHash()))
	}

	assert.NotNil(t, merkles[2])
	assert.NotEqual(t, merkles[0], merkles[2])

	for i, merkle := range merkles {
		assert.Equal(t, len(merkle), 32)
		t.Logf("%d:%v", i, EncodeB64(merkle))
	}
}

func TestMerkle3Tx(t *testing.T) {
	t.Log("TestMerkle3Tx")
	beforeTest(3)

	t.Logf("lentxs %d", len(tms))
	merkles := CalculateMerkleTree(tms)

	totalCount := 7
	assert.Equal(t, len(merkles), totalCount)

	for i, tm := range tms {
		assert.True(t, bytes.Equal(merkles[i], tm.GetHash()))
	}

	assert.True(t, bytes.Equal(merkles[2], merkles[3]))

	for i, merkle := range merkles {
		assert.NotNil(t, merkle, "nil=%d", i)
		assert.Equal(t, len(merkle), 32)
		t.Logf("%d:%v", i, EncodeB64(merkle))
	}
}

func TestMerkle32Tx(t *testing.T) {
	t.Log("TestMerkle32Tx")
	beforeTest(10)

	t.Logf("lentxs %d", len(tms))
	merkles := CalculateMerkleTree(tms)

	//totalCount := 10240
	totalCount := 31
	assert.Equal(t, len(merkles), totalCount)

	for i, tm := range tms {
		assert.True(t, bytes.Equal(merkles[i], tm.GetHash()))
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
		//t.Logf("%d:%v", i, EncodeB64(merkle))
	}

	merkleRoot := merkles[len(merkles)-1]
	assert.NotNil(t, merkleRoot)
}

func BenchmarkMerkle10000Tx(b *testing.B) {
	b.Log("BenchmarkMerkle10000Tx")
	beforeTest(10000)
	b.Logf("lentxs %d", len(tms))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CalculateMerkleTree(tms)
	}
}
