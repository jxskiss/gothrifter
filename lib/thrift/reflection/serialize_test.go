package reflection

import (
	"github.com/matryer/is"
	"io"
	"testing"
)

type TestObject struct {
	A string         `thrift:"a,1"`
	B int32          `thrift:"b,2"`
	C []int64        `thrift:"c,3"`
	D map[int]string `thrift:"d,4"`
	E map[int]bool   `thrift:"e,5"`
}

func TestMarshalUnmarshal(t *testing.T) {
	is := is.NewRelaxed(t)
	obj1 := TestObject{
		A: "hello",
		B: 2,
		C: []int64{31, 32, 33},
		D: map[int]string{41: "41"},
		E: map[int]bool{51: true},
	}

	b1, err := Marshal(obj1)
	is.NoErr(err)

	b2, err := Marshal(&obj1)
	is.NoErr(err)
	is.Equal(b1, b2)

	var val1 TestObject
	err = Unmarshal(b2, &val1)
	is.True(err == nil || err == io.EOF)
	is.Equal(obj1, val1)
}

func TestCompactMarshalUnmarshal(t *testing.T) {
	is := is.NewRelaxed(t)
	obj1 := TestObject{
		A: "hello",
		B: 2,
		C: []int64{31, 32, 33},
		D: map[int]string{41: "41"},
		E: map[int]bool{51: true},
	}

	b1, err := MarshalCompact(obj1)
	is.NoErr(err)

	b2, err := MarshalCompact(&obj1)
	is.NoErr(err)
	is.Equal(b1, b2)

	var val1 TestObject
	err = UnmarshalCompact(b2, &val1)
	is.True(err == nil || err == io.EOF)
	is.Equal(obj1, val1)
}
