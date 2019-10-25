package p2pcommon

import (
	"errors"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"time"
)

var (
	ErrInvalidCertVersion = errors.New("invalid certificate version")
	ErrInvalidRole = errors.New("invalid peer role") // receiver is not bp or requester is not registered agent
	ErrInvalidKey = errors.New("invalid key in certificate ")
	ErrInvalidPeerID = errors.New("invalid peer id in certificate ")
	ErrVerificationFailed = errors.New("signature verification failed")
)

const (
	CertVersion0001 uint32 = 0x01
)

// AgentCertificateV1 is a certificate issued by a block producer to guarantee that it is a trustworthy agent.
type AgentCertificateV1 struct {
	Version      uint32
	BPID         types.PeerID
	BPPubKey     *btcec.PublicKey
	CreateTime   time.Time
	ExpireTime   time.Time
	AgentID      types.PeerID
	AgentAddress []string
	Signature    *btcec.Signature
}

func (c *AgentCertificateV1) IsValidInTime(t time.Time) bool {
	// TODO consider the case is time error between peers
	return t.After(c.CreateTime) && t.Before(c.ExpireTime)
}
type CertificateManager interface {
	// methods for bp
	// CreateCertificate create certificate for the agent
	CreateCertificate(remoteMeta PeerMeta) (*AgentCertificateV1, error)

	// methods for agents
	// GetProducers return list of peer id of which this agent is charge.
	GetProducers() []types.PeerID
	GetCertificates() []*AgentCertificateV1
	AddCertificate(cert *AgentCertificateV1)
}
