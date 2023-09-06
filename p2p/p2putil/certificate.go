package p2putil

import (
	"encoding/binary"
	"time"

	"github.com/aergoio/aergo/v2/internal/network"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/minio/sha256-simd"
)

func ConvertCertToProto(w *p2pcommon.AgentCertificateV1) (*types.AgentCertificate, error) {
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
		return nil, p2pcommon.ErrInvalidCertField
	}
	protoC.AgentAddress = make([][]byte, len(w.AgentAddress))
	for i, addr := range w.AgentAddress {
		protoC.AgentAddress[i] = []byte(addr)
	}

	protoC.Signature = w.Signature.Serialize()
	if err != nil {
		return nil, err
	}
	return protoC, nil
}

func ConvertCertsToProto(cs []*p2pcommon.AgentCertificateV1) ([]*types.AgentCertificate, error) {
	var err error
	ret := make([]*types.AgentCertificate, len(cs))
	for i, c := range cs {
		ret[i], err = ConvertCertToProto(c)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// NewAgentCertV1 create certificate object
func NewAgentCertV1(bpID, agentID types.PeerID, bpKey *btcec.PrivateKey, addrs []string, ttl time.Duration) (*p2pcommon.AgentCertificateV1, error) {
	// need to truncate monotonic clock
	now := time.Now().Truncate(0)
	c := &p2pcommon.AgentCertificateV1{Version: p2pcommon.CertVersion0001, BPID: bpID, BPPubKey: bpKey.PubKey(),
		CreateTime: now, ExpireTime: now.Add(ttl), AgentID: agentID, AgentAddress: addrs}
	err := SignCert(bpKey, c)
	if err != nil {
		return nil, p2pcommon.ErrInvalidCertField
	}
	return c, nil
}

func CheckProtoCert(cert *types.AgentCertificate) (*p2pcommon.AgentCertificateV1, error) {
	switch cert.CertVersion {
	case p2pcommon.CertVersion0001:
		return CheckAndGetV1(cert)
	default:
		return nil, p2pcommon.ErrInvalidCertVersion
	}
}

func CheckAndGetV1(cert *types.AgentCertificate) (*p2pcommon.AgentCertificateV1, error) {
	var err error

	wrap := &p2pcommon.AgentCertificateV1{Version: cert.CertVersion}
	wrap.BPID, err = peer.IDFromBytes(cert.BPID)
	if err != nil {
		return nil, p2pcommon.ErrInvalidPeerID
	}
	wrap.BPPubKey, err = btcec.ParsePubKey(cert.BPPubKey, btcec.S256())
	if err != nil {
		return nil, p2pcommon.ErrInvalidKey
	}
	libp2pKey := ConvertPubToLibP2P(wrap.BPPubKey)
	generatedID, err := peer.IDFromPublicKey(libp2pKey)
	if err != nil {
		return nil, p2pcommon.ErrInvalidKey
	}
	if !types.IsSamePeerID(wrap.BPID, generatedID) {
		return nil, p2pcommon.ErrInvalidKey
	}

	wrap.CreateTime = time.Unix(0, int64(cert.CreateTime))
	wrap.ExpireTime = time.Unix(0, int64(cert.ExpireTime))
	now := time.Now()
	// check certificate is valid
	if !wrap.IsValidInTime(now, p2pcommon.TimeErrorTolerance) {
		return nil, p2pcommon.ErrInvalidCertField
	}
	wrap.AgentID, err = peer.IDFromBytes(cert.AgentID)
	if err != nil {
		return nil, p2pcommon.ErrInvalidPeerID
	}
	if len(cert.AgentAddress) == 0 {
		return nil, p2pcommon.ErrInvalidCertField
	}
	wrap.AgentAddress = make([]string, len(cert.AgentAddress))
	for i, addr := range cert.AgentAddress {
		addrStr, err := network.CheckAddress(string(addr))
		if err != nil {
			return nil, err
		}
		wrap.AgentAddress[i] = addrStr
	}
	wrap.Signature, err = btcec.ParseSignature(cert.Signature, btcec.S256())
	if err != nil {
		return nil, p2pcommon.ErrInvalidCertField
	}

	// verify seq
	if !VerifyCert(wrap) {
		return nil, p2pcommon.ErrVerificationFailed
	}
	return wrap, nil
}

func SignCert(key *btcec.PrivateKey, wrap *p2pcommon.AgentCertificateV1) error {
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

func VerifyCert(wrap *p2pcommon.AgentCertificateV1) bool {
	hash, err := calculateCertificateHash(wrap)
	if err != nil {
		return false
	} else {
		return wrap.Signature.Verify(hash, wrap.BPPubKey)
	}
}

// version, bpid, bppubkey, create time, expire time, agent id, agent address, signature
func calculateCertificateHash(cert *p2pcommon.AgentCertificateV1) ([]byte, error) {
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
