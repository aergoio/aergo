package log

import (
	"io"
	"sync"

	"github.com/mattn/go-colorable"
)

type syncWriter struct {
	sync.Mutex
	io.Writer
}

func newSyncWriter() io.Writer {
	w := colorable.NewColorableStderr()
	return &syncWriter{
		Writer: w,
	}
}
func (w *syncWriter) Write(p []byte) (n int, err error) {
	w.Lock()
	n, err = w.Writer.Write(p)
	w.Unlock()
	return n, err
}
