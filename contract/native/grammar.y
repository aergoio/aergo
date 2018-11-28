%{

/**
 * @file    grammar.y
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "error.h"
#include "parse.h"

#define YYLLOC_DEFAULT(Current, Rhs, N)                                                  \
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
        L_HEX           "hexadecimal"
        L_INT           "integer"
        L_OCTAL         "octal"
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
        K_ENUM          "enum"
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
        K_MAP           "map"
        K_NEW           "new"
        K_NULL          "null"
        K_PAYABLE       "payable"
        K_PUBLIC        "public"
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
%left   '^'
%left   '&'
%left   CMP_EQ CMP_NE
%left   CMP_LE CMP_GE '<' '>'
%left   '+' '-'
%left   SHIFT_L SHIFT_R
%left   '*' '/' '%'
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
    meta_t *meta;

    struct {
        modifier_t mod;
        meta_t *meta;
    } spec;
}

%type <id>      contract_decl
%type <blk>     contract_body
%type <array>   variable
%type <array>   var_decl
%type <array>   var_init_decl
%type <spec>    var_type
%type <spec>    var_qual
%type <meta>    var_spec
%type <type>    prim_type
%type <array>   var_name_list
%type <id>      declarator
%type <exp>     size_opt
%type <array>   var_init_list
%type <id>      compound
%type <id>      struct
%type <array>   field_list
%type <id>      enumeration
%type <array>   enum_list
%type <id>      enumerator
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
%type <id>      return_decl
%type <stmt>    statement
%type <stmt>    empty_stmt
%type <stmt>    exp_stmt
%type <stmt>    assign_stmt
%type <op>      assign_op
%type <stmt>    label_stmt
%type <stmt>    if_stmt
%type <stmt>    loop_stmt
%type <stmt>    init_stmt
%type <exp>     cond_exp
%type <stmt>    switch_stmt
%type <array>   case_list
%type <stmt>    case_stmt
%type <array>   stmt_list
%type <stmt>    jump_stmt
%type <stmt>    ddl_stmt
%type <stmt>    blk_stmt
%type <exp>     expression
%type <exp>     sql_exp
%type <sql>     sql_prefix
%type <exp>     ternary_exp
%type <exp>     or_exp
%type <exp>     and_exp
%type <exp>     bit_or_exp
%type <exp>     bit_xor_exp
%type <exp>     bit_and_exp
%type <exp>     eq_exp
%type <op>      eq_op
%type <exp>     cmp_exp
%type <op>      cmp_op
%type <exp>     shift_exp
%type <op>      shift_op
%type <exp>     add_exp
%type <op>      add_op
%type <exp>     mul_exp
%type <op>      mul_op
%type <exp>     cast_exp
%type <exp>     unary_exp
%type <op>      unary_op
%type <exp>     post_exp
%type <array>   arg_list_opt
%type <array>   arg_list
%type <exp>     prim_exp
%type <exp>     literal
%type <exp>     new_exp
%type <exp>     initializer
%type <array>   elem_list
%type <exp>     init_elem
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
        ast_blk_t *blk = blk_new_anon(&@3);

        /* add default constructor */
        id_add_last(&blk->ids, id_new_ctor($2, &@2));

        $$ = id_new_contract($2, blk, &@$);
    }
|   K_CONTRACT identifier '{' contract_body '}'
    {
        int i;
        bool exist_ctor = false;

        for (i = 0; i < array_size(&$4->ids); i++) {
            ast_id_t *id = array_get(&$4->ids, i, ast_id_t);

            if (is_ctor_id(id)) {
                if (strcmp($2, id->name) != 0)
                    ERROR(ERROR_SYNTAX, &id->pos, "syntax error, unexpected "
                          "identifier, expecting func");
                else
                    exist_ctor = true;
            }
        }

        if (!exist_ctor)
            /* add default constructor */
            id_add_last(&$4->ids, id_new_ctor($2, &@2));

        $$ = id_new_contract($2, $4, &@$);
    }
;

contract_body:
    variable
    {
        $$ = blk_new_anon(&@$);
        id_join_last(&$$->ids, $1);
    }
