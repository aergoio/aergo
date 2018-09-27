%{

/**
 * @file    grammar.y
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "error.h"
#include "parse.h"

#define YYLLOC_DEFAULT(Current, Rhs, N)                                        \
    (Current) = YYRHSLOC(Rhs, (N) > 0 ? 1 : 0)

extern int yylex(YYSTYPE *lval, YYLTYPE *lloc, void *yyscanner);
extern void yylex_set_token(void *yyscanner, int token, YYLTYPE *lloc);

static void yyerror(YYLTYPE *lloc, yyparam_t *penv, void *scanner,
                    const char *msg);

%}

%parse-param { yyparam_t *penv }
%param { void *yyscanner }
%locations
%debug
%verbose
%define api.pure full
%define parse.error verbose
%initial-action {
    errpos_init(&yylloc, penv->path);
}

/* identifier */
%token  <str>
        ID              "identifier"

/* literal */
%token  <str>
        L_FLOAT         "floating-point"
        L_HEXA          "hexadecimal"
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

%token  END 0           "EOF"

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
%union {
    bool flag;
    char *str;

    type_t type;
    scope_t scope;
    op_kind_t op;
    sql_kind_t sql;
    ddl_kind_t ddl;
    modifier_t mod;

    list_t *list;

    ast_var_t *var;
    ast_struct_t *struc;
    ast_blk_t *blk;
    ast_exp_t *exp;
    ast_stmt_t *stmt;
    ast_func_t *fn;
}

%type <blk>     contract_decl
%type <blk>     contract_body
%type <list>    variable
%type <exp>     var_type
%type <exp>     var_spec
%type <scope>   var_scope
%type <type>    prim_type
%type <list>    var_decl_list
%type <var>     var_decl
%type <exp>     declarator
%type <exp>     initializer
%type <list>    init_list
%type <struc>   struct
%type <list>    field_list
%type <fn>      constructor
%type <list>    param_list_opt
%type <list>    param_list
%type <var>     param_decl
%type <blk>     block
%type <blk>     blk_decl
%type <fn>      function
%type <mod>     modifier_opt
%type <list>    return_opt
%type <list>    return_list
%type <stmt>    statement
%type <stmt>    stmt_exp
%type <stmt>    stmt_if
%type <stmt>    stmt_loop
%type <exp>     exp_loop
%type <stmt>    stmt_switch
%type <list>    label_list
%type <stmt>    label
%type <list>    stmt_list
%type <stmt>    stmt_jump
%type <stmt>    stmt_ddl
%type <stmt>    stmt_blk
%type <ddl>     ddl_prefix
%type <exp>     expression
%type <exp>     exp_assign
%type <op>      op_assign
%type <exp>     exp_sql
%type <sql>     sql_prefix
%type <exp>     exp_cond
%type <exp>     exp_or
%type <exp>     exp_and
%type <exp>     exp_bit_or
%type <exp>     exp_bit_xor
%type <exp>     exp_bit_and
%type <exp>     exp_eq
%type <exp>     exp_cmp
%type <op>      op_cmp
%type <exp>     exp_shift
%type <exp>     exp_add
%type <exp>     exp_mul
%type <exp>     exp_unary
%type <exp>     exp_post
%type <exp>     exp_prim
%type <exp>     exp_new
%type <list>    exp_list
%type <str>     non_reserved_token
%type <str>     identifier

%start  smart_contract

%%

smart_contract:
    contract_decl
    {
        *penv->ast = ast_new();
        list_add_tail(&(*penv->ast)->blk_l, $1);
    }
|   smart_contract contract_decl
    {
        list_add_tail(&(*penv->ast)->blk_l, $2);
    }
;

contract_decl:
    K_CONTRACT identifier '{' '}'
    {
        $$ = ast_blk_new(&@$);
        $$->name = $2;
    }
|   K_CONTRACT identifier '{' contract_body '}'
    {
        $$ = $4;
        $$->name = $2;
    }
;

contract_body:
    variable
    {
        $$ = ast_blk_new(&@$);
        list_join(&$$->var_l, $1);
    }
|   struct
    {
        $$ = ast_blk_new(&@$);
        list_add_tail(&$$->struct_l, $1);
    }
|   constructor
    {
        $$ = ast_blk_new(&@$);
        list_add_tail(&$$->func_l, $1);
    }
|   function
    {
        $$ = ast_blk_new(&@$);
        list_add_tail(&$$->func_l, $1);
    }
|   contract_body variable
    {
        $$ = $1;
        list_join(&$$->var_l, $2);
    }
|   contract_body struct
    {
        $$ = $1;
        list_add_tail(&$$->struct_l, $2);
    }
|   contract_body constructor
    {
        $$ = $1;
        list_add_tail(&$$->func_l, $2);
    }
|   contract_body function
    {
        $$ = $1;
        list_add_tail(&$$->func_l, $2);
    }
;

variable:
    var_type var_decl_list ';'
    {
        list_node_t *node;
        list_foreach(node, $2) {
            ((ast_var_t *)node->item)->type_exp = $1;
        }
        $$ = $2;
    }
;

var_type:
    var_spec
    {
        $$ = $1;
    }
|   K_CONST var_spec
    {
        $$ = $2;
        $$->meta.is_const = true;
    }
|   var_scope var_type
    {
        $$ = $2;
        $$->meta.scope = $1;
    }
;

var_scope:
    K_LOCAL             { $$ = SCOPE_LOCAL; }
|   K_SHARED            { $$ = SCOPE_SHARED; }
;

var_spec:
    prim_type
    {
        $$ = exp_type_new($1, NULL, NULL, NULL, &@$);
    }
|   identifier
    {
        $$ = exp_type_new(TYPE_STRUCT, $1, NULL, NULL, &@$);
    }
|   K_MAP '(' var_spec ',' var_spec ')'
    {
        $$ = exp_type_new(TYPE_MAP, NULL, $3, $5, &@$);
    }
;

prim_type:
    K_ACCOUNT           { $$ = TYPE_ACCOUNT; }
|   K_BOOL              { $$ = TYPE_BOOL; }
|   K_BYTE              { $$ = TYPE_BYTE; }
|   K_FLOAT             { $$ = TYPE_FLOAT; }
|   K_DOUBLE            { $$ = TYPE_DOUBLE; }
|   K_INT               { $$ = TYPE_INT32; }
|   K_INT16             { $$ = TYPE_INT16; }
|   K_INT32             { $$ = TYPE_INT32; }
|   K_INT64             { $$ = TYPE_INT64; }
|   K_STRING            { $$ = TYPE_STRING; }
|   K_UINT              { $$ = TYPE_UINT32; }
|   K_UINT16            { $$ = TYPE_UINT16; }
|   K_UINT32            { $$ = TYPE_UINT32; }
|   K_UINT64            { $$ = TYPE_UINT64; }
;

var_decl_list:
    var_decl
    {
        $$ = list_new();
        list_add_tail($$, $1);
    }
|   var_decl_list ',' var_decl
    {
        $$ = $1;
        list_add_tail($$, $3);
    }
;

var_decl:
    declarator
    {
        $$ = ast_var_new(NULL, $1, NULL, &@$);
    }
|   declarator '=' initializer
    {
        $$ = ast_var_new(NULL, $1, $3, &@$);
    }
;

declarator:
    identifier
    {
        $$ = exp_id_ref_new($1, &@1);
    }
|   declarator '[' exp_add ']'
    {
        $$ = exp_array_new($1, $3, &@$);
    }
|   declarator '[' ']'
    {
        $$ = exp_array_new($1, NULL, &@$);
    }
;

initializer:
    exp_sql
|   '{' init_list '}'
    {
        $$ = exp_tuple_new(NULL, &@$);
        $$->u_tuple.exp_l = $2;
    }
|   '{' init_list ',' '}'
    {
        $$ = exp_tuple_new(NULL, &@$);
        $$->u_tuple.exp_l = $2;
    }
;

init_list:
    initializer
    {
        $$ = list_new();
        list_add_tail($$, $1);
    }
|   init_list ',' initializer
    {
        $$ = $1;
        list_add_tail($$, $3);
    }
;

struct:
    K_STRUCT identifier '{' field_list '}'
    {
        $$ = ast_struct_new($2, $4, &@$);
    }
;

field_list:
    variable
    {
        $$ = $1;
    }
|   field_list variable
    {
        $$ = $1;
        list_join($$, $2);
    }
;

constructor:
    identifier '(' param_list_opt ')' block
    {
        $$ = ast_func_new($1, MOD_GLOBAL, $3, NULL, $5, &@$);
    }
;

param_list_opt:
    /* empty */             { $$ = NULL; }
|   param_list
;

param_list:
    param_decl
    {
        $$ = list_new();
        list_add_tail($$, $1);
    }
|   param_list ',' param_decl
    {
        $$ = $1;
        list_add_tail($$, $3);
    }
;

param_decl:
    var_type declarator
    {
        $$ = ast_var_new($1, $2, NULL, &@$);
    }
;

block:
    '{' '}'             { $$ = NULL; }
|   '{' blk_decl '}'    { $$ = $2; }
;

blk_decl:
    variable
    {
        $$ = ast_blk_new(&@$);
        list_join(&$$->var_l, $1);
    }
|   struct
    {
        $$ = ast_blk_new(&@$);
        list_add_tail(&$$->struct_l, $1);
    }
|   statement
    {
        $$ = ast_blk_new(&@$);
        list_add_tail(&$$->stmt_l, $1);
    }
|   blk_decl variable
    {
        $$ = $1;
        list_join(&$$->var_l, $2);
    }
|   blk_decl struct
    {
        $$ = $1;
        list_add_tail(&$$->struct_l, $2);
    }
|   blk_decl statement
    {
        $$ = $1;
        list_add_tail(&$$->stmt_l, $2);
    }
;

function:
    modifier_opt K_FUNC identifier '(' param_list_opt ')' return_opt block
    {
        $$ = ast_func_new($3, $1, $5, $7, $8, &@$);
    }
;

modifier_opt:
    /* empty */                 { $$ = MOD_GLOBAL; }
|   K_LOCAL                     { $$ = MOD_LOCAL; }
|   K_SHARED                    { $$ = MOD_SHARED; }
|   modifier_opt K_READONLY     { flag_set($$, MOD_READONLY); }
|   modifier_opt K_TRANSFER     { flag_set($$, MOD_TRANSFER); }
;

return_opt:
    /* empty */                 { $$ = NULL; }
|   '(' return_list ')'         { $$ = $2; }
|   return_list 
;

return_list:
    var_type
    {
        $$ = list_new();
        list_add_tail($$, $1);
    }
|   return_list ',' var_type
    {
        $$ = $1;
        list_add_tail($$, $3);
    }
;

statement:
    stmt_exp
|   stmt_if
|   stmt_loop
|   stmt_switch
|   stmt_jump
|   stmt_ddl
|   stmt_blk
;

stmt_exp:
    ';'
    {
        $$ = ast_stmt_new(STMT_NULL, &@$);
    }
|   expression ';'
    {
        $$ = stmt_exp_new($1, &@$);
    }
;

stmt_if:
    K_IF '(' expression ')' block
    {
        $$ = stmt_if_new($3, $5, &@$);
    }
|   stmt_if K_ELSE K_IF '(' expression ')' block
    {
        $$ = $1;
        list_add_tail(&$$->u_if.elsif_l, stmt_if_new($5, $7, &@2));
    }
|   stmt_if K_ELSE block
    {
        $$ = $1;
        $$->u_if.else_blk = $3;
    }
;

stmt_loop:
    K_FOR block
    {
        $$ = stmt_for_new(NULL, NULL, NULL, $2, &@$);
    }
|   K_FOR '(' exp_or ')' block
    {
        $$ = stmt_for_new($3, NULL, NULL, $5, &@$);
    }
|   K_FOR '(' exp_loop exp_loop ')' block
    {
        $$ = stmt_for_new($3, $4, NULL, $6, &@$);
    }
|   K_FOR '(' exp_loop exp_loop expression ')' block
    {
        $$ = stmt_for_new($3, $4, $5, $7, &@$);
    }
|   K_FOR '(' variable exp_loop ')' block
    {
        $$ = stmt_for_new(NULL, $4, NULL, $6, &@$);
        $$->u_for.init_l = $3;
    }
|   K_FOR '(' variable exp_loop expression ')' block
    {
        $$ = stmt_for_new(NULL, $4, $5, $7, &@$);
        $$->u_for.init_l = $3;
    }
|   K_FOR '(' exp_loop K_IN exp_post ')' block
    {
        $$ = NULL;
    }
|   K_FOR '(' variable K_IN exp_post ')' block
    {
        $$ = NULL;
    }
|   K_FOREACH '(' iter_decl ',' iter_decl K_IN exp_post ')' block
    {
        $$ = NULL;
    }
;

exp_loop:
    ';'                     { $$ = NULL; }
|   expression ';'          { $$ = $1; }
;

iter_decl:
    var_type declarator
;

stmt_switch:
    K_SWITCH '{' label_list '}'
    {
        $$ = stmt_switch_new(NULL, $3, &@$);
    }
|   K_SWITCH '(' expression ')' '{' label_list '}'
    {
        $$ = stmt_switch_new($3, $6, &@$);
    }
;

label_list:
    label
    {
        $$ = list_new();
        list_add_tail($$, $1);
    }
|   label_list label
    {
        $$ = $1;
        list_add_tail($$, $2);
    }
;

label:
    K_CASE exp_eq ':' stmt_list
    {
        $$ = stmt_case_new($2, $4, &@$);
    }
|   K_DEFAULT ':' stmt_list
    {
        $$ = stmt_case_new(NULL, $3, &@$);
    }
;

stmt_list:
    statement
    {
        $$ = list_new();
        list_add_tail($$, $1);
    }
|   stmt_list statement
    {
        $$ = $1;
        list_add_tail($$, $2);
    }
;

stmt_jump:
    K_CONTINUE ';'
    {
        $$ = ast_stmt_new(STMT_CONTINUE, &@$);
    }
|   K_BREAK ';'
    {
        $$ = ast_stmt_new(STMT_BREAK, &@$);
    }
|   K_RETURN ';'
    {
        $$ = stmt_return_new(NULL, &@$);
    }
|   K_RETURN expression ';'
    {
        $$ = stmt_return_new($2, &@$);
    }
;

stmt_ddl:
    ddl_prefix error ';'
    {
        int len;
        char *ddl;
        yyerrok;
        error_pop();
        len = @$.last_offset - @$.first_offset;
        ddl = xstrndup(penv->src + @$.first_offset, len);
        $$ = stmt_ddl_new($1, ddl, &@$);
        yylex_set_token(yyscanner, ';', &@3);
        yyclearin;
    }
;

ddl_prefix:
    K_CREATE K_INDEX        { $$ = DDL_CREATE_IDX; }
|   K_CREATE K_TABLE        { $$ = DDL_CREATE_TBL; }
|   K_DROP K_INDEX          { $$ = DDL_DROP_IDX; }
|   K_DROP K_TABLE          { $$ = DDL_DROP_TBL; }
;

stmt_blk:
    block
    {
        $$ = stmt_blk_new($1, &@$);
    }
;

expression:
    exp_assign
    {
        $$ = exp_tuple_new($1, &@$);
    }
|   expression ',' exp_assign
    {
        $$ = $1;
        list_add_tail($$->u_tuple.exp_l, $3);
    }
;

exp_assign:
    exp_sql
|   exp_unary op_assign exp_assign
    {
        if ($2 == OP_ASSIGN) {
            $$ = exp_op_new($2, $1, $3, &@$);
        }
        else {
            ast_exp_t *op_exp = exp_op_new($2, $1, $3, &@2);

            $$ = exp_op_new(OP_ASSIGN, $1, op_exp, &@$);
        }
    }
;

op_assign:
    '='                 { $$ = OP_ASSIGN; }
|   ASSIGN_ADD          { $$ = OP_ADD; }
|   ASSIGN_SUB          { $$ = OP_SUB; }
|   ASSIGN_MUL          { $$ = OP_MUL; }
|   ASSIGN_DIV          { $$ = OP_DIV; }
|   ASSIGN_MOD          { $$ = OP_MOD; }
|   ASSIGN_AND          { $$ = OP_BIT_AND; }
|   ASSIGN_XOR          { $$ = OP_BIT_XOR; }
|   ASSIGN_OR           { $$ = OP_BIT_OR; }
|   ASSIGN_RS           { $$ = OP_RSHIFT; }
|   ASSIGN_LS           { $$ = OP_LSHIFT; }
;

exp_sql:
    exp_cond
|   sql_prefix error ';'
    {
        int len;
        char *sql;
        yyerrok;
        error_pop();
        len = @$.last_offset - @$.first_offset;
        sql = xstrndup(penv->src + @$.first_offset, len);
        $$ = exp_sql_new($1, sql, &@$);
        yylex_set_token(yyscanner, ';', &@3);
        yyclearin;
    }
;

sql_prefix:
    K_DELETE            { $$ = SQL_DELETE; }
|   K_INSERT            { $$ = SQL_INSERT; }
|   K_SELECT            { $$ = SQL_QUERY; }
|   K_UPDATE            { $$ = SQL_UPDATE; }
;

exp_cond:
    exp_or
|   exp_or '?' exp_sql ':' exp_cond
    {
        $$ = exp_cond_new($1, $3, $5, &@$);
    }
;

exp_or:
    exp_and
|   exp_or CMP_OR exp_and
    {
        $$ = exp_op_new(OP_OR, $1, $3, &@$);
    }
;

exp_and:
    exp_bit_or
|   exp_and CMP_AND exp_bit_or
    {
        $$ = exp_op_new(OP_AND, $1, $3, &@$);
    }
;

exp_bit_or:
    exp_bit_xor
|   exp_bit_or '|' exp_bit_xor
    {
        $$ = exp_op_new(OP_BIT_OR, $1, $3, &@$);
    }
;

exp_bit_xor:
    exp_bit_and
|   exp_bit_xor '^' exp_bit_and
    {
        $$ = exp_op_new(OP_BIT_XOR, $1, $3, &@$);
    }
;

exp_bit_and:
    exp_eq
|   exp_bit_and '&' exp_eq
    {
        $$ = exp_op_new(OP_BIT_AND, $1, $3, &@$);
    }
;

exp_eq:
    exp_cmp
|   exp_eq CMP_EQ exp_cmp
    {
        $$ = exp_op_new(OP_EQ, $1, $3, &@$);
    }
|   exp_eq CMP_NE exp_cmp
    {
        $$ = exp_op_new(OP_NE, $1, $3, &@$);
    }
;

exp_cmp:
    exp_shift
|   exp_cmp op_cmp exp_shift
    {
        $$ = exp_op_new($2, $1, $3, &@$);
    }
;

op_cmp:
    '<'             { $$ = OP_LT; }
|   '>'             { $$ = OP_GT; }
|   CMP_LE          { $$ = OP_LE; }
|   CMP_GE          { $$ = OP_GE; }
;

exp_shift:
    exp_add
|   exp_shift SHIFT_R exp_add
    {
        $$ = exp_op_new(OP_RSHIFT, $1, $3, &@$);
    }
|   exp_shift SHIFT_L exp_add
    {
        $$ = exp_op_new(OP_LSHIFT, $1, $3, &@$);
    }
;

exp_add:
    exp_mul
|   exp_add '+' exp_mul
    {
        $$ = exp_op_new(OP_ADD, $1, $3, &@$);
    }
|   exp_add '-' exp_mul
    {
        $$ = exp_op_new(OP_SUB, $1, $3, &@$);
    }
;

exp_mul:
    exp_unary
|   exp_mul '*' exp_unary
    {
        $$ = exp_op_new(OP_MUL, $1, $3, &@$);
    }
|   exp_mul '/' exp_unary
    {
        $$ = exp_op_new(OP_DIV, $1, $3, &@$);
    }
|   exp_mul '%' exp_unary
    {
        $$ = exp_op_new(OP_MOD, $1, $3, &@$);
    }
;

exp_unary:
    exp_post
|   UNARY_INC exp_unary
    {
        $$ = exp_op_new(OP_INC, $2, NULL, &@$);
    }
|   UNARY_DEC exp_unary
    {
        $$ = exp_op_new(OP_DEC, $2, NULL, &@$);
    }
|   '!' exp_unary
    {
        $$ = exp_op_new(OP_NOT, $2, NULL, &@$);
    }
;

exp_post:
    exp_prim
|   exp_post '[' expression ']'
    {
        $$ = exp_array_new($1, $3, &@$);
    }
|   exp_post '(' ')'
    {
        $$ = exp_call_new($1, NULL, &@$);
    }
|   exp_post '(' exp_list ')'
    {
        $$ = exp_call_new($1, $3, &@$);
    }
|   exp_post '.' identifier
    {
        ast_exp_t *id_exp = exp_id_ref_new($3, &@3);
        $$ = exp_access_new($1, id_exp, &@$);
    }
|   exp_post UNARY_INC
    {
        $$ = exp_op_new(OP_INC, $1, NULL, &@$);
    }
|   exp_post UNARY_DEC
    {
        $$ = exp_op_new(OP_DEC, $1, NULL, &@$);
    }
;

exp_prim:
    exp_new
|   K_NULL
    {
        $$ = exp_lit_new(LIT_NULL, NULL, &@$);
    }
|   K_TRUE
    {
        $$ = exp_lit_new(LIT_BOOL, INT2PTR(1), &@$);
    }
|   K_FALSE
    {
        $$ = exp_lit_new(LIT_BOOL, INT2PTR(0), &@$);
    }
|   L_INT
    {
        $$ = exp_lit_new(LIT_INT, $1, &@$);
    }
|   L_FLOAT
    {
        $$ = exp_lit_new(LIT_FLOAT, $1, &@$);
    }
|   L_HEXA
    {
        $$ = exp_lit_new(LIT_HEXA, $1, &@$);
    }
|   L_STR
    {
        $$ = exp_lit_new(LIT_STR, $1, &@$);
    }
|   identifier
    {
        $$ = exp_id_ref_new($1, &@$);
    }
|   '(' expression ')'
    {
        $$ = $2;
    }
;

exp_new:
    K_NEW identifier '(' ')'
    {
        ast_exp_t *id_exp = exp_id_ref_new($2, &@2);
        $$ = exp_call_new(id_exp, NULL, &@$);
    }
|   K_NEW identifier '(' exp_list ')'
    {
        ast_exp_t *id_exp = exp_id_ref_new($2, &@2);
        $$ = exp_call_new(id_exp, $4, &@$);
    }
|   K_NEW K_MAP '(' ')'
    {
        ast_exp_t *id_exp = exp_id_ref_new(xstrdup("map"), &@2);
        $$ = exp_call_new(id_exp, NULL, &@$);
    }
|   K_NEW K_MAP '(' L_INT ')'
    {
        list_t *exp_l = list_new();
        ast_exp_t *id_exp = exp_id_ref_new(xstrdup("map"), &@2);
        ast_exp_t *size_exp = exp_lit_new(LIT_INT, $4, &@4);
        list_add_tail(exp_l, size_exp);
        $$ = exp_call_new(id_exp, exp_l, &@$);
    }
;

exp_list:
    exp_cond
    {
        $$ = list_new();
        list_add_tail($$, $1);
    }
|   exp_list ',' exp_cond
    {
        $$ = $1;
        list_add_tail($$, $3);
    }
;

non_reserved_token:
    K_CONTRACT          { $$ = xstrdup("contract"); }
|   K_INDEX             { $$ = xstrdup("index"); }
|   K_TABLE             { $$ = xstrdup("table"); }
;

identifier:
    ID
|   non_reserved_token
;

%%

static void
yyerror(YYLTYPE *lloc, yyparam_t *penv, void *scanner, const char *msg)
{
    TRACE(ERROR_SYNTAX, lloc, msg);
}

/* end of grammar.y */
