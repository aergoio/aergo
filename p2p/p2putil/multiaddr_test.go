package p2putil

import (
	"fmt"
	"github.com/aergoio/aergo/v2/types"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveMultiAddress(t *testing.T) {
	type args struct {
		peerAddr types.Multiaddr
	}
	tests := []struct {
		name        string
		arg         string
		wantEquals  bool
		wantMinSize int
		wantErr     assert.ErrorAssertionFunc
	}{
		{"nope", "/dns4/nowherenothing.aergo.io/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF", false, 0, assert.Error},
		{"dnspolaris", "/dns4/polaris.aergo.io/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF", false, 1, assert.NoError},
		{"ippolaris", "/ip4/3.36.146.156/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF", true, 1, assert.NoError},
		{"dnsNaver", "/dns/naver.com/tcp/443", false, 3, assert.NoError},
		{"dns4Naver", "/dns4/naver.com/tcp/443", false, 3, assert.NoError},
		{"dns6Naver", "/dns6/naver.com/tcp/443", false, 0, assert.NoError},
		{"dnsGoogle", "/dns/google.com/tcp/443", false, 2, assert.NoError},
		{"dns4Google", "/dns4/google.com/tcp/443", false, 1, assert.NoError},
		{"dns6Google", "/dns6/google.com/tcp/443", false, 1, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peerAddr, _ := multiaddr.NewMultiaddr(tt.arg)
			got1, err := ResolveMultiAddress(peerAddr)
			if !tt.wantErr(t, err, fmt.Sprintf("ResolveMultiAddress(%v)", tt.arg)) {
				return
			}
			t.Logf("%v is resolved to %v", peerAddr, got1)
			assert.LessOrEqualf(t, tt.wantMinSize, len(got1), "ResolveMultiAddress(%v) result min size", tt.arg)
			found := false
			for _, ele := range got1 {
				if ele.Equal(peerAddr) {
					found = true
					break
				}
			}
			assert.Equal(t, tt.wantEquals, found)

			//got2, err := ResolveToBestIp4Address(peerAddr)
			//if !tt.wantErr(t, err, fmt.Sprintf("ResolveMultiAddress(%v)", tt.arg)) {
			//	return
			//}

		})
	}
}

func TestResolveToBestIp4Address(t *testing.T) {
	type args struct {
		peerAddr types.Multiaddr
	}
	tests := []struct {
		name       string
		arg        string
		wantEquals bool
		wantErr    assert.ErrorAssertionFunc
	}{
		{"nope", "/dns4/nowherenothing.aergo.io/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF", false, assert.Error},
		{"dnspolaris", "/dns4/polaris.aergo.io/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF", false, assert.NoError},
		{"ippolaris", "/ip4/3.36.146.156/tcp/8916/p2p/16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF", true, assert.NoError},
		{"dnsNaver", "/dns/naver.com/tcp/443", false, assert.NoError},
		{"dns4Naver", "/dns4/naver.com/tcp/443", false, assert.NoError},
		{"dns6Naver", "/dns6/naver.com/tcp/443", false, assert.Error},
		{"dnsGoogle", "/dns/google.com/tcp/443", false, assert.NoError},
		{"dns4Google", "/dns4/google.com/tcp/443", false, assert.NoError},
		{"dns6Google", "/dns6/google.com/tcp/443", false, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peerAddr, _ := multiaddr.NewMultiaddr(tt.arg)
			got, err := ResolveToBestIp4Address(peerAddr)
			if !tt.wantErr(t, err, fmt.Sprintf("ResolveToBestIp4Address(%v)", tt.arg)) {
				return
			}
			if err == nil {
				if !assert.NotNil(t, got) {
					return
				}
				t.Logf("%v is best resolved to %v", peerAddr, got)
				assert.Equal(t, tt.wantEquals, got.Equal(peerAddr))
			}
		})
	}
}
