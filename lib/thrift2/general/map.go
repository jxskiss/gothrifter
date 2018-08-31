package general

import (
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
)

type Map map[interface{}]interface{}

func (obj Map) Get(path ...interface{}) interface{} {
	if len(path) == 0 {
		return obj
	}
	elem := obj[path[0]]
	if len(path) == 1 {
		return elem
	}
	return elem.(Object).Get(path[1:]...)
}

func (obj *Map) Read(r thrift.Reader) error {
	keyType, valueType, size, err := r.ReadMapBegin()
	if err != nil {
		return err
	}
	if size == 0 {
		return nil
	}
	keyReader := readerOf(keyType)
	valueReader := readerOf(valueType)
	for i := 0; i < size; i++ {
		var key, value interface{}
		if key, err = keyReader(r); err != nil {
			return err
		}
		if value, err = valueReader(r); err != nil {
			return err
		}
		(*obj)[key] = value
	}
	return r.ReadMapEnd()
}

func (obj Map) sample() (interface{}, interface{}) {
	for k, v := range obj {
		return k, v
	}
	// should not come here
	panic("can't take sample from empty map")
}

func (obj Map) Write(w thrift.Writer) error {
	var length = len(obj)
	var keyType, valueType thrift.Type
	var keyWriter, valueWriter func(val interface{}, w thrift.Writer) error
	if length == 0 {
		keyType, valueType = thrift.I64, thrift.I64
	} else {
		keySample, valueSample := obj.sample()
		keyType, keyWriter = writerOf(keySample)
		valueType, valueWriter = writerOf(valueSample)
	}
	if err := w.WriteMapBegin(keyType, valueType, length); err != nil {
		return err
	}
	for k, v := range obj {
		if err := keyWriter(k, w); err != nil {
			return err
		}
		if err := valueWriter(v, w); err != nil {
			return err
		}
	}
	return w.WriteMapEnd()
}
