/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"bytes"
	"context"
	"os"
	"sort"
	"strings"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util"
	"github.com/aergoio/aergo/v2/types"
	"github.com/spf13/cobra"
)

var getpeersCmd = &cobra.Command{
	Use:   "getpeers",
	Short: "Get Peer list",
	Run:   execGetPeers,
}

var nohidden bool
var showself bool
var sortFlag string
var detailed int

const (
	sortAddr    = "addr"
	sortID      = "id"
	sortHeight  = "height"
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
		cmd.Println(util.PeerListToString(msg))
	} else if detailed > 0 {
		// TODO show long fields
		cmd.Println(util.LongPeerListToString(msg))
	} else {
		cmd.Println(util.ShortPeerListToString(msg))
	}
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
	case sortDefault:
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
