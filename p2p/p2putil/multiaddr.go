package p2putil

import (
	"context"
	"errors"
	"github.com/aergoio/aergo/v2/types"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multiaddr-dns"
	"math/rand"
)

// ResolveMultiAddress resolve address of dns name to ip address,
// It is needed to connect tcp stream since tcp protocol requires ip4 or ip6 after 2021
func ResolveMultiAddress(peerAddr types.Multiaddr) ([]multiaddr.Multiaddr, error) {
	resolver, err2 := madns.NewResolver()
	if err2 != nil {
		return nil, errors.New("Failed to get dns resolver: " + err2.Error())
	}
	resolved, err2 := resolver.Resolve(context.Background(), peerAddr)
	if err2 != nil {
		return nil, errors.New("Failed to resolve peer address: " + err2.Error())
	}
	return resolved, nil
}

// ResolveToBestIp4Address resolve multi address and then choose one with ip4
func ResolveToBestIp4Address(peerAddr types.Multiaddr) (multiaddr.Multiaddr, error) {
	addrs, err := ResolveMultiAddress(peerAddr)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errors.New("no valid ip addresses")
	}
	return addrs[rand.Intn(len(addrs))], nil
}
