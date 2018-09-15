package thrift

import (
	"bytes"
	"encoding/binary"
	"github.com/matryer/is"
	"io/ioutil"
	"math/rand"
	"testing"
)

func TestFramed(t *testing.T) {
	is := is.New(t)
	var buf bytes.Buffer
	var framed = NewFramedTransport(&buf, 1500)

	var writeb []byte
	for i := 0; i < 100; i++ {
		b := make([]byte, 10+rand.Intn(100))
		rand.Read(b)
		if len(writeb)+len(b) > 1500 {
			break
		}
		writeb = append(writeb, b...)
		n, err := framed.Write(b)
		is.NoErr(err)
		is.Equal(n, len(b))
	}
	err := framed.Flush()
	is.NoErr(err)

	b, err := ioutil.ReadAll(framed)
	is.NoErr(err)
	is.Equal(writeb, b)

	framed.Reset(new(bytes.Buffer))
	_, err = framed.Write(make([]byte, 1501))
	is.Equal(err, ErrMaxFrameSize)
	_, err = framed.Write(make([]byte, 1500))
	is.NoErr(err)

	b = make([]byte, 1505)
	binary.BigEndian.PutUint32(b[:4], uint32(1501))
	framed.Reset(bytes.NewBuffer(b))
	_, err = ioutil.ReadAll(framed)
	is.Equal(err, ErrMaxFrameSize)
}
