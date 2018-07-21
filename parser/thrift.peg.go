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
	rules  [100]func() bool
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
		/* 25 BaseType <- <(BOOL / BYTE / I8 / I16 / I32 / I64 / DOUBLE / STRING / BINARY / SLIST)> */
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
			position151, tokenIndex151 := position, tokenIndex
			{
				position152 := position
				{
					position153, tokenIndex153 := position, tokenIndex
					if !_rules[ruleMapType]() {
						goto l154
					}
					goto l153
				l154:
					position, tokenIndex = position153, tokenIndex153
					if !_rules[ruleSetType]() {
						goto l155
					}
					goto l153
				l155:
					position, tokenIndex = position153, tokenIndex153
					if !_rules[ruleListType]() {
						goto l151
					}
				}
			l153:
				add(ruleContainerType, position152)
			}
			return true
		l151:
			position, tokenIndex = position151, tokenIndex151
			return false
		},
		/* 27 MapType <- <(MAP CppType? LPOINT FieldType COMMA FieldType RPOINT)> */
		func() bool {
			position156, tokenIndex156 := position, tokenIndex
			{
				position157 := position
				if !_rules[ruleMAP]() {
					goto l156
				}
				{
					position158, tokenIndex158 := position, tokenIndex
					if !_rules[ruleCppType]() {
						goto l158
					}
					goto l159
				l158:
					position, tokenIndex = position158, tokenIndex158
				}
			l159:
				if !_rules[ruleLPOINT]() {
					goto l156
				}
				if !_rules[ruleFieldType]() {
					goto l156
				}
				if !_rules[ruleCOMMA]() {
					goto l156
				}
				if !_rules[ruleFieldType]() {
					goto l156
				}
				if !_rules[ruleRPOINT]() {
					goto l156
				}
				add(ruleMapType, position157)
			}
			return true
		l156:
			position, tokenIndex = position156, tokenIndex156
			return false
		},
		/* 28 SetType <- <(SET CppType? LPOINT FieldType RPOINT)> */
		func() bool {
			position160, tokenIndex160 := position, tokenIndex
			{
				position161 := position
				if !_rules[ruleSET]() {
					goto l160
				}
				{
					position162, tokenIndex162 := position, tokenIndex
					if !_rules[ruleCppType]() {
						goto l162
					}
					goto l163
				l162:
					position, tokenIndex = position162, tokenIndex162
				}
			l163:
				if !_rules[ruleLPOINT]() {
					goto l160
				}
				if !_rules[ruleFieldType]() {
					goto l160
				}
				if !_rules[ruleRPOINT]() {
					goto l160
				}
				add(ruleSetType, position161)
			}
			return true
		l160:
			position, tokenIndex = position160, tokenIndex160
			return false
		},
		/* 29 ListType <- <(LIST LPOINT FieldType RPOINT CppType?)> */
		func() bool {
			position164, tokenIndex164 := position, tokenIndex
			{
				position165 := position
				if !_rules[ruleLIST]() {
					goto l164
				}
				if !_rules[ruleLPOINT]() {
					goto l164
				}
				if !_rules[ruleFieldType]() {
					goto l164
				}
				if !_rules[ruleRPOINT]() {
					goto l164
				}
				{
					position166, tokenIndex166 := position, tokenIndex
					if !_rules[ruleCppType]() {
						goto l166
					}
					goto l167
				l166:
					position, tokenIndex = position166, tokenIndex166
				}
			l167:
				add(ruleListType, position165)
			}
			return true
		l164:
			position, tokenIndex = position164, tokenIndex164
			return false
		},
		/* 30 CppType <- <(CPP_TYPE Literal)> */
		func() bool {
			position168, tokenIndex168 := position, tokenIndex
			{
				position169 := position
				if !_rules[ruleCPP_TYPE]() {
					goto l168
				}
				if !_rules[ruleLiteral]() {
					goto l168
				}
				add(ruleCppType, position169)
			}
			return true
		l168:
			position, tokenIndex = position168, tokenIndex168
			return false
		},
		/* 31 ConstValue <- <(DoubleConstant / IntConstant / Literal / Identifier / ConstList / ConstMap)> */
		func() bool {
			position170, tokenIndex170 := position, tokenIndex
			{
				position171 := position
				{
					position172, tokenIndex172 := position, tokenIndex
					if !_rules[ruleDoubleConstant]() {
						goto l173
					}
					goto l172
				l173:
					position, tokenIndex = position172, tokenIndex172
					if !_rules[ruleIntConstant]() {
						goto l174
					}
					goto l172
				l174:
					position, tokenIndex = position172, tokenIndex172
					if !_rules[ruleLiteral]() {
						goto l175
					}
					goto l172
				l175:
					position, tokenIndex = position172, tokenIndex172
					if !_rules[ruleIdentifier]() {
						goto l176
					}
					goto l172
				l176:
					position, tokenIndex = position172, tokenIndex172
					if !_rules[ruleConstList]() {
						goto l177
					}
					goto l172
				l177:
					position, tokenIndex = position172, tokenIndex172
					if !_rules[ruleConstMap]() {
						goto l170
					}
				}
			l172:
				add(ruleConstValue, position171)
			}
			return true
		l170:
			position, tokenIndex = position170, tokenIndex170
			return false
		},
		/* 32 DoubleConstant <- <(<(('+' / '-')? ((Digit* '.' Digit+ Exponent?) / (Digit+ Exponent)))> Spacing)> */
		func() bool {
			position178, tokenIndex178 := position, tokenIndex
			{
				position179 := position
				{
					position180 := position
					{
						position181, tokenIndex181 := position, tokenIndex
						{
							position183, tokenIndex183 := position, tokenIndex
							if buffer[position] != rune('+') {
								goto l184
							}
							position++
							goto l183
						l184:
							position, tokenIndex = position183, tokenIndex183
							if buffer[position] != rune('-') {
								goto l181
							}
							position++
						}
					l183:
						goto l182
					l181:
						position, tokenIndex = position181, tokenIndex181
					}
				l182:
					{
						position185, tokenIndex185 := position, tokenIndex
					l187:
						{
							position188, tokenIndex188 := position, tokenIndex
							if !_rules[ruleDigit]() {
								goto l188
							}
							goto l187
						l188:
							position, tokenIndex = position188, tokenIndex188
						}
						if buffer[position] != rune('.') {
							goto l186
						}
						position++
						if !_rules[ruleDigit]() {
							goto l186
						}
					l189:
						{
							position190, tokenIndex190 := position, tokenIndex
							if !_rules[ruleDigit]() {
								goto l190
							}
							goto l189
						l190:
							position, tokenIndex = position190, tokenIndex190
						}
						{
							position191, tokenIndex191 := position, tokenIndex
							if !_rules[ruleExponent]() {
								goto l191
							}
							goto l192
						l191:
							position, tokenIndex = position191, tokenIndex191
						}
					l192:
						goto l185
					l186:
						position, tokenIndex = position185, tokenIndex185
						if !_rules[ruleDigit]() {
							goto l178
						}
					l193:
						{
							position194, tokenIndex194 := position, tokenIndex
							if !_rules[ruleDigit]() {
								goto l194
							}
							goto l193
						l194:
							position, tokenIndex = position194, tokenIndex194
						}
						if !_rules[ruleExponent]() {
							goto l178
						}
					}
				l185:
					add(rulePegText, position180)
				}
				if !_rules[ruleSpacing]() {
					goto l178
				}
				add(ruleDoubleConstant, position179)
			}
			return true
		l178:
			position, tokenIndex = position178, tokenIndex178
			return false
		},
		/* 33 Exponent <- <(('e' / 'E') ('+' / '-')? Digit+)> */
		func() bool {
			position195, tokenIndex195 := position, tokenIndex
			{
				position196 := position
				{
					position197, tokenIndex197 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l198
					}
					position++
					goto l197
				l198:
					position, tokenIndex = position197, tokenIndex197
					if buffer[position] != rune('E') {
						goto l195
					}
					position++
				}
			l197:
				{
					position199, tokenIndex199 := position, tokenIndex
					{
						position201, tokenIndex201 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l202
						}
						position++
						goto l201
					l202:
						position, tokenIndex = position201, tokenIndex201
						if buffer[position] != rune('-') {
							goto l199
						}
						position++
					}
				l201:
					goto l200
				l199:
					position, tokenIndex = position199, tokenIndex199
				}
			l200:
				if !_rules[ruleDigit]() {
					goto l195
				}
			l203:
				{
					position204, tokenIndex204 := position, tokenIndex
					if !_rules[ruleDigit]() {
						goto l204
					}
					goto l203
				l204:
					position, tokenIndex = position204, tokenIndex204
				}
				add(ruleExponent, position196)
			}
			return true
		l195:
			position, tokenIndex = position195, tokenIndex195
			return false
		},
		/* 34 IntConstant <- <(<(('+' / '-')? Digit+)> Spacing)> */
		func() bool {
			position205, tokenIndex205 := position, tokenIndex
			{
				position206 := position
				{
					position207 := position
					{
						position208, tokenIndex208 := position, tokenIndex
						{
							position210, tokenIndex210 := position, tokenIndex
							if buffer[position] != rune('+') {
								goto l211
							}
							position++
							goto l210
						l211:
							position, tokenIndex = position210, tokenIndex210
							if buffer[position] != rune('-') {
								goto l208
							}
							position++
						}
					l210:
						goto l209
					l208:
						position, tokenIndex = position208, tokenIndex208
					}
				l209:
					if !_rules[ruleDigit]() {
						goto l205
					}
				l212:
					{
						position213, tokenIndex213 := position, tokenIndex
						if !_rules[ruleDigit]() {
							goto l213
						}
						goto l212
					l213:
						position, tokenIndex = position213, tokenIndex213
					}
					add(rulePegText, position207)
				}
				if !_rules[ruleSpacing]() {
					goto l205
				}
				add(ruleIntConstant, position206)
			}
			return true
		l205:
			position, tokenIndex = position205, tokenIndex205
			return false
		},
		/* 35 ConstList <- <(LBRK (ConstValue ListSeparator?)* RBRK)> */
		func() bool {
			position214, tokenIndex214 := position, tokenIndex
			{
				position215 := position
				if !_rules[ruleLBRK]() {
					goto l214
				}
			l216:
				{
					position217, tokenIndex217 := position, tokenIndex
					if !_rules[ruleConstValue]() {
						goto l217
					}
					{
						position218, tokenIndex218 := position, tokenIndex
						if !_rules[ruleListSeparator]() {
							goto l218
						}
						goto l219
					l218:
						position, tokenIndex = position218, tokenIndex218
					}
				l219:
					goto l216
				l217:
					position, tokenIndex = position217, tokenIndex217
				}
				if !_rules[ruleRBRK]() {
					goto l214
				}
				add(ruleConstList, position215)
			}
			return true
		l214:
			position, tokenIndex = position214, tokenIndex214
			return false
		},
		/* 36 ConstMap <- <(LWING (ConstValue COLON ConstValue ListSeparator?)* RWING)> */
		func() bool {
			position220, tokenIndex220 := position, tokenIndex
			{
				position221 := position
				if !_rules[ruleLWING]() {
					goto l220
				}
			l222:
				{
					position223, tokenIndex223 := position, tokenIndex
					if !_rules[ruleConstValue]() {
						goto l223
					}
					if !_rules[ruleCOLON]() {
						goto l223
					}
					if !_rules[ruleConstValue]() {
						goto l223
					}
					{
						position224, tokenIndex224 := position, tokenIndex
						if !_rules[ruleListSeparator]() {
							goto l224
						}
						goto l225
					l224:
						position, tokenIndex = position224, tokenIndex224
					}
				l225:
					goto l222
				l223:
					position, tokenIndex = position223, tokenIndex223
				}
				if !_rules[ruleRWING]() {
					goto l220
				}
				add(ruleConstMap, position221)
			}
			return true
		l220:
			position, tokenIndex = position220, tokenIndex220
			return false
		},
		/* 37 Literal <- <((('"' <(!'"' .)*> '"') / ('\'' <(!'\'' .)*> '\'')) Spacing)> */
		func() bool {
			position226, tokenIndex226 := position, tokenIndex
			{
				position227 := position
				{
					position228, tokenIndex228 := position, tokenIndex
					if buffer[position] != rune('"') {
						goto l229
					}
					position++
					{
						position230 := position
					l231:
						{
							position232, tokenIndex232 := position, tokenIndex
							{
								position233, tokenIndex233 := position, tokenIndex
								if buffer[position] != rune('"') {
									goto l233
								}
								position++
								goto l232
							l233:
								position, tokenIndex = position233, tokenIndex233
							}
							if !matchDot() {
								goto l232
							}
							goto l231
						l232:
							position, tokenIndex = position232, tokenIndex232
						}
						add(rulePegText, position230)
					}
					if buffer[position] != rune('"') {
						goto l229
					}
					position++
					goto l228
				l229:
					position, tokenIndex = position228, tokenIndex228
					if buffer[position] != rune('\'') {
						goto l226
					}
					position++
					{
						position234 := position
					l235:
						{
							position236, tokenIndex236 := position, tokenIndex
							{
								position237, tokenIndex237 := position, tokenIndex
								if buffer[position] != rune('\'') {
									goto l237
								}
								position++
								goto l236
							l237:
								position, tokenIndex = position237, tokenIndex237
							}
							if !matchDot() {
								goto l236
							}
							goto l235
						l236:
							position, tokenIndex = position236, tokenIndex236
						}
						add(rulePegText, position234)
					}
					if buffer[position] != rune('\'') {
						goto l226
					}
					position++
				}
			l228:
				if !_rules[ruleSpacing]() {
					goto l226
				}
				add(ruleLiteral, position227)
			}
			return true
		l226:
			position, tokenIndex = position226, tokenIndex226
			return false
		},
		/* 38 Identifier <- <(<((Letter / '_') (Letter / Digit / '.' / '_')*)> Spacing)> */
		func() bool {
			position238, tokenIndex238 := position, tokenIndex
			{
				position239 := position
				{
					position240 := position
					{
						position241, tokenIndex241 := position, tokenIndex
						if !_rules[ruleLetter]() {
							goto l242
						}
						goto l241
					l242:
						position, tokenIndex = position241, tokenIndex241
						if buffer[position] != rune('_') {
							goto l238
						}
						position++
					}
				l241:
				l243:
					{
						position244, tokenIndex244 := position, tokenIndex
						{
							position245, tokenIndex245 := position, tokenIndex
							if !_rules[ruleLetter]() {
								goto l246
							}
							goto l245
						l246:
							position, tokenIndex = position245, tokenIndex245
							if !_rules[ruleDigit]() {
								goto l247
							}
							goto l245
						l247:
							position, tokenIndex = position245, tokenIndex245
							if buffer[position] != rune('.') {
								goto l248
							}
							position++
							goto l245
						l248:
							position, tokenIndex = position245, tokenIndex245
							if buffer[position] != rune('_') {
								goto l244
							}
							position++
						}
					l245:
						goto l243
					l244:
						position, tokenIndex = position244, tokenIndex244
					}
					add(rulePegText, position240)
				}
				if !_rules[ruleSpacing]() {
					goto l238
				}
				add(ruleIdentifier, position239)
			}
			return true
		l238:
			position, tokenIndex = position238, tokenIndex238
			return false
		},
		/* 39 STIdentifier <- <(<((Letter / '_') (Letter / Digit / '.' / '_' / '-')*)> Spacing)> */
		func() bool {
			position249, tokenIndex249 := position, tokenIndex
			{
				position250 := position
				{
					position251 := position
					{
						position252, tokenIndex252 := position, tokenIndex
						if !_rules[ruleLetter]() {
							goto l253
						}
						goto l252
					l253:
						position, tokenIndex = position252, tokenIndex252
						if buffer[position] != rune('_') {
							goto l249
						}
						position++
					}
				l252:
				l254:
					{
						position255, tokenIndex255 := position, tokenIndex
						{
							position256, tokenIndex256 := position, tokenIndex
							if !_rules[ruleLetter]() {
								goto l257
							}
							goto l256
						l257:
							position, tokenIndex = position256, tokenIndex256
							if !_rules[ruleDigit]() {
								goto l258
							}
							goto l256
						l258:
							position, tokenIndex = position256, tokenIndex256
							if buffer[position] != rune('.') {
								goto l259
							}
							position++
							goto l256
						l259:
							position, tokenIndex = position256, tokenIndex256
							if buffer[position] != rune('_') {
								goto l260
							}
							position++
							goto l256
						l260:
							position, tokenIndex = position256, tokenIndex256
							if buffer[position] != rune('-') {
								goto l255
							}
							position++
						}
					l256:
						goto l254
					l255:
						position, tokenIndex = position255, tokenIndex255
					}
					add(rulePegText, position251)
				}
				if !_rules[ruleSpacing]() {
					goto l249
				}
				add(ruleSTIdentifier, position250)
			}
			return true
		l249:
			position, tokenIndex = position249, tokenIndex249
			return false
		},
		/* 40 ListSeparator <- <((',' / ';') Spacing)> */
		func() bool {
			position261, tokenIndex261 := position, tokenIndex
			{
				position262 := position
				{
					position263, tokenIndex263 := position, tokenIndex
					if buffer[position] != rune(',') {
						goto l264
					}
					position++
					goto l263
				l264:
					position, tokenIndex = position263, tokenIndex263
					if buffer[position] != rune(';') {
						goto l261
					}
					position++
				}
			l263:
				if !_rules[ruleSpacing]() {
					goto l261
				}
				add(ruleListSeparator, position262)
			}
			return true
		l261:
			position, tokenIndex = position261, tokenIndex261
			return false
		},
		/* 41 Letter <- <([a-z] / [A-Z])> */
		func() bool {
			position265, tokenIndex265 := position, tokenIndex
			{
				position266 := position
				{
					position267, tokenIndex267 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l268
					}
					position++
					goto l267
				l268:
					position, tokenIndex = position267, tokenIndex267
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l265
					}
					position++
				}
			l267:
				add(ruleLetter, position266)
			}
			return true
		l265:
			position, tokenIndex = position265, tokenIndex265
			return false
		},
		/* 42 Digit <- <[0-9]> */
		func() bool {
			position269, tokenIndex269 := position, tokenIndex
			{
				position270 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l269
				}
				position++
				add(ruleDigit, position270)
			}
			return true
		l269:
			position, tokenIndex = position269, tokenIndex269
			return false
		},
		/* 43 IdChars <- <([a-z] / [A-Z] / [0-9] / ('_' / '$'))> */
		func() bool {
			position271, tokenIndex271 := position, tokenIndex
			{
				position272 := position
				{
					position273, tokenIndex273 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l274
					}
					position++
					goto l273
				l274:
					position, tokenIndex = position273, tokenIndex273
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l275
					}
					position++
					goto l273
				l275:
					position, tokenIndex = position273, tokenIndex273
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l276
					}
					position++
					goto l273
				l276:
					position, tokenIndex = position273, tokenIndex273
					{
						position277, tokenIndex277 := position, tokenIndex
						if buffer[position] != rune('_') {
							goto l278
						}
						position++
						goto l277
					l278:
						position, tokenIndex = position277, tokenIndex277
						if buffer[position] != rune('$') {
							goto l271
						}
						position++
					}
				l277:
				}
			l273:
				add(ruleIdChars, position272)
			}
			return true
		l271:
			position, tokenIndex = position271, tokenIndex271
			return false
		},
		/* 44 Spacing <- <(Whitespace / LongComment / LineComment / Pragma)*> */
		func() bool {
			{
				position280 := position
			l281:
				{
					position282, tokenIndex282 := position, tokenIndex
					{
						position283, tokenIndex283 := position, tokenIndex
						if !_rules[ruleWhitespace]() {
							goto l284
						}
						goto l283
					l284:
						position, tokenIndex = position283, tokenIndex283
						if !_rules[ruleLongComment]() {
							goto l285
						}
						goto l283
					l285:
						position, tokenIndex = position283, tokenIndex283
						if !_rules[ruleLineComment]() {
							goto l286
						}
						goto l283
					l286:
						position, tokenIndex = position283, tokenIndex283
						if !_rules[rulePragma]() {
							goto l282
						}
					}
				l283:
					goto l281
				l282:
					position, tokenIndex = position282, tokenIndex282
				}
				add(ruleSpacing, position280)
			}
			return true
		},
		/* 45 Whitespace <- <(' ' / '\t' / '\r' / '\n')+> */
		func() bool {
			position287, tokenIndex287 := position, tokenIndex
			{
				position288 := position
				{
					position291, tokenIndex291 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l292
					}
					position++
					goto l291
				l292:
					position, tokenIndex = position291, tokenIndex291
					if buffer[position] != rune('\t') {
						goto l293
					}
					position++
					goto l291
				l293:
					position, tokenIndex = position291, tokenIndex291
					if buffer[position] != rune('\r') {
						goto l294
					}
					position++
					goto l291
				l294:
					position, tokenIndex = position291, tokenIndex291
					if buffer[position] != rune('\n') {
						goto l287
					}
					position++
				}
			l291:
			l289:
				{
					position290, tokenIndex290 := position, tokenIndex
					{
						position295, tokenIndex295 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l296
						}
						position++
						goto l295
					l296:
						position, tokenIndex = position295, tokenIndex295
						if buffer[position] != rune('\t') {
							goto l297
						}
						position++
						goto l295
					l297:
						position, tokenIndex = position295, tokenIndex295
						if buffer[position] != rune('\r') {
							goto l298
						}
						position++
						goto l295
					l298:
						position, tokenIndex = position295, tokenIndex295
						if buffer[position] != rune('\n') {
							goto l290
						}
						position++
					}
				l295:
					goto l289
				l290:
					position, tokenIndex = position290, tokenIndex290
				}
				add(ruleWhitespace, position288)
			}
			return true
		l287:
			position, tokenIndex = position287, tokenIndex287
			return false
		},
		/* 46 LongComment <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		func() bool {
			position299, tokenIndex299 := position, tokenIndex
			{
				position300 := position
				if buffer[position] != rune('/') {
					goto l299
				}
				position++
				if buffer[position] != rune('*') {
					goto l299
				}
				position++
			l301:
				{
					position302, tokenIndex302 := position, tokenIndex
					{
						position303, tokenIndex303 := position, tokenIndex
						if buffer[position] != rune('*') {
							goto l303
						}
						position++
						if buffer[position] != rune('/') {
							goto l303
						}
						position++
						goto l302
					l303:
						position, tokenIndex = position303, tokenIndex303
					}
					if !matchDot() {
						goto l302
					}
					goto l301
				l302:
					position, tokenIndex = position302, tokenIndex302
				}
				if buffer[position] != rune('*') {
					goto l299
				}
				position++
				if buffer[position] != rune('/') {
					goto l299
				}
				position++
				add(ruleLongComment, position300)
			}
			return true
		l299:
			position, tokenIndex = position299, tokenIndex299
			return false
		},
		/* 47 LineComment <- <('/' '/' (!('\r' / '\n') .)* ('\r' / '\n'))> */
		func() bool {
			position304, tokenIndex304 := position, tokenIndex
			{
				position305 := position
				if buffer[position] != rune('/') {
					goto l304
				}
				position++
				if buffer[position] != rune('/') {
					goto l304
				}
				position++
			l306:
				{
					position307, tokenIndex307 := position, tokenIndex
					{
						position308, tokenIndex308 := position, tokenIndex
						{
							position309, tokenIndex309 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l310
							}
							position++
							goto l309
						l310:
							position, tokenIndex = position309, tokenIndex309
							if buffer[position] != rune('\n') {
								goto l308
							}
							position++
						}
					l309:
						goto l307
					l308:
						position, tokenIndex = position308, tokenIndex308
					}
					if !matchDot() {
						goto l307
					}
					goto l306
				l307:
					position, tokenIndex = position307, tokenIndex307
				}
				{
					position311, tokenIndex311 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l312
					}
					position++
					goto l311
				l312:
					position, tokenIndex = position311, tokenIndex311
					if buffer[position] != rune('\n') {
						goto l304
					}
					position++
				}
			l311:
				add(ruleLineComment, position305)
			}
			return true
		l304:
			position, tokenIndex = position304, tokenIndex304
			return false
		},
		/* 48 Pragma <- <('#' (!('\r' / '\n') .)* ('\r' / '\n'))> */
		func() bool {
			position313, tokenIndex313 := position, tokenIndex
			{
				position314 := position
				if buffer[position] != rune('#') {
					goto l313
				}
				position++
			l315:
				{
					position316, tokenIndex316 := position, tokenIndex
					{
						position317, tokenIndex317 := position, tokenIndex
						{
							position318, tokenIndex318 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l319
							}
							position++
							goto l318
						l319:
							position, tokenIndex = position318, tokenIndex318
							if buffer[position] != rune('\n') {
								goto l317
							}
							position++
						}
					l318:
						goto l316
					l317:
						position, tokenIndex = position317, tokenIndex317
					}
					if !matchDot() {
						goto l316
					}
					goto l315
				l316:
					position, tokenIndex = position316, tokenIndex316
				}
				{
					position320, tokenIndex320 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l321
					}
					position++
					goto l320
				l321:
					position, tokenIndex = position320, tokenIndex320
					if buffer[position] != rune('\n') {
						goto l313
					}
					position++
				}
			l320:
				add(rulePragma, position314)
			}
			return true
		l313:
			position, tokenIndex = position313, tokenIndex313
			return false
		},
		/* 49 INCLUDE <- <('i' 'n' 'c' 'l' 'u' 'd' 'e' !IdChars Spacing)> */
		func() bool {
			position322, tokenIndex322 := position, tokenIndex
			{
				position323 := position
				if buffer[position] != rune('i') {
					goto l322
				}
				position++
				if buffer[position] != rune('n') {
					goto l322
				}
				position++
				if buffer[position] != rune('c') {
					goto l322
				}
				position++
				if buffer[position] != rune('l') {
					goto l322
				}
				position++
				if buffer[position] != rune('u') {
					goto l322
				}
				position++
				if buffer[position] != rune('d') {
					goto l322
				}
				position++
				if buffer[position] != rune('e') {
					goto l322
				}
				position++
				{
					position324, tokenIndex324 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l324
					}
					goto l322
				l324:
					position, tokenIndex = position324, tokenIndex324
				}
				if !_rules[ruleSpacing]() {
					goto l322
				}
				add(ruleINCLUDE, position323)
			}
			return true
		l322:
			position, tokenIndex = position322, tokenIndex322
			return false
		},
		/* 50 CPP_INCLUDE <- <('c' 'p' 'p' '_' 'i' 'n' 'c' 'l' 'u' 'd' 'e' !IdChars Spacing)> */
		func() bool {
			position325, tokenIndex325 := position, tokenIndex
			{
				position326 := position
				if buffer[position] != rune('c') {
					goto l325
				}
				position++
				if buffer[position] != rune('p') {
					goto l325
				}
				position++
				if buffer[position] != rune('p') {
					goto l325
				}
				position++
				if buffer[position] != rune('_') {
					goto l325
				}
				position++
				if buffer[position] != rune('i') {
					goto l325
				}
				position++
				if buffer[position] != rune('n') {
					goto l325
				}
				position++
				if buffer[position] != rune('c') {
					goto l325
				}
				position++
				if buffer[position] != rune('l') {
					goto l325
				}
				position++
				if buffer[position] != rune('u') {
					goto l325
				}
				position++
				if buffer[position] != rune('d') {
					goto l325
				}
				position++
				if buffer[position] != rune('e') {
					goto l325
				}
				position++
				{
					position327, tokenIndex327 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l327
					}
					goto l325
				l327:
					position, tokenIndex = position327, tokenIndex327
				}
				if !_rules[ruleSpacing]() {
					goto l325
				}
				add(ruleCPP_INCLUDE, position326)
			}
			return true
		l325:
			position, tokenIndex = position325, tokenIndex325
			return false
		},
		/* 51 NAMESPACE <- <('n' 'a' 'm' 'e' 's' 'p' 'a' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position328, tokenIndex328 := position, tokenIndex
			{
				position329 := position
				if buffer[position] != rune('n') {
					goto l328
				}
				position++
				if buffer[position] != rune('a') {
					goto l328
				}
				position++
				if buffer[position] != rune('m') {
					goto l328
				}
				position++
				if buffer[position] != rune('e') {
					goto l328
				}
				position++
				if buffer[position] != rune('s') {
					goto l328
				}
				position++
				if buffer[position] != rune('p') {
					goto l328
				}
				position++
				if buffer[position] != rune('a') {
					goto l328
				}
				position++
				if buffer[position] != rune('c') {
					goto l328
				}
				position++
				if buffer[position] != rune('e') {
					goto l328
				}
				position++
				{
					position330, tokenIndex330 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l330
					}
					goto l328
				l330:
					position, tokenIndex = position330, tokenIndex330
				}
				if !_rules[ruleSpacing]() {
					goto l328
				}
				add(ruleNAMESPACE, position329)
			}
			return true
		l328:
			position, tokenIndex = position328, tokenIndex328
			return false
		},
		/* 52 SMALLTALK_CATEGORY <- <('s' 'm' 'a' 'l' 'l' 't' 'a' 'l' 'k' '.' 'c' 'a' 't' 'e' 'g' 'o' 'r' 'y' !IdChars Spacing)> */
		func() bool {
			position331, tokenIndex331 := position, tokenIndex
			{
				position332 := position
				if buffer[position] != rune('s') {
					goto l331
				}
				position++
				if buffer[position] != rune('m') {
					goto l331
				}
				position++
				if buffer[position] != rune('a') {
					goto l331
				}
				position++
				if buffer[position] != rune('l') {
					goto l331
				}
				position++
				if buffer[position] != rune('l') {
					goto l331
				}
				position++
				if buffer[position] != rune('t') {
					goto l331
				}
				position++
				if buffer[position] != rune('a') {
					goto l331
				}
				position++
				if buffer[position] != rune('l') {
					goto l331
				}
				position++
				if buffer[position] != rune('k') {
					goto l331
				}
				position++
				if buffer[position] != rune('.') {
					goto l331
				}
				position++
				if buffer[position] != rune('c') {
					goto l331
				}
				position++
				if buffer[position] != rune('a') {
					goto l331
				}
				position++
				if buffer[position] != rune('t') {
					goto l331
				}
				position++
				if buffer[position] != rune('e') {
					goto l331
				}
				position++
				if buffer[position] != rune('g') {
					goto l331
				}
				position++
				if buffer[position] != rune('o') {
					goto l331
				}
				position++
				if buffer[position] != rune('r') {
					goto l331
				}
				position++
				if buffer[position] != rune('y') {
					goto l331
				}
				position++
				{
					position333, tokenIndex333 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l333
					}
					goto l331
				l333:
					position, tokenIndex = position333, tokenIndex333
				}
				if !_rules[ruleSpacing]() {
					goto l331
				}
				add(ruleSMALLTALK_CATEGORY, position332)
			}
			return true
		l331:
			position, tokenIndex = position331, tokenIndex331
			return false
		},
		/* 53 SMALLTALK_PREFIX <- <('s' 'm' 'a' 'l' 'l' 't' 'a' 'l' 'k' '.' 'p' 'r' 'e' 'f' 'i' 'x' !IdChars Spacing)> */
		func() bool {
			position334, tokenIndex334 := position, tokenIndex
			{
				position335 := position
				if buffer[position] != rune('s') {
					goto l334
				}
				position++
				if buffer[position] != rune('m') {
					goto l334
				}
				position++
				if buffer[position] != rune('a') {
					goto l334
				}
				position++
				if buffer[position] != rune('l') {
					goto l334
				}
				position++
				if buffer[position] != rune('l') {
					goto l334
				}
				position++
				if buffer[position] != rune('t') {
					goto l334
				}
				position++
				if buffer[position] != rune('a') {
					goto l334
				}
				position++
				if buffer[position] != rune('l') {
					goto l334
				}
				position++
				if buffer[position] != rune('k') {
					goto l334
				}
				position++
				if buffer[position] != rune('.') {
					goto l334
				}
				position++
				if buffer[position] != rune('p') {
					goto l334
				}
				position++
				if buffer[position] != rune('r') {
					goto l334
				}
				position++
				if buffer[position] != rune('e') {
					goto l334
				}
				position++
				if buffer[position] != rune('f') {
					goto l334
				}
				position++
				if buffer[position] != rune('i') {
					goto l334
				}
				position++
				if buffer[position] != rune('x') {
					goto l334
				}
				position++
				{
					position336, tokenIndex336 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l336
					}
					goto l334
				l336:
					position, tokenIndex = position336, tokenIndex336
				}
				if !_rules[ruleSpacing]() {
					goto l334
				}
				add(ruleSMALLTALK_PREFIX, position335)
			}
			return true
		l334:
			position, tokenIndex = position334, tokenIndex334
			return false
		},
		/* 54 PHP_NAMESPACE <- <('p' 'h' 'p' '_' 'n' 'a' 'm' 'e' 's' 'p' 'a' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position337, tokenIndex337 := position, tokenIndex
			{
				position338 := position
				if buffer[position] != rune('p') {
					goto l337
				}
				position++
				if buffer[position] != rune('h') {
					goto l337
				}
				position++
				if buffer[position] != rune('p') {
					goto l337
				}
				position++
				if buffer[position] != rune('_') {
					goto l337
				}
				position++
				if buffer[position] != rune('n') {
					goto l337
				}
				position++
				if buffer[position] != rune('a') {
					goto l337
				}
				position++
				if buffer[position] != rune('m') {
					goto l337
				}
				position++
				if buffer[position] != rune('e') {
					goto l337
				}
				position++
				if buffer[position] != rune('s') {
					goto l337
				}
				position++
				if buffer[position] != rune('p') {
					goto l337
				}
				position++
				if buffer[position] != rune('a') {
					goto l337
				}
				position++
				if buffer[position] != rune('c') {
					goto l337
				}
				position++
				if buffer[position] != rune('e') {
					goto l337
				}
				position++
				{
					position339, tokenIndex339 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l339
					}
					goto l337
				l339:
					position, tokenIndex = position339, tokenIndex339
				}
				if !_rules[ruleSpacing]() {
					goto l337
				}
				add(rulePHP_NAMESPACE, position338)
			}
			return true
		l337:
			position, tokenIndex = position337, tokenIndex337
			return false
		},
		/* 55 XSD_NAMESPACE <- <('x' 's' 'd' '_' 'n' 'a' 'm' 'e' 's' 'p' 'a' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position340, tokenIndex340 := position, tokenIndex
			{
				position341 := position
				if buffer[position] != rune('x') {
					goto l340
				}
				position++
				if buffer[position] != rune('s') {
					goto l340
				}
				position++
				if buffer[position] != rune('d') {
					goto l340
				}
				position++
				if buffer[position] != rune('_') {
					goto l340
				}
				position++
				if buffer[position] != rune('n') {
					goto l340
				}
				position++
				if buffer[position] != rune('a') {
					goto l340
				}
				position++
				if buffer[position] != rune('m') {
					goto l340
				}
				position++
				if buffer[position] != rune('e') {
					goto l340
				}
				position++
				if buffer[position] != rune('s') {
					goto l340
				}
				position++
				if buffer[position] != rune('p') {
					goto l340
				}
				position++
				if buffer[position] != rune('a') {
					goto l340
				}
				position++
				if buffer[position] != rune('c') {
					goto l340
				}
				position++
				if buffer[position] != rune('e') {
					goto l340
				}
				position++
				{
					position342, tokenIndex342 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l342
					}
					goto l340
				l342:
					position, tokenIndex = position342, tokenIndex342
				}
				if !_rules[ruleSpacing]() {
					goto l340
				}
				add(ruleXSD_NAMESPACE, position341)
			}
			return true
		l340:
			position, tokenIndex = position340, tokenIndex340
			return false
		},
		/* 56 CONST <- <('c' 'o' 'n' 's' 't' !IdChars Spacing)> */
		func() bool {
			position343, tokenIndex343 := position, tokenIndex
			{
				position344 := position
				if buffer[position] != rune('c') {
					goto l343
				}
				position++
				if buffer[position] != rune('o') {
					goto l343
				}
				position++
				if buffer[position] != rune('n') {
					goto l343
				}
				position++
				if buffer[position] != rune('s') {
					goto l343
				}
				position++
				if buffer[position] != rune('t') {
					goto l343
				}
				position++
				{
					position345, tokenIndex345 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l345
					}
					goto l343
				l345:
					position, tokenIndex = position345, tokenIndex345
				}
				if !_rules[ruleSpacing]() {
					goto l343
				}
				add(ruleCONST, position344)
			}
			return true
		l343:
			position, tokenIndex = position343, tokenIndex343
			return false
		},
		/* 57 TYPEDEF <- <('t' 'y' 'p' 'e' 'd' 'e' 'f' !IdChars Spacing)> */
		func() bool {
			position346, tokenIndex346 := position, tokenIndex
			{
				position347 := position
				if buffer[position] != rune('t') {
					goto l346
				}
				position++
				if buffer[position] != rune('y') {
					goto l346
				}
				position++
				if buffer[position] != rune('p') {
					goto l346
				}
				position++
				if buffer[position] != rune('e') {
					goto l346
				}
				position++
				if buffer[position] != rune('d') {
					goto l346
				}
				position++
				if buffer[position] != rune('e') {
					goto l346
				}
				position++
				if buffer[position] != rune('f') {
					goto l346
				}
				position++
				{
					position348, tokenIndex348 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l348
					}
					goto l346
				l348:
					position, tokenIndex = position348, tokenIndex348
				}
				if !_rules[ruleSpacing]() {
					goto l346
				}
				add(ruleTYPEDEF, position347)
			}
			return true
		l346:
			position, tokenIndex = position346, tokenIndex346
			return false
		},
		/* 58 ENUM <- <('e' 'n' 'u' 'm' !IdChars Spacing)> */
		func() bool {
			position349, tokenIndex349 := position, tokenIndex
			{
				position350 := position
				if buffer[position] != rune('e') {
					goto l349
				}
				position++
				if buffer[position] != rune('n') {
					goto l349
				}
				position++
				if buffer[position] != rune('u') {
					goto l349
				}
				position++
				if buffer[position] != rune('m') {
					goto l349
				}
				position++
				{
					position351, tokenIndex351 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l351
					}
					goto l349
				l351:
					position, tokenIndex = position351, tokenIndex351
				}
				if !_rules[ruleSpacing]() {
					goto l349
				}
				add(ruleENUM, position350)
			}
			return true
		l349:
			position, tokenIndex = position349, tokenIndex349
			return false
		},
		/* 59 SENUM <- <('s' 'e' 'n' 'u' 'm' !IdChars Spacing)> */
		func() bool {
			position352, tokenIndex352 := position, tokenIndex
			{
				position353 := position
				if buffer[position] != rune('s') {
					goto l352
				}
				position++
				if buffer[position] != rune('e') {
					goto l352
				}
				position++
				if buffer[position] != rune('n') {
					goto l352
				}
				position++
				if buffer[position] != rune('u') {
					goto l352
				}
				position++
				if buffer[position] != rune('m') {
					goto l352
				}
				position++
				{
					position354, tokenIndex354 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l354
					}
					goto l352
				l354:
					position, tokenIndex = position354, tokenIndex354
				}
				if !_rules[ruleSpacing]() {
					goto l352
				}
				add(ruleSENUM, position353)
			}
			return true
		l352:
			position, tokenIndex = position352, tokenIndex352
			return false
		},
		/* 60 STRUCT <- <('s' 't' 'r' 'u' 'c' 't' !IdChars Spacing)> */
		func() bool {
			position355, tokenIndex355 := position, tokenIndex
			{
				position356 := position
				if buffer[position] != rune('s') {
					goto l355
				}
				position++
				if buffer[position] != rune('t') {
					goto l355
				}
				position++
				if buffer[position] != rune('r') {
					goto l355
				}
				position++
				if buffer[position] != rune('u') {
					goto l355
				}
				position++
				if buffer[position] != rune('c') {
					goto l355
				}
				position++
				if buffer[position] != rune('t') {
					goto l355
				}
				position++
				{
					position357, tokenIndex357 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l357
					}
					goto l355
				l357:
					position, tokenIndex = position357, tokenIndex357
				}
				if !_rules[ruleSpacing]() {
					goto l355
				}
				add(ruleSTRUCT, position356)
			}
			return true
		l355:
			position, tokenIndex = position355, tokenIndex355
			return false
		},
		/* 61 UNION <- <('u' 'n' 'i' 'o' 'n' !IdChars Spacing)> */
		func() bool {
			position358, tokenIndex358 := position, tokenIndex
			{
				position359 := position
				if buffer[position] != rune('u') {
					goto l358
				}
				position++
				if buffer[position] != rune('n') {
					goto l358
				}
				position++
				if buffer[position] != rune('i') {
					goto l358
				}
				position++
				if buffer[position] != rune('o') {
					goto l358
				}
				position++
				if buffer[position] != rune('n') {
					goto l358
				}
				position++
				{
					position360, tokenIndex360 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l360
					}
					goto l358
				l360:
					position, tokenIndex = position360, tokenIndex360
				}
				if !_rules[ruleSpacing]() {
					goto l358
				}
				add(ruleUNION, position359)
			}
			return true
		l358:
			position, tokenIndex = position358, tokenIndex358
			return false
		},
		/* 62 SERVICE <- <('s' 'e' 'r' 'v' 'i' 'c' 'e' !IdChars Spacing)> */
		func() bool {
			position361, tokenIndex361 := position, tokenIndex
			{
				position362 := position
				if buffer[position] != rune('s') {
					goto l361
				}
				position++
				if buffer[position] != rune('e') {
					goto l361
				}
				position++
				if buffer[position] != rune('r') {
					goto l361
				}
				position++
				if buffer[position] != rune('v') {
					goto l361
				}
				position++
				if buffer[position] != rune('i') {
					goto l361
				}
				position++
				if buffer[position] != rune('c') {
					goto l361
				}
				position++
				if buffer[position] != rune('e') {
					goto l361
				}
				position++
				{
					position363, tokenIndex363 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l363
					}
					goto l361
				l363:
					position, tokenIndex = position363, tokenIndex363
				}
				if !_rules[ruleSpacing]() {
					goto l361
				}
				add(ruleSERVICE, position362)
			}
			return true
		l361:
			position, tokenIndex = position361, tokenIndex361
			return false
		},
		/* 63 EXTENDS <- <('e' 'x' 't' 'e' 'n' 'd' 's' !IdChars Spacing)> */
		func() bool {
			position364, tokenIndex364 := position, tokenIndex
			{
				position365 := position
				if buffer[position] != rune('e') {
					goto l364
				}
				position++
				if buffer[position] != rune('x') {
					goto l364
				}
				position++
				if buffer[position] != rune('t') {
					goto l364
				}
				position++
				if buffer[position] != rune('e') {
					goto l364
				}
				position++
				if buffer[position] != rune('n') {
					goto l364
				}
				position++
				if buffer[position] != rune('d') {
					goto l364
				}
				position++
				if buffer[position] != rune('s') {
					goto l364
				}
				position++
				{
					position366, tokenIndex366 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l366
					}
					goto l364
				l366:
					position, tokenIndex = position366, tokenIndex366
				}
				if !_rules[ruleSpacing]() {
					goto l364
				}
				add(ruleEXTENDS, position365)
			}
			return true
		l364:
			position, tokenIndex = position364, tokenIndex364
			return false
		},
		/* 64 EXCEPTION <- <('e' 'x' 'c' 'e' 'p' 't' 'i' 'o' 'n' !IdChars Spacing)> */
		func() bool {
			position367, tokenIndex367 := position, tokenIndex
			{
				position368 := position
				if buffer[position] != rune('e') {
					goto l367
				}
				position++
				if buffer[position] != rune('x') {
					goto l367
				}
				position++
				if buffer[position] != rune('c') {
					goto l367
				}
				position++
				if buffer[position] != rune('e') {
					goto l367
				}
				position++
				if buffer[position] != rune('p') {
					goto l367
				}
				position++
				if buffer[position] != rune('t') {
					goto l367
				}
				position++
				if buffer[position] != rune('i') {
					goto l367
				}
				position++
				if buffer[position] != rune('o') {
					goto l367
				}
				position++
				if buffer[position] != rune('n') {
					goto l367
				}
				position++
				{
					position369, tokenIndex369 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l369
					}
					goto l367
				l369:
					position, tokenIndex = position369, tokenIndex369
				}
				if !_rules[ruleSpacing]() {
					goto l367
				}
				add(ruleEXCEPTION, position368)
			}
			return true
		l367:
			position, tokenIndex = position367, tokenIndex367
			return false
		},
		/* 65 ONEWAY <- <('o' 'n' 'e' 'w' 'a' 'y' !IdChars Spacing)> */
		func() bool {
			position370, tokenIndex370 := position, tokenIndex
			{
				position371 := position
				if buffer[position] != rune('o') {
					goto l370
				}
				position++
				if buffer[position] != rune('n') {
					goto l370
				}
				position++
				if buffer[position] != rune('e') {
					goto l370
				}
				position++
				if buffer[position] != rune('w') {
					goto l370
				}
				position++
				if buffer[position] != rune('a') {
					goto l370
				}
				position++
				if buffer[position] != rune('y') {
					goto l370
				}
				position++
				{
					position372, tokenIndex372 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l372
					}
					goto l370
				l372:
					position, tokenIndex = position372, tokenIndex372
				}
				if !_rules[ruleSpacing]() {
					goto l370
				}
				add(ruleONEWAY, position371)
			}
			return true
		l370:
			position, tokenIndex = position370, tokenIndex370
			return false
		},
		/* 66 THROWS <- <('t' 'h' 'r' 'o' 'w' 's' !IdChars Spacing)> */
		func() bool {
			position373, tokenIndex373 := position, tokenIndex
			{
				position374 := position
				if buffer[position] != rune('t') {
					goto l373
				}
				position++
				if buffer[position] != rune('h') {
					goto l373
				}
				position++
				if buffer[position] != rune('r') {
					goto l373
				}
				position++
				if buffer[position] != rune('o') {
					goto l373
				}
				position++
				if buffer[position] != rune('w') {
					goto l373
				}
				position++
				if buffer[position] != rune('s') {
					goto l373
				}
				position++
				{
					position375, tokenIndex375 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l375
					}
					goto l373
				l375:
					position, tokenIndex = position375, tokenIndex375
				}
				if !_rules[ruleSpacing]() {
					goto l373
				}
				add(ruleTHROWS, position374)
			}
			return true
		l373:
			position, tokenIndex = position373, tokenIndex373
			return false
		},
		/* 67 CPP_TYPE <- <('c' 'p' 'p' '_' 't' 'y' 'p' 'e' !IdChars Spacing)> */
		func() bool {
			position376, tokenIndex376 := position, tokenIndex
			{
				position377 := position
				if buffer[position] != rune('c') {
					goto l376
				}
				position++
				if buffer[position] != rune('p') {
					goto l376
				}
				position++
				if buffer[position] != rune('p') {
					goto l376
				}
				position++
				if buffer[position] != rune('_') {
					goto l376
				}
				position++
				if buffer[position] != rune('t') {
					goto l376
				}
				position++
				if buffer[position] != rune('y') {
					goto l376
				}
				position++
				if buffer[position] != rune('p') {
					goto l376
				}
				position++
				if buffer[position] != rune('e') {
					goto l376
				}
				position++
				{
					position378, tokenIndex378 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l378
					}
					goto l376
				l378:
					position, tokenIndex = position378, tokenIndex378
				}
				if !_rules[ruleSpacing]() {
					goto l376
				}
				add(ruleCPP_TYPE, position377)
			}
			return true
		l376:
			position, tokenIndex = position376, tokenIndex376
			return false
		},
		/* 68 XSD_ALL <- <('x' 's' 'd' '_' 'a' 'l' 'l' !IdChars Spacing)> */
		func() bool {
			position379, tokenIndex379 := position, tokenIndex
			{
				position380 := position
				if buffer[position] != rune('x') {
					goto l379
				}
				position++
				if buffer[position] != rune('s') {
					goto l379
				}
				position++
				if buffer[position] != rune('d') {
					goto l379
				}
				position++
				if buffer[position] != rune('_') {
					goto l379
				}
				position++
				if buffer[position] != rune('a') {
					goto l379
				}
				position++
				if buffer[position] != rune('l') {
					goto l379
				}
				position++
				if buffer[position] != rune('l') {
					goto l379
				}
				position++
				{
					position381, tokenIndex381 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l381
					}
					goto l379
				l381:
					position, tokenIndex = position381, tokenIndex381
				}
				if !_rules[ruleSpacing]() {
					goto l379
				}
				add(ruleXSD_ALL, position380)
			}
			return true
		l379:
			position, tokenIndex = position379, tokenIndex379
			return false
		},
		/* 69 XSD_OPTIONAL <- <('x' 's' 'd' '_' 'o' 'p' 't' 'i' 'o' 'n' 'a' 'l' !IdChars Spacing)> */
		func() bool {
			position382, tokenIndex382 := position, tokenIndex
			{
				position383 := position
				if buffer[position] != rune('x') {
					goto l382
				}
				position++
				if buffer[position] != rune('s') {
					goto l382
				}
				position++
				if buffer[position] != rune('d') {
					goto l382
				}
				position++
				if buffer[position] != rune('_') {
					goto l382
				}
				position++
				if buffer[position] != rune('o') {
					goto l382
				}
				position++
				if buffer[position] != rune('p') {
					goto l382
				}
				position++
				if buffer[position] != rune('t') {
					goto l382
				}
				position++
				if buffer[position] != rune('i') {
					goto l382
				}
				position++
				if buffer[position] != rune('o') {
					goto l382
				}
				position++
				if buffer[position] != rune('n') {
					goto l382
				}
				position++
				if buffer[position] != rune('a') {
					goto l382
				}
				position++
				if buffer[position] != rune('l') {
					goto l382
				}
				position++
				{
					position384, tokenIndex384 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l384
					}
					goto l382
				l384:
					position, tokenIndex = position384, tokenIndex384
				}
				if !_rules[ruleSpacing]() {
					goto l382
				}
				add(ruleXSD_OPTIONAL, position383)
			}
			return true
		l382:
			position, tokenIndex = position382, tokenIndex382
			return false
		},
		/* 70 XSD_NILLABLE <- <('x' 's' 'd' '_' 'n' 'i' 'l' 'l' 'a' 'b' 'l' 'e' !IdChars Spacing)> */
		func() bool {
			position385, tokenIndex385 := position, tokenIndex
			{
				position386 := position
				if buffer[position] != rune('x') {
					goto l385
				}
				position++
				if buffer[position] != rune('s') {
					goto l385
				}
				position++
				if buffer[position] != rune('d') {
					goto l385
				}
				position++
				if buffer[position] != rune('_') {
					goto l385
				}
				position++
				if buffer[position] != rune('n') {
					goto l385
				}
				position++
				if buffer[position] != rune('i') {
					goto l385
				}
				position++
				if buffer[position] != rune('l') {
					goto l385
				}
				position++
				if buffer[position] != rune('l') {
					goto l385
				}
				position++
				if buffer[position] != rune('a') {
					goto l385
				}
				position++
				if buffer[position] != rune('b') {
					goto l385
				}
				position++
				if buffer[position] != rune('l') {
					goto l385
				}
				position++
				if buffer[position] != rune('e') {
					goto l385
				}
				position++
				{
					position387, tokenIndex387 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l387
					}
					goto l385
				l387:
					position, tokenIndex = position387, tokenIndex387
				}
				if !_rules[ruleSpacing]() {
					goto l385
				}
				add(ruleXSD_NILLABLE, position386)
			}
			return true
		l385:
			position, tokenIndex = position385, tokenIndex385
			return false
		},
		/* 71 XSD_ATTRS <- <('x' 's' 'd' '_' 'a' 't' 't' 'r' 's' !IdChars Spacing)> */
		func() bool {
			position388, tokenIndex388 := position, tokenIndex
			{
				position389 := position
				if buffer[position] != rune('x') {
					goto l388
				}
				position++
				if buffer[position] != rune('s') {
					goto l388
				}
				position++
				if buffer[position] != rune('d') {
					goto l388
				}
				position++
				if buffer[position] != rune('_') {
					goto l388
				}
				position++
				if buffer[position] != rune('a') {
					goto l388
				}
				position++
				if buffer[position] != rune('t') {
					goto l388
				}
				position++
				if buffer[position] != rune('t') {
					goto l388
				}
				position++
				if buffer[position] != rune('r') {
					goto l388
				}
				position++
				if buffer[position] != rune('s') {
					goto l388
				}
				position++
				{
					position390, tokenIndex390 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l390
					}
					goto l388
				l390:
					position, tokenIndex = position390, tokenIndex390
				}
				if !_rules[ruleSpacing]() {
					goto l388
				}
				add(ruleXSD_ATTRS, position389)
			}
			return true
		l388:
			position, tokenIndex = position388, tokenIndex388
			return false
		},
		/* 72 VOID <- <('v' 'o' 'i' 'd' !IdChars Spacing)> */
		func() bool {
			position391, tokenIndex391 := position, tokenIndex
			{
				position392 := position
				if buffer[position] != rune('v') {
					goto l391
				}
				position++
				if buffer[position] != rune('o') {
					goto l391
				}
				position++
				if buffer[position] != rune('i') {
					goto l391
				}
				position++
				if buffer[position] != rune('d') {
					goto l391
				}
				position++
				{
					position393, tokenIndex393 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l393
					}
					goto l391
				l393:
					position, tokenIndex = position393, tokenIndex393
				}
				if !_rules[ruleSpacing]() {
					goto l391
				}
				add(ruleVOID, position392)
			}
			return true
		l391:
			position, tokenIndex = position391, tokenIndex391
			return false
		},
		/* 73 MAP <- <('m' 'a' 'p' !IdChars Spacing)> */
		func() bool {
			position394, tokenIndex394 := position, tokenIndex
			{
				position395 := position
				if buffer[position] != rune('m') {
					goto l394
				}
				position++
				if buffer[position] != rune('a') {
					goto l394
				}
				position++
				if buffer[position] != rune('p') {
					goto l394
				}
				position++
				{
					position396, tokenIndex396 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l396
					}
					goto l394
				l396:
					position, tokenIndex = position396, tokenIndex396
				}
				if !_rules[ruleSpacing]() {
					goto l394
				}
				add(ruleMAP, position395)
			}
			return true
		l394:
			position, tokenIndex = position394, tokenIndex394
			return false
		},
		/* 74 SET <- <('s' 'e' 't' !IdChars Spacing)> */
		func() bool {
			position397, tokenIndex397 := position, tokenIndex
			{
				position398 := position
				if buffer[position] != rune('s') {
					goto l397
				}
				position++
				if buffer[position] != rune('e') {
					goto l397
				}
				position++
				if buffer[position] != rune('t') {
					goto l397
				}
				position++
				{
					position399, tokenIndex399 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l399
					}
					goto l397
				l399:
					position, tokenIndex = position399, tokenIndex399
				}
				if !_rules[ruleSpacing]() {
					goto l397
				}
				add(ruleSET, position398)
			}
			return true
		l397:
			position, tokenIndex = position397, tokenIndex397
			return false
		},
		/* 75 LIST <- <('l' 'i' 's' 't' !IdChars Spacing)> */
		func() bool {
			position400, tokenIndex400 := position, tokenIndex
			{
				position401 := position
				if buffer[position] != rune('l') {
					goto l400
				}
				position++
				if buffer[position] != rune('i') {
					goto l400
				}
				position++
				if buffer[position] != rune('s') {
					goto l400
				}
				position++
				if buffer[position] != rune('t') {
					goto l400
				}
				position++
				{
					position402, tokenIndex402 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l402
					}
					goto l400
				l402:
					position, tokenIndex = position402, tokenIndex402
				}
				if !_rules[ruleSpacing]() {
					goto l400
				}
				add(ruleLIST, position401)
			}
			return true
		l400:
			position, tokenIndex = position400, tokenIndex400
			return false
		},
		/* 76 BOOL <- <(<('b' 'o' 'o' 'l')> !IdChars Spacing)> */
		func() bool {
			position403, tokenIndex403 := position, tokenIndex
			{
				position404 := position
				{
					position405 := position
					if buffer[position] != rune('b') {
						goto l403
					}
					position++
					if buffer[position] != rune('o') {
						goto l403
					}
					position++
					if buffer[position] != rune('o') {
						goto l403
					}
					position++
					if buffer[position] != rune('l') {
						goto l403
					}
					position++
					add(rulePegText, position405)
				}
				{
					position406, tokenIndex406 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l406
					}
					goto l403
				l406:
					position, tokenIndex = position406, tokenIndex406
				}
				if !_rules[ruleSpacing]() {
					goto l403
				}
				add(ruleBOOL, position404)
			}
			return true
		l403:
			position, tokenIndex = position403, tokenIndex403
			return false
		},
		/* 77 BYTE <- <(<('b' 'y' 't' 'e')> !IdChars Spacing)> */
		func() bool {
			position407, tokenIndex407 := position, tokenIndex
			{
				position408 := position
				{
					position409 := position
					if buffer[position] != rune('b') {
						goto l407
					}
					position++
					if buffer[position] != rune('y') {
						goto l407
					}
					position++
					if buffer[position] != rune('t') {
						goto l407
					}
					position++
					if buffer[position] != rune('e') {
						goto l407
					}
					position++
					add(rulePegText, position409)
				}
				{
					position410, tokenIndex410 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l410
					}
					goto l407
				l410:
					position, tokenIndex = position410, tokenIndex410
				}
				if !_rules[ruleSpacing]() {
					goto l407
				}
				add(ruleBYTE, position408)
			}
			return true
		l407:
			position, tokenIndex = position407, tokenIndex407
			return false
		},
		/* 78 I8 <- <(<('i' '8')> !IdChars Spacing)> */
		func() bool {
			position411, tokenIndex411 := position, tokenIndex
			{
				position412 := position
				{
					position413 := position
					if buffer[position] != rune('i') {
						goto l411
					}
					position++
					if buffer[position] != rune('8') {
						goto l411
					}
					position++
					add(rulePegText, position413)
				}
				{
					position414, tokenIndex414 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l414
					}
					goto l411
				l414:
					position, tokenIndex = position414, tokenIndex414
				}
				if !_rules[ruleSpacing]() {
					goto l411
				}
				add(ruleI8, position412)
			}
			return true
		l411:
			position, tokenIndex = position411, tokenIndex411
			return false
		},
		/* 79 I16 <- <(<('i' '1' '6')> !IdChars Spacing)> */
		func() bool {
			position415, tokenIndex415 := position, tokenIndex
			{
				position416 := position
				{
					position417 := position
					if buffer[position] != rune('i') {
						goto l415
					}
					position++
					if buffer[position] != rune('1') {
						goto l415
					}
					position++
					if buffer[position] != rune('6') {
						goto l415
					}
					position++
					add(rulePegText, position417)
				}
				{
					position418, tokenIndex418 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l418
					}
					goto l415
				l418:
					position, tokenIndex = position418, tokenIndex418
				}
				if !_rules[ruleSpacing]() {
					goto l415
				}
				add(ruleI16, position416)
			}
			return true
		l415:
			position, tokenIndex = position415, tokenIndex415
			return false
		},
		/* 80 I32 <- <(<('i' '3' '2')> !IdChars Spacing)> */
		func() bool {
			position419, tokenIndex419 := position, tokenIndex
			{
				position420 := position
				{
					position421 := position
					if buffer[position] != rune('i') {
						goto l419
					}
					position++
					if buffer[position] != rune('3') {
						goto l419
					}
					position++
					if buffer[position] != rune('2') {
						goto l419
					}
					position++
					add(rulePegText, position421)
				}
				{
					position422, tokenIndex422 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l422
					}
					goto l419
				l422:
					position, tokenIndex = position422, tokenIndex422
				}
				if !_rules[ruleSpacing]() {
					goto l419
				}
				add(ruleI32, position420)
			}
			return true
		l419:
			position, tokenIndex = position419, tokenIndex419
			return false
		},
		/* 81 I64 <- <(<('i' '6' '4')> !IdChars Spacing)> */
		func() bool {
			position423, tokenIndex423 := position, tokenIndex
			{
				position424 := position
				{
					position425 := position
					if buffer[position] != rune('i') {
						goto l423
					}
					position++
					if buffer[position] != rune('6') {
						goto l423
					}
					position++
					if buffer[position] != rune('4') {
						goto l423
					}
					position++
					add(rulePegText, position425)
				}
				{
					position426, tokenIndex426 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l426
					}
					goto l423
				l426:
					position, tokenIndex = position426, tokenIndex426
				}
				if !_rules[ruleSpacing]() {
					goto l423
				}
				add(ruleI64, position424)
			}
			return true
		l423:
			position, tokenIndex = position423, tokenIndex423
			return false
		},
		/* 82 DOUBLE <- <(<('d' 'o' 'u' 'b' 'l' 'e')> !IdChars Spacing)> */
		func() bool {
			position427, tokenIndex427 := position, tokenIndex
			{
				position428 := position
				{
					position429 := position
					if buffer[position] != rune('d') {
						goto l427
					}
					position++
					if buffer[position] != rune('o') {
						goto l427
					}
					position++
					if buffer[position] != rune('u') {
						goto l427
					}
					position++
					if buffer[position] != rune('b') {
						goto l427
					}
					position++
					if buffer[position] != rune('l') {
						goto l427
					}
					position++
					if buffer[position] != rune('e') {
						goto l427
					}
					position++
					add(rulePegText, position429)
				}
				{
					position430, tokenIndex430 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l430
					}
					goto l427
				l430:
					position, tokenIndex = position430, tokenIndex430
				}
				if !_rules[ruleSpacing]() {
					goto l427
				}
				add(ruleDOUBLE, position428)
			}
			return true
		l427:
			position, tokenIndex = position427, tokenIndex427
			return false
		},
		/* 83 STRING <- <(<('s' 't' 'r' 'i' 'n' 'g')> !IdChars Spacing)> */
		func() bool {
			position431, tokenIndex431 := position, tokenIndex
			{
				position432 := position
				{
					position433 := position
					if buffer[position] != rune('s') {
						goto l431
					}
					position++
					if buffer[position] != rune('t') {
						goto l431
					}
					position++
					if buffer[position] != rune('r') {
						goto l431
					}
					position++
					if buffer[position] != rune('i') {
						goto l431
					}
					position++
					if buffer[position] != rune('n') {
						goto l431
					}
					position++
					if buffer[position] != rune('g') {
						goto l431
					}
					position++
					add(rulePegText, position433)
				}
				{
					position434, tokenIndex434 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l434
					}
					goto l431
				l434:
					position, tokenIndex = position434, tokenIndex434
				}
				if !_rules[ruleSpacing]() {
					goto l431
				}
				add(ruleSTRING, position432)
			}
			return true
		l431:
			position, tokenIndex = position431, tokenIndex431
			return false
		},
		/* 84 BINARY <- <(<('b' 'i' 'n' 'a' 'r' 'y')> !IdChars Spacing)> */
		func() bool {
			position435, tokenIndex435 := position, tokenIndex
			{
				position436 := position
				{
					position437 := position
					if buffer[position] != rune('b') {
						goto l435
					}
					position++
					if buffer[position] != rune('i') {
						goto l435
					}
					position++
					if buffer[position] != rune('n') {
						goto l435
					}
					position++
					if buffer[position] != rune('a') {
						goto l435
					}
					position++
					if buffer[position] != rune('r') {
						goto l435
					}
					position++
					if buffer[position] != rune('y') {
						goto l435
					}
					position++
					add(rulePegText, position437)
				}
				{
					position438, tokenIndex438 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l438
					}
					goto l435
				l438:
					position, tokenIndex = position438, tokenIndex438
				}
				if !_rules[ruleSpacing]() {
					goto l435
				}
				add(ruleBINARY, position436)
			}
			return true
		l435:
			position, tokenIndex = position435, tokenIndex435
			return false
		},
		/* 85 SLIST <- <(<('s' 'l' 'i' 's' 't')> !IdChars Spacing)> */
		func() bool {
			position439, tokenIndex439 := position, tokenIndex
			{
				position440 := position
				{
					position441 := position
					if buffer[position] != rune('s') {
						goto l439
					}
					position++
					if buffer[position] != rune('l') {
						goto l439
					}
					position++
					if buffer[position] != rune('i') {
						goto l439
					}
					position++
					if buffer[position] != rune('s') {
						goto l439
					}
					position++
					if buffer[position] != rune('t') {
						goto l439
					}
					position++
					add(rulePegText, position441)
				}
				{
					position442, tokenIndex442 := position, tokenIndex
					if !_rules[ruleIdChars]() {
						goto l442
					}
					goto l439
				l442:
					position, tokenIndex = position442, tokenIndex442
				}
				if !_rules[ruleSpacing]() {
					goto l439
				}
				add(ruleSLIST, position440)
			}
			return true
		l439:
			position, tokenIndex = position439, tokenIndex439
			return false
		},
		/* 86 LBRK <- <('[' Spacing)> */
		func() bool {
			position443, tokenIndex443 := position, tokenIndex
			{
				position444 := position
				if buffer[position] != rune('[') {
					goto l443
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l443
				}
				add(ruleLBRK, position444)
			}
			return true
		l443:
			position, tokenIndex = position443, tokenIndex443
			return false
		},
		/* 87 RBRK <- <(']' Spacing)> */
		func() bool {
			position445, tokenIndex445 := position, tokenIndex
			{
				position446 := position
				if buffer[position] != rune(']') {
					goto l445
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l445
				}
				add(ruleRBRK, position446)
			}
			return true
		l445:
			position, tokenIndex = position445, tokenIndex445
			return false
		},
		/* 88 LPAR <- <('(' Spacing)> */
		func() bool {
			position447, tokenIndex447 := position, tokenIndex
			{
				position448 := position
				if buffer[position] != rune('(') {
					goto l447
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l447
				}
				add(ruleLPAR, position448)
			}
			return true
		l447:
			position, tokenIndex = position447, tokenIndex447
			return false
		},
		/* 89 RPAR <- <(')' Spacing)> */
		func() bool {
			position449, tokenIndex449 := position, tokenIndex
			{
				position450 := position
				if buffer[position] != rune(')') {
					goto l449
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l449
				}
				add(ruleRPAR, position450)
			}
			return true
		l449:
			position, tokenIndex = position449, tokenIndex449
			return false
		},
		/* 90 LWING <- <('{' Spacing)> */
		func() bool {
			position451, tokenIndex451 := position, tokenIndex
			{
				position452 := position
				if buffer[position] != rune('{') {
					goto l451
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l451
				}
				add(ruleLWING, position452)
			}
			return true
		l451:
			position, tokenIndex = position451, tokenIndex451
			return false
		},
		/* 91 RWING <- <('}' Spacing)> */
		func() bool {
			position453, tokenIndex453 := position, tokenIndex
			{
				position454 := position
				if buffer[position] != rune('}') {
					goto l453
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l453
				}
				add(ruleRWING, position454)
			}
			return true
		l453:
			position, tokenIndex = position453, tokenIndex453
			return false
		},
		/* 92 LPOINT <- <('<' Spacing)> */
		func() bool {
			position455, tokenIndex455 := position, tokenIndex
			{
				position456 := position
				if buffer[position] != rune('<') {
					goto l455
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l455
				}
				add(ruleLPOINT, position456)
			}
			return true
		l455:
			position, tokenIndex = position455, tokenIndex455
			return false
		},
		/* 93 RPOINT <- <('>' Spacing)> */
		func() bool {
			position457, tokenIndex457 := position, tokenIndex
			{
				position458 := position
				if buffer[position] != rune('>') {
					goto l457
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l457
				}
				add(ruleRPOINT, position458)
			}
			return true
		l457:
			position, tokenIndex = position457, tokenIndex457
			return false
		},
		/* 94 EQUAL <- <('=' !'=' Spacing)> */
		func() bool {
			position459, tokenIndex459 := position, tokenIndex
			{
				position460 := position
				if buffer[position] != rune('=') {
					goto l459
				}
				position++
				{
					position461, tokenIndex461 := position, tokenIndex
					if buffer[position] != rune('=') {
						goto l461
					}
					position++
					goto l459
				l461:
					position, tokenIndex = position461, tokenIndex461
				}
				if !_rules[ruleSpacing]() {
					goto l459
				}
				add(ruleEQUAL, position460)
			}
			return true
		l459:
			position, tokenIndex = position459, tokenIndex459
			return false
		},
		/* 95 COMMA <- <(',' Spacing)> */
		func() bool {
			position462, tokenIndex462 := position, tokenIndex
			{
				position463 := position
				if buffer[position] != rune(',') {
					goto l462
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l462
				}
				add(ruleCOMMA, position463)
			}
			return true
		l462:
			position, tokenIndex = position462, tokenIndex462
			return false
		},
		/* 96 COLON <- <(':' Spacing)> */
		func() bool {
			position464, tokenIndex464 := position, tokenIndex
			{
				position465 := position
				if buffer[position] != rune(':') {
					goto l464
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l464
				}
				add(ruleCOLON, position465)
			}
			return true
		l464:
			position, tokenIndex = position464, tokenIndex464
			return false
		},
		/* 97 EOT <- <!.> */
		func() bool {
			position466, tokenIndex466 := position, tokenIndex
			{
				position467 := position
				{
					position468, tokenIndex468 := position, tokenIndex
					if !matchDot() {
						goto l468
					}
					goto l466
				l468:
					position, tokenIndex = position468, tokenIndex468
				}
				add(ruleEOT, position467)
			}
			return true
		l466:
			position, tokenIndex = position466, tokenIndex466
			return false
		},
		nil,
	}
	p.rules = _rules
}
