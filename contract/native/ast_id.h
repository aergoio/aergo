/**
 * @file    ast_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_ID_H
#define _AST_ID_H

#include "common.h"

#include "ast.h"
#include "enum.h"

#define is_var_id(id)               ((id)->kind == ID_VAR)
#define is_struct_id(id)            ((id)->kind == ID_STRUCT)
#define is_func_id(id)              ((id)->kind == ID_FUNC)
#define is_contract_id(id)          ((id)->kind == ID_CONTRACT)

#define is_local_id(id)             flag_on((id)->mod, MOD_LOCAL)
#define is_payable_id(id)           flag_on((id)->mod, MOD_PAYABLE)
#define is_readonly_id(id)          flag_on((id)->mod, MOD_READONLY)
#define is_ctor_id(id)              flag_on((id)->mod, MOD_CTOR)

#define id_ctor_new(name, params, blk, pos)                                    \
    id_func_new((name), MOD_CTOR, (params), NULL, (blk), (pos))

#define id_add_first(ids, new_id)   id_add((ids), 0, (new_id))
#define id_add_last(ids, new_id)    id_add((ids), (ids)->cnt, (new_id))

#define id_join_first(ids, new_ids) id_join((ids), 0, (new_ids))
#define id_join_last(ids, new_ids)  id_join((ids), (ids)->cnt, (new_ids))

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

typedef struct id_var_s {
    ast_exp_t *type_exp;
    ast_exp_t *arr_exp;
    ast_exp_t *init_exp;
} id_var_t;

typedef struct id_struct_s {
    array_t *fld_ids;
} id_struct_t;

typedef struct id_func_s {
    array_t *param_ids;
    array_t *ret_exps;

    ast_blk_t *blk;
} id_func_t;

typedef struct id_cont_s {
    ast_blk_t *blk;
} id_cont_t;

struct ast_id_s {
    AST_NODE_DECL;

    id_kind_t kind;
    modifier_t mod;
    char *name;

    union {
        id_var_t u_var;
        id_struct_t u_st;
        id_func_t u_func;
        id_cont_t u_cont;
    };

    // results of semantic checker
    bool is_used;
    meta_t meta;
};

ast_id_t *id_var_new(char *name, src_pos_t *pos);
ast_id_t *id_struct_new(char *name, array_t *fld_ids, src_pos_t *pos);
ast_id_t *id_func_new(char *name, modifier_t mod, array_t *param_ids,
                      array_t *ret_exps, ast_blk_t *blk, src_pos_t *pos);
ast_id_t *id_contract_new(char *name, ast_blk_t *blk, src_pos_t *pos);

ast_id_t *id_search_name(ast_blk_t *blk, int num, char *name);
ast_id_t *id_search_fld(ast_id_t *id, char *name);
ast_id_t *id_search_param(ast_id_t *id, char *name);
ast_id_t *id_search_ctor(array_t *ids);

void id_add(array_t *ids, int idx, ast_id_t *new_id);
void id_join(array_t *ids, int idx, array_t *new_ids);

void ast_id_dump(ast_id_t *id, int indent);

#endif /* ! _AST_ID_H */
