/**
 * @file    ast_var.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_exp.h"

#include "ast_var.h"

ast_var_t *
ast_var_new(ast_exp_t *id_exp, ast_exp_t *init_exp, errpos_t *pos)
{
    ast_var_t *var = xmalloc(sizeof(ast_var_t));

    ASSERT(id_exp != NULL);

    list_link_init(&var->link);
    var->pos = *pos;

    var->meta = id_exp->meta;

    var->type_exp = NULL;
    var->id_exp = id_exp;
    var->init_exp = init_exp;

    return var;
} 

/* end of ast_var.c */
