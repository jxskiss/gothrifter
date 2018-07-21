/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package thrift

import (
	"errors"
	"io"
)

var errTransportInterrupted = errors.New("Transport Interrupted")

// Flusher is the interface that wraps the basic Flush method
type Flusher interface {
	Flush() (err error)
}

// ReadSizeProvider is the interface that wraps the basic RemainingBytes method
type ReadSizeProvider interface {
	RemainingBytes() (numBytes uint64)
}

// Transport is an encapsulation of the I/O layer
type Transport interface {
	io.ReadWriteCloser
	Flusher
	ReadSizeProvider

	// Opens the transport for communication
	Open() error

	// Returns true if the transport is open
	IsOpen() bool
}

type stringWriter interface {
	WriteString(s string) (n int, err error)
}

// RichTransport is an "enhanced" transport with extra capabilities.
// You need to use one of these to construct protocol.
// Notably, Socket does not implement this interface, and it is always a mistake to use
// Socket directly in protocol.
type RichTransport interface {
	io.ReadWriter
	io.ByteReader
	io.ByteWriter
	stringWriter
	Flusher
	ReadSizeProvider
}

// TransportFactory is used to create wrapped instance of Transports.
// This is used primarily in servers, which get Transports from
// a ServerTransport and then may want to mutate them (i.e. create
// a BufferedTransport from the underlying base transport)
type TransportFactory interface {
	GetTransport(trans Transport) Transport
}

type transportFactory struct{}

// Return a wrapped instance of the base Transport.
func (p *transportFactory) GetTransport(trans Transport) Transport {
	return trans
}

// NewTransportFactory returns a new TransportFactory
func NewTransportFactory() TransportFactory {
	return &transportFactory{}
}

// Server transport. Object which provides client transports.
type ServerTransport interface {
	Listen() error
	Accept() (Transport, error)
	Close() error

	// Optional method implementation. This signals to the Server transport
	// that it should break out of any accept() or listen() that it is currently
	// blocked on. This method, if implemented, MUST be thread safe, as it may
	// be called from a different thread context than the other ServerTransport
	// methods.
	Interrupt() error
}