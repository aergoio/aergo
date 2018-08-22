/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import "time"

type CompStatReq struct {
	SentTime time.Time
}

type CompStatRsp map[string]interface{}

type Status = uint32

const (
	StartedStatus Status = 1 + iota
	StoppingStatus
	StoppedStatus
	RestartingStatus
)
