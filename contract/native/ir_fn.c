/**
 * @file    ir_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ir_bb.h"
#include "ir_abi.h"
#include "gen.h"

#include "ir_fn.h"

ir_fn_t *
fn_new(ast_id_t *id)
{
    char name[NAME_MAX_LEN * 2 + 2];
    ir_fn_t *fn = xmalloc(sizeof(ir_fn_t));

    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);

    snprintf(name, sizeof(name), "%s$%s", id->up->name, id->name);

    fn->name = xstrdup(name);
    fn->exp_name = is_public_id(id) ? id->name : NULL;

    fn->abi = NULL;

    array_init(&fn->locals);
    array_init(&fn->bbs);

    fn->entry_bb = bb_new();
    fn->exit_bb = bb_new();

    fn->heap_idx = -1;
    fn->stack_idx = -1;
    fn->reloop_idx = -1;

    fn->usage = 0;

    return fn;
}

void
fn_add_local(ir_fn_t *fn, ast_id_t *id)
{
    ir_abi_t *abi = fn->abi;

    //ASSERT1(is_var_id(id) || is_return_id(id), id->kind);
    ASSERT1(is_var_id(id), id->kind);
    ASSERT(abi != NULL);

    id->idx = abi->param_cnt + array_size(&fn->locals);

    array_add_last(&fn->locals, id);
}

void
fn_add_stack(ir_fn_t *fn, meta_t *meta)
{
    ASSERT(fn->stack_idx >= 0);

    fn->usage = ALIGN(fn->usage, meta_align(meta));

    meta->base_idx = fn->stack_idx;
    meta->rel_addr = fn->usage;

    fn->usage += meta_size(meta);
}

void
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    array_add_last(&fn->bbs, bb);
}

int
fn_add_tmp_var(ir_fn_t *fn, char *name, type_t type)
{
    ast_id_t *tmp_id;
    ir_abi_t *abi = fn->abi;

    ASSERT(abi != NULL);

    tmp_id = id_new_tmp_var(name);
    meta_set(&tmp_id->meta, type);

    array_add_last(&fn->locals, tmp_id);

    return abi->param_cnt + array_size(&fn->locals) - 1;
}

/* end of ir_fn.c */
