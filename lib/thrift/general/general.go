package general

import (
	thrift "github.com/jxskiss/gothrifter/lib/thrift"
	"reflect"
)

type Object interface {
	Get(path ...interface{}) interface{}
}

// General header and message representation.

type Message struct {
	Header    MessageHeader
	Arguments Struct
}

func (msg *Message) Read(r thrift.Reader) error {
	if err := msg.Header.Read(r); err != nil {
		return err
	}
	return msg.Arguments.Read(r)
}

func (msg *Message) Write(w thrift.Writer) error {
	if err := msg.Header.Write(w); err != nil {
		return err
	}
	if err := msg.Arguments.Write(w); err != nil {
		return err
	}
	return w.WriteMessageEnd()
}

type MessageHeader struct {
	MessageName string
	MessageType thrift.MessageType
	SeqId       int32
}

func (h *MessageHeader) Read(r thrift.Reader) error {
	name, typeId, seqid, err := r.ReadMessageBegin()
	if err != nil {
		return err
	}
	h.MessageName = name
	h.MessageType = typeId
	h.SeqId = seqid
	return nil
}

func (h *MessageHeader) Write(w thrift.Writer) error {
	return w.WriteMessageBegin(h.MessageName, h.MessageType, h.SeqId)
}

func readerOf(ttype thrift.Type) func(r thrift.Reader) (interface{}, error) {
	switch ttype {
	case thrift.BOOL:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadBool() }
	case thrift.I08:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadByte() }
	case thrift.I16:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadI16() }
	case thrift.I32:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadI32() }
	case thrift.I64:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadI64() }
	case thrift.STRING:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadString() }
	case thrift.BINARY:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadBinary() }
	case thrift.DOUBLE:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadDouble() }
	case thrift.FLOAT:
		return func(r thrift.Reader) (interface{}, error) { return r.ReadFloat() }
	case thrift.LIST:
		return func(r thrift.Reader) (interface{}, error) {
			obj := List(make([]interface{}, 0))
			err := obj.Read(r)
			return obj, err
		}
	case thrift.MAP:
		return func(r thrift.Reader) (interface{}, error) {
			obj := Map(make(map[interface{}]interface{}))
			err := obj.Read(r)
			return obj, err
		}
	case thrift.SET:
		return func(r thrift.Reader) (interface{}, error) {
			obj := Set(make(map[interface{}]bool))
			err := obj.Read(r)
			return obj, err
		}
	case thrift.STRUCT:
		return func(r thrift.Reader) (interface{}, error) {
			obj := Struct{}
			err := obj.Read(r)
			return obj, err
		}
	default:
		panic("unsupported type: " + ttype.String())
	}
}

func writerOf(sample interface{}) (thrift.Type, func(val interface{}, w thrift.Writer) error) {
	switch sample.(type) {
	case bool:
		return thrift.BOOL, func(val interface{}, w thrift.Writer) error { return w.WriteBool(val.(bool)) }
	case int8:
		return thrift.I08, func(val interface{}, w thrift.Writer) error { return w.WriteByte(byte(val.(int8))) }
	case uint8:
		return thrift.I08, func(val interface{}, w thrift.Writer) error { return w.WriteByte(byte(val.(uint8))) }
	case int16:
		return thrift.I16, func(val interface{}, w thrift.Writer) error { return w.WriteI16(val.(int16)) }
	case uint16:
		return thrift.I16, func(val interface{}, w thrift.Writer) error { return w.WriteI16(int16(val.(uint16))) }
	case int32:
		return thrift.I32, func(val interface{}, w thrift.Writer) error { return w.WriteI32(val.(int32)) }
	case uint32:
		return thrift.I32, func(val interface{}, w thrift.Writer) error { return w.WriteI32(int32(val.(uint32))) }
	case int64:
		return thrift.I64, func(val interface{}, w thrift.Writer) error { return w.WriteI64(val.(int64)) }
	case uint64:
		return thrift.I64, func(val interface{}, w thrift.Writer) error { return w.WriteI64(int64(val.(uint64))) }
	case int:
		return thrift.I64, func(val interface{}, w thrift.Writer) error { return w.WriteI64(int64(val.(int))) }
	case uint:
		return thrift.I64, func(val interface{}, w thrift.Writer) error { return w.WriteI64(int64(val.(uint))) }
	case string:
		return thrift.STRING, func(val interface{}, w thrift.Writer) error { return w.WriteString(val.(string)) }
	case []byte:
		return thrift.BINARY, func(val interface{}, w thrift.Writer) error { return w.WriteBinary(val.([]byte)) }
	case float64:
		return thrift.DOUBLE, func(val interface{}, w thrift.Writer) error { return w.WriteDouble(val.(float64)) }
	case float32:
		return thrift.FLOAT, func(val interface{}, w thrift.Writer) error { return w.WriteFloat(val.(float32)) }
	case List:
		return thrift.LIST, func(val interface{}, w thrift.Writer) error { return val.(List).Write(w) }
	case *List:
		return thrift.LIST, func(val interface{}, w thrift.Writer) error { return val.(*List).Write(w) }
	case Map:
		return thrift.MAP, func(val interface{}, w thrift.Writer) error { return val.(Map).Write(w) }
	case *Map:
		return thrift.MAP, func(val interface{}, w thrift.Writer) error { return val.(*Map).Write(w) }
	case Set:
		return thrift.SET, func(val interface{}, w thrift.Writer) error { return val.(Set).Write(w) }
	case *Set:
		return thrift.SET, func(val interface{}, w thrift.Writer) error { return val.(*Set).Write(w) }
	case Struct:
		return thrift.STRUCT, func(val interface{}, w thrift.Writer) error { return val.(Struct).Write(w) }
	case *Struct:
		return thrift.STRUCT, func(val interface{}, w thrift.Writer) error { return val.(*Struct).Write(w) }
	default:
		panic("unsupported type: " + reflect.TypeOf(sample).String())
	}
}
