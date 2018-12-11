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
stmt_check_assign(check_t *check, ast_stmt_t *stmt)
{
    int i;
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT1(is_assign_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_assign.l_exp != NULL);
    ASSERT(stmt->u_assign.r_exp != NULL);

    l_exp = stmt->u_assign.l_exp;
    l_meta = &l_exp->meta;

    CHECK(exp_check(check, l_exp));

    r_exp = stmt->u_assign.r_exp;
    r_meta = &r_exp->meta;

    CHECK(exp_check(check, r_exp));

    if (is_tuple_exp(l_exp)) {
        array_t *var_exps = l_exp->u_tup.exps;

        for (i = 0; i < array_size(var_exps); i++) {
            ast_exp_t *var_exp = array_get(var_exps, i, ast_exp_t);

            if (!is_usable_lval(var_exp))
                RETURN(ERROR_INVALID_LVALUE, &var_exp->pos);
        }
    }
    else if (!is_usable_lval(l_exp)) {
        RETURN(ERROR_INVALID_LVALUE, &l_exp->pos);
    }

    CHECK(meta_cmp(l_meta, r_meta));

    if (is_tuple_exp(l_exp)) {
        array_t *var_exps = l_exp->u_tup.exps;
        array_t *val_exps = r_exp->u_tup.exps;

        if (is_tuple_exp(r_exp) && array_size(var_exps) == array_size(val_exps)) {
            for (i = 0; i < array_size(var_exps); i++) {
                ast_exp_t *var_exp = array_get(var_exps, i, ast_exp_t);
                ast_exp_t *val_exp = array_get(val_exps, i, ast_exp_t);

                if (is_lit_exp(val_exp) &&
                    !value_fit(&val_exp->u_lit.val, &var_exp->meta))
                    RETURN(ERROR_NUMERIC_OVERFLOW, &val_exp->pos,
                           meta_to_str(&var_exp->meta));
            }
        }
    }
    else if (is_lit_exp(r_exp) && !value_fit(&r_exp->u_lit.val, l_meta)) {
        RETURN(ERROR_NUMERIC_OVERFLOW, &r_exp->pos, meta_to_str(l_meta));
    }

    return NO_ERROR;
}

static int
stmt_check_if(check_t *check, ast_stmt_t *stmt)
{
    int i;
    ast_exp_t *cond_exp;
    meta_t *cond_meta;
    array_t *elif_stmts;

    ASSERT1(is_if_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_if.cond_exp != NULL);

    cond_exp = stmt->u_if.cond_exp;
    cond_meta = &cond_exp->meta;

    CHECK(exp_check(check, cond_exp));

    if (!is_bool_type(cond_meta))
        RETURN(ERROR_INVALID_COND_TYPE, &cond_exp->pos, meta_to_str(cond_meta));

    if (stmt->u_if.if_blk != NULL)
        blk_check(check, stmt->u_if.if_blk);

    elif_stmts = &stmt->u_if.elif_stmts;

    if (stmt->u_if.else_blk != NULL) {
        ASSERT1(array_size(elif_stmts) == 0, array_size(elif_stmts));
        blk_check(check, stmt->u_if.else_blk);
    }
    else {
        for (i = 0; i < array_size(elif_stmts); i++) {
            stmt_check_if(check, array_get(elif_stmts, i, ast_stmt_t));
        }
    }

    return NO_ERROR;
}

static int
stmt_check_for_loop(check_t *check, ast_stmt_t *stmt)
{
    ast_exp_t *cond_exp = stmt->u_loop.cond_exp;
    ast_exp_t *loop_exp = stmt->u_loop.loop_exp;
    ast_blk_t *blk = stmt->u_loop.blk;

    /* for-loop is converted like this:
     *
     *  init_stmt;
     *  loop {
     *      if (!cond_exp)
     *          break;
     *      ...
     *      loop_exp;
     *  }
     */

    if (cond_exp != NULL) {
        ast_blk_t *if_blk;
        ast_stmt_t *break_stmt;
        ast_exp_t *not_exp;
        ast_stmt_t *if_stmt;

        if_blk = blk_new_normal(&cond_exp->pos);

        break_stmt = stmt_new_jump(STMT_BREAK, &cond_exp->pos);
        array_add_last(&if_blk->stmts, break_stmt);

        not_exp = exp_new_unary(OP_NOT, cond_exp, &cond_exp->pos);

        if_stmt = stmt_new_if(not_exp, if_blk, &cond_exp->pos);
        array_add_first(&blk->stmts, if_stmt);
    }

    if (stmt->u_loop.init_ids != NULL) {
        ASSERT(stmt->u_loop.init_stmt == NULL);
        id_join_first(&blk->ids, stmt->u_loop.init_ids);
    }
    else if (stmt->u_loop.init_stmt != NULL) {
        stmt_check(check, stmt->u_loop.init_stmt);
    }

    if (loop_exp != NULL)
        array_add_last(&blk->stmts, stmt_new_exp(loop_exp, &loop_exp->pos));

    return NO_ERROR;
}

