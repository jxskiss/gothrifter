package thrift2

import (
	"bufio"
	"io"
)

type bufReader struct {
	rd  *bufio.Reader
	tmp [10]byte
	raw []byte

	// for compact protocol
	fieldIdStack     []int16
	lastFieldId      int16
	pendingBoolField uint8

	prot *Protocol
}

func (b *bufReader) ReadByte() (c byte, err error) {
	if c, err = b.rd.ReadByte(); err != nil {
		return
	}
	if b.raw != nil {
		b.raw = append(b.raw, c)
	}
	return
}

func (b *bufReader) Read(p []byte) (n int, err error) {
	n, err = b.rd.Read(p)
	if err != nil {
		return
	}
	if b.raw != nil {
		b.raw = append(b.raw, p...)
	}
	return
}

func (b *bufReader) Peek(n int) ([]byte, error) {
	return b.rd.Peek(n)
}

func (b *bufReader) Reset(r io.Reader) {
	b.rd.Reset(r)
	b.raw = nil
	b.fieldIdStack = b.fieldIdStack[:0]
	b.lastFieldId = 0
	b.pendingBoolField = 0
}

type bufWriter struct {
	*bufio.Writer
	tmp [10]byte

	// for compact protocol
	fieldIdStack     []int16
	lastFieldId      int16
	boolFieldId      int16
	boolFieldPending bool

	prot *Protocol
}

func (b *bufWriter) Reset(w io.Writer) {
	b.Writer.Reset(w)
	b.fieldIdStack = b.fieldIdStack[:0]
	b.lastFieldId = 0
	b.boolFieldId = 0
	b.boolFieldPending = false
}
