/**
 * @file    trans_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_bb.h"
#include "ir_fn.h"
#include "trans_id.h"
#include "trans_blk.h"
#include "trans_exp.h"

#include "trans_stmt.h"

static void
add_exp_stmt(trans_t *trans, ast_exp_t *exp)
{
    if (is_null_exp(exp))
        return;

    exp_trans(trans, exp);

    /* For unary increase/decrease expressions, which are postfixes,
     * add them as piggybacked statements */
    if (has_piggyback(trans->bb)) {
        int i;
        array_t *pgbacks = &trans->bb->pgbacks;

        array_foreach(pgbacks, i) {
            bb_add_stmt(trans->bb, array_get_stmt(pgbacks, i));
        }

        array_reset(pgbacks);
    }
}

static void
stmt_trans_exp(trans_t *trans, ast_stmt_t *stmt)
{
    ast_exp_t *exp = stmt->u_exp.exp;

    if (is_tuple_exp(exp)) {
        int i;

        array_foreach(exp->u_tup.elem_exps, i) {
            add_exp_stmt(trans, array_get_exp(exp->u_tup.elem_exps, i));
        }
    }
    else {
        add_exp_stmt(trans, exp);
    }
}

static void
resolve_rel(ast_exp_t *var_exp, ast_exp_t *val_exp)
{
    meta_t *meta = &var_exp->meta;

    if (is_init_exp(val_exp)) {
        ASSERT(var_exp->id != NULL);

        /* This is set here because we need the base address to store
         * each value of the initializer expression */
        val_exp->id = var_exp->id;
    }

    if (val_exp->id != NULL && is_object_meta(meta) &&
        meta->type_id != NULL && is_itf_id(meta->type_id)) {
        /* Override the meta of the variable declared as the interface type
         * with the rvalue (== contract) meta */
        if (is_fn_id(val_exp->id))
            meta_set_object(meta, val_exp->id->up);
        else
            meta_set_object(meta, val_exp->id->meta.type_id);
    }
}

static void
stmt_trans_assign(trans_t *trans, ast_stmt_t *stmt)
{
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_exp_t *r_exp = stmt->u_assign.r_exp;

    /* TODO: When assigning to a struct variable,
     * we must replace it with an assignment for each field */

    exp_trans(trans, l_exp);
    exp_trans(trans, r_exp);

    if (is_tuple_exp(l_exp) && is_tuple_exp(r_exp)) {
        /* Separate combinations of each expression made up of tuples
         * into separate assign statements */
        int i, j;
        int var_idx = 0;
        src_pos_t *pos = &stmt->pos;
        array_t *var_exps = l_exp->u_tup.elem_exps;
        array_t *val_exps = r_exp->u_tup.elem_exps;
        ast_exp_t *var_exp, *val_exp;

        /* If rvalue has a function that returns multiple values,
         * the number of left and right expressions may be different */
        if (array_size(var_exps) == array_size(val_exps)) {
            array_foreach(val_exps, i) {
                var_exp = array_get_exp(var_exps, i);
                val_exp = array_get_exp(val_exps, i);

                resolve_rel(var_exp, val_exp);
                bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, pos));
            }
            return;
        }

        /* For a function that returns the multiple value mentioned above,
         * an expression is generated for each value in the transformer and
         * finally a tuple expression is created. (see exp_trans_call()) */
        array_foreach(val_exps, i) {
            ASSERT1(var_idx < array_size(var_exps), var_idx);

            val_exp = array_get_exp(val_exps, i);

            if (is_tuple_exp(val_exp)) {
                ast_exp_t *elem_exp;

                array_foreach(val_exp->u_tup.elem_exps, j) {
                    var_exp = array_get_exp(var_exps, var_idx++);
                    elem_exp = array_get_exp(val_exp->u_tup.elem_exps, j);

                    resolve_rel(var_exp, elem_exp);
                    bb_add_stmt(trans->bb, stmt_new_assign(var_exp, elem_exp, pos));
                }
            }
            else {
                var_exp = array_get_exp(var_exps, var_idx++);

                resolve_rel(var_exp, val_exp);
                bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, pos));
            }
        }
    }
    else {
        ASSERT(!is_tuple_exp(l_exp));

        resolve_rel(l_exp, r_exp);
        bb_add_stmt(trans->bb, stmt);
    }
}

