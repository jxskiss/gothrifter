package reflection

import (
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func isEnumType(valType reflect.Type) bool {
	if valType.Kind() != reflect.Int64 {
		return false
	}
	_, hasStringMethod := valType.MethodByName("String")
	return hasStringMethod
}

func parseFieldId(refField reflect.StructField) int16 {
	if !unicode.IsUpper(rune(refField.Name[0])) {
		return -1
	}
	thriftTag := refField.Tag.Get("thrift")
	if thriftTag == "" {
		return -1
	}
	tags := strings.Split(thriftTag, ",")
	if len(tags) < 2 {
		return -1
	}
	fieldId, err := strconv.Atoi(tags[1])
	if err != nil {
		return -1
	}
	return int16(fieldId)
}

func parseMapType(refField reflect.StructField) thrift.Type {
	if refField.Type.Elem().Kind() != reflect.Bool {
		return thrift.MAP
	}
	thriftTag := refField.Tag.Get("thrift")
	tags := strings.Split(thriftTag, ",")
	if len(tags) > 2 {
		for _, tag := range tags[2:] {
			if strings.TrimSpace(tag) == "map" {
				return thrift.MAP
			}
		}
	}
	// By default, consider map with boolean value as SET.
	return thrift.SET
}
