package blacklist

import (
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
)

type Blacklist struct {
	sourcelist []string        // account address (b58 encoded like Am...) or id (32 bytes in hex = 64 bytes)
	blocked    map[string]bool // all above converted to account id (32 bytes)
}

var globalBlacklist *Blacklist

// Initialize sets up the blacklist with the given addresses.
// This function should be called only once at the start.
func Initialize(addresses []string) {
	conf := &Blacklist{}
	conf.sourcelist = make([]string, len(addresses))
	copy(conf.sourcelist, addresses)
	conf.blocked = make(map[string]bool)
	for _, v := range addresses {
		key, err := toKey(v)
		if err == nil {
			conf.blocked[key] = true
		} else {
			// Handle invalid address, log or take other actions as needed
		}
	}
	globalBlacklist = conf
}

func Check(address string) bool {
	if globalBlacklist == nil {
		return false
	}
	key, err := toKey(address)
	if err != nil {
		return false
	}
	return globalBlacklist.blocked[key]
}

func toKey(address string) (string, error) {
	var key []byte
	var err error
	if len(address) == 64 {
		key, err = hex.Decode(address)
	} else {
		var addr []byte
		addr, err = types.DecodeAddress(address)
		if err != nil {
			return "", err
		}
		key = common.Hasher(addr)
	}
	return string(key), err
}
