package thrift2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var ErrMaxFrameSize = errors.New("thrift: max frame size exceeded")

type FramedReadWriter struct {
	rw    io.ReadWriter
	wbuf  bytes.Buffer
	wsize [4]byte
	rsize [4]byte
	rleft int

	maxFramesize int
}

func (t *FramedReadWriter) Reset(rw io.ReadWriter) {
	t.rw = rw
	t.wbuf.Reset()
	t.rleft = 0
}

func (t *FramedReadWriter) Read(p []byte) (int, error) {
	if t.rleft <= 0 {
		if _, err := io.ReadFull(t.rw, t.rsize[:]); err != nil {
			return 0, err
		}
		n := int(binary.BigEndian.Uint32(t.rsize[:]))
		if n > t.maxFramesize {
			return 0, ErrMaxFrameSize
		}
		t.rleft = n
	}
	if len(p) > t.rleft {
		p = p[:t.rleft]
	}
	n, err := t.rw.Read(p)
	t.rleft -= n
	return n, err
}

func (t *FramedReadWriter) Write(p []byte) (int, error) {
	if t.wbuf.Len()+len(p) > t.maxFramesize {
		return 0, ErrMaxFrameSize
	}
	return t.wbuf.Write(p)
}

func (t *FramedReadWriter) Flush() error {
	binary.BigEndian.PutUint32(t.wsize[:], uint32(t.wbuf.Len()))
	if _, err := t.rw.Write(t.wsize[:]); err != nil {
		return err
	}
	_, err := t.wbuf.WriteTo(t.rw)
	return err
}
