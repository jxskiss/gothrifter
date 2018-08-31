package thrift2

import (
	"errors"
)

var (
	ErrReflectionNotImported = errors.New("thrift: reflection not imported")

	ReadReflect  func(val interface{}, r Reader) error
	WriteReflect func(val interface{}, w Writer) error
)

func Read(val interface{}, r Reader) error {
	if x, ok := val.(Readable); ok {
		return x.Read(r)
	}
	if ReadReflect != nil {
		return ReadReflect(val, r)
	}
	return ErrReflectionNotImported
}

func Write(val interface{}, w Writer) error {
	if x, ok := val.(Writable); ok {
		return x.Write(w)
	}
	if WriteReflect != nil {
		return WriteReflect(val, w)
	}
	return ErrReflectionNotImported
}
