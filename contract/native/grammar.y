%{

/**
 * @file    grammar.y
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "error.h"
#include "parser.h"

#define YYLLOC_DEFAULT(Current, Rhs, N)                                        \
    (Current) = YYRHSLOC(Rhs, (N) > 0 ? 1 : 0)

extern int yylex(YYSTYPE *lval, YYLTYPE *lloc, void *yyscanner);
extern void yylex_set_token(void *yyscanner, int token, YYLTYPE *lloc);

static void yyerror(YYLTYPE *lloc, yyparam_t *param, void *scanner,
                    const char *msg);

%}

%parse-param { yyparam_t *param }
%param { void *yyscanner }
%locations
%debug
%verbose
%define api.pure full
%define parse.error verbose
%initial-action {
    yylloc.first.line = 1;
    yylloc.first.col = 1;
    yylloc.first.offset = 0;
    yylloc.last.line = 1;
    yylloc.last.col = 1;
    yylloc.last.offset = 0;
}

%union {
    int ival;
    char *str;
}

/* identifier */
%token  <str>
        ID              "identifier"

/* literal */
%token  <str>
        L_FLOAT         "floating-point number"
        L_HEXA          "hexadecimal number"
        L_INT           "integer"
        L_STR           "characters"

/* operator */
%token  ASSIGN_ADD      "+="
        ASSIGN_SUB      "-="
        ASSIGN_MUL      "*="
        ASSIGN_DIV      "/="
        ASSIGN_MOD      "%="
        ASSIGN_AND      "&="
        ASSIGN_XOR      "^="
        ASSIGN_OR       "|="
        ASSIGN_RS       ">>="
        ASSIGN_LS       "<<="
        SHIFT_R         ">>"
        SHIFT_L         "<<"
        CMP_AND         "&&"
        CMP_OR          "||"
        CMP_LE          "<="
        CMP_GE          ">="
        CMP_EQ          "=="
        CMP_NE          "!="
        UNARY_INC       "++"
        UNARY_DEC       "--"

/* keyword */
%token  K_ACCOUNT       "account"
        K_BOOL          "bool"
        K_BREAK         "break"
        K_BYTE          "byte"
        K_CASE          "case"
        K_CONST         "const"
        K_CONTINUE      "continue"
        K_CONTRACT      "contract"
        K_CREATE        "create"
        K_DEFAULT       "default"
        K_DELETE        "delete"
        K_DOUBLE        "double"
        K_DROP          "drop"
        K_ELSE          "else"
        K_FALSE         "false"
        K_FLOAT         "float"
        K_FOR           "for"
        K_FOREACH       "foreach"
        K_FUNC          "func"
        K_IF            "if"
        K_IN            "in"
        K_INDEX         "index"
        K_INSERT        "insert"
        K_INT           "int"
        K_INT16         "int16"
        K_INT32         "int32"
        K_INT64         "int64"
        K_LOCAL         "local"
        K_MAP           "map"
        K_NEW           "new"
        K_NULL          "null"
        K_READONLY      "readonly"
        K_RETURN        "return"
        K_SELECT        "select"
        K_SHARED        "shared"
        K_STRING        "string"
        K_STRUCT        "struct"
        K_SWITCH        "switch"
        K_TABLE         "table"
        K_TRANSFER      "transfer"
        K_TRUE          "true"
        K_UINT          "uint"
        K_UINT16        "uint16"
        K_UINT32        "uint32"
        K_UINT64        "uint64"
        K_UPDATE        "update"

/* precedences */
%left   CMP_OR
%left   CMP_AND
%right  '!'
%left   '|'
%left   '&'
%left   CMP_EQ CMP_NE
%left   CMP_LE CMP_GE '<' '>'
%left   '+' '-' '%'
%left   '*' '/'
%left   UNARY_INC UNARY_DEC
%left   '.'

/* types */
%type   <str>           non_reserved_token
%type   <str>           identifier

%start  smart_contract

%%

smart_contract:
    contract_decl
|   smart_contract contract_decl
;

contract_decl:
    K_CONTRACT identifier '{' '}'
|   K_CONTRACT identifier '{' contract_body '}'
;

contract_body:
    variable
|   contract_body variable
|   struct
|   contract_body struct
|   constructor
|   contract_body constructor
|   function
|   contract_body function
;

variable:
    var_type var_decl_list ';'
;

var_type:
    type_spec
|   type_qual type_spec
|   var_scope var_type
;

type_qual:
    K_CONST
;

var_scope:
    K_LOCAL
|   K_SHARED
;

type_spec:
    K_ACCOUNT
|   K_BOOL
|   K_BYTE
|   K_FLOAT
|   K_DOUBLE
|   K_INT
|   K_INT16
|   K_INT32
|   K_INT64
|   K_STRING
|   K_UINT
|   K_UINT16
|   K_UINT32
|   K_UINT64
|   map_spec
|   identifier
;

map_spec:
    K_MAP '(' type_spec ',' type_spec ')'
;

var_decl_list:
    var_decl
|   var_decl_list ',' var_decl
;

var_decl:
    declarator
|   declarator '=' initializer
;

declarator:
    identifier
|   declarator '[' expr_add ']'
|   declarator '[' ']'
;

initializer:
    expr_sql
|   '{' init_list '}'
|   '{' init_list ',' '}'
;

init_list:
    initializer
|   init_list ',' initializer
;

struct:
    K_STRUCT identifier '{' field_list '}'
;

field_list:
    variable
|   field_list variable
;

constructor:
    identifier '(' param_list_opt ')' stmt_blk
;

param_list_opt:
    /* empty */
|   param_list
;

param_list:
    param_decl
|   param_list ',' param_decl
;

param_decl:
    var_type declarator
;

function:
    modifier_opt K_FUNC identifier '(' param_list_opt ')' return_opt stmt_blk
;

modifier_opt:
    /* empty */
|   K_LOCAL
|   K_SHARED
|   modifier_opt K_READONLY
|   modifier_opt K_TRANSFER
;

return_opt:
    /* empty */
|   return_list
|   '(' return_list ')'
;

return_list:
    var_type
|   return_list ',' var_type
;

statement:
    stmt_expr
|   stmt_if
|   stmt_loop
|   stmt_switch
|   stmt_jump
|   stmt_ddl
|   stmt_blk
;

stmt_expr:
    ';'
|   expression ';'
;

stmt_if:
    K_IF '(' expression ')' stmt_blk
|   K_ELSE stmt_if
|   K_ELSE stmt_blk
;

stmt_loop:
    K_FOR '(' stmt_expr stmt_expr ')' stmt_blk
|   K_FOR '(' stmt_expr stmt_expr expression ')' stmt_blk
|   K_FOR '(' variable stmt_expr ')' stmt_blk
|   K_FOR '(' variable stmt_expr expression ')' stmt_blk
|   K_FOR stmt_blk
|   K_FOREACH '(' iter_decl ',' iter_decl K_IN expr_post ')' stmt_blk
;

iter_decl:
    var_type declarator
;

stmt_switch:
    K_SWITCH '{' switch_blk '}'
|   K_SWITCH '(' expression ')' '{' switch_blk '}'
;

switch_blk:
    stmt_label
|   switch_blk stmt_label
;

stmt_label:
    K_CASE expr_eq ':' stmt_list
|   K_DEFAULT ':' stmt_list
;

stmt_list:
    statement
|   stmt_list statement
;

stmt_jump:
    K_CONTINUE ';'
|   K_BREAK ';'
|   K_RETURN ';'
|   K_RETURN expression ';'
;

stmt_ddl:
    sql_ddl error ';'
    {
        yyerrok;
        error_pop();
        yylex_set_token(yyscanner, ';', &@3);
        yyclearin;
    }
;

sql_ddl:
    K_CREATE K_INDEX
|   K_CREATE K_TABLE
|   K_DROP K_INDEX
|   K_DROP K_TABLE
;

stmt_blk:
    '{' '}'
|   '{' blk_decl_list '}'
;

blk_decl_list:
    blk_decl
|   blk_decl_list blk_decl
;

blk_decl:
    variable
|   struct
|   statement
;

expression:
    expr_assign
|   expression ',' expr_assign
;

expr_assign:
    expr_sql
|   expr_unary op_assign expr_assign
;

op_assign:
    '='
|   ASSIGN_ADD
|   ASSIGN_SUB
|   ASSIGN_MUL
|   ASSIGN_DIV
|   ASSIGN_MOD
|   ASSIGN_AND
|   ASSIGN_XOR
|   ASSIGN_OR
|   ASSIGN_RS
|   ASSIGN_LS
;

expr_sql:
    expr_cond
|   sql_prefix error ';'
    {
        yyerrok;
        error_pop();
        yylex_set_token(yyscanner, ';', &@3);
        yyclearin;
    }
;

sql_prefix:
    K_DELETE
|   K_INSERT
|   K_SELECT
|   K_UPDATE
;

expr_cond:
    expr_or
|   expr_or '?' expression ':' expr_cond
;

expr_or:
    expr_and
|   expr_or CMP_OR expr_and
;

expr_and:
    expr_bit_or
|   expr_and CMP_AND expr_bit_or
;

expr_bit_or:
    expr_bit_xor
|   expr_bit_or '|' expr_bit_xor
;

expr_bit_xor:
    expr_bit_and
|   expr_bit_xor '^' expr_bit_and
;

expr_bit_and:
    expr_eq
|   expr_bit_and '&' expr_eq
;

expr_eq:
    expr_cmp
|   expr_eq CMP_EQ expr_cmp
|   expr_eq CMP_NE expr_cmp
;

expr_cmp:
    expr_shift
|   expr_cmp op_cmp expr_shift
;

op_cmp:
    '<'
|   '>'
|   CMP_LE
|   CMP_GE
;

expr_shift:
    expr_add
|   expr_shift SHIFT_R expr_add
|   expr_shift SHIFT_L expr_add
;

expr_add:
    expr_mul
|   expr_add '+' expr_mul
|   expr_add '-' expr_mul
;

expr_mul:
    expr_unary
|   expr_mul '*' expr_unary
|   expr_mul '/' expr_unary
|   expr_mul '%' expr_unary
;

expr_unary:
    expr_post
|   UNARY_INC expr_unary
|   UNARY_DEC expr_unary
|   op_unary expr_unary
;

op_unary:
    '+'
|   '-'
|   '!'
;

expr_post:
    expr_prim
|   expr_new
|   expr_post '[' expression ']'
|   expr_post '(' ')'
|   expr_post '(' expr_list ')'
|   expr_post '.' identifier
|   expr_post UNARY_INC
|   expr_post UNARY_DEC
;

expr_prim:
    K_NULL
|   K_TRUE
|   K_FALSE
|   L_INT
|   L_FLOAT
|   L_HEXA
|   L_STR
|   identifier
|   '(' expression ')'
;

expr_new:
    K_NEW identifier '(' ')'
|   K_NEW identifier '(' expr_list ')'
|   K_NEW K_MAP '(' ')'
|   K_NEW K_MAP '(' L_INT ')'
;

expr_list:
    expr_cond
|   expr_list ',' expr_cond
;

non_reserved_token:
    K_CONTRACT          { $$ = xstrdup("contract"); }
|   K_IN                { $$ = xstrdup("in"); }
|   K_INDEX             { $$ = xstrdup("index"); }
|   K_TABLE             { $$ = xstrdup("table"); }
;

identifier:
    ID                  { $$ = $1; }
|   non_reserved_token  { $$ = $1; }
;

%%

static void
yyerror(YYLTYPE *lloc, yyparam_t *param, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, FILENAME(param->path), lloc->first.line, msg,
          make_trace(param->path, lloc));
}

/* end of grammar.y */
