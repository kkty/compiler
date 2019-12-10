// Code generated by goyacc grammar.y. DO NOT EDIT.

//line grammar.y:2
package parser

import __yyfmt__ "fmt"

//line grammar.y:2

import "github.com/kkty/compiler/ast"

//line grammar.y:7
type yySymType struct {
	yys  int
	val  interface{}
	node ast.Node
}

const BOOL = 57346
const INT = 57347
const FLOAT = 57348
const NOT = 57349
const MINUS = 57350
const PLUS = 57351
const MINUS_DOT = 57352
const PLUS_DOT = 57353
const AST_DOT = 57354
const SLASH_DOT = 57355
const EQUAL = 57356
const LESS_GREATER = 57357
const LESS_EQUAL = 57358
const GREATER_EQUAL = 57359
const LESS = 57360
const GREATER = 57361
const IF = 57362
const THEN = 57363
const ELSE = 57364
const IDENT = 57365
const LET = 57366
const IN = 57367
const REC = 57368
const COMMA = 57369
const ARRAY_CREATE = 57370
const READ_INT = 57371
const READ_FLOAT = 57372
const READ_BYTE = 57373
const PRINT_INT = 57374
const PRINT_CHAR = 57375
const INT_TO_FLOAT = 57376
const FLOAT_TO_INT = 57377
const SQRT = 57378
const DOT = 57379
const LESS_MINUS = 57380
const SEMICOLON = 57381
const LPAREN = 57382
const RPAREN = 57383
const EOF = 57384
const prec_let = 57385
const prec_if = 57386
const prec_tuple = 57387
const prec_unary_minus = 57388
const prec_app = 57389

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"BOOL",
	"INT",
	"FLOAT",
	"NOT",
	"MINUS",
	"PLUS",
	"MINUS_DOT",
	"PLUS_DOT",
	"AST_DOT",
	"SLASH_DOT",
	"EQUAL",
	"LESS_GREATER",
	"LESS_EQUAL",
	"GREATER_EQUAL",
	"LESS",
	"GREATER",
	"IF",
	"THEN",
	"ELSE",
	"IDENT",
	"LET",
	"IN",
	"REC",
	"COMMA",
	"ARRAY_CREATE",
	"READ_INT",
	"READ_FLOAT",
	"READ_BYTE",
	"PRINT_INT",
	"PRINT_CHAR",
	"INT_TO_FLOAT",
	"FLOAT_TO_INT",
	"SQRT",
	"DOT",
	"LESS_MINUS",
	"SEMICOLON",
	"LPAREN",
	"RPAREN",
	"EOF",
	"prec_let",
	"prec_if",
	"prec_tuple",
	"prec_unary_minus",
	"prec_app",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 473

