/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

type StatusReq struct{}

type StatusRsp struct {
	Status Status
	MsgLat int64
	MsgNum map[string]int
}

type Status int

const (
	StartedStatus Status = 1 + iota
	StoppingStatus
	StoppedStatus
	RestartingStatus
)
