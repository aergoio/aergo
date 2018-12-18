/**
 * @file    trans_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "trans_id.h"
#include "trans_exp.h"
#include "trans_blk.h"

#include "trans_stmt.h"

static void
stmt_trans_assign(trans_t *trans, ast_stmt_t *stmt)
{
    if (is_tuple_exp(l_exp)) {
        ERROR(ERROR_NOT_SUPPORTED, &stmt->pos);
        return;
    }

    bb_add_instr(trans->bb, stmt);
}

static void
stmt_trans_if(trans_t *trans, ast_stmt_t *stmt)
{
    int i;
    ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *next_bb = bb_new();

    trans->bb = bb_new();

    if (stmt->u_if.if_blk != NULL)
        blk_trans(trans, stmt->u_if.if_blk);

    bb_add_branch(prev_bb, stmt->u_if.cond_exp, trans->bb);
    bb_add_branch(trans->bb, NULL, next_bb);

    for (i = 0; i < array_size(elif_stmts); i++) {
        ast_stmt_t *elif_stmt = array_get(elif_stmts, i, ast_stmt_t);

        trans->bb = bb_new();

        if (elif_stmt->u_if.if_blk != NULL)
            blk_trans(trans, elif_stmt->u_if.if_blk);

        bb_add_branch(prev_bb, elif_stmt->u_if.cond_exp, trans->bb);
        bb_add_branch(trans->bb, NULL, next_bb);
    }

    /* XXX: Fix parsing & checking mechanism */

    if (stmt->u_if.else_blk != NULL) {
        trans->bb = bb_new();
        blk_trans(trans, stmt->u_if.else_blk);

        bb_add_branch(prev_bb, NULL, trans->bb);
        bb_add_branch(trans->bb, NULL, next_bb);
    }

    fn_add_basic_blk(trans->fn, prev_bb);

    trans->bb = next_bb;
}

static void
stmt_trans_for_loop(trans_t *trans, ast_stmt_t *stmt)
{
    ast_exp_t *cond_exp = stmt->u_loop.cond_exp;
    ir_bb_t *cond_bb = bb_new();
    ir_bb_t *loop_bb = bb_new();
    ir_bb_t *next_bb = bb_new();

    /* XXX: Fix parsing & checking mechanism */

    if (stmt->u_loop.init_stmt != NULL)
        stmt_trans(trans, stmt->u_loop.init_stmt);

    /* previous basic block */
    bb_add_branch(trans->bb, NULL, cond_bb);
    fn_add_basic_blk(trans->fn, trans->bb);

    trans->bb = cond_bb;

    if (cond_exp != NULL) {
        bb_add_branch(cond_bb, cond_exp, loop_bb);
        bb_add_branch(cond_bb, NULL, next_bb);
    }

    if (stmt->u_loop.blk != NULL) {
        trans->bb = loop_bb;
        blk_trans(trans, stmt->u_loop.blk);
    }

    /* make loop */
    bb_add_branch(trans->bb, NULL, cond_bb);

    fn_add_basic_blk(trans->fn, trans->bb);

    if (cond_bb != trans->bb)
        fn_add_basic_blk(trans->fn, cond_bb);

    trans->bb = next_bb;
}

static void
stmt_trans_array_loop(trans_t *trans, ast_stmt_t *stmt)
{
    ERROR(ERROR_NOT_SUPPORTED, &stmt->pos);
}

static void
stmt_trans_loop(trans_t *trans, ast_stmt_t *stmt)
{
    switch (stmt->u_loop.kind) {
    case LOOP_FOR:
        stmt_trans_for_loop(trans, stmt);
        break;

    case LOOP_ARRAY:
        stmt_trans_array_loop(trans, stmt);
        break;

    default:
        ASSERT1(!"invalid loop", stmt->u_loop.kind);
    }
}

static void
stmt_trans_switch(trans_t *trans, ast_stmt_t *stmt)
{
    int i, j;
    ast_blk_t *blk = stmt->u_sw.blk;
    ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *next_bb = bb_new();

    fn_add_basic_blk(trans->fn, prev_bb);

    for (i = 0; i < array_size(&blk->stmts); i++) {
        ast_stmt_t *case_stmt = array_get(&blk->stmts, i, ast_stmt_t);

        trans->bb = bb_new();
        bb_add_branch(prev_bb, case_stmt->u_case.val_exp, trans->bb);

        for (j = 0; j < array_size(case_stmt->u_case.stmts); j++) {
            stmt_trans(trans, array_get(case_stmt->u_case.stmts, j, ast_stmt_t));
        }

        bb_add_branch(trans->bb, NULL, next_bb);
        fn_add_basic_blk(trans->fn, trans->bb);
    }

    trans->bb = next_bb;
}

