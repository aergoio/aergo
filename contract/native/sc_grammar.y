%{

/**
 * @file    sc.grammar.y
 * @copyright defined in aergo/LICENSE.txt
 */

#include "sc_common.h"

#include "sc_parser.h"

#define YYLLOC_DEFAULT(Current, Rhs, N)                                        \
    (Current) = YYRHSLOC(Rhs, (N) > 0 ? 1 : 0)

extern int sc_yylex(YYSTYPE *lval, YYLTYPE *lloc, void *yyscanner);

static void sc_yyerror(YYLTYPE *lloc, sc_yacc_t *yacc, const char *msg);

%}

%define api.pure full
%define parse.error verbose
%locations
%parse-param { void *yyscanner }
%lex-param { yyscan_t yyscanner }
%name-prefix "sc_yy"
%debug
%verbose
%initial-action {
    yylloc.line = 1;
    yylloc.column = 1;
    yylloc.offset = 0;
}

%union {
    int ival;
    char *str;
}

/* identifier */
%token  <str>           ID

/* expr_lit */
%token  <str>           L_FLOAT         L_INT           L_STR

/* operator */
%token  OP_ADD_ASSIGN   OP_SUB_ASSIGN   OP_MUL_ASSIGN   OP_DIV_ASSIGN
        OP_MOD_ASSIGN   OP_AND_ASSIGN   OP_XOR_ASSIGN   OP_OR_ASSIGN
        OP_RSHIFT       OP_LSHIFT       OP_INC          OP_DEC
        OP_AND          OP_OR           OP_LE           OP_GE
        OP_EQ           OP_NE

/* keyword */
%token  /* A */
        K_ACCOUNT
        /* B */
        K_BOOL          K_BREAK         K_BYTE          
        /* C */
        K_CASE          K_CONST         K_CONSTRUCTOR   K_CONTINUE      
        K_CONTRACT      
        /* D */
        K_DEFAULT       K_DOUBLE        
        /* E */
        K_ELSE
        /* F */
        K_FALSE         K_FILE          K_FLOAT         K_FOR           
        K_FUNC
        /* G */
        /* H */
        /* I */
        K_IF            K_INT           K_INT16         K_INT32         
        K_INT64         K_INT8          
        /* L */
        /* M */
        K_MAP           
        /* N */
        K_NEW           K_NULL
        /* O */
        /* P */
        K_PAYABLE       K_PRAGMA        K_PRIVATE       K_PUBLIC
        /* Q */
        /* R */
        K_RETURN        
        /* S */
        K_STRING        K_STRUCT        K_SWITCH        
        /* T */
        K_TRUE
        /* U */
        K_UINT          K_UINT16        K_UINT32        K_UINT64        
        K_UINT8         
        /* V */
        K_VERSION
        /* W */
        K_WHILE
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
%type   <str>           id

%start  smart_contract

%%

smart_contract: 
    /* empty */
|   smart_contract pragma_decl
|   smart_contract contract_decl
;

pragma_decl: 
    K_PRAGMA directive
;

directive: 
    K_VERSION L_STR
;

contract_decl: 
    K_CONTRACT id '{' contract_body '}'
;

contract_body: 
    decl_list_opt constructor_opt func_decl_list_opt
;

decl_list_opt: 
    /* empty */
|   decl_list
;

decl_list: 
    prim_decl 
|   decl_list prim_decl
|   struct_decl 
|   decl_list struct_decl
;

prim_decl:
    prim_spec ';'
|   prim_spec init_decl_list ';'
;

prim_spec:
    type_spec
|   type_spec prim_spec
|   type_qual
|   type_qual prim_spec
;

type_spec: 
    K_ACCOUNT
|   K_BOOL
|   K_BYTE
|   K_FLOAT
|   K_DOUBLE
|   K_INT8
|   K_INT16
|   K_INT32
|   K_INT64
|   K_MAP
|   K_STRING
|   K_UINT8
|   K_UINT16
|   K_UINT32
|   K_UINT64
;

init_decl_list:
    init_decl
|   init_decl_list ',' init_decl
;

init_decl:
    var_decl
|   var_decl '=' initializer
;

var_decl:
    id
|   '(' var_decl ')'
|   var_decl '[' expr_add ']'
;

