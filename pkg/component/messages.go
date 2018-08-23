/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import "time"

type CompStatReq struct {
	SentTime time.Time
}

type CompStatRsp struct {
	Status            string      `json:"status"`
	ProcessedMsg      uint64      `json:"acc_processed_msg"`
	QueuedMsg         uint64      `json:"acc_queued_msg"`
	MsgProcessLatency string      `json:"msg_latency"`
	Error             string      `json:"error"`
	Actor             interface{} `json:"actor"`
}
type Status = uint32

const (
	StartedStatus Status = 1 + iota
	StoppingStatus
	StoppedStatus
	RestartingStatus
)
