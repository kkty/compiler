package parser

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/kkty/mincaml-go/ast"
)

type lexer struct {
	program string
	result  ast.Node
}

func atoi(s string) int32 {
	i, err := strconv.Atoi(s)

	if err != nil {
		log.Fatal(err)
	}

	return int32(i)
}

func atof(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)

	if err != nil {
		log.Fatal(err)
	}

	return float32(f)
}

func (l *lexer) Lex(lval *yySymType) int {
	advance := func(i int) {
		l.program = l.program[i:]
	}

	hasPrefix := func(s string) bool {
		return strings.HasPrefix(l.program, s)
	}

	skipWhitespaces := func() bool {
		modified := false
		for hasPrefix(" ") || hasPrefix("\n") || hasPrefix("\t") {
			modified = true
			advance(1)
		}
		return modified
	}

	skipComments := func() bool {
		modified := false
		if hasPrefix("(*") {
			advance(2)

			for !hasPrefix("*)") {
				advance(1)
			}

			advance(2)

			modified = true
		}
		return modified
	}

	for skipComments() || skipWhitespaces() {
	}

	if len(l.program) == 0 {
		// 0 stands for EOF.
		return 0
	}

	patterns := []struct {
		pattern string
		token   int
		f       func(s string)
	}{
		{"\\(", LPAREN, nil},
		{"\\)", RPAREN, nil},
		{"true", BOOL, func(s string) { lval.val = true }},
		{"false", BOOL, func(s string) { lval.val = false }},
		{"not", NOT, nil},
		{"[0-9]+", INT, func(s string) { lval.val = atoi(s) }},
		{"[0-9]+(\\.[0-9]*)?([eE][\\+\\-]?[0-9]+)?", FLOAT, func(s string) { lval.val = atof(s) }},
		{"-", MINUS, nil},
		{"\\+", PLUS, nil},
		{"-\\.", MINUS_DOT, nil},
		{"\\+\\.", PLUS_DOT, nil},
		{"\\*\\.", AST_DOT, nil},
		{"/\\.", SLASH_DOT, nil},
		{"=", EQUAL, nil},
		{"<>", LESS_GREATER, nil},
		{"<=", LESS_EQUAL, nil},
		{">=", GREATER_EQUAL, nil},
		{"<", LESS, nil},
		{">", GREATER, nil},
		{"if", IF, nil},
		{"then", THEN, nil},
		{"else", ELSE, nil},
		{"let", LET, nil},
		{"in", IN, nil},
		{"rec", REC, nil},
		{",", COMMA, nil},
		{"_", IDENT, func(s string) { lval.val = "" }},
		{"create_array", ARRAY_CREATE, nil},
		{"read_int", READ_INT, nil},
		{"read_float", READ_FLOAT, nil},
		{"print_int", PRINT_INT, nil},
		{"print_char", PRINT_CHAR, nil},
		{"\\.", DOT, nil},
		{"<-", LESS_MINUS, nil},
		{";", SEMICOLON, nil},
		{"[a-z][0-9a-zA-Z_]*", IDENT, func(s string) { lval.val = s }},
	}

	longestMatch := struct {
		pattern string
		found   string
		token   int
		f       func(s string)
	}{}

	for _, pattern := range patterns {
		found := regexp.MustCompile("^" + pattern.pattern).FindString(l.program)

		if len(found) > len(longestMatch.found) {
			longestMatch.pattern = pattern.pattern
			longestMatch.token = pattern.token
			longestMatch.found = found
			longestMatch.f = pattern.f
		}
	}

	if longestMatch.pattern == "" {
		log.Fatal("no matching token")
	}

	if f := longestMatch.f; f != nil {
		f(longestMatch.found)
	}

	advance(len(longestMatch.found))

	return longestMatch.token

}

func (l *lexer) Error(e string) {
	log.Fatal(e)
}

func Parse(program string) ast.Node {
	l := lexer{program: program}
	yyParse(&l)
	return l.result
}
