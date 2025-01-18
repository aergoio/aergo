package p2pcommon

import (
	"errors"
	"time"

	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

var (
	ErrInvalidCertVersion = errors.New("invalid certificate version")
	ErrInvalidRole        = errors.New("invalid peer role") // receiver is not bp or requester is not registered agent
	ErrInvalidKey         = errors.New("invalid key in certificate ")
	ErrInvalidPeerID      = errors.New("invalid peer id in certificate ")
	ErrVerificationFailed = errors.New("signature verification failed")
	ErrMalformedCert      = errors.New("malformed certificate data")
	ErrInvalidCertField   = errors.New("invalid field in certificate ")
)

const (
	CertVersion0001 uint32 = 0x01
)

const (
	timeErrorTolerance = time.Minute
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
	Signature    *ecdsa.Signature
}

// IsValidInTime check if this certificate is expired
func (c *AgentCertificateV1) IsValidInTime(t time.Time, errTolerance time.Duration) bool {
	return (c.CreateTime.Sub(t) < errTolerance) && t.Before(c.ExpireTime)
}

// IsNeedUpdate check if this certificate need to be renewed.
func (c *AgentCertificateV1) IsNeedUpdate(t time.Time, bufTerm time.Duration) bool {
	// bufTerm is expected to be huge with respect to time error, so that no need to consider time error.
	return c.ExpireTime.Sub(t) < bufTerm
}

// CertificateManager manages local peer's certificates and related information
type CertificateManager interface {
	PeerEventListener

	Start()
	Stop()

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

	CanHandle(bpID types.PeerID) bool
}

//go:generate sh -c "mockgen github.com/aergoio/aergo/v2/p2p/p2pcommon CertificateManager | sed -e 's/^package mock_p2pcommon/package p2pmock/g' > ../p2pmock/mock_certificate.go"
