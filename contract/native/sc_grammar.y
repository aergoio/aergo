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
%token  ID

/* literal */
%token  L_FLOAT         L_INT           L_STR

/* operator */
%token  OP_ADD_ASSIGN   OP_SUB_ASSIGN   OP_MUL_ASSIGN   OP_DIV_ASSIGN
        OP_MOD_ASSIGN   OP_AND_ASSIGN   OP_XOR_ASSIGN   OP_OR_ASSIGN
        OP_RSHIFT       OP_LSHIFT       OP_INC          OP_DEC
        OP_AND          OP_OR           OP_LE           OP_GE
        OP_EQ           OP_NE

/* keyword */
%token  K_BREAK         K_BYTE          K_CASE          K_CONST         
        K_CONTINUE      K_CONTRACT      K_DOUBLE        K_ELSE
        K_FLOAT         K_FOR           K_IF            K_INT           
        K_INT16         K_INT32         K_INT64         K_INT8          
        K_NEW           K_PRIVATE       K_PUBLIC        K_RETURN        
        K_STRING        K_STRUCT        K_SWITCH        K_UINT          
        K_UINT16        K_UINT32        K_UINT64        K_UINT8         
        K_WHILE

/* precedences */
%left   OP_OR
%left   OP_AND
%left   '|'
%left   '&'
%left   OP_EQ OP_NE
%left   OP_LE OP_GE '<' '>'
%left   '+' '-' 
%left   '*' '/'

%%

start:
    /* empty */
;

%%

static void
sc_yyerror(YYLTYPE *lloc, sc_yacc_t *yacc, const char *msg)
{
}

/* end of sc_grammar.y */
