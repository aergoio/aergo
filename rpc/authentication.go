/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */
package rpc

import (
	"context"
	"strings"

	"github.com/aergoio/aergo/v2/types"
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

func (rpc *AergoRPCService) setClientAuth(conf *types.EnterpriseConfig) {
	rpc.clientAuthLock.Lock()
	defer rpc.clientAuthLock.Unlock()

	rpc.clientAuth, rpc.clientAuthOn = parseConf(conf)
}

func (rpc *AergoRPCService) setClientAuthOn(v bool) {
	rpc.clientAuthLock.Lock()
	defer rpc.clientAuthLock.Unlock()

	logger.Info().Bool("value", v).Msg("rpc permission check")
	rpc.clientAuthOn = v
}

func (rpc *AergoRPCService) setClientAuthMap(v []string) {
	rpc.clientAuthLock.Lock()

	for _, perm := range v {
		key, value := parseValue(perm)
		rpc.clientAuth[key] = value
	}
	rpc.clientAuthLock.Unlock()
}

func (rpc *AergoRPCService) checkAuth(ctx context.Context, auth Authentication) error {
	rpc.clientAuthLock.RLock()
	defer rpc.clientAuthLock.RUnlock()

	if !rpc.clientAuthOn || len(rpc.clientAuth) == 0 {
		return nil
	}

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

	for _, id := range tlsAuth.State.PeerCertificates {
		key := types.EncodeB64(id.Raw)
		if (rpc.clientAuth[key] & auth) != 0 {
			return nil
		}
	}

	return status.Error(codes.Unauthenticated, "permission forbidden")
}

func parseValue(perm string) (string, int) {
	v := strings.Split(perm, ":")
	if len(v) != 2 {
		logger.Warn().Str("value", perm).Msg("invalid rpc client config")
		return "", 0
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
	return v[0], permission
}

func parseConf(conf *types.EnterpriseConfig) (map[string]Authentication, bool) {
	ret := map[string]Authentication{}

	for _, perm := range conf.GetValues() {
		key, value := parseValue(perm)
		ret[key] = value
	}

	return ret, conf.GetOn()
}
