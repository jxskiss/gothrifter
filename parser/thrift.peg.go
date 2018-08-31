package parser

//go:generate peg thrift.peg

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleDocument
	ruleHeader
	ruleInclude
	ruleCppInclude
	ruleNamespace
	ruleNamespaceScope
	ruleDefinition
	ruleConst
	ruleTypedef
	ruleEnum
	ruleSenum
	ruleStruct
	ruleUnion
	ruleException
	ruleService
	ruleField
	ruleFieldID
	ruleFieldReq
	ruleXsdFieldOptions
	ruleXsdAttrs
	ruleFunction
	ruleFunctionType
	ruleThrows
	ruleFieldType
	ruleDefinitionType
	ruleBaseType
	ruleContainerType
	ruleMapType
	ruleSetType
	ruleListType
	ruleCppType
	ruleConstValue
	ruleDoubleConstant
	ruleExponent
	ruleIntConstant
	ruleConstList
	ruleConstMap
	ruleLiteral
	ruleIdentifier
	ruleSTIdentifier
	ruleListSeparator
	ruleLetter
	ruleDigit
	ruleIdChars
	ruleSpacing
	ruleWhitespace
	ruleLongComment
	ruleLineComment
	rulePragma
	ruleINCLUDE
	ruleCPP_INCLUDE
	ruleNAMESPACE
	ruleSMALLTALK_CATEGORY
	ruleSMALLTALK_PREFIX
	rulePHP_NAMESPACE
	ruleXSD_NAMESPACE
	ruleCONST
	ruleTYPEDEF
	ruleENUM
	ruleSENUM
	ruleSTRUCT
	ruleUNION
	ruleSERVICE
	ruleEXTENDS
	ruleEXCEPTION
	ruleONEWAY
	ruleTHROWS
	ruleCPP_TYPE
	ruleXSD_ALL
	ruleXSD_OPTIONAL
	ruleXSD_NILLABLE
	ruleXSD_ATTRS
	ruleVOID
	ruleMAP
	ruleSET
	ruleLIST
	ruleBOOL
	ruleBYTE
	ruleI8
	ruleI16
	ruleI32
	ruleI64
	ruleDOUBLE
	ruleSTRING
	ruleBINARY
	ruleSLIST
	ruleFLOAT
	ruleLBRK
	ruleRBRK
	ruleLPAR
	ruleRPAR
	ruleLWING
	ruleRWING
	ruleLPOINT
	ruleRPOINT
	ruleEQUAL
	ruleCOMMA
	ruleCOLON
	ruleEOT
	rulePegText
)

