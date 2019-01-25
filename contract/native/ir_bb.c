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
    ir_bb_t *bb = xmalloc(sizeof(ir_bb_t));

    vector_init(&bb->stmts);
    vector_init(&bb->brs);
    vector_init(&bb->pgbacks);

    bb->rb = NULL;

    return bb;
}

void
bb_add_stmt(ir_bb_t *bb, ast_stmt_t *stmt)
{
    ASSERT(bb != NULL);

    vector_add_last(&bb->stmts, stmt);

    if (has_piggyback(bb)) {
        vector_join_last(&bb->stmts, &bb->pgbacks);
        vector_reset(&bb->pgbacks);
    }
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
}

void
bb_add_piggyback(ir_bb_t *bb, ast_stmt_t *stmt)
{
    ASSERT(stmt != NULL);

    vector_add_last(&bb->pgbacks, stmt);
}

/* end of ir_bb.c */
