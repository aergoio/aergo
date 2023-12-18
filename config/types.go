package config

// Config should be read-only in outer world, but golang doesn't have any simple solution for that.
// A Developer MUST NOT modify config value in caller code.
const (
	defaultAergoHomePath       = ".aergo"
	defaultAergoConfigFileName = "config.toml"

	EnvironmentPrefix = "AG"

	//defaultLogFileName = "aergo.log"
)

// Config defines configurations of each services
type Config struct {
	BaseConfig `mapstructure:",squash"`
	RPC        *RPCConfig        `mapstructure:"rpc"`
	P2P        *P2PConfig        `mapstructure:"p2p"`
	Polaris    *PolarisConfig    `mapstructure:"polaris"`
	Blockchain *BlockchainConfig `mapstructure:"blockchain"`
	Mempool    *MempoolConfig    `mapstructure:"mempool"`
	Consensus  *ConsensusConfig  `mapstructure:"consensus"`
	Monitor    *MonitorConfig    `mapstructure:"monitor"`
	Account    *AccountConfig    `mapstructure:"account"`
	Auth       *AuthConfig       `mapstructure:"auth"`
	Hardfork   *HardforkConfig   `mapstructure:"hardfork"`
	SQL        *SQLConfig        `mapstructure:"sql"`
}

// BaseConfig defines base configurations for aergo server
type BaseConfig struct {
	DataDir        string `mapstructure:"datadir" description:"Directory to store datafiles"`
	DbType         string `mapstructure:"dbtype" description:"db implementation to store data"`
	EnableDump     bool   `mapstructure:"enabledump" description:"enable dump feature for debugging"`
	DumpPort       int    `mapstructure:"dumpport" description:"dump port (default:7070)"`
	EnableTestmode bool   `mapstructure:"enabletestmode" description:"enable unsafe test mode"`
	UseTestnet     bool   `mapstructure:"usetestnet" description:"need description"`
	Personal       bool   `mapstructure:"personal" description:"enable personal account service"`
	AuthDir        string `mapstructure:"authdir" description:"Directory to store files for auth"`
}

// RPCConfig defines configurations for rpc service
type RPCConfig struct {
	// RPC and REST
	NetServiceAddr  string `mapstructure:"netserviceaddr" description:"RPC service address"`
	NetServicePort  int    `mapstructure:"netserviceport" description:"RPC service port"`
	NetServicePath  string `mapstructure:"netservicepath" description:"UNIX domain socket path (only for admin commands)"`
	NetServiceTrace bool   `mapstructure:"netservicetrace" description:"Trace RPC service"`
	// RPC API with TLS
	NSEnableTLS bool   `mapstructure:"nstls" description:"Enable TLS on RPC or REST API"`
	NSCert      string `mapstructure:"nscert" description:"Server Certificate file for RPC or REST API"`
	NSKey       string `mapstructure:"nskey" description:"Private Key file for RPC or REST API"`
	NSCACert    string `mapstructure:"nscacert" description:"CA Certificate file for RPC or REST API"`
	NSAllowCORS bool   `mapstructure:"nsallowcors" description:"Allow CORS to RPC or REST API"`
}

