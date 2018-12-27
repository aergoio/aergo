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

    if (id->u_fn.ret_id != NULL) {
        ast_id_t *ret_id = id->u_fn.ret_id;

        if (is_tuple_id(ret_id)) {
            for (i = 0; i < array_size(&ret_id->u_tup.var_ids); i++) {
                ast_id_t *var_id = array_get_id(&ret_id->u_tup.var_ids, i);

                array_add_last(&fn->params, var_id);
                var_id->idx = j++;
            }
        }
        else {
            array_add_last(&fn->params, ret_id);
            ret_id->idx = j++;
        }
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
    id->addr = fn->usage;

    fn->usage += meta_size(&id->meta);
}

void 
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    array_add_last(&fn->bbs, bb);
}

/* end of ir_fn.c */
