package p2putil

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/golang/protobuf/proto"
)

func TestNewAgentCertV1(t *testing.T) {
	pk, _ := btcec.NewPrivateKey(btcec.S256())
	pid1, pid2 := types.RandomPeerID(), types.RandomPeerID()
	addr0 := "192.168.0.2"
	addr1 := "2001:0db8:85a3:08d3:1319:8a2e:370:7334"
	addr2 := "tester.aergo.io"
	DAY := time.Hour * 24
	type args struct {
		bpID    types.PeerID
		agentID types.PeerID
		bpKey   *btcec.PrivateKey
		addrs   []string
		ttl     time.Duration
	}
	tests := []struct {
		name string
		args args

		wantErr bool
	}{
		{"TSucc", args{pid1, pid2, pk, []string{addr0}, DAY}, false},
		{"TMultiID", args{pid1, pid2, pk, []string{addr0, addr1, addr2}, DAY}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAgentCertV1(tt.args.bpID, tt.args.agentID, tt.args.bpKey, tt.args.addrs, tt.args.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAgentCertV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !types.IsSamePeerID(got.BPID, tt.args.bpID) {
					t.Errorf("NewAgentCertV1() bpID = %v, want %v", got.BPID, tt.args.bpID)
				}
				if !types.IsSamePeerID(got.AgentID, tt.args.agentID) {
					t.Errorf("NewAgentCertV1() bpID = %v, want %v", got.AgentID, tt.args.agentID)
				}
				if !got.BPPubKey.IsEqual(tt.args.bpKey.PubKey()) {
					t.Errorf("NewAgentCertV1() pubKey = %v, want %v", enc.ToString(got.BPPubKey.SerializeCompressed()), enc.ToString(tt.args.bpKey.PubKey().SerializeCompressed()))
				}
				if !types.IsSamePeerID(got.BPID, tt.args.bpID) {
					t.Errorf("NewAgentCertV1() bpID = %v, want %v", got.BPID, tt.args.bpID)
				}

			}
		})
	}
}

