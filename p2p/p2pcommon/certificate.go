package p2pcommon

import (
	"errors"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"time"
)

var (
	ErrInvalidCertVersion = errors.New("invalid certificate version")
	ErrInvalidKey = errors.New("invalid key in certificate ")
	ErrInvalidPeerID = errors.New("invalid peer id in certificate ")
	ErrVerificationFailed = errors.New("signature verification failed")
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

type CertificateManager interface {
	GetProducers() []types.PeerID
	GetCertificates() []*types.AgentCertificate
	AddCertificate(cert*types.AgentCertificate)
}