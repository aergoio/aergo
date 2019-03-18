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

#define AST             (parse->ast)
#define ROOT            AST->root
#define LABELS          (&parse->labels)

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
        K_ALTER         "alter"
        K_BOOL          "bool"
        K_BREAK         "break"
        K_BYTE          "byte"
        K_CASE          "case"
        K_CONST         "const"
        K_CONTINUE      "continue"
        K_CONTRACT      "contract"
        K_CREATE        "create"
        K_CURSOR        "cursor"
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
        K_IMPLEMENTS    "implements"
        K_IMPORT        "import"
        K_IN            "in"
        K_INDEX         "index"
        K_INSERT        "insert"
        K_INT           "int"
        K_INT8          "int8"
        K_INT16         "int16"
        K_INT32         "int32"
        K_INT64         "int64"
        K_INT128        "int128"
        K_INT256        "int256"
        K_INTERFACE     "interface"
        K_LIBRARY       "library"
        K_MAP           "map"
        K_NEW           "new"
        K_NULL          "null"
        K_PAYABLE       "payable"
        K_PUBLIC        "public"
        K_REPLACE       "replace"
        K_RETURN        "return"
        K_SELECT        "select"
        K_STRING        "string"
        K_STRUCT        "struct"
        K_SWITCH        "switch"
        K_TABLE         "table"
        K_TRUE          "true"
        K_TYPE          "type"
        K_UINT          "uint"
        K_UINT8         "uint8"
        K_UINT16        "uint16"
        K_UINT32        "uint32"
        K_UINT64        "uint64"
        K_UINT128       "uint128"
        K_UINT256       "uint256"
        K_UPDATE        "update"
        K_VIEW          "view"

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
    vector_t *vect;

    type_t type;
    op_kind_t op;
    sql_kind_t sql;
    modifier_t mod;

    ast_id_t *id;
    ast_blk_t *blk;
    ast_exp_t *exp;
    ast_stmt_t *stmt;
    meta_t *meta;
}

%type <exp>     impl_opt
%type <blk>     contract_body
%type <id>      variable
%type <id>      var_qual
%type <id>      var_decl
%type <id>      var_spec
%type <exp>     var_type
%type <type>    prim_type
%type <id>      declarator_list
%type <id>      declarator
%type <exp>     size_opt
%type <exp>     var_init
%type <id>      compound
%type <id>      struct
%type <vect>    field_list
%type <id>      enumeration
%type <vect>    enum_list
%type <id>      enumerator
%type <id>      function
%type <id>      func_spec
%type <id>      ctor_spec
%type <vect>    param_list_opt
%type <vect>    param_list
%type <id>      param_decl
%type <id>      udf_spec
%type <mod>     modifier_opt
%type <id>      return_opt
%type <id>      return_list
%type <id>      return_decl
%type <blk>     block
%type <blk>     blk_decl
%type <blk>     interface_body
%type <blk>     library_body
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
%type <blk>     switch_blk
%type <blk>     case_blk
%type <stmt>    jump_stmt
%type <stmt>    ddl_stmt
%type <stmt>    blk_stmt
%type <exp>     expression
%type <exp>     sql_exp
%type <sql>     sql_prefix
%type <exp>     new_exp
%type <exp>     alloc_exp
%type <exp>     initializer
%type <vect>    elem_list
%type <exp>     init_elem
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
%type <exp>     add_exp
%type <op>      add_op
%type <exp>     shift_exp
%type <op>      shift_op
%type <exp>     mul_exp
%type <op>      mul_op
%type <exp>     cast_exp
%type <exp>     unary_exp
%type <op>      unary_op
%type <exp>     post_exp
%type <vect>    arg_list_opt
%type <vect>    arg_list
%type <exp>     prim_exp
%type <exp>     literal
%type <str>     non_reserved_token
%type <str>     identifier

%start  root

%%

root:
    component
|   library
;

component:
    import