|   compound
    {
        $$ = blk_new_anon(&@$);
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
    var_decl ';'
|   var_init_decl ';'
|   var_type error ';'
    {
        $$ = NULL;
    }
;

var_decl:
    var_type var_name_list
    {
        int i;

        for (i = 0; i < array_size($2); i++) {
            ast_id_t *id = array_get($2, i, ast_id_t);

            id->mod = $1.mod;
            id->u_var.type_meta = $1.meta;
        }
        $$ = $2;
    }
;

var_init_decl:
    var_type var_name_list '=' var_init_list
    {
        int i;

        if (array_size($2) != array_size($4)) {
            ERROR(ERROR_MISMATCHED_COUNT, &@4, "declaration", array_size($2),
                  array_size($4));
        }
        else {
            for (i = 0; i < array_size($2); i++) {
                ast_id_t *id = array_get($2, i, ast_id_t);

                id->mod = $1.mod;
                id->u_var.type_meta = $1.meta;
                id->u_var.init_exp = array_get($4, i, ast_exp_t);
            }
        }
        $$ = $2;
    }
;

var_type:
    var_qual
|   K_PUBLIC var_type
    {
        $$ = $2;
        flag_set($$.mod, MOD_PUBLIC);
    }
;

var_qual:
    var_spec
    {
        $$.mod = MOD_PRIVATE;
        $$.meta = $1;
    }
|   K_CONST var_spec
    {
        $$.mod = MOD_CONST;
        $$.meta = $2;
    }
;

var_spec:
    prim_type
    {
        $$ = meta_new($1, &@1);
    }
|   identifier
    {
        $$ = meta_new(TYPE_STRUCT, &@1);
        $$->name = $1;
    }
|   K_MAP '(' var_spec ',' var_spec ')'
    {
        $$ = meta_new(TYPE_MAP, &@1);
        meta_set_map($$, $3, $5);
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
        $$ = id_new_var($1, MOD_PRIVATE, &@1);
    }
|   declarator '[' size_opt ']'
    {
        $$ = $1;

        if ($$->u_var.size_exps == NULL)
            $$->u_var.size_exps = array_new();

        exp_add_last($$->u_var.size_exps, $3);
    }
;

size_opt:
    /* empty */         { $$ = exp_new_null(&@$); }
|   add_exp
;

var_init_list:
    sql_exp
    {
        $$ = array_new();
        exp_add_last($$, $1);
    }
|   var_init_list ',' sql_exp
    {
        $$ = $1;
        exp_add_last($$, $3);
    }
;

compound:
    struct
|   enumeration
|   constructor
|   function
;

struct:
    K_STRUCT identifier '{' field_list '}'
    {
        $$ = id_new_struct($2, $4, &@$);
    }
|   K_STRUCT error '}'
    {
        $$ = NULL;
    }
;

field_list:
    var_decl ';'
|   field_list var_decl ';'
    {
        $$ = $1;
        id_join_last($$, $2);
    }
;

enumeration:
    K_ENUM identifier '{' enum_list comma_opt '}'
    {
        $$ = id_new_enum($2, $4, &@$);
    }
|   K_ENUM error ')'
    {
        $$ = NULL;
    }
;

enum_list:
    enumerator
    {
        $$ = array_new();
        id_add_last($$, $1);
    }
|   enum_list ',' enumerator
    {
        $$ = $1;
        id_add_last($$, $3);
    }
;

enumerator:
    identifier
    {
        $$ = id_new_var($1, MOD_PUBLIC | MOD_CONST, &@1);
    }
|   identifier '=' or_exp
    {
        $$ = id_new_var($1, MOD_PUBLIC | MOD_CONST, &@1);
        $$->u_var.init_exp = $3;
    }
;

comma_opt:
    /* empty */
|   ','
;

constructor:
    identifier '(' param_list_opt ')' block
    {
        $$ = id_new_func($1, MOD_PUBLIC | MOD_CTOR, $3, NULL, $5, &@$);
    }
;

param_list_opt:
    /* empty */         { $$ = NULL; }
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
    var_qual declarator
    {
        $$ = $2;
        $$->mod = $1.mod;
        $$->u_var.type_meta = $1.meta;
    }
;

block:
    '{' '}'             { $$ = NULL; }
|   '{' blk_decl '}'    { $$ = $2; }
;

blk_decl:
    variable
    {
        $$ = blk_new_anon(&@$);
        id_join_last(&$$->ids, $1);
    }
|   struct
    {
        $$ = blk_new_anon(&@$);
        id_add_last(&$$->ids, $1);
    }
|   statement
    {
        $$ = blk_new_anon(&@$);
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
        $$ = id_new_func($3, $1, $5, $7, $8, &@3);
    }
;

modifier_opt:
    /* empty */             { $$ = MOD_PRIVATE; }
|   K_PUBLIC                { $$ = MOD_PUBLIC; }
|   modifier_opt K_READONLY
    {
        if (flag_on($1, MOD_READONLY))
            ERROR(ERROR_SYNTAX, &@2, "syntax error, unexpected readonly, "
                  "expecting func or payable");

        $$ = $1;
        flag_set($$, MOD_READONLY);
    }
|   modifier_opt K_PAYABLE
    {
        if (flag_on($1, MOD_PAYABLE))
            ERROR(ERROR_SYNTAX, &@2, "syntax error, unexpected payable, "
                  "expecting func or readonly");

        $$ = $1;
        flag_set($$, MOD_PAYABLE);
    }
;

return_opt:
    /* empty */             { $$ = NULL; }
|   '(' return_list ')'     { $$ = $2; }
|   return_list
;

return_list:
    return_decl
    {
        $$ = array_new();
        id_add_last($$, $1);
    }
|   return_list ',' return_decl
    {
        $$ = $1;
        id_add_last($$, $3);
    }
;

return_decl:
    var_spec
    {
        char name[256];

        snprintf(name, sizeof(name), "__return_var_%d", node_num_);

        $$ = id_new_var(xstrdup(name), MOD_PRIVATE, &@1);
        $$->u_var.type_meta = $1;
    }
|   var_spec declarator
    {
        $$ = $2;
        $$->u_var.type_meta = $1;
    }
;

statement:
    empty_stmt
|   exp_stmt
|   assign_stmt
|   label_stmt
|   if_stmt
|   loop_stmt
|   switch_stmt
|   jump_stmt
|   ddl_stmt
|   blk_stmt
;

empty_stmt:
    ';'
    {
        $$ = stmt_new_null(&@$);
    }
;

exp_stmt:
    expression ';'
    {
        $$ = stmt_new_exp($1, &@$);
    }
|   error ';'           { $$ = NULL; }
;

assign_stmt:
    expression '=' expression ';'
    {
        $$ = stmt_new_assign($1, $3, &@2);
    }
|   unary_exp assign_op expression ';'
    {
        $$ = stmt_new_assign($1, exp_new_op($2, $1, $3, &@2), &@2);
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
        $$ = stmt_new_if($3, $5, &@$);
    }
|   if_stmt K_ELSE K_IF '(' expression ')' block
    {
        $$ = $1;
        stmt_add_last(&$$->u_if.elif_stmts, stmt_new_if($5, $7, &@2));
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
        $$ = stmt_new_loop(LOOP_FOR, NULL, NULL, $2, &@$);
    }
|   K_FOR '(' or_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, $3, NULL, $5, &@$);
    }
