/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pkey

import (
	"os"
	"path/filepath"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
	"github.com/libp2p/go-libp2p/core/crypto"
)

type nodeInfo struct {
	id      types.PeerID
	sid     string
	pubKey  crypto.PubKey
	privKey crypto.PrivKey

	version   string
	startTime time.Time
}

var ni *nodeInfo

// InitNodeInfo initializes node-specific information like node id.
// Caution: this must be called before all the goroutines are started.
func InitNodeInfo(baseCfg *config.BaseConfig, p2pCfg *config.P2PConfig, version string, logger *log.Logger) {
	// check Key and address
	var (
		priv crypto.PrivKey
		pub  crypto.PubKey
		err  error
	)

	if !p2pcommon.CheckVersion(version) {
		logger.Warn().Str("minVersion", p2pcommon.MinimumAergoVersion).Str("maxVersion", p2pcommon.MaximumAergoVersion).Str("version", version).Msg("min/max version range is not set properly. change constant in source and then rebuild it")
	}

	if p2pCfg.NPKey != "" {
		priv, pub, err = p2putil.LoadKeyFile(p2pCfg.NPKey)
		if err != nil {
			panic("Failed to load Keyfile '" + p2pCfg.NPKey + "' " + err.Error())
		}
	} else {
		logger.Info().Msg("No private key file is configured, so use auto-generated pk file instead.")

		autogenFilePath := filepath.Join(baseCfg.AuthDir, p2pcommon.DefaultPkKeyPrefix+p2pcommon.DefaultPkKeyExt)
		if _, err := os.Stat(autogenFilePath); os.IsNotExist(err) {
			logger.Info().Str("pk_file", autogenFilePath).Msg("Generate new private key file.")
			priv, pub, err = p2putil.GenerateKeyFile(baseCfg.AuthDir, p2pcommon.DefaultPkKeyPrefix)
			if err != nil {
				panic("Failed to generate new pk file: " + err.Error())
			}
		} else {
			logger.Info().Str("pk_file", autogenFilePath).Msg("Load existing generated private key file.")
			priv, pub, err = p2putil.LoadKeyFile(autogenFilePath)
			if err != nil {
				panic("Failed to load generated pk file '" + autogenFilePath + "' " + err.Error())
			}
		}
	}
	id, _ := types.IDFromPublicKey(pub)

	ni = &nodeInfo{
		id:        id,
		sid:       base58.Encode([]byte(id)),
		pubKey:    pub,
		privKey:   priv,
		version:   version,
		startTime: time.Now(),
	}

	p2putil.UseFullID = p2pCfg.LogFullPeerID
}

// NodeID returns the node id.
func NodeID() types.PeerID {
	return ni.id
}

// NodeSID returns the string representation of the node id.
func NodeSID() string {
	if ni == nil {
		return ""
	}
	return ni.sid
}

// NodePrivKey returns the private key of the node.
func NodePrivKey() crypto.PrivKey {
	return ni.privKey
}

// NodePubKey returns the public key of the node.
func NodePubKey() crypto.PubKey {
	return ni.pubKey
}

// NodeVersion returns the version of this binary. TODO: It's not good that version info is in p2pkey package
func NodeVersion() string {
	return ni.version
}

func StartTime() time.Time {
	return ni.startTime
}

func GetHostAccessor() types.HostAccessor {
	return simpleHostAccessor{}
}

type simpleHostAccessor struct{}

func (simpleHostAccessor) Version() string {
	return ni.version
}

func (simpleHostAccessor) StartTime() time.Time {
	return ni.startTime
}
