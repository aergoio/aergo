/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"github.com/aergoio/aergo/types"
)

type TxNoticeTracer interface {
	RegisterTxNotice(txIDs []types.TxID, cnt int, alreadySent []types.PeerID)
	ReportSend(txIDs []types.TxID, peerID types.PeerID)
	ReportNotSend(txIDs []types.TxID, cnt int)
}
//go:generate mockgen -source=txnotice.go  -package=p2pmock -destination=../p2pmock/mock_txnotice.go

type ReportType int

const (
	Send ReportType = iota
	Fail
	Skip
)
