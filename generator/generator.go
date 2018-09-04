package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/gobuffalo/packr"
	"github.com/jxskiss/gothrifter/parser"
)

var (
	_      = format.Source
	logger = log.New(os.Stderr, "[THRIFTERC] ", log.LstdFlags)
	//pprint  = spew.ConfigState{DisableMethods: true, Indent: "  "}
	tmplBox    = packr.NewBox("./templates")
	mlineRegex = regexp.MustCompile(`([\r\n]+\s*){2,}([\w\}\]\)/]+)`)
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
	// clear the massive blank lines
	code = mlineRegex.ReplaceAllFunc(code, func(old []byte) (new []byte) {
		word := old[bytes.LastIndexAny(old, "\r\n\t ")+1:]
		var sep = make([]byte, 0, 16)
		if old[0] == '\r' && old[1] == '\n' {
			sep = append(sep, old[:2]...)
		} else {
			sep = append(sep, old[0])
		}
		if w := string(word); w == "package" || w == "func" || w == "type" {
			sep = append(sep, sep...)
		}
		new = append(sep, word...)
		return
	})

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
		"isPtrType":       g.isPtrType,
		"formatValue":     g.formatValue,
		"formatType":      g.formatType,
		"formatStructTag": g.formatStructTag,
		"formatReturn":    g.formatReturn,
		"formatArguments": g.formatArguments,
		"formatRead":      g.formatRead,
		"formatWrite":     g.formatWrite,
		"reqChecker":      g.reqChecker,
		"toCamelCase":     ToCamelCase,
		"toSnakeCase":     ToSnakeCase,
		"TODO":            func() string { return "TODO" },
		"VERSION":         func() string { return Version },
	}
}

func (g *Generator) isPtrField(field *parser.Field) bool {
	if field.Type.Category == parser.TypeContainer {
		return false
	}
	if field.Optional && field.Default == nil {
		return true
	}
	return g.isPtrType(field.Type)
}

func (g *Generator) isPtrType(typ *parser.Type) bool {
	if typ.Category == parser.TypeIdentifier {
		finalType := typ.GetFinalType()
		if typ, ok := finalType.(*parser.Type); ok {
			if typ.Category != parser.TypeIdentifier {
				return false
			}
		}
		if _, ok := finalType.(*parser.Enum); ok {
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
			if g.isPtrType(typ.ValueType) {
				vt = "*" + vt
			}
			return fmt.Sprintf("map[%v]%v", kt, vt), nil
		case "list":
			if vt, err = g.formatType(typ.ValueType); err != nil {
				return "", err
			}
			if g.isPtrType(typ.ValueType) {
				vt = "*" + vt
			}
			return fmt.Sprintf("[]%v", vt), nil
		}
	}

	return typ.Name, nil
}

func (g *Generator) formatStructTag(field *parser.Field) (str string, err error) {
	defer func(err *error) {
		if e := recover(); e != nil {
			*err = e.(error)
		}
	}(&err)
	must := func(n int, err error) {
		if err != nil {
			panic(err)
		}
	}

	var buf bytes.Buffer
	must(buf.WriteString(`thrift:"`))
	must(buf.WriteString(field.Name))
	must(buf.WriteString(","))
	must(buf.WriteString(strconv.Itoa(field.ID)))
	if field.Requiredness != "" && field.Requiredness != "default" {
		must(buf.WriteString(","))
		must(buf.WriteString(field.Requiredness))
	} else if t := field.Type.TType(); t == parser.MAP || t == parser.SET {
		must(buf.WriteString(","))
	}
	if field.Type.TType() == parser.MAP {
		must(buf.WriteString(",map"))
	} else if field.Type.TType() == parser.SET {
		must(buf.WriteString(",set"))
	}
	must(buf.WriteString(`" json:"`))
	must(buf.WriteString(ToSnakeCase(field.Name)))
	if field.Optional {
		must(buf.WriteString(",omitempty"))
	}
	must(buf.WriteRune('"'))
	return buf.String(), nil
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
	argStructs, err := g.parseArguments(svc)
	if err != nil {
		return "", err
	}
	pkg := &Package{
		Document: &parser.Document{
			RefName:  svc.D.RefName,
			Includes: svc.D.Includes,
			Structs:  argStructs,
		},
	}
	var buf bytes.Buffer
	if err := g.tmpl("structs.tmpl").Execute(&buf, pkg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (g *Generator) parseArguments(svc *parser.Service) ([]*parser.Struct, error) {
	argStructs := make([]*parser.Struct, 0)
	for _, method := range svc.Methods {
		// arguments
		s := &parser.Struct{
			Name:   svc.Name + ToCamelCase(method.Name) + "Args",
			Fields: make([]*parser.Field, 0, len(method.Arguments)),
		}
		for _, f := range method.Arguments {
			if g.isPtrType(f.Type) {
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
				Requiredness: parser.ReqOptional,
			}
			if !g.isPtrType(r.Type) {
				r.Optional = false
				r.Requiredness = parser.ReqRequired
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
	return argStructs, nil
}

func (g *Generator) reqChecker(s *parser.Struct) *ReqChecker {
	requiredFields := make([]*parser.Field, 0)
	for _, f := range s.Fields {
		if f.Requiredness == parser.ReqRequired {
			requiredFields = append(requiredFields, f)
		}
	}
	return NewReqChecker(requiredFields)
}

type checkField struct {
	f    *parser.Field
	id   int
	mask uint64
}

type ReqChecker struct {
	fields map[int]checkField
	mask   []uint64
}

func NewReqChecker(requiredFields []*parser.Field) *ReqChecker {
	rc := &ReqChecker{fields: make(map[int]checkField)}
	preMask := uint64(0)
	for i, f := range requiredFields {
		var mask uint64
		if i%64 == 0 {
			rc.mask = append(rc.mask, 0)
			preMask, mask = 1, 1
		} else {
			mask = preMask << 1
			preMask = mask
		}
		rc.fields[f.ID] = checkField{f: f, id: i, mask: mask}
		rc.mask[i/64] |= mask
	}
	return rc
}

func (rc *ReqChecker) Init() string {
	if len(rc.fields) == 0 {
		return ""
	}
	tmpl := "var issetmask [%v]uint64"
	return fmt.Sprintf(tmpl, len(rc.fields)/64+1)
}

func (rc *ReqChecker) Set(fieldId int) string {
	req, ok := rc.fields[fieldId]
	if !ok {
		return ""
	}
	tmpl := "issetmask[%v] |= 0x%x"
	return fmt.Sprintf(tmpl, req.id/64, req.mask)
}

func (rc *ReqChecker) Check() string {
	if len(rc.fields) == 0 {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("if issetmask != [%v]uint64{", len(rc.mask)))
	for _, x := range rc.mask {
		buf.WriteString(fmt.Sprintf("0x%x, ", x))
	}
	buf.WriteString("} {\n")
	stmts := make([]string, len(rc.fields))
	for _, r := range rc.fields {
		tmpl := `if issetmask[%v] & 0x%x == 0 { return thrift.NewApplicationException(thrift.INVALID_DATA, "required field %v is not set") }`
		stmts[r.id] = fmt.Sprintf(tmpl, r.id/64, r.mask, ToCamelCase(r.f.Name))
	}
	for _, s := range stmts {
		buf.WriteString(s)
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	return buf.String()
}