static int
stmt_check_array_loop(check_t *check, ast_stmt_t *stmt)
{
    /*
    char name[128];
    ast_id_t *id;
    ast_exp_t *inc_exp;
    ast_exp_t *arr_exp;
    ast_exp_t *assign_exp;
    ast_exp_t *loop_exp;
    ast_stmt_t *null_stmt;
    ast_blk_t *blk = stmt->u_loop.blk;
    src_pos_t *pos = &stmt->pos;
    */

    /* TODO: map & sql iteration */

    /* array-loop is converted like this:
     *
     *      int array_idx_xxx = 0;
     *
     *      init_exp;               // only if init_exp != NULL
     *
     *  for_loop_xxx:
     *      if (array_idx_xxx >= array.size)
     *          goto for_exit_xxx;
     *
     *      variable = arr_exp[array_idx_xxx];
     *
     *      ...
     *
     *  for_cont_xxx:
     *      array_idx_xxx++;
     *      goto for_loop_xxx;
     *
     *  for_exit_xxx:
     *      ;
     */

    /* TODO: we need to know array.size */
    RETURN(ERROR_NOT_SUPPORTED, &stmt->pos);
#if 0
    loop_exp = stmt->u_loop.loop_exp;
    ASSERT(loop_exp != NULL);

    /* make "int i = 0" */
    snprintf(name, sizeof(name), "array_idx_%d", blk->num);

    id = id_new_var(xstrdup(name), MOD_PRIVATE, pos);

    id->u_var.type_exp = exp_new_type(TYPE_INT32, pos);
    id->u_var.size_exps = NULL;
    id->u_var.init_exp = exp_new_lit(pos);
    value_set_i64(&id->u_var.init_exp->u_lit.val, 0);

    id_add_last(&blk->ids, id);

    inc_exp = exp_new_op(OP_INC, exp_new_ref(xstrdup(name), pos), NULL, pos);
    arr_exp = exp_new_array(loop_exp, inc_exp, &loop_exp->pos);

    if (stmt->u_loop.init_ids != NULL) {
        int i;
        array_t *var_ids = stmt->u_loop.init_ids;

        if (array_size(var_ids) > 1)
            RETURN(ERROR_NOT_SUPPORTED, &array_get(var_ids, 1, ast_id_t)->pos);

        /* make "variable = loop_exp[i++]" */
        for (i = 0; i < array_size(var_ids); i++) {
            ast_id_t *var_id = array_get(var_ids, i, ast_id_t);
            ast_exp_t *id_exp;

            id_exp = exp_new_ref(var_id->name, pos);
            assign_exp = exp_new_op(OP_ASSIGN, id_exp, arr_exp, &loop_exp->pos);

            array_add_first(&blk->stmts, stmt_new_exp(assign_exp, pos));
        }

        id_join_first(&blk->ids, var_ids);
    }
    else {
        ast_exp_t *init_exp = stmt->u_loop.init_exp;

        ASSERT(init_exp != NULL);

        if (is_tuple_exp(init_exp))
            RETURN(ERROR_NOT_SUPPORTED, &init_exp->pos);

        /* make "init_exp = loop_exp[i++]" */
        assign_exp = exp_new_op(OP_ASSIGN, init_exp, arr_exp, &loop_exp->pos);

        array_add_first(&blk->stmts, stmt_new_exp(assign_exp, pos));
    }

    null_stmt = stmt_new_null(&stmt->pos);
    null_stmt->label = blk->loop_label;

    array_add_first(&blk->stmts, null_stmt);
#endif

    return NO_ERROR;
}

static int
stmt_check_loop(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(is_loop_stmt(stmt), stmt->kind);

    if (stmt->u_loop.blk == NULL)
        stmt->u_loop.blk = blk_new_loop(&stmt->pos);

    switch (stmt->u_loop.kind) {
    case LOOP_FOR:
        stmt_check_for_loop(check, stmt);
        break;

    case LOOP_ARRAY:
        stmt_check_array_loop(check, stmt);
        break;

    default:
        ASSERT1(!"invalid loop", stmt->u_loop.kind);
    }

    blk_check(check, stmt->u_loop.blk);

    return NO_ERROR;
}

static int
stmt_check_case(check_t *check, ast_stmt_t *stmt, meta_t *meta)
{
    int i;
    ast_exp_t *val_exp;
    array_t *stmts;

    ASSERT1(is_case_stmt(stmt), stmt->kind);

    val_exp = stmt->u_case.val_exp;

    if (val_exp != NULL) {
        meta_t *val_meta = &val_exp->meta;

        exp_check(check, val_exp);

        if (meta == NULL) {
            if (!is_bool_type(val_meta))
                RETURN(ERROR_INVALID_COND_TYPE, &val_exp->pos, meta_to_str(val_meta));
        }
        else {
            meta_cmp(meta, val_meta);
        }
    }

    stmts = stmt->u_case.stmts;

    for (i = 0; i < array_size(stmts); i++) {
        stmt_check(check, array_get(stmts, i, ast_stmt_t));
    }

    return NO_ERROR;
}

