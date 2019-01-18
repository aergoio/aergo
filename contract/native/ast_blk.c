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
ast_blk_new(blk_kind_t kind)
{
    ast_blk_t *blk = xcalloc(sizeof(ast_blk_t));

    blk->kind = kind;

    array_init(&blk->ids);
    array_init(&blk->nodes);

    return blk;
}

ast_blk_t *
blk_new_normal(void)
{
    return ast_blk_new(BLK_NORMAL);
}

ast_blk_t *
blk_new_root(void)
{
    return ast_blk_new(BLK_ROOT);
}

ast_blk_t *
blk_new_contract(void)
{
    return ast_blk_new(BLK_CONT);
}

ast_blk_t *
blk_new_interface(void)
{
    return ast_blk_new(BLK_ITF);
}

ast_blk_t *
blk_new_fn(void)
{
    return ast_blk_new(BLK_FN);
}

ast_blk_t *
blk_new_loop(void)
{
    return ast_blk_new(BLK_LOOP);
}

ast_blk_t *
blk_new_switch(void)
{
    return ast_blk_new(BLK_SWITCH);
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

static bool
is_visible_id(ast_blk_t *blk, ast_id_t *id, char *name, int num, bool is_type)
{
    if (is_root_blk(blk))
        return strcmp(name, id->name) == 0;

    if (is_cont_blk(blk)) {
        if (is_type)
            return is_struct_id(id) && strcmp(name, id->name) == 0;

        if (is_fn_id(id))
            return strcmp(name, id->name) == 0;
    }

    return id->num < num && strcmp(name, id->name) == 0;
}

ast_id_t *
blk_search_id(ast_blk_t *blk, char *name, int num, bool is_type)
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

                    if (is_visible_id(blk, elem_id, name, num, is_type))
                        return elem_id;
                }
            }
            else if (is_visible_id(blk, id, name, num, is_type)) {
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
