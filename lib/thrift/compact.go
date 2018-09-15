package thrift

import (
	"encoding/binary"
	"math"
	"unsafe"
)

type compactReader bufReader

func (r *compactReader) version() byte {
	return byte(r.prot.compactVer)
}

func (r *compactReader) ReadByte() (c byte, err error) {
	return (*bufReader)(r).ReadByte()
}

func (r *compactReader) Read(p []byte) (n int, err error) {
	return (*bufReader)(r).Read(p)
}

func (r *compactReader) ReadMessageBegin() (name string, typeId MessageType, seqid int32, err error) {
	var protoID ProtocolID
	if protoID, err = r.prot.preReadMessageBegin(); err != nil {
		return
	}
	// the protocol may be changed during preReadMessageBegin
	if protoID != ProtocolIDCompact {
		return r.prot.ReadMessageBegin()
	}

	b := r.tmp[:2]
	if _, err = r.Read(b); err != nil {
		return
	}
	verAndType := b[1]
	typeId = MessageType((verAndType >> 5) & COMPACT_TYPE_BITS)
	// version has already been checked in preReadMessageBegin, don't need to check again
	if seqid, err = r.readVarInt32(); err != nil {
		return
	}
	if name, err = r.ReadString(); err != nil {
		return
	}
	return
}

func (r *compactReader) ReadMessageEnd() error {
	return nil
}

func (r *compactReader) ReadStructBegin() (name string, err error) {
	r.fieldIdStack = append(r.fieldIdStack, r.lastFieldId)
	r.lastFieldId = 0
	return
}

// Doesn't actually consume any wire data, just remove the last field id
// for this struct from the field stack.
func (r *compactReader) ReadStructEnd() error {
	// consume the last field we read off the wire.
	r.lastFieldId = r.fieldIdStack[len(r.fieldIdStack)-1]
	r.fieldIdStack = r.fieldIdStack[:len(r.fieldIdStack)-1]
	return nil
}

func (r *compactReader) ReadFieldBegin() (name string, typeId Type, id int16, err error) {
	t, err := r.ReadByte()
	if err != nil {
		return
	}
	// if it's a stop, then we can return immediately, as the struct is over.
	if (t & 0x0f) == STOP {
		return "", STOP, 0, nil
	}

	// mask off the 4 MSB of the type header. it could contain a field id delta.
	modifier := int16((t & 0xf0) >> 4)
	if modifier == 0 {
		// not a delta. look ahead for the zigzag varint field id.
		id, err = r.ReadI16()
		if err != nil {
			return
		}
	} else {
		// has a delta. add the delta to the last read field id.
		id = int16(r.lastFieldId) + modifier
	}
	typeId = compactType(t & 0x0f).toType()

	// if this happens to be a boolean field, the value is encoded in the type
	if typeId == BOOL {
		r.pendingBoolField = t
	}

	// push the new field onto the field stack so we can keep the deltas going.
	r.lastFieldId = id
	return
}

func (r *compactReader) isBoolType(b byte) bool {
	return (b&0x0f) == COMPACT_BOOLEAN_TRUE || (b&0x0f) == COMPACT_BOOLEAN_FALSE
}

func (r *compactReader) ReadFieldEnd() error {
	return nil
}

func (r *compactReader) ReadMapBegin() (keyType Type, valueType Type, size int, err error) {
	size32, e := r.readVarInt32()
	if e != nil {
		err = e // TODO
		return
	}
	if size32 < 0 {
		err = ErrDataLength
		return
	}
	size = int(size32)
	keyAndValueType, e := r.ReadByte()
	if e != nil {
		err = e // TODO
		return
	}
	keyType = compactType(keyAndValueType >> 4).toType()
	valueType = compactType(keyAndValueType & 0xf).toType()
	return
}

func (r *compactReader) ReadMapEnd() error {
	return nil
}

func (r *compactReader) ReadListBegin() (elemType Type, size int, err error) {
	return r.readCollectionBegin()
}

func (r *compactReader) ReadListEnd() error {
	return nil
}

func (r *compactReader) ReadSetBegin() (elemType Type, size int, err error) {
	return r.readCollectionBegin()
}

func (r *compactReader) ReadSetEnd() error {
	return nil
}

func (r *compactReader) readCollectionBegin() (elemType Type, size int, err error) {
	var lenAndType byte
	if lenAndType, err = r.ReadByte(); err != nil {
		return
	}
	size = int((lenAndType >> 4) & 0x0f)
	if size == 15 {
		var size2 int32
		if size2, err = r.readVarInt32(); err != nil {
			return
		}
		size = int(size2)
	}
	elemType = compactType(lenAndType).toType()
	return
}

