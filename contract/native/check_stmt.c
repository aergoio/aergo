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
    int i;
    ast_exp_t *cond_exp;
    ast_meta_t *cond_meta;
    array_t *elif_stmts;

    ASSERT1(stmt_is_if(stmt), stmt->kind);
    ASSERT(stmt->u_if.cond_exp != NULL);

    cond_exp = stmt->u_if.cond_exp;
    cond_meta = &cond_exp->meta;

    TRY(check_exp(check, cond_exp));

    if (!meta_is_bool(cond_meta))
        THROW(ERROR_INVALID_COND_TYPE, &cond_exp->trc,
              TYPE_NAME(cond_meta->type));

    if (stmt->u_if.if_blk != NULL)
        check_blk(check, stmt->u_if.if_blk);

    elif_stmts = &stmt->u_if.elif_stmts;

    for (i = 0; i < array_size(elif_stmts); i++) {
        check_stmt(check, array_item(elif_stmts, i, ast_stmt_t));
    }

    if (stmt->u_if.else_blk != NULL)
        check_blk(check, stmt->u_if.else_blk);

    return NO_ERROR;
}

static int
stmt_for_check(check_t *check, ast_stmt_t *stmt)
{
    ast_exp_t *cond_exp;

    ASSERT1(stmt_is_for(stmt), stmt->kind);

    if (stmt->u_for.init_exp != NULL)
        check_exp(check, stmt->u_for.init_exp);

    cond_exp = stmt->u_for.cond_exp;

    if (cond_exp != NULL) {
        ast_meta_t *cond_meta = &cond_exp->meta;

        check_exp(check, cond_exp);

        if (exp_is_tuple(cond_exp)) {
            int i;
            array_t *exps = cond_exp->u_tup.exps;

            for (i = 0; i < array_size(exps); i++) {
                ast_exp_t *exp = array_item(exps, i, ast_exp_t);

                if (!meta_is_bool(&exp->meta))
                    THROW(ERROR_INVALID_COND_TYPE, &exp->trc,
                          TYPE_NAME(exp->meta.type));
            }
        }
        else if (!meta_is_bool(cond_meta)) {
            THROW(ERROR_INVALID_COND_TYPE, &cond_exp->trc,
                  TYPE_NAME(cond_meta->type));
        }
    }

    if (stmt->u_for.loop_exp != NULL)
        check_exp(check, stmt->u_for.loop_exp);

    if (stmt->u_for.blk != NULL)
        check_blk(check, stmt->u_for.blk);

    return NO_ERROR;
}

static int
stmt_case_check(check_t *check, ast_stmt_t *stmt, ast_meta_t *meta)
{
    int i;
    ast_exp_t *val_exp;
    array_t *stmts;

    ASSERT1(stmt_is_case(stmt), stmt->kind);

    val_exp = stmt->u_case.val_exp;

    if (val_exp != NULL) {
        ast_meta_t *val_meta = &val_exp->meta;

        check_exp(check, val_exp);

        if (meta == NULL) {
            if (!meta_is_bool(val_meta))
                THROW(ERROR_INVALID_COND_TYPE, &val_exp->trc,
                      TYPE_NAME(val_meta->type));
        }
        else if (!meta_is_compatible(meta, val_meta)) {
            THROW(ERROR_MISMATCHED_TYPE, &val_exp->trc,
                  TYPE_NAME(meta->type), TYPE_NAME(val_meta->type));
        }
    }

    stmts = stmt->u_case.stmts;

    for (i = 0; i < array_size(stmts); i++) {
        check_stmt(check, array_item(stmts, i, ast_stmt_t));
    }

    return NO_ERROR;
}

static int
stmt_switch_check(check_t *check, ast_stmt_t *stmt)
{
    int i;
    ast_exp_t *cond_exp;
    ast_meta_t *cond_meta = NULL;
    array_t *case_stmts;

    ASSERT1(stmt_is_switch(stmt), stmt->kind);

    cond_exp = stmt->u_sw.cond_exp;

    if (cond_exp != NULL) {
        cond_meta = &cond_exp->meta;

        check_exp(check, cond_exp);

        if (!meta_is_comparable(cond_meta))
            THROW(ERROR_NOT_COMPARABLE_TYPE, &cond_exp->trc,
                  TYPE_NAME(cond_meta->type));
    }

    case_stmts = stmt->u_sw.case_stmts;

    for (i = 0; i < array_size(case_stmts); i++) {
        stmt_case_check(check, array_item(case_stmts, i, ast_stmt_t),
                        cond_meta);
    }

    return NO_ERROR;
}

static int
stmt_return_check(check_t *check, ast_stmt_t *stmt)
{
    ast_exp_t *arg_exp;
    ast_id_t *fn_id;
    array_t *fn_ret_exps;

    ASSERT1(stmt_is_return(stmt), stmt->kind);

    arg_exp = stmt->u_ret.arg_exp;

    fn_id = check->fn_id;
    ASSERT(fn_id != NULL);
    ASSERT1(id_is_func(fn_id), fn_id->kind);

    fn_ret_exps = fn_id->u_func.ret_exps;

    if (arg_exp != NULL) {
        if (fn_ret_exps == NULL)
            THROW(ERROR_MISMATCHED_RETURN, &arg_exp->trc);

        check_exp(check, arg_exp);

        if (exp_is_tuple(arg_exp)) {
            int i;
            array_t *ret_exps = arg_exp->u_tup.exps;

            if (array_size(ret_exps) != array_size(fn_ret_exps))
                THROW(ERROR_MISMATCHED_RETURN, &arg_exp->trc);

            for (i = 0; i < array_size(ret_exps); i++) {
                ast_exp_t *ret_exp = array_item(ret_exps, i, ast_exp_t);
                ast_meta_t *ret_meta = &ret_exp->meta;
                ast_exp_t *fn_ret_exp = array_item(fn_ret_exps, i, ast_exp_t);
                ast_meta_t *fn_ret_meta = &fn_ret_exp->meta;

                if ((exp_is_lit(ret_exp) &&
                     !meta_is_compatible(ret_meta, fn_ret_meta)) ||
                    meta_equals(ret_meta, fn_ret_meta))
                    THROW(ERROR_MISMATCHED_TYPE, &arg_exp->trc,
                          TYPE_NAME(fn_ret_meta->type),
                          TYPE_NAME(ret_meta->type));
            }
        }
        else {
            ast_meta_t *fn_ret_meta;
            ast_meta_t *arg_meta = &arg_exp->meta;

            if (array_size(fn_ret_exps) != 1)
                THROW(ERROR_MISMATCHED_RETURN, &arg_exp->trc);

            fn_ret_meta = &array_item(fn_ret_exps, 0, ast_exp_t)->meta;

            if ((exp_is_lit(arg_exp) &&
                 !meta_is_compatible(arg_meta, fn_ret_meta)) ||
                meta_equals(arg_meta, fn_ret_meta))
                THROW(ERROR_MISMATCHED_TYPE, &arg_exp->trc,
                      TYPE_NAME(fn_ret_meta->type), TYPE_NAME(arg_meta->type));
        }
    }
    else {
        if (fn_ret_exps != NULL)
            THROW(ERROR_MISMATCHED_RETURN, &stmt->trc);
    }

    return NO_ERROR;
}

static int
stmt_ddl_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(stmt_is_ddl(stmt), stmt->kind);
    ASSERT(stmt->u_ddl.ddl != NULL);

    return NO_ERROR;
}

static int
stmt_blk_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(stmt_is_blk(stmt), stmt->kind);

    if (stmt->u_blk.blk != NULL)
        check_blk(check, stmt->u_blk.blk);

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
