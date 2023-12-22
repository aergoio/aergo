package jsonrpc

import (
	"math/big"

	"github.com/aergoio/aergo/v2/types"
)

func ConvAccount(msg *types.Account) *InOutAccount {
	if msg == nil {
		return nil
	}

	a := &InOutAccount{}
	a.Address = types.EncodeAddress(msg.GetAddress())

	return a
}

type InOutAccount struct {
	Address string `json:"address,omitempty"`
}

func ConvAccounts(msg *types.AccountList) *InOutAccountList {
	if msg == nil {
		return nil
	}

	al := &InOutAccountList{}
	al.Accounts = make([]*InOutAccount, len(msg.Accounts))
	for i, account := range msg.Accounts {
		al.Accounts[i] = ConvAccount(account)
	}

	return al
}

type InOutAccountList struct {
	Accounts []*InOutAccount `json:"accounts,omitempty"`
}

func ConvState(msg *types.State) *InOutState {
	if msg == nil {
		return nil
	}

	s := &InOutState{}
	s.Nonce = msg.GetNonce()
	s.Balance = new(big.Int).SetBytes(msg.GetBalance()).String()

	return s
}

type InOutState struct {
	Nonce       uint64 `json:"nonce"`
	Balance     string `json:"balance"`
	Account     string `json:"account,omitempty"`
	Stake       string `json:"stake,omitempty"`
	VotingPower string `json:"votingpower,omitempty"`
	NextAction  uint64 `json:"nextaction,omitempty"`
}

func ConvStateAndPoof(msg *types.AccountProof) *InOutStateAndPoof {
	if msg == nil {
		return nil
	}

	ap := &InOutStateAndPoof{}
	ap.Nonce = msg.GetState().GetNonce()
	ap.Balance = new(big.Int).SetBytes(msg.State.GetBalance()).String()
	ap.Included = msg.GetInclusion()
	ap.MerkleProofLength = len(msg.GetAuditPath())
	ap.Height = msg.GetHeight()

	return ap
}

type InOutStateAndPoof struct {
	Nonce             uint64 `json:"nonce"`
	Balance           string `json:"balance"`
	Account           string `json:"account,omitempty"`
	Included          bool   `json:"included,omitempty"`
	MerkleProofLength int    `json:"merkle proof length,omitempty"`
	Height            uint32 `json:"height,omitempty"`
}

func ConvNameInfo(msg *types.NameInfo) *InOutNameInfo {
	if msg == nil {
		return nil
	}

	ni := &InOutNameInfo{}
	ni.Name = msg.Name.Name
	ni.Owner = types.EncodeAddress(msg.Owner)
	ni.Destination = types.EncodeAddress(msg.Destination)

	return ni
}

type InOutNameInfo struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Destination string `json:"destination"`
}

func ConvBalance(msg *types.State) *InOutBalance {
	if msg == nil {
		return nil
	}

	b := &InOutBalance{}
	state := ConvState(msg)
	b.Balance = state.Balance

	return b
}

type InOutBalance struct {
	Balance string `json:"balance"`
}
