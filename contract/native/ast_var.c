/**
 * @file    ast_var.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_exp.h"
#include "util.h"

#include "ast_var.h"

ast_var_t *
ast_var_new(ast_exp_t *type_exp, ast_exp_t *id_exp, ast_exp_t *init_exp,
            errpos_t *pos)
{
    ast_var_t *var = xmalloc(sizeof(ast_var_t));

    ASSERT(id_exp != NULL);

    ast_node_init(var, pos);

    var->type_exp = type_exp;
    var->id_exp = id_exp;
    var->init_exp = init_exp;

    var->name = NULL;

    return var;
}

void
ast_var_dump(ast_var_t *var, int indent)
{
}

/* end of ast_var.c */
