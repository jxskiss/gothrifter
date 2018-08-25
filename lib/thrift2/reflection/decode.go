package reflection

import (
	"bytes"
	"fmt"
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unsafe"
)

var decodersCache sync.Map

var byteSliceType = reflect.TypeOf(([]byte)(nil))

func Unmarshal(data []byte, val interface{}) error {
	var buf bytes.Buffer
	var p = thrift.NewProtocol(&buf, thrift.DefaultOptions)
	if _, err := buf.Write(data); err != nil {
		return err
	}
	return Read(p, val)
}

func Read(r thrift.Reader, val interface{}) error {
	if x, ok := val.(thrift.Readable); ok {
		return x.Read(r)
	}
	decoder := DecoderOf(reflect.TypeOf(val))
	return decoder.Decode(val, r)
}

type Decoder interface {
	Decode(val interface{}, r thrift.Reader) error
}

func DecoderOf(valType reflect.Type) Decoder {
	if decoder, ok := decodersCache.Load(valType); ok {
		return decoder.(Decoder)
	}
	// make new decoder and cache it
	var decoder Decoder
	if valType.Kind() != reflect.Ptr {
		decoder = &valDecoderAdapter{&unknownDecoder{
			prefix: "non-pointer type", valType: valType,
		}}
	} else {
		decoder = &valDecoderAdapter{decoderOf("", valType.Elem())}
	}
	decodersCache.Store(valType, decoder)
	return decoder
}

func decoderOf(prefix string, valType reflect.Type) internalDecoder {
	if byteSliceType == valType {
		return &binaryDecoder{}
	}
	if isEnumType(valType) {
		return &int32Decoder{}
	}
	switch valType.Kind() {
	case reflect.Bool:
		return &boolDecoder{}
	case reflect.Float64:
		return &float64Decoder{}
	case reflect.Int:
		return &intDecoder{}
	case reflect.Uint:
		return &uintDecoder{}
	case reflect.Int8:
		return &int8Decoder{}
	case reflect.Uint8:
		return &uint8Decoder{}
	case reflect.Int16:
		return &int16Decoder{}
	case reflect.Uint16:
		return &uint16Decoder{}
	case reflect.Int32:
		return &int32Decoder{}
	case reflect.Uint32:
		return &uint32Decoder{}
	case reflect.Int64:
		return &int64Decoder{}
	case reflect.Uint64:
		return &uint64Decoder{}
	case reflect.String:
		return &stringDecoder{}
	case reflect.Ptr:
		return &pointerDecoder{
			valType:    valType.Elem(),
			valDecoder: decoderOf(prefix+" [ptrElem]", valType.Elem()),
		}
	case reflect.Slice:
		return &sliceDecoder{
			elemType:    valType.Elem(),
			sliceType:   valType,
			elemDecoder: decoderOf(prefix+" [sliceElem]", valType.Elem()),
		}
	case reflect.Map:
		sampleObj := reflect.New(valType).Interface()
		decoder := &mapDecoder{
			keyType:      valType.Key(),
			keyDecoder:   decoderOf(prefix+" [mapKey]", valType.Key()),
			elemType:     valType.Elem(),
			elemDecoder:  decoderOf(prefix+" [mapElem]", valType.Elem()),
			mapType:      valType,
			mapInterface: *(*emptyInterface)(unsafe.Pointer(&sampleObj)),
			tType:        thrift.MAP,
		}
		// FIXME: is there any reasonable way to auto distinct map and set?
		if valType.Elem().Kind() == reflect.Bool {
			decoder.tType = thrift.SET
		}
		return decoder
	case reflect.Struct:
		decoderFields := make([]structDecoderField, 0, valType.NumField())
		decoderFieldMap := map[int16]structDecoderField{}
		for i := 0; i < valType.NumField(); i++ {
			refField := valType.Field(i)
			fieldId := parseFieldId(refField)
			if fieldId == -1 {
				continue
			}
			decoderField := structDecoderField{
				offset:  refField.Offset,
				fieldId: fieldId,
				decoder: decoderOf(prefix+" "+refField.Name, refField.Type),
			}
			if refField.Type.Kind() == reflect.Map && refField.Type.Elem().Kind() == reflect.Bool {
				decoderField.decoder.(*mapDecoder).tType = parseSetType(refField)
			}
			decoderFields = append(decoderFields, decoderField)
			decoderFieldMap[fieldId] = decoderField
		}
		return &structDecoder{
			fields:   decoderFields,
			fieldMap: decoderFieldMap,
		}
	}
	return &unknownDecoder{prefix, valType}
}

func isEnumType(valType reflect.Type) bool {
	if valType.Kind() != reflect.Int64 {
		return false
	}
	_, hasStringMethod := valType.MethodByName("String")
	return hasStringMethod
}

func parseFieldId(refField reflect.StructField) int16 {
	if !unicode.IsUpper(rune(refField.Name[0])) {
		return -1
	}
	thriftTag := refField.Tag.Get("thrift")
	if thriftTag == "" {
		return -1
	}
	parts := strings.Split(thriftTag, ",")
	if len(parts) < 2 {
		return -1
	}
	fieldId, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}
	return int16(fieldId)
}

func parseSetType(refField reflect.StructField) thrift.Type {
	tag := refField.Tag.Get("thrift")
	parts := strings.Split(tag, ",")
	if len(parts) >= 4 && strings.TrimSpace(parts[3]) == "map" {
		return thrift.MAP
	}
	return thrift.SET
}

type unknownDecoder struct {
	prefix  string
	valType reflect.Type
}

func (decoder *unknownDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	return fmt.Errorf("%v: do not known how to decode %v", decoder.prefix, decoder.valType.String())
}