// P2PConfig defines configurations for p2p service
type P2PConfig struct {
	// N2N (peer-to-peer) network
	NetProtocolAddr string   `mapstructure:"netprotocoladdr" description:"N2N listen address to which other peer can connect. This address is advertized to other peers."`
	NetProtocolPort int      `mapstructure:"netprotocolport" description:"N2N listen port to which other peer can connect. This port is advertized to other peers."`
	NPBindAddr      string   `mapstructure:"npbindaddr" description:"N2N bind address. If it was set, it only accept connection to this addresse only"`
	NPBindPort      int      `mapstructure:"npbindport" description:"N2N bind port. It not set, bind port is same as netprotocolport. Set if server is configured with NAT and port is differ."`
	NPEnableTLS     bool     `mapstructure:"nptls" description:"Enable TLS on N2N network"`
	NPCert          string   `mapstructure:"npcert" description:"Certificate file for N2N network"`
	NPKey           string   `mapstructure:"npkey" description:"Private Key file for N2N network"`
	NPAddPeers      []string `mapstructure:"npaddpeers" description'':"Add peers to connect to at startup"`
	NPHiddenPeers   []string `mapstructure:"nphiddenpeers" description:"List of peerids which will not show to other peers"`
	NPDiscoverPeers bool     `mapstructure:"npdiscoverpeers" description:"Whether to discover from polaris or other nodes and connects"`
	NPMaxPeers      int      `mapstructure:"npmaxpeers" description:"Maximum number of remote peers to keep"`
	NPPeerPool      int      `mapstructure:"nppeerpool" description:"Max peer pool size"`

	NPExposeSelf   bool     `mapstructure:"npexposeself" description:"Whether to request expose self to polaris and other connected node"`
	NPUsePolaris   bool     `mapstructure:"npusepolaris" description:"Whether to connect and get node list from polaris"`
	NPAddPolarises []string `mapstructure:"npaddpolarises" description:"Add addresses of polarises if default polaris is not sufficient"`

	LogFullPeerID bool `mapstructure:"logfullpeerid" description:"Whether to use full legnth peerID or short form"`

	PeerRole      string   `mapstructure:"peerrole" description:"Role of peer. It must be sync with enablebp field in consensus config "`
	Producers     []string `mapstructure:"producers" description:"List of peer ids of block producers, only meaningful when peer is agent"`
	InternalZones []string `mapstructure:"internalzones" description:"List of address ranges that are recognised as inner zone of agent. defined by CIDR notation."`
	Agent         string   `mapstructure:"agent" description:"Peer id of agent that delegates this producer, only available when local peer is producer"`
	AllowLegacy   bool     `mapstructure:"allowlegacy" description:"Whether to allow legacy security protocols"`
}

// AuthConfig defines configuration for auditing
type AuthConfig struct {
	EnableLocalConf bool `mapstructure:"enablelocalconf" description:"apply local white/blacklist file or not"`
}

// PolarisConfig defines configuration for polaris server and client (i.e. polarisConnect)
type PolarisConfig struct {
	AllowPrivate    bool   `mapstructure:"allowprivate" description:"allow peer to have private address. for private network and test"`
	GenesisFile     string `mapstructure:"genesisfile" description:"json file containing informations of genesisblock to which polaris refer "`
	EnableBlacklist bool   `mapstructure:"enableblacklist" description:"allow peer to have private address. for private network and test"`
}

// BlockchainConfig defines configurations for blockchain service
type BlockchainConfig struct {
	MaxBlockSize     uint32 `mapstructure:"maxblocksize"  description:"maximum block size in bytes"`
	CoinbaseAccount  string `mapstructure:"coinbaseaccount" description:"wallet address for coinbase"`
	MaxAnchorCount   int    `mapstructure:"maxanchorcount" description:"maximum anchor count for sync"`
	VerifierCount    int    `mapstructure:"verifiercount" description:"maximum transaction verifier count"`
	ForceResetHeight uint64 `mapstructure:"forceresetheight" description:"best height to reset chain manually"`
	ZeroFee          bool   `mapstructure:"zerofee" description:"enable zero-fee mode(deprecated)"`
	VerifyOnly       bool   `mapstructure:"verifyonly" description:"In verify only mode, server verifies block chain of disk. server never modifies block chain'"`
	StateTrace       uint64 `mapstructure:"statetrace" description:"dump trace of setting state"`
	VerifyBlock      uint64 `mapstructure:"verifyblock" description:"In verify only mode, server verifies given block of disk. server never modifies block chain'"`
	NumWorkers       int    `mapstructure:"numworkers" description:"maximum worker count for chainservice"`
	NumLStateClosers int    `mapstructure:"numclosers" description:"maximum LuaVM state closer count for chainservice"`
	CloseLimit       int    `mapstructure:"closelimit" description:"number of LuaVM states which a LuaVM state closer closes at one time"`
}

