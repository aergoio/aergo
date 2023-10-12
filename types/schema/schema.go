// package schema contains a key prefix collection of low level database accessors.
package schema

// chain
const (
	// BlockHeaderPrefix    = "h" // headerPrefix + num (uint64 big endian) + hash -> header
	// BlockNumByHashPrefix = "n" // blockNumberPrefix + hash -> num (uint64 big endian)
	// BlockBodyPrefix      = "b" // blockBodyPrefix + num (uint64 big endian) + hash -> block body
	// BlockReceiptPrefix   = "r" // blockReceiptsPrefix + num (uint64 big endian) + hash -> block receipts
	// txLookupPrefix = "t" // txLookupPrefix + hash -> transaction/receipt lookup metadata
	ReceiptsPrefix = "r"
)

// metadata
const (
	ChainDBName       = "chain"
	GenesisKey        = ChainDBName + ".genesisInfo"
	GenesisBalanceKey = ChainDBName + ".genesisBalance"
	LatestKey         = ChainDBName + ".latest"
	HardForkKey       = "hardfork"
	ReOrgKey          = "_reorg_marker_"

	// dpos
	DposLibStatusKey = "dpos.LibStatus" // LibStatusKey is the key when a LIB information is put into the chain DB.

	// raft
	RaftPrefix                   = "r_"
	RaftIdentityKey              = RaftPrefix + "identity"
	RaftStateKey                 = RaftPrefix + "state"
	RaftSnapKey                  = RaftPrefix + "snap"
	RaftEntryLastIdxKey          = RaftPrefix + "last"
	RaftEntryPrefix              = RaftPrefix + "entry."
	RaftEntryInvertPrefix        = RaftPrefix + "inv."
	RaftConfChangeProgressPrefix = RaftPrefix + "ccstatus."
)

// contract
const (
	EnterpriseAdmins     = "ADMINS"
	EnterpriseConfPrefix = "conf\\"

	NamePrefix = "name"

	SystemParamPrefix  = "param\\"
	SystemProposal     = "proposal"
	SystemStaking      = "staking"
	SystemStakingTotal = "stakingtotal"
	SystemVote         = "vote"
	SystemVoteTotal    = "total"
	SystemVoteSort     = "sort"
	SystemVpr          = "VotingPowerBucket/"

	CreatorMetaKey = "Creator"
)

// state
const (
// codePrefix = "c" // CodePrefix + code hash -> account code
// TriePrefix = "s"
// TrieAccountPrefix = "A"
// TrieStoragePrefix = "O"
)
