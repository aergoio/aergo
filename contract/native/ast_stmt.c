/**
 * @file    ast_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_stmt.h"

ast_stmt_t *
ast_stmt_new(stmt_kind_t kind, errpos_t *pos)
{
    ast_stmt_t *stmt = xmalloc(sizeof(ast_stmt_t));

    ast_node_init(stmt, pos);

    stmt->kind = kind;

    return stmt;
}

ast_stmt_t *
stmt_exp_new(ast_exp_t *exp, errpos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_EXP, pos);

    stmt->u_exp.exp = exp;

    return stmt;
}

ast_stmt_t *
stmt_if_new(ast_exp_t *cond_exp, ast_blk_t *if_blk, errpos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_IF, pos);

    stmt->u_if.cond_exp = cond_exp;
    stmt->u_if.if_blk = if_blk;
    stmt->u_if.else_blk = NULL;
    array_init(&stmt->u_if.elif_stmts);

    return stmt;
}

ast_stmt_t *
stmt_for_new(ast_exp_t *init_exp, ast_exp_t *cond_exp, ast_exp_t *loop_exp, 
             ast_blk_t *blk, errpos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_FOR, pos);

    stmt->u_for.init_exp = init_exp;
    stmt->u_for.cond_exp = cond_exp;
    stmt->u_for.loop_exp = loop_exp;
    stmt->u_for.blk = blk;

    return stmt;
}

ast_stmt_t *
stmt_switch_new(ast_exp_t *cond_exp, array_t *case_stmts, errpos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_SWITCH, pos);

    stmt->u_sw.cond_exp = cond_exp;
    stmt->u_sw.case_stmts = case_stmts;

    return stmt;
}

ast_stmt_t *
stmt_case_new(ast_exp_t *val_exp, array_t *stmts, errpos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_CASE, pos);

    stmt->u_case.val_exp = val_exp;
    stmt->u_case.stmts = stmts;

    return stmt;
}

ast_stmt_t *
stmt_return_new(ast_exp_t *arg_exp, errpos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_RETURN, pos);

    stmt->u_ret.arg_exp = arg_exp;

    return stmt;
}

ast_stmt_t *
stmt_ddl_new(ddl_kind_t kind, char *ddl, errpos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_DDL, pos);

    stmt->u_ddl.kind = kind;
    stmt->u_ddl.ddl = ddl;

    return stmt;
}

ast_stmt_t *
stmt_blk_new(ast_blk_t *blk, errpos_t *pos)
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
