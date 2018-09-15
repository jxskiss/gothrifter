package reflection

import (
	"unsafe"
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
)

type internalDecoder interface {
	decode(ptr unsafe.Pointer, r thrift.Reader) error
}

type valDecoderAdapter struct {
	decoder internalDecoder
}

func (d *valDecoderAdapter) Decode(val interface{}, r thrift.Reader) error {
	ptr := (*emptyInterface)(unsafe.Pointer(&val)).word
	return d.decoder.decode(ptr, r)
}

type internalDecoderAdapter struct {
	decoder           Decoder
	valEmptyInterface emptyInterface
}

func (d *internalDecoderAdapter) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	valEmptyInterface := d.valEmptyInterface
	valEmptyInterface.word = ptr
	valObj := *(*interface{})((unsafe.Pointer(&valEmptyInterface)))
	return d.decoder.Decode(valObj, r)
}

type internalEncoder interface {
	encode(ptr unsafe.Pointer, w thrift.Writer) error
	thriftType() thrift.Type
}

type valEncoderAdapter struct {
	encoder internalEncoder
}

func (e *valEncoderAdapter) Encode(val interface{}, w thrift.Writer) error {
	ptr := (*emptyInterface)(unsafe.Pointer(&val)).word
	return e.encoder.encode(ptr, w)
}

func (e *valEncoderAdapter) ThriftType() thrift.Type {
	return e.encoder.thriftType()
}

type ptrEncoderAdapter struct {
	encoder internalEncoder
}

func (e *ptrEncoderAdapter) Encode(val interface{}, w thrift.Writer) error {
	ptr := (*emptyInterface)(unsafe.Pointer(&val)).word
	return e.encoder.encode(unsafe.Pointer(&ptr), w)
}

func (e *ptrEncoderAdapter) ThriftType() thrift.Type {
	return e.encoder.thriftType()
}

type internalEncoderAdapter struct {
	encoder           Encoder
	valEmptyInterface emptyInterface
}

func (e *internalEncoderAdapter) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	valEmptyInterface := e.valEmptyInterface
	valEmptyInterface.word = ptr
	valObj := *(*interface{})((unsafe.Pointer(&valEmptyInterface)))
	return e.encoder.Encode(valObj, w)
}

func (e *internalEncoderAdapter) thriftType() thrift.Type {
	return e.encoder.ThriftType()
}

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  unsafe.Pointer
	word unsafe.Pointer
}

// sliceHeader is a safe version of SliceHeader used within this package.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
