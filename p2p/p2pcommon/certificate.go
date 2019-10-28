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
	ErrMalformedCert    = errors.New("malformed certificate data")
	ErrInvalidCertField = errors.New("invalid field in certificate ")
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
	// CreateCertificate create certificate for the agent. It will return ErrInvalidRole error if local peer is not block producer
	CreateCertificate(remoteMeta PeerMeta) (*AgentCertificateV1, error)

	// methods for agents
	// GetProducers return list of peer id of which this agent is charge.
	GetProducers() []types.PeerID
	// GetCertificates returns my certificates
	GetCertificates() []*AgentCertificateV1
	// AddCertificate add to my certificate list
	AddCertificate(cert *AgentCertificateV1)
}
//go:generate mockgen -source=certificate.go -package=p2pmock -destination=../p2pmock/mock_certificate.go
