/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/jsonrpc"
	"github.com/spf13/cobra"
)

var getpeersCmd = &cobra.Command{
	Use:   "getpeers",
	Short: "Get Peer list",
	Run:   execGetPeers,
}

const AergoDir = ".aergo"
const NodeAliasFile = "nodeAliases.json"

var nohidden bool
var showself bool
var sortFlag string
var detailed int
var nameAliases map[string]string

const (
	sortAddr    = "addr"
	sortID      = "id"
	sortHeight  = "height"
	sortAlias   = "alias"
	sortDefault = "no"
)

const (
	DetailShort   int = -1
	DetailDefault     = 0
	DetailLong        = 1
)

func init() {
	rootCmd.AddCommand(getpeersCmd)
	getpeersCmd.Flags().BoolVar(&nohidden, "nohidden", false, "exclude hidden peers")
	getpeersCmd.Flags().BoolVar(&showself, "self", false, "show self peer info")
	getpeersCmd.Flags().StringVar(&sortFlag, "sort", "no", "sort peers by address, id or other")
	getpeersCmd.Flags().IntVar(&detailed, "detail", 0, "detail level")

	nameAliases = map[string]string{}
}

func execGetPeers(cmd *cobra.Command, args []string) {
	sorter := GetSorter(cmd, sortFlag)
	msg, err := client.GetPeers(context.Background(), &types.PeersParams{NoHidden: nohidden, ShowSelf: showself})
	if err != nil {
		cmd.Printf("Failed to get peer from server: %s\n", err.Error())
		return
	}
	// address and peerid should be encoded, respectively
	sorter.Sort(msg.Peers)
	if detailed == 0 {
		res := jsonrpc.ConvPeerList(msg)
		cmd.Println(jsonrpc.MarshalJSON(res))
	} else if detailed > 0 {
		// TODO show long fields
		res := jsonrpc.ConvLongPeerList(msg)
		cmd.Println(jsonrpc.MarshalJSON(res))
	} else if detailed == -2 {
		showPrettyAndShort(cmd, msg)
	} else {
		res := jsonrpc.ConvShortPeerList(msg)
		cmd.Println(jsonrpc.MarshalJSON(res))
	}
}

func showPrettyAndShort(cmd *cobra.Command, msg *types.PeerList) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	aliasPath := filepath.Join(dir, AergoDir, NodeAliasFile)
	// If the file doesn't exist, create it, or append to the file
	f, err := os.Open(aliasPath)
	if err == nil {
		defer f.Close()
		jsonString, _ := io.ReadAll(f)
		json.Unmarshal(jsonString, &nameAliases)
	}

	res := convShort2PeerList(msg, nameAliases)
	cmd.Println(jsonrpc.MarshalJSON(res))
}

type shortPeerInfo struct {
	alias    string
	addr     string
	height   uint64
	roleName string
	peer     *types.Peer
}

func convShort2PeerList(msg *types.PeerList, aliases map[string]string) *jsonrpc.InOutShortPeerList {
	if msg == nil {
		return nil
	}
	p := &jsonrpc.InOutShortPeerList{}
	peersCount := len(msg.Peers)
	shortPeers := make([]shortPeerInfo, peersCount)
	nameSize := 0
	for i, peer := range msg.Peers {
		pa := peer.Address
		peerName := aliases[types.PeerID(pa.PeerID).String()]
		if len(peerName) == 0 {
			peerName = ""
		}
		if len(peerName) > nameSize {
			nameSize = len(peerName)
		}
		shortPeers[i] = shortPeerInfo{peerName, p2putil.ShortForm(types.PeerID(peer.Address.PeerID)), peer.Bestblock.BlockNo, peer.AcceptedRole.String(), peer}
	}
	// TODO sorting by alias cannot be done in original algorithms, so ad-hoc is added here for now.
	if sortFlag == sortAlias {
		sort.Sort(byAliasName(shortPeers))
	}
	p.Peers = make([]string, peersCount)
	for i, sp := range shortPeers {
		if nameSize > 0 {
			p.Peers[i] = fmt.Sprintf("%*s;%5s;%15s/%5d;%9s;%10d", nameSize, sp.alias, sp.addr,
				sp.peer.Address.Address, sp.peer.Address.Port, sp.roleName, sp.height)
		} else {
			p.Peers[i] = fmt.Sprintf("%5s;%15s/%5d;%9s;%10d", sp.addr,
				sp.peer.Address.Address, sp.peer.Address.Port, sp.roleName, sp.height)
		}
	}
	return p
}

func Must(a0 string, _ error) string {
	return a0
}

func GetSorter(cmd *cobra.Command, flag string) peerSorter {
	switch flag {
	case sortAddr:
		return addrSorter{}
	case sortID:
		return idSorter{}
	case sortHeight:
		return heightSorter{}
	case sortDefault, sortAlias: // TODO fix ad-hoc code later
		return noSorter{}
	default:
		cmd.Println("Invalid sort type", flag)
		os.Exit(1)
		return noSorter{}
	}
}

type peerSorter interface {
	Sort([]*types.Peer)
}
type addrSorter struct{}

func (addrSorter) Sort(peerArr []*types.Peer) {
	sort.Sort(byAddr(peerArr))
}

type idSorter struct{}

func (idSorter) Sort(peerArr []*types.Peer) {
	sort.Sort(byID(peerArr))
}

type heightSorter struct{}

func (heightSorter) Sort(peerArr []*types.Peer) {
	sort.Sort(byHeight(peerArr))
}

type noSorter struct{}

func (noSorter) Sort(peerArr []*types.Peer) {
	// Do nothing. no sort
}

type byAddr []*types.Peer

func (s byAddr) Len() int {
	return len(s)
}

func (s byAddr) Less(i, j int) bool {
	result := strings.Compare(s[i].Address.Address, s[j].Address.Address)
	if result == 0 {
		result = int(s[i].Address.Port) - int(s[j].Address.Port)
	}
	return result < 0
}

func (s byAddr) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type byID []*types.Peer

func (s byID) Len() int {
	return len(s)
}

func (s byID) Less(i, j int) bool {
	result := bytes.Compare(s[i].Address.PeerID, s[j].Address.PeerID)
	return result < 0
}

func (s byID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type byHeight []*types.Peer

func (s byHeight) Len() int {
	return len(s)
}

func (s byHeight) Less(i, j int) bool {
	if s[i].Bestblock.BlockNo < s[j].Bestblock.BlockNo {
		return true
	} else if s[i].Bestblock.BlockNo > s[j].Bestblock.BlockNo {
		return false
	} else {
		return bytes.Compare(s[i].Bestblock.GetBlockHash(), s[j].Bestblock.GetBlockHash()) < 0
	}
}

func (s byHeight) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type byAliasName []shortPeerInfo

func (s byAliasName) Len() int {
	return len(s)
}

func (s byAliasName) Less(i, j int) bool {
	result := strings.Compare(s[i].alias, s[j].alias)
	return result < 0
}

func (s byAliasName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
