/**
 * @file    check_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"
#include "check_exp.h"
#include "check_blk.h"

#include "check_stmt.h"

static bool
stmt_check_id(check_t *check, ast_stmt_t *stmt)
{
    ast_id_t *id = stmt->u_id.id;
    ast_blk_t *blk = check->blk;

    ASSERT1(is_id_stmt(stmt), stmt->kind);
    ASSERT(id != NULL);
    ASSERT(blk != NULL);

    id_check(check, id);

    id_add(&blk->ids, id);

    return true;
}

static bool
stmt_check_exp(check_t *check, ast_stmt_t *stmt)
{
    ast_exp_t *exp = stmt->u_exp.exp;

    ASSERT1(is_exp_stmt(stmt), stmt->kind);
    ASSERT(exp != NULL);

    exp_check(check, exp);

    if (is_tuple_exp(exp)) {
        int i;

        vector_foreach(exp->u_tup.elem_exps, i) {
            ast_exp_t *elem_exp = vector_get_exp(exp->u_tup.elem_exps, i);

            if (!is_usable_stmt(elem_exp)) {
                elem_exp->kind = EXP_NULL;
                WARN(ERROR_IGNORED_STMT, &elem_exp->pos);
            }
        }
    }
    else if (!is_usable_stmt(exp)) {
        exp->kind = EXP_NULL;
        WARN(ERROR_IGNORED_STMT, &stmt->pos);
    }

    return true;
}

static void
check_overflow(ast_exp_t *l_exp, ast_exp_t *r_exp)
{
    if (is_tuple_exp(l_exp)) {
        int i;
        vector_t *var_exps = l_exp->u_tup.elem_exps;
        vector_t *val_exps = r_exp->u_tup.elem_exps;

        ASSERT1(is_tuple_exp(r_exp), r_exp->kind);
        ASSERT2(vector_size(var_exps) == vector_size(val_exps), vector_size(var_exps),
                vector_size(val_exps));

        vector_foreach(var_exps, i) {
            ast_exp_t *var_exp = vector_get_exp(var_exps, i);
            ast_exp_t *val_exp = vector_get_exp(val_exps, i);

            ASSERT2(meta_cmp(&var_exp->meta, &val_exp->meta), var_exp->meta.type,
                    val_exp->meta.type);

            exp_check_overflow(val_exp, &var_exp->meta);
        }
    }
    else {
        exp_check_overflow(r_exp, &l_exp->meta);
    }
}
// TODO: multiple return values
#if 0
static void
check_overflow(ast_exp_t *l_exp, ast_exp_t *r_exp)
{
    if (is_tuple_exp(l_exp) && is_tuple_exp(r_exp)) {
        int i;
        vector_t *var_exps = l_exp->u_tup.elem_exps;
        vector_t *val_exps = r_exp->u_tup.elem_exps;

        if (vector_size(var_exps) == vector_size(val_exps)) {
            vector_foreach(var_exps, i) {
                ast_exp_t *var_exp = vector_get_exp(var_exps, i);
                ast_exp_t *val_exp = vector_get_exp(val_exps, i);

                ASSERT2(meta_cmp(&var_exp->meta, &val_exp->meta), var_exp->meta.type,
                        val_exp->meta.type);

                exp_check_overflow(val_exp, &var_exp->meta);
            }
        }
        else {
            int var_idx = 0;

            ASSERT2(vector_size(var_exps) > vector_size(val_exps), vector_size(var_exps),
                    vector_size(val_exps));

            vector_foreach(val_exps, i) {
                ast_exp_t *val_exp = vector_get_exp(val_exps, i);

                /* If the value expression is a tuple, it cannot be a literal */
                if (is_tuple_meta(&val_exp->meta)) {
                    var_idx += val_exp->meta.elem_cnt;
                }
                else {
                    ast_exp_t *var_exp = vector_get_exp(var_exps, var_idx++);

                    ASSERT2(meta_cmp(&var_exp->meta, &val_exp->meta),
                            var_exp->meta.type, val_exp->meta.type);

                    exp_check_overflow(val_exp, &var_exp->meta);
                }
            }
        }
    }
    else {
        exp_check_overflow(r_exp, &l_exp->meta);
    }
}
#endif

