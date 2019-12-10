%{
package parser

import "github.com/kkty/compiler/ast"
%}

%union{
  val interface{}
  node ast.Node
}

%token<val> BOOL
%token<val> INT
%token<val> FLOAT
%token<> NOT
%token<> MINUS
%token<> PLUS
%token<> MINUS_DOT
%token<> PLUS_DOT
%token<> AST_DOT
%token<> SLASH_DOT
%token<> EQUAL
%token<> LESS_GREATER
%token<> LESS_EQUAL
%token<> GREATER_EQUAL
%token<> LESS
%token<> GREATER
%token<> IF
%token<> THEN
%token<> ELSE
%token<val> IDENT
%token<> LET
%token<> IN
%token<> REC
%token<> COMMA
%token<> ARRAY_CREATE
%token<> READ_INT
%token<> READ_FLOAT
%token<> READ_BYTE
%token<> PRINT_INT
%token<> PRINT_CHAR
%token<> INT_TO_FLOAT
%token<> FLOAT_TO_INT
%token<> SQRT
%token<> DOT
%token<> LESS_MINUS
%token<> SEMICOLON
%token<> LPAREN
%token<> RPAREN
%token<> EOF

%nonassoc IN
%right prec_let
%right SEMICOLON
%right prec_if
%right LESS_MINUS
%nonassoc prec_tuple
%left COMMA
%left EQUAL LESS_GREATER LESS GREATER LESS_EQUAL GREATER_EQUAL
%left PLUS MINUS PLUS_DOT MINUS_DOT
%left AST_DOT SLASH_DOT
%right prec_unary_minus
%left prec_app
%left DOT

%type<> program
%type<node> exp
%type<node> simple_exp
%type<val> formal_args
%type<val> actual_args
%type<val> elems
%type<val> pat

%start program

%%

program: exp
  { yylex.(*lexer).result = $1 }

simple_exp: LPAREN exp RPAREN
  { $$ = $2 }
| LPAREN RPAREN
  { $$ = &ast.Unit{} }
| BOOL
  { $$ = &ast.Bool{$1.(bool)} }
| INT
  { $$ = &ast.Int{$1.(int32)} }
| FLOAT
  { $$ = &ast.Float{$1.(float32)} }
| IDENT
  { $$ = &ast.Variable{$1.(string)} }
| simple_exp DOT LPAREN exp RPAREN
  { $$ = &ast.ArrayGet{$1, $4} }

exp: simple_exp
  { $$ = $1 }
| NOT exp
  %prec prec_app
  { $$ = &ast.Not{$2} }
| MINUS exp
  %prec prec_unary_minus
  { $$ = &ast.Neg{$2} }
| exp PLUS exp 
  { $$ = &ast.Add{$1, $3} }
| exp MINUS exp
  { $$ = &ast.Sub{$1, $3} }
| exp EQUAL exp
  { $$ = &ast.Equal{$1, $3} }
| exp LESS_GREATER exp
  { $$ = &ast.Not{&ast.Equal{$1, $3}} }
| exp LESS exp
  { $$ = &ast.LessThan{$1, $3} }
| exp GREATER exp
  { $$ = &ast.LessThan{$3, $1} }
| exp LESS_EQUAL exp
  { $$ = &ast.Not{&ast.LessThan{$3, $1}} }
| exp GREATER_EQUAL exp
  { $$ = &ast.Not{&ast.LessThan{$1, $3}} }
| IF exp THEN exp ELSE exp
  %prec prec_if
  { $$ = &ast.If{$2, $4, $6} }
| MINUS_DOT exp
  %prec prec_unary_minus
  { $$ = &ast.FloatNeg{$2} }
| exp PLUS_DOT exp
  { $$ = &ast.FloatAdd{$1, $3} }
| exp MINUS_DOT exp
  { $$ = &ast.FloatSub{$1, $3} }
| exp AST_DOT exp
  { $$ = &ast.FloatMul{$1, $3} }
| exp SLASH_DOT exp
  { $$ = &ast.FloatDiv{$1, $3} }
| LET IDENT EQUAL exp IN exp
  %prec prec_let
  { $$ = &ast.ValueBinding{$2.(string), $4, $6} }
| LET REC IDENT formal_args EQUAL exp IN exp
  %prec prec_let
  { $$ = &ast.FunctionBinding{$3.(string), $4.([]string), $6, $8} }
| IDENT actual_args
  %prec prec_app
  { $$ = &ast.Application{$1.(string), $2.([]ast.Node)} }
| elems
  %prec prec_tuple
  { $$ = &ast.Tuple{$1.([]ast.Node)} }
| LET LPAREN pat RPAREN EQUAL exp IN exp
  { $$ = &ast.TupleBinding{$3.([]string), $6, $8} }
| simple_exp DOT LPAREN exp RPAREN LESS_MINUS exp
  { $$ = &ast.ArrayPut{$1, $4, $7} }
| exp SEMICOLON exp
  { $$ = &ast.ValueBinding{"", $1, $3} }
| ARRAY_CREATE simple_exp simple_exp
  %prec prec_app
  { $$ = &ast.ArrayCreate{$2, $3} }
| READ_INT LPAREN RPAREN
  %prec prec_app
  { $$ = &ast.ReadInt{} }
| READ_FLOAT LPAREN RPAREN
  %prec prec_app
  { $$ = &ast.ReadFloat{} }
| READ_BYTE LPAREN RPAREN
  %prec prec_app
  { $$ = &ast.ReadByte{} }
| PRINT_INT simple_exp
  %prec prec_app
  { $$ = &ast.PrintInt{$2} }
| PRINT_CHAR simple_exp
  %prec prec_app
  { $$ = &ast.WriteByte{$2} }
| INT_TO_FLOAT simple_exp
  %prec prec_app
  { $$ = &ast.IntToFloat{$2} }
| FLOAT_TO_INT simple_exp
  %prec prec_app
  { $$ = &ast.FloatToInt{$2} }
| SQRT simple_exp
  %prec prec_app
  { $$ = &ast.Sqrt{$2} }

formal_args: IDENT formal_args
  { $$ = append([]string{$1.(string)}, $2.([]string)...) }
| IDENT
  { $$ = []string{$1.(string)} }

actual_args: actual_args simple_exp
  %prec prec_app
  { $$ = append($1.([]ast.Node), $2) }
| simple_exp
  %prec prec_app
  { $$ = []ast.Node{$1} }

elems: elems COMMA exp
  { $$ = append($1.([]ast.Node), $3) }
| exp COMMA exp
  { $$ = append([]ast.Node{$1}, $3) }

pat: pat COMMA IDENT
  { $$ = append($1.([]string), $3.(string)) }
| IDENT COMMA IDENT
  { $$ = append([]string{$1.(string)}, $3.(string)) } 
