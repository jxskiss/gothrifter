package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"

	//"github.com/davecgh/go-spew/spew"
	"github.com/jxskiss/gothrifter/parser"
)

var (
	_      = format.Source
	logger = log.New(os.Stderr, "[THRIFTERC] ", log.LstdFlags)
	//pprint = spew.ConfigState{DisableMethods: true, Indent: "  "}
)

var GitRevision = "????"
var Version = "0.1.0-" + GitRevision

type Generator struct {
	Filename string
	Prefix   string
	Output   string

	GenAll       bool
	DebugMode    bool
	RootPkg      *Package
	ImportedPkgs map[string]*Package // key: absolute path to thrift file

	tmplCache sync.Map
}

func New(filename, prefix, output string) *Generator {
	gen := &Generator{
		Filename:     filename,
		Prefix:       prefix,
		Output:       output,
		ImportedPkgs: make(map[string]*Package),
	}
	return gen
}

func (g *Generator) Parse() error {
	logger.Println("parsing:", g.Filename)

	doc, err := parser.Parse(g.Filename)
	if err != nil {
		return err
	}
	g.RootPkg = &Package{Document: doc, G: g}

	if err = g.parseIncludes(); err != nil {
		return err
	}
	if err = g.resolveTypes(); err != nil {
		return err
	}

	return nil
}

func (g *Generator) parseIncludes() error {
	if g.RootPkg == nil {
		return nil
	}
	includes := make([]string, 0)
	for _, inc := range g.RootPkg.Document.Includes {
		includes = append(includes, inc.AbsPath)
	}
	for len(includes) > 0 {
		fn := includes[0]
		includes = includes[1:]

		logger.Println("parsing (include):", fn)
		if g.ImportedPkgs[fn] != nil {
			continue
		}
		doc, err := parser.Parse(fn)
		if err != nil {
			return err
		}
		g.ImportedPkgs[fn] = &Package{Document: doc, G: g}
		for _, inc := range doc.Includes {
			if g.ImportedPkgs[inc.AbsPath] == nil {
				includes = append(includes, fn)
			}
		}
	}
	return nil
}

