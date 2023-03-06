package contract

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
)

type EthState struct {
	StateCache state.Database
	VMcfg      vm.Config
	TempState  *state.StateDB
}

func NewEthState(ethDb ethdb.Database) *EthState {
	return &EthState{
		StateCache: state.NewDatabase(ethDb),
		VMcfg: vm.Config{
			Debug: true,
		},
	}
}

func TestEVM(t *testing.T) {
	ethState := NewEthState(nil)
	if ethState == nil {
		t.Errorf("eth state not created")
	}
	t.Log("testing")
}
