/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package rpc

import (
	"context"
	"encoding/pem"
	"strings"

	"github.com/aergoio/aergo/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Authentication = int

const (
	ReadBlockChain  Authentication = 1
	WriteBlockChain Authentication = 2
	ShowNode        Authentication = 4
	ControlNode     Authentication = 8
)

func (rpc *AergoRPCService) setClientAuth(m map[string]Authentication) {
	rpc.clientAuthLock.Lock()
	defer rpc.clientAuthLock.Unlock()

	rpc.clientAuth = m
}

func (rpc *AergoRPCService) checkAuth(ctx context.Context, auth Authentication) error {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no peer found")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return status.Error(codes.Unauthenticated, "unexpected peer transport credentials")
	}

	if len(tlsAuth.State.VerifiedChains) == 0 || len(tlsAuth.State.VerifiedChains[0]) == 0 {
		return status.Error(codes.Unauthenticated, "could not verify peer certificate")
	}

	rpc.clientAuthLock.RLock()
	defer rpc.clientAuthLock.RUnlock()

	for _, id := range tlsAuth.State.PeerCertificates {
		if (rpc.clientAuth[string(id.Raw)] & auth) != 0 {
			return nil
		}
	}

	return status.Error(codes.Unauthenticated, "permission forbidden")
}

func parseConf(conf *types.EnterpriseConfig) map[string]Authentication {
	ret := map[string]Authentication{}
	if !conf.GetOn() {
		return nil
	}
	for _, c := range conf.GetValues() {
		v := strings.Split(c, ":")
		if len(v) != 2 {
			logger.Warn().Str("value", c).Msg("invalid rpc client config")
			continue
		}

		permission := 0
		if strings.Contains(v[1], "R") {
			permission |= ReadBlockChain
		}
		if strings.Contains(v[1], "W") {
			permission |= WriteBlockChain
		}
		if strings.Contains(v[1], "C") {
			permission |= ControlNode
		}
		if strings.Contains(v[1], "S") {
			permission |= ShowNode
		}
		block, err := pem.Decode([]byte(v[0]))
		if err != nil {
			logger.Warn().Str("value", c).Msg("invalid pem format in rpc client config")
			continue
		}
		ret[string(block.Bytes)] = permission
	}
	return ret
}
