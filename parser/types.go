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
	D         *Document
	FinalType interface{}
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
	if finalType := t.GetFinalType(); finalType != nil {
		switch x := finalType.(type) {
		case *Type:
			return x.TType()
		case *Enum:
			return I32
		default: // Struct, Union, Exception
			return STRUCT
		}
	}
	return ToTType(t.Name)
}

func (t *Type) GetFinalType() interface{} {
	if t.FinalType == nil {
		if t.D != nil && t.D.IdentTypes != nil {
			if x := t.D.IdentTypes[t.Name]; x != nil {
				t.FinalType = x.FinalType
			}
		}
	}
	return t.FinalType
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

	ReqRequired = "required"
	ReqOptional = "optional"
	ReqDefault  = "default"
)

type ConstValue struct {
	Type  string // Double, Int, Literal or Identifier
	Value string
}

func (c *ConstValue) IsZero() bool {
	switch c.Type {
	case ConstTypeInt:
		return c.Value == "0"
	case ConstTypeDouble:
		return c.Value == "0" || c.Value == ".0" || c.Value == "0.0"
	case ConstTypeLiteral:
		return c.Value == ""
	}
	return false
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

func (f *Field) IsDefaultZero() bool {
	if c, ok := f.Default.(ConstValue); ok {
		if f.Type.Name == "bool" {
			return c.Value == "false"
		}
		return c.IsZero()
	}
	return false
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
	IdentTypes map[string]*Type // for generator
	Typedefs   []*Typedef
	Constants  []*Constant
	Enums      []*Enum
	Structs    []*Struct
	Exceptions []*Struct
	Unions     []*Struct
	Services   []*Service
}
