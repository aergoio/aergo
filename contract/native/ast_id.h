/**
 * @file    ast_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_ID_H
#define _AST_ID_H

#include "common.h"

#include "ast.h"
#include "vector.h"
#include "enum.h"
#include "meta.h"
#include "value.h"

#define is_var_id(id)               ((id)->kind == ID_VAR)
#define is_struct_id(id)            ((id)->kind == ID_STRUCT)
#define is_enum_id(id)              ((id)->kind == ID_ENUM)
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

#define is_type_id(id)              (is_struct_id(id) || is_cont_id(id) || is_itf_id(id))
#define is_param_id(id)             (is_var_id(id) && (id)->u_var.kind != PARAM_NONE)

#define is_in_param(id)             (is_var_id(id) && (id)->u_var.kind == PARAM_IN)
#define is_out_param(id)            (is_var_id(id) && (id)->u_var.kind == PARAM_OUT)

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

#ifndef _IR_ABI_T
#define _IR_ABI_T
typedef struct ir_abi_s ir_abi_t;
#endif /* ! _IR_ABI_T */

typedef struct id_var_s {
    param_kind_t kind;
    ast_exp_t *type_exp;
    vector_t *size_exps;
    ast_exp_t *dflt_exp;
} id_var_t;

typedef struct id_struct_s {
    vector_t *fld_ids;
} id_struct_t;

typedef struct id_enum_s {
    vector_t *elem_ids;
} id_enum_t;

typedef struct id_fn_s {
    vector_t *param_ids;
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
    ast_exp_t *type_exp;
    vector_t *elem_ids;
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
        id_fn_t u_fn;
        id_cont_t u_cont;
        id_itf_t u_itf;
        id_label_t u_lab;
        id_tuple_t u_tup;
    };

    bool is_used;       /* whether it is referenced */
    bool is_checked;    /* whether it is checked */

    meta_t meta;
    value_t *val;       /* constant value */

    int idx;            /* local or function index */

    ast_id_t *up;
    ir_abi_t *abi;

    AST_NODE_DECL;
};

ast_id_t *id_new_var(char *name, modifier_t mod, src_pos_t *pos);
ast_id_t *id_new_param(param_kind_t kind, char *name, ast_exp_t *type_exp,
                       src_pos_t *pos);
ast_id_t *id_new_struct(char *name, vector_t *fld_ids, src_pos_t *pos);
ast_id_t *id_new_enum(char *name, vector_t *elem_ids, src_pos_t *pos);
ast_id_t *id_new_func(char *name, modifier_t mod, vector_t *param_ids, ast_id_t *ret_id,
                      ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_ctor(char *name, vector_t *param_ids, ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_contract(char *name, ast_exp_t *impl_exp, ast_blk_t *blk,
                          src_pos_t *pos);
ast_id_t *id_new_interface(char *name, ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_label(char *name, ast_stmt_t *stmt, src_pos_t *pos);
ast_id_t *id_new_tuple(src_pos_t *pos);

ast_id_t *id_new_tmp_var(char *name);

ast_id_t *id_search_fld(ast_id_t *id, char *name, bool is_self);
ast_id_t *id_search_param(ast_id_t *id, char *name);

void id_add(vector_t *ids, ast_id_t *new_id);
void id_join(vector_t *ids, vector_t *new_ids);

vector_t *id_strip(ast_id_t *id);

bool id_cmp(ast_id_t *x, ast_id_t *y);

#endif /* ! _AST_ID_H */
