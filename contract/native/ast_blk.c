/**
 * @file    ast_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_id.h"
#include "ast_stmt.h"

#include "ast_blk.h"

static ast_blk_t *
ast_blk_new(blk_kind_t kind, src_pos_t *pos)
{
    ast_blk_t *blk = xcalloc(sizeof(ast_blk_t));

    ast_node_init(blk, pos);

    blk->kind = kind;

    array_init(&blk->ids);
    array_init(&blk->stmts);

    return blk;
}

ast_blk_t *
blk_new_normal(src_pos_t *pos)
{
    return ast_blk_new(BLK_NORMAL, pos);
}

ast_blk_t *
blk_new_root(src_pos_t *pos)
{
    return ast_blk_new(BLK_ROOT, pos);
}

ast_blk_t *
blk_new_loop(src_pos_t *pos)
{
    return ast_blk_new(BLK_LOOP, pos);
}

ast_blk_t *
blk_new_switch(src_pos_t *pos)
{
    return ast_blk_new(BLK_SWITCH, pos);
}

ast_blk_t *
blk_search(ast_blk_t *blk, blk_kind_t kind)
{
    ASSERT(blk != NULL);

    do {
        if (blk->kind == kind)
            return blk;
    } while ((blk = blk->up) != NULL);

    return NULL;
}

ast_id_t *
blk_search_id(ast_blk_t *blk, char *name, int num)
{
    int i, j;

    ASSERT(name != NULL);

    if (blk == NULL)
        return NULL;

    do {
        /* TODO: better to skip if it is equal to the current contract id */
        array_foreach(&blk->ids, i) {
            ast_id_t *id = array_get_id(&blk->ids, i);

            if (is_label_id(id))
                continue;

            if (is_tuple_id(id)) {
                array_foreach(id->u_tup.elem_ids, j) {
                    ast_id_t *elem_id = array_get_id(id->u_tup.elem_ids, j);

                    if (elem_id->num < num && strcmp(elem_id->name, name) == 0)
                        return elem_id;
                }
            }
            else if (id->num < num && strcmp(id->name, name) == 0) {
                return id;
            }
        }
    } while ((blk = blk->up) != NULL);

    return NULL;
}

ast_id_t *
blk_search_label(ast_blk_t *blk, char *name)
{
    int i;

    ASSERT(name != NULL);

    if (blk == NULL)
        return NULL;

    array_foreach(&blk->ids, i) {
        ast_id_t *id = array_get_id(&blk->ids, i);

        if (is_label_id(id) && strcmp(id->name, name) == 0)
            return id;
    }

    return NULL;
}

/* end of ast_blk.c */
