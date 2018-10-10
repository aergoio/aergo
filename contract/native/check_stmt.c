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
    meta_t *cond_meta;
    array_t *elif_stmts;

    ASSERT1(is_if_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_if.cond_exp != NULL);

    cond_exp = stmt->u_if.cond_exp;
    cond_meta = &cond_exp->meta;

    TRY(check_exp(check, cond_exp));

    if (!is_bool_meta(cond_meta))
        THROW(ERROR_INVALID_COND_TYPE, exp_pos(cond_exp),
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
    char begin_label[128];
    char end_label[128];
    ast_exp_t *cond_exp;
    ast_exp_t *loop_exp;
    ast_stmt_t *null_stmt;
    ast_blk_t *blk;

    ASSERT1(is_for_stmt(stmt), stmt->kind);

    if (stmt->u_for.blk == NULL)
        stmt->u_for.blk = ast_blk_new(stmt_pos(stmt));

    blk = stmt->u_for.blk;

    snprintf(begin_label, sizeof(begin_label), "begin_for_loop_%d", stmt->num);
    snprintf(end_label, sizeof(end_label), "end_for_loop_%d", stmt->num);

    if (stmt->u_for.init_vars != NULL) {
        ASSERT(stmt->u_for.init_exp == NULL);
        array_join(&blk->ids, stmt->u_for.init_vars);
    }
    else {
        ast_exp_t *init_exp = stmt->u_for.init_exp;

        if (init_exp != NULL) {
            ast_stmt_t *exp_stmt = stmt_exp_new(init_exp, exp_pos(init_exp));
            array_add_head(&blk->stmts, exp_stmt);
        }
    }

    cond_exp = stmt->u_for.cond_exp;

    if (cond_exp != NULL) {
        ast_blk_t *if_blk;
        ast_stmt_t *goto_stmt;
        ast_exp_t *not_exp;
        ast_stmt_t *if_stmt;

        goto_stmt = stmt_goto_new(xstrdup(end_label), exp_pos(cond_exp));

        if_blk = ast_blk_new(exp_pos(cond_exp));
        array_add_tail(&if_blk->stmts, goto_stmt);

        not_exp = exp_op_new(OP_NOT, cond_exp, NULL, exp_pos(cond_exp));

        if_stmt = stmt_if_new(not_exp, if_blk, exp_pos(cond_exp));
        if_stmt->label = xstrdup(begin_label);

        array_add_head(&blk->stmts, if_stmt);
    }

    loop_exp = stmt->u_for.loop_exp;

    if (loop_exp != NULL) {
        ast_stmt_t *exp_stmt = stmt_exp_new(loop_exp, exp_pos(loop_exp));
        array_add_tail(&blk->stmts, exp_stmt);
    }

    null_stmt = stmt_null_new(stmt_pos(stmt));
    null_stmt->label = xstrdup(end_label);

    array_add_tail(&blk->stmts, null_stmt);

    check_blk(check, blk);

    return NO_ERROR;
}