static void
stmt_trans_return(trans_t *trans, ast_stmt_t *stmt)
{
    bb_add_instr(trans->bb, stmt);
}

static void
stmt_trans_continue(trans_t *trans, ast_stmt_t *stmt)
{
    /* XXX: branch to cond_bb */
    ast_blk_t *blk;

    ASSERT1(is_continue_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_jump.cond_exp == NULL);

    blk = blk_search(trans->blk, BLK_LOOP);
    if (blk == NULL)
        RETURN(ERROR_INVALID_JUMP_STMT, &stmt->pos, STMT_KIND(stmt));

    stmt->u_jump.label = blk->name;

    return NO_ERROR;
}

static void
stmt_trans_break(trans_t *trans, ast_stmt_t *stmt)
{
    /* XXX: branch to next_bb */
    ast_exp_t *cond_exp;
    ast_blk_t *blk;

    ASSERT1(is_break_stmt(stmt), stmt->kind);

    cond_exp = stmt->u_jump.cond_exp;

    if (cond_exp != NULL) {
        meta_t *cond_meta = &cond_exp->meta;

        CHECK(exp_trans(trans, cond_exp));

        if (!is_bool_type(cond_meta))
            RETURN(ERROR_INVALID_COND_TYPE, &cond_exp->pos, meta_to_str(cond_meta));
    }

    blk = blk_search(trans->blk, BLK_LOOP);
    if (blk == NULL) {
        blk = blk_search(trans->blk, BLK_SWITCH);
        if (blk == NULL)
            RETURN(ERROR_INVALID_JUMP_STMT, &stmt->pos, STMT_KIND(stmt));
    }

    stmt->u_jump.label = blk->name;

    return NO_ERROR;
}

static void
stmt_trans_goto(trans_t *trans, ast_stmt_t *stmt)
{
    /* XXX: branch to label_bb */
    ast_id_t *label_id;

    ASSERT1(is_goto_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_goto.label != NULL);

    label_id = id_search_label(trans->blk, stmt->u_goto.label);
    if (label_id == NULL)
        RETURN(ERROR_UNDEFINED_LABEL, &stmt->pos, stmt->u_goto.label);

    return NO_ERROR;
}

static void
stmt_trans_ddl(trans_t *trans, ast_stmt_t *stmt)
{
    bb_add_instr(trans->bb, stmt);
}

static void
stmt_trans_blk(trans_t *trans, ast_stmt_t *stmt)
{
    if (stmt->u_blk.blk != NULL)
        blk_trans(trans, stmt->u_blk.blk);
}

void
stmt_trans(trans_t *trans, ast_stmt_t *stmt)
{
    if (stmt->label_id != NULL) {
        ir_bb_t *next_bb = bb_new();

        if (trans->bb != NULL) {
            bb_add_branch(trans->bb, NULL, next_bb);
            fn_add_basic_blk(trans->fn, trans->bb);
        }

        trans->bb = next_bb;
    }

    ASSERT(trans->bb != NULL);

    switch (stmt->kind) {
    case STMT_NULL:
        break;

    case STMT_EXP:
        bb_add_instr(trans->bb, stmt);
        break;

    case STMT_ASSIGN:
        stmt_trans_assign(trans, stmt);
        break;

    case STMT_IF:
        stmt_trans_if(trans, stmt);
        break;

    case STMT_LOOP:
        stmt_trans_loop(trans, stmt);
        break;

    case STMT_SWITCH:
        stmt_trans_switch(trans, stmt);
        break;

    case STMT_CASE:
        stmt_trans_case(trans, stmt);
        break;

    case STMT_RETURN:
        stmt_trans_return(trans, stmt);
        break;

    case STMT_CONTINUE:
        stmt_trans_continue(trans, stmt);
        break;

    case STMT_BREAK:
        stmt_trans_break(trans, stmt);
        break;

    case STMT_GOTO:
        stmt_trans_goto(trans, stmt);
        break;

    case STMT_DDL:
        stmt_trans_ddl(trans, stmt);
        break;

    case STMT_BLK:
        stmt_trans_blk(trans, stmt);
        break;

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }
}

/* end of trans_stmt.c */
