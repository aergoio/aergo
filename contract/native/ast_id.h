/**
 * @file    ast_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_ID_H
#define _AST_ID_H

#include "common.h"

#include "ast.h"
#include "array.h"
#include "enum.h"
#include "meta.h"
#include "value.h"

#define is_var_id(id)               ((id)->kind == ID_VAR)
#define is_struct_id(id)            ((id)->kind == ID_STRUCT)
#define is_enum_id(id)              ((id)->kind == ID_ENUM)
#define is_return_id(id)            ((id)->kind == ID_RETURN)
#define is_fn_id(id)                ((id)->kind == ID_FN)
#define is_cont_id(id)              ((id)->kind == ID_CONT)
#define is_itf_id(id)               ((id)->kind == ID_ITF)
#define is_label_id(id)             ((id)->kind == ID_LABEL)
#define is_tuple_id(id)             ((id)->kind == ID_TUPLE)

#define is_public_id(id)            flag_on((id)->mod, MOD_PUBLIC)
#define is_private_id(id)           flag_on((id)->mod, MOD_PRIVATE)
#define is_payable_id(id)           flag_on((id)->mod, MOD_PAYABLE)
#define is_readonly_id(id)          flag_on((id)->mod, MOD_READONLY)
#define is_const_id(id)             flag_on((id)->mod, MOD_CONST)
#define is_ctor_id(id)              flag_on((id)->mod, MOD_CTOR)

#define is_global_id(id)            (id->up != NULL && is_cont_id(id->up))
#define is_local_id(id)             (id->up != NULL && !is_cont_id(id->up))

#define is_stack_id(id)                                                                  \
    (!(id)->is_param && (is_array_type(&(id)->meta) || !is_primitive_type(&(id)->meta)))

#define is_type_id(id)              (is_struct_id(id) || is_cont_id(id) || is_itf_id(id))

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_STMT_T
#define _AST_STMT_T
typedef struct ast_stmt_s ast_stmt_t;
#endif /* ! _AST_STMT_T */

typedef struct id_var_s {
    meta_t *type_meta;
    array_t *size_exps;
    ast_exp_t *dflt_exp;
} id_var_t;

typedef struct id_struct_s {
    array_t *fld_ids;
} id_struct_t;

typedef struct id_enum_s {
    array_t *elem_ids;
} id_enum_t;

typedef struct id_return_s {
    meta_t *type_meta;
    array_t *size_exps;
} id_return_t;

typedef struct id_fn_s {
    array_t *param_ids;
    ast_id_t *ret_id;
    ast_blk_t *blk;
} id_fn_t;

typedef struct id_cont_s {
    ast_exp_t *impl_exp;
    ast_blk_t *blk;
} id_cont_t;

typedef struct id_itf_s {
    ast_blk_t *blk;
} id_itf_t;

typedef struct id_label_s {
    ast_stmt_t *stmt;
} id_label_t;

typedef struct id_tuple_s {
    meta_t *type_meta;
    array_t *elem_ids;
    ast_exp_t *dflt_exp;
} id_tuple_t;

struct ast_id_s {
    id_kind_t kind;
    modifier_t mod;     /* public or const */

    char *name;

    union {
        id_var_t u_var;
        id_struct_t u_struc;
        id_enum_t u_enum;
        id_return_t u_ret;
        id_fn_t u_fn;
        id_cont_t u_cont;
        id_itf_t u_itf;
        id_label_t u_lab;
        id_tuple_t u_tup;
    };

    bool is_used;       /* whether it is referenced */
    bool is_checked;    /* whether it is checked */
    bool is_param;      /* whether it is a parameter */

    meta_t meta;
    value_t *val;       /* constant value */

    int idx;            /* local index */
    int addr;           /* relative address */
    int offset;         /* offset (from addr) */

    ast_id_t *up;

    AST_NODE_DECL;
};

ast_id_t *id_new_var(char *name, modifier_t mod, src_pos_t *pos);
ast_id_t *id_new_struct(char *name, array_t *fld_ids, src_pos_t *pos);
ast_id_t *id_new_enum(char *name, array_t *elem_ids, src_pos_t *pos);
ast_id_t *id_new_return(char *name, meta_t *type_meta, src_pos_t *pos);
ast_id_t *id_new_func(char *name, modifier_t mod, array_t *param_ids, ast_id_t *ret_id,
                      ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_ctor(char *name, array_t *param_ids, ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_contract(char *name, ast_exp_t *impl_exp, ast_blk_t *blk,
                          src_pos_t *pos);
ast_id_t *id_new_interface(char *name, ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_label(char *name, ast_stmt_t *stmt, src_pos_t *pos);
ast_id_t *id_new_tuple(src_pos_t *pos);

ast_id_t *id_lookup_fld(ast_id_t *id, char *name, bool is_self);
ast_id_t *id_lookup_param(ast_id_t *id, char *name);

void id_add(array_t *ids, ast_id_t *new_id);
void id_join(array_t *ids, array_t *new_ids);

array_t *id_strip(ast_id_t *id);

bool id_cmp(ast_id_t *x, ast_id_t *y);

#endif /* ! _AST_ID_H */