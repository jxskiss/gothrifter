package parser

// Type constants in the Thrift protocol
type TType byte

const (
	STOP   TType = 0
	VOID   TType = 1
	BOOL   TType = 2
	BYTE   TType = 3
	I08    TType = 3
	DOUBLE TType = 4
	I16    TType = 6
	I32    TType = 8
	I64    TType = 10
	STRING TType = 11
	UTF7   TType = 11
	STRUCT TType = 12
	MAP    TType = 13
	SET    TType = 14
	LIST   TType = 15
	UTF8   TType = 16
	UTF16  TType = 17
	BINARY TType = 18

	UNKNOWN TType = 255
)

func (p TType) String() string {
	switch p {
	case STOP:
		return "STOP"
	case VOID:
		return "VOID"
	case BOOL:
		return "BOOL"
	case BYTE:
		return "BYTE"
	case DOUBLE:
		return "DOUBLE"
	case I16:
		return "I16"
	case I32:
		return "I32"
	case I64:
		return "I64"
	case STRING:
		return "STRING"
	case STRUCT:
		return "STRUCT"
	case MAP:
		return "MAP"
	case SET:
		return "SET"
	case LIST:
		return "LIST"
	case UTF8:
		return "UTF8"
	case UTF16:
		return "UTF16"
	case BINARY:
		return "BINARY"
	}
	return "UNKNOWN"
}

var tTypeMap = map[string]TType{
	"bool":   BOOL,
	"byte":   BYTE,
	"i8":     I08,
	"i16":    I16,
	"i32":    I32,
	"i64":    I64,
	"double": DOUBLE,
	"string": STRING,
	"binary": BINARY,
	"map":    MAP,
	"set":    SET,
	"list":   LIST,
}

func ToTType(name string) TType {
	if t, ok := tTypeMap[name]; ok {
		return t
	}
	return UNKNOWN
}
