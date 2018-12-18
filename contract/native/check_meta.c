/**
 * @file    check_meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"

#include "check_meta.h"

int
meta_check(check_t *check, meta_t *meta)
{
    if (is_struct_type(meta)) {
        ast_id_t *id;

        if (meta->elem_cnt > 0)
            return NO_ERROR;

        ASSERT(meta->name != NULL);

        if (check->qual_id != NULL)
            id = id_search_fld(check->qual_id, meta->name,
                               check->qual_id == check->cont_id);
        else
            id = blk_search_id(check->blk, meta->name, meta->num);

        if (id == NULL || (!is_struct_id(id) && !is_cont_id(id)))
            RETURN(ERROR_UNDEFINED_TYPE, &meta->pos, meta->name);

        id->is_used = true;
        *meta = id->meta;
    }
    else if (is_map_type(meta)) {
        meta_t *k_meta, *v_meta;

        ASSERT1(meta->elem_cnt == 2, meta->elem_cnt);

        k_meta = meta->elems[0];
        v_meta = meta->elems[1];

        CHECK(meta_check(check, k_meta));
        CHECK(meta_check(check, v_meta));

        if (!is_comparable_type(k_meta))
            RETURN(ERROR_NOT_COMPARABLE_TYPE, &k_meta->pos, meta_to_str(k_meta));

        ASSERT(!is_tuple_type(v_meta));
    }
    else {
        ASSERT(meta->name == NULL);
        ASSERT1(meta->elem_cnt == 0, meta->elem_cnt);
    }

    return NO_ERROR;
}

/* end of check_meta.c */