initializer:
    expr_cond
|   '{' init_list '}'
;

init_list:
;

struct_decl: 
    K_STRUCT id '{' struct_decl_list '}'
|   K_STRUCT id
;

struct_decl_list:
    prim_decl
|   struct_decl_list prim_decl
;

/*
decl_list: 
    prim_decl 
|   decl_list prim_decl 
|   struct_decl
|   decl_list struct_decl
;

prim_decl: 
    type_decl id ';'
|   K_CONST type_decl id ';'
;

type_decl: 
    type_spec 
|   type_spec '[' ']'
|   type_spec '[' L_INT ']'
;

type_spec: 
    K_ACCOUNT
|   K_BOOL
|   K_BYTE
|   K_FLOAT
|   K_DOUBLE
|   K_INT8
|   K_INT16
|   K_INT32
|   K_INT64
|   K_MAP
|   K_STRING
|   K_UINT8
|   K_UINT16
|   K_UINT32
|   K_UINT64
;

struct_decl: 
    K_STRUCT id '{' struct_decl_list '}'
;

struct_decl_list: 
    prim_decl
|   struct_decl_list prim_decl
;
*/

constructor_opt: 
    /* empty */
|   constructor
;

constructor: 
    K_CONSTRUCTOR '(' param_list_opt ')' stmt_block
;

param_list_opt: 
    /* empty */
|   param_list
;

param_list: 
    prim_decl
|   param_list ',' prim_decl
;

func_decl_list_opt: 
    /* empty */
|   func_decl_list
;

func_decl_list: 
    func_decl
|   func_decl_list func_decl
;

func_decl: 
    modifier_opt K_FUNC id '(' param_list_opt ')' return_type_opt stmt_block
;

modifier_opt: 
    /* empty */
|   modifier_opt K_PAYABLE
|   modifier_opt K_PRIVATE
|   modifier_opt K_PUBLIC
;

return_type_opt: 
    /* empty */
|   type_decl
|   '(' return_type_list ')'
;

return_type_list: 
    type_decl
|   return_type_list ',' type_decl
;

statement: 
    stmt_expr
|   stmt_if
|   stmt_for
|   stmt_switch
|   stmt_label
|   stmt_jump
|   stmt_block
;

stmt_expr: 
    ';'
|   expression ';'
;

stmt_if: 
    K_IF '(' expression ')' statement
|   K_IF '(' expression ')' statement K_ELSE statement
;

stmt_for: 
    K_FOR '(' stmt_expr stmt_expr ')' statement
|   K_FOR '(' stmt_expr stmt_expr expression ')' statement
|   K_FOR '(' prim_decl stmt_expr ')' statement
|   K_FOR '(' prim_decl stmt_expr expression ')' statement
;

stmt_switch: 
    K_SWITCH '(' expression ')' statement
;

stmt_label: 
    K_CASE expr_const ':' statement
|   K_DEFAULT ':' statement
;

stmt_jump: 
    K_CONTINUE ';'
|   K_BREAK ';'
|   K_RETURN ';'
|   K_RETURN expression ';'
;

stmt_block: 
    '{' '}'
|   '{' block_item_list '}'
;

block_item_list: 
    block_item
|   block_item_list block_item
;

block_item: 
    prim_decl
|   struct_decl
|   statement
;

expression: 
    expr_assign
|   expression ',' expr_assign
;

expr_assign: 
    expr_cond 
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
;

expr_post:
    expr_prim
|   expr_post '[' expression ']'
|   expr_post '(' ')'
|   expr_post '(' expr_list ')'
|   expr_post '.' id
|   expr_post OP_INC
|   expr_post OP_DEC
;

expr_prim:
    K_NULL
|   K_TRUE
|   K_FALSE
|   L_INT
|   L_FLOAT
|   L_STR
|   id
|   '(' expression ')'
;

expr_list: 
    expr_assign
|   expr_list ',' expr_assign
;

expr_const:
    expr_cond
;

id: 
    ID              { $$ = $1; }
|   K_VERSION       { $$ = strdup("VERSION"); }
;

%%

static void
sc_yyerror(YYLTYPE *lloc, sc_yacc_t *yacc, const char *msg)
{
}

/* end of sc_grammar.y */