static int
stmt_case_check(check_t *check, ast_stmt_t *stmt, meta_t *meta)
{
    int i;
    ast_exp_t *val_exp;
    array_t *stmts;

    ASSERT1(is_case_stmt(stmt), stmt->kind);

    val_exp = stmt->u_case.val_exp;

    if (val_exp != NULL) {
        meta_t *val_meta = &val_exp->meta;

        check_exp(check, val_exp);

        if (meta == NULL) {
            if (!is_bool_meta(val_meta))
                THROW(ERROR_INVALID_COND_TYPE, exp_pos(val_exp),
                      TYPE_NAME(val_meta->type));
        }
        else if (!meta_equals(meta, val_meta)) {
            THROW(ERROR_MISMATCHED_TYPE, exp_pos(val_exp),
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
    meta_t *cond_meta = NULL;
    array_t *case_stmts;

    ASSERT1(is_switch_stmt(stmt), stmt->kind);

    cond_exp = stmt->u_sw.cond_exp;

    if (cond_exp != NULL) {
        cond_meta = &cond_exp->meta;

        check_exp(check, cond_exp);

        if (!is_comparable_meta(cond_meta))
            THROW(ERROR_NOT_COMPARABLE_TYPE, exp_pos(cond_exp),
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
    ast_id_t *fn_id;
    meta_t *fn_meta;
    ast_exp_t *arg_exp;

    ASSERT1(is_return_stmt(stmt), stmt->kind);
    ASSERT(check->fn_id != NULL);

    fn_id = check->fn_id;
    fn_meta = &fn_id->meta;

    ASSERT1(is_func_id(fn_id), fn_id->kind);

    arg_exp = stmt->u_ret.arg_exp;

    if (arg_exp != NULL) {
        ASSERT1(is_tuple_meta(fn_meta), fn_meta->type);

        if (is_void_meta(fn_meta))
            THROW(ERROR_MISMATCHED_COUNT, exp_pos(arg_exp), 0,
                  meta_size(&arg_exp->meta));

        check_exp(check, arg_exp);

        if (is_tuple_meta(&arg_exp->meta)) {
            int i;
            array_t *arg_metas = arg_exp->meta.u_tup.metas;
            array_t *ret_metas = fn_meta->u_tup.metas;

            if (array_size(arg_metas) != array_size(ret_metas))
                THROW(ERROR_MISMATCHED_COUNT, exp_pos(arg_exp),
                      array_size(ret_metas), array_size(arg_metas));

            for (i = 0; i < array_size(arg_metas); i++) {
                meta_t *arg_meta = array_item(arg_metas, i, meta_t);
                meta_t *ret_meta = array_item(ret_metas, i, meta_t);

                if (!meta_equals(ret_meta, arg_meta))
                    THROW(ERROR_MISMATCHED_TYPE, exp_pos(arg_exp),
                          TYPE_NAME(ret_meta->type),
                          TYPE_NAME(arg_meta->type));
            }
        }
        else {
            meta_t *arg_meta = &arg_exp->meta;
            array_t *ret_metas = fn_meta->u_tup.metas;
            meta_t *ret_meta;

            if (array_size(ret_metas) != 1)
                THROW(ERROR_MISMATCHED_COUNT, exp_pos(arg_exp),
                      array_size(ret_metas), 1);

            ret_meta = array_item(fn_meta->u_tup.metas, 0, meta_t);

            if (!meta_equals(arg_meta, ret_meta))
                THROW(ERROR_MISMATCHED_TYPE, exp_pos(arg_exp),
                      TYPE_NAME(ret_meta->type), TYPE_NAME(arg_meta->type));
        }
    }
    else if (!is_void_meta(fn_meta)) {
        THROW(ERROR_MISMATCHED_COUNT, stmt_pos(stmt), meta_size(fn_meta), 0);
    }

    return NO_ERROR;
}

static int
stmt_goto_check(check_t *check, ast_stmt_t *stmt)
{
    int i;
    int stmt_cnt;
    bool has_found = false;
    ast_blk_t *blk = check->blk;

    ASSERT1(is_goto_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_goto.label != NULL);
    ASSERT(blk != NULL);

    do {
        stmt_cnt = array_size(&blk->stmts);

        for (i = 0; i < stmt_cnt; i++) {
            ast_stmt_t *prev = array_item(&blk->stmts, i, ast_stmt_t);

            if (prev->label != NULL &&
                strcmp(prev->label, stmt->u_goto.label) == 0) {
                has_found = true;
                break;
            }
        }
    } while ((blk = blk->up) != NULL);

    if (!has_found)
        THROW(ERROR_UNDEFINED_LABEL, stmt_pos(stmt), stmt->u_goto.label);

    return NO_ERROR;
}

static int
stmt_ddl_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(is_ddl_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_ddl.ddl != NULL);

    return NO_ERROR;
}

static int
stmt_blk_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(is_blk_stmt(stmt), stmt->kind);

    if (stmt->u_blk.blk != NULL)
        check_blk(check, stmt->u_blk.blk);

    return NO_ERROR;
}

void
check_stmt(check_t *check, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_NULL:
    case STMT_CONTINUE:
    case STMT_BREAK:
        break;

    case STMT_EXP:
        check_exp(check, stmt->u_exp.exp);
        break;

    case STMT_IF:
        stmt_if_check(check, stmt);
        break;

    case STMT_FOR:
        stmt_for_check(check, stmt);
        break;

    case STMT_SWITCH:
        stmt_switch_check(check, stmt);
        break;

    case STMT_RETURN:
        stmt_return_check(check, stmt);
        break;

    case STMT_GOTO:
        stmt_goto_check(check, stmt);
        break;

    case STMT_DDL:
        stmt_ddl_check(check, stmt);
        break;

    case STMT_BLK:
        stmt_blk_check(check, stmt);
        break;

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }
}

/* end of check_stmt.c */
