/**
 * @file    ir_bb.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_bb.h"

ir_bb_t *
bb_new(void)
{
    ir_bb_t *bb = xmalloc(sizeof(ir_bb_t));

    array_init(&bb->ids);
    array_init(&bb->stmts);
    array_init(&bb->brs);

    bb->rb = NULL;

    return bb;
}

void
bb_add_id(ir_bb_t *bb, ast_id_t *id)
{
    ASSERT(bb != NULL);

    array_add_last(&bb->ids, id);
}

void
bb_add_stmt(ir_bb_t *bb, ast_stmt_t *stmt)
{
    ASSERT(bb != NULL);

    array_add_last(&bb->stmts, stmt);
}

static ir_br_t *
br_new(ast_exp_t *cond_exp, ir_bb_t *bb)
{
    ir_br_t *br = xmalloc(sizeof(ir_br_t));

    br->cond_exp = cond_exp;
    br->bb = bb;

    return br;
}

void
bb_add_branch(ir_bb_t *bb, ast_exp_t *cond_exp, ir_bb_t *br_bb)
{
    ASSERT(bb != NULL);

    array_add_last(&bb->brs, br_new(cond_exp, br_bb));
}

/* end of ir_bb.c */
