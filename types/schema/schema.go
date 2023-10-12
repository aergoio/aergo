// package schema contains a key prefix collection of low level database accessors.
package schema

// chain
const (
	ReceiptsPrefix = "r"
)

// metadata
const (
	ChainDBName    = "chain"
	Genesis        = ChainDBName + ".genesisInfo"
	GenesisBalance = ChainDBName + ".genesisBalance"
	Latest         = ChainDBName + ".latest"
	HardFork       = "hardfork"
	ReOrg          = "_reorg_marker_"

	// dpos
	DposLibStatus = "dpos.LibStatus" // LibStatusKey is the key when a LIB information is put into the chain DB.

	// raft
	RaftPrefix             = "r_"
	RaftIdentity           = RaftPrefix + "identity"
	RaftState              = RaftPrefix + "state"
	RaftSnap               = RaftPrefix + "snap"
	RaftEntryLastIdx       = RaftPrefix + "last"
	RaftEntry              = RaftPrefix + "entry."
	RaftEntryInvert        = RaftPrefix + "inv."
	RaftConfChangeProgress = RaftPrefix + "ccstatus."
)

// governance
const (
	EnterpriseAdmins = "ADMINS"
	EnterpriseConf   = "conf\\"

	Name = "name"

	SystemParam        = "param\\"
	SystemProposal     = "proposal"
	SystemStaking      = "staking"
	SystemStakingTotal = "stakingtotal"
	SystemVote         = "vote"
	SystemVoteTotal    = "total"
	SystemVoteSort     = "sort"
	SystemVpr          = "VotingPowerBucket/"

	CreatorMeta = "Creator"
)
