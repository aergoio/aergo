package types

import "github.com/multiformats/go-multiaddr"

type AddrType int
const (
	InvalidAddrType = 0
	AddrTypeIP4  = multiaddr.P_IP4
	AddrTypeIP6  = multiaddr.P_IP6
	AddrTypeDNS4 = multiaddr.P_DNS4
	AddrTypeDNS6 = multiaddr.P_DNS6
)

type NetAddress struct {
	Type AddrType
	Address string
	Port uint32
}
