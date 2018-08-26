package thrift2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var ErrMaxFrameSize = errors.New("thrift: max frame size exceeded")

type FramedTransport struct {
	transport io.ReadWriter
	wbuf      *bytes.Buffer
	wsize     [4]byte
	rsize     [4]byte
	rleft     int
	max       int
}

func NewFramedTransport(transport io.ReadWriter, max int) *FramedTransport {
	return &FramedTransport{
		transport: transport,
		wbuf:      bytes.NewBuffer(nil),
		max:       max,
	}
}

func (t *FramedTransport) Reset(rw io.ReadWriter) {
	t.transport = rw
	t.wbuf.Reset()
	t.rleft = 0
}

func (t *FramedTransport) Read(p []byte) (int, error) {
	if t.rleft <= 0 {
		if _, err := io.ReadFull(t.transport, t.rsize[:]); err != nil {
			return 0, err
		}
		n := int(binary.BigEndian.Uint32(t.rsize[:]))
		if n > t.max {
			return 0, ErrMaxFrameSize
		}
		t.rleft = n
	}
	if len(p) > t.rleft {
		p = p[:t.rleft]
	}
	n, err := t.transport.Read(p)
	t.rleft -= n
	return n, err
}

func (t *FramedTransport) Write(p []byte) (int, error) {
	if t.wbuf.Len()+len(p) > t.max {
		return 0, ErrMaxFrameSize
	}
	return t.wbuf.Write(p)
}

func (t *FramedTransport) Flush() error {
	binary.BigEndian.PutUint32(t.wsize[:], uint32(t.wbuf.Len()))
	if _, err := t.transport.Write(t.wsize[:]); err != nil {
		return err
	}
	_, err := t.wbuf.WriteTo(t.transport)
	return err
}
