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
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Generic Thrift exception
type Exception interface {
	error
}

// Application exception types.
const (
	UNKNOWN_APPLICATION_EXCEPTION  = 0
	UNKNOWN_METHOD                 = 1
	INVALID_MESSAGE_TYPE_EXCEPTION = 2
	WRONG_METHOD_NAME              = 3
	BAD_SEQUENCE_ID                = 4
	MISSING_RESULT                 = 5
	INTERNAL_ERROR                 = 6
	PROTOCOL_ERROR                 = 7
)

type ApplicationException struct {
	Message string `thrift:"message,1"`
	TypeId  int32  `thrift:"type_id,2"`
}

func NewApplicationException(t int32, m string) error {
	return &ApplicationException{TypeId: t, Message: m}
}

func NewApplicationExceptionFromError(err error) *ApplicationException {
	if ae, ok := err.(*ApplicationException); ok {
		return ae
	}
	return &ApplicationException{Message: err.Error()}
}

func (e *ApplicationException) Error() string {
	return fmt.Sprintf("ApplicationException: %s", e.Message)
}

// Protocol exception types.
const (
	UNKNOWN_PROTOCOL_EXCEPTION = 0
	INVALID_DATA               = 1
	NEGATIVE_SIZE              = 2
	SIZE_LIMIT                 = 3
	BAD_VERSION                = 4
	NOT_IMPLEMENTED            = 5
	DEPTH_LIMIT                = 6
)

type ProtocolException interface {
	Exception
	TypeId() int
}

type protocolException struct {
	typeId  int
	message string
}

func (e *protocolException) TypeId() int { return e.typeId }

func (e *protocolException) String() string { return e.message }

func (e *protocolException) Error() string { return e.message }

func NewProtocolException(err error) ProtocolException {
	if err == nil {
		return nil
	}
	if e, ok := err.(ProtocolException); ok {
		return e
	}
	if _, ok := err.(base64.CorruptInputError); ok {
		return &protocolException{typeId: INVALID_DATA, message: err.Error()}
	}
	return &protocolException{typeId: UNKNOWN_PROTOCOL_EXCEPTION, message: err.Error()}
}

func NewProtocolExceptionWithType(typeId int, err error) ProtocolException {
	if err == nil {
		return nil
	}
	return &protocolException{typeId: typeId, message: err.Error()}
}

// Transport exception types.
const (
	UNKNOWN_TRANSPORT_EXCEPTION = 0
	NOT_OPEN                    = 1
	ALREADY_OPEN                = 2
	TIMED_OUT                   = 3
	END_OF_FILE                 = 4
	INTERRUPTED                 = 5
	BAD_ARGS                    = 6
	CORRUPTED_DATA              = 7
	NOT_SUPPORTED               = 9
	INVALID_STATE               = 10
	INVALID_FRAME_SIZE          = 11
	SSL_ERROR                   = 12
	COULD_NOT_BIND              = 13
	SASL_HANDSHAKE_TIMEOUT      = 14
	NETWORK_ERROR               = 15
)

type TransportException interface {
	Exception
	TypeId() int
	Err() error
}

type transportException struct {
	typeId int
	err    error
}

type timeoutable interface {
	Timeout() bool
}

func (e *transportException) Error() string { return e.err.Error() }

func (e *transportException) TypeId() int { return e.typeId }

func (e *transportException) Err() error { return e.err }

func NewTransportException(t int, e string) TransportException {
	return &transportException{typeId: t, err: errors.New(e)}
}

func NewTransportExceptionFromError(e error) TransportException {
	if e == nil {
		return nil
	}
	if t, ok := e.(TransportException); ok {
		return t
	}

	switch v := e.(type) {
	case TransportException:
		return v
	case timeoutable:
		if v.Timeout() {
			return &transportException{typeId: TIMED_OUT, err: e}
		}
	}

	if e == io.EOF || strings.Contains(e.Error(), "EOF") {
		return &transportException{typeId: END_OF_FILE, err: e}
	}

	return &transportException{typeId: UNKNOWN_TRANSPORT_EXCEPTION, err: e}
}
