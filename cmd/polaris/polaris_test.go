package main

import (
	"testing"

	"github.com/aergoio/aergo/v2/cmd/polaris/common"
	"github.com/aergoio/aergo/v2/config"
)

func Test_arrangeDefaultCfgForPolaris(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"TNormal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg = config.NewServerContext("/", "/").GetDefaultConfig().(*config.Config)
			if cfg.RPC.NetServicePort != 7845 {
				t.Errorf("Assumption failure: default cfg.RPC.NetServicePort = %d, want %d", cfg.RPC.NetServicePort, 7845)
			}
			if cfg.P2P.NetProtocolPort != 7846 {
				t.Errorf("Assumption failure: default cfg.P2P.NetProtocolPort = %d, want %d", cfg.P2P.NetProtocolPort, 7846)
			}
			arrangeDefaultCfgForPolaris(cfg)
			if cfg.RPC.NetServicePort != common.DefaultRPCPort {
				t.Errorf("cfg.RPC.NetServicePort = %d, want %d", cfg.RPC.NetServicePort, common.DefaultRPCPort)
			}
			if cfg.P2P.NetProtocolPort != common.DefaultSrvPort {
				t.Errorf("cfg.P2P.NetProtocolPort = %d, want %d", cfg.P2P.NetProtocolPort, common.DefaultSrvPort)
			}
		})
	}
}
