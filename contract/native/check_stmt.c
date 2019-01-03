/**
 * @file    check_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"
#include "check_exp.h"
#include "check_blk.h"

#include "check_stmt.h"

static void
check_symm_assign(check_t *check, array_t *var_exps, array_t *val_exps)
{
    int i;

    array_foreach(var_exps, i) {
        ast_exp_t *var_exp = array_get_exp(var_exps, i);
        ast_exp_t *val_exp = array_get_exp(val_exps, i);

        ASSERT2(meta_cmp(&var_exp->meta, &val_exp->meta) == 0,
                var_exp->meta.type, val_exp->meta.type);

        exp_check_overflow(val_exp, &var_exp->meta);
    }
}

static void
check_asymm_assign(check_t *check, array_t *var_exps, array_t *val_exps)
{
    int i, j;
    int var_idx = 0;

    ASSERT2(array_size(var_exps) > array_size(val_exps), array_size(var_exps),
            array_size(val_exps));

    array_foreach(val_exps, i) {
        ast_exp_t *val_exp = array_get_exp(val_exps, i);

        /* The number of value expressions is known only through meta.
         * However, after the transform, we can access it in expression form.
         * See stmt_trans_assign() */

        if (is_tuple_type(&val_exp->meta)) {
            for (j = 0; j < val_exp->meta.elem_cnt; j++) {
                ast_exp_t *var_exp = array_get_exp(var_exps, var_idx++);

                ASSERT2(meta_cmp(&var_exp->meta, val_exp->meta.elems[j]) == 0,
                        var_exp->meta.type, val_exp->meta.elems[j]->type);
            }
        }
        else {
            ast_exp_t *var_exp = array_get_exp(var_exps, var_idx++);

            ASSERT2(meta_cmp(&var_exp->meta, &val_exp->meta) == 0,
                    var_exp->meta.type, val_exp->meta.type);

            exp_check_overflow(val_exp, &var_exp->meta);
        }
    }
}

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
        array_t *var_exps = l_exp->u_tup.elem_exps;

        array_foreach(var_exps, i) {
            ast_exp_t *var_exp = array_get_exp(var_exps, i);

            if (!is_usable_lval(var_exp))
                RETURN(ERROR_INVALID_LVALUE, &var_exp->pos);
        }
    }
    else if (!is_usable_lval(l_exp)) {
        RETURN(ERROR_INVALID_LVALUE, &l_exp->pos);
    }

    CHECK(meta_cmp(l_meta, r_meta));

    meta_eval(l_meta, r_meta);

    if (is_tuple_exp(l_exp) && is_tuple_exp(r_exp)) {
        array_t *var_exps = l_exp->u_tup.elem_exps;
        array_t *val_exps = r_exp->u_tup.elem_exps;

        if (array_size(var_exps) == array_size(val_exps))
            check_symm_assign(check, var_exps, val_exps);
        else
            check_asymm_assign(check, var_exps, val_exps);
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

    array_foreach(elif_stmts, i) {
        stmt_check_if(check, array_get_stmt(elif_stmts, i));
    }

    if (stmt->u_if.else_blk != NULL)
        blk_check(check, stmt->u_if.else_blk);

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
        ast_exp_t *not_exp;
        ast_stmt_t *break_stmt;

        not_exp = exp_new_unary(OP_NOT, true, cond_exp, &cond_exp->pos);

        break_stmt = stmt_new_jump(STMT_BREAK, not_exp, &cond_exp->pos);
        array_add_first(&blk->stmts, break_stmt);
    }

    if (stmt->u_loop.init_id != NULL) {
        ASSERT(stmt->u_loop.init_stmt == NULL);
        array_add_first(&blk->ids, stmt->u_loop.init_id);
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

    inc_exp = exp_new_op(OP_INC, exp_new_id_ref(xstrdup(name), pos), NULL, pos);
    arr_exp = exp_new_array(loop_exp, inc_exp, &loop_exp->pos);

    if (stmt->u_loop.init_ids != NULL) {
        int i;
        array_t *elem_ids = stmt->u_loop.init_ids;

        if (array_size(elem_ids) > 1)
            RETURN(ERROR_NOT_SUPPORTED, &array_get_id(elem_ids, 1)->pos);

        /* make "variable = loop_exp[i++]" */
        array_foreach(elem_ids, i) {
            ast_id_t *var_id = array_get_id(elem_ids, i);
            ast_exp_t *id_exp;

            id_exp = exp_new_id_ref(var_id->name, pos);
            assign_exp = exp_new_op(OP_ASSIGN, id_exp, arr_exp, &loop_exp->pos);

            array_add_first(&blk->stmts, stmt_new_exp(assign_exp, pos));
        }

        id_join_first(&blk->ids, elem_ids);
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
stmt_check_switch(check_t *check, ast_stmt_t *stmt)
{
    int i, j;
    ast_blk_t *blk;
    ast_exp_t *cond_exp;

    ASSERT1(is_switch_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_sw.blk != NULL);

    blk = stmt->u_sw.blk;
    cond_exp = stmt->u_sw.cond_exp;

    array_foreach(&blk->stmts, i) {
        ast_stmt_t *case_stmt = array_get_stmt(&blk->stmts, i);
        ast_exp_t *val_exp = case_stmt->u_case.val_exp;

        ASSERT1(is_case_stmt(case_stmt), case_stmt->kind);

        if (val_exp == NULL) {
            if (stmt->u_sw.has_dflt)
                RETURN(ERROR_DUPLICATED_LABEL, &case_stmt->pos, "default");

            stmt->u_sw.has_dflt = true;
        }
        else {
            for (j = i + 1; j < array_size(&blk->stmts); j++) {
                ast_stmt_t *next_case = array_get_stmt(&blk->stmts, j);
                ast_exp_t *next_val = next_case->u_case.val_exp;

                if (next_val != NULL && exp_equals(val_exp, next_val))
                    RETURN(ERROR_DUPLICATED_CASE, &next_val->pos);
            }

            if (cond_exp != NULL)
                case_stmt->u_case.val_exp =
                    exp_new_binary(OP_EQ, cond_exp, val_exp, &val_exp->pos);
        }
    }

    blk_check(check, blk);

    return NO_ERROR;
}

static int
stmt_check_case(check_t *check, ast_stmt_t *stmt)
{
    int i;
    ast_exp_t *val_exp;
    array_t *stmts;

    ASSERT1(is_case_stmt(stmt), stmt->kind);

    val_exp = stmt->u_case.val_exp;

    if (val_exp != NULL) {
        meta_t *val_meta = &val_exp->meta;

        exp_check(check, val_exp);

        if (!is_bool_type(val_meta))
            RETURN(ERROR_INVALID_COND_TYPE, &val_exp->pos, meta_to_str(val_meta));
    }

    stmts = stmt->u_case.stmts;

    array_foreach(stmts, i) {
        stmt_check(check, array_get_stmt(stmts, i));
    }

    return NO_ERROR;
}

static int
stmt_check_return(check_t *check, ast_stmt_t *stmt)
{
    ast_id_t *fn_id;
    meta_t *fn_meta;
    ast_exp_t *arg_exp;

    ASSERT1(is_return_stmt(stmt), stmt->kind);
    ASSERT(check->fn_id != NULL);

    fn_id = check->fn_id;
    fn_meta = &fn_id->meta;

    ASSERT1(is_fn_id(fn_id), fn_id->kind);

    arg_exp = stmt->u_ret.arg_exp;

    if (arg_exp != NULL) {
        if (is_void_type(fn_meta) || is_ctor_id(fn_id))
            RETURN(ERROR_MISMATCHED_COUNT, &arg_exp->pos, "return", 0,
                   meta_cnt(&arg_exp->meta));

        CHECK(exp_check(check, arg_exp));
        CHECK(meta_cmp(fn_meta, &arg_exp->meta));

        meta_eval(fn_meta, &arg_exp->meta);

        if (is_tuple_exp(arg_exp)) {
            int i;

            ASSERT1(is_tuple_type(fn_meta), fn_meta->type);
            ASSERT2(array_size(arg_exp->u_tup.elem_exps) == fn_meta->elem_cnt,
                    array_size(arg_exp->u_tup.elem_exps), fn_meta->elem_cnt);

            array_foreach(arg_exp->u_tup.elem_exps, i) {
                meta_t *var_meta = fn_meta->elems[i];
                ast_exp_t *val_exp = array_get_exp(arg_exp->u_tup.elem_exps, i);

                exp_check_overflow(val_exp, var_meta);
            }
        }
        else {
            exp_check_overflow(arg_exp, fn_meta);
        }
    }
    else if (!is_void_type(fn_meta) && !is_ctor_id(fn_id)) {
        RETURN(ERROR_MISMATCHED_COUNT, &stmt->pos, "return", meta_cnt(fn_meta), 0);
    }

    stmt->u_ret.ret_id = fn_id->u_fn.ret_id;

    return NO_ERROR;
}

static int
stmt_check_continue(check_t *check, ast_stmt_t *stmt)
{
    ast_blk_t *blk;

    ASSERT1(is_continue_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_jump.cond_exp == NULL);

    blk = blk_search(check->blk, BLK_LOOP);
    if (blk == NULL)
        RETURN(ERROR_INVALID_JUMP_STMT, &stmt->pos, STMT_KIND(stmt));

    return NO_ERROR;
}

static int
stmt_check_break(check_t *check, ast_stmt_t *stmt)
{
    ast_exp_t *cond_exp;
    ast_blk_t *blk;

    ASSERT1(is_break_stmt(stmt), stmt->kind);

    cond_exp = stmt->u_jump.cond_exp;

    if (cond_exp != NULL) {
        meta_t *cond_meta = &cond_exp->meta;

        CHECK(exp_check(check, cond_exp));

        if (!is_bool_type(cond_meta))
            RETURN(ERROR_INVALID_COND_TYPE, &cond_exp->pos, meta_to_str(cond_meta));
    }

    blk = blk_search(check->blk, BLK_LOOP);
    if (blk == NULL) {
        blk = blk_search(check->blk, BLK_SWITCH);
        if (blk == NULL)
            RETURN(ERROR_INVALID_JUMP_STMT, &stmt->pos, STMT_KIND(stmt));
    }

    return NO_ERROR;
}

static int
stmt_check_goto(check_t *check, ast_stmt_t *stmt)
{
    ast_id_t *fn_id;

    ASSERT1(is_goto_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_goto.label != NULL);

    fn_id = check->fn_id;
    ASSERT(fn_id != NULL);

    stmt->u_goto.jump_id = blk_search_label(fn_id->u_fn.blk, stmt->u_goto.label);
    if (stmt->u_goto.jump_id == NULL)
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

int
stmt_check(check_t *check, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_NULL:
        return NO_ERROR;

    case STMT_EXP:
        return exp_check(check, stmt->u_exp.exp);

    case STMT_ASSIGN:
        return stmt_check_assign(check, stmt);

    case STMT_IF:
        return stmt_check_if(check, stmt);

    case STMT_LOOP:
        return stmt_check_loop(check, stmt);

    case STMT_SWITCH:
        return stmt_check_switch(check, stmt);

    case STMT_CASE:
        return stmt_check_case(check, stmt);

    case STMT_RETURN:
        return stmt_check_return(check, stmt);

    case STMT_CONTINUE:
        return stmt_check_continue(check, stmt);

    case STMT_BREAK:
        return stmt_check_break(check, stmt);

    case STMT_GOTO:
        return stmt_check_goto(check, stmt);

    case STMT_DDL:
        return stmt_check_ddl(check, stmt);

    case STMT_BLK:
        return stmt_check_blk(check, stmt);

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }

    return NO_ERROR;
}

/* end of check_stmt.c */