func (r *compactReader) ReadBool() (value bool, err error) {
	if r.pendingBoolField == 0 {
		v, err := r.ReadByte()
		if err != nil {
			return false, err
		}
		return v == 1, nil
	}
	return r.pendingBoolField == 1, nil
}

func (r *compactReader) ReadI16() (value int16, err error) {
	var i32 int32
	if i32, err = r.ReadI32(); err != nil {
		return
	}
	return int16(i32), nil
}

func (r *compactReader) ReadI32() (value int32, err error) {
	var i64 int64
	if i64, err = r.readVarInt64(); err == nil {
		u32 := uint32(i64)
		value = int32(u32>>1) ^ -(int32(i64) & 1)
	}
	return
}

func (r *compactReader) readVarInt32() (value int32, err error) {
	v, err := r.readVarInt64()
	return int32(v), err
}

func (r *compactReader) ReadI64() (value int64, err error) {
	var i64 int64
	if i64, err = r.readVarInt64(); err == nil {
		u64 := uint64(i64)
		value = int64(u64>>1) ^ -(i64 & 1)
	}
	return
}

func (r *compactReader) readVarInt64() (value int64, err error) {
	shift := uint(0)
	result := int64(0)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= int64(b&0x7f) << shift
		if (b & 0x80) != 0x80 {
			break
		}
		shift += 7
	}
	return result, nil
}

func (r *compactReader) ReadDouble() (value float64, err error) {
	b := r.tmp[:8]
	if _, err = r.Read(b); err != nil {
		return
	}
	if r.version() == COMPACT_VERSION {
		return math.Float64frombits(binary.LittleEndian.Uint64(b)), nil
	} else {
		return math.Float64frombits(binary.BigEndian.Uint64(b)), nil
	}
}

func (r *compactReader) ReadFloat() (value float32, err error) {
	b := r.tmp[:4]
	if _, err = r.Read(b); err == nil {
		value = math.Float32frombits(binary.BigEndian.Uint32(b))
	}
	return
}

func (r *compactReader) ReadString() (value string, err error) {
	var length int32
	if length, err = r.readVarInt32(); err != nil {
		return
	}
	var b = make([]byte, length)
	if _, err = r.Read(b); err != nil {
		return
	}
	// no need to copy the memory
	return *(*string)(unsafe.Pointer(&b)), nil
}

func (r *compactReader) ReadBinary() (value []byte, err error) {
	var length int32
	if length, err = r.readVarInt32(); err != nil {
		return
	}
	value = make([]byte, length)
	_, err = r.Read(value)
	return
}

func (r *compactReader) Skip(fieldType Type) (err error) {
	return SkipDefaultDepth(r, fieldType)
}

func (r *compactReader) ReadRaw(fieldType Type) (raw []byte, err error) {
	return ReadRaw((*bufReader)(r), func() error { return r.Skip(fieldType) })
}

type compactWriter bufWriter

func (w *compactWriter) version() byte {
	return byte(w.prot.compactVer)
}

func (w *compactWriter) WriteMessageBegin(name string, typeId MessageType, seqid int32) error {
	protoID, err := w.prot.preWriteMessageBegin(name, typeId, seqid)
	if err != nil {
		return err
	}
	// the protocol may be changed during preWriteMessageBegin
	if protoID != ProtocolIDCompact {
		return w.prot.WriteMessageBegin(name, typeId, seqid)
	}

	if err := w.WriteByte(COMPACT_PROTOCOL_ID); err != nil {
		return err
	}
	if err := w.WriteByte((w.version() & COMPACT_VERSION_MASK) | (byte(typeId<<5) & COMPACT_TYPE_MASK)); err != nil {
		return err
	}
	if err := w.writeVarInt32(seqid); err != nil {
		return err
	}
	return w.WriteString(name)
}

func (w *compactWriter) WriteMessageEnd() error {
	return nil
}

func (w *compactWriter) WriteStructBegin(name string) error {
	w.fieldIdStack = append(w.fieldIdStack, w.lastFieldId)
	w.lastFieldId = 0
	return nil
}

func (w *compactWriter) WriteStructEnd() error {
	w.lastFieldId = w.fieldIdStack[len(w.fieldIdStack)-1]
	w.fieldIdStack = w.fieldIdStack[:len(w.fieldIdStack)-1]
	return nil
}

func (w *compactWriter) WriteFieldBegin(name string, typeId Type, id int16) error {
	if typeId == BOOL {
		w.boolFieldId, w.boolFieldPending = id, true
		return nil
	}
	compactType := uint8(compactTypes[typeId])
	// check if we can use delta encoding for the field id
	if id > w.lastFieldId && id-w.lastFieldId <= 15 {
		if err := w.WriteByte(byte((id-w.lastFieldId)<<4) | compactType); err != nil {
			return err
		}
	} else {
		if err := w.WriteByte(compactType); err != nil {
			return err
		}
		if err := w.WriteI16(id); err != nil {
			return err
		}
	}
	w.lastFieldId = id
	return nil
}

