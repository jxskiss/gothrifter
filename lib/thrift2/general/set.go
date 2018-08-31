package general

import (
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
)

type Set map[interface{}]bool

func (obj Set) Get(path ...interface{}) interface{} {
	if len(path) == 0 {
		return obj
	}
	return obj[path[0]]
}

func (obj *Set) Read(r thrift.Reader) error {
	elemType, size, err := r.ReadSetBegin()
	if err != nil {
		return err
	}
	if size == 0 {
		return nil
	}
	elemReader := readerOf(elemType)
	for i := 0; i < size; i++ {
		elem, err := elemReader(r)
		if err != nil {
			return err
		}
		(*obj)[elem] = true
	}
	return r.ReadSetEnd()
}

func (obj Set) sample() interface{} {
	for elem := range obj {
		return elem
	}
	// shall not go to here
	panic("can't take sample from empty set")
}

func (obj Set) Write(w thrift.Writer) error {
	var length = len(obj)
	var elemType thrift.Type
	var elemWriter func(val interface{}, w thrift.Writer) error
	if length == 0 {
		elemType = thrift.I64
	} else {
		elemSample := obj.sample()
		elemType, elemWriter = writerOf(elemSample)
	}
	if err := w.WriteSetBegin(elemType, length); err != nil {
		return err
	}
	for elem := range obj {
		if err := elemWriter(elem, w); err != nil {
			return err
		}
	}
	return w.WriteSetEnd()
}