|   contract
|   interface
|   root import
|   root contract
|   root interface
;

import:
    K_IMPORT L_STR
;

contract:
    K_CONTRACT identifier impl_opt '{' '}'
    {
        ast_blk_t *blk = blk_new_contract(&@4);

        /* add default constructor */
        id_add(&blk->ids, id_new_ctor($2, NULL, NULL, &@2));

        id_add(&ROOT->ids, id_new_contract($2, $3, blk, &@$));
    }
|   K_CONTRACT identifier impl_opt '{' contract_body '}'
    {
        int i;
        bool has_ctor = false;

        vector_foreach(&$5->ids, i) {
            if (is_ctor_id(vector_get_id(&$5->ids, i)))
                has_ctor = true;
        }

        if (!has_ctor)
            /* add default constructor */
            id_add(&$5->ids, id_new_ctor($2, NULL, NULL, &@2));

        id_add(&ROOT->ids, id_new_contract($2, $3, $5, &@$));
    }
;

impl_opt:
    /* empty */                 { $$ = NULL; }
|   K_IMPLEMENTS identifier
    {
        $$ = exp_new_id($2, &@2);
    }
;

contract_body:
    variable
    {
        $$ = blk_new_contract(&@$);
        stmt_add(&$$->stmts, stmt_new_id($1, &@1));
    }
|   compound
    {
        $$ = blk_new_contract(&@$);
        id_add(&$$->ids, $1);
    }
|   function
    {
        $$ = blk_new_contract(&@$);
        id_add(&$$->ids, $1);
    }
|   contract_body variable
    {
        $$ = $1;
        stmt_add(&$$->stmts, stmt_new_id($2, &@2));
    }
|   contract_body compound
    {
        $$ = $1;
        id_add(&$$->ids, $2);
    }
|   contract_body function
    {
        $$ = $1;
        id_add(&$$->ids, $2);
    }
;

variable:
    var_qual
|   K_PUBLIC var_qual
    {
        $$ = $2;
        $$->mod |= MOD_PUBLIC;
    }
|   var_type error ';'
    {
        yyerrok;
        $$ = NULL;
    }
;

var_qual:
    var_decl
|   K_CONST var_decl
    {
        $$ = $2;
        $$->mod |= MOD_CONST;
    }
;

var_decl:
    var_spec ';'
|   var_spec '=' var_init ';'
    {
        $$ = $1;

        if (is_var_id($$))
            $$->u_var.dflt_exp = $3;
        else
            $$->u_tup.dflt_exp = $3;
    }
;

var_spec:
    var_type declarator_list
    {
        $$ = $2;

        if (is_var_id($$))
            $$->u_var.type_exp = $1;
        else
            $$->u_tup.type_exp = $1;
    }
;

var_type:
    prim_type
    {
        $$ = exp_new_type($1, &@1);
    }
|   identifier
    {
        $$ = exp_new_type(TYPE_NONE, &@1);

        $$->u_type.name = $1;
    }
|   K_MAP '(' var_type ',' var_type ')'
    {
        $$ = exp_new_type(TYPE_MAP, &@1);

        $$->u_type.k_exp = $3;
        $$->u_type.v_exp = $5;
    }
;

prim_type:
    K_ACCOUNT           { $$ = TYPE_ACCOUNT; }
|   K_BOOL              { $$ = TYPE_BOOL; }
|   K_BYTE              { $$ = TYPE_UINT8; }
|   K_INT8              { $$ = TYPE_INT8; }
|   K_INT16             { $$ = TYPE_INT16; }
|   K_INT32             { $$ = TYPE_INT32; }
|   K_INT64             { $$ = TYPE_INT64; }
|   K_INT128            { $$ = TYPE_INT128; }
|   K_INT256            { $$ = TYPE_INT256; }
|   K_INT               { $$ = TYPE_INT32; }
|   K_UINT8             { $$ = TYPE_UINT8; }
|   K_UINT16            { $$ = TYPE_UINT16; }
|   K_UINT32            { $$ = TYPE_UINT32; }
|   K_UINT64            { $$ = TYPE_UINT64; }
|   K_UINT128           { $$ = TYPE_UINT128; }
|   K_UINT256           { $$ = TYPE_UINT256; }
|   K_UINT              { $$ = TYPE_UINT32; }
//|   K_FLOAT             { $$ = TYPE_FLOAT; }
//|   K_DOUBLE            { $$ = TYPE_DOUBLE; }
|   K_STRING            { $$ = TYPE_STRING; }
|   K_CURSOR            { $$ = TYPE_CURSOR; }
;

