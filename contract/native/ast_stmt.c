/**
 * @file    ast_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"

#include "ast_stmt.h"

static ast_stmt_t *
ast_stmt_new(stmt_kind_t kind, src_pos_t *pos)
{
    ast_stmt_t *stmt = xcalloc(sizeof(ast_stmt_t));

    ast_node_init(stmt, pos);

    stmt->kind = kind;

    return stmt;
}

ast_stmt_t *
stmt_new_null(src_pos_t *pos)
{
    return ast_stmt_new(STMT_NULL, pos);
}

ast_stmt_t *
stmt_new_exp(ast_exp_t *exp, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_EXP, pos);

    stmt->u_exp.exp = exp;

    return stmt;
}

ast_stmt_t *
stmt_new_if(ast_exp_t *cond_exp, ast_blk_t *if_blk, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_IF, pos);

    stmt->u_if.cond_exp = cond_exp;
    stmt->u_if.if_blk = if_blk;
    stmt->u_if.else_blk = NULL;
    array_init(&stmt->u_if.elif_stmts);

    return stmt;
}

ast_stmt_t *
stmt_new_loop(loop_kind_t kind, ast_exp_t *cond_exp, ast_exp_t *loop_exp,
              ast_blk_t *blk, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_LOOP, pos);

    stmt->u_loop.kind = kind;
    stmt->u_loop.cond_exp = cond_exp;
    stmt->u_loop.loop_exp = loop_exp;
    stmt->u_loop.blk = blk;

    if (stmt->u_loop.blk != NULL)
        stmt->u_loop.blk->kind = BLK_LOOP;

    return stmt;
}

ast_stmt_t *
stmt_new_switch(ast_exp_t *cond_exp, array_t *case_stmts, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_SWITCH, pos);

    stmt->u_sw.cond_exp = cond_exp;
    stmt->u_sw.case_stmts = case_stmts;

    return stmt;
}

ast_stmt_t *
stmt_new_case(ast_exp_t *val_exp, array_t *stmts, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_CASE, pos);

    stmt->u_case.val_exp = val_exp;
    stmt->u_case.stmts = stmts;

    return stmt;
}

ast_stmt_t *
stmt_new_return(ast_exp_t *arg_exp, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_RETURN, pos);

    stmt->u_ret.arg_exp = arg_exp;

    return stmt;
}

ast_stmt_t *
stmt_new_goto(char *label, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_GOTO, pos);

    stmt->u_goto.label = label;

    return stmt;
}

ast_stmt_t *
stmt_new_jump(stmt_kind_t kind, src_pos_t *pos)
{
    return ast_stmt_new(kind, pos);
}

ast_stmt_t *
stmt_new_ddl(char *ddl, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_DDL, pos);

    stmt->u_ddl.ddl = ddl;

    return stmt;
}

ast_stmt_t *
stmt_new_blk(ast_blk_t *blk, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_BLK, pos);

    stmt->u_blk.blk = blk;

    return stmt;
}

void
ast_stmt_dump(ast_stmt_t *stmt, int indent)
{
}

/* end of ast_stmt.c */
