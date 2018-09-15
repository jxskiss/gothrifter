package reflection

import (
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
	"unsafe"
)

type binaryDecoder struct {
}

func (decoder *binaryDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	b, err := r.ReadBinary()
	if err == nil {
		*(*[]byte)(ptr) = b
	}
	return err
}

type stringDecoder struct {
}

func (decoder *stringDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	s, err := r.ReadString()
	if err == nil {
		*(*string)(ptr) = s
	}
	return err
}

type boolDecoder struct {
}

func (decoder *boolDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	b, err := r.ReadBool()
	if err == nil {
		*(*bool)(ptr) = b
	}
	return err
}

type int8Decoder struct {
}

func (decoder *int8Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadByte()
	if err == nil {
		*(*int8)(ptr) = int8(v)
	}
	return err
}

type uint8Decoder struct {
}

func (decoder *uint8Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadByte()
	if err == nil {
		*(*uint8)(ptr) = uint8(v)
	}
	return err
}

type int16Decoder struct {
}

func (decoder *int16Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI16()
	if err == nil {
		*(*int16)(ptr) = v
	}
	return err
}

type uint16Decoder struct {
}

func (decoder *uint16Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI16()
	if err == nil {
		*(*uint16)(ptr) = uint16(v)
	}
	return err
}

type int32Decoder struct {
}

func (decoder *int32Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI32()
	if err == nil {
		*(*int32)(ptr) = v
	}
	return err
}

type uint32Decoder struct {
}

func (decoder *uint32Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI32()
	if err == nil {
		*(*uint32)(ptr) = uint32(v)
	}
	return err
}

type int64Decoder struct {
}

func (decoder *int64Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI64()
	if err == nil {
		*(*int64)(ptr) = v
	}
	return err
}

type uint64Decoder struct {
}

func (decoder *uint64Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI64()
	if err == nil {
		*(*uint64)(ptr) = uint64(v)
	}
	return err
}

type intDecoder struct {
}

func (decoder *intDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI64()
	if err == nil {
		*(*int)(ptr) = int(v)
	}
	return err
}

type uintDecoder struct {
}

func (decoder *uintDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadI64()
	if err == nil {
		*(*uint)(ptr) = uint(v)
	}
	return err
}

type float64Decoder struct {
}

func (decoder *float64Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadDouble()
	if err == nil {
		*(*float64)(ptr) = v
	}
	return err
}

type float32Decoder struct {
}

func (decoder *float32Decoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	v, err := r.ReadFloat()
	if err == nil {
		*(*float32)(ptr) = v
	}
	return err
}

// Encoders

type binaryEncoder struct {
}

func (encoder *binaryEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteBinary(*(*[]byte)(ptr))
}

func (encoder *binaryEncoder) thriftType() thrift.Type {
	return thrift.STRING
}

type stringEncoder struct {
}

func (encoder *stringEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteString(*(*string)(ptr))
}

func (encoder *stringEncoder) thriftType() thrift.Type {
	return thrift.STRING
}

type boolEncoder struct {
}

func (encoder *boolEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteBool(*(*bool)(ptr))
}

func (encoder *boolEncoder) thriftType() thrift.Type {
	return thrift.BOOL
}

type int8Encoder struct {
}

func (encoder *int8Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteByte(byte(*(*int8)(ptr)))
}

func (encoder *int8Encoder) thriftType() thrift.Type {
	return thrift.BYTE
}

type uint8Encoder struct {
}

func (encoder *uint8Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteByte(byte(*(*uint8)(ptr)))
}

func (encoder *uint8Encoder) thriftType() thrift.Type {
	return thrift.BYTE
}

type int16Encoder struct {
}

func (encoder *int16Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI16(*(*int16)(ptr))
}

func (encoder *int16Encoder) thriftType() thrift.Type {
	return thrift.I16
}

type uint16Encoder struct {
}

func (encoder *uint16Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI16(int16(*(*uint16)(ptr)))
}

func (encoder *uint16Encoder) thriftType() thrift.Type {
	return thrift.I16
}

type int32Encoder struct {
}

func (encoder *int32Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI32(*(*int32)(ptr))
}

func (encoder *int32Encoder) thriftType() thrift.Type {
	return thrift.I32
}

type uint32Encoder struct {
}

func (encoder *uint32Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI32(int32(*(*uint32)(ptr)))
}

func (encoder *uint32Encoder) thriftType() thrift.Type {
	return thrift.I32
}

type int64Encoder struct {
}

func (encoder *int64Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI64(*(*int64)(ptr))
}

func (encoder *int64Encoder) thriftType() thrift.Type {
	return thrift.I64
}

type uint64Encoder struct {
}

func (encoder *uint64Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI64(int64(*(*uint64)(ptr)))
}

func (encoder *uint64Encoder) thriftType() thrift.Type {
	return thrift.I64
}

type intEncoder struct {
}

func (encoder *intEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI64(int64(*(*int)(ptr)))
}

func (encoder *intEncoder) thriftType() thrift.Type {
	return thrift.I64
}

type uintEncoder struct {
}

func (encoder *uintEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteI64(int64(*(*uint)(ptr)))
}

func (encoder *uintEncoder) thriftType() thrift.Type {
	return thrift.I64
}

type float64Encoder struct {
}

func (encoder *float64Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteDouble(*(*float64)(ptr))
}

func (encoder *float64Encoder) thriftType() thrift.Type {
	return thrift.DOUBLE
}

type float32Encoder struct {
}

func (encoder *float32Encoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return w.WriteFloat(*(*float32)(ptr))
}

func (encoder *float32Encoder) thriftType() thrift.Type {
	return thrift.FLOAT
}
