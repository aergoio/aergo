package fee

import (
	"fmt"
	"math/big"
)

const defaultFixedTxFee = "1000000000"

var fixedTxFee *big.Int

func SetUserTxFee(fee string) error {
	fixedTxFee, _ := new(big.Int).SetString(fee, 10)
	if fixedTxFee == nil || fixedTxFee.Sign() == -1 {
		return fmt.Errorf("fail to set the fixed transaction fee: %s", fee)
	}
	return nil
}

func SetFixedTxFee(pubNet bool) {
	if pubNet {
		fixedTxFee, _ = new(big.Int).SetString(defaultFixedTxFee, 10)
	} else {
		if fixedTxFee == nil {
			fixedTxFee = big.NewInt(0)
		}
	}
}

func FixedTxFee() *big.Int {
	return fixedTxFee
}
