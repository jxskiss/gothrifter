package generator

import (
	"bytes"
	"github.com/jxskiss/thriftkit/parser"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

	var code []byte
	if code, err = p.gencode(); err != nil {
		return err
	}
	if code, err = p.G.formatCode(code); err != nil {
		return err
	}
	if err = ioutil.WriteFile(outfile, code, 0644); err != nil {
		return err
	}

	// struct decoders
	decoderFile, err := filepath.Abs(filepath.Join(outDir, "decoder.go"))
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	if err = p.G.tmpl("header.tmpl").Execute(&buf, p); err != nil {
		log.Println("decoder:", err)
		return err
	}
	for _, x := range p.Structs {
		if err = p.G.tmpl("read_struct.tmpl").Execute(&buf, x); err != nil {
			log.Println("decoder:", err)
			return err
		}
	}
	for _, x := range p.Unions {
		if err = p.G.tmpl("read_struct.tmpl").Execute(&buf, (*parser.Struct)(x)); err != nil {
			log.Println("decoder:", err)
			return err
		}
	}
	for _, svc := range p.Services {
		argStructs, err := p.G.parseArguments(svc)
		if err != nil {
			return err
		}
		for _, x := range argStructs {
			if err = p.G.tmpl("read_struct.tmpl").Execute(&buf, x); err != nil {
				log.Println("decoder:", err)
				return err
			}
		}
	}
	if code, err = p.G.formatCode(buf.Bytes()); err != nil {
		return err
	}
	if err = ioutil.WriteFile(decoderFile, code, 0644); err != nil {
		return err
	}

	// struct encoders
	encoderFile, err := filepath.Abs(filepath.Join(outDir, "encoder.go"))
	if err != nil {
		return err
	}
	buf.Reset()
	if err = p.G.tmpl("header.tmpl").Execute(&buf, p); err != nil {
		log.Println("encoder:", err)
		return err
	}
	for _, x := range p.Structs {
		if err = p.G.tmpl("write_struct.tmpl").Execute(&buf, x); err != nil {
			log.Println("encoder:", err)
			return err
		}
	}
	for _, x := range p.Unions {
		if err = p.G.tmpl("write_union.tmpl").Execute(&buf, x); err != nil {
			log.Println("encoder:", err)
			return err
		}
	}
	for _, svc := range p.Services {
		argStructs, err := p.G.parseArguments(svc)
		if err != nil {
			return err
		}
		for _, x := range argStructs {
			if err = p.G.tmpl("write_struct.tmpl").Execute(&buf, x); err != nil {
				log.Println("encoder:", err)
				return err
			}
		}
	}
	if code, err = p.G.formatCode(buf.Bytes()); err != nil {
		return err
	}
	if err = ioutil.WriteFile(encoderFile, code, 0644); err != nil {
		return err
	}

	if len(p.Services) > 0 {
		// go-kit server
		kitserverFile, err := filepath.Abs(filepath.Join(outDir, "kitserver.go"))
		if err != nil {
			return err
		}
		buf.Reset()
		if err = p.G.tmpl("kitserver.tmpl").Execute(&buf, p); err != nil {
			log.Println("kitserver:", err)
			return err
		}
		if code, err = p.G.formatCode(buf.Bytes()); err != nil {
			return err
		}
		if err = ioutil.WriteFile(kitserverFile, code, 0644); err != nil {
			return err
		}

		// go-kit client
		kitclientFile, err := filepath.Abs(filepath.Join(outDir, "kitclient.go"))
		if err != nil {
			return err
		}
		buf.Reset()
		if err = p.G.tmpl("kitclient.tmpl").Execute(&buf, p); err != nil {
			log.Println("kitclient:", err)
			return err
		}
		if code, err = p.G.formatCode(buf.Bytes()); err != nil {
			return err
		}
		if err = ioutil.WriteFile(kitclientFile, code, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (p *Package) gencode() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	if err = p.G.tmpl("header.tmpl").Execute(&buf, p); err != nil {
		return nil, err
	}
	if _, err = buf.WriteString("var GoUnusedProtection__ int"); err != nil {
		return nil, err
	}

	if err = p.G.tmpl("consts.tmpl").Execute(&buf, p); err != nil {
		return nil, err
	}

	if err = p.G.tmpl("typedefs.tmpl").Execute(&buf, p); err != nil {
		return nil, err
	}

	// Structs, Exceptions
	p.Structs = append(p.Structs, p.Exceptions...)

	// Unions
	// Field of union should all be optional.
	for _, un := range p.Unions {
		for _, f := range un.Fields {
			if f.Requiredness != parser.ReqDefault {
				log.Printf("union %v field %v: union members must be optional, ignoring specified requiredness\n", un.Name, f.Name)
				f.Requiredness = parser.ReqDefault
			}
			f.Optional = true
		}
	}

	if err = p.G.tmpl("structs.tmpl").Execute(&buf, p); err != nil {
		return nil, err
	}

	if err = p.G.tmpl("unions.tmpl").Execute(&buf, p); err != nil {
		return nil, err
	}

	if err = p.G.tmpl("exceptions.tmpl").Execute(&buf, p); err != nil {
		return nil, err
	}

	if err = p.G.tmpl("service.tmpl").Execute(&buf, p); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
