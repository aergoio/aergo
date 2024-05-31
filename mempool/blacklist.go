package mempool

import (
	"sync"

	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
)

type blacklistConf struct {
	sync.RWMutex
	sourcelist []string        // account address (b58 encoded like Am...) or id (32 bytes in hex = 64 bytes)
	blocked    map[string]bool // all above converted to account id (32 bytes)
	mp         *MemPool        // for log
}

func newBlacklistConf(mp *MemPool, addresses []string) *blacklistConf {
	out := &blacklistConf{}
	out.mp = mp
	out.SetBlacklist(addresses)
	return out
}

func (b *blacklistConf) GetBlacklist() []string {
	if b == nil {
		return []string{}
	}
	b.RLock()
	defer b.RUnlock()
	ret := make([]string, len(b.sourcelist))
	copy(ret, b.sourcelist)
	return ret
}

func (b *blacklistConf) SetBlacklist(addresses []string) {
	b.Lock()
	defer b.Unlock()
	b.sourcelist = make([]string, len(addresses))
	copy(b.sourcelist, addresses)
	b.blocked = make(map[string]bool)
	for _, v := range addresses {
		b.mp.Debug().Str("address", v).Msg("set account blacklist")
		key, err := toKey(v)
		if err == nil {
			b.blocked[key] = true
		} else {
			b.mp.Debug().Str("address", v).Msg("invalid account address or id for blacklist")
		}
	}
}

func (b *blacklistConf) Check(address string) bool {
	if b == nil {
		return false
	}
	key, err := toKey(address)
	if err != nil {
		return false
	}
	b.RLock()
	defer b.RUnlock()
	return b.blocked[key]
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