declarator_list:
    declarator
|   declarator_list ',' declarator
    {
        if (is_tuple_id($1)) {
            $$ = $1;
        }
        else {
            $$ = id_new_tuple(&@1);
            id_add($$->u_tup.elem_ids, $1);
        }

        id_add($$->u_tup.elem_ids, $3);
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
            $$->u_var.size_exps = vector_new();

        exp_add($$->u_var.size_exps, $3);
    }
;

size_opt:
    /* empty */         { $$ = exp_new_null(&@$); }
|   add_exp
;

var_init:
    sql_exp
    {
        $$ = $1;
    }
|   var_init ',' sql_exp
    {
        if (is_tuple_exp($1)) {
            $$ = $1;
        }
        else {
            $$ = exp_new_tuple(vector_new(), &@1);
            exp_add($$->u_tup.elem_exps, $1);
        }
        exp_add($$->u_tup.elem_exps, $3);
    }
;

compound:
    struct
|   enumeration
;

struct:
    K_TYPE identifier K_STRUCT '{' field_list '}'
    {
        $$ = id_new_struct($2, $5, &@$);
    }
|   K_TYPE error '}'
    {
        yyerrok;
        $$ = NULL;
    }
;

field_list:
    var_spec ';'
    {
        $$ = vector_new();

        if (is_var_id($1))
            id_add($$, $1);
        else
            id_join($$, id_strip($1));
    }
|   field_list var_spec ';'
    {
        $$ = $1;

        if (is_var_id($2))
            id_add($$, $2);
        else
            id_join($$, id_strip($2));
    }
;

enumeration:
    K_ENUM identifier '{' enum_list comma_opt '}'
    {
        $$ = id_new_enum($2, $4, &@$);
    }
|   K_ENUM error '}'
    {
        yyerrok;
        $$ = NULL;
    }
;

enum_list:
    enumerator
    {
        $$ = vector_new();
        id_add($$, $1);
    }
|   enum_list ',' enumerator
    {
        $$ = $1;
        id_add($$, $3);
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
        $$->u_var.dflt_exp = $3;
    }
;

comma_opt:
    /* empty */
|   ','
;

function:
    func_spec block
    {
        $$ = $1;
        $$->u_fn.blk = $2;

        /* The label is added to the topmost block because it can be referenced
         * regardless of the order of declaration. */
        if (!is_empty_vector(LABELS)) {
            ASSERT($2 != NULL);
            id_join(&$2->ids, LABELS);
            vector_reset(LABELS);
        }
    }
;

func_spec:
    ctor_spec
|   udf_spec
;

ctor_spec:
    identifier '(' param_list_opt ')'
    {
        $$ = id_new_ctor($1, $3, NULL, &@1);
    }
;

param_list_opt:
    /* empty */         { $$ = NULL; }
|   param_list
;

param_list:
    param_decl
    {
        $$ = vector_new();
        exp_add($$, $1);
    }
|   param_list ',' param_decl
    {
        $$ = $1;
        exp_add($$, $3);
    }
;

param_decl:
    var_type declarator
    {
        $$ = $2;
        $$->u_var.is_param = true;
        $$->u_var.type_exp = $1;
    }
;

udf_spec:
    modifier_opt K_FUNC identifier '(' param_list_opt ')' return_opt
    {
        $$ = id_new_func($3, $1, $5, $7, NULL, &@3);
    }
