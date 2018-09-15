package reflection

import (
	"bytes"
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
)

func Marshal(val interface{}) ([]byte, error) {
	var p = thrift.DefaultProtocolPool.Get().(*thrift.Protocol)
	_ = p.UseBinary()
	defer thrift.DefaultProtocolPool.Put(p)

	var buf bytes.Buffer
	p.Reset(&buf)
	if err := marshal(val, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MarshalCompact(val interface{}) ([]byte, error) {
	var p = thrift.DefaultProtocolPool.Get().(*thrift.Protocol)
	_ = p.UseCompact(thrift.COMPACT_VERSION)
	defer thrift.DefaultProtocolPool.Put(p)

	var buf bytes.Buffer
	p.Reset(&buf)
	if err := marshal(val, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshal(val interface{}, w thrift.Writer) error {
	if x, ok := val.(thrift.Writable); ok {
		if err := x.Write(w); err != nil {
			return err
		}
	} else {
		if err := Write(val, w); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func Unmarshal(data []byte, val interface{}) error {
	var p = thrift.DefaultProtocolPool.Get().(*thrift.Protocol)
	_ = p.UseBinary()
	defer thrift.DefaultProtocolPool.Put(p)

	var buf = bytes.NewBuffer(data)
	p.Reset(buf)
	if x, ok := val.(thrift.Readable); ok {
		return x.Read(p)
	}
	return Read(val, p)
}

func UnmarshalCompact(data []byte, val interface{}) error {
	var p = thrift.DefaultProtocolPool.Get().(*thrift.Protocol)
	_ = p.UseCompact(thrift.COMPACT_VERSION)
	defer thrift.DefaultProtocolPool.Put(p)

	var buf = bytes.NewBuffer(data)
	p.Reset(buf)
	if x, ok := val.(thrift.Readable); ok {
		return x.Read(p)
	}
	return Read(val, p)
}
