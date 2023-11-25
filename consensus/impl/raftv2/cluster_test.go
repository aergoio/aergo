package raftv2

import (
	"encoding/json"
	"testing"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var (
	testMbrs   []*consensus.Member
	testPeerID types.PeerID
	testEncID  string

	testSnapData   *consensus.SnapshotData
	testPeerIDs    []types.PeerID
	testPeerIDStrs = []string{
		"16Uiu2HAkvaAMCHkd9hZ6hQkdDLKoXP4eLJSqkMF1YqkSNy5v9SVn",
		"16Uiu2HAmJqEp9f9WAbzFxkLrnHnW4EuUDM69xkCDPF26HmNCsib6",
		"16Uiu2HAmA2ysmFxoQ37sk1Zk2sMrPysqTmwYAFrACyf3LtP3gxpJ",
		"16Uiu2HAmQti7HLHC9rXqkeABtauv2YsCPG3Uo1WLqbXmbuxpbjmF",
	}
)

func init() {
	testEncID = "16Uiu2HAkxVB65cmCWceTu4HsHnz8WkUKknZXwr7PYdg2vy1fjDcU"
	testPeerIDs = make([]types.PeerID, len(testPeerIDStrs))
	testPeerID, _ = types.IDB58Decode(testEncID)

	for i, peerStr := range testPeerIDStrs {
		peerID, _ := types.IDB58Decode(peerStr)
		testPeerIDs[i] = peerID
	}

	testMbrs = []*consensus.Member{
		{MemberAttr: types.MemberAttr{
			ID:      1,
			Name:    "testm1",
			Address: "/ip4/127.0.0.1/13001",
			PeerID:  []byte(testPeerIDs[0]),
		}},
		{MemberAttr: types.MemberAttr{
			ID:      2,
			Name:    "testm2",
			Address: "/ip4/127.0.0.1/tcp/13002",
			PeerID:  []byte(testPeerIDs[1]),
		}},
		{MemberAttr: types.MemberAttr{
			ID:      3,
			Name:    "testm3",
			Address: "/ip4/127.0.0.1/tcp/13003",
			PeerID:  []byte(testPeerIDs[2]),
		}},
	}

	testBlock := types.NewBlock(types.EmptyBlockHeaderInfo, nil, nil, nil, nil, nil)

	testSnapData = consensus.NewSnapshotData(testMbrs, nil, testBlock)
}

func TestMemberJson(t *testing.T) {
	mbr := testMbrs[0]

	data, err := json.Marshal(mbr)
	assert.NoError(t, err)

	var newMbr = consensus.Member{}
	err = json.Unmarshal(data, &newMbr)
	assert.NoError(t, err)

	assert.NoError(t, err)
	//t.Logf("peer=%s", types.IDB58Encode(newMbr.GetPeerID()))

	assert.True(t, mbr.Equal(&newMbr))

	mbrRemove := &consensus.Member{MemberAttr: types.MemberAttr{ID: 1}}
	data, err = json.Marshal(mbrRemove)
	assert.NoError(t, err)

	newMbr = consensus.Member{}
	err = json.Unmarshal(data, &newMbr)
	assert.NoError(t, err)
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

func TestClusterConfChange(t *testing.T) {
	// init cluster
	serverCtx := config.NewServerContext("", "")
	testCfg := serverCtx.GetDefaultConfig().(*config.Config)
	testCfg.Consensus.Raft = &config.RaftConfig{
		Name: "test1",
		/*
			BPs: []config.RaftBPConfig{
				{"test1", "/ip4/127.0.0.1/tcp/10001", testPeerIDs[0]},
				{"test2", "/ip4/127.0.0.1/tcp/10002", testPeerIDs[1]},
				{"test3", "/ip4/127.0.0.1/tcp/10003", testPeerIDs[2]},
			},*/
	}

	mbrs := []*types.MemberAttr{
		{ID: 0, Name: "test1", Address: "/ip4/127.0.0.1/tcp/10001", PeerID: []byte(testPeerIDs[0])},
		{ID: 1, Name: "test2", Address: "/ip4/127.0.0.1/tcp/10002", PeerID: []byte(testPeerIDs[1])},
		{ID: 2, Name: "test3", Address: "/ip4/127.0.0.1/tcp/10003", PeerID: []byte(testPeerIDs[2])},
	}

	cl := NewCluster([]byte("test"), nil, "test1", testPeerIDs[0], 0, nil)

	err := cl.AddInitialMembers(mbrs)
	assert.NoError(t, err)

	// add applied members
	for _, m := range cl.Members().ToArray() {
		err = cl.addMember(m, true)
		assert.NoError(t, err)
	}

	// normal case
	req := &types.MembershipChange{
		Type: types.MembershipChangeType_ADD_MEMBER,
		Attr: &types.MemberAttr{ID: 3, Name: "test4", Address: "/ip4/127.0.0.1/tcp/10004", PeerID: []byte(testPeerIDs[3])},
	}
	_, err = cl.makeProposal(req, true)
	assert.NoError(t, err)

	id := cl.getNodeID("test3")
	req = &types.MembershipChange{
		Type: types.MembershipChangeType_REMOVE_MEMBER,
		Attr: &types.MemberAttr{ID: id},
	}

	_, err = cl.makeProposal(req, true)
	assert.NoError(t, err)

	// failed case
	req = &types.MembershipChange{
		Type: types.MembershipChangeType_ADD_MEMBER,
		Attr: &types.MemberAttr{Address: "/ip4/127.0.0.1/tcp/10004", PeerID: []byte(testPeerIDs[3])},
	}
	_, err = cl.makeProposal(req, true)
	assert.Error(t, err, "no name")

	req = &types.MembershipChange{
		Type: types.MembershipChangeType_ADD_MEMBER,
		Attr: &types.MemberAttr{Name: "test4", Address: "/ip4/127.0.0.1/tcp/10004", PeerID: []byte(testPeerIDs[0])},
	}
	_, err = cl.makeProposal(req, true)
	assert.Error(t, err, "duplicate peerid")

	req = &types.MembershipChange{
		Type: types.MembershipChangeType_REMOVE_MEMBER,
		Attr: &types.MemberAttr{Name: "test4", Address: "/ip4/127.0.0.1/tcp/10004", PeerID: []byte(testPeerIDs[3])},
	}
	_, err = cl.makeProposal(req, true)
	assert.Error(t, err, "no id to remove")

}

func TestClusterEqual(t *testing.T) {
	//isAllMembersEqual
	cl := NewCluster([]byte("test"), nil, "testm1", testPeerIDs[0], 0, nil)
	for _, m := range testMbrs {
		err := cl.addMember(m, true)
		assert.NoError(t, err)
	}

	assert.True(t, cl.isAllMembersEqual(testMbrs, nil))

	rm := testMbrs[2]
	err := cl.removeMember(rm)
	assert.NoError(t, err)

	rmMembers := []*consensus.Member{rm}
	assert.True(t, cl.isAllMembersEqual(testMbrs[0:2], rmMembers))

	cl = NewCluster([]byte("test"), nil, "testm1", testPeerIDs[0], 0, nil)
	for _, m := range testMbrs {
		newM := *m
		newM.Address = "invalidaddress"
		err := cl.addMember(&newM, true)
		assert.NoError(t, err)
	}

	assert.False(t, cl.isAllMembersEqual(testMbrs, nil))

	cl = NewCluster([]byte("test"), nil, "testm1", testPeerIDs[0], 0, nil)
	for i, m := range testMbrs {
		newM := *m
		newM.ID = uint64(i) + 100
		err := cl.addMember(&newM, true)
		assert.NoError(t, err)
	}

	assert.False(t, cl.isAllMembersEqual(testMbrs, nil))
}