var rul3s = [...]string{
	"Unknown",
	"Document",
	"Header",
	"Include",
	"CppInclude",
	"Namespace",
	"NamespaceScope",
	"Definition",
	"Const",
	"Typedef",
	"Enum",
	"Senum",
	"Struct",
	"Union",
	"Exception",
	"Service",
	"Field",
	"FieldID",
	"FieldReq",
	"XsdFieldOptions",
	"XsdAttrs",
	"Function",
	"FunctionType",
	"Throws",
	"FieldType",
	"DefinitionType",
	"BaseType",
	"ContainerType",
	"MapType",
	"SetType",
	"ListType",
	"CppType",
	"ConstValue",
	"DoubleConstant",
	"Exponent",
	"IntConstant",
	"ConstList",
	"ConstMap",
	"Literal",
	"Identifier",
	"STIdentifier",
	"ListSeparator",
	"Letter",
	"Digit",
	"IdChars",
	"Spacing",
	"Whitespace",
	"LongComment",
	"LineComment",
	"Pragma",
	"INCLUDE",
	"CPP_INCLUDE",
	"NAMESPACE",
	"SMALLTALK_CATEGORY",
	"SMALLTALK_PREFIX",
	"PHP_NAMESPACE",
	"XSD_NAMESPACE",
	"CONST",
	"TYPEDEF",
	"ENUM",
	"SENUM",
	"STRUCT",
	"UNION",
	"SERVICE",
	"EXTENDS",
	"EXCEPTION",
	"ONEWAY",
	"THROWS",
	"CPP_TYPE",
	"XSD_ALL",
	"XSD_OPTIONAL",
	"XSD_NILLABLE",
	"XSD_ATTRS",
	"VOID",
	"MAP",
	"SET",
	"LIST",
	"BOOL",
	"BYTE",
	"I8",
	"I16",
	"I32",
	"I64",
	"DOUBLE",
	"STRING",
	"BINARY",
	"SLIST",
	"FLOAT",
	"LBRK",
	"RBRK",
	"LPAR",
	"RPAR",
	"LWING",
	"RWING",
	"LPOINT",
	"RPOINT",
	"EQUAL",
	"COMMA",
	"COLON",
	"EOT",
	"PegText",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type Thrift struct {
	D *Document

	Buffer string
	buffer []rune
	rules  [101]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *Thrift) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *Thrift) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *Thrift
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *Thrift) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *Thrift) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Document <- <(Spacing Header* Definition* EOT)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleSpacing]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3 := position, tokenIndex
					if !_rules[ruleHeader]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex = position3, tokenIndex3
				}
			l4:
				{
					position5, tokenIndex5 := position, tokenIndex
					if !_rules[ruleDefinition]() {
						goto l5
					}
					goto l4
				l5:
					position, tokenIndex = position5, tokenIndex5
				}
				if !_rules[ruleEOT]() {
					goto l0
				}
				add(ruleDocument, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Header <- <(Include / CppInclude / Namespace)> */
		func() bool {
			position6, tokenIndex6 := position, tokenIndex
			{
				position7 := position
				{
					position8, tokenIndex8 := position, tokenIndex
					if !_rules[ruleInclude]() {
						goto l9
					}
					goto l8
				l9:
					position, tokenIndex = position8, tokenIndex8
					if !_rules[ruleCppInclude]() {
						goto l10
					}
					goto l8
				l10:
					position, tokenIndex = position8, tokenIndex8
					if !_rules[ruleNamespace]() {
						goto l6
					}
				}
			l8:
				add(ruleHeader, position7)
			}
			return true
		l6:
			position, tokenIndex = position6, tokenIndex6
			return false
		},
		/* 2 Include <- <(INCLUDE Literal)> */
		func() bool {
			position11, tokenIndex11 := position, tokenIndex
			{
				position12 := position
				if !_rules[ruleINCLUDE]() {
					goto l11
				}
				if !_rules[ruleLiteral]() {
					goto l11
				}
				add(ruleInclude, position12)
			}
			return true
		l11:
			position, tokenIndex = position11, tokenIndex11
			return false
		},
		/* 3 CppInclude <- <(CPP_INCLUDE Literal)> */
		func() bool {
			position13, tokenIndex13 := position, tokenIndex
			{
				position14 := position
				if !_rules[ruleCPP_INCLUDE]() {
					goto l13
				}
				if !_rules[ruleLiteral]() {
					goto l13
				}
				add(ruleCppInclude, position14)
			}
			return true
		l13:
			position, tokenIndex = position13, tokenIndex13
			return false
		},
		/* 4 Namespace <- <((NAMESPACE (NamespaceScope Identifier)) / (SMALLTALK_CATEGORY STIdentifier) / (SMALLTALK_PREFIX Identifier) / (PHP_NAMESPACE Literal) / (XSD_NAMESPACE Literal))> */
		func() bool {
			position15, tokenIndex15 := position, tokenIndex
			{
				position16 := position
				{
					position17, tokenIndex17 := position, tokenIndex
					if !_rules[ruleNAMESPACE]() {
						goto l18
					}
					if !_rules[ruleNamespaceScope]() {
						goto l18
					}
					if !_rules[ruleIdentifier]() {
						goto l18
					}
					goto l17
				l18:
					position, tokenIndex = position17, tokenIndex17
					if !_rules[ruleSMALLTALK_CATEGORY]() {
						goto l19
					}
					if !_rules[ruleSTIdentifier]() {
						goto l19
					}
					goto l17
				l19:
					position, tokenIndex = position17, tokenIndex17
					if !_rules[ruleSMALLTALK_PREFIX]() {
						goto l20
					}
					if !_rules[ruleIdentifier]() {
						goto l20
					}
					goto l17
				l20:
					position, tokenIndex = position17, tokenIndex17
					if !_rules[rulePHP_NAMESPACE]() {
						goto l21
					}
					if !_rules[ruleLiteral]() {
						goto l21
					}
					goto l17
				l21:
					position, tokenIndex = position17, tokenIndex17
					if !_rules[ruleXSD_NAMESPACE]() {
						goto l15
					}
					if !_rules[ruleLiteral]() {
						goto l15
					}
				}
			l17:
				add(ruleNamespace, position16)
			}
			return true
		l15:
			position, tokenIndex = position15, tokenIndex15
			return false
		},
		/* 5 NamespaceScope <- <((<'*'> Spacing) / Identifier)> */
		func() bool {
			position22, tokenIndex22 := position, tokenIndex
			{
				position23 := position
				{
					position24, tokenIndex24 := position, tokenIndex
					{
						position26 := position
						if buffer[position] != rune('*') {
							goto l25
						}
						position++
						add(rulePegText, position26)
					}
					if !_rules[ruleSpacing]() {
						goto l25
					}
					goto l24
				l25:
					position, tokenIndex = position24, tokenIndex24
					if !_rules[ruleIdentifier]() {
						goto l22
					}
				}
			l24:
				add(ruleNamespaceScope, position23)
			}
			return true
		l22:
			position, tokenIndex = position22, tokenIndex22
			return false
		},
		/* 6 Definition <- <(Const / Typedef / Enum / Senum / Struct / Union / Service / Exception)> */
		func() bool {
			position27, tokenIndex27 := position, tokenIndex
			{
				position28 := position
				{
					position29, tokenIndex29 := position, tokenIndex
					if !_rules[ruleConst]() {
						goto l30
					}
					goto l29
				l30:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleTypedef]() {
						goto l31
					}
					goto l29
				l31:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleEnum]() {
						goto l32
					}
					goto l29
				l32:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleSenum]() {
						goto l33
					}
					goto l29
				l33:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleStruct]() {
						goto l34
					}
					goto l29
				l34:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleUnion]() {
						goto l35
					}
					goto l29
				l35:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleService]() {
						goto l36
					}
					goto l29
				l36:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleException]() {
						goto l27
					}
				}
			l29:
				add(ruleDefinition, position28)
			}
			return true
		l27:
			position, tokenIndex = position27, tokenIndex27
			return false
		},
		/* 7 Const <- <(CONST FieldType Identifier EQUAL ConstValue ListSeparator?)> */
		func() bool {
			position37, tokenIndex37 := position, tokenIndex
			{
				position38 := position
				if !_rules[ruleCONST]() {
					goto l37
				}
				if !_rules[ruleFieldType]() {
					goto l37
				}
				if !_rules[ruleIdentifier]() {
					goto l37
				}
				if !_rules[ruleEQUAL]() {
					goto l37
				}
				if !_rules[ruleConstValue]() {
					goto l37
				}
				{
					position39, tokenIndex39 := position, tokenIndex
					if !_rules[ruleListSeparator]() {
						goto l39
					}
					goto l40
				l39:
					position, tokenIndex = position39, tokenIndex39
				}
			l40:
				add(ruleConst, position38)
			}
			return true
		l37:
			position, tokenIndex = position37, tokenIndex37
			return false
		},
		/* 8 Typedef <- <(TYPEDEF DefinitionType Identifier)> */
		func() bool {
			position41, tokenIndex41 := position, tokenIndex
			{
				position42 := position
				if !_rules[ruleTYPEDEF]() {
					goto l41
				}
				if !_rules[ruleDefinitionType]() {
					goto l41
				}
				if !_rules[ruleIdentifier]() {
					goto l41
				}
				add(ruleTypedef, position42)
			}
			return true
		l41:
			position, tokenIndex = position41, tokenIndex41
			return false
		},
		/* 9 Enum <- <(ENUM Identifier LWING (Identifier (EQUAL IntConstant)? ListSeparator?)* RWING)> */
		func() bool {
			position43, tokenIndex43 := position, tokenIndex
			{
				position44 := position
				if !_rules[ruleENUM]() {
					goto l43
				}
				if !_rules[ruleIdentifier]() {
					goto l43
				}
				if !_rules[ruleLWING]() {
					goto l43
				}
			l45:
				{
					position46, tokenIndex46 := position, tokenIndex
					if !_rules[ruleIdentifier]() {
						goto l46
					}
					{
						position47, tokenIndex47 := position, tokenIndex
						if !_rules[ruleEQUAL]() {
							goto l47
						}
						if !_rules[ruleIntConstant]() {
							goto l47
						}
						goto l48
					l47:
						position, tokenIndex = position47, tokenIndex47
					}
				l48:
					{
						position49, tokenIndex49 := position, tokenIndex
						if !_rules[ruleListSeparator]() {
							goto l49
						}
						goto l50
					l49:
						position, tokenIndex = position49, tokenIndex49
					}
				l50:
					goto l45
				l46:
					position, tokenIndex = position46, tokenIndex46
				}
				if !_rules[ruleRWING]() {
					goto l43
				}
				add(ruleEnum, position44)
			}
			return true
		l43:
			position, tokenIndex = position43, tokenIndex43
			return false
		},
		/* 10 Senum <- <(SENUM Identifier LWING (Literal ListSeparator?)* RWING)> */
		func() bool {
			position51, tokenIndex51 := position, tokenIndex
			{
				position52 := position
				if !_rules[ruleSENUM]() {
					goto l51
				}
				if !_rules[ruleIdentifier]() {
					goto l51
				}
				if !_rules[ruleLWING]() {
					goto l51
				}
			l53:
				{
					position54, tokenIndex54 := position, tokenIndex
					if !_rules[ruleLiteral]() {
						goto l54
					}
					{
						position55, tokenIndex55 := position, tokenIndex
						if !_rules[ruleListSeparator]() {
							goto l55
						}
						goto l56
					l55:
						position, tokenIndex = position55, tokenIndex55
					}
				l56:
					goto l53
				l54:
					position, tokenIndex = position54, tokenIndex54
				}
				if !_rules[ruleRWING]() {
					goto l51
				}
				add(ruleSenum, position52)
			}
			return true
		l51:
			position, tokenIndex = position51, tokenIndex51
			return false
		},
		/* 11 Struct <- <(STRUCT Identifier XSD_ALL? LWING Field* RWING)> */
		func() bool {
			position57, tokenIndex57 := position, tokenIndex
			{
				position58 := position
				if !_rules[ruleSTRUCT]() {
					goto l57
				}
				if !_rules[ruleIdentifier]() {
					goto l57
				}
				{
					position59, tokenIndex59 := position, tokenIndex
					if !_rules[ruleXSD_ALL]() {
						goto l59
					}
					goto l60
				l59:
					position, tokenIndex = position59, tokenIndex59
				}
			l60:
				if !_rules[ruleLWING]() {
					goto l57
				}
			l61:
				{
					position62, tokenIndex62 := position, tokenIndex
					if !_rules[ruleField]() {
						goto l62
					}
					goto l61
				l62:
					position, tokenIndex = position62, tokenIndex62
				}
				if !_rules[ruleRWING]() {
					goto l57
				}
				add(ruleStruct, position58)
			}
			return true
		l57:
			position, tokenIndex = position57, tokenIndex57
			return false
		},
		/* 12 Union <- <(UNION Identifier XSD_ALL? LWING Field* RWING)> */
		func() bool {
			position63, tokenIndex63 := position, tokenIndex
			{
				position64 := position
				if !_rules[ruleUNION]() {
					goto l63
				}
				if !_rules[ruleIdentifier]() {
					goto l63
				}
				{
					position65, tokenIndex65 := position, tokenIndex
					if !_rules[ruleXSD_ALL]() {
						goto l65
					}
					goto l66
				l65:
					position, tokenIndex = position65, tokenIndex65
				}
			l66:
				if !_rules[ruleLWING]() {
					goto l63
				}
			l67:
				{
					position68, tokenIndex68 := position, tokenIndex
					if !_rules[ruleField]() {
						goto l68
					}
					goto l67
				l68:
					position, tokenIndex = position68, tokenIndex68
				}
				if !_rules[ruleRWING]() {
					goto l63
				}
				add(ruleUnion, position64)
			}
			return true
		l63:
			position, tokenIndex = position63, tokenIndex63
			return false
		},
		/* 13 Exception <- <(EXCEPTION Identifier LWING Field* RWING)> */
		func() bool {
			position69, tokenIndex69 := position, tokenIndex
			{
				position70 := position
				if !_rules[ruleEXCEPTION]() {
					goto l69
				}
				if !_rules[ruleIdentifier]() {
					goto l69
				}
				if !_rules[ruleLWING]() {
					goto l69
				}
			l71:
				{
					position72, tokenIndex72 := position, tokenIndex
					if !_rules[ruleField]() {
						goto l72
					}
					goto l71
				l72:
					position, tokenIndex = position72, tokenIndex72
				}
				if !_rules[ruleRWING]() {
					goto l69
				}
				add(ruleException, position70)
			}
			return true
		l69:
			position, tokenIndex = position69, tokenIndex69
			return false
		},
		/* 14 Service <- <(SERVICE Identifier (EXTENDS Identifier)? LWING Function* RWING)> */
		func() bool {
			position73, tokenIndex73 := position, tokenIndex
			{
				position74 := position
				if !_rules[ruleSERVICE]() {
					goto l73
				}
				if !_rules[ruleIdentifier]() {
					goto l73
				}
				{
					position75, tokenIndex75 := position, tokenIndex
					if !_rules[ruleEXTENDS]() {
						goto l75
					}
					if !_rules[ruleIdentifier]() {
						goto l75
					}
					goto l76
				l75:
					position, tokenIndex = position75, tokenIndex75
				}
			l76:
				if !_rules[ruleLWING]() {
					goto l73
				}
			l77:
				{
					position78, tokenIndex78 := position, tokenIndex
					if !_rules[ruleFunction]() {
						goto l78
					}
					goto l77
				l78:
					position, tokenIndex = position78, tokenIndex78
				}
				if !_rules[ruleRWING]() {
					goto l73
				}
				add(ruleService, position74)
			}
			return true
		l73:
			position, tokenIndex = position73, tokenIndex73
			return false
		},
		/* 15 Field <- <(FieldID? FieldReq? FieldType Identifier (EQUAL ConstValue)? XsdFieldOptions ListSeparator?)> */
		func() bool {
			position79, tokenIndex79 := position, tokenIndex
			{
				position80 := position
				{
					position81, tokenIndex81 := position, tokenIndex
					if !_rules[ruleFieldID]() {
						goto l81
					}
					goto l82
				l81:
					position, tokenIndex = position81, tokenIndex81
				}
			l82:
				{
					position83, tokenIndex83 := position, tokenIndex
					if !_rules[ruleFieldReq]() {
						goto l83
					}
					goto l84
				l83:
					position, tokenIndex = position83, tokenIndex83
				}
			l84:
				if !_rules[ruleFieldType]() {
					goto l79
				}
				if !_rules[ruleIdentifier]() {
					goto l79
				}
				{
					position85, tokenIndex85 := position, tokenIndex
					if !_rules[ruleEQUAL]() {
						goto l85
					}
					if !_rules[ruleConstValue]() {
						goto l85
					}
					goto l86
				l85:
					position, tokenIndex = position85, tokenIndex85
				}
			l86:
				if !_rules[ruleXsdFieldOptions]() {
					goto l79
				}
				{
					position87, tokenIndex87 := position, tokenIndex
					if !_rules[ruleListSeparator]() {
						goto l87
					}
					goto l88
				l87:
					position, tokenIndex = position87, tokenIndex87
				}
			l88:
				add(ruleField, position80)
			}
			return true
		l79:
			position, tokenIndex = position79, tokenIndex79
			return false
		},
		/* 16 FieldID <- <(IntConstant COLON)> */
		func() bool {
			position89, tokenIndex89 := position, tokenIndex
			{
				position90 := position
				if !_rules[ruleIntConstant]() {
					goto l89
				}
				if !_rules[ruleCOLON]() {
					goto l89
				}
				add(ruleFieldID, position90)
			}
			return true
		l89:
			position, tokenIndex = position89, tokenIndex89
			return false
		},
		/* 17 FieldReq <- <(<(('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd') / ('o' 'p' 't' 'i' 'o' 'n' 'a' 'l'))> Spacing)> */
		func() bool {
			position91, tokenIndex91 := position, tokenIndex
			{
				position92 := position
				{
					position93 := position
					{
						position94, tokenIndex94 := position, tokenIndex
						if buffer[position] != rune('r') {
							goto l95
						}
						position++
						if buffer[position] != rune('e') {
							goto l95
						}
						position++
						if buffer[position] != rune('q') {
							goto l95
						}
						position++
						if buffer[position] != rune('u') {
							goto l95
						}
						position++
						if buffer[position] != rune('i') {
							goto l95
						}
						position++
						if buffer[position] != rune('r') {
							goto l95
						}
						position++
						if buffer[position] != rune('e') {
							goto l95
						}
						position++
						if buffer[position] != rune('d') {
							goto l95
						}
						position++
						goto l94
					l95:
						position, tokenIndex = position94, tokenIndex94
						if buffer[position] != rune('o') {
							goto l91
						}
						position++
						if buffer[position] != rune('p') {
							goto l91
						}
						position++
						if buffer[position] != rune('t') {
							goto l91
						}
						position++
						if buffer[position] != rune('i') {
							goto l91
						}
						position++
						if buffer[position] != rune('o') {
							goto l91
						}
						position++
						if buffer[position] != rune('n') {
							goto l91
						}
						position++
						if buffer[position] != rune('a') {
							goto l91
						}
						position++
						if buffer[position] != rune('l') {
							goto l91
						}
						position++
					}
				l94:
					add(rulePegText, position93)
				}
				if !_rules[ruleSpacing]() {
					goto l91
				}
				add(ruleFieldReq, position92)
			}
			return true
		l91:
			position, tokenIndex = position91, tokenIndex91
			return false
		},
		/* 18 XsdFieldOptions <- <(XSD_OPTIONAL? / XSD_NILLABLE? / XsdAttrs?)> */
		func() bool {
			{
				position97 := position
				{
					position98, tokenIndex98 := position, tokenIndex
					{
						position100, tokenIndex100 := position, tokenIndex
						if !_rules[ruleXSD_OPTIONAL]() {
							goto l100
						}
						goto l101
					l100:
						position, tokenIndex = position100, tokenIndex100
					}
				l101:
					goto l98

					position, tokenIndex = position98, tokenIndex98
					{
						position103, tokenIndex103 := position, tokenIndex
						if !_rules[ruleXSD_NILLABLE]() {
							goto l103
						}
						goto l104
					l103:
						position, tokenIndex = position103, tokenIndex103
					}
				l104:
					goto l98

					position, tokenIndex = position98, tokenIndex98
					{
						position105, tokenIndex105 := position, tokenIndex
						if !_rules[ruleXsdAttrs]() {
							goto l105
						}
						goto l106
					l105:
						position, tokenIndex = position105, tokenIndex105
					}
				l106:
				}
			l98:
				add(ruleXsdFieldOptions, position97)
			}
			return true
		},
		/* 19 XsdAttrs <- <(XSD_ATTRS LWING Field* RWING)> */
		func() bool {
			position107, tokenIndex107 := position, tokenIndex
			{
				position108 := position
				if !_rules[ruleXSD_ATTRS]() {
					goto l107
				}
				if !_rules[ruleLWING]() {
					goto l107
				}
			l109:
				{
					position110, tokenIndex110 := position, tokenIndex
					if !_rules[ruleField]() {
						goto l110
					}
					goto l109
				l110:
					position, tokenIndex = position110, tokenIndex110
				}
				if !_rules[ruleRWING]() {
					goto l107
				}
				add(ruleXsdAttrs, position108)
			}
			return true
		l107:
			position, tokenIndex = position107, tokenIndex107
			return false
		},
		/* 20 Function <- <(ONEWAY? FunctionType Identifier LPAR Field* RPAR Throws? ListSeparator?)> */
		func() bool {
			position111, tokenIndex111 := position, tokenIndex
			{
				position112 := position
				{
					position113, tokenIndex113 := position, tokenIndex
					if !_rules[ruleONEWAY]() {
						goto l113
					}
					goto l114
				l113:
					position, tokenIndex = position113, tokenIndex113
				}
			l114:
				if !_rules[ruleFunctionType]() {
					goto l111
				}
				if !_rules[ruleIdentifier]() {
					goto l111
				}
				if !_rules[ruleLPAR]() {
					goto l111
				}
			l115:
				{
					position116, tokenIndex116 := position, tokenIndex
					if !_rules[ruleField]() {
						goto l116
					}
					goto l115
				l116:
					position, tokenIndex = position116, tokenIndex116
				}
				if !_rules[ruleRPAR]() {
					goto l111
				}
				{
					position117, tokenIndex117 := position, tokenIndex
					if !_rules[ruleThrows]() {
						goto l117
					}
					goto l118
				l117:
					position, tokenIndex = position117, tokenIndex117
				}
			l118:
				{
					position119, tokenIndex119 := position, tokenIndex
					if !_rules[ruleListSeparator]() {
						goto l119
					}
					goto l120
				l119:
					position, tokenIndex = position119, tokenIndex119
				}
			l120:
				add(ruleFunction, position112)
			}
			return true
		l111:
			position, tokenIndex = position111, tokenIndex111
			return false
		},
		/* 21 FunctionType <- <(VOID / FieldType)> */
		func() bool {
			position121, tokenIndex121 := position, tokenIndex
			{
				position122 := position
				{
					position123, tokenIndex123 := position, tokenIndex
					if !_rules[ruleVOID]() {
						goto l124
					}
					goto l123
				l124:
					position, tokenIndex = position123, tokenIndex123
					if !_rules[ruleFieldType]() {
						goto l121
					}
				}
			l123:
				add(ruleFunctionType, position122)
			}
			return true
		l121:
			position, tokenIndex = position121, tokenIndex121
			return false
		},
		/* 22 Throws <- <(THROWS LPAR Field* RPAR)> */
		func() bool {
			position125, tokenIndex125 := position, tokenIndex
			{
				position126 := position
				if !_rules[ruleTHROWS]() {
					goto l125
				}
				if !_rules[ruleLPAR]() {
					goto l125
				}
			l127:
				{
					position128, tokenIndex128 := position, tokenIndex
					if !_rules[ruleField]() {
						goto l128
					}
					goto l127
				l128:
					position, tokenIndex = position128, tokenIndex128
				}
				if !_rules[ruleRPAR]() {
					goto l125
				}
				add(ruleThrows, position126)
			}
			return true
		l125:
			position, tokenIndex = position125, tokenIndex125
			return false
		},
		/* 23 FieldType <- <(BaseType / ContainerType / Identifier)> */
		func() bool {
			position129, tokenIndex129 := position, tokenIndex
			{
				position130 := position
				{
					position131, tokenIndex131 := position, tokenIndex
					if !_rules[ruleBaseType]() {
						goto l132
					}
					goto l131
				l132:
					position, tokenIndex = position131, tokenIndex131
					if !_rules[ruleContainerType]() {
						goto l133
					}
					goto l131
				l133:
					position, tokenIndex = position131, tokenIndex131
					if !_rules[ruleIdentifier]() {
						goto l129
					}
				}
			l131:
				add(ruleFieldType, position130)
			}
			return true
		l129:
			position, tokenIndex = position129, tokenIndex129
			return false
		},
		/* 24 DefinitionType <- <(BaseType / ContainerType / Identifier)> */
		func() bool {
			position134, tokenIndex134 := position, tokenIndex
			{
				position135 := position
				{
					position136, tokenIndex136 := position, tokenIndex
					if !_rules[ruleBaseType]() {
						goto l137
					}
					goto l136
				l137:
					position, tokenIndex = position136, tokenIndex136
					if !_rules[ruleContainerType]() {
						goto l138
					}
					goto l136
				l138:
					position, tokenIndex = position136, tokenIndex136
					if !_rules[ruleIdentifier]() {
						goto l134
					}
				}
			l136:
				add(ruleDefinitionType, position135)
			}
			return true
		l134:
			position, tokenIndex = position134, tokenIndex134
			return false
		},
		/* 25 BaseType <- <(BOOL / BYTE / I8 / I16 / I32 / I64 / DOUBLE / STRING / BINARY / SLIST / FLOAT)> */
		func() bool {
			position139, tokenIndex139 := position, tokenIndex
			{
				position140 := position
				{
					position141, tokenIndex141 := position, tokenIndex
					if !_rules[ruleBOOL]() {
						goto l142
					}
					goto l141
				l142:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleBYTE]() {
						goto l143
					}
					goto l141
				l143:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleI8]() {
						goto l144
					}
					goto l141
				l144:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleI16]() {
						goto l145
					}
					goto l141
				l145:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleI32]() {
						goto l146
					}
					goto l141
				l146:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleI64]() {
						goto l147
					}
					goto l141
				l147:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleDOUBLE]() {
						goto l148
					}
					goto l141
				l148:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleSTRING]() {
						goto l149
					}
					goto l141
				l149:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleBINARY]() {
						goto l150
					}
					goto l141
				l150:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleSLIST]() {
						goto l151
					}
					goto l141
				l151:
					position, tokenIndex = position141, tokenIndex141
					if !_rules[ruleFLOAT]() {
						goto l139
					}
				}
			l141:
				add(ruleBaseType, position140)
			}
			return true
		l139:
			position, tokenIndex = position139, tokenIndex139
			return false
		},
		/* 26 ContainerType <- <(MapType / SetType / ListType)> */
		func() bool {
			position152, tokenIndex152 := position, tokenIndex
			{
				position153 := position
				{
					position154, tokenIndex154 := position, tokenIndex
					if !_rules[ruleMapType]() {
						goto l155
					}
					goto l154
				l155:
					position, tokenIndex = position154, tokenIndex154
					if !_rules[ruleSetType]() {
						goto l156
					}
					goto l154
				l156:
					position, tokenIndex = position154, tokenIndex154
					if !_rules[ruleListType]() {
						goto l152
					}
				}
			l154:
				add(ruleContainerType, position153)
			}
			return true
		l152:
			position, tokenIndex = position152, tokenIndex152
			return false
		},
		/* 27 MapType <- <(MAP CppType? LPOINT FieldType COMMA FieldType RPOINT)> */
		func() bool {
			position157, tokenIndex157 := position, tokenIndex
			{
				position158 := position
				if !_rules[ruleMAP]() {
					goto l157
				}
				{
					position159, tokenIndex159 := position, tokenIndex
					if !_rules[ruleCppType]() {
						goto l159
					}
					goto l160
				l159:
					position, tokenIndex = position159, tokenIndex159
				}
			l160:
				if !_rules[ruleLPOINT]() {
					goto l157
				}
				if !_rules[ruleFieldType]() {
					goto l157
				}
				if !_rules[ruleCOMMA]() {
					goto l157
				}
				if !_rules[ruleFieldType]() {
					goto l157
				}
				if !_rules[ruleRPOINT]() {
					goto l157
				}
				add(ruleMapType, position158)
			}
			return true
		l157:
			position, tokenIndex = position157, tokenIndex157
			return false
		},
		/* 28 SetType <- <(SET CppType? LPOINT FieldType RPOINT)> */
		func() bool {
			position161, tokenIndex161 := position, tokenIndex
			{
				position162 := position
				if !_rules[ruleSET]() {
					goto l161
				}
				{
					position163, tokenIndex163 := position, tokenIndex
					if !_rules[ruleCppType]() {
						goto l163
					}
					goto l164
				l163:
					position, tokenIndex = position163, tokenIndex163
				}
			l164:
				if !_rules[ruleLPOINT]() {
					goto l161
				}
				if !_rules[ruleFieldType]() {
					goto l161
				}
				if !_rules[ruleRPOINT]() {
					goto l161
				}
				add(ruleSetType, position162)
			}
			return true
		l161:
			position, tokenIndex = position161, tokenIndex161
			return false
		},
		/* 29 ListType <- <(LIST LPOINT FieldType RPOINT CppType?)> */
		func() bool {
			position165, tokenIndex165 := position, tokenIndex
			{
				position166 := position
				if !_rules[ruleLIST]() {
					goto l165
				}
				if !_rules[ruleLPOINT]() {
					goto l165
				}
				if !_rules[ruleFieldType]() {
					goto l165
				}
				if !_rules[ruleRPOINT]() {
					goto l165
				}
				{
					position167, tokenIndex167 := position, tokenIndex
					if !_rules[ruleCppType]() {
						goto l167
					}
					goto l168
				l167:
					position, tokenIndex = position167, tokenIndex167
				}
			l168:
				add(ruleListType, position166)
			}
			return true
		l165:
			position, tokenIndex = position165, tokenIndex165
			return false
		},
		/* 30 CppType <- <(CPP_TYPE Literal)> */
		func() bool {
			position169, tokenIndex169 := position, tokenIndex
			{
				position170 := position
				if !_rules[ruleCPP_TYPE]() {
					goto l169
				}
				if !_rules[ruleLiteral]() {
					goto l169
				}
				add(ruleCppType, position170)
			}
			return true
		l169:
			position, tokenIndex = position169, tokenIndex169
			return false
		},
		/* 31 ConstValue <- <(DoubleConstant / IntConstant / Literal / Identifier / ConstList / ConstMap)> */
		func() bool {
			position171, tokenIndex171 := position, tokenIndex
			{
				position172 := position
				{
					position173, tokenIndex173 := position, tokenIndex
					if !_rules[ruleDoubleConstant]() {
						goto l174
					}
					goto l173
				l174:
					position, tokenIndex = position173, tokenIndex173
					if !_rules[ruleIntConstant]() {
						goto l175
					}
					goto l173
				l175:
					position, tokenIndex = position173, tokenIndex173
					if !_rules[ruleLiteral]() {
						goto l176
					}
					goto l173
				l176:
					position, tokenIndex = position173, tokenIndex173
					if !_rules[ruleIdentifier]() {
						goto l177
					}
					goto l173
				l177:
					position, tokenIndex = position173, tokenIndex173
					if !_rules[ruleConstList]() {
						goto l178
					}
					goto l173
				l178:
					position, tokenIndex = position173, tokenIndex173
					if !_rules[ruleConstMap]() {
						goto l171
					}
				}
			l173:
				add(ruleConstValue, position172)
			}
			return true
		l171:
			position, tokenIndex = position171, tokenIndex171
			return false
		},
		/* 32 DoubleConstant <- <(<(('+' / '-')? ((Digit* '.' Digit+ Exponent?) / (Digit+ Exponent)))> Spacing)> */
		func() bool {
			position179, tokenIndex179 := position, tokenIndex
			{
				position180 := position
				{
					position181 := position
					{
						position182, tokenIndex182 := position, tokenIndex
						{
							position184, tokenIndex184 := position, tokenIndex
							if buffer[position] != rune('+') {
								goto l185
							}
							position++
							goto l184
						l185:
							position, tokenIndex = position184, tokenIndex184
							if buffer[position] != rune('-') {
								goto l182
							}
							position++
						}
					l184:
						goto l183
					l182:
						position, tokenIndex = position182, tokenIndex182
					}
				l183:
					{
						position186, tokenIndex186 := position, tokenIndex
					l188:
						{
							position189, tokenIndex189 := position, tokenIndex
							if !_rules[ruleDigit]() {
								goto l189
							}
							goto l188
						l189:
							position, tokenIndex = position189, tokenIndex189
						}
						if buffer[position] != rune('.') {
							goto l187
						}
						position++
						if !_rules[ruleDigit]() {
							goto l187
						}
					l190:
						{
							position191, tokenIndex191 := position, tokenIndex
							if !_rules[ruleDigit]() {
								goto l191
							}
							goto l190
						l191:
							position, tokenIndex = position191, tokenIndex191
						}
						{
							position192, tokenIndex192 := position, tokenIndex
							if !_rules[ruleExponent]() {
								goto l192
							}
							goto l193
						l192:
							position, tokenIndex = position192, tokenIndex192
						}
					l193:
						goto l186
					l187:
						position, tokenIndex = position186, tokenIndex186
						if !_rules[ruleDigit]() {
							goto l179
						}
					l194:
						{
							position195, tokenIndex195 := position, tokenIndex
							if !_rules[ruleDigit]() {
								goto l195
							}
							goto l194
						l195:
							position, tokenIndex = position195, tokenIndex195
						}
						if !_rules[ruleExponent]() {
							goto l179
						}
					}
				l186:
					add(rulePegText, position181)
				}
				if !_rules[ruleSpacing]() {
					goto l179
				}
				add(ruleDoubleConstant, position180)
			}
			return true
		l179:
			position, tokenIndex = position179, tokenIndex179
			return false
		},
		/* 33 Exponent <- <(('e' / 'E') ('+' / '-')? Digit+)> */
		func() bool {
			position196, tokenIndex196 := position, tokenIndex
			{
				position197 := position
				{
					position198, tokenIndex198 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l199
					}
					position++
					goto l198
				l199:
					position, tokenIndex = position198, tokenIndex198
					if buffer[position] != rune('E') {
						goto l196
					}
					position++
				}
			l198:
				{
					position200, tokenIndex200 := position, tokenIndex
					{
						position202, tokenIndex202 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l203
						}
						position++
						goto l202
					l203:
						position, tokenIndex = position202, tokenIndex202
						if buffer[position] != rune('-') {
							goto l200
						}
						position++
					}
				l202:
					goto l201
				l200:
					position, tokenIndex = position200, tokenIndex200
				}
			l201:
				if !_rules[ruleDigit]() {
					goto l196
				}
			l204:
				{
					position205, tokenIndex205 := position, tokenIndex
					if !_rules[ruleDigit]() {
						goto l205
					}
					goto l204
				l205:
					position, tokenIndex = position205, tokenIndex205
				}
				add(ruleExponent, position197)
			}
			return true
		l196:
			position, tokenIndex = position196, tokenIndex196
			return false
		},
		/* 34 IntConstant <- <(<(('+' / '-')? Digit+)> Spacing)> */
		func() bool {
			position206, tokenIndex206 := position, tokenIndex
			{
				position207 := position
				{
					position208 := position
					{
						position209, tokenIndex209 := position, tokenIndex
						{
							position211, tokenIndex211 := position, tokenIndex
							if buffer[position] != rune('+') {
								goto l212
							}
							position++
							goto l211
						l212:
							position, tokenIndex = position211, tokenIndex211
							if buffer[position] != rune('-') {
								goto l209
							}
							position++
						}
					l211:
						goto l210
					l209:
						position, tokenIndex = position209, tokenIndex209
					}
				l210:
					if !_rules[ruleDigit]() {
						goto l206
					}
				l213:
					{
						position214, tokenIndex214 := position, tokenIndex
						if !_rules[ruleDigit]() {
							goto l214
						}
						goto l213
					l214:
						position, tokenIndex = position214, tokenIndex214
					}
					add(rulePegText, position208)
				}
				if !_rules[ruleSpacing]() {
					goto l206
				}
				add(ruleIntConstant, position207)
			}
			return true
		l206:
			position, tokenIndex = position206, tokenIndex206
			return false
		},
		/* 35 ConstList <- <(LBRK (ConstValue ListSeparator?)* RBRK)> */
		func() bool {
			position215, tokenIndex215 := position, tokenIndex
			{
				position216 := position
				if !_rules[ruleLBRK]() {
					goto l215
				}
			l217:
				{
					position218, tokenIndex218 := position, tokenIndex
					if !_rules[ruleConstValue]() {
						goto l218
					}
					{
						position219, tokenIndex219 := position, tokenIndex
						if !_rules[ruleListSeparator]() {
							goto l219
						}
						goto l220
					l219:
						position, tokenIndex = position219, tokenIndex219
					}
				l220:
					goto l217
				l218:
					position, tokenIndex = position218, tokenIndex218
				}
				if !_rules[ruleRBRK]() {
					goto l215
				}
				add(ruleConstList, position216)
			}
			return true
		l215:
			position, tokenIndex = position215, tokenIndex215
			return false
		},
		/* 36 ConstMap <- <(LWING (ConstValue COLON ConstValue ListSeparator?)* RWING)> */
		func() bool {
			position221, tokenIndex221 := position, tokenIndex
			{
				position222 := position
				if !_rules[ruleLWING]() {
					goto l221
				}
			l223:
				{
					position224, tokenIndex224 := position, tokenIndex
					if !_rules[ruleConstValue]() {
						goto l224
					}
					if !_rules[ruleCOLON]() {
						goto l224
					}
					if !_rules[ruleConstValue]() {
						goto l224
					}
					{
						position225, tokenIndex225 := position, tokenIndex
						if !_rules[ruleListSeparator]() {
							goto l225
						}
						goto l226
					l225:
						position, tokenIndex = position225, tokenIndex225
					}
				l226:
					goto l223
				l224:
					position, tokenIndex = position224, tokenIndex224
				}
				if !_rules[ruleRWING]() {
					goto l221
				}
				add(ruleConstMap, position222)
			}
			return true
		l221:
			position, tokenIndex = position221, tokenIndex221
			return false
		},
		/* 37 Literal <- <((('"' <(!'"' .)*> '"') / ('\'' <(!'\'' .)*> '\'')) Spacing)> */
		func() bool {
			position227, tokenIndex227 := position, tokenIndex
			{
				position228 := position
				{
					position229, tokenIndex229 := position, tokenIndex
					if buffer[position] != rune('"') {
						goto l230
					}
					position++
					{
						position231 := position
					l232:
						{
							position233, tokenIndex233 := position, tokenIndex
							{
								position234, tokenIndex234 := position, tokenIndex
								if buffer[position] != rune('"') {
									goto l234
								}
								position++
								goto l233
							l234:
								position, tokenIndex = position234, tokenIndex234
							}
							if !matchDot() {
								goto l233
							}
							goto l232
						l233:
							position, tokenIndex = position233, tokenIndex233
						}
						add(rulePegText, position231)
					}
					if buffer[position] != rune('"') {
						goto l230
					}
					position++
					goto l229
				l230:
					position, tokenIndex = position229, tokenIndex229
					if buffer[position] != rune('\'') {
						goto l227
					}
					position++
					{
						position235 := position
					l236:
						{
							position237, tokenIndex237 := position, tokenIndex
							{
								position238, tokenIndex238 := position, tokenIndex
								if buffer[position] != rune('\'') {
									goto l238
								}
								position++
								goto l237
							l238:
								position, tokenIndex = position238, tokenIndex238
							}
							if !matchDot() {
								goto l237
							}
							goto l236
						l237:
							position, tokenIndex = position237, tokenIndex237
						}
						add(rulePegText, position235)
					}
					if buffer[position] != rune('\'') {
						goto l227
					}
					position++
				}
			l229:
				if !_rules[ruleSpacing]() {
					goto l227
				}
				add(ruleLiteral, position228)
			}
			return true
		l227:
			position, tokenIndex = position227, tokenIndex227
			return false
		},
		/* 38 Identifier <- <(<((Letter / '_') (Letter / Digit / '.' / '_')*)> Spacing)> */
		func() bool {
			position239, tokenIndex239 := position, tokenIndex
			{
				position240 := position
				{
					position241 := position
					{
						position242, tokenIndex242 := position, tokenIndex
						if !_rules[ruleLetter]() {
							goto l243
						}
						goto l242
					l243:
						position, tokenIndex = position242, tokenIndex242
						if buffer[position] != rune('_') {
							goto l239
						}
						position++
					}
				l242:
				l244:
					{
						position245, tokenIndex245 := position, tokenIndex
						{
							position246, tokenIndex246 := position, tokenIndex
							if !_rules[ruleLetter]() {
								goto l247
							}
							goto l246
						l247:
							position, tokenIndex = position246, tokenIndex246
							if !_rules[ruleDigit]() {
								goto l248
							}
							goto l246
						l248:
							position, tokenIndex = position246, tokenIndex246
							if buffer[position] != rune('.') {
								goto l249
							}
							position++
							goto l246
						l249:
							position, tokenIndex = position246, tokenIndex246
							if buffer[position] != rune('_') {
								goto l245
							}
							position++
						}
					l246:
						goto l244
					l245:
						position, tokenIndex = position245, tokenIndex245
					}
					add(rulePegText, position241)
				}
				if !_rules[ruleSpacing]() {
					goto l239
				}
				add(ruleIdentifier, position240)
			}
			return true
		l239:
			position, tokenIndex = position239, tokenIndex239
			return false
		},
		/* 39 STIdentifier <- <(<((Letter / '_') (Letter / Digit / '.' / '_' / '-')*)> Spacing)> */
		func() bool {
			position250, tokenIndex250 := position, tokenIndex
			{
				position251 := position
				{
					position252 := position
					{
						position253, tokenIndex253 := position, tokenIndex
						if !_rules[ruleLetter]() {
							goto l254
						}
						goto l253
					l254:
						position, tokenIndex = position253, tokenIndex253
						if buffer[position] != rune('_') {
							goto l250
						}
						position++
					}
				l253:
				l255:
					{
						position256, tokenIndex256 := position, tokenIndex
						{
							position257, tokenIndex257 := position, tokenIndex
							if !_rules[ruleLetter]() {
								goto l258
							}
							goto l257
						l258:
							position, tokenIndex = position257, tokenIndex257
							if !_rules[ruleDigit]() {
								goto l259
							}
							goto l257
						l259:
							position, tokenIndex = position257, tokenIndex257
							if buffer[position] != rune('.') {
								goto l260
							}
							position++
							goto l257
						l260:
							position, tokenIndex = position257, tokenIndex257
							if buffer[position] != rune('_') {
								goto l261
							}
							position++
							goto l257
						l261:
							position, tokenIndex = position257, tokenIndex257
							if buffer[position] != rune('-') {
								goto l256
							}
							position++
						}
					l257:
						goto l255
					l256:
						position, tokenIndex = position256, tokenIndex256
					}
					add(rulePegText, position252)
				}
				if !_rules[ruleSpacing]() {
					goto l250
				}
				add(ruleSTIdentifier, position251)
			}
			return true
		l250:
			position, tokenIndex = position250, tokenIndex250
			return false
		},
		/* 40 ListSeparator <- <((',' / ';') Spacing)> */
		func() bool {
			position262, tokenIndex262 := position, tokenIndex
			{
				position263 := position
				{
					position264, tokenIndex264 := position, tokenIndex
					if buffer[position] != rune(',') {
						goto l265
					}
					position++
					goto l264
				l265:
					position, tokenIndex = position264, tokenIndex264
					if buffer[position] != rune(';') {
						goto l262
					}
					position++
				}
			l264:
				if !_rules[ruleSpacing]() {
					goto l262
				}
				add(ruleListSeparator, position263)
			}
			return true
		l262:
			position, tokenIndex = position262, tokenIndex262
			return false
		},
		/* 41 Letter <- <([a-z] / [A-Z])> */
		func() bool {
			position266, tokenIndex266 := position, tokenIndex
			{
				position267 := position
				{
					position268, tokenIndex268 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l269
					}
					position++
					goto l268
				l269:
					position, tokenIndex = position268, tokenIndex268
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l266
					}
					position++
				}
			l268:
				add(ruleLetter, position267)
			}
			return true
		l266:
			position, tokenIndex = position266, tokenIndex266
			return false
		},
		/* 42 Digit <- <[0-9]> */
		func() bool {
			position270, tokenIndex270 := position, tokenIndex
			{
				position271 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l270
				}
				position++
				add(ruleDigit, position271)
			}
			return true
		l270:
			position, tokenIndex = position270, tokenIndex270
			return false
		},
		/* 43 IdChars <- <([a-z] / [A-Z] / [0-9] / ('_' / '$'))> */
		func() bool {
			position272, tokenIndex272 := position, tokenIndex
			{
				position273 := position
				{
					position274, tokenIndex274 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l275
					}
					position++
					goto l274
				l275:
					position, tokenIndex = position274, tokenIndex274
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l276
					}
					position++
					goto l274
				l276:
					position, tokenIndex = position274, tokenIndex274
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l277
					}
					position++
					goto l274
				l277:
					position, tokenIndex = position274, tokenIndex274
					{
						position278, tokenIndex278 := position, tokenIndex
						if buffer[position] != rune('_') {
							goto l279
						}
						position++
						goto l278
					l279:
						position, tokenIndex = position278, tokenIndex278
						if buffer[position] != rune('$') {
							goto l272
						}
						position++
					}
				l278:
				}
			l274:
				add(ruleIdChars, position273)
			}
			return true
		l272:
			position, tokenIndex = position272, tokenIndex272
			return false
		},
		/* 44 Spacing <- <(Whitespace / LongComment / LineComment / Pragma)*> */
		func() bool {
			{
				position281 := position
			l282:
				{
					position283, tokenIndex283 := position, tokenIndex
					{
						position284, tokenIndex284 := position, tokenIndex
						if !_rules[ruleWhitespace]() {
							goto l285
						}
						goto l284
					l285:
						position, tokenIndex = position284, tokenIndex284
						if !_rules[ruleLongComment]() {
							goto l286
						}
						goto l284
					l286:
						position, tokenIndex = position284, tokenIndex284
						if !_rules[ruleLineComment]() {
							goto l287
						}
						goto l284
					l287:
						position, tokenIndex = position284, tokenIndex284
						if !_rules[rulePragma]() {
							goto l283
						}
					}
				l284:
					goto l282
				l283:
					position, tokenIndex = position283, tokenIndex283
				}
				add(ruleSpacing, position281)
			}
			return true
		},
		/* 45 Whitespace <- <(' ' / '\t' / '\r' / '\n')+> */
		func() bool {
			position288, tokenIndex288 := position, tokenIndex
			{
				position289 := position
				{
					position292, tokenIndex292 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l293
					}
					position++
					goto l292
				l293:
					position, tokenIndex = position292, tokenIndex292
					if buffer[position] != rune('\t') {
						goto l294
					}
					position++
					goto l292
				l294:
					position, tokenIndex = position292, tokenIndex292
					if buffer[position] != rune('\r') {
						goto l295
					}
					position++
					goto l292
				l295:
					position, tokenIndex = position292, tokenIndex292
					if buffer[position] != rune('\n') {
						goto l288
					}
					position++
				}
			l292:
			l290:
				{
					position291, tokenIndex291 := position, tokenIndex
					{
						position296, tokenIndex296 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l297
						}
						position++
						goto l296
					l297:
						position, tokenIndex = position296, tokenIndex296
						if buffer[position] != rune('\t') {
							goto l298
						}
						position++
						goto l296
					l298:
						position, tokenIndex = position296, tokenIndex296
						if buffer[position] != rune('\r') {
							goto l299
						}
						position++
						goto l296
					l299:
						position, tokenIndex = position296, tokenIndex296
						if buffer[position] != rune('\n') {
							goto l291
						}
						position++
					}
				l296:
					goto l290
				l291:
					position, tokenIndex = position291, tokenIndex291
				}
				add(ruleWhitespace, position289)
			}
			return true
		l288:
			position, tokenIndex = position288, tokenIndex288
			return false
		},
		/* 46 LongComment <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		func() bool {
			position300, tokenIndex300 := position, tokenIndex
			{
				position301 := position
				if buffer[position] != rune('/') {
					goto l300
				}
				position++
				if buffer[position] != rune('*') {
					goto l300
				}
				position++
			l302:
				{
					position303, tokenIndex303 := position, tokenIndex
					{
						position304, tokenIndex304 := position, tokenIndex
						if buffer[position] != rune('*') {
							goto l304
						}
						position++
						if buffer[position] != rune('/') {
							goto l304
						}
						position++
						goto l303
					l304:
						position, tokenIndex = position304, tokenIndex304
					}
					if !matchDot() {
						goto l303
					}
					goto l302
				l303:
					position, tokenIndex = position303, tokenIndex303
				}
				if buffer[position] != rune('*') {
					goto l300
				}
				position++
				if buffer[position] != rune('/') {
					goto l300
				}
				position++
				add(ruleLongComment, position301)
			}
			return true
		l300:
			position, tokenIndex = position300, tokenIndex300
			return false
		},
		/* 47 LineComment <- <('/' '/' (!('\r' / '\n') .)* ('\r' / '\n'))> */
		func() bool {
			position305, tokenIndex305 := position, tokenIndex
			{
				position306 := position
				if buffer[position] != rune('/') {
					goto l305
				}
				position++
				if buffer[position] != rune('/') {
					goto l305
				}
				position++
			l307:
				{
					position308, tokenIndex308 := position, tokenIndex
					{
						position309, tokenIndex309 := position, tokenIndex
						{
							position310, tokenIndex310 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l311
							}
							position++
							goto l310
						l311:
							position, tokenIndex = position310, tokenIndex310
							if buffer[position] != rune('\n') {
								goto l309
							}
							position++
						}
					l310:
						goto l308
					l309:
						position, tokenIndex = position309, tokenIndex309
					}
					if !matchDot() {
						goto l308
					}
					goto l307
				l308:
					position, tokenIndex = position308, tokenIndex308
				}
				{
					position312, tokenIndex312 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l313
					}
					position++
					goto l312
				l313:
					position, tokenIndex = position312, tokenIndex312
					if buffer[position] != rune('\n') {
						goto l305
					}
					position++
				}
			l312:
				add(ruleLineComment, position306)
			}
			return true
		l305:
			position, tokenIndex = position305, tokenIndex305
			return false
		},
		/* 48 Pragma <- <('#' (!('\r' / '\n') .)* ('\r' / '\n'))> */
		func() bool {
			position314, tokenIndex314 := position, tokenIndex
			{
				position315 := position
				if buffer[position] != rune('#') {
					goto l314
				}
				position++
			l316:
				{
					position317, tokenIndex317 := position, tokenIndex
					{
						position318, tokenIndex318 := position, tokenIndex
						{
							position319, tokenIndex319 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l320
							}
							position++
							goto l319
						l320:
							position, tokenIndex = position319, tokenIndex319
							if buffer[position] != rune('\n') {
								goto l318
							}
							position++
						}
					l319:
						goto l317
					l318:
						position, tokenIndex = position318, tokenIndex318
					}
					if !matchDot() {
						goto l317
					}
					goto l316
				l317:
					position, tokenIndex = position317, tokenIndex317
				}
				{
					position321, tokenIndex321 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l322
					}
					position++
					goto l321
				l322:
					position, tokenIndex = position321, tokenIndex321
					if buffer[position] != rune('\n') {
						goto l314
					}
					position++
				}
			l321:
				add(rulePragma, position315)
			}
			return true
		l314:
			position, tokenIndex = position314, tokenIndex314
			return false
		},
		/* 49 INCLUDE <- <('i' 'n' 'c' 'l' 'u' 'd' 'e' !IdChars Spacing)> */
		func() bool {
			position323, tokenIndex323 := position, tokenIndex
			{
				position324 := position
				if buffer[position] != rune('i') {
					goto l323
				}
				position++
				if buffer[position] != rune('n') {
					goto l323
				}
				position++
				if buffer[position] != rune('c') {
					goto l323
				}
				position++
				if buffer[position] != rune('l') {
					goto l323
				}
				position++
				if buffer[position] != rune('u') {
					goto l323
				}
				position++
				if buffer[position] != rune('d') {
					goto l323
				}
				position++
				if buffer[position] != rune('e') {
					goto l323
				}
				position++
				{
					position325, tokenIndex325 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l325
					}
					goto l323
				l325:
					position, tokenIndex = position325, tokenIndex325
				}
				if !_rules[ruleSpacing]() {
					goto l323
				}
				add(ruleINCLUDE, position324)
			}
			return true
		l323:
			position, tokenIndex = position323, tokenIndex323
			return false
		},
		/* 50 CPP_INCLUDE <- <('c' 'p' 'p' '_' 'i' 'n' 'c' 'l' 'u' 'd' 'e' !IdChars Spacing)> */
		func() bool {
			position326, tokenIndex326 := position, tokenIndex
			{
				position327 := position
				if buffer[position] != rune('c') {
					goto l326
				}
				position++
				if buffer[position] != rune('p') {
					goto l326
				}
				position++
				if buffer[position] != rune('p') {
					goto l326
				}
				position++
				if buffer[position] != rune('_') {
					goto l326
				}
				position++
				if buffer[position] != rune('i') {
					goto l326
				}
				position++
				if buffer[position] != rune('n') {
					goto l326
				}
				position++
				if buffer[position] != rune('c') {
					goto l326
				}
				position++
				if buffer[position] != rune('l') {
					goto l326
				}
				position++
				if buffer[position] != rune('u') {
					goto l326
				}
				position++
				if buffer[position] != rune('d') {
					goto l326
				}
				position++
				if buffer[position] != rune('e') {
					goto l326
				}
				position++
				{
					position328, tokenIndex328 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l328
					}
					goto l326
				l328:
					position, tokenIndex = position328, tokenIndex328
				}
				if !_rules[ruleSpacing]() {
					goto l326
				}
				add(ruleCPP_INCLUDE, position327)
			}
			return true
		l326:
			position, tokenIndex = position326, tokenIndex326
			return false
		},
		/* 51 NAMESPACE <- <('n' 'a' 'm' 'e' 's' 'p' 'a' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position329, tokenIndex329 := position, tokenIndex
			{
				position330 := position
				if buffer[position] != rune('n') {
					goto l329
				}
				position++
				if buffer[position] != rune('a') {
					goto l329
				}
				position++
				if buffer[position] != rune('m') {
					goto l329
				}
				position++
				if buffer[position] != rune('e') {
					goto l329
				}
				position++
				if buffer[position] != rune('s') {
					goto l329
				}
				position++
				if buffer[position] != rune('p') {
					goto l329
				}
				position++
				if buffer[position] != rune('a') {
					goto l329
				}
				position++
				if buffer[position] != rune('c') {
					goto l329
				}
				position++
				if buffer[position] != rune('e') {
					goto l329
				}
				position++
				{
					position331, tokenIndex331 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l331
					}
					goto l329
				l331:
					position, tokenIndex = position331, tokenIndex331
				}
				if !_rules[ruleSpacing]() {
					goto l329
				}
				add(ruleNAMESPACE, position330)
			}
			return true
		l329:
			position, tokenIndex = position329, tokenIndex329
			return false
		},
		/* 52 SMALLTALK_CATEGORY <- <('s' 'm' 'a' 'l' 'l' 't' 'a' 'l' 'k' '.' 'c' 'a' 't' 'e' 'g' 'o' 'r' 'y' !IdChars Spacing)> */
		func() bool {
			position332, tokenIndex332 := position, tokenIndex
			{
				position333 := position
				if buffer[position] != rune('s') {
					goto l332
				}
				position++
				if buffer[position] != rune('m') {
					goto l332
				}
				position++
				if buffer[position] != rune('a') {
					goto l332
				}
				position++
				if buffer[position] != rune('l') {
					goto l332
				}
				position++
				if buffer[position] != rune('l') {
					goto l332
				}
				position++
				if buffer[position] != rune('t') {
					goto l332
				}
				position++
				if buffer[position] != rune('a') {
					goto l332
				}
				position++
				if buffer[position] != rune('l') {
					goto l332
				}
				position++
				if buffer[position] != rune('k') {
					goto l332
				}
				position++
				if buffer[position] != rune('.') {
					goto l332
				}
				position++
				if buffer[position] != rune('c') {
					goto l332
				}
				position++
				if buffer[position] != rune('a') {
					goto l332
				}
				position++
				if buffer[position] != rune('t') {
					goto l332
				}
				position++
				if buffer[position] != rune('e') {
					goto l332
				}
				position++
				if buffer[position] != rune('g') {
					goto l332
				}
				position++
				if buffer[position] != rune('o') {
					goto l332
				}
				position++
				if buffer[position] != rune('r') {
					goto l332
				}
				position++
				if buffer[position] != rune('y') {
					goto l332
				}
				position++
				{
					position334, tokenIndex334 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l334
					}
					goto l332
				l334:
					position, tokenIndex = position334, tokenIndex334
				}
				if !_rules[ruleSpacing]() {
					goto l332
				}
				add(ruleSMALLTALK_CATEGORY, position333)
			}
			return true
		l332:
			position, tokenIndex = position332, tokenIndex332
			return false
		},
		/* 53 SMALLTALK_PREFIX <- <('s' 'm' 'a' 'l' 'l' 't' 'a' 'l' 'k' '.' 'p' 'r' 'e' 'f' 'i' 'x' !IdChars Spacing)> */
		func() bool {
			position335, tokenIndex335 := position, tokenIndex
			{
				position336 := position
				if buffer[position] != rune('s') {
					goto l335
				}
				position++
				if buffer[position] != rune('m') {
					goto l335
				}
				position++
				if buffer[position] != rune('a') {
					goto l335
				}
				position++
				if buffer[position] != rune('l') {
					goto l335
				}
				position++
				if buffer[position] != rune('l') {
					goto l335
				}
				position++
				if buffer[position] != rune('t') {
					goto l335
				}
				position++
				if buffer[position] != rune('a') {
					goto l335
				}
				position++
				if buffer[position] != rune('l') {
					goto l335
				}
				position++
				if buffer[position] != rune('k') {
					goto l335
				}
				position++
				if buffer[position] != rune('.') {
					goto l335
				}
				position++
				if buffer[position] != rune('p') {
					goto l335
				}
				position++
				if buffer[position] != rune('r') {
					goto l335
				}
				position++
				if buffer[position] != rune('e') {
					goto l335
				}
				position++
				if buffer[position] != rune('f') {
					goto l335
				}
				position++
				if buffer[position] != rune('i') {
					goto l335
				}
				position++
				if buffer[position] != rune('x') {
					goto l335
				}
				position++
				{
					position337, tokenIndex337 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l337
					}
					goto l335
				l337:
					position, tokenIndex = position337, tokenIndex337
				}
				if !_rules[ruleSpacing]() {
					goto l335
				}
				add(ruleSMALLTALK_PREFIX, position336)
			}
			return true
		l335:
			position, tokenIndex = position335, tokenIndex335
			return false
		},
		/* 54 PHP_NAMESPACE <- <('p' 'h' 'p' '_' 'n' 'a' 'm' 'e' 's' 'p' 'a' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position338, tokenIndex338 := position, tokenIndex
			{
				position339 := position
				if buffer[position] != rune('p') {
					goto l338
				}
				position++
				if buffer[position] != rune('h') {
					goto l338
				}
				position++
				if buffer[position] != rune('p') {
					goto l338
				}
				position++
				if buffer[position] != rune('_') {
					goto l338
				}
				position++
				if buffer[position] != rune('n') {
					goto l338
				}
				position++
				if buffer[position] != rune('a') {
					goto l338
				}
				position++
				if buffer[position] != rune('m') {
					goto l338
				}
				position++
				if buffer[position] != rune('e') {
					goto l338
				}
				position++
				if buffer[position] != rune('s') {
					goto l338
				}
				position++
				if buffer[position] != rune('p') {
					goto l338
				}
				position++
				if buffer[position] != rune('a') {
					goto l338
				}
				position++
				if buffer[position] != rune('c') {
					goto l338
				}
				position++
				if buffer[position] != rune('e') {
					goto l338
				}
				position++
				{
					position340, tokenIndex340 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l340
					}
					goto l338
				l340:
					position, tokenIndex = position340, tokenIndex340
				}
				if !_rules[ruleSpacing]() {
					goto l338
				}
				add(rulePHP_NAMESPACE, position339)
			}
			return true
		l338:
			position, tokenIndex = position338, tokenIndex338
			return false
		},
		/* 55 XSD_NAMESPACE <- <('x' 's' 'd' '_' 'n' 'a' 'm' 'e' 's' 'p' 'a' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position341, tokenIndex341 := position, tokenIndex
			{
				position342 := position
				if buffer[position] != rune('x') {
					goto l341
				}
				position++
				if buffer[position] != rune('s') {
					goto l341
				}
				position++
				if buffer[position] != rune('d') {
					goto l341
				}
				position++
				if buffer[position] != rune('_') {
					goto l341
				}
				position++
				if buffer[position] != rune('n') {
					goto l341
				}
				position++
				if buffer[position] != rune('a') {
					goto l341
				}
				position++
				if buffer[position] != rune('m') {
					goto l341
				}
				position++
				if buffer[position] != rune('e') {
					goto l341
				}
				position++
				if buffer[position] != rune('s') {
					goto l341
				}
				position++
				if buffer[position] != rune('p') {
					goto l341
				}
				position++
				if buffer[position] != rune('a') {
					goto l341
				}
				position++
				if buffer[position] != rune('c') {
					goto l341
				}
				position++
				if buffer[position] != rune('e') {
					goto l341
				}
				position++
				{
					position343, tokenIndex343 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l343
					}
					goto l341
				l343:
					position, tokenIndex = position343, tokenIndex343
				}
				if !_rules[ruleSpacing]() {
					goto l341
				}
				add(ruleXSD_NAMESPACE, position342)
			}
			return true
		l341:
			position, tokenIndex = position341, tokenIndex341
			return false
		},
		/* 56 CONST <- <('c' 'o' 'n' 's' 't' !IdChars Spacing)> */
		func() bool {
			position344, tokenIndex344 := position, tokenIndex
			{
				position345 := position
				if buffer[position] != rune('c') {
					goto l344
				}
				position++
				if buffer[position] != rune('o') {
					goto l344
				}
				position++
				if buffer[position] != rune('n') {
					goto l344
				}
				position++
				if buffer[position] != rune('s') {
					goto l344
				}
				position++
				if buffer[position] != rune('t') {
					goto l344
				}
				position++
				{
					position346, tokenIndex346 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l346
					}
					goto l344
				l346:
					position, tokenIndex = position346, tokenIndex346
				}
				if !_rules[ruleSpacing]() {
					goto l344
				}
				add(ruleCONST, position345)
			}
			return true
		l344:
			position, tokenIndex = position344, tokenIndex344
			return false
		},
		/* 57 TYPEDEF <- <('t' 'y' 'p' 'e' 'd' 'e' 'f' !IdChars Spacing)> */
		func() bool {
			position347, tokenIndex347 := position, tokenIndex
			{
				position348 := position
				if buffer[position] != rune('t') {
					goto l347
				}
				position++
				if buffer[position] != rune('y') {
					goto l347
				}
				position++
				if buffer[position] != rune('p') {
					goto l347
				}
				position++
				if buffer[position] != rune('e') {
					goto l347
				}
				position++
				if buffer[position] != rune('d') {
					goto l347
				}
				position++
				if buffer[position] != rune('e') {
					goto l347
				}
				position++
				if buffer[position] != rune('f') {
					goto l347
				}
				position++
				{
					position349, tokenIndex349 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l349
					}
					goto l347
				l349:
					position, tokenIndex = position349, tokenIndex349
				}
				if !_rules[ruleSpacing]() {
					goto l347
				}
				add(ruleTYPEDEF, position348)
			}
			return true
		l347:
			position, tokenIndex = position347, tokenIndex347
			return false
		},
		/* 58 ENUM <- <('e' 'n' 'u' 'm' !IdChars Spacing)> */
		func() bool {
			position350, tokenIndex350 := position, tokenIndex
			{
				position351 := position
				if buffer[position] != rune('e') {
					goto l350
				}
				position++
				if buffer[position] != rune('n') {
					goto l350
				}
				position++
				if buffer[position] != rune('u') {
					goto l350
				}
				position++
				if buffer[position] != rune('m') {
					goto l350
				}
				position++
				{
					position352, tokenIndex352 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l352
					}
					goto l350
				l352:
					position, tokenIndex = position352, tokenIndex352
				}
				if !_rules[ruleSpacing]() {
					goto l350
				}
				add(ruleENUM, position351)
			}
			return true
		l350:
			position, tokenIndex = position350, tokenIndex350
			return false
		},
		/* 59 SENUM <- <('s' 'e' 'n' 'u' 'm' !IdChars Spacing)> */
		func() bool {
			position353, tokenIndex353 := position, tokenIndex
			{
				position354 := position
				if buffer[position] != rune('s') {
					goto l353
				}
				position++
				if buffer[position] != rune('e') {
					goto l353
				}
				position++
				if buffer[position] != rune('n') {
					goto l353
				}
				position++
				if buffer[position] != rune('u') {
					goto l353
				}
				position++
				if buffer[position] != rune('m') {
					goto l353
				}
				position++
				{
					position355, tokenIndex355 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l355
					}
					goto l353
				l355:
					position, tokenIndex = position355, tokenIndex355
				}
				if !_rules[ruleSpacing]() {
					goto l353
				}
				add(ruleSENUM, position354)
			}
			return true
		l353:
			position, tokenIndex = position353, tokenIndex353
			return false
		},
		/* 60 STRUCT <- <('s' 't' 'r' 'u' 'c' 't' !IdChars Spacing)> */
		func() bool {
			position356, tokenIndex356 := position, tokenIndex
			{
				position357 := position
				if buffer[position] != rune('s') {
					goto l356
				}
				position++
				if buffer[position] != rune('t') {
					goto l356
				}
				position++
				if buffer[position] != rune('r') {
					goto l356
				}
				position++
				if buffer[position] != rune('u') {
					goto l356
				}
				position++
				if buffer[position] != rune('c') {
					goto l356
				}
				position++
				if buffer[position] != rune('t') {
					goto l356
				}
				position++
				{
					position358, tokenIndex358 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l358
					}
					goto l356
				l358:
					position, tokenIndex = position358, tokenIndex358
				}
				if !_rules[ruleSpacing]() {
					goto l356
				}
				add(ruleSTRUCT, position357)
			}
			return true
		l356:
			position, tokenIndex = position356, tokenIndex356
			return false
		},
		/* 61 UNION <- <('u' 'n' 'i' 'o' 'n' !IdChars Spacing)> */
		func() bool {
			position359, tokenIndex359 := position, tokenIndex
			{
				position360 := position
				if buffer[position] != rune('u') {
					goto l359
				}
				position++
				if buffer[position] != rune('n') {
					goto l359
				}
				position++
				if buffer[position] != rune('i') {
					goto l359
				}
				position++
				if buffer[position] != rune('o') {
					goto l359
				}
				position++
				if buffer[position] != rune('n') {
					goto l359
				}
				position++
				{
					position361, tokenIndex361 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l361
					}
					goto l359
				l361:
					position, tokenIndex = position361, tokenIndex361
				}
				if !_rules[ruleSpacing]() {
					goto l359
				}
				add(ruleUNION, position360)
			}
			return true
		l359:
			position, tokenIndex = position359, tokenIndex359
			return false
		},
		/* 62 SERVICE <- <('s' 'e' 'r' 'v' 'i' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position362, tokenIndex362 := position, tokenIndex
			{
				position363 := position
				if buffer[position] != rune('s') {
					goto l362
				}
				position++
				if buffer[position] != rune('e') {
					goto l362
				}
				position++
				if buffer[position] != rune('r') {
					goto l362
				}
				position++
				if buffer[position] != rune('v') {
					goto l362
				}
				position++
				if buffer[position] != rune('i') {
					goto l362
				}
				position++
				if buffer[position] != rune('c') {
					goto l362
				}
				position++
				if buffer[position] != rune('e') {
					goto l362
				}
				position++
				{
					position364, tokenIndex364 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l364
					}
					goto l362
				l364:
					position, tokenIndex = position364, tokenIndex364
				}
				if !_rules[ruleSpacing]() {
					goto l362
				}
				add(ruleSERVICE, position363)
			}
			return true
		l362:
			position, tokenIndex = position362, tokenIndex362
			return false
		},
		/* 63 EXTENDS <- <('e' 'x' 't' 'e' 'n' 'd' 's' !IdChars Spacing)> */
		func() bool {
			position365, tokenIndex365 := position, tokenIndex
			{
				position366 := position
				if buffer[position] != rune('e') {
					goto l365
				}
				position++
				if buffer[position] != rune('x') {
					goto l365
				}
				position++
				if buffer[position] != rune('t') {
					goto l365
				}
				position++
				if buffer[position] != rune('e') {
					goto l365
				}
				position++
				if buffer[position] != rune('n') {
					goto l365
				}
				position++
				if buffer[position] != rune('d') {
					goto l365
				}
				position++
				if buffer[position] != rune('s') {
					goto l365
				}
				position++
				{
					position367, tokenIndex367 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l367
					}
					goto l365
				l367:
					position, tokenIndex = position367, tokenIndex367
				}
				if !_rules[ruleSpacing]() {
					goto l365
				}
				add(ruleEXTENDS, position366)
			}
			return true
		l365:
			position, tokenIndex = position365, tokenIndex365
			return false
		},
		/* 64 EXCEPTION <- <('e' 'x' 'c' 'e' 'p' 't' 'i' 'o' 'n' !IdChars Spacing)> */
		func() bool {
			position368, tokenIndex368 := position, tokenIndex
			{
				position369 := position
				if buffer[position] != rune('e') {
					goto l368
				}
				position++
				if buffer[position] != rune('x') {
					goto l368
				}
				position++
				if buffer[position] != rune('c') {
					goto l368
				}
				position++
				if buffer[position] != rune('e') {
					goto l368
				}
				position++
				if buffer[position] != rune('p') {
					goto l368
				}
				position++
				if buffer[position] != rune('t') {
					goto l368
				}
				position++
				if buffer[position] != rune('i') {
					goto l368
				}
				position++
				if buffer[position] != rune('o') {
					goto l368
				}
				position++
				if buffer[position] != rune('n') {
					goto l368
				}
				position++
				{
					position370, tokenIndex370 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l370
					}
					goto l368
				l370:
					position, tokenIndex = position370, tokenIndex370
				}
				if !_rules[ruleSpacing]() {
					goto l368
				}
				add(ruleEXCEPTION, position369)
			}
			return true
		l368:
			position, tokenIndex = position368, tokenIndex368
			return false
		},
		/* 65 ONEWAY <- <('o' 'n' 'e' 'w' 'a' 'y' !IdChars Spacing)> */
		func() bool {
			position371, tokenIndex371 := position, tokenIndex
			{
				position372 := position
				if buffer[position] != rune('o') {
					goto l371
				}
				position++
				if buffer[position] != rune('n') {
					goto l371
				}
				position++
				if buffer[position] != rune('e') {
					goto l371
				}
				position++
				if buffer[position] != rune('w') {
					goto l371
				}
				position++
				if buffer[position] != rune('a') {
					goto l371
				}
				position++
				if buffer[position] != rune('y') {
					goto l371
				}
				position++
				{
					position373, tokenIndex373 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l373
					}
					goto l371
				l373:
					position, tokenIndex = position373, tokenIndex373
				}
				if !_rules[ruleSpacing]() {
					goto l371
				}
				add(ruleONEWAY, position372)
			}
			return true
		l371:
			position, tokenIndex = position371, tokenIndex371
			return false
		},
		/* 66 THROWS <- <('t' 'h' 'r' 'o' 'w' 's' !IdChars Spacing)> */
		func() bool {
			position374, tokenIndex374 := position, tokenIndex
			{
				position375 := position
				if buffer[position] != rune('t') {
					goto l374
				}
				position++
				if buffer[position] != rune('h') {
					goto l374
				}
				position++
				if buffer[position] != rune('r') {
					goto l374
				}
				position++
				if buffer[position] != rune('o') {
					goto l374
				}
				position++
				if buffer[position] != rune('w') {
					goto l374
				}
				position++
				if buffer[position] != rune('s') {
					goto l374
				}
				position++
				{
					position376, tokenIndex376 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l376
					}
					goto l374
				l376:
					position, tokenIndex = position376, tokenIndex376
				}
				if !_rules[ruleSpacing]() {
					goto l374
				}
				add(ruleTHROWS, position375)
			}
			return true
		l374:
			position, tokenIndex = position374, tokenIndex374
			return false
		},
		/* 67 CPP_TYPE <- <('c' 'p' 'p' '_' 't' 'y' 'p' 'e' !IdChars Spacing)> */
		func() bool {
			position377, tokenIndex377 := position, tokenIndex
			{
				position378 := position
				if buffer[position] != rune('c') {
					goto l377
				}
				position++
				if buffer[position] != rune('p') {
					goto l377
				}
				position++
				if buffer[position] != rune('p') {
					goto l377
				}
				position++
				if buffer[position] != rune('_') {
					goto l377
				}
				position++
				if buffer[position] != rune('t') {
					goto l377
				}
				position++
				if buffer[position] != rune('y') {
					goto l377
				}
				position++
				if buffer[position] != rune('p') {
					goto l377
				}
				position++
				if buffer[position] != rune('e') {
					goto l377
				}
				position++
				{
					position379, tokenIndex379 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l379
					}
					goto l377
				l379:
					position, tokenIndex = position379, tokenIndex379
				}
				if !_rules[ruleSpacing]() {
					goto l377
				}
				add(ruleCPP_TYPE, position378)
			}
			return true
		l377:
			position, tokenIndex = position377, tokenIndex377
			return false
		},
		/* 68 XSD_ALL <- <('x' 's' 'd' '_' 'a' 'l' 'l' !IdChars Spacing)> */
		func() bool {
			position380, tokenIndex380 := position, tokenIndex
			{
				position381 := position
				if buffer[position] != rune('x') {
					goto l380
				}
				position++
				if buffer[position] != rune('s') {
					goto l380
				}
				position++
				if buffer[position] != rune('d') {
					goto l380
				}
				position++
				if buffer[position] != rune('_') {
					goto l380
				}
				position++
				if buffer[position] != rune('a') {
					goto l380
				}
				position++
				if buffer[position] != rune('l') {
					goto l380
				}
				position++
				if buffer[position] != rune('l') {
					goto l380
				}
				position++
				{
					position382, tokenIndex382 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l382
					}
					goto l380
				l382:
					position, tokenIndex = position382, tokenIndex382
				}
				if !_rules[ruleSpacing]() {
					goto l380
				}
				add(ruleXSD_ALL, position381)
			}
			return true
		l380:
			position, tokenIndex = position380, tokenIndex380
			return false
		},
		/* 69 XSD_OPTIONAL <- <('x' 's' 'd' '_' 'o' 'p' 't' 'i' 'o' 'n' 'a' 'l' !IdChars Spacing)> */
		func() bool {
			position383, tokenIndex383 := position, tokenIndex
			{
				position384 := position
				if buffer[position] != rune('x') {
					goto l383
				}
				position++
				if buffer[position] != rune('s') {
					goto l383
				}
				position++
				if buffer[position] != rune('d') {
					goto l383
				}
				position++
				if buffer[position] != rune('_') {
					goto l383
				}
				position++
				if buffer[position] != rune('o') {
					goto l383
				}
				position++
				if buffer[position] != rune('p') {
					goto l383
				}
				position++
				if buffer[position] != rune('t') {
					goto l383
				}
				position++
				if buffer[position] != rune('i') {
					goto l383
				}
				position++
				if buffer[position] != rune('o') {
					goto l383
				}
				position++
				if buffer[position] != rune('n') {
					goto l383
				}
				position++
				if buffer[position] != rune('a') {
					goto l383
				}
				position++
				if buffer[position] != rune('l') {
					goto l383
				}
				position++
				{
					position385, tokenIndex385 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l385
					}
					goto l383
				l385:
					position, tokenIndex = position385, tokenIndex385
				}
				if !_rules[ruleSpacing]() {
					goto l383
				}
				add(ruleXSD_OPTIONAL, position384)
			}
			return true
		l383:
			position, tokenIndex = position383, tokenIndex383
			return false
		},
		/* 70 XSD_NILLABLE <- <('x' 's' 'd' '_' 'n' 'i' 'l' 'l' 'a' 'b' 'l' 'e' !IdChars Spacing)> */
		func() bool {
			position386, tokenIndex386 := position, tokenIndex
			{
				position387 := position
				if buffer[position] != rune('x') {
					goto l386
				}
				position++
				if buffer[position] != rune('s') {
					goto l386
				}
				position++
				if buffer[position] != rune('d') {
					goto l386
				}
				position++
				if buffer[position] != rune('_') {
					goto l386
				}
				position++
				if buffer[position] != rune('n') {
					goto l386
				}
				position++
				if buffer[position] != rune('i') {
					goto l386
				}
				position++
				if buffer[position] != rune('l') {
					goto l386
				}
				position++
				if buffer[position] != rune('l') {
					goto l386
				}
				position++
				if buffer[position] != rune('a') {
					goto l386
				}
				position++
				if buffer[position] != rune('b') {
					goto l386
				}
				position++
				if buffer[position] != rune('l') {
					goto l386
				}
				position++
				if buffer[position] != rune('e') {
					goto l386
				}
				position++
				{
					position388, tokenIndex388 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l388
					}
					goto l386
				l388:
					position, tokenIndex = position388, tokenIndex388
				}
				if !_rules[ruleSpacing]() {
					goto l386
				}
				add(ruleXSD_NILLABLE, position387)
			}
			return true
		l386:
			position, tokenIndex = position386, tokenIndex386
			return false
		},
		/* 71 XSD_ATTRS <- <('x' 's' 'd' '_' 'a' 't' 't' 'r' 's' !IdChars Spacing)> */
		func() bool {
			position389, tokenIndex389 := position, tokenIndex
			{
				position390 := position
				if buffer[position] != rune('x') {
					goto l389
				}
				position++
				if buffer[position] != rune('s') {
					goto l389
				}
				position++
				if buffer[position] != rune('d') {
					goto l389
				}
				position++
				if buffer[position] != rune('_') {
					goto l389
				}
				position++
				if buffer[position] != rune('a') {
					goto l389
				}
				position++
				if buffer[position] != rune('t') {
					goto l389
				}
				position++
				if buffer[position] != rune('t') {
					goto l389
				}
				position++
				if buffer[position] != rune('r') {
					goto l389
				}
				position++
				if buffer[position] != rune('s') {
					goto l389
				}
				position++
				{
					position391, tokenIndex391 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l391
					}
					goto l389
				l391:
					position, tokenIndex = position391, tokenIndex391
				}
				if !_rules[ruleSpacing]() {
					goto l389
				}
				add(ruleXSD_ATTRS, position390)
			}
			return true
		l389:
			position, tokenIndex = position389, tokenIndex389
			return false
		},
		/* 72 VOID <- <('v' 'o' 'i' 'd' !IdChars Spacing)> */
		func() bool {
			position392, tokenIndex392 := position, tokenIndex
			{
				position393 := position
				if buffer[position] != rune('v') {
					goto l392
				}
				position++
				if buffer[position] != rune('o') {
					goto l392
				}
				position++
				if buffer[position] != rune('i') {
					goto l392
				}
				position++
				if buffer[position] != rune('d') {
					goto l392
				}
				position++
				{
					position394, tokenIndex394 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l394
					}
					goto l392
				l394:
					position, tokenIndex = position394, tokenIndex394
				}
				if !_rules[ruleSpacing]() {
					goto l392
				}
				add(ruleVOID, position393)
			}
			return true
		l392:
			position, tokenIndex = position392, tokenIndex392
			return false
		},
		/* 73 MAP <- <('m' 'a' 'p' !IdChars Spacing)> */
		func() bool {
			position395, tokenIndex395 := position, tokenIndex
			{
				position396 := position
				if buffer[position] != rune('m') {
					goto l395
				}
				position++
				if buffer[position] != rune('a') {
					goto l395
				}
				position++
				if buffer[position] != rune('p') {
					goto l395
				}
				position++
				{
					position397, tokenIndex397 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l397
					}
					goto l395
				l397:
					position, tokenIndex = position397, tokenIndex397
				}
				if !_rules[ruleSpacing]() {
					goto l395
				}
				add(ruleMAP, position396)
			}
			return true
		l395:
			position, tokenIndex = position395, tokenIndex395
			return false
		},
		/* 74 SET <- <('s' 'e' 't' !IdChars Spacing)> */
		func() bool {
			position398, tokenIndex398 := position, tokenIndex
			{
				position399 := position
				if buffer[position] != rune('s') {
					goto l398
				}
				position++
				if buffer[position] != rune('e') {
					goto l398
				}
				position++
				if buffer[position] != rune('t') {
					goto l398
				}
				position++
				{
					position400, tokenIndex400 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l400
					}
					goto l398
				l400:
					position, tokenIndex = position400, tokenIndex400
				}
				if !_rules[ruleSpacing]() {
					goto l398
				}
				add(ruleSET, position399)
			}
			return true
		l398:
			position, tokenIndex = position398, tokenIndex398
			return false
		},
		/* 75 LIST <- <('l' 'i' 's' 't' !IdChars Spacing)> */
		func() bool {
			position401, tokenIndex401 := position, tokenIndex
			{
				position402 := position
				if buffer[position] != rune('l') {
					goto l401
				}
				position++
				if buffer[position] != rune('i') {
					goto l401
				}
				position++
				if buffer[position] != rune('s') {
					goto l401
				}
				position++
				if buffer[position] != rune('t') {
					goto l401
				}
				position++
				{
					position403, tokenIndex403 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l403
					}
					goto l401
				l403:
					position, tokenIndex = position403, tokenIndex403
				}
				if !_rules[ruleSpacing]() {
					goto l401
				}
				add(ruleLIST, position402)
			}
			return true
		l401:
			position, tokenIndex = position401, tokenIndex401
			return false
		},
		/* 76 BOOL <- <(<('b' 'o' 'o' 'l')> !IdChars Spacing)> */
		func() bool {
			position404, tokenIndex404 := position, tokenIndex
			{
				position405 := position
				{
					position406 := position
					if buffer[position] != rune('b') {
						goto l404
					}
					position++
					if buffer[position] != rune('o') {
						goto l404
					}
					position++
					if buffer[position] != rune('o') {
						goto l404
					}
					position++
					if buffer[position] != rune('l') {
						goto l404
					}
					position++
					add(rulePegText, position406)
				}
				{
					position407, tokenIndex407 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l407
					}
					goto l404
				l407:
					position, tokenIndex = position407, tokenIndex407
				}
				if !_rules[ruleSpacing]() {
					goto l404
				}
				add(ruleBOOL, position405)
			}
			return true
		l404:
			position, tokenIndex = position404, tokenIndex404
			return false
		},
		/* 77 BYTE <- <(<('b' 'y' 't' 'e')> !IdChars Spacing)> */
		func() bool {
			position408, tokenIndex408 := position, tokenIndex
			{
				position409 := position
				{
					position410 := position
					if buffer[position] != rune('b') {
						goto l408
					}
					position++
					if buffer[position] != rune('y') {
						goto l408
					}
					position++
					if buffer[position] != rune('t') {
						goto l408
					}
					position++
					if buffer[position] != rune('e') {
						goto l408
					}
					position++
					add(rulePegText, position410)
				}
				{
					position411, tokenIndex411 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l411
					}
					goto l408
				l411:
					position, tokenIndex = position411, tokenIndex411
				}
				if !_rules[ruleSpacing]() {
					goto l408
				}
				add(ruleBYTE, position409)
			}
			return true
		l408:
			position, tokenIndex = position408, tokenIndex408
			return false
		},
		/* 78 I8 <- <(<('i' '8')> !IdChars Spacing)> */
		func() bool {
			position412, tokenIndex412 := position, tokenIndex
			{
				position413 := position
				{
					position414 := position
					if buffer[position] != rune('i') {
						goto l412
					}
					position++
					if buffer[position] != rune('8') {
						goto l412
					}
					position++
					add(rulePegText, position414)
				}
				{
					position415, tokenIndex415 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l415
					}
					goto l412
				l415:
					position, tokenIndex = position415, tokenIndex415
				}
				if !_rules[ruleSpacing]() {
					goto l412
				}
				add(ruleI8, position413)
			}
			return true
		l412:
			position, tokenIndex = position412, tokenIndex412
			return false
		},
		/* 79 I16 <- <(<('i' '1' '6')> !IdChars Spacing)> */
		func() bool {
			position416, tokenIndex416 := position, tokenIndex
			{
				position417 := position
				{
					position418 := position
					if buffer[position] != rune('i') {
						goto l416
					}
					position++
					if buffer[position] != rune('1') {
						goto l416
					}
					position++
					if buffer[position] != rune('6') {
						goto l416
					}
					position++
					add(rulePegText, position418)
				}
				{
					position419, tokenIndex419 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l419
					}
					goto l416
				l419:
					position, tokenIndex = position419, tokenIndex419
				}
				if !_rules[ruleSpacing]() {
					goto l416
				}
				add(ruleI16, position417)
			}
			return true
		l416:
			position, tokenIndex = position416, tokenIndex416
			return false
		},
		/* 80 I32 <- <(<('i' '3' '2')> !IdChars Spacing)> */
		func() bool {
			position420, tokenIndex420 := position, tokenIndex
			{
				position421 := position
				{
					position422 := position
					if buffer[position] != rune('i') {
						goto l420
					}
					position++
					if buffer[position] != rune('3') {
						goto l420
					}
					position++
					if buffer[position] != rune('2') {
						goto l420
					}
					position++
					add(rulePegText, position422)
				}
				{
					position423, tokenIndex423 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l423
					}
					goto l420
				l423:
					position, tokenIndex = position423, tokenIndex423
				}
				if !_rules[ruleSpacing]() {
					goto l420
				}
				add(ruleI32, position421)
			}
			return true
		l420:
			position, tokenIndex = position420, tokenIndex420
			return false
		},
		/* 81 I64 <- <(<('i' '6' '4')> !IdChars Spacing)> */
		func() bool {
			position424, tokenIndex424 := position, tokenIndex
			{
				position425 := position
				{
					position426 := position
					if buffer[position] != rune('i') {
						goto l424
					}
					position++
					if buffer[position] != rune('6') {
						goto l424
					}
					position++
					if buffer[position] != rune('4') {
						goto l424
					}
					position++
					add(rulePegText, position426)
				}
				{
					position427, tokenIndex427 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l427
					}
					goto l424
				l427:
					position, tokenIndex = position427, tokenIndex427
				}
				if !_rules[ruleSpacing]() {
					goto l424
				}
				add(ruleI64, position425)
			}
			return true
		l424:
			position, tokenIndex = position424, tokenIndex424
			return false
		},
		/* 82 DOUBLE <- <(<('d' 'o' 'u' 'b' 'l' 'e')> !IdChars Spacing)> */
		func() bool {
			position428, tokenIndex428 := position, tokenIndex
			{
				position429 := position
				{
					position430 := position
					if buffer[position] != rune('d') {
						goto l428
					}
					position++
					if buffer[position] != rune('o') {
						goto l428
					}
					position++
					if buffer[position] != rune('u') {
						goto l428
					}
					position++
					if buffer[position] != rune('b') {
						goto l428
					}
					position++
					if buffer[position] != rune('l') {
						goto l428
					}
					position++
					if buffer[position] != rune('e') {
						goto l428
					}
					position++
					add(rulePegText, position430)
				}
				{
					position431, tokenIndex431 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l431
					}
					goto l428
				l431:
					position, tokenIndex = position431, tokenIndex431
				}
				if !_rules[ruleSpacing]() {
					goto l428
				}
				add(ruleDOUBLE, position429)
			}
			return true
		l428:
			position, tokenIndex = position428, tokenIndex428
			return false
		},
		/* 83 STRING <- <(<('s' 't' 'r' 'i' 'n' 'g')> !IdChars Spacing)> */
		func() bool {
			position432, tokenIndex432 := position, tokenIndex
			{
				position433 := position
				{
					position434 := position
					if buffer[position] != rune('s') {
						goto l432
					}
					position++
					if buffer[position] != rune('t') {
						goto l432
					}
					position++
					if buffer[position] != rune('r') {
						goto l432
					}
					position++
					if buffer[position] != rune('i') {
						goto l432
					}
					position++
					if buffer[position] != rune('n') {
						goto l432
					}
					position++
					if buffer[position] != rune('g') {
						goto l432
					}
					position++
					add(rulePegText, position434)
				}
				{
					position435, tokenIndex435 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l435
					}
					goto l432
				l435:
					position, tokenIndex = position435, tokenIndex435
				}
				if !_rules[ruleSpacing]() {
					goto l432
				}
				add(ruleSTRING, position433)
			}
			return true
		l432:
			position, tokenIndex = position432, tokenIndex432
			return false
		},
		/* 84 BINARY <- <(<('b' 'i' 'n' 'a' 'r' 'y')> !IdChars Spacing)> */
		func() bool {
			position436, tokenIndex436 := position, tokenIndex
			{
				position437 := position
				{
					position438 := position
					if buffer[position] != rune('b') {
						goto l436
					}
					position++
					if buffer[position] != rune('i') {
						goto l436
					}
					position++
					if buffer[position] != rune('n') {
						goto l436
					}
					position++
					if buffer[position] != rune('a') {
						goto l436
					}
					position++
					if buffer[position] != rune('r') {
						goto l436
					}
					position++
					if buffer[position] != rune('y') {
						goto l436
					}
					position++
					add(rulePegText, position438)
				}
				{
					position439, tokenIndex439 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l439
					}
					goto l436
				l439:
					position, tokenIndex = position439, tokenIndex439
				}
				if !_rules[ruleSpacing]() {
					goto l436
				}
				add(ruleBINARY, position437)
			}
			return true
		l436:
			position, tokenIndex = position436, tokenIndex436
			return false
		},
		/* 85 SLIST <- <(<('s' 'l' 'i' 's' 't')> !IdChars Spacing)> */
		func() bool {
			position440, tokenIndex440 := position, tokenIndex
			{
				position441 := position
				{
					position442 := position
					if buffer[position] != rune('s') {
						goto l440
					}
					position++
					if buffer[position] != rune('l') {
						goto l440
					}
					position++
					if buffer[position] != rune('i') {
						goto l440
					}
					position++
					if buffer[position] != rune('s') {
						goto l440
					}
					position++
					if buffer[position] != rune('t') {
						goto l440
					}
					position++
					add(rulePegText, position442)
				}
				{
					position443, tokenIndex443 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l443
					}
					goto l440
				l443:
					position, tokenIndex = position443, tokenIndex443
				}
				if !_rules[ruleSpacing]() {
					goto l440
				}
				add(ruleSLIST, position441)
			}
			return true
		l440:
			position, tokenIndex = position440, tokenIndex440
			return false
		},
		/* 86 FLOAT <- <(<('f' 'l' 'o' 'a' 't')> !IdChars Spacing)> */
		func() bool {
			position444, tokenIndex444 := position, tokenIndex
			{
				position445 := position
				{
					position446 := position
					if buffer[position] != rune('f') {
						goto l444
					}
					position++
					if buffer[position] != rune('l') {
						goto l444
					}
					position++
					if buffer[position] != rune('o') {
						goto l444
					}
					position++
					if buffer[position] != rune('a') {
						goto l444
					}
					position++
					if buffer[position] != rune('t') {
						goto l444
					}
					position++
					add(rulePegText, position446)
				}
				{
					position447, tokenIndex447 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l447
					}
					goto l444
				l447:
					position, tokenIndex = position447, tokenIndex447
				}
				if !_rules[ruleSpacing]() {
					goto l444
				}
				add(ruleFLOAT, position445)
			}
			return true
		l444:
			position, tokenIndex = position444, tokenIndex444
			return false
		},
		/* 87 LBRK <- <('[' Spacing)> */
		func() bool {
			position448, tokenIndex448 := position, tokenIndex
			{
				position449 := position
				if buffer[position] != rune('[') {
					goto l448
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l448
				}
				add(ruleLBRK, position449)
			}
			return true
		l448:
			position, tokenIndex = position448, tokenIndex448
			return false
		},
		/* 88 RBRK <- <(']' Spacing)> */
		func() bool {
			position450, tokenIndex450 := position, tokenIndex
			{
				position451 := position
				if buffer[position] != rune(']') {
					goto l450
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l450
				}
				add(ruleRBRK, position451)
			}
			return true
		l450:
			position, tokenIndex = position450, tokenIndex450
			return false
		},
		/* 89 LPAR <- <('(' Spacing)> */
		func() bool {
			position452, tokenIndex452 := position, tokenIndex
			{
				position453 := position
				if buffer[position] != rune('(') {
					goto l452
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l452
				}
				add(ruleLPAR, position453)
			}
			return true
		l452:
			position, tokenIndex = position452, tokenIndex452
			return false
		},
		/* 90 RPAR <- <(')' Spacing)> */
		func() bool {
			position454, tokenIndex454 := position, tokenIndex
			{
				position455 := position
				if buffer[position] != rune(')') {
					goto l454
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l454
				}
				add(ruleRPAR, position455)
			}
			return true
		l454:
			position, tokenIndex = position454, tokenIndex454
			return false
		},
		/* 91 LWING <- <('{' Spacing)> */
		func() bool {
			position456, tokenIndex456 := position, tokenIndex
			{
				position457 := position
				if buffer[position] != rune('{') {
					goto l456
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l456
				}
				add(ruleLWING, position457)
			}
			return true
		l456:
			position, tokenIndex = position456, tokenIndex456
			return false
		},
		/* 92 RWING <- <('}' Spacing)> */
		func() bool {
			position458, tokenIndex458 := position, tokenIndex
			{
				position459 := position
				if buffer[position] != rune('}') {
					goto l458
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l458
				}
				add(ruleRWING, position459)
			}
			return true
		l458:
			position, tokenIndex = position458, tokenIndex458
			return false
		},
		/* 93 LPOINT <- <('<' Spacing)> */
		func() bool {
			position460, tokenIndex460 := position, tokenIndex
			{
				position461 := position
				if buffer[position] != rune('<') {
					goto l460
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l460
				}
				add(ruleLPOINT, position461)
			}
			return true
		l460:
			position, tokenIndex = position460, tokenIndex460
			return false
		},
		/* 94 RPOINT <- <('>' Spacing)> */
		func() bool {
			position462, tokenIndex462 := position, tokenIndex
			{
				position463 := position
				if buffer[position] != rune('>') {
					goto l462
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l462
				}
				add(ruleRPOINT, position463)
			}
			return true
		l462:
			position, tokenIndex = position462, tokenIndex462
			return false
		},
		/* 95 EQUAL <- <('=' !'=' Spacing)> */
		func() bool {
			position464, tokenIndex464 := position, tokenIndex
			{
				position465 := position
				if buffer[position] != rune('=') {
					goto l464
				}
				position++
				{
					position466, tokenIndex466 := position, tokenIndex
					if buffer[position] != rune('=') {
						goto l466
					}
					position++
					goto l464
				l466:
					position, tokenIndex = position466, tokenIndex466
				}
				if !_rules[ruleSpacing]() {
					goto l464
				}
				add(ruleEQUAL, position465)
			}
			return true
		l464:
			position, tokenIndex = position464, tokenIndex464
			return false
		},
		/* 96 COMMA <- <(',' Spacing)> */
		func() bool {
			position467, tokenIndex467 := position, tokenIndex
			{
				position468 := position
				if buffer[position] != rune(',') {
					goto l467
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l467
				}
				add(ruleCOMMA, position468)
			}
			return true
		l467:
			position, tokenIndex = position467, tokenIndex467
			return false
		},
		/* 97 COLON <- <(':' Spacing)> */
		func() bool {
			position469, tokenIndex469 := position, tokenIndex
			{
				position470 := position
				if buffer[position] != rune(':') {
					goto l469
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l469
				}
				add(ruleCOLON, position470)
			}
			return true
		l469:
			position, tokenIndex = position469, tokenIndex469
			return false
		},
		/* 98 EOT <- <!.> */
		func() bool {
			position471, tokenIndex471 := position, tokenIndex
			{
				position472 := position
				{
					position473, tokenIndex473 := position, tokenIndex
					if !matchDot() {
						goto l473
					}
					goto l471
				l473:
					position, tokenIndex = position473, tokenIndex473
				}
				add(ruleEOT, position472)
			}
			return true
		l471:
			position, tokenIndex = position471, tokenIndex471
			return false
		},
		nil,
	}
	p.rules = _rules
}
