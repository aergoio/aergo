package p2putil

import (
	"bytes"
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func TestConvertPKToLibP2P(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"TNormal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btcPK, err := btcec.NewPrivateKey(btcec.S256())
			if err != nil {
				t.Fatalf("Failed to create test input pk: %v", err.Error())
			}
			got := ConvertPKToLibP2P(btcPK)
			if got == nil {
				t.Fatalf("ConvertPKToLibP2P() return nil ")
			}
			raw, err := got.Raw()
			if !bytes.Equal(raw, btcPK.Serialize()) {
				t.Errorf("ConvertPKToLibP2P() pk = %v, want %v", hex.Encode(raw), hex.Encode(btcPK.Serialize()))
			}
			rev := ConvertPKToBTCEC(got)
			if !bytes.Equal(rev.Serialize(), btcPK.Serialize()) {
				t.Errorf("ConvertPKToBTCEC() pk = %v, want %v", hex.Encode(rev.Serialize()), hex.Encode(btcPK.Serialize()))
			}

			marshaled, err := crypto.MarshalPrivateKey(got)
			if err != nil {
				t.Fatalf("Failed to create test input pk: %v", err.Error())
			}
			bs, err := crypto.MarshalPrivateKey(got)
			if err != nil {
				t.Fatalf("Failed to create test input pk: %v", err.Error())
			}
			if !bytes.Equal(marshaled, bs) {
				t.Fatalf("libp2p crypto Marshal and Bytes() is differ!")
			}
		})
	}
}

func TestConvertPubKeyToLibP2P(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"TNormal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btcPK, err := btcec.NewPrivateKey(btcec.S256())
			if err != nil {
				t.Fatalf("Failed to create test input pk: %v", err.Error())
			}
			pubKey := btcPK.PubKey()
			got := ConvertPubToLibP2P(pubKey)
			if got == nil {
				t.Fatalf("ConvertPubToLibP2P() return nil ")
			}
			raw, err := got.Raw()
			if !bytes.Equal(raw, pubKey.SerializeCompressed()) {
				t.Errorf("ConvertPubToLibP2P() pk = %v, want %v", hex.Encode(raw), hex.Encode(pubKey.SerializeCompressed()))
			}
			rev := ConvertPubKeyToBTCEC(got)
			if !bytes.Equal(rev.SerializeCompressed(), pubKey.SerializeCompressed()) {
				t.Errorf("ConvertPubKeyToBTCEC() pk = %v, want %v", hex.Encode(rev.SerializeCompressed()), hex.Encode(pubKey.SerializeCompressed()))
			}
		})
	}
}