;

modifier_opt:
    /* empty */             { $$ = MOD_PRIVATE; }
|   K_PUBLIC                { $$ = MOD_PUBLIC; }
|   K_PAYABLE               { $$ = MOD_PAYABLE; }
|   K_PUBLIC K_PAYABLE      { $$ = MOD_PUBLIC | MOD_PAYABLE; }
;

return_opt:
    /* empty */             { $$ = NULL; }
|   '(' return_list ')'     { $$ = $2; }
|   return_list
;

return_list:
    return_decl
|   return_list ',' return_decl
    {
        yyerror(&@3, parse, yyscanner, "not supported yet");

        /* TODO multiple return values
         *
         * Since WebAssembly does not support this syntax at present, it is better to
         * implement it when it is formally supported than implementing arbitarily
         */
#if 0
        /* The reason for making the return list as tuple is because of
         * the convenience of meta comparison. If vector_t is used,
         * it must be looped for each id and compared directly,
         * but for tuples, meta_cmp() is sufficient */

        if (is_tuple_id($1)) {
            $$ = $1;
        }
        else {
            $$ = id_new_tuple(&@1);
            id_add($$->u_tup.elem_ids, $1);
        }

        id_add($$->u_tup.elem_ids, $3);
#endif
    }
;

return_decl:
    var_type
    {
        /* We wanted to use a type expression, but we can not store size expressions
         * and declare it as an identifier. */
        $$ = id_new_param(NULL, $1, &@1);
    }
|   return_decl '[' size_opt ']'
    {
        $$ = $1;

        if ($$->u_var.size_exps == NULL)
            $$->u_var.size_exps = vector_new();

        exp_add($$->u_var.size_exps, $3);
    }
;

block:
    '{' '}'             { $$ = NULL; }
|   '{' blk_decl '}'    { $$ = $2; }
;

blk_decl:
    var_qual
    {
        $$ = blk_new_normal(&@$);
        stmt_add(&$$->stmts, stmt_new_id($1, &@1));
    }
|   compound
    {
        $$ = blk_new_normal(&@$);
        /* Unlike state variables, local variables are referenced according to their
         * order of declaration. */
        stmt_add(&$$->stmts, stmt_new_id($1, &@1));
    }
|   statement
    {
        $$ = blk_new_normal(&@$);
        stmt_add(&$$->stmts, $1);
    }
|   blk_decl var_qual
    {
        $$ = $1;
        stmt_add(&$$->stmts, stmt_new_id($2, &@2));
    }
|   blk_decl compound
    {
        $$ = $1;
        stmt_add(&$$->stmts, stmt_new_id($2, &@2));
    }
|   blk_decl statement
    {
        $$ = $1;
        stmt_add(&$$->stmts, $2);
    }
;

interface:
    K_INTERFACE identifier '{' interface_body '}'
    {
        id_add(&ROOT->ids, id_new_interface($2, $4, &@2));
    }
;

interface_body:
    udf_spec ';'
    {
        $$ = blk_new_interface(&@$);
        id_add(&$$->ids, $1);
    }
|   interface_body udf_spec ';'
    {
        $$ = $1;
        id_add(&$$->ids, $2);
    }
;

library:
    K_LIBRARY identifier '{' library_body '}'
    {
        id_add(&ROOT->ids, id_new_library($2, $4, &@2));
    }
;

library_body:
    udf_spec ':' L_STR ';'
    {
        $$ = blk_new_library(&@$);

        $1->mod = MOD_SYSTEM;
        $1->u_fn.alias = $3;

        vector_add_last(&$$->ids, $1);
    }
|   udf_spec block
    {
        $$ = blk_new_library(&@$);

        $1->mod = MOD_SYSTEM;
        $1->u_fn.blk = $2;

        vector_add_last(&$$->ids, $1);
    }