static void
stmt_trans_if(trans_t *trans, ast_stmt_t *stmt)
{
    int i;
    ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *next_bb = bb_new();
    array_t *elif_stmts = &stmt->u_if.elif_stmts;

    /* The if statement is transformed to a combination of basic blocks,
     * each condition is used as a branch condition, and the else block is
     * transformed by an unconditional branch
     *
     *         .---------------------------.
     *         |         prev_bb           |
     *         '---------------------------'
     *         /           / \              \
     *  .------. .---------. .---------.     .------.
     *  |  if  | | else if | | else if | ... | else |
     *  '------' '---------' '---------'     '------'
     *         \           \ /              /
     *         .---------------------------.
     *         |         next_bb           |
     *         '---------------------------'
     */

    fn_add_basic_blk(trans->fn, prev_bb);

    trans->bb = bb_new();
    bb_add_branch(prev_bb, stmt->u_if.cond_exp, trans->bb);

    exp_trans(trans, stmt->u_if.cond_exp);

    if (stmt->u_if.if_blk != NULL)
        blk_trans(trans, stmt->u_if.if_blk);

    bb_add_branch(trans->bb, NULL, next_bb);

    fn_add_basic_blk(trans->fn, trans->bb);

    array_foreach(elif_stmts, i) {
        ast_stmt_t *elif_stmt = array_get_stmt(elif_stmts, i);

        trans->bb = bb_new();
        bb_add_branch(prev_bb, elif_stmt->u_if.cond_exp, trans->bb);

        exp_trans(trans, elif_stmt->u_if.cond_exp);

        if (elif_stmt->u_if.if_blk != NULL)
            blk_trans(trans, elif_stmt->u_if.if_blk);

        bb_add_branch(trans->bb, NULL, next_bb);

        fn_add_basic_blk(trans->fn, trans->bb);
    }

    if (stmt->u_if.else_blk != NULL) {
        trans->bb = bb_new();
        bb_add_branch(prev_bb, NULL, trans->bb);

        blk_trans(trans, stmt->u_if.else_blk);

        bb_add_branch(trans->bb, NULL, next_bb);

        fn_add_basic_blk(trans->fn, trans->bb);
    }
    else {
        bb_add_branch(prev_bb, NULL, next_bb);
    }

    trans->bb = next_bb;
}

static void
stmt_trans_for_loop(trans_t *trans, ast_stmt_t *stmt)
{
    ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *cond_bb = bb_new();
    ir_bb_t *next_bb = bb_new();

    /* Each expression of the for-loop statement is separated,
     * the initial expression is added to the end of prev_bb,
     * the conditional expression is added at the beginning of cond_bb,
     * and the afterthought expression is added at the end of the loop block
     *
     *         .---------------------.
     *         | prev_bb + init_stmt |
     *         '---------------------'
     *                    |
     *              .-----------.
     *              |  cond_bb  |<---------.
     *              '-----------'          |
     *                  /   \              |
     *       .-----------. .------------.  |
     *       |  next_bb  | |  loop blk  |--'
     *       '-----------' '------------'
     */

    if (stmt->u_loop.init_stmt != NULL)
        stmt_trans(trans, stmt->u_loop.init_stmt);

    /* previous basic block */
    bb_add_branch(prev_bb, NULL, cond_bb);

    fn_add_basic_blk(trans->fn, prev_bb);

    trans->bb = cond_bb;

    trans->cont_bb = cond_bb;
    trans->break_bb = next_bb;

    blk_trans(trans, stmt->u_loop.blk);

    trans->cont_bb = NULL;
    trans->break_bb = NULL;

    if (trans->bb != NULL) {
        /* Make loop using last block and entry block */
        bb_add_branch(trans->bb, NULL, cond_bb);

        fn_add_basic_blk(trans->fn, trans->bb);
    }
    else {
        /* Make loop using self block in case of an empty loop without loop_exp */
        bb_add_branch(cond_bb, NULL, cond_bb);
    }

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
    int i;
    ast_blk_t *blk = stmt->u_sw.blk;
    ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *next_bb = bb_new();

    /* In a switch-case statement, each case block is transformed to
     * a single basic block, and the switch condition and the case value are
     * compared and used as a branch condition
     *
     *         .---------------------------.
     *         |         prev_bb           |
     *         '---------------------------'
     *            /          |           \
     *    .----------. .----------.     .---------.
     *    |  case 1  | |  case 2  | ... | default |
     *    '----------' '----------'     '---------'
     *            \          |           /
     *         .---------------------------.
     *         |         next_bb           |
     *         '---------------------------'
     */

    fn_add_basic_blk(trans->fn, prev_bb);

    trans->bb = NULL;
    trans->cont_bb = NULL;
    trans->break_bb = next_bb;

    array_foreach(&blk->stmts, i) {
        ast_stmt_t *case_stmt = array_get_stmt(&blk->stmts, i);

        /* The case statement means the start of a case block or default block,
         * and the remaining statements are included in the corresponding block */
        if (is_case_stmt(case_stmt)) {
            ir_bb_t *case_bb = bb_new();

            if (trans->bb != NULL) {
                bb_add_branch(trans->bb, NULL, case_bb);
                fn_add_basic_blk(trans->fn, trans->bb);
            }

            trans->bb = case_bb;

            /* The value of the default label can be null */
            if (case_stmt->u_case.val_exp != NULL)
                exp_trans(trans, case_stmt->u_case.val_exp);

            bb_add_branch(prev_bb, case_stmt->u_case.val_exp, trans->bb);
        }
        else {
            stmt_trans(trans, case_stmt);
        }
    }

    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, next_bb);
        fn_add_basic_blk(trans->fn, trans->bb);
    }

    if (!stmt->u_sw.has_dflt)
        bb_add_branch(prev_bb, NULL, next_bb);

    trans->break_bb = NULL;
    trans->bb = next_bb;
}

