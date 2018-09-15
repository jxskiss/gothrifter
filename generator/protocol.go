package generator

import (
	"bytes"
	"fmt"
	"github.com/jxskiss/thriftkit/parser"
	"strings"
)

func (g *Generator) parseRefType(typ *parser.Type) (refName, typeName string) {
	parts := strings.SplitN(typ.Name, ".", 2)
	if len(parts) == 1 {
		return "", ToCamelCase(parts[0])
	}
	return parts[0], ToCamelCase(parts[1])
}

func (g *Generator) formatRead(typ *parser.Type, variable string) (string, error) {
	//fieldName := ToCamelCase(field.Name)
	//typeName := ToCamelCase(typ.Name)
	var buf bytes.Buffer
	switch typ.Category {
	case parser.TypeBasic:
		readFunc := ""
		switch typ.TType() {
		case parser.BOOL:
			readFunc = "ReadBool"
		case parser.BYTE:
			readFunc = "ReadByte"
		case parser.I16:
			readFunc = "ReadI16"
		case parser.I32:
			readFunc = "ReadI32"
		case parser.I64:
			readFunc = "ReadI64"
		case parser.DOUBLE:
			readFunc = "ReadDouble"
		case parser.STRING:
			readFunc = "ReadString"
		case parser.BINARY:
			readFunc = "ReadBinary"
		case parser.FLOAT:
			readFunc = "ReadFloat"
		default:
			return "", fmt.Errorf("unsupported type: %v", typ.Name)
		}
		tmpl := "if %v, err = r.%v(); err != nil { return err }"
		return fmt.Sprintf(tmpl, variable, readFunc), nil
	case parser.TypeContainer:
		switch typ.Name {
		case "map":
			tmpl := "%v\n %v = m"
			if err := g.tmpl("read_map.tmpl").Execute(&buf, typ); err != nil {
				return "", err
			}
			return fmt.Sprintf(tmpl, buf.String(), variable), nil
		case "list":
			tmpl := "%v\n  %v = lst"
			if err := g.tmpl("read_list.tmpl").Execute(&buf, typ); err != nil {
				return "", err
			}
			return fmt.Sprintf(tmpl, buf.String(), variable), nil
		case "set":
			tmpl := "%v\n  %v = m"
			if err := g.tmpl("read_set.tmpl").Execute(&buf, typ); err != nil {
				return "", err
			}
			return fmt.Sprintf(tmpl, buf.String(), variable), nil
		}
		return "", fmt.Errorf("unsupported type: %v", typ.Name)
	case parser.TypeIdentifier:
		if g.isPtrType(typ) {
			refName, typeName := g.parseRefType(typ)
			if refName == "" {
				tmpl := "%v = New%v()\n  if err = %v.Read(r); err != nil { return err }"
				return fmt.Sprintf(tmpl, variable, typeName, variable), nil
			}
			tmpl := "%v = %v.New%v()\n  if err = %v.Read(r); err != nil { return err }"
			return fmt.Sprintf(tmpl, variable, refName, typeName, variable), nil
		}
		if finalType := typ.GetFinalType(); finalType != nil {
			if typ, ok := finalType.(*parser.Type); ok {
				return g.formatRead(typ, variable)
			}
			if _, ok := finalType.(*parser.Enum); ok {
				tt, _ := g.formatType(typ)
				tmpl := "if x, err := r.ReadI32(); err != nil { return err } else { %v = %v(x) }"
				return fmt.Sprintf(tmpl, variable, tt), nil
			}
			panic("not implemented")
		}
		return "", fmt.Errorf("unknown non-ptr type: %v", typ.Name)
	default:
		return "", fmt.Errorf("unknown type category: %v", typ.Category)
	}
}

func (g *Generator) formatWrite(typ *parser.Type, variable string) (string, error) {
	//fieldName := ToCamelCase(field.Name)
	typeName := ToCamelCase(typ.Name)
	switch typ.Category {
	case parser.TypeBasic:
		writeFunc := ""
		switch typ.TType() {
		case parser.BOOL:
			writeFunc = "WriteBool(%v)"
		case parser.BYTE:
			writeFunc = "WriteByte(byte(%v))"
		case parser.I16:
			writeFunc = "WriteI16(%v)"
		case parser.I32:
			writeFunc = "WriteI32(%v)"
		case parser.I64:
			writeFunc = "WriteI64(%v)"
		case parser.DOUBLE:
			writeFunc = "WriteDouble(%v)"
		case parser.STRING:
			writeFunc = "WriteString(%v)"
		case parser.BINARY:
			writeFunc = "WriteBinary(%v)"
		case parser.FLOAT:
			writeFunc = "WriteFloat(%v)"
		default:
			return "", fmt.Errorf("unsupported type: %v", typ.Name)
		}
		writeFunc = fmt.Sprintf(writeFunc, variable)
		tmpl := "if err = w.%v; err != nil { return err }"
		return fmt.Sprintf(tmpl, writeFunc), nil
	case parser.TypeContainer:
		var buf bytes.Buffer
		switch typ.Name {
		case "map":
			tmpl := "m := %v\n  %v"
			if err := g.tmpl("write_map.tmpl").Execute(&buf, typ); err != nil {
				return "", err
			}
			return fmt.Sprintf(tmpl, variable, buf.String()), nil
		case "list":
			tmpl := "lst := %v\n  %v"
			if err := g.tmpl("write_list.tmpl").Execute(&buf, typ); err != nil {
				return "", err
			}
			return fmt.Sprintf(tmpl, variable, buf.String()), nil
		case "set":
			tmpl := "m := %v\n  %v"
			if err := g.tmpl("write_set.tmpl").Execute(&buf, typ); err != nil {
				return "", nil
			}
			return fmt.Sprintf(tmpl, variable, buf.String()), nil
		}
		return "", fmt.Errorf("unsupported type: %v", typ.Name)
	case parser.TypeIdentifier:
		if g.isPtrType(typ) {
			tmpl := "if err = %v.Write(w); err != nil { return err }"
			return fmt.Sprintf(tmpl, variable), nil
		}
		if finalType := typ.GetFinalType(); finalType != nil {
			if typ, ok := finalType.(*parser.Type); ok {
				return g.formatWrite(typ, variable)
			}
			if _, ok := finalType.(*parser.Enum); ok {
				tmpl := "if err = w.WriteI32(int32(%v)); err != nil { return err }"
				return fmt.Sprintf(tmpl, variable), nil
			}
			panic(" not implemented")
		}
		return "", fmt.Errorf("unknown non-ptr type: %v", typ.Name)
	default:
		return "", fmt.Errorf("unknown type category: %v", typ.Category)
	}

	_ = typeName
	return "// TODO", nil
}
