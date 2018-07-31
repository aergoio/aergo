package log

import (
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	defaultTimestampFormat = "01-02|15:04:05.000"
)

type DebugFormatter struct {
	logrus.TextFormatter
}

func NewDebugFormatter() *DebugFormatter {
	return &DebugFormatter{
		TextFormatter: logrus.TextFormatter{
			ForceColors:      isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd()),
			DisableTimestamp: false,
			FullTimestamp:    true,
			TimestampFormat:  defaultTimestampFormat,
		},
	}
}
