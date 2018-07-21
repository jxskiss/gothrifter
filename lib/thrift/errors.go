package thrift

import "errors"

var (
	ErrMaxBufferLen    = errors.New("thrift: max buffer len exceeded")
	ErrMaxMapElements  = errors.New("thrift: max map elements exceeded")
	ErrMaxSetElements  = errors.New("thrift: max set elements exceeded")
	ErrMaxListElements = errors.New("thrift: max list elements exceeded")
	ErrUnknownFunction = errors.New("thrift: unknown function")
	ErrMessageType     = errors.New("thrift: error message type")
	ErrFieldType       = errors.New("thrift: error field type")
	ErrVersion         = errors.New("thrift: unknown version")
	ErrSeqMismatch     = errors.New("thrift: seq mismatch")

	errTooManyConn = errors.New("thrift: too many connections")
	errPeerClosed  = errors.New("thrift: peer closed")
	errClosed      = errors.New("thrift: conn closed")
	errWorking     = errors.New("thrift: working")
)