static bool
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
        vector_t *var_exps = l_exp->u_tup.elem_exps;

        vector_foreach(var_exps, i) {
            ast_exp_t *var_exp = vector_get_exp(var_exps, i);

            if (!is_usable_lval(var_exp))
                RETURN(ERROR_INVALID_LVALUE, &var_exp->pos);
        }
    }
    else if (!is_usable_lval(l_exp)) {
        RETURN(ERROR_INVALID_LVALUE, &l_exp->pos);
    }

    CHECK(meta_cmp(l_meta, r_meta));

    meta_eval(l_meta, r_meta);

    check_overflow(l_exp, r_exp);

    return true;
}

static bool
stmt_check_if(check_t *check, ast_stmt_t *stmt)
{
    int i;
    ast_exp_t *cond_exp;
    meta_t *cond_meta;
    vector_t *elif_stmts;

    ASSERT1(is_if_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_if.cond_exp != NULL);

    cond_exp = stmt->u_if.cond_exp;
    cond_meta = &cond_exp->meta;

    CHECK(exp_check(check, cond_exp));

    if (!is_bool_meta(cond_meta))
        RETURN(ERROR_INVALID_COND_TYPE, &cond_exp->pos, meta_to_str(cond_meta));

    if (stmt->u_if.if_blk != NULL)
        blk_check(check, stmt->u_if.if_blk);

    elif_stmts = &stmt->u_if.elif_stmts;

    vector_foreach(elif_stmts, i) {
        stmt_check_if(check, vector_get_stmt(elif_stmts, i));
    }

    if (stmt->u_if.else_blk != NULL)
        blk_check(check, stmt->u_if.else_blk);

    return true;
}

static bool
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
        vector_add_first(&blk->stmts, break_stmt);
    }

    if (stmt->u_loop.init_stmt != NULL)
        vector_add_first(&blk->stmts, stmt->u_loop.init_stmt);

    if (loop_exp != NULL)
        vector_add_last(&blk->stmts, stmt_new_exp(loop_exp, &loop_exp->pos));

    return true;
}

static bool
stmt_check_array_loop(check_t *check, ast_stmt_t *stmt)
{
    ast_id_t *iter_id;
    ast_exp_t *iter_exp;
    ast_exp_t *val_exp;
    ast_exp_t *arr_exp;
    ast_exp_t *size_exp;
    ast_exp_t *cmp_exp;
    ast_exp_t *loop_exp;
    ast_stmt_t *init_stmt = stmt->u_loop.init_stmt;
    ast_exp_t *cond_exp = stmt->u_loop.cond_exp;
    ast_blk_t *blk = stmt->u_loop.blk;
    src_pos_t *pos = &stmt->pos;

    ASSERT(init_stmt != NULL);
    ASSERT(cond_exp != NULL);
    ASSERT(stmt->u_loop.loop_exp == NULL);

    /* TODO: map & sql iteration */
    CHECK(exp_check(check, cond_exp));

    if (is_map_meta(&cond_exp->meta))
        RETURN(ERROR_NOT_SUPPORTED, &cond_exp->pos);

    /* array-loop is converted like this:
     *
     *  int arr_idx = 0;
     *  loop {
     *      if (arr_idx >= array.size)
     *          break;
     *      val = array[arr_idx];
     *      ...
     *      arr_idx++;
     *  }
     */

    /* Make "int arr_idx = 0" */
    iter_id = id_new_var("array$idx", MOD_PRIVATE, pos);

    iter_id->u_var.type_exp = exp_new_type(TYPE_INT32, pos);
    iter_id->u_var.size_exps = NULL;
    iter_id->u_var.dflt_exp = exp_new_lit_i64(0, pos);

    /* Make "arr_idx" expression */
    iter_exp = exp_new_id(iter_id->name, pos);

    /* Make "array" expression */
    if (is_id_stmt(init_stmt)) {
        val_exp = exp_new_id(init_stmt->u_id.id->name, pos);
    }
    else {
        ASSERT1(is_exp_stmt(init_stmt), init_stmt->kind);
        val_exp = init_stmt->u_exp.exp;
    }

    /* Make "val = array[arr_idx]" statement */
    arr_exp = exp_new_array(cond_exp, iter_exp, pos);
    vector_add_first(&blk->stmts, stmt_new_assign(val_exp, arr_exp, pos));

    /* Make "break if arr_idx >= array.size" statement */
    size_exp = exp_new_access(cond_exp, exp_new_id("size", pos), pos);
    cmp_exp = exp_new_binary(OP_GE, iter_exp, size_exp, pos);

    vector_add_first(&blk->stmts, stmt_new_jump(STMT_BREAK, cmp_exp, pos));

    if (is_id_stmt(init_stmt))
        vector_add_first(&blk->stmts, init_stmt);

    vector_add_first(&blk->stmts, stmt_new_id(iter_id, pos));

    /* Make "arr_idx++" expression */
    loop_exp = exp_new_unary(OP_INC, false, iter_exp, pos);
    vector_add_last(&blk->stmts, stmt_new_exp(loop_exp, pos));

    return true;
}

