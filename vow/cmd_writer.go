package vow

import (
	"io"
	"sync/atomic"
)

type cmdWriter struct {
	closed *int32
	w      io.Writer
}

func newCdmWriter(w io.Writer) *cmdWriter {
	return &cmdWriter{
		w:      w,
		closed: new(int32),
	}
}

func (cw *cmdWriter) Write(p []byte) (n int, err error) {
	if atomic.LoadInt32(cw.closed) == 0 {
		return cw.w.Write(p)
	}
	return 0, io.EOF
}

func (cw *cmdWriter) Close() error {
	atomic.StoreInt32(cw.closed, 1)
	return nil
}
