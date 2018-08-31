package reflection

import (
	"fmt"
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
	"reflect"
	"sync"
	"unsafe"
)

func init() {
	thrift.WriteReflect = Write
}

var encodersCache sync.Map

func Write(val interface{}, w thrift.Writer) error {
	encoder := EncoderOf(reflect.TypeOf(val))
	return encoder.Encode(val, w)
}

type Encoder interface {
	Encode(val interface{}, w thrift.Writer) error
	ThriftType() thrift.Type
}

func EncoderOf(valType reflect.Type) Encoder {
	if encoder, ok := encodersCache.Load(valType); ok {
		return encoder.(Encoder)
	}
	// make new encoder and cache it
	var encoder Encoder
	isPtr := valType.Kind() == reflect.Ptr
	isOnePtrArray := valType.Kind() == reflect.Array && valType.Len() == 1 &&
		valType.Elem().Kind() == reflect.Ptr
	isOnePtrStruct := valType.Kind() == reflect.Struct && valType.NumField() == 1 &&
		valType.Field(0).Type.Kind() == reflect.Ptr
	isOneMapStruct := valType.Kind() == reflect.Struct && valType.NumField() == 1 &&
		valType.Field(0).Type.Kind() == reflect.Map
	if isPtr || isOnePtrArray || isOnePtrStruct || isOneMapStruct {
		encoder = &ptrEncoderAdapter{encoderOf("", valType)}
	} else {
		encoder = &valEncoderAdapter{encoderOf("", valType)}
	}
	encodersCache.Store(valType, encoder)
	return encoder
}

func encoderOf(prefix string, valType reflect.Type) internalEncoder {
	if byteSliceType == valType {
		return &binaryEncoder{}
	}
	if isEnumType(valType) {
		return &int32Encoder{}
	}
	switch valType.Kind() {
	case reflect.String:
		return &stringEncoder{}
	case reflect.Bool:
		return &boolEncoder{}
	case reflect.Int8:
		return &int8Encoder{}
	case reflect.Uint8:
		return &uint8Encoder{}
	case reflect.Int16:
		return &int16Encoder{}
	case reflect.Uint16:
		return &uint16Encoder{}
	case reflect.Int32:
		return &int32Encoder{}
	case reflect.Uint32:
		return &uint32Encoder{}
	case reflect.Int64:
		return &int64Encoder{}
	case reflect.Uint64:
		return &uint64Encoder{}
	case reflect.Int:
		return &intEncoder{}
	case reflect.Uint:
		return &uintEncoder{}
	case reflect.Float32:
		return &float32Encoder{}
	case reflect.Float64:
		return &float64Encoder{}
	case reflect.Slice:
		return &sliceEncoder{
			sliceType:   valType,
			elemType:    valType.Elem(),
			elemEncoder: encoderOf(prefix+" [sliceElem]", valType.Elem()),
		}
	case reflect.Map:
		sampleObj := reflect.New(valType).Elem().Interface()
		encoder := &mapEncoder{
			keyEncoder:   encoderOf(prefix+" [mapKey]", valType.Key()),
			elemEncoder:  encoderOf(prefix+" [mapElem]", valType.Elem()),
			mapInterface: *(*emptyInterface)(unsafe.Pointer(&sampleObj)),
			tType:        thrift.MAP,
		}
		// FIXME: is there any reasonable way to auto distinct map and set?
		if valType.Elem().Kind() == reflect.Bool {
			encoder.tType = thrift.SET
		}
		return encoder
	case reflect.Struct:
		encoderFields := make([]structEncoderField, 0, valType.NumField())
		for i := 0; i < valType.NumField(); i++ {
			refField := valType.Field(i)
			fieldId := parseFieldId(refField)
			if fieldId == -1 {
				continue
			}
			encoderField := structEncoderField{
				offset:  refField.Offset,
				fieldId: fieldId,
				encoder: encoderOf(prefix+" "+refField.Name, refField.Type),
			}
			if refField.Type.Kind() == reflect.Map && refField.Type.Elem().Kind() == reflect.Bool {
				encoderField.encoder.(*mapEncoder).tType = parseSetType(refField)
			}
			encoderFields = append(encoderFields, encoderField)
		}
		return &structEncoder{
			fields: encoderFields,
		}
	case reflect.Ptr:
		return &pointerEncoder{
			valType:    valType.Elem(),
			valEncoder: encoderOf(prefix+" [ptrElem]", valType.Elem()),
		}
	}
	return &unknownEncoder{prefix, valType}
}

type unknownEncoder struct {
	prefix  string
	valType reflect.Type
}

func (encoder *unknownEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	return fmt.Errorf("%v: do not know how to encode %v", encoder.prefix, encoder.valType.String())
}

func (encoder *unknownEncoder) thriftType() thrift.Type {
	return thrift.STOP
}
