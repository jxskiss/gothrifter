package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	//"github.com/davecgh/go-spew/spew"
	"github.com/jxskiss/gothrifter/parser"
)

var (
	_      = format.Source
	logger = log.New(os.Stderr, "[THRIFTERC] ", log.LstdFlags)
	//pprint = spew.ConfigState{DisableMethods: true, Indent: "  "}
)

var Version = "0.1.0"

type Package struct {
	*parser.Document
	G *Generator
}

func (p *Package) Name() string {
	name := p.fullname()
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func (p *Package) ImportPath() string {
	name := p.fullname()
	path := filepath.Join(p.G.Prefix, filepath.Join(strings.Split(name, ".")...))
	return strings.Replace(path, string(os.PathSeparator), "/", -1)
}

func (p *Package) fullname() string {
	if n := p.Document.Namespaces["go"]; n != nil {
		return n.Name
	}
	if n := p.Document.Namespaces["*"]; n != nil {
		return n.Name
	}
	return p.Document.RefName
}

type include struct {
	Name       string
	ImportPath string
}

func (p *Package) Includes() []include {
	var r []include
	for _, inc := range p.Document.Includes {
		r = append(r, include{
			Name:       p.G.ImportedPkgs[inc.AbsPath].Name(),
			ImportPath: p.G.ImportedPkgs[inc.AbsPath].ImportPath(),
		})
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i].Name < r[j].Name
	})
	return r
}

func (p *Package) Generate() error {
	outDir := filepath.Join(p.G.Output, filepath.Join(strings.Split(p.fullname(), ".")...))
	if err := os.MkdirAll(outDir, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	outfile, err := filepath.Abs(filepath.Join(outDir, p.RefName+".thrift.go"))
	if err != nil {
		return err
	}

	code, err := p.gencode()
	if err != nil {
		return err
	}
	formattedCode, err := format.Source(code)
	if err != nil {
		log.Println("format:", err)
		if !p.G.DebugMode {
			return err
		}
	} else {
		code = formattedCode
	}
	if err = ioutil.WriteFile(outfile, code, 0644); err != nil {
		return err
	}

	return nil
}

func (p *Package) gencode() ([]byte, error) {
	var buf bytes.Buffer
	var err error
	var funcMap = p.G.funcMap()
	var tmpl *template.Template

	tmpl = template.Must(template.New("header").Funcs(funcMap).Parse(headerTmpl))
	if err = tmpl.Execute(&buf, p); err != nil {
		return nil, err
	}

	tmpl = template.Must(template.New("consts").Funcs(funcMap).Parse(constsTmpl))
	if err = tmpl.Execute(&buf, p); err != nil {
		return nil, err
	}

	tmpl = template.Must(template.New("typedefs").Funcs(funcMap).Parse(typedefsTmpl))
	if err = tmpl.Execute(&buf, p); err != nil {
		return nil, err
	}

	// Structs, Exceptions, Unions
	p.Structs = append(p.Structs, p.Exceptions...)
	p.Structs = append(p.Structs, p.Unions...)
	tmpl = template.Must(template.New("structs").Funcs(funcMap).Parse(structsTmpl))
	if err = tmpl.Execute(&buf, p); err != nil {
		return nil, err
	}

	tmpl = template.Must(template.New("exceptions").Funcs(funcMap).Parse(exceptionsTmpl))
	if err = tmpl.Execute(&buf, p); err != nil {
		return nil, err
	}

	tmpl = template.Must(template.New("services").Funcs(funcMap).Parse(servicesTmpl))
	if err = tmpl.Execute(&buf, p); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type Generator struct {
	Filename string
	Prefix   string
	Output   string

	GenAll       bool
	DebugMode    bool
	RootPkg      *Package
	ImportedPkgs map[string]*Package // key: absolute path to thrift file
}

func New(filename, prefix, output string) *Generator {
	return &Generator{
		Filename:     filename,
		Prefix:       prefix,
		Output:       output,
		ImportedPkgs: make(map[string]*Package),
	}
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

func (g *Generator) funcMap() template.FuncMap {
	return template.FuncMap{
		"formatValue":     g.formatValue,
		"formatType":      g.formatType,
		"formatReturn":    g.formatReturn,
		"formatArguments": g.formatArguments,
		"toCamelCase":     ToCamelCase,
		"toSnakeCase":     ToSnakeCase,
	}
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
