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

#define AST             (*parse->ast)
#define ROOT            AST->root

extern int yylex(YYSTYPE *yylval, YYLTYPE *yylloc, void *yyscanner);
extern void yylex_set_token(void *yyscanner, int token, YYLTYPE *yylloc);

static void yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner,
                    const char *msg);

%}

%parse-param { parse_t *parse }
%param { void *yyscanner }
%locations
%debug
%verbose
%define api.pure full
%define parse.error verbose
%initial-action {
    src_pos_init(&yylloc, parse->src, parse->path);
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
        K_FUNC          "func"
        K_GOTO          "goto"
        K_IF            "if"
        K_IN            "in"
        K_INDEX         "index"
        K_INSERT        "insert"
        K_INT           "int"
        K_INT16         "int16"
        K_INT32         "int32"
        K_INT64         "int64"
        K_INT8          "int8"
        K_LOCAL         "local"
        K_MAP           "map"
        K_NEW           "new"
        K_NULL          "null"
        K_PAYABLE       "payable"
        K_READONLY      "readonly"
        K_RETURN        "return"
        K_SELECT        "select"
        K_STRING        "string"
        K_STRUCT        "struct"
        K_SWITCH        "switch"
        K_TABLE         "table"
        K_TRUE          "true"
        K_UINT          "uint"
        K_UINT16        "uint16"
        K_UINT32        "uint32"
        K_UINT64        "uint64"
        K_UINT8         "uint8"
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
%left   '(' ')' '.'
%right  UNARY_INC UNARY_DEC

/* type */
%union {
    bool flag;
    char *str;
    array_t *array;

    type_t type;
    op_kind_t op;
    sql_kind_t sql;
    modifier_t mod;

    ast_id_t *id;
    ast_blk_t *blk;
    ast_exp_t *exp;
    ast_stmt_t *stmt;
}

%type <id>      contract_decl
%type <blk>     contract_body
%type <array>   variable
%type <array>   var_decl
%type <array>   var_init_decl
%type <exp>     var_type
%type <exp>     var_spec
%type <type>    prim_type
%type <array>   var_name_list
%type <id>      declarator
%type <array>   var_init_list
%type <exp>     initializer
%type <array>   elem_list
%type <id>      compound
%type <id>      struct
%type <array>   field_list
%type <id>      constructor
%type <array>   param_list_opt
%type <array>   param_list
%type <id>      param_decl
%type <blk>     block
%type <blk>     blk_decl
%type <id>      function
%type <mod>     modifier_opt
%type <array>   return_opt
%type <array>   return_list
%type <stmt>    statement
%type <stmt>    exp_stmt
%type <stmt>    label_stmt
%type <stmt>    if_stmt
%type <stmt>    loop_stmt
%type <exp>     init_exp
%type <exp>     cond_exp
%type <stmt>    switch_stmt
%type <array>   case_list
%type <stmt>    case_stmt
%type <array>   stmt_list
%type <stmt>    jump_stmt
%type <stmt>    ddl_stmt
%type <stmt>    blk_stmt
%type <exp>     expression
%type <op>      assign_op
%type <exp>     tuple_exp
%type <exp>     sql_exp
%type <sql>     sql_prefix
%type <exp>     ternary_exp
%type <exp>     or_exp
%type <exp>     and_exp
%type <exp>     bit_or_exp
%type <exp>     bit_xor_exp
%type <exp>     bit_and_exp
%type <exp>     eq_exp
%type <exp>     cmp_exp
%type <op>      cmp_op
%type <exp>     shift_exp
%type <exp>     add_exp
%type <exp>     mul_exp
%type <exp>     unary_exp
%type <exp>     post_exp
%type <exp>     prim_exp
%type <exp>     new_exp
%type <array>   arg_list
%type <str>     non_reserved_token
%type <str>     identifier

%start  smart_contract

%%

smart_contract:
    contract_decl
    {
        AST = ast_new();
        id_add_last(&ROOT->ids, $1);
    }
|   smart_contract contract_decl
    {
        id_add_last(&ROOT->ids, $2);
    }
;

contract_decl:
    K_CONTRACT identifier '{' '}'
    {
        ast_blk_t *blk = blk_anon_new(&@3);

        id_add_last(&blk->ids, id_ctor_new($2, NULL, NULL, &@2));

        $$ = id_contract_new($2, blk, &@$);
    }
|   K_CONTRACT identifier '{' contract_body '}'
    {
        int i;
        bool exist_ctor = false;

        for (i = 0; i < array_size(&$4->ids); i++) {
            ast_id_t *id = array_item(&$4->ids, i, ast_id_t);

            if (is_ctor_id(id)) {
                if (strcmp($2, id->name) != 0)
                    ERROR(ERROR_SYNTAX, &@1, "syntax error, unexpected "
                          "identifier, expecting 'func'");
                else
                    exist_ctor = true;
            }
        }

        if (!exist_ctor)
            id_add_last(&$4->ids, id_ctor_new($2, NULL, NULL, &@2));

        $$ = id_contract_new($2, $4, &@$);
    }
;

contract_body:
    variable
    {
        $$ = blk_anon_new(&@$);
        id_join_last(&$$->ids, $1);
    }
|   compound
    {
        $$ = blk_anon_new(&@$);
        id_add_last(&$$->ids, $1);
    }
|   contract_body variable
    {
        $$ = $1;
        id_join_last(&$$->ids, $2);
    }
|   contract_body compound
    {
        $$ = $1;
        id_add_last(&$$->ids, $2);
    }
;

variable:
    var_decl ';'                { $$ = $1; }
|   var_init_decl ';'           { $$ = $1; }
|   var_type error ';'          { $$ = NULL; }
;

var_decl:
    var_type var_name_list
    {
        int i;

        for (i = 0; i < array_size($2); i++) {
            ast_id_t *id = array_item($2, i, ast_id_t);

            ASSERT1(is_var_id(id), id->kind);
            id->u_var.type_exp = $1;
        }
        $$ = $2;
    }
;

var_init_decl:
    var_type var_name_list '=' var_init_list
    {
        int i;

        if (array_size($2) != array_size($4)) {
            ERROR(ERROR_MISMATCHED_INIT_CNT, &@4, array_size($2), array_size($4));
        }
        else {
            for (i = 0; i < array_size($2); i++) {
                ast_id_t *id = array_item($2, i, ast_id_t);

                ASSERT1(is_var_id(id), id->kind);

                id->u_var.type_exp = $1;
                id->u_var.init_exp = array_item($4, i, ast_exp_t);
            }
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
|   K_LOCAL var_type
    {
        ASSERT1(is_type_exp($2), $2->kind);

        $$ = $2;
        $$->u_type.is_local = true;
    }
;

var_spec:
    prim_type
    {
        $$ = exp_type_new($1, &@$);
    }
|   identifier
    {
        $$ = exp_type_new(TYPE_STRUCT, &@$);
        $$->u_type.name = $1;
    }
|   K_MAP '(' var_spec ',' var_spec ')'
    {
        $$ = exp_type_new(TYPE_MAP, &@$);
        $$->u_type.k_exp = $3;
        $$->u_type.v_exp = $5;
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
|   K_INT8              { $$ = TYPE_INT8; }
|   K_STRING            { $$ = TYPE_STRING; }
|   K_UINT              { $$ = TYPE_UINT32; }
|   K_UINT16            { $$ = TYPE_UINT16; }
|   K_UINT32            { $$ = TYPE_UINT32; }
|   K_UINT64            { $$ = TYPE_UINT64; }
|   K_UINT8             { $$ = TYPE_UINT8; }
;

var_name_list:
    declarator
    {
        $$ = array_new();
        id_add_last($$, $1);
    }
|   var_name_list ',' declarator
    {
        $$ = $1;
        id_add_last($$, $3);
    }
;

declarator:
    identifier
    {
        $$ = id_var_new($1, &@1);
    }
|   declarator '[' add_exp ']'
    {
        $$ = $1;
        $$->u_var.arr_exp = $3;
    }
|   declarator '[' ']'
    {
        $$ = $1;
        $$->u_var.arr_exp = exp_null_new(&@2);
    }
;

var_init_list:
    initializer
    {
        $$ = array_new();
        exp_add_last($$, $1);
    }
|   var_init_list ',' initializer
    {
        $$ = $1;
        exp_add_last($$, $3);
    }
;

initializer:
    sql_exp
|   '{' elem_list '}'
    {
        $$ = exp_tuple_new($2, &@$);
    }
|   '{' elem_list ',' '}'
    {
        $$ = exp_tuple_new($2, &@$);
    }
;

elem_list:
    initializer
    {
        $$ = array_new();
        exp_add_last($$, $1);
    }
|   elem_list ',' initializer
    {
        $$ = $1;
        exp_add_last($$, $3);
    }
;

compound:
    struct
|   constructor
|   function
;

struct:
    K_STRUCT identifier '{' field_list '}'
    {
        $$ = id_struct_new($2, $4, &@$);
    }
|   K_STRUCT error '}'          { $$ = NULL; }
;

field_list:
    variable
    {
        $$ = $1;
    }
|   field_list variable
    {
        $$ = $1;
        id_join_last($$, $2);
    }
;

constructor:
    identifier '(' param_list_opt ')' block
    {
        $$ = id_ctor_new($1, $3, $5, &@$);
    }
;

param_list_opt:
    /* empty */             { $$ = NULL; }
|   param_list
;

param_list:
    param_decl
    {
        $$ = array_new();
        exp_add_last($$, $1);
    }
|   param_list ',' param_decl
    {
        $$ = $1;
        exp_add_last($$, $3);
    }
;

param_decl:
    var_type declarator
    {
        $$ = $2;
        $$->u_var.type_exp = $1;
    }
;

block:
    '{' '}'                 { $$ = NULL; }
|   '{' blk_decl '}'        { $$ = $2; }
;

blk_decl:
    variable
    {
        $$ = blk_anon_new(&@$);
        id_join_last(&$$->ids, $1);
    }
|   struct
    {
        $$ = blk_anon_new(&@$);
        id_add_last(&$$->ids, $1);
    }
|   statement
    {
        $$ = blk_anon_new(&@$);
        stmt_add_last(&$$->stmts, $1);
    }
|   blk_decl variable
    {
        $$ = $1;
        id_join_last(&$$->ids, $2);
    }
|   blk_decl struct
    {
        $$ = $1;
        id_add_last(&$$->ids, $2);
    }
|   blk_decl statement
    {
        $$ = $1;
        stmt_add_last(&$$->stmts, $2);
    }
;

function:
    modifier_opt K_FUNC identifier '(' param_list_opt ')' return_opt block
    {
        $$ = id_func_new($3, $1, $5, $7, $8, &@3);
    }
;

modifier_opt:
    /* empty */                 { $$ = MOD_GLOBAL; }
|   K_LOCAL                     { $$ = MOD_LOCAL; }
|   modifier_opt K_READONLY
    {
        $$ = $1;
        flag_set($$, MOD_READONLY);
    }
|   modifier_opt K_PAYABLE
    {
        $$ = $1;
        flag_set($$, MOD_PAYABLE);
    }
;

return_opt:
    /* empty */                 { $$ = NULL; }
|   '(' return_list ')'         { $$ = $2; }
|   return_list
;

return_list:
    var_type
    {
        $$ = array_new();
        exp_add_last($$, $1);
    }
|   return_list ',' var_type
    {
        $$ = $1;
        exp_add_last($$, $3);
    }
;

statement:
    exp_stmt
|   label_stmt
|   if_stmt
|   loop_stmt
|   switch_stmt
|   jump_stmt
|   ddl_stmt
|   blk_stmt
;

exp_stmt:
    ';'
    {
        $$ = stmt_null_new(&@$);
    }
|   expression ';'
    {
        $$ = stmt_exp_new($1, &@$);
    }
|   error ';'       { $$ = NULL; }
;

label_stmt:
    identifier ':' statement
    {
        $$ = $3;
        $$->label = $1;
    }
;

if_stmt:
    K_IF '(' expression ')' block
    {
        $$ = stmt_if_new($3, $5, &@$);
    }
|   if_stmt K_ELSE K_IF '(' expression ')' block
    {
        $$ = $1;
        stmt_add_last(&$$->u_if.elif_stmts, stmt_if_new($5, $7, &@2));
    }
|   if_stmt K_ELSE block
    {
        $$ = $1;
        $$->u_if.else_blk = $3;
    }
|   K_IF error '}'
    {
        $$ = NULL;
    }
;

loop_stmt:
    K_FOR block
    {
        $$ = stmt_loop_new(LOOP_FOR, NULL, NULL, $2, &@$);
    }
|   K_FOR '(' or_exp ')' block
    {
        $$ = stmt_loop_new(LOOP_FOR, $3, NULL, $5, &@$);
    }
|   K_FOR '(' init_exp cond_exp ')' block
    {
        $$ = stmt_loop_new(LOOP_FOR, $4, NULL, $6, &@$);
        $$->u_loop.init_exp = $3;
    }
|   K_FOR '(' init_exp cond_exp expression ')' block
    {
        $$ = stmt_loop_new(LOOP_FOR, $4, $5, $7, &@$);
        $$->u_loop.init_exp = $3;
    }
|   K_FOR '(' variable cond_exp ')' block
    {
        $$ = stmt_loop_new(LOOP_FOR, $4, NULL, $6, &@$);
        $$->u_loop.init_ids = $3;
    }
|   K_FOR '(' variable cond_exp expression ')' block
    {
        $$ = stmt_loop_new(LOOP_FOR, $4, $5, $7, &@$);
        $$->u_loop.init_ids = $3;
    }
|   K_FOR '(' expression K_IN post_exp ')' block
    {
        $$ = stmt_loop_new(LOOP_EACH, NULL, $5, $7, &@$);
        $$->u_loop.init_exp = $3;
    }
|   K_FOR '(' var_decl K_IN post_exp ')' block
    {
        $$ = stmt_loop_new(LOOP_EACH, NULL, $5, $7, &@$);
        $$->u_loop.init_ids = $3;
    }
|   K_FOR error '}'
    {
        $$ = NULL;
    }
;

init_exp:
    ';'             { $$ = NULL; }
|   expression ';'
;

cond_exp:
    ';'             { $$ = NULL; }
|   or_exp ';'
;

switch_stmt:
    K_SWITCH '{' case_list '}'
    {
        $$ = stmt_switch_new(NULL, $3, &@$);
    }
|   K_SWITCH '(' expression ')' '{' case_list '}'
    {
        $$ = stmt_switch_new($3, $6, &@$);
    }
|   K_SWITCH error '}'
    {
        $$ = NULL;
    }
;

case_list:
    case_stmt
    {
        $$ = array_new();
        stmt_add_last($$, $1);
    }
|   case_list case_stmt
    {
        $$ = $1;
        stmt_add_last($$, $2);
    }
;

case_stmt:
    K_CASE eq_exp ':' stmt_list
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
        $$ = array_new();
        stmt_add_last($$, $1);
    }
|   stmt_list statement
    {
        $$ = $1;
        stmt_add_last($$, $2);
    }
;

jump_stmt:
    K_CONTINUE ';'
    {
        $$ = stmt_jump_new(STMT_CONTINUE, &@$);
    }
|   K_BREAK ';'
    {
        $$ = stmt_jump_new(STMT_BREAK, &@$);
    }
|   K_RETURN ';'
    {
        $$ = stmt_return_new(NULL, &@$);
    }
|   K_RETURN expression ';'
    {
        $$ = stmt_return_new($2, &@$);
    }
|   K_GOTO identifier ';'
    {
        $$ = stmt_goto_new($2, &@2);
    }
;

ddl_stmt:
    ddl_prefix error ';'
    {
        int len;
        char *ddl;
        yyerrok;
        error_pop();
        len = @$.abs.last_offset - @$.abs.first_offset;
        ddl = xstrndup(parse->src + @$.abs.first_offset, len);
        $$ = stmt_ddl_new(ddl, &@$);
        yylex_set_token(yyscanner, ';', &@3);
        yyclearin;
    }
;

ddl_prefix:
    K_CREATE K_INDEX
|   K_CREATE K_TABLE
|   K_DROP K_INDEX
|   K_DROP K_TABLE
;

blk_stmt:
    block
    {
        $$ = stmt_blk_new($1, &@$);
    }
;

expression:
    tuple_exp
|   tuple_exp '=' tuple_exp
    {
        $$ = exp_op_new(OP_ASSIGN, $1, $3, &@2);
    }
|   unary_exp assign_op sql_exp
    {
        $$ = exp_op_new(OP_ASSIGN, $1, exp_op_new($2, $1, $3, &@2), &@2);
    }
;

assign_op:
    ASSIGN_ADD          { $$ = OP_ADD; }
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

tuple_exp:
    sql_exp
|   tuple_exp ',' sql_exp
    {
        if (is_tuple_exp($1)) {
            $$ = $1;
        }
        else {
            $$ = exp_tuple_new(NULL, &@$);
            exp_add_last($$->u_tup.exps, $1);
        }
        exp_add_last($$->u_tup.exps, $3);
    }
;

sql_exp:
    ternary_exp
|   sql_prefix error ';'
    {
        int len;
        char *sql;
        yyerrok;
        error_pop();
        len = @$.abs.last_offset - @$.abs.first_offset;
        sql = xstrndup(parse->src + @$.abs.first_offset, len);
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

ternary_exp:
    or_exp
|   or_exp '?' sql_exp ':' ternary_exp
    {
        $$ = exp_ternary_new($1, $3, $5, &@$);
    }
;

or_exp:
    and_exp
|   or_exp CMP_OR and_exp
    {
        $$ = exp_op_new(OP_OR, $1, $3, &@2);
    }
;

and_exp:
    bit_or_exp
|   and_exp CMP_AND bit_or_exp
    {
        $$ = exp_op_new(OP_AND, $1, $3, &@2);
    }
;

bit_or_exp:
    bit_xor_exp
|   bit_or_exp '|' bit_xor_exp
    {
        $$ = exp_op_new(OP_BIT_OR, $1, $3, &@2);
    }
;

bit_xor_exp:
    bit_and_exp
|   bit_xor_exp '^' bit_and_exp
    {
        $$ = exp_op_new(OP_BIT_XOR, $1, $3, &@2);
    }
;

bit_and_exp:
    eq_exp
|   bit_and_exp '&' eq_exp
    {
        $$ = exp_op_new(OP_BIT_AND, $1, $3, &@2);
    }
;

eq_exp:
    cmp_exp
|   eq_exp CMP_EQ cmp_exp
    {
        $$ = exp_op_new(OP_EQ, $1, $3, &@2);
    }
|   eq_exp CMP_NE cmp_exp
    {
        $$ = exp_op_new(OP_NE, $1, $3, &@2);
    }
;

cmp_exp:
    shift_exp
|   cmp_exp cmp_op shift_exp
    {
        $$ = exp_op_new($2, $1, $3, &@2);
    }
;

cmp_op:
    '<'             { $$ = OP_LT; }
|   '>'             { $$ = OP_GT; }
|   CMP_LE          { $$ = OP_LE; }
|   CMP_GE          { $$ = OP_GE; }
;

shift_exp:
    add_exp
|   shift_exp SHIFT_R add_exp
    {
        $$ = exp_op_new(OP_RSHIFT, $1, $3, &@2);
    }
|   shift_exp SHIFT_L add_exp
    {
        $$ = exp_op_new(OP_LSHIFT, $1, $3, &@2);
    }
;

add_exp:
    mul_exp
|   add_exp '+' mul_exp
    {
        $$ = exp_op_new(OP_ADD, $1, $3, &@2);
    }
|   add_exp '-' mul_exp
    {
        $$ = exp_op_new(OP_SUB, $1, $3, &@2);
    }
;

mul_exp:
    unary_exp
|   mul_exp '*' unary_exp
    {
        $$ = exp_op_new(OP_MUL, $1, $3, &@2);
    }
|   mul_exp '/' unary_exp
    {
        $$ = exp_op_new(OP_DIV, $1, $3, &@2);
    }
|   mul_exp '%' unary_exp
    {
        $$ = exp_op_new(OP_MOD, $1, $3, &@2);
    }
;

unary_exp:
    post_exp
|   UNARY_INC unary_exp
    {
        $$ = exp_op_new(OP_INC, $2, NULL, &@$);
    }
|   UNARY_DEC unary_exp
    {
        $$ = exp_op_new(OP_DEC, $2, NULL, &@$);
    }
|   '!' unary_exp
    {
        $$ = exp_op_new(OP_NOT, $2, NULL, &@$);
    }
;

post_exp:
    prim_exp
|   post_exp '[' ternary_exp ']'
    {
        $$ = exp_array_new($1, $3, &@$);
    }
|   post_exp '(' ')'
    {
        $$ = exp_call_new($1, NULL, &@$);
    }
|   post_exp '(' arg_list ')'
    {
        $$ = exp_call_new($1, $3, &@$);
    }
|   post_exp '.' identifier
    {
        $$ = exp_access_new($1, exp_id_new($3, &@3), &@$);
    }
|   post_exp UNARY_INC
    {
        $$ = exp_op_new(OP_INC, $1, NULL, &@$);
    }
|   post_exp UNARY_DEC
    {
        $$ = exp_op_new(OP_DEC, $1, NULL, &@$);
    }
;

prim_exp:
    new_exp
|   K_NULL
    {
        $$ = exp_val_new(&@$);
    }
|   K_TRUE
    {
        $$ = exp_val_new(&@$);
        value_set_bool(&$$->u_val.val, true);
    }
|   K_FALSE
    {
        $$ = exp_val_new(&@$);
        value_set_bool(&$$->u_val.val, false);
    }
|   L_INT
    {
        int64_t v;
        $$ = exp_val_new(&@$);
        sscanf($1, "%"SCNd64, &v);
        value_set_int(&$$->u_val.val, v);
    }
|   L_FLOAT
    {
        double v;
        $$ = exp_val_new(&@$);
        sscanf($1, "%lf", &v);
        value_set_double(&$$->u_val.val, v);
    }
|   L_HEXA
    {
        int64_t v;
        $$ = exp_val_new(&@$);
        sscanf($1, "%"SCNd64, &v);
        value_set_int(&$$->u_val.val, v);
    }
|   L_STR
    {
        $$ = exp_val_new(&@$);
        value_set_str(&$$->u_val.val, $1);
    }
|   identifier
    {
        $$ = exp_id_new($1, &@$);
    }
|   '(' expression ')'
    {
        $$ = $2;
    }
;

arg_list:
    ternary_exp
    {
        $$ = array_new();
        exp_add_last($$, $1);
    }
|   arg_list ',' ternary_exp
    {
        $$ = $1;
        exp_add_last($$, $3);
    }
;

new_exp:
    K_NEW identifier '(' ')'
    {
        $$ = exp_call_new(exp_id_new($2, &@2), NULL, &@$);
    }
|   K_NEW identifier '(' arg_list ')'
    {
        $$ = exp_call_new(exp_id_new($2, &@2), $4, &@$);
    }
|   K_NEW K_MAP '(' ')'
    {
        $$ = exp_call_new(exp_id_new(xstrdup("map"), &@2), NULL, &@$);
    }
|   K_NEW K_MAP '(' L_INT ')'
    {
        int64_t v;
        array_t *exps;
        ast_exp_t *size_exp;

        size_exp = exp_val_new(&@4);
        sscanf($4, "%"SCNd64, &v);
        value_set_int(&size_exp->u_val.val, v);

        exps = array_new();
        exp_add_last(exps, size_exp);

        $$ = exp_call_new(exp_id_new(xstrdup("map"), &@2), exps, &@$);
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
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
