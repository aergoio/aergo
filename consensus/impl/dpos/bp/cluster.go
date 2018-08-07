/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package bp

import (
	"fmt"

	"github.com/aergoio/aergo/consensus/impl/dpos/param"
	"github.com/libp2p/go-libp2p-peer"
)

type errBpSize struct {
	required uint16
	given    uint16
}

func (e errBpSize) Error() string {
	return fmt.Sprintf("insufficient or redundant block producers  - %v (required - %v)", e.given, e.required)
}

// Cluster represents a cluster of block producers.
type Cluster struct {
	size   uint16
	member map[uint16]*blockProducer
	index  map[peer.ID]uint16
}

// blockProducer represents one member in the block producer cluster.
type blockProducer struct {
	id peer.ID
}

// NewCluster returns a new bp.Cluster.
func NewCluster(ids []string) (*Cluster, error) {
	if len(ids) != param.BlockProducers {
		return nil, errBpSize{required: param.BlockProducers, given: uint16(len(ids))}
	}

	c := &Cluster{
		size:   param.BlockProducers,
		member: make(map[uint16]*blockProducer),
		index:  make(map[peer.ID]uint16),
	}

	for i, id := range ids {
		bpID, err := peer.IDB58Decode(id)
		if err != nil {
			return nil, fmt.Errorf("invalid node ID[%d]: %s", i, err.Error())
		}

		index := uint16(i)
		c.member[index] = newBlockProducer(bpID)
		c.index[bpID] = index
	}

	if len(c.index) != param.BlockProducers {
		return nil, errBpSize{required: param.BlockProducers, given: uint16(len(ids))}
	}

	return c, nil
}

func newBlockProducer(id peer.ID) *blockProducer {
	return &blockProducer{id: id}
}

// BpIndex2ID returns the ID correspinding to idx.
func (c *Cluster) BpIndex2ID(idx uint16) (peer.ID, bool) {
	if bp, exist := c.member[idx]; exist {
		return bp.id, exist
	}
	return peer.ID(""), false
}

// BpID2Index returns the index corresponding to id.
func (c *Cluster) BpID2Index(id peer.ID) (uint16, bool) {
	idx, exist := c.index[id]
	return idx, exist
}

// Has reports whether c includes id or not
func (c *Cluster) Has(id peer.ID) bool {
	_, exist := c.index[id]
	return exist
}
