package thrift

import (
	"bytes"
	"sync"
)

var DefaultProtocolPool = sync.Pool{
	New: func() interface{} {
		return NewProtocol(nil, DefaultOptions)
	},
}

func Marshal(val Writable) ([]byte, error) {
	var p = DefaultProtocolPool.Get().(*Protocol)
	_ = p.UseBinary()
	defer DefaultProtocolPool.Put(p)

	var buf bytes.Buffer
	p.Reset(&buf)
	if err := val.Write(p); err != nil {
		return nil, err
	}
	if err := p.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MarshalCompact(val Writable) ([]byte, error) {
	var p = DefaultProtocolPool.Get().(*Protocol)
	_ = p.UseCompact(COMPACT_VERSION)
	defer DefaultProtocolPool.Put(p)

	var buf bytes.Buffer
	p.Reset(&buf)
	if err := val.Write(p); err != nil {
		return nil, err
	}
	if err := p.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, val Readable) error {
	var p = DefaultProtocolPool.Get().(*Protocol)
	_ = p.UseBinary()
	defer DefaultProtocolPool.Put(p)

	var buf = bytes.NewBuffer(data)
	p.Reset(buf)
	return val.Read(p)
}

func UnmarshalCompact(data []byte, val Readable) error {
	var p = DefaultProtocolPool.Get().(*Protocol)
	_ = p.UseCompact(COMPACT_VERSION)
	defer DefaultProtocolPool.Put(p)

	var buf = bytes.NewBuffer(data)
	p.Reset(buf)
	return val.Read(p)
}

type Serializer struct {
	buf  *bytes.Buffer
	prot *Protocol
}

// NewSerializer create a new serializer using the binary protocol.
func NewSerializer() *Serializer {
	s := &Serializer{buf: &bytes.Buffer{}}
	s.prot = NewProtocol(s.buf, DefaultOptions)
	return s
}

// NewCompactSerializer create a new serializer using the compact protocol.
func NewCompactSerializer() *Serializer {
	s := &Serializer{buf: &bytes.Buffer{}}
	s.prot = NewProtocol(s.buf, WithCompact()(DefaultOptions))
	return s
}

// WriteString writes msg to the serializer and returns it as a string.
func (s *Serializer) WriteString(msg Writable) (str string, err error) {
	s.buf.Reset()
	s.prot.Reset(s.buf)
	if err = msg.Write(s.prot); err != nil {
		return
	}
	if err = s.prot.Flush(); err != nil {
		return
	}
	return s.buf.String(), nil
}

// Write writes msg to the serializer and returns it as a byte slice.
func (s *Serializer) Write(msg Writable) (b []byte, err error) {
	s.buf.Reset()
	s.prot.Reset(s.buf)
	if err = msg.Write(s.prot); err != nil {
		return
	}
	if err = s.prot.Flush(); err != nil {
		return
	}
	return s.buf.Bytes(), nil
}

type Deserializer struct {
	buf  *bytes.Buffer
	prot *Protocol
}

// NewDeserializer create a new deserializer using the binary protocol.
func NewDeserializer() *Deserializer {
	ds := &Deserializer{buf: &bytes.Buffer{}}
	ds.prot = NewProtocol(ds.buf, DefaultOptions)
	return ds
}

// NewCompactDeserializer create a new deserializer using the compact protocol.
func NewCompactDeserializer() *Deserializer {
	ds := &Deserializer{buf: &bytes.Buffer{}}
	ds.prot = NewProtocol(ds.buf, WithCompact()(DefaultOptions))
	return ds
}

func (ds *Deserializer) ReadString(msg Readable, s string) (err error) {
	ds.buf.Reset()
	ds.prot.Reset(ds.buf)
	if _, err = ds.buf.WriteString(s); err != nil {
		return err
	}
	return msg.Read(ds.prot)
}

func (ds *Deserializer) Read(msg Readable, b []byte) (err error) {
	ds.buf.Reset()
	ds.prot.Reset(ds.buf)
	if _, err := ds.buf.Write(b); err != nil {
		return err
	}
	return msg.Read(ds.prot)
}
