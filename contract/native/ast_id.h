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
#define is_func_id(id)              ((id)->kind == ID_FUNC)
#define is_contract_id(id)          ((id)->kind == ID_CONTRACT)
#define is_label_id(id)             ((id)->kind == ID_LABEL)

#define is_public_id(id)            flag_on((id)->mod, MOD_PUBLIC)
#define is_private_id(id)           flag_on((id)->mod, MOD_PRIVATE)
#define is_payable_id(id)           flag_on((id)->mod, MOD_PAYABLE)
#define is_readonly_id(id)          flag_on((id)->mod, MOD_READONLY)
#define is_const_id(id)             flag_on((id)->mod, MOD_CONST)
#define is_ctor_id(id)              flag_on((id)->mod, MOD_CTOR)

#define is_global_id(id)            ((id)->scope == SCOPE_GLOBAL)
#define is_local_id(id)             ((id)->scope == SCOPE_LOCAL)

#define id_new_ctor(name, pos)                                                           \
    id_new_func((name), MOD_PUBLIC | MOD_CTOR, NULL, NULL, NULL, (pos))

#define id_add_first(ids, new_id)   id_add((ids), 0, (new_id))
#define id_add_last(ids, new_id)    id_add((ids), array_size(ids), (new_id))

#define id_join_first(ids, new_ids) id_join((ids), 0, (new_ids))
#define id_join_last(ids, new_ids)  id_join((ids), array_size(ids), (new_ids))

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
    ast_exp_t *dflt_exp;
    array_t *size_exps;
} id_var_t;

typedef struct id_struct_s {
    array_t *fld_ids;
} id_struct_t;

typedef struct id_enum_s {
    array_t *elem_ids;
} id_enum_t;

typedef struct id_func_s {
    array_t *param_ids;
    array_t *ret_ids;
    ast_blk_t *blk;
} id_func_t;

typedef struct id_cont_s {
    ast_blk_t *blk;
} id_cont_t;

typedef struct id_label_s {
    ast_stmt_t *stmt;
} id_label_t;

struct ast_id_s {
    AST_NODE_DECL;

    id_kind_t kind;
    modifier_t mod;
    char *name;

    union {
        id_var_t u_var;
        id_struct_t u_struc;
        id_enum_t u_enum;
        id_func_t u_func;
        id_cont_t u_cont;
        id_label_t u_label;
    };

    // results of semantic checker
    bool is_used;       /* whether it is referenced */
    bool is_checked;    /* whether it is checked */

    scope_t scope;      /* local or global */

    meta_t meta;        /* identifier metadata */
    value_t *val;       /* constant value */

    int idx;            /* local index */
    int offset;         /* relative offset */
};

ast_id_t *id_new_var(char *name, modifier_t mod, src_pos_t *pos);
ast_id_t *id_new_struct(char *name, array_t *fld_ids, src_pos_t *pos);
ast_id_t *id_new_enum(char *name, array_t *elem_ids, src_pos_t *pos);
ast_id_t *id_new_func(char *name, modifier_t mod, array_t *param_ids, array_t *ret_ids,
                      ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_contract(char *name, ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_new_label(char *name, ast_stmt_t *stmt, src_pos_t *pos);

ast_id_t *id_search_name(ast_blk_t *blk, char *name, int num);
ast_id_t *id_search_fld(ast_id_t *id, char *name, bool is_self);
ast_id_t *id_search_param(ast_id_t *id, char *name);
ast_id_t *id_search_label(ast_blk_t *blk, char *name);

void id_add(array_t *ids, int idx, ast_id_t *new_id);
void id_join(array_t *ids, int idx, array_t *new_ids);

void ast_id_dump(ast_id_t *id, int indent);

#endif /* ! _AST_ID_H */
