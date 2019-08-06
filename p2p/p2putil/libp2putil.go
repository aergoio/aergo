/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo/internal/network"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/test"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)


func RandomPeerID() (types.PeerID, error) {
	return test.RandPeerID()
}
// PeerMetaToMultiAddr make libp2p compatible Multiaddr object from peermeta
func PeerMetaToMultiAddr(m p2pcommon.PeerMeta) (multiaddr.Multiaddr, error) {
	ipAddr, err := network.GetSingleIPAddress(m.IPAddress)
	if err != nil {
		return nil, err
	}
	return types.ToMultiAddr(ipAddr, m.Port)
}

func FromMultiAddr(targetAddr multiaddr.Multiaddr) (p2pcommon.PeerMeta, error) {
	meta := p2pcommon.PeerMeta{}
	split := strings.Split(targetAddr.String(), "/")
	if len(split) != 7 {
		return meta, fmt.Errorf("invalid NPAddPeer addr format %s", targetAddr.String())
	}
	addrType := split[1]
	if addrType != "ip4" && addrType != "ip6" {
		return meta, fmt.Errorf("invalid NPAddPeer addr type %s", addrType)
	}
	peerAddrString := split[2]
	peerPortString := split[4]
	peerPort, err := strconv.Atoi(peerPortString)
	if err != nil {
		return meta, fmt.Errorf("invalid Peer port %s", peerPortString)
	}
	peerIDString := split[6]
	peerID, err := types.IDB58Decode(peerIDString)
	if err != nil {
		return meta, fmt.Errorf("invalid PeerID %s", peerIDString)
	}
	meta = p2pcommon.PeerMeta{
		ID:        peerID,
		Port:      uint32(peerPort),
		IPAddress: peerAddrString,
	}
	return meta, nil
}

func FromMultiAddrString(str string) (p2pcommon.PeerMeta, error) {
	ma, err := types.ParseMultiaddrWithResolve(str)
	if err != nil {
		return p2pcommon.PeerMeta{}, err
	}
	return FromMultiAddr(ma)
}


func FromMultiAddrStringWithPID(str string, id types.PeerID) (p2pcommon.PeerMeta, error) {
	addr1, err := types.ParseMultiaddrWithResolve(str)
	if err != nil {
		return p2pcommon.PeerMeta{}, err
	}
	pidAddr, err := multiaddr.NewComponent(multiaddr.ProtocolWithCode(multiaddr.P_P2P).Name, id.Pretty())
	if err != nil {
		return p2pcommon.PeerMeta{}, err
	}
	ma := multiaddr.Join(addr1, pidAddr)
	return FromMultiAddr(ma)
}

// ExtractIPAddress returns ip address from multiaddr. it return null if ma has no ip field.
func ExtractIPAddress(ma multiaddr.Multiaddr) net.IP {
	ipStr, err := ma.ValueForProtocol(multiaddr.P_IP4)
	if err == nil {
		return net.ParseIP(ipStr)
	}
	ipStr, err = ma.ValueForProtocol(multiaddr.P_IP6)
	if err == nil {
		return net.ParseIP(ipStr)
	}
	return nil
}

func LoadKeyFile(keyFile string) (crypto.PrivKey, crypto.PubKey, error) {
	dat, err := ioutil.ReadFile(keyFile)
	if err == nil {
		priv, err := crypto.UnmarshalPrivateKey(dat)
		if err != nil {
			return nil,nil, fmt.Errorf("invalid keyfile. It's not private key file")
		}
		return priv, priv.GetPublic(), nil
	} else {
		return nil, nil, fmt.Errorf("Invalid keyfile path '"+ keyFile +"'. Check the key file exists.")
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
	pkBytes, err := priv.Bytes()
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
		stop := len(sli)-1
		for i:=0 ; i<stop; i++ {
			sb.WriteString(string(sli[i]))
			sb.WriteByte(',')
		}
		sb.WriteString(string(sli[stop]))
	}
	sb.WriteByte(']')
	return sb.String()
}