func (g *Generator) resolveTypes() error {
	pkgs := []*Package{g.RootPkg}
	for _, pkg := range g.ImportedPkgs {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range pkgs {
		for _, typ := range pkg.IdentTypes {
			if typ.Category != parser.TypeIdentifier || typ.FinalType != nil {
				continue
			}
			typ.FinalType = g.resolveIdentifierType(pkg, typ)
		}
	}
	return nil
}

// TODO: what about recursive definitions?
func (g *Generator) resolveIdentifierType(pkg *Package, typ *parser.Type) interface{} {
	if typ.Category != parser.TypeIdentifier {
		panic("can not resolve non-identifier type: " + typ.Category)
	}
	refNames := strings.SplitN(typ.Name, ".", 2)
	var ref interface{}
	if len(refNames) == 1 {
		ref = pkg.Document.ResolveIdentifierType(refNames[0])
	} else {
		pkg = g.ImportedPkgs[pkg.Document.Includes[refNames[0]].AbsPath]
		ref = pkg.Document.ResolveIdentifierType(refNames[1])
	}
	if typ, ok := ref.(*parser.Type); ok {
		if typ.Category != parser.TypeIdentifier {
			return typ
		}
		return g.resolveIdentifierType(pkg, typ)
	}
	return ref
}

func (g *Generator) Generate() error {
	var err error
	if err = g.RootPkg.Generate(); err != nil {
		return err
	}
	if !g.GenAll {
		return nil
	}
	for _, pkg := range g.ImportedPkgs {
		if err = pkg.Generate(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) formatCode(code []byte) ([]byte, error) {
	formatted, err := format.Source(code)
	if err != nil {
		log.Println("format:", err)
		if g.DebugMode {
			return code, nil
		}
		return nil, err
	}
	return formatted, nil
}

func (g *Generator) tmpl(name string) *template.Template {
	if t, ok := g.tmplCache.Load(name); ok {
		return t.(*template.Template)
	}
	t := template.Must(template.New(name).Funcs(g.funcMap()).Parse(tmplBox.String(name)))
	g.tmplCache.Store(name, t)
	return t
}

func (g *Generator) funcMap() template.FuncMap {
	return template.FuncMap{
		"isPtrField":      g.isPtrField,
		"formatValue":     g.formatValue,
		"formatType":      g.formatType,
		"formatReturn":    g.formatReturn,
		"formatArguments": g.formatArguments,
		"formatRead":      g.formatRead,
		"formatWrite":     g.formatWrite,
		"toCamelCase":     ToCamelCase,
		"toSnakeCase":     ToSnakeCase,
		"TODO":            func() string { return "TODO" },
		"VERSION":         func() string { return Version },
	}
}

func (g *Generator) isPtrField(field *parser.Field) bool {
	if field.Optional && field.Default == nil {
		return true
	}
	return g.isPtrType(field.Type)
}

func (g *Generator) isPtrType(typ *parser.Type) bool {
	if typ.Category == parser.TypeIdentifier {
		if typ, ok := typ.FinalType.(*parser.Type); ok {
			if typ.Category != parser.TypeIdentifier {
				return false
			}
		}
		if _, ok := typ.FinalType.(*parser.Enum); ok {
			return false
		}
		return true
	}
	return false
}

func (g *Generator) formatValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case parser.ConstValue:
		switch v.Type {
		case parser.ConstTypeLiteral:
			return fmt.Sprintf("%q", v.Value), nil
		case parser.ConstTypeIdentifier:
			ss := strings.Split(v.Value, ".")
			if len(ss) <= 2 {
				return v.Value, nil
			}
			return strings.Join(ss[0:len(ss)-1], ".") + "_" + ss[len(ss)-1], nil
		}
		return v.Value, nil
	case parser.ListConstValue:
		values := make([]string, len(v))
		for i, item := range v {
			v, vErr := g.formatValue(item)
			if vErr != nil {
				return "", vErr
			}
			values[i] = v
		}
		return fmt.Sprintf("{%v}", strings.Join(values, ", ")), nil
	case []parser.MapConstValue:
		kvs := make([]string, len(v))
		for i, item := range v {
			k, kErr := g.formatValue(item.Key)
			if kErr != nil {
				return "", kErr
			}
			v, vErr := g.formatValue(item.Value)
			if vErr != nil {
				return "", vErr
			}
			kvs[i] = fmt.Sprintf("%v: %v", k, v)
		}
		return fmt.Sprintf("{%v}", strings.Join(kvs, ", ")), nil
	default:
		return "", fmt.Errorf("undefined const type: %t", v)
	}
}

func (g *Generator) formatType(typ *parser.Type) (string, error) {
	if typ.Category == parser.TypeBasic {
		switch typ.Name {
		case "i8", "byte":
			return "int8", nil
		case "i16":
			return "int16", nil
		case "i32":
			return "int32", nil
		case "i64":
			return "int64", nil
		case "double":
			return "float64", nil
		case "binary":
			return "[]byte", nil
		case "string", "slist":
			return "string", nil
		case "float":
			return "float32", nil
		}
		return typ.Name, nil
	}

	if typ.Category == parser.TypeIdentifier {
		// TODO: handle potential name conflict
		parts := strings.SplitN(typ.Name, ".", 2)
		if len(parts) == 1 {
			return typ.Name, nil
		}
		refName, typName := parts[0], parts[1]
		inc, ok := typ.D.Includes[refName]
		if !ok {
			return typ.Name, nil
		}
		incPkg := g.ImportedPkgs[inc.AbsPath]
		if incPkg == nil {
			return typ.Name, nil
		}
		return incPkg.Name() + "." + typName, nil
	}

	if typ.Category == parser.TypeContainer {
		var kt, vt string
		var err error
		switch typ.Name {
		case "set":
			if kt, err = g.formatType(typ.ValueType); err != nil {
				return "", err
			}
			return fmt.Sprintf("map[%v]bool", kt), nil
		case "map":
			if kt, err = g.formatType(typ.KeyType); err != nil {
				return "", err
			}
			if vt, err = g.formatType(typ.ValueType); err != nil {
				return "", err
			}
			if typ.ValueType.Category == parser.TypeIdentifier {
				vt = "*" + vt
			}
			return fmt.Sprintf("map[%v]%v", kt, vt), nil
		case "list":
			if vt, err = g.formatType(typ.ValueType); err != nil {
				return "", err
			}
			if typ.ValueType.Category == parser.TypeIdentifier {
				vt = "*" + vt
			}
			return fmt.Sprintf("[]%v", vt), nil
		}
	}

	return typ.Name, nil
}

func (g *Generator) formatReturn(typ *parser.Type) (string, error) {
	switch typ.Category {
	case parser.TypeBasic, parser.TypeContainer:
		return g.formatType(typ)
	}

	ret, err := g.formatType(typ)
	if err != nil {
		return "", err
	}
	return "*" + ret, nil
}

func (g *Generator) formatArguments(svc *parser.Service) (string, error) {
	argStructs := make([]*parser.Struct, 0)
	for _, method := range svc.Methods {
		// arguments
		s := &parser.Struct{
			Name:   svc.Name + ToCamelCase(method.Name) + "Args",
			Fields: make([]*parser.Field, 0, len(method.Arguments)),
		}
		for _, f := range method.Arguments {
			if f.Type.Category == parser.TypeIdentifier {
				f.Optional = true
				f.Requiredness = "optional"
			}
			s.Fields = append(s.Fields, f)
		}
		argStructs = append(argStructs, s)

		// return and throws
		if method.Oneway {
			continue
		}
		s = &parser.Struct{
			Name:   svc.Name + ToCamelCase(method.Name) + "Result",
			Fields: make([]*parser.Field, 0),
		}
		if method.ReturnType.Name != "void" {
			r := &parser.Field{
				ID:           0,
				Name:         "success",
				Type:         method.ReturnType,
				Optional:     true,
				Requiredness: parser.RequirednessOptional,
			}
			if r.Type.Category != parser.TypeIdentifier {
				r.Optional = false
				r.Requiredness = parser.RequirednessRequired
			}
			s.Fields = append(s.Fields, r)
		}
		for _, f := range method.Exceptions {
			// any exception should be considered optional
			f.Optional = true
			s.Fields = append(s.Fields, f)
		}
		argStructs = append(argStructs, s)
	}
	pkg := &Package{
		Document: &parser.Document{
			RefName:  svc.D.RefName,
			Includes: svc.D.Includes,
			Structs:  argStructs,
		},
	}
	tmpl := template.Must(template.New("structs").Funcs(g.funcMap()).Parse(structsTmpl))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pkg); err != nil {
		return "", err
	}
	return buf.String(), nil
}
