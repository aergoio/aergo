package mempool

import (
	"sync"
)

type whitelistConf struct {
	sync.RWMutex
	whitelist map[string]bool
	on        bool
	mp        *MemPool //for log
}

func newWhitelistConf(mp *MemPool, addresses []string, on bool) *whitelistConf {
	out := &whitelistConf{}
	out.mp = mp
	out.on = on
	out.SetWhitelist(addresses)
	return out
}

func (w *whitelistConf) GetWhitelist() []string {
	w.RLock()
	defer w.RUnlock()
	ret := []string{}
	for k := range w.whitelist {
		ret = append(ret, k)
	}
	return ret
}

func (w *whitelistConf) GetOn() bool {
	w.RLock()
	defer w.RUnlock()
	return w.on
}

func (w *whitelistConf) SetWhitelist(addresses []string) {
	w.Lock()
	defer w.Unlock()
	w.whitelist = make(map[string]bool)
	for _, v := range addresses {
		w.whitelist[v] = true
		w.mp.Debug().Str("address", v).Msg("set account white list")
	}
}

func (w *whitelistConf) Check(address string) bool {
	if w == nil {
		return true
	}
	w.RLock()
	defer w.RUnlock()
	if !w.on {
		return true
	}
	return w.whitelist[address]
}

func (w *whitelistConf) Enable(enable bool) {
	w.Lock()
	defer w.Unlock()
	w.on = enable
	w.mp.Debug().Bool("enable", enable).Msg("switch account white list")
}
