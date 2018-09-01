package thrift2

import (
	"encoding/binary"
	"math"
	"unsafe"
)

type binaryReader bufReader

func (r *binaryReader) ReadByte() (c byte, err error) {
	return (*bufReader)(r).ReadByte()
}

func (r *binaryReader) Read(p []byte) (n int, err error) {
	return (*bufReader)(r).Read(p)
}

func (r *binaryReader) ReadMessageBegin() (name string, typeId MessageType, seqid int32, err error) {
	var protoID ProtocolID
	if protoID, err = r.prot.preReadMessageBegin(); err != nil {
		return
	}
	// the protocol may be changed during preReadMessageBegin
	if protoID != ProtocolIDBinary {
		return r.prot.ReadMessageBegin()
	}

	var n int32
	if n, err = r.ReadI32(); err != nil {
		return
	}
	typeId = MessageType(uint32(n) & 0x0ff)
	if version := uint32(n) & BinaryVersionMask; version != BinaryVersion1 {
		err = ErrBinaryVersion
		return
	}
	if name, err = r.ReadString(); err != nil {
		return
	}
	if seqid, err = r.ReadI32(); err != nil {
		return
	}
	if typeId == EXCEPTION {
		var ex ApplicationException
		if err = ex.Read(r); err == nil {
			err = &ex
		}
	}
	return
}

func (r *binaryReader) ReadMessageEnd() error {
	return nil
}

func (r *binaryReader) ReadStructBegin() (name string, err error) {
	return
}

func (r *binaryReader) ReadStructEnd() error {
	return nil
}

func (r *binaryReader) ReadFieldBegin() (name string, typeId Type, id int16, err error) {
	t, err := r.ReadByte()
	if err != nil {
		return
	}
	typeId = Type(t)
	if typeId != STOP {
		id, err = r.ReadI16()
	}
	return
}

func (r *binaryReader) ReadFieldEnd() error {
	return nil
}

func (r *binaryReader) ReadMapBegin() (keyType Type, valueType Type, size int, err error) {
	b := r.tmp[:6]
	if _, err = r.Read(b); err != nil {
		return
	}
	keyType, valueType = Type(b[0]), Type(b[1])
	size = int(uint32(b[5]) | uint32(b[4])<<8 | uint32(b[3])<<16 | uint32(b[2])<<24)
	return
}

func (r *binaryReader) ReadMapEnd() error {
	return nil
}

func (r *binaryReader) ReadListBegin() (elemType Type, size int, err error) {
	return r.readCollectionBegin()
}

func (r *binaryReader) ReadListEnd() error {
	return nil
}

func (r *binaryReader) ReadSetBegin() (elemType Type, size int, err error) {
	return r.readCollectionBegin()
}

func (r *binaryReader) ReadSetEnd() error {
	return nil
}

func (r *binaryReader) readCollectionBegin() (elemType Type, size int, err error) {
	b := r.tmp[:5]
	if _, err = r.Read(b); err != nil {
		return
	}
	elemType = Type(b[0])
	size = int(uint32(b[4]) | uint32(b[3])<<8 | uint32(b[2])<<16 | uint32(b[1])<<24)
	return
}

func (r *binaryReader) ReadBool() (value bool, err error) {
	if v, err := r.ReadByte(); err == nil {
		value = v > 0
	}
	return
}

func (r *binaryReader) ReadI16() (value int16, err error) {
	b := r.tmp[:2]
	if _, err = r.Read(b); err == nil {
		value = int16(uint16(b[1]) | uint16(b[0])<<8)
	}
	return
}

func (r *binaryReader) ReadI32() (value int32, err error) {
	b := r.tmp[:4]
	if _, err = r.Read(b); err == nil {
		value = int32(uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24)
	}
	return
}

func (r *binaryReader) ReadI64() (value int64, err error) {
	b := r.tmp[:8]
	if _, err = r.Read(b); err == nil {
		value = int64(
			uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
				uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56)
	}
	return
}

func (r *binaryReader) ReadDouble() (value float64, err error) {
	b := r.tmp[:8]
	if _, err = r.Read(b); err == nil {
		value = math.Float64frombits(binary.BigEndian.Uint64(b))
	}
	return
}

func (r *binaryReader) ReadFloat() (value float32, err error) {
	b := r.tmp[:4]
	if _, err = r.Read(b); err == nil {
		value = math.Float32frombits(binary.BigEndian.Uint32(b))
	}
	return
}

func (r *binaryReader) ReadString() (value string, err error) {
	var length int32
	if length, err = r.ReadI32(); err != nil {
		return
	}
	var b = make([]byte, length)
	if _, err = r.Read(b); err != nil {
		return
	}
	// no need to copy the memory
	return *(*string)(unsafe.Pointer(&b)), nil
}

