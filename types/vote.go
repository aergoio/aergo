package types

import (
	"fmt"
	"math/big"
)

const (
	AergoSystem     = "aergo.system"
	AergoName       = "aergo.name"
	AergoEnterprise = "aergo.enterprise"

	MaxCandidates = 30
)

//go:generate stringer -type=OpSysTx
// OpSysTx represents a kind of a system transaction.
type OpSysTx int

const (
	// For compatibility with the old version, in which the first character of
	// each voting type is lower, the constant name does not follow go naming
	// convertion.

	// OpvoteBP corresponds to a voting transaction for a BP election.
	OpvoteBP OpSysTx = iota
	// OpvoteProposal corresponds to a proposal transaction for a system parameter change.
	OpvoteProposal
	// Opstake represents a staking tranaction.
	Opstake
	// Opunstake represents a unstaking tranaction.
	Opunstake
	// OpcreateProposal represents a transaction creating a proposal.
	OpcreateProposal
	// OpSysTxMax is the maximum of system tx OP numbers.
	OpSysTxMax

	version = 1
)

var cmdToOp = make(map[string]OpSysTx, OpSysTxMax)

func init() {
	for i := OpvoteBP; i < OpSysTxMax; i++ {
		cmdToOp[i.Cmd()] = i
	}
}

// GetOpSysTx returns a OpSysTx value corresponding to vName.
func GetOpSysTx(vName string) OpSysTx {
	return cmdToOp[vName]
}

// Name returns a unprefixed name corresponding to op.
func (op OpSysTx) Name() string {
	const prefixLen = 2 // prefix = "Op"

	if op < OpSysTxMax && op >= 0 {
		return op.String()[prefixLen:]
	}
	return ""
}

// Cmd returns a string representation for op.
func (op OpSysTx) Cmd() string {
	name := op.Name()
	if len(name) == 0 {
		return name
	}
	return fmt.Sprintf("v%d%s", version, name)
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
