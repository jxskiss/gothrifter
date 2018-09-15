package reflection

import (
	thrift "github.com/jxskiss/gothrifter/lib/thrift"
	"reflect"
	"unsafe"
)

var reflectTrueValue = reflect.ValueOf(true)

type mapDecoder struct {
	mapType      reflect.Type
	mapInterface emptyInterface
	keyType      reflect.Type
	keyDecoder   internalDecoder
	elemType     reflect.Type
	elemDecoder  internalDecoder
	tType        thrift.Type
}

func (decoder *mapDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	if decoder.tType == thrift.SET {
		return decoder.decodeSet(ptr, r)
	}
	return decoder.decodeMap(ptr, r)
}

func (decoder *mapDecoder) decodeMap(ptr unsafe.Pointer, r thrift.Reader) error {
	mapInterface := decoder.mapInterface
	mapInterface.word = ptr
	realInterface := (*interface{})(unsafe.Pointer(&mapInterface))
	mapVal := reflect.ValueOf(*realInterface).Elem()
	if mapVal.IsNil() {
		mapVal.Set(reflect.MakeMap(decoder.mapType))
	}
	_, _, length, err := r.ReadMapBegin()
	if err != nil {
		return err
	}
	return decoder.readMap(mapVal, length, r)
}

func (decoder *mapDecoder) decodeSet(ptr unsafe.Pointer, r thrift.Reader) error {
	mapInterface := decoder.mapInterface
	mapInterface.word = ptr
	realInterface := (*interface{})(unsafe.Pointer(&mapInterface))
	mapVal := reflect.ValueOf(*realInterface).Elem()
	if mapVal.IsNil() {
		mapVal.Set(reflect.MakeMap(decoder.mapType))
	}
	_, length, err := r.ReadSetBegin()
	if err != nil {
		return err
	}
	return decoder.readSet(mapVal, length, r)
}

func (decoder *mapDecoder) readMap(mapVal reflect.Value, length int, r thrift.Reader) error {
	for i := 0; i < length; i++ {
		keyVal := reflect.New(decoder.keyType)
		if err := decoder.keyDecoder.decode(unsafe.Pointer(keyVal.Pointer()), r); err != nil {
			return err
		}
		elemVal := reflect.New(decoder.elemType)
		if err := decoder.elemDecoder.decode(unsafe.Pointer(elemVal.Pointer()), r); err != nil {
			return err
		}
		mapVal.SetMapIndex(keyVal.Elem(), elemVal.Elem())
	}
	return nil
}

func (decoder *mapDecoder) readSet(mapVal reflect.Value, length int, r thrift.Reader) error {
	for i := 0; i < length; i++ {
		keyVal := reflect.New(decoder.keyType)
		if err := decoder.keyDecoder.decode(unsafe.Pointer(keyVal.Pointer()), r); err != nil {
			return err
		}
		mapVal.SetMapIndex(keyVal.Elem(), reflectTrueValue)
	}
	return nil
}

type mapEncoder struct {
	mapInterface emptyInterface
	keyEncoder   internalEncoder
	elemEncoder  internalEncoder
	tType        thrift.Type
}

func (encoder *mapEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	if encoder.tType == thrift.SET {
		return encoder.encodeSet(ptr, w)
	}
	return encoder.encodeMap(ptr, w)
}

func (encoder *mapEncoder) encodeMap(ptr unsafe.Pointer, w thrift.Writer) error {
	mapInterface := encoder.mapInterface
	mapInterface.word = ptr
	realInterface := (*interface{})(unsafe.Pointer(&mapInterface))
	mapVal := reflect.ValueOf(*realInterface)
	keys := mapVal.MapKeys()
	keyType := encoder.keyEncoder.thriftType()
	elemType := encoder.elemEncoder.thriftType()
	if err := w.WriteMapBegin(keyType, elemType, len(keys)); err != nil {
		return err
	}
	for _, key := range keys {
		keyObj := key.Interface()
		keyInf := (*emptyInterface)(unsafe.Pointer(&keyObj))
		if err := encoder.keyEncoder.encode(keyInf.word, w); err != nil {
			return err
		}
		elem := mapVal.MapIndex(key)
		elemObj := elem.Interface()
		elemInf := (*emptyInterface)(unsafe.Pointer(&elemObj))
		if err := encoder.elemEncoder.encode(elemInf.word, w); err != nil {
			return err
		}
	}
	return nil
}

func (encoder *mapEncoder) encodeSet(ptr unsafe.Pointer, w thrift.Writer) error {
	mapInterface := encoder.mapInterface
	mapInterface.word = ptr
	realInterface := (*interface{})(unsafe.Pointer(&mapInterface))
	mapVal := reflect.ValueOf(*realInterface)
	keys := mapVal.MapKeys()
	if err := w.WriteSetBegin(encoder.keyEncoder.thriftType(), len(keys)); err != nil {
		return err
	}
	for _, key := range keys {
		keyObj := key.Interface()
		keyInf := (*emptyInterface)(unsafe.Pointer(&keyObj))
		if err := encoder.keyEncoder.encode(keyInf.word, w); err != nil {
			return err
		}
	}
	return nil
}

func (encoder *mapEncoder) thriftType() thrift.Type {
	if encoder.tType == thrift.SET {
		return thrift.SET
	}
	return thrift.MAP
}
