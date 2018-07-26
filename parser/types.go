// Copyright 2012-2015 Samuel Stauffer. All rights reserved.
// Use of this source code is governed by a 3-clause BSD
// license that can be found in the LICENSE file.
//
// https://github.com/samuel/go-thrift/blob/master/parser/types.go

package parser

import "fmt"

const (
	TypeBasic      = "basic"
	TypeContainer  = "container"
	TypeIdentifier = "identifier"
)

type Type struct {
	Name        string
	Category    string
	KeyType     *Type // map
	ValueType   *Type // map, list, or set
	Annotations []*Annotation

	// for generator
	D *Document
}

func (t *Type) String() string {
	switch t.Name {
	case "map":
		return fmt.Sprintf("map<%v,%v>", t.KeyType, t.ValueType)
	case "set":
		return fmt.Sprintf("set<%v>", t.ValueType)
	case "list":
		return fmt.Sprintf("list<%v>", t.ValueType)
	}
	return t.Name
}

func (t *Type) TType() TType {
	return ToTType(t.Name)
}

type Namespace struct {
	Scope string
	Name  string
}

type Include struct {
	filename string

	// for generator
	RefName string
	AbsPath string
}

//type Include string

type Exception *Struct

type Union *Struct

type Typedef struct {
	*Type

	Alias       string
	Annotations []*Annotation
}

type EnumValue struct {
	Name        string
	Value       int
	Annotations []*Annotation
}

type Enum struct {
	Name        string
	Values      []*EnumValue
	Annotations []*Annotation
}

const (
	ConstTypeDouble     = "Double"
	ConstTypeInt        = "Int"
	ConstTypeLiteral    = "Literal"
	ConstTypeIdentifier = "Identifier"

	RequirednessRequired = "required"
	RequirednessOptional = "optional"
	RequirednessDefault  = "default"
)

type ConstValue struct {
	Type  string // Double, Int, Literal or Identifier
	Value string
}

type ListConstValue []interface{}

type MapConstValue struct {
	// for const map value
	Key, Value interface{}
}

type Constant struct {
	Name  string
	Type  *Type
	Value interface{} // ConstValue, ListConstValue or []MapConstValue
}

type Field struct {
	ID           int
	Name         string
	Optional     bool
	Requiredness string
	Type         *Type
	Default      interface{}
	Annotations  []*Annotation
}

type Struct struct {
	Name        string
	Fields      []*Field
	Annotations []*Annotation
}

type Method struct {
	Comment     string
	Name        string
	Oneway      bool
	ReturnType  *Type
	Arguments   []*Field
	Exceptions  []*Field
	Annotations []*Annotation
}

type Service struct {
	Name        string
	Extends     string
	Methods     []*Method
	Annotations []*Annotation

	// for generator
	D *Document
}

type KeyValue struct {
	Key, Value interface{}
}

type Annotation struct {
	Name  string
	Value string
}

type Document struct {
	Filename string
	RefName  string

	Includes   map[string]*Include
	Namespaces map[string]*Namespace
	Typedefs   []*Typedef
	Constants  []*Constant
	Enums      []*Enum
	Structs    []*Struct
	Exceptions []*Struct
	Unions     []*Struct
	Services   []*Service
}
