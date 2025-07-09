package raftv2

import (
	"fmt"
	"strings"

	"github.com/aergoio/aergo-lib/log"
)

// Logger is a logging unit. It controls the flow of messages to a given
// (swappable) backend.
type RaftLogger struct {
	logger *log.Logger
}

func NewRaftLogger(logger *log.Logger) *RaftLogger {
	if logger == nil {
		panic("base logger of raft is nil")
	}

	return &RaftLogger{logger: logger.WithSkipFrameCount(3)}
}
func (l RaftLogger) Fatal(args ...interface{}) {
	s := fmt.Sprint(args...)
	l.logger.Fatal().Msgf("%s", s)
}

func (l *RaftLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

func (l *RaftLogger) Panic(args ...interface{}) {
	s := fmt.Sprint(args...)
	l.logger.Panic().Msgf("%s", s)
}

func (l *RaftLogger) Panicf(format string, args ...interface{}) {
	l.logger.Panic().Msgf(format, args...)
}

func (l *RaftLogger) Error(args ...interface{}) {
	l.logger.Error().Msgf(defaultArgsFormat(len(args)), args...)
}

func (l *RaftLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *RaftLogger) Warning(args ...interface{}) {
	l.logger.Warn().Msgf(defaultArgsFormat(len(args)), args...)
}

func (l *RaftLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *RaftLogger) Info(args ...interface{}) {
	l.logger.Info().Msgf(defaultArgsFormat(len(args)), args...)
}

func (l *RaftLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *RaftLogger) Debug(args ...interface{}) {
	l.logger.Debug().Msgf(defaultArgsFormat(len(args)), args...)
}

func (l *RaftLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func defaultArgsFormat(argc int) string {
	f := strings.Repeat("%s ", argc)
	if argc > 0 {
		f = f[:len(f)-1]
	}
	return f
}
