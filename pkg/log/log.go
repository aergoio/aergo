/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package log

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

var defaultLevel = logrus.DebugLevel

var (
	loggers     []*Logger
	loggersLock sync.Mutex
)

func addLogger(logger *Logger) {
	loggersLock.Lock()
	loggers = append(loggers, logger)
	loggersLock.Unlock()
}

func setLoggerLevels() {
	loggersLock.Lock()
	for _, v := range loggers {
		v.setLevel(moduleLevels[v.module])
	}
	defer loggersLock.Unlock()

}

type Logger struct {
	entry  *logrus.Entry
	ctx    []interface{}
	module Module
}

func NewLogger(module Module) *Logger {
	lLogger := &logrus.Logger{
		Level:     moduleLevels[module],
		Out:       newSyncWriter(),
		Formatter: NewDebugFormatter(),
	}
	logger := &Logger{
		entry:  lLogger.WithField("module", module.String()),
		module: module,
	}
	addLogger(logger)

	return logger
}

func (l *Logger) Debug(msg ...interface{}) {
	if l.level() >= logrus.DebugLevel {
		l.withFields().Debug(msg...)
	}
}

func (l *Logger) Info(msg ...interface{}) {
	if l.level() >= logrus.InfoLevel {
		l.withFields().Info(msg...)
	}
}

func (l *Logger) Warn(msg ...interface{}) {
	if l.level() >= logrus.WarnLevel {
		l.withFields().Warn(msg...)
	}
}

func (l *Logger) Error(msg ...interface{}) {
	if l.level() >= logrus.ErrorLevel {
		l.withFields().Error(msg...)
	}
}

func (l *Logger) Fatal(msg ...interface{}) {
	if l.level() >= logrus.FatalLevel {
		l.withFields().Fatal(msg...)
	}
}

func (l *Logger) Debugf(formattedMsg string, args ...interface{}) {
	if l.level() >= logrus.DebugLevel {
		l.withFields().Debugf(formattedMsg, args...)
	}
}

func (l *Logger) Infof(formattedMsg string, args ...interface{}) {
	if l.level() >= logrus.InfoLevel {
		l.withFields().Infof(formattedMsg, args...)
	}
}

func (l *Logger) Warnf(formattedMsg string, args ...interface{}) {
	if l.level() >= logrus.WarnLevel {
		l.withFields().Warnf(formattedMsg, args...)
	}
}

func (l *Logger) Errorf(formattedMsg string, args ...interface{}) {
	if l.level() >= logrus.ErrorLevel {
		l.withFields().Errorf(formattedMsg, args...)
	}
}

func (l *Logger) Fatalf(formattedMsg string, args ...interface{}) {
	if l.level() >= logrus.FatalLevel {
		l.withFields().Fatalf(formattedMsg, args...)
	}
}

func (l *Logger) WithCtx(keyValues ...interface{}) *Logger {
	ctx := make([]interface{}, len(l.ctx)+len(keyValues))
	copy(ctx, l.ctx[:])
	copy(ctx[len(l.ctx):], keyValues[:])
	return &Logger{
		entry:  l.entry,
		ctx:    ctx,
		module: l.module,
	}
}

func (l *Logger) withFields() *logrus.Entry {
	ctx := l.ctx
	s := len(ctx)
	fields := make(logrus.Fields, len(l.entry.Data)+s/2)
	for i := 0; i < s; i += 2 {
		switch v := ctx[i].(type) {
		case string:
			fields[v] = ctx[i+1]
		}
	}
	for k, v := range l.entry.Data {
		fields[k] = v
	}
	var ok bool
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	fields["src"] = fmt.Sprintf("%s:%d", file, line)
	return l.entry.WithFields(fields)
}

func (l *Logger) SetLevel(level string) {
	n, err := logrus.ParseLevel(level)
	if err == nil {
		l.setLevel(n)
	} else {
		l.Error(err)
		l.setLevel(defaultLevel)
	}
}

func (l *Logger) setLevel(level logrus.Level) {
	l.entry.Logger.SetLevel(level)
}

func (l *Logger) Level() string {
	return l.entry.Logger.Level.String()
}

func (l *Logger) level() logrus.Level {
	return l.entry.Logger.Level
}

func (l *Logger) IsDebugEnabled() bool {
	return l.level() >= logrus.DebugLevel
}
