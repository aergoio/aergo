/**
 * @file    ast_meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_exp.h"

#include "ast_meta.h"

void
meta_set_map(ast_meta_t *meta, type_t k_type, ast_meta_t *v_meta)
{
    ast_meta_init(meta, TYPE_MAP);

    meta->u_map.k_type = k_type;
    meta->u_map.v_meta = v_meta;
}

void
meta_set_tuple(ast_meta_t *meta, array_t *exps)
{
    int i;
    ast_exp_t *exp;

    ASSERT(exps != NULL);

    ast_meta_init(meta, TYPE_TUPLE);

    meta->u_tup.size = array_size(exps);
    meta->u_tup.metas = xmalloc(sizeof(ast_meta_t *) * meta->u_tup.size);

    for (i = 0; i < array_size(exps); i++) {
        ast_exp_t *exp = array_item(exps, i, ast_exp_t);

        ASSERT1(type_is_valid(exp->meta.type), exp->meta.type);
        meta->u_tup.metas[i++] = &exp->meta;
    }
}

void 
ast_meta_dump(ast_meta_t *meta, int indent)
{
}

/* end of ast_meta.c */
