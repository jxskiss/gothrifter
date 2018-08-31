package general

import (
	"bytes"
	"encoding/hex"
	thrift "github.com/jxskiss/gothrifter/lib/thrift2"
	"github.com/matryer/is"
	"io"
	"testing"
)

func Test_General_Objects(t *testing.T) {
	is := is.New(t)
	obj := Struct{
		0: true,
		1:  int16(1),
		2:  int32(2),
		3:  int64(3),
		4:  "4",
		5:  []byte("5"),
		6:  Set{int64(61): true},
		8:  Map{int32(81): "81"},
		10: List{int64(101), int64(102), int64(103)},
		12: Struct{
			1201: "1201",
			1202: int64(1202),
		},
	}

	var buf bytes.Buffer
	var p = thrift.NewProtocol(&buf, thrift.DefaultOptions)
	var err error
	var r = readerOf(thrift.STRUCT)

	{
		buf.Reset()
		ttype, w := writerOf(obj)
		is.True(ttype == thrift.STRUCT)
		err = w(obj, p)
		is.NoErr(err)
		err = p.Flush()
		is.NoErr(err)
		b1 := buf.Bytes()

		buf.Reset()
		buf.Write(b1)
		val1, err := r(p)
		is.True(err == nil || err == io.EOF)
		is.Equal(val1, obj)
	}
	{
		buf.Reset()
		ttype, w := writerOf(&obj) // pointer of struct
		is.True(ttype == thrift.STRUCT)
		err = w(&obj, p)
		is.NoErr(err)
		err = p.Flush()
		is.NoErr(err)
		b2 := buf.Bytes()

		buf.Reset()
		buf.Write(b2)
		val2, err := r(p)
		is.True(err == nil || err == io.EOF)
		is.Equal(val2, obj)
	}
}

func Test_Struct_Field(t *testing.T) {
	is := is.New(t)

	obj := Struct{
		0: true,
		12: Struct{
			1201: "1201",
			1202: int64(1202),
		},
	}
	s := "0c000c0b04b100000004313230310a04b200000000000004b2000200000100"
	b, err := hex.DecodeString(s)
	is.NoErr(err)

	var buf = bytes.NewBuffer(b)
	var p = thrift.NewProtocol(buf, thrift.DefaultOptions)
	r := readerOf(thrift.STRUCT)
	val, err := r(p)
	is.True(err == nil || err == io.EOF)
	is.Equal(val, obj)
}
