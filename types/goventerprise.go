package types

const (
	SetConf       = "setConf"
	AppendConf    = "appendConf"
	RemoveConf    = "removeConf"
	AppendAdmin   = "appendAdmin"
	RemoveAdmin   = "removeAdmin"
	EnableConf    = "enableConf"
	DisableConf   = "disableConf"
	ChangeCluster = "changeCluster"
)

const (
	RPCPermissions = "RPCPERMISSIONS"
	P2PWhite       = "P2PWHITE"
	P2PBlack       = "P2PBLACK"
	AccountWhite   = "ACCOUNTWHITE"
)

// EnterpriseKeyDict is represent allowed key list and used when validate tx, int values are meaningless.
var EnterpriseKeyDict = map[string]int{
	RPCPermissions: 1,
	P2PWhite:       2,
	P2PBlack:       3,
	AccountWhite:   4,
}
