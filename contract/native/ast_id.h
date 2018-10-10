/**
 * @file    ast_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_ID_H
#define _AST_ID_H

#include "common.h"

#include "ast.h"

#define is_var_id(id)               ((id)->kind == ID_VAR)
#define is_struct_id(id)            ((id)->kind == ID_STRUCT)
#define is_func_id(id)              ((id)->kind == ID_FUNC)
#define is_contract_id(id)          ((id)->kind == ID_CONTRACT)

#define id_ctor_new(name, params, blk, trc)                                    \
    id_func_new((name), MOD_INITIAL, params, NULL, blk, (trc))

#define id_pos(id)                  (&(id)->meta.trc)

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

typedef enum id_kind_e {
    ID_VAR          = 0,
    ID_STRUCT,
    ID_FUNC,
    ID_CONTRACT,
    ID_MAX
} id_kind_t;

typedef enum modifier_e {
    MOD_GLOBAL      = 0x00,
    MOD_LOCAL       = 0x01,
    MOD_PAYABLE     = 0x02,
    MOD_READONLY    = 0x04,
    MOD_INITIAL     = 0x08
} modifier_t;

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
};

ast_id_t *ast_id_new(id_kind_t kind, modifier_t mod, char *name, trace_t *trc);

ast_id_t *id_var_new(char *name, trace_t *trc);
ast_id_t *id_struct_new(char *name, array_t *fld_ids, trace_t *trc);
ast_id_t *id_func_new(char *name, modifier_t mod, array_t *param_ids,
                      array_t *ret_exps, ast_blk_t *blk, trace_t *trc);
ast_id_t *id_contract_new(char *name, ast_blk_t *blk, trace_t *trc);

ast_id_t *ast_id_search_fld(ast_id_t *id, int num, char *name);
ast_id_t *ast_id_search_blk(ast_blk_t *blk, int num, char *name);
ast_id_t *ast_id_search_param(ast_id_t *id, int num, char *name);

void ast_id_add(array_t *ids, ast_id_t *new_id);
void ast_id_join(array_t *ids, array_t *new_ids);

void ast_id_dump(ast_id_t *id, int indent);

#endif /* ! _AST_ID_H */