func (w *compactWriter) WriteFieldEnd() error {
	return nil
}

func (w *compactWriter) WriteFieldStop() error {
	return w.WriteByte(STOP)
}

func (w *compactWriter) WriteMapBegin(keyType Type, valueType Type, size int) error {
	if size == 0 {
		return w.WriteByte(0)
	}
	if err := w.writeVarInt32(int32(size)); err != nil {
		return err
	}
	return w.WriteByte(byte(compactTypes[keyType]<<4) | byte(compactTypes[valueType]))
}

func (w *compactWriter) WriteMapEnd() error {
	return nil
}

func (w *compactWriter) WriteListBegin(elemType Type, size int) error {
	return w.writeCollectionBegin(elemType, size)
}

func (w *compactWriter) WriteListEnd() error {
	return nil
}

func (w *compactWriter) WriteSetBegin(elemType Type, size int) error {
	return w.writeCollectionBegin(elemType, size)
}

func (w *compactWriter) WriteSetEnd() error {
	return nil
}

func (w *compactWriter) writeCollectionBegin(elemType Type, size int) error {
	if size <= 14 {
		return w.WriteByte(byte(int32(size<<4) | int32(compactTypes[elemType])))
	}
	if err := w.WriteByte(0xf0 | uint8(compactTypes[elemType])); err != nil {
		return err
	}
	return w.writeVarInt32(int32(size))
}

func (w *compactWriter) WriteBool(value bool) error {
	var compactType byte = COMPACT_BOOLEAN_FALSE
	if value {
		compactType = COMPACT_BOOLEAN_TRUE
	}
	// we are not part of a field, so just write the value
	if !w.boolFieldPending {
		return w.WriteByte(compactType)
	}

	id := w.boolFieldId
	// check if we can use delta encoding for the field id
	if id > w.lastFieldId && id-w.lastFieldId <= 15 {
		if err := w.WriteByte(byte((id-w.lastFieldId)<<4) | compactType); err != nil {
			return err
		}
	} else {
		if err := w.WriteByte(compactType); err != nil {
			return err
		}
		if err := w.WriteI16(id); err != nil {
			return err
		}
	}
	w.lastFieldId = id
	w.boolFieldPending = false
	return nil
}

func (w *compactWriter) WriteI16(value int16) error {
	return w.WriteI32(int32(value))
}

func (w *compactWriter) WriteI32(value int32) error {
	zigzag := (value << 1) ^ (value >> 31)
	return w.writeVarInt32(zigzag)
}

func (w *compactWriter) writeVarInt32(n int32) error {
	b := w.tmp[:0]
	for {
		if (n & ^0x7F) == 0 {
			b = append(b, byte(n))
			break
		}
		b = append(b, byte(n&0x7F)|0x80)
		n = int32(uint32(n) >> 7)
	}
	_, err := w.Write(b)
	return err
}

func (w *compactWriter) WriteI64(value int64) error {
	zigzag := (value << 1) ^ (value >> 63)
	return w.writeVarInt64(zigzag)
}

func (w *compactWriter) writeVarInt64(n int64) error {
	b := w.tmp[:0]
	for {
		if (n & ^0x7F) == 0 {
			b = append(b, byte(n))
			break
		}
		b = append(b, byte(n&0x7F)|0x80)
		n = int64(uint64(n) >> 7)
	}
	_, err := w.Write(b)
	return err
}

func (w *compactWriter) WriteDouble(value float64) error {
	b := w.tmp[:8]
	if w.version() == COMPACT_VERSION {
		binary.LittleEndian.PutUint64(b, math.Float64bits(value))
	} else {
		binary.BigEndian.PutUint64(b, math.Float64bits(value))
	}
	_, err := w.Write(b)
	return err
}

func (w *compactWriter) WriteFloat(value float32) error {
	b := w.tmp[:4]
	binary.BigEndian.PutUint32(b, math.Float32bits(value))
	_, err := w.Write(b)
	return err
}

func (w *compactWriter) WriteString(value string) error {
	if err := w.writeVarInt32(int32(len(value))); err != nil {
		return err
	}
	if len(value) == 0 {
		return nil
	}
	_, err := w.Writer.WriteString(value)
	return err
}

func (w *compactWriter) WriteBinary(value []byte) error {
	if err := w.writeVarInt32(int32(len(value))); err != nil {
		return err
	}
	if len(value) == 0 {
		return nil
	}
	_, err := w.Write(value)
	return err
}

func (w *compactWriter) Flush() error {
	if err := w.Writer.Flush(); err != nil {
		return err
	}
	return w.prot.postFlush()
}