static bool
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

    return true;
}

static bool
stmt_check_switch(check_t *check, ast_stmt_t *stmt)
{
    int i, j;
    bool is_case_blk = false;
    ast_blk_t *blk;
    ast_exp_t *cond_exp;

    ASSERT1(is_switch_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_sw.blk != NULL);

    blk = stmt->u_sw.blk;
    cond_exp = stmt->u_sw.cond_exp;

    vector_foreach(&blk->stmts, i) {
        ast_stmt_t *elem_stmt = vector_get_stmt(&blk->stmts, i);

        if (is_case_stmt(elem_stmt)) {
            ast_exp_t *val_exp = elem_stmt->u_case.val_exp;

            if (val_exp == NULL) {
                if (stmt->u_sw.has_dflt)
                    RETURN(ERROR_DUPLICATED_LABEL, &elem_stmt->pos, "default");

                stmt->u_sw.has_dflt = true;
            }
            else {
                for (j = i + 1; j < vector_size(&blk->stmts); j++) {
                    ast_stmt_t *next_stmt = vector_get_stmt(&blk->stmts, j);
                    ast_exp_t *next_exp = next_stmt->u_case.val_exp;

                    if (!is_case_stmt(next_stmt))
                        continue;

                    if (next_exp != NULL && exp_equals(val_exp, next_exp))
                        RETURN(ERROR_DUPLICATED_CASE, &next_exp->pos);
                }

                if (cond_exp != NULL)
                    elem_stmt->u_case.val_exp =
                        exp_new_binary(OP_EQ, cond_exp, val_exp, &val_exp->pos);
            }

            is_case_blk = true;
        }
        else if (is_break_stmt(elem_stmt) || is_return_stmt(elem_stmt)) {
            is_case_blk = false;
        }
        else if (!is_case_blk) {
            ERROR(ERROR_INVALID_CASE, &elem_stmt->pos);
        }
    }

    blk_check(check, blk);

    return true;
}

static bool
stmt_check_case(check_t *check, ast_stmt_t *stmt)
{
    ast_exp_t *val_exp = stmt->u_case.val_exp;

    ASSERT1(is_case_stmt(stmt), stmt->kind);

    if (val_exp != NULL) {
        meta_t *val_meta = &val_exp->meta;

        exp_check(check, val_exp);

        if (!is_bool_meta(val_meta))
            RETURN(ERROR_INVALID_COND_TYPE, &val_exp->pos, meta_to_str(val_meta));
    }

    return true;
}

static bool
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
        if (is_void_meta(fn_meta) || is_ctor_id(fn_id))
            RETURN(ERROR_MISMATCHED_COUNT, &arg_exp->pos, "return", 0,
                   meta_cnt(&arg_exp->meta));

        CHECK(exp_check(check, arg_exp));
        CHECK(meta_cmp(fn_meta, &arg_exp->meta));

        meta_eval(fn_meta, &arg_exp->meta);

        exp_check_overflow(arg_exp, fn_meta);

        // TODO: multiple return values