static int
stmt_check_switch(check_t *check, ast_stmt_t *stmt)
{
    int i;
    ast_exp_t *cond_exp;
    meta_t *cond_meta = NULL;
    array_t *case_stmts;

    ASSERT1(is_switch_stmt(stmt), stmt->kind);

    cond_exp = stmt->u_sw.cond_exp;

    if (cond_exp != NULL) {
        cond_meta = &cond_exp->meta;

        exp_check(check, cond_exp);

        if (!is_comparable_type(cond_meta))
            RETURN(ERROR_NOT_COMPARABLE_TYPE, &cond_exp->pos, meta_to_str(cond_meta));
    }

    case_stmts = stmt->u_sw.case_stmts;

    check->is_in_switch = true;

    for (i = 0; i < array_size(case_stmts); i++) {
        stmt_check_case(check, array_get(case_stmts, i, ast_stmt_t), cond_meta);
    }

    check->is_in_switch = false;

    return NO_ERROR;
}

static int
stmt_check_return(check_t *check, ast_stmt_t *stmt)
{
    ast_id_t *func_id;
    meta_t *fn_meta;
    ast_exp_t *arg_exp;

    ASSERT1(is_return_stmt(stmt), stmt->kind);
    ASSERT(check->func_id != NULL);

    func_id = check->func_id;
    fn_meta = &func_id->meta;

    ASSERT1(is_func_id(func_id), func_id->kind);

    arg_exp = stmt->u_ret.arg_exp;

    if (arg_exp != NULL) {
        if (is_void_type(fn_meta))
            RETURN(ERROR_MISMATCHED_COUNT, &arg_exp->pos, "argument", 0,
                   meta_cnt(&arg_exp->meta));

        exp_check(check, arg_exp);

        return meta_cmp(fn_meta, &arg_exp->meta);
    }
    else if (!is_void_type(fn_meta)) {
        RETURN(ERROR_MISMATCHED_COUNT, &stmt->pos, "argument",
               meta_cnt(fn_meta), 0);
    }

    return NO_ERROR;
}

static int
stmt_check_jump(check_t *check, ast_stmt_t *stmt)
{
    ast_blk_t *blk;

    ASSERT1(is_continue_stmt(stmt) || is_break_stmt(stmt), stmt->kind);

    blk = blk_search_loop(check->blk);
    if (!check->is_in_switch && (blk == NULL || blk->name[0] == '\0'))
        RETURN(ERROR_INVALID_JUMP_STMT, &stmt->pos, STMT_KIND(stmt));

    stmt->u_jump.label = blk->name;

    return NO_ERROR;
}

static int
stmt_check_goto(check_t *check, ast_stmt_t *stmt)
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
            ast_stmt_t *prev = array_get(&blk->stmts, i, ast_stmt_t);

            if (prev->label != NULL && strcmp(prev->label, stmt->u_goto.label) == 0) {
                has_found = true;
                break;
            }
        }
    } while ((blk = blk->up) != NULL);

    if (!has_found)
        RETURN(ERROR_UNDEFINED_LABEL, &stmt->pos, stmt->u_goto.label);

    return NO_ERROR;
}

static int
stmt_check_ddl(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(is_ddl_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_ddl.ddl != NULL);

    return NO_ERROR;
}

static int
stmt_check_blk(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(is_blk_stmt(stmt), stmt->kind);

    if (stmt->u_blk.blk != NULL)
        blk_check(check, stmt->u_blk.blk);

    return NO_ERROR;
}

void
stmt_check(check_t *check, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_NULL:
        break;

    case STMT_EXP:
        exp_check(check, stmt->u_exp.exp);
        break;

    case STMT_ASSIGN:
        stmt_check_assign(check, stmt);
        break;

    case STMT_IF:
        stmt_check_if(check, stmt);
        break;

    case STMT_LOOP:
        stmt_check_loop(check, stmt);
        break;

    case STMT_SWITCH:
        stmt_check_switch(check, stmt);
        break;

    case STMT_RETURN:
        stmt_check_return(check, stmt);
        break;

    case STMT_CONTINUE:
        stmt_check_jump(check, stmt);
        break;

    case STMT_BREAK:
        stmt_check_jump(check, stmt);
        break;

    case STMT_GOTO:
        stmt_check_goto(check, stmt);
        break;

    case STMT_DDL:
        stmt_check_ddl(check, stmt);
        break;

    case STMT_BLK:
        stmt_check_blk(check, stmt);
        break;

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }
}

/* end of check_stmt.c */
