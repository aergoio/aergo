/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var InvalidArgument = fmt.Errorf("invalid argument")

// ToMultiAddr make libp2p compatible Multiaddr object
func ToMultiAddr(ipAddr net.IP, port uint32) (multiaddr.Multiaddr, error) {
	var addrString string
	if ipAddr.To4() != nil {
		addrString = fmt.Sprintf("/ip4/%s/tcp/%d", ipAddr.String(), port)
	} else if ipAddr.To16() != nil {
		addrString = fmt.Sprintf("/ip6/%s/tcp/%d", ipAddr.String(), port)
	} else {
		return nil, InvalidArgument
	}
	peerAddr, err := multiaddr.NewMultiaddr(addrString)
	if err != nil {
		return nil, err
	}
	return peerAddr, nil
}

// PeerMetaToMultiAddr make libp2p compatible Multiaddr object from peermeta
func PeerMetaToMultiAddr(m p2pcommon.PeerMeta) (multiaddr.Multiaddr, error) {
	ipAddr, err := p2putil.GetSingleIPAddress(m.IPAddress)
	if err != nil {
		return nil, err
	}
	return ToMultiAddr(ipAddr, m.Port)
}

func FromMultiAddr(targetAddr multiaddr.Multiaddr) (p2pcommon.PeerMeta, error) {
	meta := p2pcommon.PeerMeta{}
	splitted := strings.Split(targetAddr.String(), "/")
	if len(splitted) != 7 {
		return meta, fmt.Errorf("invalid NPAddPeer addr format %s", targetAddr.String())
	}
	addrType := splitted[1]
	if addrType != "ip4" && addrType != "ip6" {
		return meta, fmt.Errorf("invalid NPAddPeer addr type %s", addrType)
	}
	peerAddrString := splitted[2]
	peerPortString := splitted[4]
	peerPort, err := strconv.Atoi(peerPortString)
	if err != nil {
		return meta, fmt.Errorf("invalid Peer port %s", peerPortString)
	}
	peerIDString := splitted[6]
	peerID, err := peer.IDB58Decode(peerIDString)
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

func ParseMultiAddrString(str string) (p2pcommon.PeerMeta, error) {
	ma, err := ParseMultiaddrWithResolve(str)
	if err != nil {
		return p2pcommon.PeerMeta{}, err
	}
	return FromMultiAddr(ma)
}

// ParseMultiaddrWithResolve parse string to multiaddr, additionally accept domain name with protocol /dns
// NOTE: this function is temporarilly use until go-multiaddr start to support dns.
func ParseMultiaddrWithResolve(str string) (multiaddr.Multiaddr, error) {
	ma, err := multiaddr.NewMultiaddr(str)
	if err != nil {
		// multiaddr is not support domain name yet. change domain name to ip address manually
		splitted := strings.Split(str, "/")
		if len(splitted) < 3 || !strings.HasPrefix(splitted[1], "dns") {
			return nil, err
		}
		domainName := splitted[2]
		ips, err := p2putil.ResolveHostDomain(domainName)
		if err != nil {
			return nil, fmt.Errorf("Could not get IPs: %v\n", err)
		}
		// use ipv4 as possible.
		ipSelected := false
		for _, ip := range ips {
			if ip.To4() != nil {
				splitted[1] = "ip4"
				splitted[2] = ip.To4().String()
				ipSelected = true
				break
			}
		}
		if !ipSelected {
			for _, ip := range ips {
				if ip.To16() != nil {
					splitted[1] = "ip6"
					splitted[2] = ip.To16().String()
					ipSelected = true
					break
				}
			}
		}
		if !ipSelected {
			return nil, err
		}
		rejoin := strings.Join(splitted, "/")
		return multiaddr.NewMultiaddr(rejoin)
	}
	return ma, nil
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
	// invariant: keyfile must not exists.
	if _, err2 := os.Stat(dir); os.IsNotExist(err2) {
		mkdirerr := os.MkdirAll(dir, os.ModePerm)
		if mkdirerr != nil {
			return nil, nil, mkdirerr
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

	pkFile := filepath.Join(dir, prefix+DefaultPkKeyExt)
//	pubFile := filepath.Join(dir, prefix+".pub")
	idFile := filepath.Join(dir, prefix+DefaultPeerIDExt)

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
	pid, _ := peer.IDFromPublicKey(pub)
	idBytes := []byte(peer.IDB58Encode(pid))
	idf.Write(idBytes)
	idf.Sync()
	return nil
}