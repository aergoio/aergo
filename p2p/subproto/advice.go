/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package subproto

import (
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/rs/zerolog"
	"time"
)

func WithTimeLog(handler p2pcommon.MessageHandler, logger *log.Logger, level zerolog.Level) p2pcommon.MessageHandler {
	handler.AddAdvice(&LogHandleTimeAdvice{logger: logger, level:level})
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
		Str("elapsed", time.Since(a.timestamp).String()).
		Str("protocol", msg.Subprotocol().String()).
		Str("msgid", msg.ID().String()).
		Msg("handle takes")
}
