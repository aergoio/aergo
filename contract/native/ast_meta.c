/**
 * @file    ast_meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_exp.h"

#include "ast_meta.h"

void
meta_set_struct(ast_meta_t *meta, list_t *field_l)
{
    int i = 0;
    ast_exp_t *exp;
    list_node_t *node;

    ASSERT(field_l != NULL);

    ast_meta_init(meta, TYPE_STRUCT);

    meta->u_st.fld_cnt = list_size(field_l);
    meta->u_st.fld_metas = xmalloc(sizeof(ast_meta_t *) * meta->u_st.fld_cnt);

    list_foreach(node, field_l) {
        ast_exp_t *exp = (ast_exp_t *)node->item;

        ASSERT1(type_is_valid(exp->meta.type), exp->meta.type);

        meta->u_st.fld_metas[i++] = &exp->meta;
    }
}

void
meta_set_map(ast_meta_t *meta, type_t k_type, ast_meta_t *v_meta)
{
    ast_meta_init(meta, TYPE_MAP);

    meta->type = TYPE_MAP;
    meta->u_map.k_type = k_type;
    meta->u_map.v_meta = v_meta;
}

void 
ast_meta_dump(ast_meta_t *meta, int indent)
{
}

/* end of ast_meta.c */
