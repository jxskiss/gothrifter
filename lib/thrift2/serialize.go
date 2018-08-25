package thrift2

import "bytes"

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

// WriteString writes msg to the serializer and returns it as a string.
func (s *Serializer) WriteString(msg Struct) (str string, err error) {
	s.buf.Reset()
	if err = msg.Write(s.prot); err != nil {
		return
	}
	if err = s.prot.Flush(); err != nil {
		return
	}
	return s.buf.String(), nil
}

// Write writes msg to the serializer and returns it as a byte slice.
func (s *Serializer) Write(msg Struct) (b []byte, err error) {
	s.buf.Reset()
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

func NewDeserialer() *Deserializer {
	ds := &Deserializer{buf: &bytes.Buffer{}}
	ds.prot = NewProtocol(ds.buf, options{})
	return ds
}

func (ds *Deserializer) ReadString(msg Struct, s string) (err error) {
	if _, err = ds.buf.WriteString(s); err != nil {
		return err
	}
	return msg.Read(ds.prot)
}

func (ds *Deserializer) Read(msg Struct, b []byte) (err error) {
	if _, err := ds.buf.Write(b); err != nil {
		return err
	}
	return msg.Read(ds.prot)
}
