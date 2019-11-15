package p2putil

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-core/crypto"
	"testing"
)

func Test(t *testing.T) {
	priv, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	//priv, pub, _ := crypto.GenerateKeyPair(crypto.Ed25519, 2048)

	t.Log("Private Key")
	marshaled, _ := crypto.MarshalPrivateKey(priv)
	PrintLibP2PKey(priv, marshaled, t)
	t.Log("Public Key")
	marshaled, _ = crypto.MarshalPublicKey(pub)
	PrintLibP2PKey(pub, marshaled, t)
}

func PrintLibP2PKey(priv crypto.Key, marshaled []byte, t *testing.T) {
	oldBytes, err := priv.Bytes()
	newBytes, err := priv.Raw()
	if err != nil {
		t.Errorf("Failed to get bytes: %v", err.Error())
	} else {
		t.Logf("BT/MAR %v", hex.EncodeToString(oldBytes))
		t.Logf("RAW    %v", hex.EncodeToString(newBytes))
	}
}

func PrintBTCPKey(priv *btcec.PrivateKey, t *testing.T) {
	oldBytes := priv.Serialize()
	t.Logf("PRIV   %v", hex.EncodeToString(oldBytes))
	t.Logf("PUBLIC %v", hex.EncodeToString(priv.PubKey().SerializeCompressed()))
}

func TestLibs(t *testing.T) {
	btcKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Fatalf("Failed to generate btcec key : %s ", err)
	}
	libp2pKey, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	if err != nil {
		t.Fatalf("Failed to generate libp2p key : %s ", err)
	}

	t.Log("BTC key")
	PrintBTCPKey(btcKey, t)
	t.Log("Private Key")
	marshaled, _ := crypto.MarshalPrivateKey(libp2pKey)
	PrintLibP2PKey(libp2pKey, marshaled, t)
	PrintLibP2PKey(pubKey, marshaled, t)
}


func TestLibs2(t *testing.T) {
	btcKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Fatalf("Failed to generate btcec key : %s ", err)
	}
	libp2pKey, err := crypto.UnmarshalSecp256k1PrivateKey(btcKey.Serialize())
	if err != nil {
		t.Fatalf("Failed to generate libp2p key : %s ", err)
	}

	t.Log("BTC key")
	PrintBTCPKey(btcKey, t)

	t.Log("LibP2P Key")
	marshaled, _ := crypto.MarshalPrivateKey(libp2pKey)
	PrintLibP2PKey(libp2pKey, marshaled, t)
	t.Log("LibP2P Public Key")
	PrintLibP2PKey(libp2pKey.GetPublic(), marshaled, t)
}
