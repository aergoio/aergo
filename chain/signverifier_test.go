package chain

import (
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/aergoio/aergo/v2/account/key"
	crypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
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

	verifier *SignVerifier
)

func TestTXs(t *testing.T) {
}

func _itobU32(argv uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, argv)
	return bs
}

func beforeTest(txCount int) error {
	if verifier == nil {
		verifier = NewSignVerifier(nil /*types.DefaultVerifierCnt*/, nil, 4, false)
	}

	for i := 0; i < maxAccount; i++ {
		privkey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}
		//gen new address
		accs[i] = crypto.GenerateAddress(privkey.PubKey().ToECDSA())
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
			Amount:    new(big.Int).SetUint64(amount).Bytes(),
		},
	}
	//tx.Hash = tx.CalculateTxHash()
	key.SignTx(&tx, sign[acc])
	return &tx
}

func TestInvalidTransactions(t *testing.T) {
	t.Log("TestInvalidTransactions")
	beforeTest(10)
	//defer afterTest()

	txslice := make([]*types.Tx, 0)
	tx := genTx(0, 1, 1, 1)
	tx.Body.Amount = new(big.Int).SetUint64(999999).Bytes()

	txslice = append(txslice, tx)

	verifier.RequestVerifyTxs(&types.TxList{Txs: txslice})
	failed, errs := verifier.WaitDone()

	assert.Equal(t, failed, true)

	if failed {
		for i, err := range errs {
			if err != nil {
				assert.Equal(t, i, 0)
				assert.Equal(t, err, types.ErrSignNotMatch)
			}
		}
	}
}

// gen sequential transactions
// bench
func TestVerifyValidTxs(t *testing.T) {
	t.Log("TestVerifyValidTxs")
	beforeTest(100)
	defer afterTest()

	t.Logf("len=%d", len(txs))

	verifier.RequestVerifyTxs(&types.TxList{Txs: txs})
	failed, errs := verifier.WaitDone()

	if failed {
		for i, err := range errs {
			if err != nil {
				t.Fatalf("failed tx %d:%s", i, err.Error())
			}
		}
	}
}

func BenchmarkVerify10000tx(b *testing.B) {
	b.Log("BenchmarkVerify10000tx")
	beforeTest(10000)
	defer afterTest()

	txslice := make([]*types.Tx, 0)
	for _, tx := range txs {
		txslice = append(txslice, tx)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		verifier.RequestVerifyTxs(&types.TxList{Txs: txslice})
		failed, errs := verifier.WaitDone()

		if failed {
			for i, err := range errs {
				if err != nil {
					b.Errorf("failed tx %d:%s", i, err.Error())
				}
			}
		}
	}
}

func BenchmarkVerify10000txSerial(b *testing.B) {
	b.Log("BenchmarkVerify10000txSerial")
	beforeTest(10000)
	defer afterTest()

	txslice := make([]*types.Tx, 0)
	for _, tx := range txs {
		txslice = append(txslice, tx)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		failed, errs := verifier.verifyTxsInplace(&types.TxList{Txs: txslice})
		if failed {
			for i, err := range errs {
				if err != nil {
					b.Errorf("failed tx %d:%s", i, err.Error())
				}
			}
		}
	}
}
