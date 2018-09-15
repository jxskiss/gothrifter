package thrift

import (
	"bytes"
	"github.com/matryer/is"
	"testing"
)

func TestBinaryProtocol(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	var p = NewProtocol(nil, DefaultOptions)
	p.Reset(&buf)

	var ok bool
	_, ok = p.Reader.(*binaryReader)
	is.True(ok)
	_, ok = p.Writer.(*binaryWriter)
	is.True(ok)

	var w = p // Writer
	var f float64 = 5.5
	w.WriteBool(true)
	w.WriteByte(2)
	w.WriteI16(3)
	w.WriteI32(4)
	w.WriteI64(5)
	w.WriteDouble(f)
	w.WriteFloat(float32(f))
	w.WriteString("6")
	w.WriteBinary([]byte("7"))
	w.WriteMessageBegin("header", CALL, 1)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}

	var r = p // Reader
	b, err := r.ReadBool()
	is.NoErr(err)
	is.True(b)

	n2, err := r.ReadByte()
	is.NoErr(err)
	is.Equal(n2, byte(2))

	n3, err := r.ReadI16()
	is.NoErr(err)
	is.True(n3 == 3)

	n4, err := r.ReadI32()
	is.NoErr(err)
	is.True(n4 == 4)

	n5, err := r.ReadI64()
	is.NoErr(err)
	is.True(n5 == 5)

	f6, err := r.ReadDouble()
	is.NoErr(err)
	is.Equal(f6, f)

	f7, err := r.ReadFloat()
	is.NoErr(err)
	is.Equal(f7, float32(f))

	s6, err := r.ReadString()
	is.NoErr(err)
	is.Equal(s6, "6")

	s7, err := r.ReadBinary()
	is.NoErr(err)
	is.Equal(s7, []byte("7"))

	name, tid, seq, err := r.ReadMessageBegin()
	is.NoErr(err)
	is.Equal(name, "header")
	is.True(tid == CALL)
	is.True(seq == 1)
}

func TestBinarySkip(t *testing.T) {
	is := is.New(t)

	var buf bytes.Buffer
	var p = NewProtocol(nil, DefaultOptions)
	p.Reset(&buf)

	var w = p // Writer
	var err error

	w.WriteFieldBegin("", STRUCT, 0)
	w.WriteStructBegin("")
	w.WriteFieldBegin("", BOOL, 1)
	w.WriteBool(false)
	w.WriteFieldBegin("", I08, 2)
	w.WriteByte(2)
	w.WriteFieldBegin("", DOUBLE, 3)
	w.WriteDouble(3.3)
	w.WriteFieldBegin("", I16, 4)
	w.WriteI16(4)
	w.WriteFieldBegin("", I32, 5)
	w.WriteI32(5)
	w.WriteFieldBegin("", I64, 6)
	w.WriteI64(6)
	w.WriteFieldBegin("", STRING, 7)
	w.WriteString("7")
	w.WriteFieldBegin("", MAP, 8)
	w.WriteMapBegin(I64, STRING, 1)
	w.WriteI64(81)
	w.WriteString("82")
	w.WriteFieldBegin("", SET, 9)
	w.WriteSetBegin(I32, 2)
	w.WriteI32(91)
	w.WriteI32(92)
	w.WriteFieldStop()
	w.WriteStructEnd()
	err = w.Flush()
	is.NoErr(err)

	rawBytes := append(([]byte)(nil), buf.Bytes()...)

	var r = p // Reader
	_, tp, id, err := r.ReadFieldBegin()
	is.NoErr(err)
	is.True(tp == STRUCT)
	is.True(id == 0)

	err = r.Skip(tp)
	is.NoErr(err)
	is.Equal(buf.Len(), 0)

	buf.Write(rawBytes[3:])
	skipped, err := r.ReadRaw(STRUCT)
	is.NoErr(err)
	is.Equal(skipped, rawBytes[3:])
	is.Equal(buf.Len(), 0)
}
