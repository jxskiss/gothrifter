package thrift2

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// Struct is the interface used to encapsulate a message that can be read and written to a protocol.
type Struct interface {
	Writable
	Readable
}

type Writable interface {
	Write(w Writer) error
}

type Readable interface {
	Read(r Reader) error
}

type flusher interface {
	Flush() error
}

type Reader interface {
	ReadMessageBegin() (name string, typeId MessageType, seqid int32, err error)
	ReadMessageEnd() error
	ReadStructBegin() (name string, err error)
	ReadStructEnd() error
	ReadFieldBegin() (name string, typeId Type, id int16, err error)
	ReadFieldEnd() error
	ReadMapBegin() (keyType Type, valueType Type, size int, err error)
	ReadMapEnd() error
	ReadListBegin() (elemType Type, size int, err error)
	ReadListEnd() error
	ReadSetBegin() (elemType Type, size int, err error)
	ReadSetEnd() error
	ReadBool() (value bool, err error)
	ReadByte() (value byte, err error)
	ReadI16() (value int16, err error)
	ReadI32() (value int32, err error)
	ReadI64() (value int64, err error)
	ReadDouble() (value float64, err error)
	ReadFloat() (value float32, err error)
	ReadString() (value string, err error)
	ReadBinary() (value []byte, err error)

	Skip(fieldType Type) (err error)
	ReadRaw(fieldType Type) (raw []byte, err error)
}

type Writer interface {
	WriteMessageBegin(name string, typeId MessageType, seqid int32) error
	WriteMessageEnd() error
	WriteStructBegin(name string) error
	WriteStructEnd() error
	WriteFieldBegin(name string, typeId Type, id int16) error
	WriteFieldEnd() error
	WriteFieldStop() error
	WriteMapBegin(keyType Type, valueType Type, size int) error
	WriteMapEnd() error
	WriteListBegin(elemType Type, size int) error
	WriteListEnd() error
	WriteSetBegin(elemType Type, size int) error
	WriteSetEnd() error
	WriteBool(value bool) error
	WriteByte(value byte) error
	WriteI16(value int16) error
	WriteI32(value int32) error
	WriteI64(value int64) error
	WriteDouble(value float64) error
	WriteFloat(value float32) error
	WriteString(value string) error
	WriteBinary(value []byte) error

	Flush() (err error)
}

type Protocol struct {
	Reader
	Writer
	protoID    ProtocolID
	compactVer int // COMPACT_VERSION / COMPACT_VERSION_BE

	bufr *bufReader
	bufw *bufWriter

	// Underlying transport, there will be at most one being non-nil.
	header *HeaderTransport
	frw    *FramedTransport
	// Flush function of the underlying transport, nil for raw socket.
	flush func() error
}

func NewProtocol(rw io.ReadWriter, opts options) *Protocol {
	var p = &Protocol{protoID: opts.protoID, compactVer: COMPACT_VERSION_BE}
	if opts.header {
		p.header = NewHeaderTransport(rw)
		p.header.protoID = opts.protoID
		if opts.maxframesize > 0 {
			p.header.maxFramesize = opts.maxframesize
		}
		rw = p.header
	} else if opts.maxframesize > 0 {
		p.frw = NewFramedTransport(rw, opts.maxframesize)
		rw = p.frw
	}
	if f, ok := rw.(flusher); ok {
		p.flush = func() error { return f.Flush() }
	}
	p.bufr = &bufReader{
		rd:           bufio.NewReaderSize(rw, opts.rbufsz),
		fieldIdStack: make([]int16, 0, 4),
		prot:         p,
	}
	p.bufw = &bufWriter{
		Writer:       bufio.NewWriterSize(rw, opts.wbufsz),
		fieldIdStack: make([]int16, 0, 4),
		prot:         p,
	}
	p.ResetProtocol()
	return p
}

func (p *Protocol) ResetProtocol() error {
	if p.Reader != nil && p.header != nil && p.protoID == p.header.protoID {
		return nil
	}
	if p.header != nil {
		p.protoID = p.header.protoID
	}
	switch p.protoID {
	case ProtocolIDBinary:
		p.Reader = (*binaryReader)(p.bufr)
		p.Writer = (*binaryWriter)(p.bufw)
	case ProtocolIDCompact:
		p.Reader = (*compactReader)(p.bufr)
		p.Writer = (*compactWriter)(p.bufw)
	default:
		return fmt.Errorf("unknow protocol id: %#x", p.protoID)
	}
	return nil
}

func (p *Protocol) Reset(rw io.ReadWriter) {
	switch {
	case p.header != nil: // header transport
		p.resetHeader(rw)
	case p.frw != nil: // framed transport
		p.resetFramed(rw)
	default: // raw socket
		p.resetRaw(rw)
	}
}

func (p *Protocol) resetHeader(rw io.ReadWriter) {
	header := NewHeaderTransport(rw)
	header.protoID = p.header.protoID
	header.maxFramesize = p.header.maxFramesize
	p.header = header
	p.bufr.Reset(p.header)
	p.bufw.Reset(p.header)
}

func (p *Protocol) resetFramed(rw io.ReadWriter) {
	p.frw.Reset(rw)
	p.bufr.Reset(p.frw)
	p.bufw.Reset(p.frw)
}

func (p *Protocol) resetRaw(rw io.ReadWriter) {
	p.bufr.Reset(rw)
	p.bufw.Reset(rw)
}