var yyAct = [...]int{

	2, 92, 95, 87, 86, 39, 40, 41, 42, 85,
	97, 75, 53, 52, 51, 107, 94, 82, 38, 96,
	49, 59, 21, 22, 23, 61, 62, 63, 64, 65,
	66, 67, 68, 69, 70, 71, 72, 73, 74, 43,
	105, 48, 44, 21, 22, 23, 4, 5, 104, 7,
	83, 93, 80, 78, 103, 82, 45, 101, 20, 6,
	34, 35, 9, 8, 77, 79, 10, 11, 12, 13,
	14, 15, 16, 17, 18, 19, 89, 90, 91, 20,
	60, 25, 24, 33, 32, 34, 35, 26, 27, 30,
	31, 28, 29, 3, 46, 102, 1, 0, 106, 0,
	108, 109, 110, 47, 111, 50, 0, 0, 113, 54,
	55, 56, 57, 58, 0, 116, 117, 0, 0, 0,
	0, 0, 21, 22, 23, 4, 5, 0, 7, 25,
	24, 33, 32, 34, 35, 0, 0, 0, 6, 0,
	81, 9, 8, 0, 84, 0, 11, 12, 13, 14,
	15, 16, 17, 18, 19, 0, 0, 0, 20, 25,
	24, 33, 32, 34, 35, 26, 27, 30, 31, 28,
	29, 0, 21, 22, 23, 0, 0, 0, 37, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	36, 48, 112, 25, 24, 33, 32, 34, 35, 26,
	27, 30, 31, 28, 29, 0, 0, 0, 20, 0,
	0, 0, 37, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 36, 0, 98, 25, 24, 33,
	32, 34, 35, 26, 27, 30, 31, 28, 29, 0,
	0, 0, 0, 0, 0, 0, 37, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 36, 0,
	88, 25, 24, 33, 32, 34, 35, 26, 27, 30,
	31, 28, 29, 0, 0, 0, 0, 0, 115, 0,
	37, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 36, 25, 24, 33, 32, 34, 35, 26,
	27, 30, 31, 28, 29, 0, 0, 0, 0, 0,
	114, 0, 37, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 36, 25, 24, 33, 32, 34,
	35, 26, 27, 30, 31, 28, 29, 0, 0, 0,
	0, 0, 100, 0, 37, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 36, 25, 24, 33,
	32, 34, 35, 26, 27, 30, 31, 28, 29, 0,
	0, 99, 0, 0, 0, 0, 37, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 36, 25,
	24, 33, 32, 34, 35, 26, 27, 30, 31, 28,
	29, 0, 76, 0, 0, 0, 0, 0, 37, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	36, 25, 24, 33, 32, 34, 35, 26, 27, 30,
	31, 28, 29, 0, 0, 0, 0, 0, 0, 0,
	37, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 36, 25, 24, 33, 32, 34, 35, 26,
	27, 30, 31, 28, 29, 0, 0, 0, 0, 0,
	0, 0, 37,
}
var yyPact = [...]int{

	118, -1000, 413, -19, 118, 118, 118, 118, 16, 168,
	-7, 168, -26, -27, -28, 168, 168, 168, 168, 168,
	39, -1000, -1000, -1000, 118, 118, 118, 118, 118, 118,
	118, 118, 118, 118, 118, 118, 118, 118, -29, -1000,
	-1000, 381, -1000, 50, 30, 29, 168, -20, -1000, 118,
	18, -32, -37, -38, -20, -20, -20, -20, -20, 219,
	-1000, 48, 48, 121, 121, 121, 121, 121, 121, 48,
	48, -1000, -1000, 413, 73, 118, 118, 118, 28, -25,
	-8, -20, -30, 73, -20, -1000, -1000, -1000, -1000, 185,
	349, 317, 43, 28, 40, 25, 17, 118, -23, 118,
	118, 118, -1000, 118, -1000, -1000, 151, 118, 445, 413,
	285, 253, -1000, 445, 118, 118, 413, 413,
}
var yyPgo = [...]int{

	0, 96, 0, 93, 1, 94, 66, 65,
}
var yyR1 = [...]int{

	0, 1, 3, 3, 3, 3, 3, 3, 3, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 4, 4, 5, 5, 6, 6, 7, 7,
}
var yyR2 = [...]int{

	0, 1, 3, 2, 1, 1, 1, 1, 5, 1,
	2, 2, 3, 3, 3, 3, 3, 3, 3, 3,
	6, 2, 3, 3, 3, 3, 6, 8, 2, 1,
	8, 7, 3, 3, 3, 3, 3, 2, 2, 2,
	2, 2, 2, 1, 2, 1, 3, 3, 3, 3,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, 7, 8, 20, 10, 24, 23,
	-6, 28, 29, 30, 31, 32, 33, 34, 35, 36,
	40, 4, 5, 6, 9, 8, 14, 15, 18, 19,
	16, 17, 11, 10, 12, 13, 39, 27, 37, -2,
	-2, -2, -2, 23, 26, 40, -5, -3, 23, 27,
	-3, 40, 40, 40, -3, -3, -3, -3, -3, -2,
	41, -2, -2, -2, -2, -2, -2, -2, -2, -2,
	-2, -2, -2, -2, -2, 40, 21, 14, 23, -7,
	23, -3, 37, -2, -3, 41, 41, 41, 41, -2,
	-2, -2, -4, 23, 41, 27, 27, 40, 41, 22,
	25, 14, -4, 14, 23, 23, -2, 38, -2, -2,
	-2, -2, 41, -2, 25, 25, -2, -2,
}
var yyDef = [...]int{

	0, -2, 1, 9, 0, 0, 0, 0, 0, 7,
	29, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 4, 5, 6, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 10,
	11, 0, 21, 0, 0, 0, 28, 45, 7, 0,
	0, 0, 0, 0, 37, 38, 39, 40, 41, 0,
	3, 12, 13, 14, 15, 16, 17, 18, 19, 22,
	23, 24, 25, 32, 47, 0, 0, 0, 0, 0,
	0, 44, 0, 46, 33, 34, 35, 36, 2, 0,
	0, 0, 0, 43, 0, 0, 0, 0, 8, 0,
	0, 0, 42, 0, 48, 49, 0, 0, 20, 26,
	0, 0, 8, 31, 0, 0, 27, 30,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:79
		{
			yylex.(*lexer).result = yyDollar[1].node
		}
	case 2:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:82
		{
			yyVAL.node = yyDollar[2].node
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:84
		{
			yyVAL.node = &ast.Unit{}
		}
	case 4:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:86
		{
			yyVAL.node = &ast.Bool{yyDollar[1].val.(bool)}
		}
	case 5:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:88
		{
			yyVAL.node = &ast.Int{yyDollar[1].val.(int32)}
		}
	case 6:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:90
		{
			yyVAL.node = &ast.Float{yyDollar[1].val.(float32)}
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:92
		{
			yyVAL.node = &ast.Variable{yyDollar[1].val.(string)}
		}
	case 8:
		yyDollar = yyS[yypt-5 : yypt+1]
//line grammar.y:94
		{
			yyVAL.node = &ast.ArrayGet{yyDollar[1].node, yyDollar[4].node}
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:97
		{
			yyVAL.node = yyDollar[1].node
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:100
		{
			yyVAL.node = &ast.Not{yyDollar[2].node}
		}
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:103
		{
			yyVAL.node = &ast.Neg{yyDollar[2].node}
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:105
		{
			yyVAL.node = &ast.Add{yyDollar[1].node, yyDollar[3].node}
		}
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:107
		{
			yyVAL.node = &ast.Sub{yyDollar[1].node, yyDollar[3].node}
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:109
		{
			yyVAL.node = &ast.Equal{yyDollar[1].node, yyDollar[3].node}
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:111
		{
			yyVAL.node = &ast.Not{&ast.Equal{yyDollar[1].node, yyDollar[3].node}}
		}
	case 16:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:113
		{
			yyVAL.node = &ast.LessThan{yyDollar[1].node, yyDollar[3].node}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:115
		{
			yyVAL.node = &ast.LessThan{yyDollar[3].node, yyDollar[1].node}
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:117
		{
			yyVAL.node = &ast.Not{&ast.LessThan{yyDollar[3].node, yyDollar[1].node}}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:119
		{
			yyVAL.node = &ast.Not{&ast.LessThan{yyDollar[1].node, yyDollar[3].node}}
		}
	case 20:
		yyDollar = yyS[yypt-6 : yypt+1]
//line grammar.y:122
		{
			yyVAL.node = &ast.If{yyDollar[2].node, yyDollar[4].node, yyDollar[6].node}
		}
	case 21:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:125
		{
			yyVAL.node = &ast.FloatNeg{yyDollar[2].node}
		}
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:127
		{
			yyVAL.node = &ast.FloatAdd{yyDollar[1].node, yyDollar[3].node}
		}
	case 23:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:129
		{
			yyVAL.node = &ast.FloatSub{yyDollar[1].node, yyDollar[3].node}
		}
	case 24:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:131
		{
			yyVAL.node = &ast.FloatMul{yyDollar[1].node, yyDollar[3].node}
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:133
		{
			yyVAL.node = &ast.FloatDiv{yyDollar[1].node, yyDollar[3].node}
		}
	case 26:
		yyDollar = yyS[yypt-6 : yypt+1]
//line grammar.y:136
		{
			yyVAL.node = &ast.ValueBinding{yyDollar[2].val.(string), yyDollar[4].node, yyDollar[6].node}
		}
	case 27:
		yyDollar = yyS[yypt-8 : yypt+1]
//line grammar.y:139
		{
			yyVAL.node = &ast.FunctionBinding{yyDollar[3].val.(string), yyDollar[4].val.([]string), yyDollar[6].node, yyDollar[8].node}
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:142
		{
			yyVAL.node = &ast.Application{yyDollar[1].val.(string), yyDollar[2].val.([]ast.Node)}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:145
		{
			yyVAL.node = &ast.Tuple{yyDollar[1].val.([]ast.Node)}
		}
	case 30:
		yyDollar = yyS[yypt-8 : yypt+1]
//line grammar.y:147
		{
			yyVAL.node = &ast.TupleBinding{yyDollar[3].val.([]string), yyDollar[6].node, yyDollar[8].node}
		}
	case 31:
		yyDollar = yyS[yypt-7 : yypt+1]
//line grammar.y:149
		{
			yyVAL.node = &ast.ArrayPut{yyDollar[1].node, yyDollar[4].node, yyDollar[7].node}
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:151
		{
			yyVAL.node = &ast.ValueBinding{"", yyDollar[1].node, yyDollar[3].node}
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:154
		{
			yyVAL.node = &ast.ArrayCreate{yyDollar[2].node, yyDollar[3].node}
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:157
		{
			yyVAL.node = &ast.ReadInt{}
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:160
		{
			yyVAL.node = &ast.ReadFloat{}
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:163
		{
			yyVAL.node = &ast.ReadByte{}
		}
	case 37:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:166
		{
			yyVAL.node = &ast.PrintInt{yyDollar[2].node}
		}
	case 38:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:169
		{
			yyVAL.node = &ast.WriteByte{yyDollar[2].node}
		}
	case 39:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:172
		{
			yyVAL.node = &ast.IntToFloat{yyDollar[2].node}
		}
	case 40:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:175
		{
			yyVAL.node = &ast.FloatToInt{yyDollar[2].node}
		}
	case 41:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:178
		{
			yyVAL.node = &ast.Sqrt{yyDollar[2].node}
		}
	case 42:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:181
		{
			yyVAL.val = append([]string{yyDollar[1].val.(string)}, yyDollar[2].val.([]string)...)
		}
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:183
		{
			yyVAL.val = []string{yyDollar[1].val.(string)}
		}
	case 44:
		yyDollar = yyS[yypt-2 : yypt+1]
//line grammar.y:187
		{
			yyVAL.val = append(yyDollar[1].val.([]ast.Node), yyDollar[2].node)
		}
	case 45:
		yyDollar = yyS[yypt-1 : yypt+1]
//line grammar.y:190
		{
			yyVAL.val = []ast.Node{yyDollar[1].node}
		}
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:193
		{
			yyVAL.val = append(yyDollar[1].val.([]ast.Node), yyDollar[3].node)
		}
	case 47:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:195
		{
			yyVAL.val = append([]ast.Node{yyDollar[1].node}, yyDollar[3].node)
		}
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:198
		{
			yyVAL.val = append(yyDollar[1].val.([]string), yyDollar[3].val.(string))
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
//line grammar.y:200
		{
			yyVAL.val = append([]string{yyDollar[1].val.(string)}, yyDollar[3].val.(string))
		}
	}
	goto yystack /* stack new state and value */
}
