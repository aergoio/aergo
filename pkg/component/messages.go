/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import "time"

// CompStatReq is an actor message; requesting component's statics
// When a component gets this message, then it collects infos and
// returns CompStatRsp to the requester
type CompStatReq struct {
	SentTime time.Time
}

// CompStatRsp contains component's internal info, used to help debugging
// - Status is a string representation of a component's status
// - AccProcessedMsg is the accumulated number of message that this component processes
// - MsgQueueLen is the current number of message at this component's mailbox
// - MsgProcessLatency is an estimated latency to process a msg
// - Error is an error msg when a requester fails to get statics
// - Actor is a reserved field to get component's internal debug info
type CompStatRsp struct {
	Status            string      `json:"status"`
	AccProcessedMsg   uint64      `json:"acc_processed_msg"`
	MsgQueueLen       uint64      `json:"msg_queue_len"`
	MsgProcessLatency string      `json:"msg_latency"`
	Error             string      `json:"error"`
	Actor             interface{} `json:"actor"`
}

// Status represents a component's current running status
type Status = uint32

const (
	// StartedStatus means a component is working
	StartedStatus Status = 1 + iota
	// StoppingStatus means a component is stopping
	StoppingStatus
	// StoppedStatus means a component is already stopped
	StoppedStatus
	// RestartingStatus means a component is now restarting
	RestartingStatus
)