func (r *binaryReader) ReadBinary() (value []byte, err error) {
	var length int32
	if length, err = r.ReadI32(); err != nil {
		return
	}
	value = make([]byte, length)
	_, err = r.Read(value)
	return
}

func (r *binaryReader) Skip(fieldType Type) (err error) {
	return SkipDefaultDepth(r, fieldType)
}

func (r *binaryReader) ReadRaw(fieldType Type) (raw []byte, err error) {
	return ReadRaw((*bufReader)(r), func() error { return r.Skip(fieldType) })
}

type binaryWriter bufWriter

func (w *binaryWriter) WriteMessageBegin(name string, typeId MessageType, seqid int32) error {
	if err := w.prot.preWriteMessageBegin(name, typeId, seqid); err != nil {
		return err
	}
	verAndType := uint32(BinaryVersion1) | uint32(typeId)
	if err := w.WriteI32(int32(verAndType)); err != nil {
		return err
	}
	if err := w.WriteString(name); err != nil {
		return err
	}
	return w.WriteI32(seqid)
}

func (w *binaryWriter) WriteMessageEnd() error {
	return nil
}

func (w *binaryWriter) WriteStructBegin(name string) error {
	return nil
}

func (w *binaryWriter) WriteStructEnd() error {
	return nil
}

func (w *binaryWriter) WriteFieldBegin(name string, typeId Type, id int16) error {
	b := w.tmp[:3]
	b[0] = byte(typeId)
	b[1], b[2] = byte(id>>8), byte(id)
	_, err := w.Write(b)
	return err
}

func (w *binaryWriter) WriteFieldEnd() error {
	return nil
}

func (w *binaryWriter) WriteFieldStop() error {
	return w.WriteByte(STOP)
}

func (w *binaryWriter) WriteMapBegin(keyType Type, valueType Type, size int) error {
	b := w.tmp[:6]
	b[0], b[1] = byte(keyType), byte(valueType)
	b[2], b[3], b[4], b[5] = byte(size>>24), byte(size>>16), byte(size>>8), byte(size)
	_, err := w.Write(b)
	return err
}

func (w *binaryWriter) WriteMapEnd() error {
	return nil
}

func (w *binaryWriter) WriteListBegin(elemType Type, size int) error {
	return w.writeCollectionBegin(elemType, size)
}

func (w *binaryWriter) WriteListEnd() error {
	return nil
}

func (w *binaryWriter) WriteSetBegin(elemType Type, size int) error {
	return w.writeCollectionBegin(elemType, size)
}

func (w *binaryWriter) WriteSetEnd() error {
	return nil
}

func (w *binaryWriter) writeCollectionBegin(elemType Type, size int) error {
	b := w.tmp[:5]
	b[0] = byte(elemType)
	b[1], b[2], b[3], b[4] = byte(size>>24), byte(size>>16), byte(size>>8), byte(size)
	_, err := w.Write(b)
	return err
}

func (w *binaryWriter) WriteBool(value bool) error {
	if value {
		return w.WriteByte(byte(1))
	}
	return w.WriteByte(byte(0))
}

func (w *binaryWriter) WriteI16(value int16) error {
	if err := w.WriteByte(byte(value) >> 8); err != nil {
		return err
	}
	return w.WriteByte(byte(value))
}

func (w *binaryWriter) WriteI32(value int32) error {
	b := w.tmp[:4]
	b[0], b[1], b[2], b[3] = byte(value>>24), byte(value>>16), byte(value>>8), byte(value)
	_, err := w.Write(b)
	return err
}

func (w *binaryWriter) WriteI64(value int64) error {
	b := w.tmp[:8]
	b[0], b[1], b[2], b[3] = byte(value>>56), byte(value>>48), byte(value>>40), byte(value>>32)
	b[4], b[5], b[6], b[7] = byte(value>>24), byte(value>>16), byte(value>>8), byte(value)
	_, err := w.Write(b)
	return err
}

func (w *binaryWriter) WriteDouble(value float64) error {
	b := w.tmp[:8]
	binary.BigEndian.PutUint64(b, math.Float64bits(value))
	_, err := w.Write(b)
	return err
}

func (w *binaryWriter) WriteFloat(value float32) error {
	return w.WriteI32(int32(math.Float32bits(value)))
}

func (w *binaryWriter) WriteString(value string) error {
	if err := w.WriteI32(int32(len(value))); err != nil {
		return err
	}
	_, err := w.Writer.WriteString(value)
	return err
}

func (w *binaryWriter) WriteBinary(value []byte) error {
	if err := w.WriteI32(int32(len(value))); err != nil {
		return err
	}
	_, err := w.Write(value)
	return err
}

func (w *binaryWriter) Flush() error {
	if err := w.Writer.Flush(); err != nil {
		return err
	}
	return w.prot.postFlush()
}