|   library_body udf_spec ':' L_STR ';'
    {
        $$ = $1;

        $2->mod = MOD_SYSTEM;
        $2->u_fn.alias = $4;

        vector_add_last(&$$->ids, $2);
    }
|   library_body udf_spec block
    {
        $$ = $1;

        $2->mod = MOD_SYSTEM;
        $2->u_fn.blk = $3;

        vector_add_last(&$$->ids, $2);
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
|   error ';'
    {
        yyerrok;
        $$ = NULL;
    }
;

assign_stmt:
    expression '=' expression ';'
    {
        $$ = stmt_new_assign($1, $3, &@2);
    }
|   unary_exp assign_op expression ';'
    {
        $$ = stmt_new_assign($1, exp_new_binary($2, $1, $3, &@2), &@2);
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
        id_add(LABELS, id_new_label($1, $3, &@1));
    }
|   K_CASE eq_exp ':'
    {
        $$ = stmt_new_case($2, &@$);
    }
|   K_DEFAULT ':'
    {
        $$ = stmt_new_case(NULL, &@$);
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
        stmt_add(&$$->u_if.elif_stmts, stmt_new_if($5, $7, &@2));
    }
|   if_stmt K_ELSE block
    {
        $$ = $1;
        $$->u_if.else_blk = $3;
    }
|   K_IF error '}'
    {
        yyerrok;
        $$ = NULL;
    }
;

loop_stmt:
    K_FOR block
    {
        $$ = stmt_new_loop(LOOP_FOR, NULL, NULL, NULL, $2, &@$);
    }
|   K_FOR '(' ternary_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, NULL, $3, NULL, $5, &@$);
    }
|   K_FOR '(' init_stmt cond_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, $3, $4, NULL, $6, &@$);
    }
|   K_FOR '(' init_stmt cond_exp expression ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, $3, $4, $5, $7, &@$);
    }
|   K_FOR '(' var_decl cond_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, stmt_new_id($3, &@3), $4, NULL, $6, &@$);
    }
|   K_FOR '(' var_decl cond_exp expression ')' block
    {
        $$ = stmt_new_loop(LOOP_FOR, stmt_new_id($3, &@3), $4, $5, $7, &@$);
    }
|   K_FOR '(' expression K_IN post_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_ARRAY, stmt_new_exp($3, &@3), $5, NULL, $7, &@$);
    }
|   K_FOR '(' var_spec K_IN post_exp ')' block
    {
        $$ = stmt_new_loop(LOOP_ARRAY, stmt_new_id($3, &@3), $5, NULL, $7, &@$);
    }
|   K_FOR error '}'
    {
        yyerrok;
        $$ = NULL;
    }
;

init_stmt:
    empty_stmt
|   assign_stmt
;

cond_exp:
    ';'                 { $$ = NULL; }
|   ternary_exp ';'
;

switch_stmt:
    K_SWITCH switch_blk
    {
        $$ = stmt_new_switch(NULL, $2, &@$);
    }
|   K_SWITCH '(' expression ')' switch_blk
    {
        $$ = stmt_new_switch($3, $5, &@$);
    }
|   K_SWITCH error '}'
    {
        yyerrok;
        $$ = NULL;
    }
;

switch_blk:
    '{' '}'             { $$ = NULL; }
|   '{' case_blk '}'    { $$ = $2; }
;

case_blk:
    label_stmt
    {
        $$ = blk_new_switch(&@$);
        stmt_add(&$$->stmts, $1);
    }
|   case_blk statement
    {
        $$ = $1;
        stmt_add(&$$->stmts, $2);
    }
;

jump_stmt:
    K_CONTINUE ';'
    {
        $$ = stmt_new_jump(STMT_CONTINUE, NULL, &@$);
    }
|   K_BREAK ';'
    {
        $$ = stmt_new_jump(STMT_BREAK, NULL, &@$);
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

        len = @$.last_offset - @$.first_offset;
        ddl = xstrndup(parse->src + @$.first_offset, len);

        $$ = stmt_new_ddl(ddl, &@$);

        yylex_set_token(yyscanner, ';', &@3);
        yyclearin;
    }
