package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

var _ fmt.Stringer

const UNSETID = -1

func (r pegRule) String() string { return rul3s[r] }

func assertRule(node *node32, rule pegRule) *node32 {
	if node.pegRule != rule {
		panic(fmt.Sprintf("mismatch rule: %v, %v", node.pegRule, rule))
	}
	return node.up
}

func (p *Thrift) text(n *node32) string {
	return strings.TrimSpace(string(p.buffer[int(n.begin):int(n.end)]))
}

func (p *Thrift) parsePegText(node *node32) string {
	for n := node; n != nil; n = n.next {
		n1 := n
		for n1 != nil && n1.pegRule != rulePegText {
			n1 = n1.up
		}
		if n1 == nil || n1.pegRule != rulePegText {
			continue
		}
		if text := p.text(n1); text != "" {
			return text
		}
	}
	return ""
}

func (p *Thrift) parseHeader(node *node32) interface{} {
	node = assertRule(node, ruleHeader)
	switch node.pegRule {
	case ruleNamespace:
		return p.parseNamespace(node)
	case ruleInclude:
		return p.parseInclude(node)
	case ruleCppInclude:
		// unimplemented
	default:
		panic("unknown header rule: " + node.pegRule.String())
	}
	return nil
}

func (p *Thrift) parseNamespace(node *node32) *Namespace {
	node = assertRule(node, ruleNamespace)
	node = node.next // skip "namespace"
	scope := p.parsePegText(node)
	name := p.parsePegText(node.next)
	return &Namespace{Scope: scope, Name: name}
}

func (p *Thrift) parseInclude(node *node32) *Include {
	node = assertRule(node, ruleInclude)
	node = node.next // skip "include"
	filename := p.parsePegText(node)
	return &Include{filename: filename}
}

func (p *Thrift) parseDefinition(node *node32) interface{} {
	node = assertRule(node, ruleDefinition)
	// Const / Typedef / Enum / Senum / Struct / Union / Exception / Service
	switch node.pegRule {
	case ruleConst:
		return p.parseConst(node)
	case ruleTypedef:
		return p.parseTypedef(node)
	case ruleEnum:
		return p.parseEnum(node)
	case ruleSenum:
		// unimplemented
	case ruleStruct:
		return p.parseStruct(node)
	case ruleUnion:
		return p.parseUnion(node)
	case ruleException:
		return p.parseException(node)
	case ruleService:
		return p.parseService(node)
	default:
		panic("unknown definition rule: " + node.pegRule.String())
	}
	return nil
}

func (p *Thrift) parseConst(node *node32) *Constant {
	node = assertRule(node, ruleConst)
	// CONST FieldType Identifier EQUAL ConstValue ListSeparator?
	node = node.next // skip "const"
	ft := p.parseFieldType(node)
	node = node.next
	name := p.parsePegText(node)
	node = node.next.next // skip "="
	value := p.parseConstValue(node)
	return &Constant{Name: name, Type: ft, Value: value}
}

func (p *Thrift) parseFieldType(node *node32) *Type {
	node = assertRule(node, ruleFieldType)
	// BaseType / ContainerType / Identifier
	switch node.pegRule {
	case ruleBaseType:
		return &Type{Name: p.parsePegText(node), Category: TypeBasic}
	case ruleContainerType:
		return p.parseContainerType(node)
	case ruleIdentifier:
		return &Type{Name: p.parsePegText(node), Category: TypeIdentifier, D: p.D}
	default:
		panic("unknown field type rule: " + node.pegRule.String())
	}
}

func (p *Thrift) parseContainerType(node *node32) *Type {
	node = assertRule(node, ruleContainerType)
	// MapType / SetType / ListType
	switch node.pegRule {
	case ruleMapType: // MAP CppType? LPOINT FieldType COMMA FieldType RPOINT
		node = node.up.next.next
		kt := p.parseFieldType(node)
		node = node.next.next
		vt := p.parseFieldType(node)
		return &Type{Name: "map", Category: TypeContainer, KeyType: kt, ValueType: vt}
	case ruleSetType: // SET CppType? LPOINT FieldType RPOINT
		node = node.up.next.next
		vt := p.parseFieldType(node)
		return &Type{Name: "set", Category: TypeContainer, ValueType: vt}
	case ruleListType: // LIST LPOINT FieldType RPOINT CppType?
		node = node.up.next.next
		vt := p.parseFieldType(node)
		return &Type{Name: "list", Category: TypeContainer, ValueType: vt}
	default:
		panic("unknown container type rule: " + node.pegRule.String())
	}
}

