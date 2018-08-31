package reflection

import (
	"bytes"
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
)

func Marshal(val interface{}) ([]byte, error) {
	var p = thrift.DefaultProtocolPool.Get().(*thrift.Protocol)
	defer thrift.DefaultProtocolPool.Put(p)

	var buf bytes.Buffer
	p.Reset(&buf)
	if x, ok := val.(thrift.Writable); ok {
		if err := x.Write(p); err != nil {
			return nil, err
		}
	} else {
		if err := Write(val, p); err != nil {
			return nil, err
		}
	}
	if err := p.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, val interface{}) error {
	var p = thrift.DefaultProtocolPool.Get().(*thrift.Protocol)
	defer thrift.DefaultProtocolPool.Put(p)

	var buf = bytes.NewBuffer(data)
	p.Reset(buf)
	if x, ok := val.(thrift.Readable); ok {
		return x.Read(p)
	}
	return Read(val, p)
}
