/**
 * @file    ast_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_EXP_H
#define _AST_EXP_H

#include "common.h"

#include "ast.h"
#include "meta.h"
#include "value.h"

#define is_null_exp(exp)            ((exp)->kind == EXP_NULL)
#define is_lit_exp(exp)             ((exp)->kind == EXP_LIT)
#define is_id_exp(exp)              ((exp)->kind == EXP_ID)
#define is_type_exp(exp)            ((exp)->kind == EXP_TYPE)
#define is_array_exp(exp)           ((exp)->kind == EXP_ARRAY)
#define is_cast_exp(exp)            ((exp)->kind == EXP_CAST)
#define is_unary_exp(exp)           ((exp)->kind == EXP_UNARY)
#define is_binary_exp(exp)          ((exp)->kind == EXP_BINARY)
#define is_ternary_exp(exp)         ((exp)->kind == EXP_TERNARY)
#define is_access_exp(exp)          ((exp)->kind == EXP_ACCESS)
#define is_call_exp(exp)            ((exp)->kind == EXP_CALL)
#define is_sql_exp(exp)             ((exp)->kind == EXP_SQL)
#define is_tuple_exp(exp)           ((exp)->kind == EXP_TUPLE)
#define is_init_exp(exp)            ((exp)->kind == EXP_INIT)
#define is_alloc_exp(exp)           ((exp)->kind == EXP_ALLOC)
#define is_global_exp(exp)          ((exp)->kind == EXP_GLOBAL)
#define is_register_exp(exp)        ((exp)->kind == EXP_REGISTER)
#define is_memory_exp(exp)          ((exp)->kind == EXP_MEMORY)

#define is_usable_lval(exp)         (exp)->usable_lval

#define is_usable_stmt(exp)                                                              \
    (is_null_exp(exp) || is_call_exp(exp) ||                                             \
     (is_sql_exp(exp) && exp->u_sql.kind != SQL_QUERY) ||                                \
     (is_unary_exp(exp) && (exp->u_un.kind == OP_INC || exp->u_un.kind == OP_DEC)))

#define exp_add                     vector_add_last

#define exp_check_overflow(exp, meta)                                                    \
    do {                                                                                 \
        if (is_lit_exp(exp) && !value_fit(&(exp)->u_lit.val, (meta)))                    \
            ERROR(ERROR_NUMERIC_OVERFLOW, &(exp)->pos, meta_to_str(meta));               \
    } while (0)

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

#ifndef _VECTOR_T
#define _VECTOR_T
typedef struct vector_s vector_t;
#endif /* ! _VECTOR_T */

/* null, true, false, 1, 1.0, 0x1, "..." */
typedef struct exp_lit_s {
    value_t val;
} exp_lit_t;

/* name */
typedef struct exp_id_s {
    char *name;
} exp_id_t;

/* primitive, struct, map */
typedef struct exp_type_s {
    type_t type;
    char *name;
    ast_exp_t *k_exp;
    ast_exp_t *v_exp;
} exp_type_t;

/* id[idx] */
typedef struct exp_array_s {
    ast_exp_t *id_exp;
    ast_exp_t *idx_exp;
} exp_array_t;

/* (type)val */
typedef struct exp_cast_s {
    ast_exp_t *val_exp;
    meta_t to_meta;
} exp_cast_t;

/* id(param, ...) */
typedef struct exp_call_s {
    bool is_ctor;
    ast_exp_t *id_exp;
    vector_t *param_exps;
} exp_call_t;

/* id.fld */
typedef struct exp_access_s {
    ast_exp_t *qual_exp;
    ast_exp_t *fld_exp;
} exp_access_t;

/* val kind */
typedef struct exp_unary_s {
    op_kind_t kind;
    bool is_prefix;
    ast_exp_t *val_exp;
} exp_unary_t;

/* l kind r */
typedef struct exp_binary_s {
    op_kind_t kind;
    ast_exp_t *l_exp;
    ast_exp_t *r_exp;
} exp_binary_t;

/* prefix ? infix : postfix */
typedef struct exp_ternary_s {
    ast_exp_t *pre_exp;
    ast_exp_t *in_exp;
    ast_exp_t *post_exp;
} exp_ternary_t;

/* dml, query */
typedef struct exp_sql_s {
    sql_kind_t kind;
    char *sql;
} exp_sql_t;