static void
stmt_trans_return(trans_t *trans, ast_stmt_t *stmt)
{
    ast_id_t *ret_id = stmt->u_ret.ret_id;
    ast_exp_t *arg_exp = stmt->u_ret.arg_exp;

    ASSERT(ret_id->up != NULL);
    ASSERT1(is_fn_id(ret_id->up), ret_id->up->kind);

    if (is_ctor_id(ret_id->up)) {
        if (arg_exp != NULL) {
            /* If "arg_exp" is not null, the "stmt" is added to exit_bb because
             * it is a return statement for contract address added forced by
             * id_trans_ctor() */
            ASSERT1(is_local_exp(arg_exp), arg_exp->kind);

            bb_add_stmt(trans->fn->exit_bb, stmt);
        }
        else {
            /* The constructor uses the return statement as is */
            bb_add_stmt(trans->bb, stmt);
        }
    }
    else if (arg_exp != NULL) {
        ast_exp_t *var_exp;

        /* Each return expression of a function corresponds to a local variable,
         * so if there is return arguments, the return statement is transformed to
         * an assign statement using the address value of each return argument */
        if (is_tuple_id(ret_id)) {
            int i;
            array_t *elem_exps = array_new();
            ast_exp_t *id_exp;

            /* Since the number of arg_exp may be smaller than the number of ret_id,
             * it is made as a tuple expression for asymmetry assignment processing */
            array_foreach(ret_id->u_tup.elem_ids, i) {
                ast_id_t *elem_id = array_get_id(ret_id->u_tup.elem_ids, i);

                id_exp = exp_new_id(elem_id->name, &elem_id->pos);

                id_exp->id = elem_id;
                meta_copy(&id_exp->meta, &elem_id->meta);

                array_add_last(elem_exps, id_exp);
            }

            var_exp = exp_new_tuple(elem_exps, &arg_exp->pos);
        }
        else {
            var_exp = exp_new_id(ret_id->name, &ret_id->pos);

            var_exp->id = ret_id;
            meta_copy(&var_exp->meta, &ret_id->meta);
        }

        stmt_trans(trans, stmt_new_assign(var_exp, arg_exp, &stmt->pos));
    }

    /* The current basic block branches directly to the exit block */
    bb_add_branch(trans->bb, NULL, trans->fn->exit_bb);

    fn_add_basic_blk(trans->fn, trans->bb);
    trans->bb = NULL;
}

static void
stmt_trans_continue(trans_t *trans, ast_stmt_t *stmt)
{
    ASSERT(trans->cont_bb != NULL);

    bb_add_branch(trans->bb, NULL, trans->cont_bb);

    fn_add_basic_blk(trans->fn, trans->bb);
    trans->bb = NULL;
}

static void
stmt_trans_break(trans_t *trans, ast_stmt_t *stmt)
{
    ir_bb_t *next_bb = bb_new();

    ASSERT(trans->break_bb != NULL);

    if (stmt->u_jump.cond_exp != NULL) {
        exp_trans(trans, stmt->u_jump.cond_exp);

        bb_add_branch(trans->bb, stmt->u_jump.cond_exp, trans->break_bb);
        bb_add_branch(trans->bb, NULL, next_bb);
    }
    else {
        bb_add_branch(trans->bb, NULL, trans->break_bb);
    }

    fn_add_basic_blk(trans->fn, trans->bb);

    trans->bb = next_bb;
}

static void
stmt_trans_goto(trans_t *trans, ast_stmt_t *stmt)
{
    ast_id_t *jump_id = stmt->u_goto.jump_id;

    ASSERT(jump_id->u_lab.stmt->label_bb != NULL);

    bb_add_branch(trans->bb, NULL, jump_id->u_lab.stmt->label_bb);
    fn_add_basic_blk(trans->fn, trans->bb);

    trans->bb = NULL;
}

static void
stmt_trans_ddl(trans_t *trans, ast_stmt_t *stmt)
{
    bb_add_stmt(trans->bb, stmt);
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
    if (stmt->label_bb != NULL) {
        /* A labeled statement always creates a new basic block */
        if (trans->bb != NULL) {
            bb_add_branch(trans->bb, NULL, stmt->label_bb);
            fn_add_basic_blk(trans->fn, trans->bb);
        }

        trans->bb = stmt->label_bb;
    }
    else if (trans->bb == NULL) {
        trans->bb = bb_new();
    }

    switch (stmt->kind) {
    case STMT_NULL:
        break;

    case STMT_EXP:
        stmt_trans_exp(trans, stmt);
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
