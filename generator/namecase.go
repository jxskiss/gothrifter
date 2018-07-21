package generator

import (
	"bytes"
	"strings"
	"unicode"
)

type strCase bool

const (
	lower strCase = false
	upper strCase = true
)

// Copied from gorm/golint
var commonInitialisms = []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "TTL", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
var commonInitialismsReplacer *strings.Replacer

func init() {
	var commonInitialismsForReplacer []string
	for _, initialism := range commonInitialisms {
		commonInitialismsForReplacer = append(commonInitialismsForReplacer, initialism, strings.Title(strings.ToLower(initialism)))
	}
	commonInitialismsReplacer = strings.NewReplacer(commonInitialismsForReplacer...)
}

// ToSnakeCase convert string to snake case.
func ToSnakeCase(name string) string {
	if name == "" {
		return ""
	}

	var value = commonInitialismsReplacer.Replace(name)
	var buf = bytes.NewBufferString("")
	var lastCase, currCase, nextCase strCase

	for i, v := range value[:len(value)-1] {
		nextCase = strCase(value[i+1] >= 'A' && value[i+1] <= 'Z')
		if i > 0 {
			if currCase == upper {
				if lastCase == upper && nextCase == upper {
					buf.WriteRune(v)
				} else {
					if value[i-1] != '_' && value[i+1] != '_' {
						buf.WriteRune('_')
					}
					buf.WriteRune(v)
				}
			} else {
				buf.WriteRune(v)
				if i == len(value)-2 && nextCase == upper {
					buf.WriteRune('_')
				}
			}
		} else {
			currCase = upper
			buf.WriteRune(v)
		}
		lastCase = currCase
		currCase = nextCase
	}

	buf.WriteByte(value[len(value)-1])

	s := strings.ToLower(buf.String())
	return s
}

// ToCamelCase convert string to camel case.
func ToCamelCase(name string) string {
	if strings.ToUpper(name) == name {
		return name
	}
	sep := true
	return strings.Map(func(r rune) rune {
		if r == '_' {
			sep = true
			return -1
		}
		if !sep {
			return r
		}
		sep = false
		return unicode.ToUpper(r)
	}, name)
}