func (p *Thrift) parseConstValue(node *node32) interface{} {
	node = assertRule(node, ruleConstValue)
	// IntConstant / DoubleConstant / Literal / Identifier / ConstList / ConstMap
	switch node.pegRule {
	case ruleIntConstant:
		return ConstValue{Type: ConstTypeInt, Value: p.parsePegText(node)}
	case ruleDoubleConstant:
		return ConstValue{Type: ConstTypeDouble, Value: p.parsePegText(node)}
	case ruleLiteral:
		return ConstValue{Type: ConstTypeLiteral, Value: p.parsePegText(node)}
	case ruleIdentifier:
		return ConstValue{Type: ConstTypeIdentifier, Value: p.parsePegText(node)}
	case ruleConstList:
		var ret []interface{}
		for n := node.up; n != nil; n = n.next {
			if n.pegRule == ruleConstValue {
				ret = append(ret, p.parseConstValue(n))
			}
		}
		return ret
	case ruleConstMap:
		node = node.up
		var ret []MapConstValue
		for n := node.next; n != nil; n = n.next {
			if n.pegRule != ruleConstValue {
				continue
			}
			k := p.parseConstValue(n)
			n = n.next.next
			v := p.parseConstValue(n)
			ret = append(ret, MapConstValue{Key: k, Value: v})
		}
		return ret
	default:
		panic("unknown const value rule: " + node.pegRule.String())
	}
}

func (p *Thrift) parseTypedef(node *node32) *Typedef {
	node = assertRule(node, ruleTypedef)
	// TYPEDEF DefinitionType Identifier
	node = node.next // skip "typedef"
	typ := p.parseDefinitionType(node)
	node = node.next
	alias := p.parsePegText(node)
	return &Typedef{Type: typ, Alias: alias}
}

func (p *Thrift) parseDefinitionType(node *node32) *Type {
	node = assertRule(node, ruleDefinitionType)
	// BaseType / ContainerType / Identifier
	switch node.pegRule {
	case ruleBaseType:
		return &Type{Name: p.parsePegText(node), Category: TypeBasic}
	case ruleContainerType:
		return p.parseContainerType(node)
	case ruleIdentifier:
		return &Type{Name: p.parsePegText(node), Category: TypeIdentifier, D: p.D}
	default:
		panic("unknown definition type rule: " + node.pegRule.String())
	}
}

func (p *Thrift) parseEnum(node *node32) *Enum {
	node = assertRule(node, ruleEnum)
	// ENUM Identifier LWING (Identifier (EQUAL IntConstant)? ListSeparator?)* RWING
	node = node.next // skip "enum"
	name := p.parsePegText(node)
	var values = make([]*EnumValue, 0)
	var preValue int
	for n := node.next.next; n != nil; n = n.next {
		if n.pegRule == ruleIdentifier {
			var v EnumValue
			v.Name = p.parsePegText(n)
			if p.parsePegText(n.next) == "=" {
				n = n.next.next
				v.Value, _ = strconv.Atoi(p.parsePegText(n))
			} else {
				if len(values) == 0 {
					v.Value = 1
				} else {
					v.Value = preValue + 1
				}
			}
			preValue = v.Value
			values = append(values, &v)
		}
	}
	return &Enum{Name: name, Values: values}
}

func (p *Thrift) parseStruct(node *node32) *Struct {
	node = assertRule(node, ruleStruct)
	// STRUCT Identifier XSD_ALL? LWING Field* RWING
	node = node.next // skip "struct"
	name := p.parsePegText(node)
	fields := p.parseFields(node.next)
	return &Struct{Name: name, Fields: fields}
}

func (p *Thrift) parseUnion(node *node32) Union {
	node = assertRule(node, ruleUnion)
	// UNION Identifier XSD_ALL? LWING Field* RWING
	node = node.next // skip "union"
	name := p.parsePegText(node)
	fields := p.parseFields(node.next)
	return &Struct{Name: name, Fields: fields}
}

// TODO
func (p *Thrift) parseException(node *node32) Exception {
	node = assertRule(node, ruleException)
	// EXCEPTION Identifier LWING Field* RWING
	node = node.next // skip "exception"
	name := p.parsePegText(node)
	fields := p.parseFields(node.next)
	return &Struct{Name: name, Fields: fields}
}

func (p *Thrift) parseFields(node *node32) []*Field {
	var fields []*Field
	for n := node; n != nil; n = n.next {
		if n.pegRule == ruleField {
			field := p.parseField(n)
			if field.ID == UNSETID {
				if len(fields) > 0 {
					field.ID = fields[len(fields)-1].ID + 1
				} else {
					field.ID = 1
				}
			}
			fields = append(fields, field)
		}
	}
	return fields
}

func (p *Thrift) parseField(node *node32) *Field {
	node = assertRule(node, ruleField)
	// FieldID? FieldReq? FieldType Identifier (EQUAL ConstValue)? XsdFieldOptions ListSeparator?
	var f Field
	f.ID = UNSETID
	f.Requiredness = RequirednessDefault
	f.Optional = false
	for n := node; n != nil; n = n.next {
		switch n.pegRule {
		case ruleFieldID:
			f.ID, _ = strconv.Atoi(strings.Split(p.parsePegText(n), ":")[0])
		case ruleFieldReq:
			f.Requiredness, f.Optional = p.parseFieldReqOptional(n)
		case ruleFieldType:
			f.Type = p.parseFieldType(n)
		case ruleIdentifier:
			f.Name = p.parsePegText(n)
		case ruleConstValue:
			f.Default = p.parseConstValue(n)
		}
	}
	return &f
}

