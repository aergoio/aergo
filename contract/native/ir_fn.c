/**
 * @file    ir_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_bb.h"

#include "ir_fn.h"

ir_fn_t *
fn_new(ast_id_t *id)
{
    int i, j = 0;
    ir_fn_t *fn = xmalloc(sizeof(ir_fn_t));

    fn->name = id->name;

    array_init(&fn->params);
    array_init(&fn->locals);

    for (i = 0; i < array_size(id->u_fn.param_ids); i++) {
        ast_id_t *param_id = array_get_id(id->u_fn.param_ids, i);

        array_add_last(&fn->params, param_id);
        param_id->idx = j++;
    }

    for (i = 0; i < array_size(id->u_fn.ret_ids); i++) {
        ast_id_t *ret_id = array_get_id(id->u_fn.ret_ids, i);

        array_add_last(&fn->params, ret_id);
        ret_id->idx = j++;
    }

    array_init(&fn->bbs);

    fn->entry_bb = bb_new();
    fn->exit_bb = bb_new();

    return fn;
}

void
fn_add_local(ir_fn_t *fn, ast_id_t *id)
{
    /* reserved for two internal variables (e.g, base stack address, relooper) */
    id->idx = array_size(&fn->params) + array_size(&fn->locals) + 2;

    array_add_last(&fn->locals, id);
}

void
fn_add_stack(ir_fn_t *fn, ast_id_t *id)
{
    int i;
    uint32_t size = meta_size(&id->meta);

    id->addr = fn->usage;

    if (is_array_type(&id->meta)) {
        for (i = 0; i < id->meta.arr_dim; i++) {
            ASSERT1(id->meta.arr_size[i] > 0, id->meta.arr_size[i]);
            size *= id->meta.arr_size[i];
        }
    }

    fn->usage += ALIGN64(size);
}

void 
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    array_add_last(&fn->bbs, bb);
}

/* end of ir_fn.c */
