package log

import (
	"github.com/sirupsen/logrus"
)

type Module int

const (
	ASVR        Module = iota // asvr
	Consensus                 // consensus
	DPOS                      // dpos
	SBP                       // sbp
	AccountsSvc               // accounts_svc
	ChainSvc                  // chain_svc
	MemPoolSvc                // mempool_svc
	P2PSvc                    // p2p_svc
	RPC                       // rpc
	StateDB                   // state_db
	TEST                      // test
	Rest                      // Rest
	moduleEnd                 // placeholder. DO NOT EDIT
)

//go:generate stringer -linecomment -type=Module

var moduleLevels [moduleEnd]logrus.Level

func init() {
	for i := Module(0); i < moduleEnd; i++ {
		moduleLevels[i] = defaultLevel
	}
}

func SetModuleLevels(levels *Levels) {
	n, err := logrus.ParseLevel(levels.Default)
	if err == nil {
		defaultLevel = n
	}
	for i := Module(0); i < moduleEnd; i++ {
		if moduleLevel, ok := levels.Module[i.String()]; ok {
			setLevel(i, moduleLevel)
		} else {
			setLevel(i, defaultLevel.String())
		}
	}
	setLoggerLevels()
}

func setLevel(module Module, level string) {
	if !module.isValid() {
		return
	}
	if l, e := logrus.ParseLevel(level); e == nil {
		moduleLevels[module] = l
	} else {
		moduleLevels[module] = defaultLevel
	}
}

func (m Module) isValid() bool {
	return m >= Module(0) || m < moduleEnd
}
