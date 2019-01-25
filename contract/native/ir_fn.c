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
    char name[NAME_MAX_LEN * 2 + 2];
    ir_fn_t *fn = xmalloc(sizeof(ir_fn_t));

    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);

    snprintf(name, sizeof(name), "%s$%s", id->up->name, id->name);

    fn->name = xstrdup(name);
    fn->exp_name = is_public_id(id) ? id->name : NULL;

    fn->abi = NULL;

    array_init(&fn->types, BinaryenType);
    vector_init(&fn->bbs);

    fn->entry_bb = bb_new();
    fn->exit_bb = bb_new();

    fn->cont_idx = -1;
    fn->heap_idx = -1;
    fn->stack_idx = -1;
    fn->reloop_idx = -1;
    fn->ret_idx = -1;

    fn->heap_usage = 0;
    fn->stack_usage = 0;

    return fn;
}

void
fn_add_global(ir_fn_t *fn, ast_id_t *id)
{
    meta_t *meta = &id->meta;

    ASSERT(fn != NULL);
    ASSERT1(is_var_id(id), id->kind);

    if (is_array_meta(meta))
        /* The array is always accessed as a reference */
        fn->heap_usage = ALIGN32(fn->heap_usage);
    else
        fn->heap_usage = ALIGN(fn->heap_usage, meta_align(meta));

    /* Heap variables are always accessed with "base_idx + rel_addr", and offset is
     * used only when accessing an vector or struct element */

    meta->base_idx = fn->cont_idx;
    meta->rel_addr = fn->heap_usage;
    meta->rel_offset = 0;

    if (is_array_meta(meta))
        fn->heap_usage += sizeof(uint32_t);
    else
        fn->heap_usage += TYPE_BYTE(meta->type);
}

void
fn_add_register(ir_fn_t *fn, ast_id_t *id)
{
    ASSERT(fn != NULL);
    ASSERT(fn->abi != NULL);
    ASSERT1(is_var_id(id), id->kind);

    id->idx = fn->abi->param_cnt + array_size(&fn->types);

    array_add(&fn->types, meta_gen(&id->meta), BinaryenType);
}

void
fn_add_heap(ir_fn_t *fn, meta_t *meta)
{
    ASSERT(fn != NULL);

    fn->heap_usage = ALIGN(fn->heap_usage, meta_align(meta));

    /* Heap variables are always accessed with "base_idx + rel_addr", and offset is
     * used only when accessing an vector or struct element */

    meta->base_idx = fn->heap_idx;
    meta->rel_addr = fn->heap_usage;
    meta->rel_offset = 0;

    fn->heap_usage += meta_size(meta);
}

void
fn_add_stack(ir_fn_t *fn, meta_t *meta)
{
    ASSERT(fn != NULL);
    ASSERT(fn->stack_idx >= 0);

    fn->stack_usage = ALIGN(fn->stack_usage, meta_align(meta));

    meta->base_idx = fn->stack_idx;
    meta->rel_addr = fn->stack_usage;

    fn->stack_usage += meta_size(meta);
}

void
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    ASSERT(fn != NULL);

    vector_add_last(&fn->bbs, bb);
}

int
fn_add_tmp_var(ir_fn_t *fn, char *name, type_t type)
{
    ast_id_t *tmp_id;

    ASSERT(fn != NULL);
    ASSERT(fn->abi != NULL);

    tmp_id = id_new_tmp_var(name);
    meta_set(&tmp_id->meta, type);

    tmp_id->idx = fn->abi->param_cnt + array_size(&fn->types);

    array_add(&fn->types, type_gen(type), BinaryenType);

    return tmp_id->idx;
}

/* end of ir_fn.c */
