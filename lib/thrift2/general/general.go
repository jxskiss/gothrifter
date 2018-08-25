package general

import (
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
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
	if err := w.WriteMessageEnd(); err != nil {
		return err
	}
	return w.Flush()
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
			var obj = new(List)
			if err := obj.Read(r); err != nil {
				return nil, err
			}
			return obj, nil
		}
	case thrift.MAP:
		return func(r thrift.Reader) (interface{}, error) {
			var obj = new(Map)
			if err := obj.Read(r); err != nil {
				return nil, err
			}
			return obj, nil
		}
	case thrift.SET:
		return func(r thrift.Reader) (interface{}, error) {
			var obj = new(Set)
			if err := obj.Read(r); err != nil {
				return nil, err
			}
			return obj, nil
		}
	default:
		panic("unsupported type: " + ttype.String())
	}
}

func writerOf(sample interface{}) (thrift.Type, func(w thrift.Writer, val interface{}) error) {
	switch sample.(type) {
	case bool:
		return thrift.BOOL, func(w thrift.Writer, val interface{}) error { return w.WriteBool(val.(bool)) }
	case int8:
		return thrift.I08, func(w thrift.Writer, val interface{}) error { return w.WriteByte(byte(val.(int8))) }
	case uint8:
		return thrift.I08, func(w thrift.Writer, val interface{}) error { return w.WriteByte(byte(val.(uint8))) }
	case int16:
		return thrift.I16, func(w thrift.Writer, val interface{}) error { return w.WriteI16(val.(int16)) }
	case uint16:
		return thrift.I16, func(w thrift.Writer, val interface{}) error { return w.WriteI16(int16(val.(uint16))) }
	case int32:
		return thrift.I32, func(w thrift.Writer, val interface{}) error { return w.WriteI32(val.(int32)) }
	case uint32:
		return thrift.I32, func(w thrift.Writer, val interface{}) error { return w.WriteI32(int32(val.(uint32))) }
	case int64:
		return thrift.I64, func(w thrift.Writer, val interface{}) error { return w.WriteI64(val.(int64)) }
	case uint64:
		return thrift.I64, func(w thrift.Writer, val interface{}) error { return w.WriteI64(int64(val.(uint64))) }
	case string:
		return thrift.STRING, func(w thrift.Writer, val interface{}) error { return w.WriteString(val.(string)) }
	case []byte:
		return thrift.BINARY, func(w thrift.Writer, val interface{}) error { return w.WriteBinary(val.([]byte)) }
	case float64:
		return thrift.DOUBLE, func(w thrift.Writer, val interface{}) error { return w.WriteDouble(val.(float64)) }
	case float32:
		return thrift.FLOAT, func(w thrift.Writer, val interface{}) error { return w.WriteFloat(val.(float32)) }
	case List:
		return thrift.LIST, func(w thrift.Writer, val interface{}) error { return val.(List).Write(w) }
	case Map:
		return thrift.MAP, func(w thrift.Writer, val interface{}) error { return val.(Map).Write(w) }
	case Set:
		return thrift.SET, func(w thrift.Writer, val interface{}) error { return val.(Set).Write(w) }
	case Struct:
		return thrift.STRUCT, func(w thrift.Writer, val interface{}) error { return val.(Struct).Write(w) }
	default:
		panic("unsupported type: " + reflect.TypeOf(sample).String())
	}
}
