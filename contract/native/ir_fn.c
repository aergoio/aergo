/**
 * @file    ir_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ir_bb.h"
#include "ir_abi.h"
#include "gen.h"
#include "gen_util.h"

#include "ir_fn.h"

ir_fn_t *
fn_new(ast_id_t *id)
{
    ir_fn_t *fn = xcalloc(sizeof(ir_fn_t));

    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);

    fn->name = id->u_fn.qname;
    fn->apiname = is_public_id(id) ? id->name : NULL;

    fn->abi = abi_new(id);

    array_init(&fn->types, BinaryenType);
    vector_init(&fn->bbs);

    fn->entry_bb = bb_new();
    fn->exit_bb = bb_new();

    fn->entry_bb->ref_cnt++;

    fn->cont_idx = -1;
    fn->heap_idx = -1;
    fn->stack_idx = -1;
    fn->reloop_idx = -1;
    fn->ret_idx = -1;

    return fn;
}

void
fn_add_global(ir_fn_t *fn, meta_t *meta)
{
    ASSERT(fn != NULL);

    fn->heap_usage = ALIGN(fn->heap_usage, meta_align(meta));

    /* Global variables are always accessed with "base_idx + rel_addr", and offset is used only 
     * when accessing an array or struct element */

    meta->base_idx = fn->cont_idx;
    meta->rel_addr = fn->heap_usage;
    meta->rel_offset = 0;

    fn->heap_usage += meta_memsz(meta);
}

uint32_t
fn_add_register(ir_fn_t *fn, meta_t *meta)
{
    uint32_t reg_idx;

    ASSERT(fn != NULL);
    ASSERT(fn->abi != NULL);

    reg_idx = fn->abi->param_cnt + array_size(&fn->types);

    array_add(&fn->types, meta_gen(meta), BinaryenType);

    return reg_idx;
}

void
fn_add_heap(ir_fn_t *fn, uint32_t size, meta_t *meta)
{
    ASSERT(fn != NULL);

    fn->heap_usage = ALIGN(fn->heap_usage, meta_align(meta));

    meta->base_idx = fn->heap_idx;
    meta->rel_addr = fn->heap_usage;
    meta->rel_offset = 0;

    fn->heap_usage += size;
}

void
fn_add_stack(ir_fn_t *fn, uint32_t size, meta_t *meta)
{
    ASSERT(fn != NULL);
    ASSERT(fn->stack_idx >= 0);

    fn->stack_usage = ALIGN(fn->stack_usage, meta_align(meta));

    meta->base_idx = fn->stack_idx;
    meta->rel_addr = fn->stack_usage;
    meta->rel_offset = 0;

    fn->stack_usage += size;
}

void
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    ASSERT(fn != NULL);

    vector_add_last(&fn->bbs, bb);
}

/* end of ir_fn.c */
