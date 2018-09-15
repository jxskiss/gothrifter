package reflection

import (
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
	"reflect"
	"unsafe"
)

type pointerDecoder struct {
	valType    reflect.Type
	valDecoder internalDecoder
}

func (decoder *pointerDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	value := reflect.New(decoder.valType).Interface()
	newPtr := (*emptyInterface)(unsafe.Pointer(&value)).word
	if err := decoder.valDecoder.decode(newPtr, r); err != nil {
		return err
	}
	*(*unsafe.Pointer)(ptr) = newPtr
	return nil
}

type pointerEncoder struct {
	valType    reflect.Type
	valEncoder internalEncoder
}

func (encoder *pointerEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	valPtr := *(*unsafe.Pointer)(ptr)
	if encoder.valType.Kind() == reflect.Map {
		valPtr = *(*unsafe.Pointer)(valPtr)
	}
	return encoder.valEncoder.encode(valPtr, w)
}

func (encoder *pointerEncoder) thriftType() thrift.Type {
	return encoder.valEncoder.thriftType()
}