;

ddl_prefix:
    K_CREATE K_TABLE
|   K_ALTER K_TABLE
|   K_DROP K_TABLE
|   K_CREATE K_INDEX
|   K_DROP K_INDEX
|   K_CREATE K_VIEW
|   K_DROP K_VIEW
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
            $$ = exp_new_tuple(vector_new(), &@1);
            exp_add($$->u_tup.elem_exps, $1);
        }
        exp_add($$->u_tup.elem_exps, $3);
    }
;

sql_exp:
    new_exp
|   sql_prefix error ';'
    {
        int len;
        char *sql;

        yyerrok;
        error_pop();

        len = @$.last_offset - @$.first_offset;
        sql = xstrndup(parse->src + @$.first_offset, len);

        $$ = exp_new_sql($1, sql, &@$);

        yylex_set_token(yyscanner, ';', &@3);
        yyclearin;
    }
;

sql_prefix:
    K_DELETE            { $$ = SQL_DELETE; }
|   K_INSERT            { $$ = SQL_INSERT; }
|   K_REPLACE           { $$ = SQL_REPLACE; }
|   K_SELECT            { $$ = SQL_QUERY; }
|   K_UPDATE            { $$ = SQL_UPDATE; }
;

new_exp:
    ternary_exp
|   K_NEW alloc_exp
    {
        $$ = $2;
    }
|   K_NEW initializer
    {
        $$ = $2;
        $$->u_init.is_outmost = true;
    }
;

alloc_exp:
    var_type
    {
        $$ = exp_new_alloc($1, &@1);
    }
|   alloc_exp '[' add_exp ']'
    {
        $$ = $1;

        if ($$->u_alloc.size_exps == NULL)
            $$->u_alloc.size_exps = vector_new();

        exp_add($$->u_alloc.size_exps, $3);
    }
;

initializer:
    '{' elem_list comma_opt '}'
    {
        $$ = exp_new_init($2, &@$);
    }
;

elem_list:
    init_elem
    {
        $$ = vector_new();
        exp_add($$, $1);
    }
|   elem_list ',' init_elem
    {
        $$ = $1;
        exp_add($$, $3);
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
        $$ = exp_new_binary(OP_OR, $1, $3, &@2);
    }
;

and_exp:
    bit_or_exp
|   and_exp CMP_AND bit_or_exp
    {
        $$ = exp_new_binary(OP_AND, $1, $3, &@2);
    }
;

bit_or_exp:
    bit_xor_exp
|   bit_or_exp '|' bit_xor_exp
    {
        $$ = exp_new_binary(OP_BIT_OR, $1, $3, &@2);
    }
;

bit_xor_exp:
    bit_and_exp
|   bit_xor_exp '^' bit_and_exp
    {
        $$ = exp_new_binary(OP_BIT_XOR, $1, $3, &@2);
    }
;

bit_and_exp:
    eq_exp
|   bit_and_exp '&' eq_exp
    {
        $$ = exp_new_binary(OP_BIT_AND, $1, $3, &@2);
    }
;

eq_exp:
    cmp_exp
|   eq_exp eq_op cmp_exp
    {
        $$ = exp_new_binary($2, $1, $3, &@2);
    }
;

eq_op:
    CMP_EQ              { $$ = OP_EQ; }
|   CMP_NE              { $$ = OP_NE; }
;

cmp_exp:
    add_exp
|   cmp_exp cmp_op add_exp
    {
        $$ = exp_new_binary($2, $1, $3, &@2);
    }
;

cmp_op:
    '<'                 { $$ = OP_LT; }
|   '>'                 { $$ = OP_GT; }
|   CMP_LE              { $$ = OP_LE; }
|   CMP_GE              { $$ = OP_GE; }
;

