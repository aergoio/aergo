/**
 * @file    ast_imp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_imp.h"

ast_imp_t *
imp_new(char *path, src_pos_t *pos)
{
    ast_imp_t *imp = xcalloc(sizeof(ast_imp_t));

    ast_node_init(imp, *pos);

    imp->path = path;

    return imp;
}

/* end of ast_imp.c */
