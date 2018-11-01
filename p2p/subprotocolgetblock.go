/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)


type blockResponseHandler struct {
	BaseMsgHandler
}

var _ MessageHandler = (*blockResponseHandler)(nil)

// newBlockRespHandler creates handler for GetBlockResponse
func newBlockRespHandler(pm PeerManager, peer RemotePeer, logger *log.Logger, actor ActorService, sm SyncManager) *blockResponseHandler {
	bh := &blockResponseHandler{BaseMsgHandler: BaseMsgHandler{protocol: GetBlocksResponse, pm: pm, sm:sm, peer: peer, actor: actor, logger: logger}}
	return bh
}

func (bh *blockResponseHandler) parsePayload(rawbytes []byte) (proto.Message, error) {
	return unmarshalAndReturn(rawbytes, &types.GetBlockResponse{})
}

func (bh *blockResponseHandler) handle(msg Message, msgBody proto.Message) {
	peerID := bh.peer.ID()
	remotePeer := bh.peer
	data := msgBody.(*types.GetBlockResponse)
	debugLogReceiveResponseMsg(bh.logger, bh.protocol, msg.ID().String(), msg.OriginalID().String(), peerID, fmt.Sprintf("blk_cnt=%d,hasNext=%t",len(data.Blocks),data.HasNext) )

	// locate request data and remove it if found
	if !remotePeer.GetReceiver(msg.OriginalID())(msg, data) {
		remotePeer.consumeRequest(msg.OriginalID())
		// TODO temporary code and will be deleted after newer syncer is made.
		if data.Status != types.ResultStatus_OK || len(data.Blocks) == 0 {
			return
		}
		bh.sm.HandleGetBlockResponse(remotePeer, msg, data)
	}
}