func (p *Thrift) parseFieldReqOptional(node *node32) (string, bool) {
	node = assertRule(node, ruleFieldReq)
	r := p.parsePegText(node)
	switch r {
	case "required":
		return RequirednessRequired, false
	case "optional":
		return RequirednessOptional, true
	case "":
		return RequirednessDefault, false
	default:
		panic("invalid field requiredness literal")
	}
}

func (p *Thrift) parseService(node *node32) *Service {
	node = assertRule(node, ruleService)
	// SERVICE Identifier ( EXTENDS Identifier )? LWING Function* RWING
	node = node.next // skip "service"
	var s = Service{D: p.D}
	s.Name = p.parsePegText(node)
	node = node.next
	if node.pegRule == ruleEXTENDS {
		s.Extends = p.parsePegText(node)
		node = node.next
	}
	for n := node.next; n != nil; n = n.next {
		if n.pegRule == ruleFunction {
			m := p.parseFunction(n.up)
			s.Methods = append(s.Methods, m)
		}
	}
	return &s
}

func (p *Thrift) parseFunction(node *node32) *Method {
	// ONEWAY? FunctionType Identifier LPAR Field* RPAR Throws? ListSeparator?
	var f Method
	for ; node != nil; node = node.next {
		switch node.pegRule {
		case ruleONEWAY:
			f.Oneway = true
		case ruleFunctionType:
			f.ReturnType = p.parseFunctionType(node)
		case ruleIdentifier:
			f.Name = p.parsePegText(node)
		case ruleField:
			field := p.parseField(node)
			if field.ID == UNSETID {
				if len(f.Arguments) > 0 {
					field.ID = f.Arguments[len(f.Arguments)-1].ID + 1
				} else {
					field.ID = 1
				}
			}
			f.Arguments = append(f.Arguments, field)
		case ruleThrows:
			f.Exceptions = p.parseThrows(node)
		}
	}
	return &f
}

func (p *Thrift) parseFunctionType(node *node32) *Type {
	node = assertRule(node, ruleFunctionType)
	// FieldType / VOID
	if node.pegRule == ruleFieldType {
		return p.parseFieldType(node)
	}
	if node.pegRule == ruleVOID {
		return &Type{Name: "void"}
	}
	panic("invalid function type: " + p.text(node))
}

func (p *Thrift) parseThrows(node *node32) []*Field {
	node = assertRule(node, ruleThrows)
	// THROWS LPAR Field* RPAR
	node = node.next // skip "throws"
	fields := p.parseFields(node)
	return fields
}

func (d *Document) Parse() error {
	d.RefName = strings.Split(filepath.Base(d.Filename), ".")[0]

	b, err := ioutil.ReadFile(d.Filename)
	if err != nil {
		return err
	}
	t := Thrift{
		D:      d,
		Buffer: string(b),
		Pretty: false,
	}
	t.Init()
	if err := t.Parse(); err != nil {
		return err
	}

	// Parse the AST tree.
	root := t.AST()
	if root.pegRule != ruleDocument {
		return fmt.Errorf("invalid root rule: %v", root.pegRule)
	}
	for n := root.up; n != nil; n = n.next {
		switch n.pegRule {
		case ruleSpacing:
			continue
		case ruleHeader:
			header := t.parseHeader(n)
			switch hh := header.(type) {
			case *Namespace:
				d.Namespaces[hh.Scope] = hh
			case *Include:
				fn := hh.filename
				hh.RefName = strings.SplitN(filepath.Base(fn), ".", 2)[0]
				absPath := AbsPath(filepath.Join(filepath.Dir(d.Filename), fn))
				hh.AbsPath = absPath
				d.Includes[hh.RefName] = hh
			}
		case ruleDefinition:
			def := t.parseDefinition(n)
			switch dd := def.(type) {
			case *Constant:
				d.Constants = append(d.Constants, dd)
			case *Typedef:
				d.Typedefs = append(d.Typedefs, dd)
			case *Enum:
				d.Enums = append(d.Enums, dd)
			case *Struct:
				d.Structs = append(d.Structs, dd)
			case Union:
				d.Unions = append(d.Unions, dd)
			case Exception:
				d.Exceptions = append(d.Exceptions, dd)
			case *Service:
				d.Services = append(d.Services, dd)
			}
		default:
			panic("unknown document rule: " + n.pegRule.String())
		}
	}

	return nil
}

func Parse(fn string) (*Document, error) {
	fn = AbsPath(fn)
	doc := &Document{
		Filename: fn,

		Includes:   make(map[string]*Include),
		Namespaces: make(map[string]*Namespace),
	}
	if err := doc.Parse(); err != nil {
		return nil, err
	}
	return doc, nil
}

func AbsPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return filepath.Clean(path)
}