#if 0
        if (is_tuple_exp(arg_exp)) {
            int i;

            ASSERT1(is_tuple_meta(fn_meta), fn_meta->type);
            ASSERT2(vector_size(arg_exp->u_tup.elem_exps) == fn_meta->elem_cnt,
                    vector_size(arg_exp->u_tup.elem_exps), fn_meta->elem_cnt);

            vector_foreach(arg_exp->u_tup.elem_exps, i) {
                meta_t *var_meta = fn_meta->elems[i];
                ast_exp_t *val_exp = vector_get_exp(arg_exp->u_tup.elem_exps, i);

                exp_check_overflow(val_exp, var_meta);
            }
        }
        else {
            exp_check_overflow(arg_exp, fn_meta);
        }
#endif
    }
    else if (!is_void_meta(fn_meta) && !is_ctor_id(fn_id)) {
        RETURN(ERROR_MISMATCHED_COUNT, &stmt->pos, "return", meta_cnt(fn_meta), 0);
    }

    /* Because each "arg_exp" may already be set to "id", we have to set the function's
     * ret_id itself in the statement */
    stmt->u_ret.ret_id = fn_id->u_fn.ret_id;

    return true;
}

static bool
stmt_check_continue(check_t *check, ast_stmt_t *stmt)
{
    ast_blk_t *blk;

    ASSERT1(is_continue_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_jump.cond_exp == NULL);

    blk = blk_search(check->blk, BLK_LOOP);
    if (blk == NULL)
        RETURN(ERROR_INVALID_CONTINUE, &stmt->pos);

    return true;
}

static bool
stmt_check_break(check_t *check, ast_stmt_t *stmt)
{
    ast_exp_t *cond_exp;
    ast_blk_t *blk;

    ASSERT1(is_break_stmt(stmt), stmt->kind);

    cond_exp = stmt->u_jump.cond_exp;

    if (cond_exp != NULL) {
        meta_t *cond_meta = &cond_exp->meta;

        CHECK(exp_check(check, cond_exp));

        if (!is_bool_meta(cond_meta))
            RETURN(ERROR_INVALID_COND_TYPE, &cond_exp->pos, meta_to_str(cond_meta));
    }

    blk = blk_search(check->blk, BLK_LOOP);
    if (blk == NULL) {
        blk = blk_search(check->blk, BLK_SWITCH);
        if (blk == NULL)
            RETURN(ERROR_INVALID_BREAK, &stmt->pos);
    }

    return true;
}

static bool
stmt_check_goto(check_t *check, ast_stmt_t *stmt)
{
    ast_id_t *fn_id = check->fn_id;

    ASSERT1(is_goto_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_goto.label != NULL);
    ASSERT(fn_id != NULL);

    stmt->u_goto.jump_id = blk_search_label(fn_id->u_fn.blk, stmt->u_goto.label);
    if (stmt->u_goto.jump_id == NULL)
        RETURN(ERROR_UNDEFINED_LABEL, &stmt->pos, stmt->u_goto.label);

    return true;
}

static bool
stmt_check_ddl(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(is_ddl_stmt(stmt), stmt->kind);
    ASSERT(stmt->u_ddl.ddl != NULL);

    /* TODO */
    return true;
}

static bool
stmt_check_blk(check_t *check, ast_stmt_t *stmt)
{
    ASSERT1(is_blk_stmt(stmt), stmt->kind);

    if (stmt->u_blk.blk != NULL)
        blk_check(check, stmt->u_blk.blk);

    return true;
}

void
stmt_check(check_t *check, ast_stmt_t *stmt)
{
    ASSERT(stmt != NULL);

    switch (stmt->kind) {
    case STMT_NULL:
        return;

    case STMT_ID:
        stmt_check_id(check, stmt);
        break;

    case STMT_EXP:
        stmt_check_exp(check, stmt);
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

    case STMT_CASE:
        stmt_check_case(check, stmt);
        break;

    case STMT_RETURN:
        stmt_check_return(check, stmt);
        break;

    case STMT_CONTINUE:
        stmt_check_continue(check, stmt);
        break;

    case STMT_BREAK:
        stmt_check_break(check, stmt);
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
