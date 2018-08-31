package general

import (
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
)

type List []interface{}

func (obj List) Get(path ...interface{}) interface{} {
	if len(path) == 0 {
		return obj
	}
	elem := obj[path[0].(int)]
	if len(path) == 1 {
		return elem
	}
	return elem.(Object).Get(path[1:]...)
}

func (obj *List) Read(r thrift.Reader) error {
	elemType, size, err := r.ReadListBegin()
	if err != nil {
		return err
	}
	elemReader := readerOf(elemType)
	for i := 0; i < size; i++ {
		elem, err := elemReader(r)
		if err != nil {
			return err
		}
		*obj = append(*obj, elem)
	}
	return r.ReadListEnd()
}

func (obj List) Write(w thrift.Writer) error {
	var length = len(obj)
	var elemType thrift.Type
	var elemWriter func(val interface{}, w thrift.Writer) error
	if length == 0 {
		elemType = thrift.I64
	} else {
		elemType, elemWriter = writerOf(obj[0])
	}
	if err := w.WriteListBegin(elemType, length); err != nil {
		return err
	}
	for _, elem := range obj {
		if err := elemWriter(elem, w); err != nil {
			return err
		}
	}
	return w.WriteListEnd()
}
