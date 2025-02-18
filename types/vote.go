package types

import (
	"bytes"
	"fmt"
	"math/big"
)

const (
	AergoSystem     = "aergo.system"
	AergoName       = "aergo.name"
	AergoEnterprise = "aergo.enterprise"
	AergoVault      = "aergo.vault" // For community reward program (i.e. voting reward)

	MaxCandidates = 30
)

// too few accounts to use map
var specialAccounts [][]byte

func init() {
	specialAccounts = make([][]byte, 0, 4)
	specialAccounts = append(specialAccounts, []byte(AergoSystem))
	specialAccounts = append(specialAccounts, []byte(AergoName))
	specialAccounts = append(specialAccounts, []byte(AergoEnterprise))
	specialAccounts = append(specialAccounts, []byte(AergoVault))
}

// IsSpecialAccount check if name is the one of special account names.
func IsSpecialAccount(name []byte) bool {
	for _, b := range specialAccounts {
		if bytes.Compare(name, b) == 0 {
			return true
		}
	}
	return false
}

type VotingIssue interface {
	ID() string
	Key() []byte
}

// OpSysTx represents a kind of a system transaction.
//
//go:generate stringer -type=OpSysTx
type OpSysTx int

const (
	// For compatibility with the old version, in which the first character of
	// each voting type is lower, the constant name does not follow go naming
	// convertion.

	// OpvoteBP corresponds to a voting transaction for a BP election.
	OpvoteBP OpSysTx = iota
	// OpvoteDAO corresponds to a proposal transaction for a system parameter change.
	OpvoteDAO
	// Opstake represents a staking transaction.
	Opstake
	// Opunstake represents the unstaking transaction.
	Opunstake
	// OpSysTxMax is the maximum of system tx OP numbers.
	OpSysTxMax

	version = 1
)

var cmdToOp map[string]OpSysTx

func initSysCmd() {
	cmdToOp = make(map[string]OpSysTx, OpSysTxMax)
	for i := OpvoteBP; i < OpSysTxMax; i++ {
		cmdToOp[i.Cmd()] = i
	}
}

func init() {
	initSysCmd()
}

// GetVotingIssues returns all the VotingIssues in this package.
func GetVotingIssues() []VotingIssue {
	return []VotingIssue{OpvoteBP}
}

// GetOpSysTx returns a OpSysTx value corresponding to vName.
func GetOpSysTx(vName string) OpSysTx {
	return cmdToOp[vName]
}

// Name returns the unprefixed name corresponding to op.
func (op OpSysTx) ID() string {
	const prefixLen = 2 // prefix = "Op"

	if op < OpSysTxMax && op >= 0 {
		return op.String()[prefixLen:]
	}
	return ""
}

// Cmd returns a string representation for op.
func (op OpSysTx) Cmd() string {
	name := op.ID()
	if len(name) == 0 {
		return name
	}
	return fmt.Sprintf("v%d%s", version, name)
}

func (op OpSysTx) Key() []byte {
	return []byte(op.ID())
}

func (vl VoteList) Len() int { return len(vl.Votes) }
func (vl VoteList) Less(i, j int) bool {
	result := new(big.Int).SetBytes(vl.Votes[i].Amount).Cmp(new(big.Int).SetBytes(vl.Votes[j].Amount))
	if result == -1 {
		return true
	} else if result == 0 {
		if len(vl.Votes[i].Candidate) == 39 /*peer id length*/ {
			return new(big.Int).SetBytes(vl.Votes[i].Candidate[7:]).Cmp(new(big.Int).SetBytes(vl.Votes[j].Candidate[7:])) > 0
		}
		return new(big.Int).SetBytes(vl.Votes[i].Candidate).Cmp(new(big.Int).SetBytes(vl.Votes[j].Candidate)) > 0
	}
	return false
}
func (vl VoteList) Swap(i, j int) { vl.Votes[i], vl.Votes[j] = vl.Votes[j], vl.Votes[i] }

func (v *Vote) GetAmountBigInt() *big.Int {
	return new(big.Int).SetBytes(v.Amount)
}
