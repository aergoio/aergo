package p2p

import (
	"encoding/binary"
	"errors"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/minio/sha256-simd"
	"time"
)

const (
	CertVer_0001 uint32 = 0x01
)
const (
	PeerIDSize = 32
)

var (
	ErrMalformedCert      = errors.New("malformed certificate data")
	ErrInvalidCertField   = errors.New("invalid field in certificate ")
	ErrInvalidCertVersion = errors.New("invalid certificate version")
	ErrInvalidKey         = errors.New("invalid key in certificate ")
	ErrInvalidPeerID      = errors.New("invalid peer id in certificate ")
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

// NewAgentCertV1 create certificate object
func NewAgentCertV1(bpID, agentID types.PeerID, bpKey *btcec.PrivateKey, addrs []string, ttl time.Duration) (*AgentCertificateV1, error) {
	// need to truncate monotonic clock
	now := time.Now().Truncate(0)
	c := &AgentCertificateV1{Version: CertVer_0001, BPID: bpID, BPPubKey: bpKey.PubKey(),
		CreateTime: now, ExpireTime: now.Add(ttl), AgentID: agentID, AgentAddress: addrs}
	err := SignCert(bpKey, c)
	if err != nil {
		return nil, ErrInvalidCertField
	}
	return c, nil
}

func CheckProtoCert(cert *types.AgentCertificate) (*AgentCertificateV1, error) {
	switch cert.CertVersion {
	case CertVer_0001:
		return CheckAndGetV1(cert)
	default:
		return nil, ErrInvalidCertVersion
	}
}

func CheckAndGetV1(cert *types.AgentCertificate) (*AgentCertificateV1, error) {
	var err error

	wrap := &AgentCertificateV1{Version: cert.CertVersion}
	wrap.BPID, err = peer.IDFromBytes(cert.BPID)
	if err != nil {
		return nil, ErrInvalidPeerID
	}
	wrap.BPPubKey, err = btcec.ParsePubKey(cert.BPPubKey, btcec.S256())
	if err != nil {
		return nil, ErrInvalidKey
	}
	libp2pKey := p2putil.ConvertPubToLibP2P(wrap.BPPubKey)
	generatedID, err := peer.IDFromPublicKey(libp2pKey)
	if err != nil {
		return nil, ErrInvalidKey
	}
	if !types.IsSamePeerID(wrap.BPID, generatedID) {
		return nil, ErrInvalidKey
	}

	wrap.CreateTime = time.Unix(0, int64(cert.CreateTime))
	wrap.ExpireTime = time.Unix(0, int64(cert.ExpireTime))
	now := time.Now()
	// TODO consider error of clock if it need more fine checking
	if wrap.CreateTime.After(now) || wrap.ExpireTime.Before(now) {
		return nil, ErrInvalidCertField
	}
	wrap.AgentID, err = peer.IDFromBytes(cert.AgentID)
	if err != nil {
		return nil, ErrInvalidPeerID
	}
	if len(cert.AgentAddress) == 0 {
		return nil, ErrInvalidCertField
	}
	wrap.AgentAddress = make([]string, len(cert.AgentAddress))
	for i, addr := range cert.AgentAddress {
		wrap.AgentAddress[i] = string(addr)
		// TODO check address
	}
	wrap.Signature, err = btcec.ParseSignature(cert.Signature, btcec.S256())
	if err != nil {
		return nil, ErrInvalidCertField
	}

	// verify seq
	if !VerifyCert(wrap) {
		return nil, ErrVerificationFailed
	}
	return wrap, nil
}

func (w *AgentCertificateV1) ToProtoCert() (*types.AgentCertificate, error) {
	var err error

	protoC := &types.AgentCertificate{CertVersion: w.Version}

	protoC.BPID = []byte(w.BPID)
	//if err != nil {
	//	return nil, ErrInvalidCertField
	//}
	protoC.BPPubKey = w.BPPubKey.SerializeCompressed()
	protoC.CreateTime = w.CreateTime.UnixNano()
	protoC.ExpireTime = w.ExpireTime.UnixNano()

	protoC.AgentID = []byte(w.AgentID)
	//if err != nil {
	//	return nil, ErrInvalidCertField
	//}
	if len(w.AgentAddress) == 0 {
		return nil, ErrInvalidCertField
	}
	protoC.AgentAddress = make([][]byte, len(w.AgentAddress))
	for i, addr := range w.AgentAddress {
		protoC.AgentAddress[i] = []byte(addr)
		// TODO check address
	}

	protoC.Signature = w.Signature.Serialize()
	if err != nil {
		return nil, err
	}
	return protoC, nil
}

func SignCert(key *btcec.PrivateKey, wrap *AgentCertificateV1) error {
	hash, err := calculateCertificateHash(wrap)
	if err != nil {
		return err
	}
	sign, err := key.Sign(hash)
	if err != nil {
		return err
	}
	wrap.BPPubKey = key.PubKey()
	wrap.Signature = sign
	return nil
}

func VerifyCert(wrap *AgentCertificateV1) bool {
	hash, err := calculateCertificateHash(wrap)
	if err != nil {
		return false
	} else {
		return wrap.Signature.Verify(hash, wrap.BPPubKey)
	}
}

// version, bpid, bppubkey, create time, expire time, agent id, agent address, signature
func calculateCertificateHash(cert *AgentCertificateV1) ([]byte, error) {
	var err error
	var bArr []byte
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, cert.Version)
	bArr, err = cert.BPID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	h.Write(bArr)
	h.Write(cert.BPPubKey.SerializeCompressed())
	binary.Write(h, binary.LittleEndian, cert.CreateTime.UnixNano())
	binary.Write(h, binary.LittleEndian, cert.ExpireTime.UnixNano())
	bArr, err = cert.AgentID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	h.Write(bArr)
	for _, addr := range cert.AgentAddress {
		h.Write([]byte(addr))
	}
	return h.Sum(nil), nil
}
