/**
 * @file    ast_var.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_VAR_H
#define _AST_VAR_H

#include "common.h"

#include "ast.h"
#include "ast_meta.h"

#ifndef _AST_VAR_T
#define _AST_VAR_T
typedef struct ast_var_s ast_var_t;
#endif /* ! _AST_VAR_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

struct ast_var_s {
    AST_NODE_DECL;

    ast_exp_t *type_exp;
    ast_exp_t *id_exp;
    ast_exp_t *init_exp;
};

ast_var_t *ast_var_new(ast_exp_t *type_exp, ast_exp_t *id_exp, 
                       ast_exp_t *init_exp, errpos_t *pos);

#endif /* ! _AST_VAR_H */