func (p *Protocol) preReadMessageBegin() (protoID ProtocolID, err error) {
	if p.header != nil { // header transport
		if err = p.header.ResetProtocol(); err != nil {
			return
		}
		if err = p.ResetProtocol(); err != nil {
			return
		}
		return p.header.protoID, nil
	}

	// auto detect binary & compact protocol
	var b []byte
	if b, err = p.bufr.Peek(2); err != nil {
		return
	}
	if b[0] == COMPACT_PROTOCOL_ID { // compact protocol
		version := int(b[1] & COMPACT_VERSION_MASK)
		if version != COMPACT_VERSION && version != COMPACT_VERSION_BE {
			err = fmt.Errorf("unexpected compact version %02x", version)
			return
		}
		if p.protoID != ProtocolIDCompact || p.compactVer != version {
			p.protoID = ProtocolIDCompact
			p.compactVer = version
			if err = p.ResetProtocol(); err != nil {
				return
			}
		}
	} else { // binary protocol
		if p.protoID != ProtocolIDBinary {
			p.protoID = ProtocolIDBinary
			if err = p.ResetProtocol(); err != nil {
				return
			}
		}
	}
	return p.protoID, nil
}

func (p *Protocol) preWriteMessageBegin(name string, typeId MessageType, seqid int32) error {
	p.ResetProtocol()
	if p.header != nil {
		if typeId == CALL || typeId == ONEWAY {
			p.header.SetSeqID(uint32(seqid))
		}
	}
	return nil
}

func (p *Protocol) postFlush() (err error) {
	if p.flush != nil {
		return p.flush()
	}
	return nil
}

// Header manipulation functions

func (p *Protocol) SetIdentity(identity string) {
	if p.header != nil {
		p.header.SetIdentity(identity)
	}
}

func (p *Protocol) Identity() (identity string) {
	if p.header != nil {
		identity = p.header.Identity()
	}
	return
}

func (p *Protocol) PeerIdentity() (identity string) {
	if p.header != nil {
		identity = p.header.PeerIdentity()
	}
	return
}

func (p *Protocol) SetPersistentHeader(key, value string) {
	if p.header != nil {
		p.header.SetPersistentHeader(key, value)
	}
}

func (p *Protocol) PersistentHeader(key string) (string, bool) {
	if p.header != nil {
		return p.header.PersistentHeader(key)
	}
	return "", false
}

func (p *Protocol) PersistentHeaders() map[string]string {
	if p.header != nil {
		return p.header.PersistentHeaders()
	}
	return nil
}

func (p *Protocol) ClearPersistentHeaders() {
	if p.header != nil {
		p.header.ClearPersistentHeaders()
	}
}

func (p *Protocol) SetHeader(key, value string) {
	if p.header != nil {
		p.header.SetHeader(key, value)
	}
}

func (p *Protocol) Header(key string) (string, bool) {
	if p.header != nil {
		return p.header.Header(key)
	}
	return "", false
}

func (p *Protocol) Headers() map[string]string {
	if p.header != nil {
		return p.header.Headers()
	}
	return nil
}

func (p *Protocol) ClearHeaders() {
	if p.header != nil {
		p.header.ClearHeaders()
	}
}

func (p *Protocol) ReadHeader(key string) (string, bool) {
	if p.header != nil {
		return p.header.ReadHeader(key)
	}
	return "", false
}

func (p *Protocol) ReadHeaders() map[string]string {
	if p.header != nil {
		return p.header.ReadHeaders()
	}
	return nil
}

func (p *Protocol) ProtocolID() ProtocolID {
	return p.protoID
}

func (p *Protocol) AddTransform(trans TransformID) error {
	if p.header != nil {
		return p.header.AddTransform(trans)
	}
	return errors.New("not header transport") // TODO
}

// The maximum recursive depth the skip() function will traverse
const DEFAULT_RECURSION_DEPTH = 64

// Skips over the next data element from the provided input Protocol object.
func SkipDefaultDepth(r Reader, typeId Type) (err error) {
	return Skip(r, typeId, DEFAULT_RECURSION_DEPTH)
}

// Skips over the next data element from the provided input Protocol object.
func Skip(self Reader, fieldType Type, maxDepth int) (err error) {

	if maxDepth <= 0 {
		return ErrDepthExceeded
	}

	switch fieldType {
	case STOP:
		return
	case BOOL:
		_, err = self.ReadBool()
		return
	case BYTE:
		_, err = self.ReadByte()
		return
	case I16:
		_, err = self.ReadI16()
		return
	case I32:
		_, err = self.ReadI32()
		return
	case I64:
		_, err = self.ReadI64()
		return
	case DOUBLE:
		_, err = self.ReadDouble()
		return
	case FLOAT:
		_, err = self.ReadFloat()
		return
	case STRING:
		_, err = self.ReadString()
		return
	case STRUCT:
		if _, err = self.ReadStructBegin(); err != nil {
			return err
		}
		for {
			_, typeId, _, _ := self.ReadFieldBegin()
			if typeId == STOP {
				break
			}
			err := Skip(self, typeId, maxDepth-1)
			if err != nil {
				return err
			}
			self.ReadFieldEnd()
		}
		return self.ReadStructEnd()
	case MAP:
		keyType, valueType, size, err := self.ReadMapBegin()
		if err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			err := Skip(self, keyType, maxDepth-1)
			if err != nil {
				return err
			}
			self.Skip(valueType)
		}
		return self.ReadMapEnd()
	case SET:
		elemType, size, err := self.ReadSetBegin()
		if err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			err := Skip(self, elemType, maxDepth-1)
			if err != nil {
				return err
			}
		}
		return self.ReadSetEnd()
	case LIST:
		elemType, size, err := self.ReadListBegin()
		if err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			err := Skip(self, elemType, maxDepth-1)
			if err != nil {
				return err
			}
		}
		return self.ReadListEnd()
	}
	return nil
}

func ReadRaw(br *bufReader, discard func() error) (raw []byte, err error) {
	raw, br.raw = br.raw, make([]byte, 0, 32)
	err = discard()
	raw, br.raw = br.raw, raw
	return
}
