package types

const AergoSystem = "aergo.system"

func (v VoteList) Len() int           { return len(v.Votes) }
func (v VoteList) Less(i, j int) bool { return v.Votes[i].Amount < v.Votes[j].Amount }
func (v VoteList) Swap(i, j int)      { v.Votes[i], v.Votes[j] = v.Votes[j], v.Votes[i] }
