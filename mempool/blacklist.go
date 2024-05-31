package mempool

import (
	"sync"
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
	b.RLock()
	defer b.RUnlock()
	ret := make([]string, len(b.sourcelist))
	copy(ret, b.sourcelist)
	return ret
}

func (b *blacklistConf) SetBlacklist(addresses []string) {
	b.Lock()
	defer b.Unlock()
	b.sourcelist := make([]string, len(addresses))
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
		return true
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
			return nil, err
		}
		key = common.Hasher(addr)
	}
	return string(key), err
}
