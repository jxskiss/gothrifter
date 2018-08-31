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

package thrift2

// Type constants in the Thrift protocol
type Type byte

const (
	STOP   = 0
	VOID   = 1
	BOOL   = 2
	BYTE   = 3
	I08    = 3
	DOUBLE = 4
	I16    = 6
	I32    = 8
	I64    = 10
	STRING = 11
	UTF7   = 11
	STRUCT = 12
	MAP    = 13
	SET    = 14
	LIST   = 15
	UTF8   = 16
	UTF16  = 17
	BINARY = 18
	FLOAT  = 19
)

var typeNames = map[int]string{
	STOP:   "STOP",
	VOID:   "VOID",
	BOOL:   "BOOL",
	BYTE:   "BYTE",
	DOUBLE: "DOUBLE",
	I16:    "I16",
	I32:    "I32",
	I64:    "I64",
	STRING: "STRING",
	STRUCT: "STRUCT",
	MAP:    "MAP",
	SET:    "SET",
	LIST:   "LIST",
	UTF8:   "UTF8",
	UTF16:  "UTF16",
	BINARY: "BINARY",
	FLOAT:  "FLOAT",
}
func (p Type) String() string {
	if s, ok := typeNames[int(p)]; ok {
		return s
	}
	return "Unknown"
}

// Message type constants in the Thrift protocol
type MessageType int32

const (
	INVALID_MESSAGE_TYPE MessageType = 0
	CALL                 MessageType = 1
	REPLY                MessageType = 2
	EXCEPTION            MessageType = 3
	ONEWAY               MessageType = 4
)

// Protocols

type ProtocolID int16

const (
	ProtocolIDBinary     ProtocolID = 0
	ProtocolIDJSON       ProtocolID = 1
	ProtocolIDCompact    ProtocolID = 2
	ProtocolIDDebug      ProtocolID = 3
	ProtocolIDVirtual    ProtocolID = 4
	ProtocolIDSimpleJSON ProtocolID = 5
)

func (p ProtocolID) String() string {
	switch p {
	case ProtocolIDBinary:
		return "binary"
	case ProtocolIDJSON:
		return "json"
	case ProtocolIDCompact:
		return "compact"
	case ProtocolIDDebug:
		return "debug"
	case ProtocolIDVirtual:
		return "virtual"
	case ProtocolIDSimpleJSON:
		return "simplejson"
	default:
		return "unknown"
	}
}

// Binary protocol

const (
	BinaryVersionMask uint32 = 0xffff0000
	BinaryVersion1    uint32 = 0x80010000

	VERSION_MASK = BinaryVersionMask // deprecated alias of BinaryVersionMask
	VERSION_1    = BinaryVersion1    // deprecated alias of BinaryVersion1
)

// Compact protocol

const (
	COMPACT_PROTOCOL_ID       = 0x082
	COMPACT_VERSION           = 0x01
	COMPACT_VERSION_BE        = 0x02
	COMPACT_VERSION_MASK      = 0x1f
	COMPACT_TYPE_MASK         = 0x0E0
	COMPACT_TYPE_BITS         = 0x07
	COMPACT_TYPE_SHIFT_AMOUNT = 5
)

type compactType byte

const (
	COMPACT_BOOLEAN_TRUE  = 0x01
	COMPACT_BOOLEAN_FALSE = 0x02
	COMPACT_BYTE          = 0x03
	COMPACT_I16           = 0x04
	COMPACT_I32           = 0x05
	COMPACT_I64           = 0x06
	COMPACT_DOUBLE        = 0x07
	COMPACT_BINARY        = 0x08
	COMPACT_LIST          = 0x09
	COMPACT_SET           = 0x0A
	COMPACT_MAP           = 0x0B
	COMPACT_STRUCT        = 0x0C
	COMPACT_FLOAT         = 0x0D
)

func (t compactType) toType() Type {
	switch byte(t) & 0x0f {
	case STOP:
		return STOP
	case COMPACT_BOOLEAN_FALSE, COMPACT_BOOLEAN_TRUE:
		return BOOL
	case COMPACT_BYTE:
		return BYTE
	case COMPACT_I16:
		return I16
	case COMPACT_I32:
		return I32
	case COMPACT_I64:
		return I64
	case COMPACT_DOUBLE:
		return DOUBLE
	case COMPACT_FLOAT:
		return FLOAT
	case COMPACT_BINARY:
		return STRING
	case COMPACT_LIST:
		return LIST
	case COMPACT_SET:
		return SET
	case COMPACT_MAP:
		return MAP
	case COMPACT_STRUCT:
		return STRUCT
	}
	return STOP
}

var compactTypes = map[Type]compactType{
	STOP:   STOP,
	BOOL:   COMPACT_BOOLEAN_TRUE,
	BYTE:   COMPACT_BYTE,
	I16:    COMPACT_I16,
	I32:    COMPACT_I32,
	I64:    COMPACT_I64,
	DOUBLE: COMPACT_DOUBLE,
	FLOAT:  COMPACT_FLOAT,
	STRING: COMPACT_BINARY,
	LIST:   COMPACT_LIST,
	SET:    COMPACT_SET,
	MAP:    COMPACT_MAP,
	STRUCT: COMPACT_STRUCT,
}
