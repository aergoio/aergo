/**
 * @file    ast_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_stmt.h"

ast_stmt_t *
ast_stmt_new(stmt_type_t type, yylloc_t *lloc)
{
    ast_stmt_t *stmt = xmalloc(sizeof(ast_stmt_t));

    stmt->type = type;
    stmt->lloc = *lloc;
    list_link_init(&stmt->link);

    return stmt;
}

ast_stmt_t *
stmt_exp_new(ast_exp_t *exp, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_EXP, lloc);

    stmt->u_exp.exp = exp;

    return stmt;
}

ast_stmt_t *
stmt_if_new(ast_exp_t *cmp_exp, ast_stmt_t *if_blk, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_IF, lloc);

    stmt->u_if.cmp_exp = cmp_exp;
    stmt->u_if.if_blk = if_blk;
    stmt->u_if.else_blk = NULL;
    list_init(&stmt->u_if.elsif_l);

    return stmt;
}

ast_stmt_t *
stmt_for_new(ast_exp_t *init_exp, ast_exp_t *check_exp, ast_exp_t *inc_exp, 
             ast_stmt_t *blk, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_FOR, lloc);

    stmt->u_for.init_l = NULL;
    stmt->u_for.init_exp = init_exp;
    stmt->u_for.check_exp = check_exp;
    stmt->u_for.inc_exp = inc_exp;
    stmt->u_for.blk = blk;

    return stmt;
}

ast_stmt_t *
stmt_switch_new(ast_exp_t *cmp_exp, list_t *case_l, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_SWITCH, lloc);

    stmt->u_sw.cmp_exp = cmp_exp;
    stmt->u_sw.case_l = case_l;

    return stmt;
}

ast_stmt_t *
stmt_case_new(ast_exp_t *cmp_exp, list_t *stmt_l, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_CASE, lloc);

    stmt->u_case.cmp_exp = cmp_exp;
    stmt->u_case.stmt_l = stmt_l;

    return stmt;
}

ast_stmt_t *
stmt_return_new(ast_exp_t *exp, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_RETURN, lloc);

    stmt->u_ret.exp = exp;

    return stmt;
}

ast_stmt_t *
stmt_ddl_new(ddl_kind_t kind, char *ddl, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_DDL, lloc);

    stmt->u_ddl.kind = kind;
    stmt->u_ddl.ddl = ddl;

    return stmt;
}

ast_stmt_t *
stmt_blk_new(ast_blk_t *blk, yylloc_t *lloc)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_BLK, lloc);

    stmt->u_blk.blk = blk;

    return stmt;
}

/* end of ast_stmt.c */
