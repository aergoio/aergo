package raftv2

import (
	"encoding/json"
	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/types"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testMbrs   []*consensus.Member
	testPeerID peer.ID
	testEncID  string

	testSnapData *consensus.SnapshotData
)

func init() {
	testEncID = "16Uiu2HAkxVB65cmCWceTu4HsHnz8WkUKknZXwr7PYdg2vy1fjDcU"
	testPeerID, _ = peer.IDB58Decode(testEncID)

	testMbrs = []*consensus.Member{
		{
			ID:     1,
			Name:   "testm1",
			Url:    "http://127.0.0.1:13001",
			PeerID: testPeerID,
		},
		{
			ID:     2,
			Name:   "testm2",
			Url:    "http://127.0.0.1:13002",
			PeerID: testPeerID,
		},
		{
			ID:     3,
			Name:   "testm3",
			Url:    "http://127.0.0.1:13003",
			PeerID: testPeerID,
		},
	}

	testBlock := types.NewBlock(nil, nil, nil, nil, nil, 0)

	testSnapData = consensus.NewSnapshotData(testMbrs, testBlock)
}

func TestMemberJson(t *testing.T) {
	mbr := testMbrs[0]

	jm := consensus.NewJsonMember(mbr)
	data, err := json.Marshal(jm)
	assert.NoError(t, err)

	var newJsonMbr = consensus.JsonMember{}
	err = json.Unmarshal(data, &newJsonMbr)

	t.Logf("peer=%s", newJsonMbr.PeerID)

	data, err = json.Marshal(mbr)
	assert.NoError(t, err)

	var newMbr = consensus.Member{}
	err = json.Unmarshal(data, &newMbr)
	assert.NoError(t, err)

	assert.NoError(t, err)
	t.Logf("peer=%s", peer.IDB58Encode(newMbr.PeerID))
}

func TestSnapDataJson(t *testing.T) {
	var snapdata = testSnapData

	data, err := snapdata.Encode()
	assert.NoError(t, err)

	var newSnapdata = &consensus.SnapshotData{}

	err = newSnapdata.Decode(data)
	assert.NoError(t, err)

	assert.True(t, snapdata.Equal(newSnapdata))
}
