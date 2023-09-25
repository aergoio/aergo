package types

//go:generate stringer -type=SystemValue
type SystemValue int

const (
	StakingTotal SystemValue = 0 + iota
	StakingMin
	GasPrice
	NamePrice
	TotalVotingPower
	VotingReward
)

/*
func (s SystemValue) String() string {
	return [...]string{"StakingTotal", "StakingMin", "GasPrice", "NamePrice"}[s]
}
*/