// MempoolConfig defines configurations for mempool service
type MempoolConfig struct {
	ShowMetrics    bool   `mapstructure:"showmetrics" description:"show mempool metric periodically"`
	EnableFadeout  bool   `mapstructure:"enablefadeout" description:"Enable transaction fadeout over timeout period"`
	FadeoutPeriod  int    `mapstructure:"fadeoutperiod" description:"time period for evict transactions(in hour)"`
	VerifierNumber int    `mapstructure:"verifiers" description:"number of concurrent verifier"`
	DumpFilePath   string `mapstructure:"dumpfilepath" description:"file path for recording mempool at process termintation"`
}

// ConsensusConfig defines configurations for consensus service
type ConsensusConfig struct {
	EnableBp            bool        `mapstructure:"enablebp" description:"enable block production"`
	BlockInterval       int64       `mapstructure:"blockinterval" description:"block production interval (sec)"`
	Raft                *RaftConfig `mapstructure:"raft"`
	NoTimeoutTxEviction bool        `mapstructure:"notte" description:"disable timeout tx eviction"`
}

type RaftConfig struct {
	Name               string        `mapstructure:"name" description:"raft node name. this value must be unique in cluster"`
	SkipEmpty          bool          `mapstructure:"skipempty" description:"skip producing block if there is no tx in block"`
	HeartbeatTick      uint          `mapstructure:"heartbeattick" description:"heartbeat tick of raft server (millisec)"`
	ElectionTickCount  int           `mapstructure:"electiontickcount" description:"number of ticks to wait to be a candidate when no heartbeat message comes from leader"`
	BlockFactoryTickMs int64         `mapstructure:"blockfactorytickms" description:"interval to check if block factory should run new task(millisec)"`
	BlockIntervalMs    int64         `mapstructure:"blockintervalms" description:"block interval for raft (millisec). It overrides BlockInterval of consensus"`
	NewCluster         bool          `mapstructure:"newcluster" description:"create a new raft cluster if it doesn't already exist"`
	UseBackup          bool          `mapstructure:"usebackup" description:"use backup datafiles when creating a new cluster or joining to a existing cluster"`
	SnapFrequency      uint64        `mapstructure:"snapfrequency" description:"frequency which raft make snapshot with log"`
	SlowNodeGap        uint          `mapstructure:"slownodegap" description:"frequency which raft make snapshot with log"`
	RecoverBP          *RaftBPConfig `mapstructure:"recoverbp" description:"bp info for creating a new cluster from backup"`
	StopDupCommit      bool          `mapstructure:"stopdupcommit" description:"stop server when commit of duplicate height block occurs. use this only for debugging'"`
}

type RaftBPConfig struct {
	Name    string `mapstructure:"name" description:"raft node name"`
	Address string `mapstructure:"address" description:"raft address"`
	PeerID  string `mapstructure:"peerid" description:"peerid of this bp"`
}

type MonitorConfig struct {
	ServerProtocol string `mapstructure:"protocol" description:"Protocol is one of next: http, https or kafka"`
	ServerEndpoint string `mapstructure:"endpoint" description:"Endpoint to send"`
}

// Account defines configurations for account service
type AccountConfig struct {
	UnlockTimeout uint `mapstructure:"unlocktimeout" description:"lock automatically after timeout (sec)"`
}

type SQLConfig struct {
	MaxDbSize uint64 `mapstructure:"maxdbsize" description:"maximum database size of a contract (MB)"`
}

