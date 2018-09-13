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

%define api.pure full
%define parse.error verbose
%locations
%parse-param { yyparam_t *param }
%param { void *yyscanner }
%debug
%verbose
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
        ID

/* expr_lit */
%token  <str>
        FLOAT           HEXA            INT             STRING

/* expr_sql */
%token  <str>
        DML             QUERY

/* operator */
%token  OP_ADD_ASSIGN   OP_SUB_ASSIGN   OP_MUL_ASSIGN   OP_DIV_ASSIGN
        OP_MOD_ASSIGN   OP_AND_ASSIGN   OP_XOR_ASSIGN   OP_OR_ASSIGN
        OP_RS_ASSIGN    OP_LS_ASSIGN    OP_RSHIFT       OP_LSHIFT
        OP_INC          OP_DEC          OP_AND          OP_OR
        OP_LE           OP_GE           OP_EQ           OP_NE

/* keyword */
%token  /* A */
        K_ACCOUNT
        /* B */
        K_BOOL          K_BREAK         K_BYTE
        /* C */
        K_CASE          K_CONST         K_CONTINUE      K_CONTRACT
        K_CREATE
        /* D */
        K_DEFAULT       K_DELETE        K_DOUBLE        K_DROP
        /* E */
        K_ELSE
        /* F */
        K_FALSE         K_FLOAT         K_FOR           K_FOREACH
        K_FUNC
        /* G */
        /* H */
        /* I */
        K_IF            K_IN            K_INDEX         K_INSERT        
        K_INT           K_INT16         K_INT32         K_INT64
        /* L */
        K_LOCAL
        /* M */
        K_MAP
        /* N */
        K_NEW           K_NULL
        /* O */
        /* P */
        /* Q */
        /* R */
        K_READONLY      K_RETURN
        /* S */
        K_SELECT        K_SHARED        K_STRING        K_STRUCT
        K_SWITCH
        /* T */
        K_TABLE         K_TRANSFER      K_TRUE
        /* U */
        K_UINT          K_UINT16        K_UINT32        K_UINT64
        K_UPDATE
        /* V */
        /* W */
        /* X */
        /* Y */
        /* Z */

/* precedences */
%left   OP_OR
%left   OP_AND
%right  '!'
%left   '|'
%left   '&'
%left   OP_EQ OP_NE
%left   OP_LE OP_GE '<' '>'
%left   '+' '-' '%'
%left   '*' '/'
%left   OP_INC OP_DEC
%left   '.'

/* types */
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
|   OP_ADD_ASSIGN
|   OP_SUB_ASSIGN
|   OP_MUL_ASSIGN
|   OP_DIV_ASSIGN
|   OP_MOD_ASSIGN
|   OP_AND_ASSIGN
|   OP_XOR_ASSIGN
|   OP_OR_ASSIGN
|   OP_RS_ASSIGN
|   OP_LS_ASSIGN
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
|   expr_or OP_OR expr_and
;

expr_and:
    expr_bit_or
|   expr_and OP_AND expr_bit_or
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
|   expr_eq OP_EQ expr_cmp
|   expr_eq OP_NE expr_cmp
;

expr_cmp:
    expr_shift
|   expr_cmp op_cmp expr_shift
;

op_cmp:
    '<'
|   '>'
|   OP_LE
|   OP_GE
;

expr_shift:
    expr_add
|   expr_shift OP_RSHIFT expr_add
|   expr_shift OP_LSHIFT expr_add
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
|   OP_INC expr_unary
|   OP_DEC expr_unary
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
|   expr_post OP_INC
|   expr_post OP_DEC
;

expr_prim:
    K_NULL
|   K_TRUE
|   K_FALSE
|   INT
|   FLOAT
|   HEXA
|   QUERY
|   STRING
|   identifier
|   '(' expression ')'
;

expr_new:
    K_NEW identifier '(' ')'
|   K_NEW identifier '(' expr_list ')'
|   K_NEW K_MAP '(' ')'
|   K_NEW K_MAP '(' expr_list ')'
;

expr_list:
    expr_assign
|   expr_list ',' expr_assign
;

identifier:
    ID              { $$ = $1; }
|   K_CONTRACT      { $$ = xstrdup("contract"); }
|   K_IN            { $$ = xstrdup("in"); }
|   K_INDEX         { $$ = xstrdup("index"); }
|   K_TABLE         { $$ = xstrdup("table"); }
;

%%

static void
yyerror(YYLTYPE *lloc, yyparam_t *param, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, FILENAME(param->path), lloc->first.line, msg,
          make_trace(param->path, lloc));
}

/* end of grammar.y */
