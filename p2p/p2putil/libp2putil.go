/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	core "github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
)

// PeerMetaToMultiAddr make libp2p compatible Multiaddr object from peermeta
// @Deprecated
func PeerMetaToMultiAddr(m p2pcommon.PeerMeta) (types.Multiaddr, error) {
	idPart, err := multiaddr.NewComponent("p2p", m.ID.String())
	if err != nil {
		return nil, err
	}
	ma := multiaddr.Join(m.Addresses[0], idPart)
	return ma, nil
}

var addrProtos = []int{multiaddr.P_IP4, multiaddr.P_IP6, multiaddr.P_DNS, multiaddr.P_DNS4, multiaddr.P_DNS6}

// FromMultiAddr returns PeerMeta from multiaddr. the multiaddr must contain address(ip or domain name),
// port and peer id.
func FromMultiAddr(targetAddr multiaddr.Multiaddr) (p2pcommon.PeerMeta, error) {
	//var protocol int   this variable will be used soon
	meta, err := FromMultiAddrToPeerInfo(targetAddr)
	if err != nil {
		return meta, fmt.Errorf("invalid addr format %s", targetAddr.String())
	}
	return meta, nil
}

func FromMultiAddrString(str string) (p2pcommon.PeerMeta, error) {
	ma, err := types.ParseMultiaddr(str)
	if err != nil {
		return p2pcommon.PeerMeta{}, err
	}
	return FromMultiAddr(ma)
}

// FromMultiAddrToPeerInfo create PeerStaticInfo from qualified multiaddr which contains if and only if address, port and peerID
func FromMultiAddrToPeerInfo(ma types.Multiaddr) (p2pcommon.PeerMeta, error) {
	meta := p2pcommon.PeerMeta{}
	// check if ma contains all required information (address, port, peerID)
	protos := ma.Protocols()
	// check address
	if len(protos) != 3 || !types.IsAddress(protos[0]) ||
		protos[1].Code != multiaddr.P_TCP ||
		protos[2].Code != multiaddr.P_P2P {
		return meta, fmt.Errorf("invalid Peer addr format %s", ma.String())
	}
	addrPortPart, idPart := multiaddr.SplitLast(ma)
	peerIDString, err := idPart.ValueForProtocol(multiaddr.P_P2P)
	peerID, err := types.IDB58Decode(peerIDString)
	if err != nil {
		return meta, fmt.Errorf("invalid PeerID %s", peerIDString)
	}

	meta = p2pcommon.PeerMeta{ID: peerID, Addresses: []types.Multiaddr{addrPortPart}}
	return meta, nil
}

// FromMultiAddrStringWithPID
func FromMultiAddrStringWithPID(str string, id types.PeerID) (p2pcommon.PeerMeta, error) {
	addr1, err := types.ParseMultiaddr(str)
	if err != nil {
		return p2pcommon.PeerMeta{}, err
	}
	pidAddr, err := multiaddr.NewComponent(multiaddr.ProtocolWithCode(multiaddr.P_P2P).Name, id.String())
	if err != nil {
		return p2pcommon.PeerMeta{}, err
	}
	ma := multiaddr.Join(addr1, pidAddr)
	return FromMultiAddr(ma)
}

func LoadKeyFile(keyFile string) (crypto.PrivKey, crypto.PubKey, error) {
	dat, err := ioutil.ReadFile(keyFile)
	if err == nil {
		priv, err := crypto.UnmarshalPrivateKey(dat)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid keyfile. It's not private key file")
		}
		return priv, priv.GetPublic(), nil
	} else {
		return nil, nil, fmt.Errorf("Invalid keyfile path '" + keyFile + "'. Check the key file exists.")
	}
}

func GenerateKeyFile(dir, prefix string) (crypto.PrivKey, crypto.PubKey, error) {
	// invariant: key file must not exists.
	if _, err2 := os.Stat(dir); os.IsNotExist(err2) {
		mkdirErr := os.MkdirAll(dir, os.ModePerm)
		if mkdirErr != nil {
			return nil, nil, mkdirErr
		}
	}
	// file not exist and create new file
	priv, pub, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	if err != nil {
		return nil, nil, err
	}
	err = writeToKeyFiles(priv, pub, dir, prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate files %s.{key,id}: %v", prefix, err.Error())
	}

	return priv, priv.GetPublic(), nil
}

func writeToKeyFiles(priv crypto.PrivKey, pub crypto.PubKey, dir, prefix string) error {

	pkFile := filepath.Join(dir, prefix+p2pcommon.DefaultPkKeyExt)
	//	pubFile := filepath.Join(dir, prefix+".pub")
	idFile := filepath.Join(dir, prefix+p2pcommon.DefaultPeerIDExt)

	// Write private key file
	pkf, err := os.Create(pkFile)
	if err != nil {
		return err
	}
	pkBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return err
	}
	pkf.Write(pkBytes)
	pkf.Sync()

	// Write id file
	idf, err := os.Create(idFile)
	if err != nil {
		return err
	}
	pid, _ := types.IDFromPublicKey(pub)
	idBytes := []byte(types.IDB58Encode(pid))
	idf.Write(idBytes)
	idf.Sync()
	return nil
}

func ProtocolIDsToString(sli []core.ProtocolID) string {
	sb := bytes.NewBuffer(nil)
	sb.WriteByte('[')
	if len(sli) > 0 {
		stop := len(sli) - 1
		for i := 0; i < stop; i++ {
			sb.WriteString(string(sli[i]))
			sb.WriteByte(',')
		}
		sb.WriteString(string(sli[stop]))
	}
	sb.WriteByte(']')
	return sb.String()
}
