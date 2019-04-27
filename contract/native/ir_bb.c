/**
 * @file    ir_bb.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_exp.h"
#include "ast_stmt.h"

#include "ir_bb.h"

ir_bb_t *
bb_new(void)
{
    ir_bb_t *bb = xcalloc(sizeof(ir_bb_t));

    vector_init(&bb->stmts);
    vector_init(&bb->brs);

    return bb;
}

void
bb_add_stmt(ir_bb_t *bb, ast_stmt_t *stmt)
{
    ASSERT(bb != NULL);

    vector_add_last(&bb->stmts, stmt);
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

    vector_add_last(&bb->brs, br_new(cond_exp, br_bb));

    br_bb->ref_cnt++;
}

/* end of ir_bb.c */
