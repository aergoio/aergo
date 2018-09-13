package blockchain

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	maxAccount   = 1000
	maxRecipient = 1000
	maxTx        = 10000
)

var (
	accs      [maxAccount][]byte
	sign      [maxAccount]*btcec.PrivateKey
	recipient [maxRecipient][]byte
	txs       []*types.Tx

	verifier *SignVerifier
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

func beforeTest(accCount int, txCountPerAcc int) error {
	if verifier == nil {
		verifier = NewSignVerifier(DefaultVerifierCnt)
	}

	for i := 0; i < accCount; i++ {
		privkey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			return err
		}
		//gen new address
		accs[i] = key.GenerateAddress(&privkey.PublicKey)
		sign[i] = privkey
		recipient[i] = _itobU32(uint32(i))
	}

	txs = make([]*types.Tx, 0, accCount*txCountPerAcc)

	// gen Tx
	nonce := make([]uint64, txCountPerAcc)
	for i := 0; i < txCountPerAcc; i++ {
		nonce[i] = uint64(i + 1)
	}
	for i := 0; i < accCount; i++ {
		for j := 0; j < txCountPerAcc; j++ {
			tmp := genTx(i, j, nonce[j], uint64(i+1))
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

func TestInvalidTransactions(t *testing.T) {
	t.Log("TestInvalidTransactions")
	beforeTest(10, 1)
	//defer afterTest()

	txslice := make([]*types.Tx, 0)
	tx := genTx(0, 1, 1, 1)
	tx.Body.Amount = 999999

	txslice = append(txslice, tx)

	failed, errors := verifier.VerifyTxs(&types.TxList{Txs: txslice})

	assert.Equal(t, failed, true)

	if failed {
		for i, error := range errors {
			if error != nil {
				assert.Equal(t, i, 0)
				assert.Equal(t, error, message.ErrSignNotMatch)
			}
		}
	}
}

// gen sequential transactions
// bench
func TestVerifyValidTxs(t *testing.T) {
	t.Log("TestVerifyValidTxs")
	beforeTest(3, 3)
	defer afterTest()

	txslice := make([]*types.Tx, 0)
	for _, tx := range txs {
		txslice = append(txslice, tx)
	}

	failed, errors := verifier.VerifyTxs(&types.TxList{Txs: txslice})
	if failed {
		for i, error := range errors {
			if error != nil {
				t.Fatalf("failed tx %d:%s", i, error.Error())
			}
		}
	}
}

func BenchmarkVerify10000tx(b *testing.B) {
	b.Log("BenchmarkVerify10000tx")
	beforeTest(100, 100)
	defer afterTest()

	txslice := make([]*types.Tx, 0)
	for _, tx := range txs {
		txslice = append(txslice, tx)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		failed, errors := verifier.VerifyTxs(&types.TxList{Txs: txslice})
		if failed {
			for i, error := range errors {
				if error != nil {
					b.Errorf("failed tx %d:%s", i, error.Error())
				}
			}
		}
	}
}

func BenchmarkVerify10000txSerial(b *testing.B) {
	b.Log("BenchmarkVerify10000txSerial")
	beforeTest(100, 100)
	defer afterTest()

	txslice := make([]*types.Tx, 0)
	for _, tx := range txs {
		txslice = append(txslice, tx)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		failed, errors := verifier.verifyTxsInplace(&types.TxList{Txs: txslice})
		if failed {
			for i, error := range errors {
				if error != nil {
					b.Errorf("failed tx %d:%s", i, error.Error())
				}
			}
		}
	}
}