|   K_FOR '(' init_stmt cond_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, $4, NULL, $6, &@$);
        $$->u_loop.init_stmt = $3;
    }
|   K_FOR '(' init_stmt cond_exp expression ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, $4, $5, $7, &@$);
        $$->u_loop.init_stmt = $3;
    }
|   K_FOR '(' variable cond_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, $4, NULL, $6, &@$);
        $$->u_loop.init_ids = $3;
    }
|   K_FOR '(' variable cond_exp expression ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, $4, $5, $7, &@$);
        $$->u_loop.init_ids = $3;
    }
|   K_FOR '(' expression K_IN post_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_ARRAY, NULL, $5, $7, &@$);
        $$->u_loop.init_stmt = stmt_new_exp($3, &@3);
    }
|   K_FOR '(' var_decl K_IN post_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_ARRAY, NULL, $5, $7, &@$);
        $$->u_loop.init_ids = $3;
    }
|   K_FOR error '}'
    {
        $$ = NULL;
    }
;

init_stmt:
    empty_stmt
|   assign_stmt
;

cond_exp:
    ';'                 { $$ = NULL; }
|   expression ';'
;

switch_stmt:
    K_SWITCH '{' case_list '}'
    {
        $$ = stmt_new_switch(NULL, $3, &@$);
    }
