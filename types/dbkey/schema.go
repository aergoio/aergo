// package dbkey contains a key prefix collection of low level database accessors.
package dbkey

// chain
const (
	receiptsPrefix = "r"
	ethAddrPrefix  = "e"
)

// metadata
const (
	ChainDBName = "chain"

	genesis        = ChainDBName + ".genesisInfo"
	genesisBalance = ChainDBName + ".genesisBalance"
	latestBlock    = ChainDBName + ".latest"
	hardFork       = "hardfork"
	reOrg          = "_reorg_marker_"

	// ethereum state db root hash
	ethRoot = "ethRoot"

	// dpos
	dposLibStatus = "dpos.LibStatus" // LibStatusKey is the key when a LIB information is put into the chain DB.

	// raft
	raftPrefix             = "r_"
	raftIdentity           = raftPrefix + "identity"
	raftState              = raftPrefix + "state"
	raftSnap               = raftPrefix + "snap"
	raftEntryLastIdx       = raftPrefix + "last"
	raftEntry              = raftPrefix + "entry."
	raftEntryInvert        = raftPrefix + "inv."
	raftConfChangeProgress = raftPrefix + "ccstatus."
)

// governance
const (
	// enterprise
	enterpriseAdmins = "ADMINS"
	enterpriseConf   = "conf\\"

	// name
	name = "name"

	// system
	systemParam        = "param\\"
	systemProposal     = "proposal"
	systemStaking      = "staking"
	systemStakingTotal = "stakingtotal"
	systemVote         = "vote"
	systemVoteTotal    = "total"
	systemVoteSort     = "sort"
	systemVpr          = "VotingPowerBucket/"

	creatorMeta = "Creator"
)