/* (exp, exp, exp, ...) */
typedef struct exp_tuple_s {
    vector_t *elem_exps;
} exp_tuple_t;

/* new {exp, exp, exp, ...} */
typedef struct exp_init_s {
    bool is_aggr;
    vector_t *elem_exps;
} exp_init_t;

/* new type, new type[] */
typedef struct exp_alloc_s {
    ast_exp_t *type_exp;
    vector_t *size_exps;
} exp_alloc_t;

typedef struct exp_global_s {
    char *name;
} exp_global_t;

typedef struct exp_register_s {
    type_t type;
    uint32_t idx;
} exp_register_t;

typedef struct exp_memory_s {
    type_t type;
    uint32_t base;
    uint32_t addr;
    uint32_t offset;
} exp_memory_t;

struct ast_exp_s {
    exp_kind_t kind;

    union {
        exp_lit_t u_lit;
        exp_id_t u_id;
        exp_type_t u_type;
        exp_array_t u_arr;
        exp_cast_t u_cast;
        exp_call_t u_call;
        exp_access_t u_acc;
        exp_unary_t u_un;
        exp_binary_t u_bin;
        exp_ternary_t u_tern;
        exp_sql_t u_sql;
        exp_tuple_t u_tup;
        exp_init_t u_init;
        exp_alloc_t u_alloc;
        exp_global_t u_glob;
        exp_register_t u_reg;
        exp_memory_t u_mem;
    };

    ast_id_t *id;           /* referenced identifier */
    meta_t meta;

    bool usable_lval;       /* whether is used as lvalue */

    AST_NODE_DECL;
};

ast_exp_t *exp_new_null(src_pos_t *pos);
ast_exp_t *exp_new_lit_null(src_pos_t *pos);
ast_exp_t *exp_new_lit_bool(bool v, src_pos_t *pos);
ast_exp_t *exp_new_lit_i64(uint64_t v, src_pos_t *pos);
ast_exp_t *exp_new_lit_f64(double v, src_pos_t *pos);
ast_exp_t *exp_new_lit_str(char *v, src_pos_t *pos);
ast_exp_t *exp_new_id(char *name, src_pos_t *pos);
ast_exp_t *exp_new_type(type_t type, src_pos_t *pos);
ast_exp_t *exp_new_array(ast_exp_t *id_exp, ast_exp_t *idx_exp, src_pos_t *pos);
ast_exp_t *exp_new_cast(type_t type, ast_exp_t *val_exp, src_pos_t *pos);
ast_exp_t *exp_new_call(bool is_ctor, ast_exp_t *id_exp, vector_t *param_exps,
                        src_pos_t *pos);
ast_exp_t *exp_new_access(ast_exp_t *qual_exp, ast_exp_t *fld_exp, src_pos_t *pos);
ast_exp_t *exp_new_unary(op_kind_t kind, bool is_prefix, ast_exp_t *val_exp,
                         src_pos_t *pos);
ast_exp_t *exp_new_binary(op_kind_t kind, ast_exp_t *l_exp, ast_exp_t *r_exp,
                          src_pos_t *pos);
ast_exp_t *exp_new_ternary(ast_exp_t *pre_exp, ast_exp_t *in_exp, ast_exp_t *post_exp,
                           src_pos_t *pos);
ast_exp_t *exp_new_sql(sql_kind_t kind, char *sql, src_pos_t *pos);
ast_exp_t *exp_new_tuple(vector_t *elem_exps, src_pos_t *pos);
ast_exp_t *exp_new_init(vector_t *elem_exps, src_pos_t *pos);
ast_exp_t *exp_new_alloc(ast_exp_t *type_exp, src_pos_t *pos);

ast_exp_t *exp_new_global(char *name);
ast_exp_t *exp_new_register(type_t type, uint32_t idx);
ast_exp_t *exp_new_memory(type_t type, uint32_t base, uint32_t addr, uint32_t offset);

void exp_set_lit(ast_exp_t *exp, value_t *val);
void exp_set_register(ast_exp_t *exp, uint32_t idx);
void exp_set_memory(ast_exp_t *exp, uint32_t base, uint32_t addr, uint32_t offset);

ast_exp_t *exp_clone(ast_exp_t *exp);

bool exp_equals(ast_exp_t *x, ast_exp_t *y);

#endif /* ! _AST_EXP_H */