|   K_SWITCH '(' expression ')' '{' case_list '}'
    {
        $$ = stmt_new_switch($3, $6, &@$);
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
        $$ = stmt_new_case($2, $4, &@$);
    }
|   K_DEFAULT ':' stmt_list
    {
        $$ = stmt_new_case(NULL, $3, &@$);
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
        $$ = stmt_new_jump(STMT_CONTINUE, &@$);
    }
|   K_BREAK ';'
    {
        $$ = stmt_new_jump(STMT_BREAK, &@$);
    }
|   K_RETURN ';'
    {
        $$ = stmt_new_return(NULL, &@$);
    }
|   K_RETURN expression ';'
    {
        $$ = stmt_new_return($2, &@$);
    }
|   K_GOTO identifier ';'
    {
        $$ = stmt_new_goto($2, &@2);
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

        $$ = stmt_new_ddl(ddl, &@$);

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
        $$ = stmt_new_blk($1, &@$);
    }
;

expression:
    sql_exp
|   expression ',' sql_exp
    {
        if (is_tuple_exp($1)) {
            $$ = $1;
        }
        else {
            $$ = exp_new_tuple(NULL, &@1);
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

        $$ = exp_new_sql($1, sql, &@$);

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
|   or_exp '?' expression ':' ternary_exp
    {
        $$ = exp_new_ternary($1, $3, $5, &@$);
    }
;

or_exp:
    and_exp
|   or_exp CMP_OR and_exp
    {
        $$ = exp_new_op(OP_OR, $1, $3, &@2);
    }
;

and_exp:
    bit_or_exp
|   and_exp CMP_AND bit_or_exp
    {
        $$ = exp_new_op(OP_AND, $1, $3, &@2);
    }
;

bit_or_exp:
    bit_xor_exp
|   bit_or_exp '|' bit_xor_exp
    {
        $$ = exp_new_op(OP_BIT_OR, $1, $3, &@2);
    }
;

bit_xor_exp:
    bit_and_exp
|   bit_xor_exp '^' bit_and_exp
    {
        $$ = exp_new_op(OP_BIT_XOR, $1, $3, &@2);
    }
;

bit_and_exp:
    eq_exp
|   bit_and_exp '&' eq_exp
    {
        $$ = exp_new_op(OP_BIT_AND, $1, $3, &@2);
    }
;

eq_exp:
    cmp_exp
|   eq_exp eq_op cmp_exp
    {
        $$ = exp_new_op($2, $1, $3, &@2);
    }
;

eq_op:
    CMP_EQ              { $$ = OP_EQ; }
|   CMP_NE              { $$ = OP_NE; }
;

cmp_exp:
    shift_exp
|   cmp_exp cmp_op shift_exp
    {
        $$ = exp_new_op($2, $1, $3, &@2);
    }
;

cmp_op:
    '<'                 { $$ = OP_LT; }
|   '>'                 { $$ = OP_GT; }
|   CMP_LE              { $$ = OP_LE; }
|   CMP_GE              { $$ = OP_GE; }
;

shift_exp:
    add_exp
|   shift_exp shift_op add_exp
    {
        $$ = exp_new_op($2, $1, $3, &@2);
    }
;

shift_op:
    SHIFT_R             { $$ = OP_RSHIFT; }
|   SHIFT_L             { $$ = OP_LSHIFT; }
;

add_exp:
    mul_exp
|   add_exp add_op mul_exp
    {
        $$ = exp_new_op($2, $1, $3, &@2);
    }
;

add_op:
    '+'                 { $$ = OP_ADD; }
|   '-'                 { $$ = OP_SUB; }
;

mul_exp:
    cast_exp
|   mul_exp mul_op cast_exp
    {
        $$ = exp_new_op($2, $1, $3, &@2);
    }
;

mul_op:
    '*'                 { $$ = OP_MUL; }
|   '/'                 { $$ = OP_DIV; }
|   '%'                 { $$ = OP_MOD; }
;

cast_exp:
    unary_exp
|   '(' prim_type ')' unary_exp
    {
        $$ = exp_new_cast($2, $4, &@2);
    }
;

unary_exp:
    post_exp
|   unary_op unary_exp
    {
        $$ = exp_new_op($1, $2, NULL, &@$);
    }
|   '+' unary_exp
    {
        $$ = $2;
    }
;

unary_op:
    UNARY_INC           { $$ = OP_INC; }
|   UNARY_DEC           { $$ = OP_DEC; }
|   '-'                 { $$ = OP_NEG; }
|   '!'                 { $$ = OP_NOT; }
;

post_exp:
    new_exp
|   post_exp '[' ternary_exp ']'
    {
        $$ = exp_new_array($1, $3, &@$);
    }
|   post_exp '(' arg_list_opt ')'
    {
        $$ = exp_new_call($1, $3, &@$);
    }
|   post_exp '.' identifier
    {
        $$ = exp_new_access($1, exp_new_ref($3, &@3), &@$);
    }
|   post_exp UNARY_INC
    {
        $$ = exp_new_op(OP_INC, $1, NULL, &@$);
    }
|   post_exp UNARY_DEC
    {
        $$ = exp_new_op(OP_DEC, $1, NULL, &@$);
    }
;

arg_list_opt:
    /* empty */         { $$ = NULL; }
|   arg_list
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
    prim_exp
|   K_NEW identifier '(' arg_list_opt ')'
    {
        $$ = exp_new_call(exp_new_ref($2, &@2), $4, &@$);
    }
|   K_NEW K_MAP '(' arg_list_opt ')'
    {
        $$ = exp_new_call(exp_new_ref(xstrdup("map"), &@2), $4, &@$);
    }
|   K_NEW initializer
    {
        $$ = $2;
    }
;

initializer:
    '{' elem_list comma_opt '}'
    {
        $$ = exp_new_tuple($2, &@$);
    }
;

elem_list:
    init_elem
    {
        $$ = array_new();
        exp_add_last($$, $1);
    }
|   elem_list ',' init_elem
    {
        $$ = $1;
        exp_add_last($$, $3);
    }
;

init_elem:
    ternary_exp
|   initializer
|   identifier ':' init_elem
    {
        $$ = $3;
    }
;

prim_exp:
    literal
|   identifier
    {
        $$ = exp_new_ref($1, &@$);
    }
|   '(' expression ')'
    {
        $$ = $2;
    }
;

literal:
    K_NULL
    {
        $$ = exp_new_lit(&@$);
        value_set_obj(&$$->u_lit.val, NULL);
    }
|   K_TRUE
    {
        $$ = exp_new_lit(&@$);
        value_set_bool(&$$->u_lit.val, true);
    }
|   K_FALSE
    {
        $$ = exp_new_lit(&@$);
        value_set_bool(&$$->u_lit.val, false);
    }
|   L_INT
    {
        uint64_t v;

        $$ = exp_new_lit(&@$);
        sscanf($1, "%"SCNu64, &v);
        value_set_int(&$$->u_lit.val, v);
    }
|   L_OCTAL
    {
        uint64_t v;

        $$ = exp_new_lit(&@$);
        sscanf($1, "%"SCNo64, &v);
        value_set_int(&$$->u_lit.val, v);
    }
|   L_HEX
    {
        uint64_t v;

        $$ = exp_new_lit(&@$);
        sscanf($1, "%"SCNx64, &v);
        value_set_int(&$$->u_lit.val, v);
    }
|   L_FLOAT
    {
        double v;

        $$ = exp_new_lit(&@$);
        sscanf($1, "%lf", &v);
        value_set_fp(&$$->u_lit.val, v);
    }
|   L_STR
    {
        $$ = exp_new_lit(&@$);
        value_set_str(&$$->u_lit.val, $1);
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
