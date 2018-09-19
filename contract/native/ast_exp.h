/**
 * @file    ast_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_EXP_H
#define _AST_EXP_H

#include "common.h"

#include "ast.h"
#include "ast_meta.h"

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_VAR_T
#define _AST_VAR_T
typedef struct ast_var_s ast_var_t;
#endif /* ! _AST_VAR_T */

typedef enum expn_e {
    EXP_ID          = 0,
    EXP_LIT,
    EXP_TYPE,
    EXP_ARRAY,
    EXP_OP,
    EXP_ACCESS,
    EXP_CALL,
    EXP_SQL,
    EXP_COND,
    EXP_TUPLE,
    EXP_MAX
} expn_t;

typedef enum op_e {
    OP_ASSIGN       = 0,
    OP_ADD,
    OP_SUB,
    OP_MUL,
    OP_DIV,
    OP_MOD,
    OP_AND,
    OP_OR,
    OP_BIT_AND,
    OP_BIT_OR,
    OP_BIT_XOR,
    OP_EQ,
    OP_NE,
    OP_LT,
    OP_GT,
    OP_LE,
    OP_GE,
    OP_RSHIFT,
    OP_LSHIFT,
    OP_INC,
    OP_DEC,
    OP_NOT,
    OP_MAX
} op_t;

typedef enum sql_kind_e {
    SQL_QUERY       = 0,
    SQL_INSERT,
    SQL_UPDATE,
    SQL_DELETE,
    SQL_MAX
} sql_kind_t;

typedef enum lit_kind_e {
    LIT_NULL        = 0,
    LIT_BOOL,
    LIT_INT,
    LIT_FLOAT,
    LIT_HEXA,
    LIT_STR,
    LIT_MAX
} lit_kind_t;

// null, true, false, 1, 1.0, 0x1, "..."
typedef struct exp_lit_s {
    lit_kind_t kind;
    char *val;
} exp_lit_t;

// primitive, struct, map
typedef struct exp_type_s {
    char *name;
    ast_exp_t *k_exp;
    ast_exp_t *v_exp;
} exp_type_t;

// name
typedef struct exp_id_ref_s {
    char *name;
} exp_id_ref_t;

// id[param]
typedef struct exp_array_s {
    ast_exp_t *id_exp;
    ast_exp_t *param_exp;
} exp_array_t;

// id(param...)
typedef struct exp_call_s {
    ast_exp_t *id_exp;
    list_t *param_l;
} exp_call_t;

// id.memb
typedef struct exp_access_s {
    ast_exp_t *id_exp;
    ast_exp_t *memb_exp;
} exp_access_t;

// l op r
typedef struct exp_op_s {
    op_t op;
    ast_exp_t *l_exp;
    ast_exp_t *r_exp;
} exp_op_t;

// cond ? true : false
typedef struct exp_cond_s {
    ast_exp_t *cond_exp;
    ast_exp_t *t_exp;
    ast_exp_t *f_exp;
} exp_cond_t;

// dml, query
typedef struct exp_sql_s {
    sql_kind_t kind;
    char *sql;
} exp_sql_t;

typedef struct exp_tuple_s {
    list_t *exp_l;
} exp_tuple_t;

struct ast_exp_s {
    AST_NODE_DECL;

    expn_t type;
    ast_meta_t meta;

    union {
        exp_lit_t u_lit;
        exp_type_t u_type;
        exp_id_ref_t u_id;
        exp_array_t u_arr;
        exp_call_t u_call;
        exp_access_t u_acc;
        exp_op_t u_op;
        exp_cond_t u_cond;
        exp_sql_t u_sql;
        exp_tuple_t u_tuple;
    };
};

ast_exp_t *exp_lit_new(lit_kind_t kind, char *val, errpos_t *pos);
ast_exp_t *exp_type_new(type_t type, char *name, ast_exp_t *k_exp,
                        ast_exp_t *v_exp, errpos_t *pos);
ast_exp_t *exp_id_ref_new(char *name, errpos_t *pos);
ast_exp_t *exp_array_new(ast_exp_t *id_exp, ast_exp_t *param_exp, 
                         errpos_t *pos);
ast_exp_t *exp_call_new(ast_exp_t *id_exp, list_t *param_l, errpos_t *pos);
ast_exp_t *exp_access_new(ast_exp_t *id_exp, ast_exp_t *memb_exp, 
                          errpos_t *pos);
ast_exp_t *exp_op_new(op_t op, ast_exp_t *l_exp, ast_exp_t *r_exp, 
                      errpos_t *pos);
ast_exp_t *exp_cond_new(ast_exp_t *cond_exp, ast_exp_t *t_exp, 
                        ast_exp_t *f_exp, errpos_t *pos);
ast_exp_t *exp_sql_new(sql_kind_t kind, char *sql, errpos_t *pos);
ast_exp_t *exp_tuple_new(ast_exp_t *exp, errpos_t *pos);

#endif /* ! _AST_EXP_H */
