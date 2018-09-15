package reflection

import (
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
	"unsafe"
)

type structDecoder struct {
	fields   []structDecoderField
	fieldMap map[int16]structDecoderField
}

type structDecoderField struct {
	offset  uintptr
	fieldId int16
	decoder internalDecoder
}

func (decoder *structDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	if _, err := r.ReadStructBegin(); err != nil {
		return err
	}
	for _, field := range decoder.fields {
		_, fieldType, fieldId, err := r.ReadFieldBegin()
		if err != nil {
			return err
		}
		if field.fieldId == fieldId {
			if err := field.decoder.decode(unsafe.Pointer(uintptr(ptr)+field.offset), r); err != nil {
				return err
			}
		} else {
			if err := decoder.decodeByMap(ptr, r, fieldType, fieldId); err != nil {
				return err
			}
			return r.ReadStructEnd()
		}
	}
	_, fieldType, fieldId, err := r.ReadFieldBegin()
	if err != nil {
		return err
	}
	if err := decoder.decodeByMap(ptr, r, fieldType, fieldId); err != nil {
		return err
	}
	return r.ReadStructEnd()
}

func (decoder *structDecoder) decodeByMap(ptr unsafe.Pointer, r thrift.Reader,
	fieldType thrift.Type, fieldId int16) (err error) {
	for {
		if fieldType == thrift.STOP {
			return nil
		}
		field, isFound := decoder.fieldMap[fieldId]
		if isFound {
			if err = field.decoder.decode(unsafe.Pointer(uintptr(ptr)+field.offset), r); err != nil {
				return err
			}
		} else {
			if err := r.Skip(fieldType); err != nil {
				return err
			}
		}
		if _, fieldType, fieldId, err = r.ReadFieldBegin(); err != nil {
			return err
		}
	}
}

type structEncoder struct {
	fields []structEncoderField
}

type structEncoderField struct {
	offset  uintptr
	fieldId int16
	encoder internalEncoder
}

func (encoder *structEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	if err := w.WriteStructBegin(""); err != nil {
		return err
	}
	for _, field := range encoder.fields {
		fieldPtr := unsafe.Pointer(uintptr(ptr) + field.offset)
		switch field.encoder.(type) {
		case *pointerEncoder, *sliceEncoder:
			if *(*unsafe.Pointer)(fieldPtr) == nil {
				continue
			}
		case *mapEncoder:
			if *(*unsafe.Pointer)(fieldPtr) == nil {
				continue
			}
			fieldPtr = *(*unsafe.Pointer)(fieldPtr)
		}
		if err := w.WriteFieldBegin("", field.encoder.thriftType(), field.fieldId); err != nil {
			return err
		}
		if err := field.encoder.encode(fieldPtr, w); err != nil {
			return err
		}
	}
	if err := w.WriteFieldStop(); err != nil {
		return err
	}
	return w.WriteStructEnd()
}

func (encoder *structEncoder) thriftType() thrift.Type {
	return thrift.STRUCT
}
