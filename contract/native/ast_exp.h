/**
 * @file    ast_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_EXP_H
#define _AST_EXP_H

#include "common.h"

#include "ast.h"
#include "value.h"

#define is_null_exp(exp)            ((exp)->kind == EXP_NULL)
#define is_val_exp(exp)             ((exp)->kind == EXP_VAL)
#define is_type_exp(exp)            ((exp)->kind == EXP_TYPE)
#define is_id_exp(exp)              ((exp)->kind == EXP_ID)
#define is_array_exp(exp)           ((exp)->kind == EXP_ARRAY)
#define is_op_exp(exp)              ((exp)->kind == EXP_OP)
#define is_access_exp(exp)          ((exp)->kind == EXP_ACCESS)
#define is_call_exp(exp)            ((exp)->kind == EXP_CALL)
#define is_sql_exp(exp)             ((exp)->kind == EXP_SQL)
#define is_ternary_exp(exp)         ((exp)->kind == EXP_TERNARY)
#define is_tuple_exp(exp)           ((exp)->kind == EXP_TUPLE)

#define is_usable_lval(exp)                                                              \
    ((exp)->id != NULL && !is_const_id((exp)->id) && !is_untyped_meta(&(exp)->id->meta))

#define exp_add_first               array_add_first
#define exp_add_last                array_add_last

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

// null, true, false, 1, 1.0, 0x1, "..."
typedef struct exp_val_s {
    value_t val;
} exp_val_t;

// primitive, struct, map
typedef struct exp_type_s {
    type_t type;
    char *name;
    modifier_t mod;
    ast_exp_t *k_exp;
    ast_exp_t *v_exp;
} exp_type_t;

// name
typedef struct exp_id_s {
    char *name;
} exp_id_t;

// id[idx]
typedef struct exp_array_s {
    ast_exp_t *id_exp;
    ast_exp_t *idx_exp;
} exp_array_t;

// id(param, ...)
typedef struct exp_call_s {
    ast_exp_t *id_exp;
    array_t *param_exps;
} exp_call_t;

// id.fld
typedef struct exp_access_s {
    ast_exp_t *id_exp;
    ast_exp_t *fld_exp;
} exp_access_t;

// l kind r
typedef struct exp_op_s {
    op_kind_t kind;
    ast_exp_t *l_exp;
    ast_exp_t *r_exp;
} exp_op_t;

// prefix ? infix : postfix
typedef struct exp_ternary_s {
    ast_exp_t *pre_exp;
    ast_exp_t *in_exp;
    ast_exp_t *post_exp;
} exp_ternary_t;

// dml, query
typedef struct exp_sql_s {
    sql_kind_t kind;
    char *sql;
} exp_sql_t;

// (exp, exp, exp, ...)
typedef struct exp_tuple_s {
    array_t *exps;
} exp_tuple_t;

struct ast_exp_s {
    AST_NODE_DECL;

    exp_kind_t kind;

    union {
        exp_val_t u_val;
        exp_type_t u_type;
        exp_id_t u_id;
        exp_array_t u_arr;
        exp_call_t u_call;
        exp_access_t u_acc;
        exp_op_t u_op;
        exp_ternary_t u_tern;
        exp_sql_t u_sql;
        exp_tuple_t u_tup;
    };

    /* results of semantic checker */
    ast_id_t *id;
    meta_t meta;
};

ast_exp_t *exp_new_null(src_pos_t *pos);
ast_exp_t *exp_new_val(src_pos_t *pos);
ast_exp_t *exp_new_type(type_t type, src_pos_t *pos);
ast_exp_t *exp_new_id(char *name, src_pos_t *pos);
ast_exp_t *exp_new_array(ast_exp_t *id_exp, ast_exp_t *idx_exp, src_pos_t *pos);
ast_exp_t *exp_new_call(ast_exp_t *id_exp, array_t *param_exps, src_pos_t *pos);
ast_exp_t *exp_new_access(ast_exp_t *id_exp, ast_exp_t *fld_exp, src_pos_t *pos);
ast_exp_t *exp_new_op(op_kind_t kind, ast_exp_t *l_exp, ast_exp_t *r_exp,
                      src_pos_t *pos);
ast_exp_t *exp_new_ternary(ast_exp_t *pre_exp, ast_exp_t *in_exp, ast_exp_t *post_exp, 
                           src_pos_t *pos);
ast_exp_t *exp_new_sql(sql_kind_t kind, char *sql, src_pos_t *pos);
ast_exp_t *exp_new_tuple(array_t *exps, src_pos_t *pos);

int exp_eval_const(ast_exp_t *exp, meta_t *meta);

void ast_exp_dump(ast_exp_t *exp, int indent);

#endif /* ! _AST_EXP_H */