/*
How to write this template
=======================================

string_type = "{{.STRUCT.FILED}}"
bool/number_type = {{.STRUCT.FILED}}
string_array_type = [{{range .STRUCT.FILED}}
"{{.}}", {{end}}
]
bool/number_array_type = [{{range .STRUCT.FILED}}
{{.}}, {{end}}
]
map = does not support
*/
const tomlConfigFileTemplate = `# aergo TOML Configuration File (https://github.com/toml-lang/toml)
# base configurations
datadir = "{{.BaseConfig.DataDir}}"
dbtype = "{{.BaseConfig.DbType}}"
enabledump = {{.BaseConfig.EnableDump}}
dumpport = {{.BaseConfig.DumpPort}}
personal = {{.BaseConfig.Personal}}
authdir = "{{.BaseConfig.AuthDir}}"

[rpc]
netserviceaddr = "{{.RPC.NetServiceAddr}}"
netserviceport = {{.RPC.NetServicePort}}
netservicetrace = {{.RPC.NetServiceTrace}}
nstls = {{.RPC.NSEnableTLS}}
nscert = "{{.RPC.NSCert}}"
nskey = "{{.RPC.NSKey}}"
nscacert = "{{.RPC.NSCACert}}"
nsallowcors = {{.RPC.NSAllowCORS}}

[p2p]
# Set address and port to which the inbound peers connect, and don't set loopback address or private network unless used in local network 
netprotocoladdr = "{{.P2P.NetProtocolAddr}}"
netprotocolport = {{.P2P.NetProtocolPort}}
npbindaddr = "{{.P2P.NPBindAddr}}"
npbindport = {{.P2P.NPBindPort}}
# TLS and certificate is not applied in alpha release.
nptls = {{.P2P.NPEnableTLS}}
npcert = "{{.P2P.NPCert}}"
# Set file path of key file
npkey = "{{.P2P.NPKey}}"
npaddpeers = [{{range .P2P.NPAddPeers}}
"{{.}}", {{end}}
]
npdiscoverpeers = true
npmaxpeers = "{{.P2P.NPMaxPeers}}"
nppeerpool = "{{.P2P.NPPeerPool}}"
npexposeself = true
npusepolaris = {{.P2P.NPUsePolaris}}
npaddpolarises = [{{range .P2P.NPAddPolarises}}
"{{.}}", {{end}}
]
peerrole = "{{.P2P.PeerRole}}"
allowlegacy = "{{.P2P.AllowLegacy}}"

[polaris]
allowprivate = {{.Polaris.AllowPrivate}}
genesisfile = "{{.Polaris.GenesisFile}}"
enableblacklist = "{{.Polaris.EnableBlacklist}}"

[blockchain]
# blockchain configurations
maxblocksize = {{.Blockchain.MaxBlockSize}}
coinbaseaccount = "{{.Blockchain.CoinbaseAccount}}"
maxanchorcount = "{{.Blockchain.MaxAnchorCount}}"
verifiercount = "{{.Blockchain.VerifierCount}}"
forceresetheight = "{{.Blockchain.ForceResetHeight}}"
numworkers = "{{.Blockchain.NumWorkers}}"
numclosers = "{{.Blockchain.NumLStateClosers}}"
closelimit = "{{.Blockchain.CloseLimit}}"

[mempool]
showmetrics = {{.Mempool.ShowMetrics}}
enablefadeout = {{.Mempool.EnableFadeout}}
fadeoutperiod = {{.Mempool.FadeoutPeriod}}
verifiers = {{.Mempool.VerifierNumber}}
dumpfilepath = "{{.Mempool.DumpFilePath}}"

[consensus]
enablebp = {{.Consensus.EnableBp}}
blockinterval = {{.Consensus.BlockInterval}}

[monitor]
protocol = "{{.Monitor.ServerProtocol}}"
endpoint = "{{.Monitor.ServerEndpoint}}"

[account]
unlocktimeout = "{{.Account.UnlockTimeout}}"

[auth]
enablelocalconf = "{{.Auth.EnableLocalConf}}"

` + hardforkConfigTmpl