add_exp:
    shift_exp
|   add_exp add_op shift_exp
    {
        $$ = exp_new_binary($2, $1, $3, &@2);
    }
;

add_op:
    '+'                 { $$ = OP_ADD; }
|   '-'                 { $$ = OP_SUB; }
;

shift_exp:
    mul_exp
|   shift_exp shift_op mul_exp
    {
        $$ = exp_new_binary($2, $1, $3, &@2);
    }
;

shift_op:
    SHIFT_R             { $$ = OP_RSHIFT; }
|   SHIFT_L             { $$ = OP_LSHIFT; }
;

mul_exp:
    cast_exp
|   mul_exp mul_op cast_exp
    {
        $$ = exp_new_binary($2, $1, $3, &@2);
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
        $$ = exp_new_unary($1, true, $2, &@$);
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
    prim_exp
|   post_exp '[' ternary_exp ']'
    {
        $$ = exp_new_array($1, $3, &@$);
    }
|   post_exp '(' arg_list_opt ')'
    {
        $$ = exp_new_call(false, $1, $3, &@$);
    }
|   post_exp '.' identifier
    {
        $$ = exp_new_access($1, exp_new_id($3, &@3), &@$);
    }
|   post_exp UNARY_INC
    {
        $$ = exp_new_unary(OP_INC, false, $1, &@$);
    }
|   post_exp UNARY_DEC
    {
        $$ = exp_new_unary(OP_DEC, false, $1, &@$);
    }
;

arg_list_opt:
    /* empty */         { $$ = NULL; }
|   arg_list
;

arg_list:
    new_exp
    {
        $$ = vector_new();
        exp_add($$, $1);
    }
|   arg_list ',' new_exp
    {
        $$ = $1;
        exp_add($$, $3);
    }
;

prim_exp:
    literal
|   identifier
    {
        $$ = exp_new_id($1, &@$);
    }
|   '(' expression ')'
    {
        $$ = $2;
    }
|   K_NEW identifier '(' arg_list_opt ')'
    {
        $$ = exp_new_call(true, exp_new_id($2, &@2), $4, &@2);
    }
;

literal:
    K_NULL
    {
        $$ = exp_new_lit_null(&@$);
    }
|   K_TRUE
    {
        $$ = exp_new_lit_bool(true, &@$);
    }
|   K_FALSE
    {
        $$ = exp_new_lit_bool(false, &@$);
    }
|   L_INT
    {
        uint64_t v;

        sscanf($1, "%"SCNu64, &v);
        $$ = exp_new_lit_i64(v, &@$);
    }
|   L_OCTAL
    {
        uint64_t v;

        sscanf($1, "%"SCNo64, &v);
        $$ = exp_new_lit_i64(v, &@$);
    }
|   L_HEX
    {
        uint64_t v;

        sscanf($1, "%"SCNx64, &v);
        $$ = exp_new_lit_i64(v, &@$);
    }
|   L_FLOAT
    {
        double v;

        sscanf($1, "%lf", &v);
        $$ = exp_new_lit_f64(v, &@$);
    }
|   L_STR
    {
        $$ = exp_new_lit_str($1, &@$);
    }
;

non_reserved_token:
    K_CONTRACT          { $$ = xstrdup("contract"); }
|   K_IMPORT            { $$ = xstrdup("import"); }
|   K_INDEX             { $$ = xstrdup("index"); }
|   K_INTERFACE         { $$ = xstrdup("interface"); }
|   K_LIBRARY           { $$ = xstrdup("library"); }
|   K_TABLE             { $$ = xstrdup("table"); }
|   K_VIEW              { $$ = xstrdup("view"); }
;

identifier:
    ID
    {
        if (strlen($1) > NAME_MAX_LEN)
            ERROR(ERROR_TOO_LONG_ID, &@1, NAME_MAX_LEN, strlen($1));

        $$ = $1;
    }
|   non_reserved_token
;

%%

static void
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
