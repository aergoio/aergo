/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/rs/zerolog"
)

func WithTimeLog(handler p2pcommon.MessageHandler, logger *log.Logger, level zerolog.Level) p2pcommon.MessageHandler {
	handler.AddAdvice(&LogHandleTimeAdvice{logger: logger, level: level})
	return handler
}

type LogHandleTimeAdvice struct {
	logger    *log.Logger
	level     zerolog.Level
	timestamp time.Time
}

func (a *LogHandleTimeAdvice) PreHandle() {
	a.timestamp = time.Now()
}

func (a *LogHandleTimeAdvice) PostHandle(msg p2pcommon.Message, msgBody p2pcommon.MessageBody) {
	a.logger.WithLevel(a.level).
		Int64("elapsed", time.Since(a.timestamp).Nanoseconds()/1000).
		Stringer(p2putil.LogProtoID, msg.Subprotocol()).
		Stringer(p2putil.LogMsgID, msg.ID()).
		Msg("handle takes")
}
