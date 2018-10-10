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

#define is_usable_lval(exp)                                                    \
    (!is_const_meta(&(exp)->meta) && !is_untyped_meta(&(exp)->meta) &&         \
     (is_id_exp(exp) || is_array_exp(exp) || is_access_exp(exp)))

#define exp_pos(exp)                (&(exp)->meta.trc)

#define ast_exp_add                 array_add_tail
#define ast_exp_merge               array_join

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

typedef enum exp_kind_e {
    EXP_NULL        = 0,
    EXP_VAL,
    EXP_TYPE,
    EXP_ID,
    EXP_ARRAY,
    EXP_OP,
    EXP_ACCESS,
    EXP_CALL,
    EXP_SQL,
    EXP_TERNARY,
    EXP_TUPLE,
    EXP_MAX
} exp_kind_t;

typedef enum op_kind_e {
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
} op_kind_t;

typedef enum sql_kind_e {
    SQL_QUERY       = 0,
    SQL_INSERT,
    SQL_UPDATE,
    SQL_DELETE,
    SQL_MAX
} sql_kind_t;

// null, true, false, 1, 1.0, 0x1, "..."
typedef struct exp_val_s {
    value_t val;
} exp_val_t;

// primitive, struct, map
typedef struct exp_type_s {
    type_t type;
    char *name;
    bool is_local;
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

    // results of semantic checker (might be a part of ir_exp_t in later)
    ast_id_t *id;
    value_t val;
};

ast_exp_t *ast_exp_new(exp_kind_t kind, trace_t *trc);

ast_exp_t *exp_null_new(trace_t *trc);
ast_exp_t *exp_val_new(trace_t *trc);
ast_exp_t *exp_type_new(type_t type, char *name, ast_exp_t *k_exp,
                        ast_exp_t *v_exp, trace_t *trc);
ast_exp_t *exp_id_new(char *name, trace_t *trc);
ast_exp_t *exp_array_new(ast_exp_t *id_exp, ast_exp_t *idx_exp,
                         trace_t *trc);
ast_exp_t *exp_call_new(ast_exp_t *id_exp, array_t *param_exps, trace_t *trc);
ast_exp_t *exp_access_new(ast_exp_t *id_exp, ast_exp_t *fld_exp,
                          trace_t *trc);
ast_exp_t *exp_op_new(op_kind_t kind, ast_exp_t *l_exp, ast_exp_t *r_exp,
                      trace_t *trc);
ast_exp_t *exp_ternary_new(ast_exp_t *pre_exp, ast_exp_t *in_exp,
                           ast_exp_t *post_exp, trace_t *trc);
ast_exp_t *exp_sql_new(sql_kind_t kind, char *sql, trace_t *trc);
ast_exp_t *exp_tuple_new(ast_exp_t *exp, trace_t *trc);

void ast_exp_dump(ast_exp_t *exp, int indent);

#endif /* ! _AST_EXP_H */