func TestCheckAndGetV1(t *testing.T) {
	pk1, _ := btcec.NewPrivateKey(btcec.S256())
	libp2pKey1 := ConvertPKToLibP2P(pk1)
	//pk2, _ := btcec.NewPrivateKey(btcec.S256())
	pid1, _ := types.IDFromPrivateKey(libp2pKey1)
	pid2 := types.RandomPeerID()
	addrs := []string{"192.168.0.2", "2001:0db8:85a3:08d3:1319:8a2e:370:7334", "tester.aergo.io"}
	DAY := time.Hour * 24
	w, _ := NewAgentCertV1(pid1, pid2, pk1, addrs, DAY)
	tmpl, err := ConvertCertToProto(w)
	if err != nil {
		t.Fatalf("Failed to create test input. %s ", err.Error())
	}
	w2, _ := NewAgentCertV1(pid1, pid2, pk1, addrs, time.Second)
	w3, _ := NewAgentCertV1(pid1, pid2, pk1, addrs, DAY)
	w3.CreateTime = time.Now().Add(time.Second * 10)
	SignCert(pk1, w3)
	nftmpl, _ := ConvertCertToProto(w3)
	w3.CreateTime = time.Now().Add(time.Hour)
	SignCert(pk1, w3)
	lftmpl, _ := ConvertCertToProto(w3)
	if w.Signature.IsEqual(w2.Signature) {
		t.Fatalf("Something is strange")
	}
	w3.CreateTime = time.Now().Add(-1 * time.Hour * 30)
	w3.ExpireTime = time.Now().Add(-1 * time.Minute)
	SignCert(pk1, w3)
	extmpl, _ := ConvertCertToProto(w3)

	type args struct {
	}
	tests := []struct {
		name    string
		cert    *types.AgentCertificate
		wantErr error
	}{
		{"TSucc", NCB(tmpl).Build(), nil},
		{"TEmptyBPID", NCB(tmpl).bpid(nil).Build(), p2pcommon.ErrInvalidPeerID},
		{"TEmptyKey", NCB(tmpl).pubk(nil).Build(), p2pcommon.ErrInvalidKey},
		{"TEmptyAgentID", NCB(tmpl).agid(nil).Build(), p2pcommon.ErrInvalidPeerID},
		{"TEmptyAddrs", NCB(tmpl).addr([][]byte{}).Build(), p2pcommon.ErrInvalidCertField},
		{"TEmptySignature", NCB(tmpl).sig([]byte{}).Build(), p2pcommon.ErrInvalidCertField},
		{"TDiffSignature", NCB(tmpl).sig(w2.Signature.Serialize()).Build(), p2pcommon.ErrVerificationFailed},
		{"TDiffKeyAndID", NCB(tmpl).bpid([]byte(pid2)).Build(), p2pcommon.ErrInvalidKey},
		{"TTinyTimeErrCert", nftmpl, nil},
		{"TFutureCert", lftmpl, p2pcommon.ErrInvalidCertField},
		{"TExpiredCert", extmpl, p2pcommon.ErrInvalidCertField},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckAndGetV1(tt.cert)
			if err != tt.wantErr {
				t.Errorf("CheckAndGetV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil {
				p, err := ConvertCertToProto(got)
				if err != nil {
					t.Fatalf("CheckAndGetV1() wrong obj %v, failed to convert to protobuf obj %v", got, err.Error())
				}
				if !proto.Equal(p, tt.cert) {
					t.Fatalf("CheckAndGetV1() protobuf  %v, want %v", p.String(), tt.cert.String())
				}
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("CheckAndGetV1() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

type cb struct {
	tmpl *types.AgentCertificate
	copy *types.AgentCertificate
}

func NCB(tmpl *types.AgentCertificate) *cb {
	copy := proto.Clone(tmpl)
	b := &cb{tmpl: tmpl, copy: copy.(*types.AgentCertificate)}
	return b
}
func (b *cb) bpid(k []byte) *cb {
	b.copy.BPID = k
	return b
}
func (b *cb) pubk(k []byte) *cb {
	b.copy.BPPubKey = k
	return b
}
func (b *cb) agid(k []byte) *cb {
	b.copy.AgentID = k
	return b
}
func (b *cb) addr(k [][]byte) *cb {
	b.copy.AgentAddress = k
	return b
}
func (b *cb) sig(k []byte) *cb {
	b.copy.Signature = k
	return b
}
func (b *cb) ctime(t time.Time) *cb {
	b.copy.CreateTime = t.UnixNano()
	return b
}
func (b *cb) etime(t time.Time) *cb {
	b.copy.ExpireTime = t.UnixNano()
	return b
}

func (b *cb) Build() *types.AgentCertificate {
	return b.copy
}

func TestAgentCertificateV1_Convert(t *testing.T) {
	pk1, _ := btcec.NewPrivateKey(btcec.S256())
	//pk2, _ := btcec.NewPrivateKey(btcec.S256())

	pid1, _ := types.IDFromPrivateKey(ConvertPKToLibP2P(pk1))
	pid2 := types.RandomPeerID()
	addr0 := "192.168.0.2"
	addr1 := "2001:0db8:85a3:08d3:1319:8a2e:370:7334"
	addr2 := "tester.aergo.io"
	addrs := []string{addr0, addr1, addr2}

	DAY := time.Hour * 24

	type args struct {
		BPID         types.PeerID
		pk           *btcec.PrivateKey
		ttl          time.Duration
		AgentID      types.PeerID
		AgentAddress []string
	}
	tests := []struct {
		name string
		args args

		wantErr bool
	}{
		{"TSucc", args{pid1, pk1, DAY, pid2, addrs}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewAgentCertV1(tt.args.BPID, tt.args.AgentID, tt.args.pk, tt.args.AgentAddress, tt.args.ttl)

			got, err := ConvertCertToProto(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertCertToProto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			rev, err := CheckAndGetV1(got)
			if err != nil {
				t.Fatalf("CheckAndGetV1() error = %v, want nil", err)
			}
			if !reflect.DeepEqual(rev, w) {
				t.Errorf("ConvertCertToProto()->CheckAndGetV1() = %v, want %v", rev, w)
			}
			got2, err := ConvertCertToProto(rev)
			if !proto.Equal(got, got2) {
				t.Errorf("proto cert is differ %v, wantErr %v", got, got2)
			}
		})
	}
}

func Test_calculateCertificateHash(t *testing.T) {
	pk1, _ := btcec.NewPrivateKey(btcec.S256())
	//pk2, _ := btcec.NewPrivateKey(btcec.S256())
	pid1, pid2 := types.RandomPeerID(), types.RandomPeerID()
	addrs := []string{"192.168.0.2", "2001:0db8:85a3:08d3:1319:8a2e:370:7334", "tester.aergo.io"}
	DAY := time.Hour * 24
	w, _ := NewAgentCertV1(pid1, pid2, pk1, addrs, DAY)
	w2, _ := NewAgentCertV1(pid1, pid2, pk1, addrs, time.Hour)

	_, err := ConvertCertToProto(w)
	if err != nil {
		t.Fatalf("Failed to create test input. %s ", err.Error())
	}

	h1, err := calculateCertificateHash(w)
	if err != nil {
		t.Fatalf("Failed to create test input1. %s ", err.Error())
	}

	h11, err := calculateCertificateHash(w)
	if err != nil {
		t.Fatalf("Failed to create test input2. %s ", err.Error())
	}

	if !bytes.Equal(h1, h11) {
		t.Fatalf("calculated hash is differ! %v , want %v ", enc.ToString(h11), enc.ToString(h1))
	}
	h2, err := calculateCertificateHash(w2)
	if err != nil {
		t.Fatalf("Failed to create test input2. %s ", err.Error())
	}

	if bytes.Equal(h1, h2) {
		t.Fatalf("calculated hash is same! %v , want different ", enc.ToString(h2))
	}

}
