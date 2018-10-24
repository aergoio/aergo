/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package config

import (
	"runtime"

	"github.com/aergoio/aergo-lib/config"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"
)

type ServerContext struct {
	config.BaseContext
}

func NewServerContext(homePath string, configFilePath string) *ServerContext {
	serverCxt := &ServerContext{}
	serverCxt.BaseContext = config.NewBaseContext(serverCxt, homePath, configFilePath, EnvironmentPrefix)

	return serverCxt
}

func (ctx *ServerContext) GetHomePath() string {
	return defaultAergoHomePath
}

func (ctx *ServerContext) GetConfigFileName() string {
	return defaultAergoConfigFileName
}

func (ctx *ServerContext) GetTemplate() string {
	return tomlConfigFileTemplate
}

func (ctx *ServerContext) GetDefaultConfig() interface{} {
	return &Config{
		BaseConfig: ctx.GetDefaultBaseConfig(),
		RPC:        ctx.GetDefaultRPCConfig(),
		REST:       ctx.GetDefaultRESTConfig(),
		P2P:        ctx.GetDefaultP2PConfig(),
		Blockchain: ctx.GetDefaultBlockchainConfig(),
		Mempool:    ctx.GetDefaultMempoolConfig(),
		Consensus:  ctx.GetDefaultConsensusConfig(),
	}
}

func (ctx *ServerContext) GetDefaultBaseConfig() BaseConfig {
	return BaseConfig{
		DataDir:        ctx.ExpandPathEnv("$HOME/data"),
		EnableProfile:  false,
		ProfilePort:    6060,
		EnableRest:     false,
		EnableTestmode: false,
	}
}

func (ctx *ServerContext) GetDefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		NetServiceAddr: "127.0.0.1",
		NetServicePort: 7845,
		NSKey:          "",
	}
}

func (ctx *ServerContext) GetDefaultRESTConfig() *RESTConfig {
	return &RESTConfig{
		RestPort: 8080,
	}
}

func (ctx *ServerContext) GetDefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		NetProtocolAddr: "0.0.0.0",
		NetProtocolPort: 7846,
		NPEnableTLS:     false,
		NPCert:          "",
		NPKey:           "",
		NPAddPeers:      nil,
		NPMaxPeers:      100,
		NPPeerPool:      100,
	}
}

func (ctx *ServerContext) GetDefaultBlockchainConfig() *BlockchainConfig {
	return &BlockchainConfig{
		MaxBlockSize:    types.DefaultMaxBlockSize,
		CoinbaseAccount: "",
	}
}

func (ctx *ServerContext) GetDefaultMempoolConfig() *MempoolConfig {
	return &MempoolConfig{
		ShowMetrics:    false,
		VerifierNumber: runtime.NumCPU(),
		DumpFilePath:   ctx.ExpandPathEnv("$HOME/mempool.dump"),
	}
}

func (ctx *ServerContext) GetDefaultConsensusConfig() *ConsensusConfig {
	return &ConsensusConfig{
		EnableBp:      true,
		BlockInterval: consensus.DefaultBlockIntervalSec,
		DposBpNumber:  consensus.DefaultDposBpNumber,
		BpIds:         nil,
	}
}
