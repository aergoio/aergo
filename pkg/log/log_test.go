package log

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogger_IsDebugEnabled(t *testing.T) {
	type fields struct {
		llog  *logrus.Logger
		entry *logrus.Entry
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "panic", fields: fields{NewLog(logrus.PanicLevel), nil}, want: false},
		{name: "fatal", fields: fields{NewLog(logrus.FatalLevel), nil}, want: false},
		{name: "error", fields: fields{NewLog(logrus.ErrorLevel), nil}, want: false},
		{name: "warn", fields: fields{NewLog(logrus.WarnLevel), nil}, want: false},
		{name: "info", fields: fields{NewLog(logrus.InfoLevel), nil}, want: false},
		{name: "debug", fields: fields{NewLog(logrus.DebugLevel), nil}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(TEST).WithCtx("test", tt.name)
			logger.SetLevel(tt.name)
			if got := logger.IsDebugEnabled(); got != tt.want {
				t.Errorf("Logger.IsDebugEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func NewLog(level logrus.Level) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(level)
	return logger
}
