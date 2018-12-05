/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import "io"

type ListenAdded func(size int)

type MetricReader struct {
	rd io.Reader

	listeners []ListenAdded
}

func NewReader(rd io.Reader) *MetricReader {
	return &MetricReader{rd:rd}
}

func (r *MetricReader) AddListener(listener ListenAdded) {
	r.listeners = append(r.listeners, listener)
}

func (r *MetricReader) Read(p []byte) (n int, err error) {
	n, err = r.rd.Read(p)
	if err == nil {
		for _, listener := range r.listeners {
			listener(n)
		}
	}
	return n, err
}

type MetricWriter struct {
	wt io.Writer

	listeners []ListenAdded
}

func NewWriter(wt io.Writer) *MetricWriter {
	return &MetricWriter{wt:wt}
}

func (w *MetricWriter) Write(p []byte) (n int, err error) {
	n, err = w.wt.Write(p)
	if err == nil {
		for _, listener := range w.listeners {
			listener(n)
		}
	}
	return n, err
}

func (w *MetricWriter) AddListener(listener ListenAdded) {
	w.listeners = append(w.listeners, listener)
}
