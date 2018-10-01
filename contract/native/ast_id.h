/**
 * @file    ast_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_ID_H
#define _AST_ID_H

#include "common.h"

#include "ast.h"

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_META_T
#define _AST_META_T
typedef struct ast_meta_s ast_meta_t;
#endif /* ! _AST_META_T */

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
    MOD_TRANSFER    = 0x02,
    MOD_READONLY    = 0x04
} modifier_t;

typedef struct id_var_s {
    ast_exp_t *type_exp;
    ast_exp_t *id_exp;
    ast_exp_t *init_exp;
} id_var_t;

typedef struct id_struct_s {
    list_t *field_l;
} id_struct_t;

typedef struct id_func_s {
    list_t *param_l;
    list_t *return_l;

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
    ast_meta_t *meta;
};

ast_id_t *ast_id_new(id_kind_t kind, modifier_t mod, errpos_t *pos);

ast_id_t *id_var_new(ast_exp_t *type_exp, ast_exp_t *id_exp,
                     ast_exp_t *init_exp, errpos_t *pos);

ast_id_t *id_struct_new(char *name, list_t *field_l, errpos_t *pos);

ast_id_t *id_func_new(char *name, modifier_t mod, list_t *param_l,
                      list_t *return_l, ast_blk_t *blk, errpos_t *pos);

ast_id_t *id_contract_new(char *name, ast_blk_t *blk, errpos_t *pos);

ast_id_t *ast_id_search(ast_blk_t *blk, int num, char *name);

void ast_id_dump(ast_id_t *id, int indent);

#endif /* ! _AST_ID_H */
