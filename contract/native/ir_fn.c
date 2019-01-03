/**
 * @file    ir_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_bb.h"
#include "gen_util.h"

#include "ir_fn.h"

ir_fn_t *
fn_new(ast_id_t *id)
{
    int i, j = 0;
    ast_id_t *ret_id = id->u_fn.ret_id;
    ir_fn_t *fn = xmalloc(sizeof(ir_fn_t));

    ASSERT(id->u_fn.cont_id != NULL);

    fn->name = id->u_fn.qname;

    fn->param_cnt = array_size(id->u_fn.param_ids);

    if (ret_id != NULL) {
        if (is_tuple_id(ret_id))
            fn->param_cnt += array_size(&ret_id->u_tup.elem_ids);
        else
            fn->param_cnt++;
    }

    fn->params = xmalloc(sizeof(BinaryenType) * fn->param_cnt);

    array_foreach(id->u_fn.param_ids, i) {
        ast_id_t *param_id = array_get_id(id->u_fn.param_ids, i);

        fn->params[j] = meta_gen(&param_id->meta);
        param_id->idx = j++;
    }

    /* The return value is always passed as an address */
    if (ret_id != NULL) {
        if (is_tuple_id(ret_id)) {
            array_foreach(&ret_id->u_tup.elem_ids, i) {
                ast_id_t *elem_id = array_get_id(&ret_id->u_tup.elem_ids, i);

                fn->params[j] = BinaryenTypeInt32();
                elem_id->idx = j++;
            }
        }
        else {
            fn->params[j] = BinaryenTypeInt32();
            ret_id->idx = j;
        }
    }

    array_init(&fn->locals);
    array_init(&fn->bbs);

    fn->entry_bb = bb_new();
    fn->exit_bb = bb_new();

    return fn;
}

void
fn_add_local(ir_fn_t *fn, ast_id_t *id)
{
    ASSERT1(is_var_id(id) || is_return_id(id), id->kind);

    /* reserved for two internal variables (e.g, base stack address, relooper) */
    id->idx = fn->param_cnt + array_size(&fn->locals) + 2;

    array_add_last(&fn->locals, id);
}

void
fn_add_stack(ir_fn_t *fn, ast_id_t *id)
{
    ASSERT1(is_var_id(id) || is_return_id(id), id->kind);

    id->addr = fn->usage;

    fn->usage += meta_size(&id->meta);
}

void 
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    array_add_last(&fn->bbs, bb);
}

/* end of ir_fn.c */
