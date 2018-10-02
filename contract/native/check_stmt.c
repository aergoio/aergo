/**
 * @file    check_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"
#include "check_exp.h"
#include "check_blk.h"

#include "check_stmt.h"

static int
stmt_if_check(check_t *check, ast_stmt_t *stmt)
{
    array_t *elif_stmts;

    ASSERT1(stmt_is_if(stmt), stmt->kind);
    ASSERT(stmt->u_if.cmp_exp != NULL);

    TRY(check_exp(check, stmt->u_if.cmp_exp));

    if (stmt->u_if.if_blk != NULL)
        check_blk(check, stmt->u_if.if_blk);

    elif_stmts = &stmt->u_if.elif_stmts;

    if (!array_empty(elif_stmts)) {
        int i;

        for (i = 0; i < array_size(elif_stmts); i++) {
            check_stmt(check, array_item(elif_stmts, i, ast_stmt_t));
        }
    }

    if (stmt->u_if.else_blk != NULL) 
        check_blk(check, stmt->u_if.else_blk);

    return NO_ERROR;
}

static int
stmt_for_check(check_t *check, ast_stmt_t *stmt)
{
    array_t *init_ids;

    ASSERT1(stmt_is_for(stmt), stmt->kind);

    init_ids = stmt->u_for.init_ids;

    if (!array_empty(init_ids)) {
        int i;

        for (i = 0; i < array_size(init_ids); i++) {
            check_id(check, array_item(init_ids, i, ast_id_t));
        }
    }

    if (stmt->u_for.init_exp != NULL)
        check_exp(check, stmt->u_for.init_exp);

    if (stmt->u_for.check_exp != NULL)
        check_exp(check, stmt->u_for.check_exp);

    if (stmt->u_for.inc_exp != NULL)
        check_exp(check, stmt->u_for.inc_exp);

    if (stmt->u_for.blk != NULL)
        check_blk(check, stmt->u_for.blk);

    return NO_ERROR;
}

static int
stmt_case_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(stmt_is_case(stmt), stmt->kind);

    return NO_ERROR;
}

static int
stmt_switch_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(stmt_is_switch(stmt), stmt->kind);

    return NO_ERROR;
}

static int
stmt_return_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(stmt_is_return(stmt), stmt->kind);

    return NO_ERROR;
}

static int
stmt_ddl_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(stmt_is_ddl(stmt), stmt->kind);

    return NO_ERROR;
}

static int
stmt_blk_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(stmt_is_blk(stmt), stmt->kind);

    return NO_ERROR;
}

int
check_stmt(check_t *check, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_NULL:
    case STMT_CONTINUE:
    case STMT_BREAK:
        return NO_ERROR;

    case STMT_EXP:
        return check_exp(check, stmt->u_exp.exp);

    case STMT_IF:
        return stmt_if_check(check, stmt);

    case STMT_FOR:
        return stmt_for_check(check, stmt);

    case STMT_SWITCH:
        return stmt_switch_check(check, stmt);

    case STMT_RETURN:
        return stmt_return_check(check, stmt);

    case STMT_DDL:
        return stmt_ddl_check(check, stmt);

    case STMT_BLK:
        return stmt_blk_check(check, stmt);

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }

    return NO_ERROR;
}

/* end of check_stmt.c */
