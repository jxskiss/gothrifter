package general

import (
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
)

type Field struct {
	Name  string
	Value interface{}
}

type Struct struct {
	Name   string
	Fields map[int16]Field
}

func (obj Struct) Get(path ...interface{}) interface{} {
	if len(path) == 0 {
		return obj
	}
	fieldId, ok := path[0].(int16)
	if !ok {
		fieldId = int16(path[0].(int))
	}
	elem := obj.Fields[fieldId]
	if len(path) == 1 {
		return elem
	}
	return elem.Value.(Object).Get(path[1:]...)
}

func (obj *Struct) Read(r thrift.Reader) error {
	obj.Fields = make(map[int16]Field)
	name, err := r.ReadStructBegin()
	if err != nil {
		return err
	}
	obj.Name = name
	for {
		fieldName, fieldType, fieldId, err := r.ReadFieldBegin()
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
		obj.Fields[fieldId] = Field{
			Name:  fieldName,
			Value: val,
		}
	}
	return r.ReadStructEnd()
}

func (obj *Struct) Write(w thrift.Writer) error {
	if err := w.WriteStructBegin(obj.Name); err != nil {
		return err
	}
	for fieldId, field := range obj.Fields {
		fieldType, fieldWriter := writerOf(field.Value)
		if err := w.WriteFieldBegin(field.Name, fieldType, fieldId); err != nil {
			return err
		}
		if err := fieldWriter(w, field.Value); err != nil {
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
