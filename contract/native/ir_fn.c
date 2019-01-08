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
fn_new(ast_id_t *id, ir_abi_t *abi)
{
    char name[NAME_MAX_LEN * 2 + 2];
    ir_fn_t *fn = xmalloc(sizeof(ir_fn_t));

    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);
    ASSERT(abi != NULL);

    snprintf(name, sizeof(name), "%s$%s", id->up->name, id->name);

    fn->name = xstrdup(name);
    fn->exp_name = is_public_id(id) ? id->name : NULL;

    fn->abi = abi;

    array_init(&fn->locals);
    array_init(&fn->bbs);

    fn->entry_bb = bb_new();
    fn->exit_bb = bb_new();

    fn->usage = 0;

    return fn;
}

void
fn_add_local(ir_fn_t *fn, ast_id_t *id)
{
    ir_abi_t *abi = fn->abi;

    ASSERT1(is_var_id(id) || is_return_id(id), id->kind);
    ASSERT(abi != NULL);

    /* reserved for two internal variables (e.g, base stack address, relooper) */
    id->idx = abi->param_cnt + array_size(&fn->locals) + 2;

    array_add_last(&fn->locals, id);
}

void
fn_add_stack(ir_fn_t *fn, ast_id_t *id)
{
    ASSERT1(is_var_id(id) || is_return_id(id), id->kind);

    fn->usage = ALIGN(fn->usage, meta_align(&id->meta));

    id->addr = fn->usage;

    fn->usage += meta_size(&id->meta);
}

void
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    array_add_last(&fn->bbs, bb);
}

/* end of ir_fn.c */
