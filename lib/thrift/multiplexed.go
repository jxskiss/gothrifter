package thrift

import (
	"github.com/thrift-iterator/go/protocol"
	"github.com/thrift-iterator/go/spi"
)

type MultiplexedStream struct {
	spi.Stream
	ServiceName string
}

const MULTIPLEXED_SEPARATOR = ":"

func NewMultiplexedStream(stream spi.Stream, serviceName string) *MultiplexedStream {
	return &MultiplexedStream{
		Stream:      stream,
		ServiceName: serviceName,
	}
}

func (s *MultiplexedStream) WriteMessageHeader(header protocol.MessageHeader) {
	header.MessageName = s.ServiceName + MULTIPLEXED_SEPARATOR + header.MessageName
	s.Stream.WriteMessageHeader(header)
}