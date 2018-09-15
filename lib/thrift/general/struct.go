package general

import (
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
)

type Struct map[int16]interface{}

func (obj Struct) Get(path ...interface{}) interface{} {
	if len(path) == 0 {
		return obj
	}
	fieldId, ok := path[0].(int16)
	if !ok {
		fieldId = int16(path[0].(int))
	}
	elem := obj[fieldId]
	if len(path) == 1 {
		return elem
	}
	return elem.(Object).Get(path[1:]...)
}

func (obj *Struct) Read(r thrift.Reader) error {
	_, err := r.ReadStructBegin()
	if err != nil {
		return err
	}
	for {
		_, fieldType, fieldId, err := r.ReadFieldBegin()
		if err != nil {
			return err
		}
		if fieldType == thrift.STOP {
			break
		}
		fieldReader := readerOf(fieldType)
		val, err := fieldReader(r)
		if err != nil {
			return err
		}
		if err = r.ReadFieldEnd(); err != nil {
			return err
		}
		(*obj)[fieldId] = val
	}
	return r.ReadStructEnd()
}

func (obj Struct) Write(w thrift.Writer) error {
	if err := w.WriteStructBegin(""); err != nil {
		return err
	}
	for fieldId, field := range obj {
		fieldType, fieldWriter := writerOf(field)
		if err := w.WriteFieldBegin("", fieldType, fieldId); err != nil {
			return err
		}
		if err := fieldWriter(field, w); err != nil {
			return err
		}
		if err := w.WriteFieldEnd(); err != nil {
			return err
		}
	}
	if err := w.WriteFieldStop(); err != nil {
		return err
	}
	return w.WriteStructEnd()
}
