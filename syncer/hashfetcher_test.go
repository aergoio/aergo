package syncer

import (
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashFetcher_normal(t *testing.T) {
	//make remoteBlockChain
	remoteChain := initStubBlockChain(nil, 10)

	ancestor := remoteChain.GetBlockByNo(0)

	ctx := types.NewSyncCtx("p1", 5, 1)
	ctx.SetAncestor(ancestor)

	syncer := NewStubSyncer(ctx, false, true, false, nil, remoteChain)
	syncer.hf.Start()

	//hashset 1~3, 4~5
	//receive GetHash message
	msg := syncer.testhub.recvMessage()
	assert.IsTypef(t, &message.GetHashes{}, msg, "invalid message from hf")
	syncer.handleMessageManual(t, msg, nil)

	//when pop result msg, hashfetcher send new request
	resHashSet := syncer.getResultFromHashFetcher()
	assert.Equal(t, int(TestMaxHashReq), resHashSet.Count)

	msg = syncer.testhub.recvMessage()
	syncer.handleMessageManual(t, msg, nil)

	//when pop result msg, hashfetcher send new request

	resHashSet = syncer.getResultFromHashFetcher()
	assert.Equal(t, 2, resHashSet.Count)

	//receive close hashfetcher message
	msg = syncer.testhub.recvMessage()
	assert.IsTypef(t, &message.CloseFetcher{}, msg, "need syncer close hashfetcher msg")
	syncer.handleMessageManual(t, msg, nil)
	assert.Nil(t, syncer.hf, "hashfetcher set nil")

	syncer.stop(t)
}

func TestHashFetcher_ResponseError(t *testing.T) {
	//make remoteBlockChain
	remoteChain := initStubBlockChain(nil, 10)
	ancestor := remoteChain.GetBlockByNo(0)

	ctx := types.NewSyncCtx("p1", 5, 1)
	ctx.SetAncestor(ancestor)

	syncer := NewStubSyncer(ctx, false, true, false, nil, remoteChain)
	syncer.hf.Start()

	//hashset 2~4, 5~7, 8~9
	//receive GetHash message
	msg := syncer.testhub.recvMessage()
	assert.IsTypef(t, &message.GetHashes{}, msg, "invalid message from hf")
	syncer.handleMessageManual(t, msg, ErrGetHashesRspError)

	//stop
	msg = syncer.testhub.recvMessage()
	assert.IsTypef(t, &message.SyncStop{}, msg, "need syncer stop msg")
	syncer.handleMessageManual(t, msg, nil)

	assert.True(t, syncer.isStop, "hashfetcher finished")
}
