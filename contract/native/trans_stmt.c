/**
 * @file    trans_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_bb.h"
#include "ir_fn.h"
#include "ir_abi.h"
#include "trans_id.h"
#include "trans_blk.h"
#include "trans_exp.h"

#include "trans_stmt.h"

static void
stmt_trans_id(trans_t *trans, ast_stmt_t *stmt)
{
    int i;
    ast_id_t *id = stmt->u_id.id;

    ASSERT(id != NULL);

    if (is_global_id(id))
        /* Initialization of the global variable will be done in the constructor */
        return;

    if (is_var_id(id)) {
        if (id->u_var.dflt_exp != NULL)
            stmt_trans(trans, stmt_make_assign(id, id->u_var.dflt_exp));
    }
    else if (is_tuple_id(id)) {
        vector_foreach(id->u_tup.elem_ids, i) {
            ast_id_t *elem_id = vector_get_id(id->u_tup.elem_ids, i);

            ASSERT1(is_var_id(elem_id), elem_id->kind);

            if (elem_id->u_var.dflt_exp != NULL)
                stmt_trans(trans, stmt_make_assign(elem_id, elem_id->u_var.dflt_exp));
        }
    }
}

static void
stmt_trans_exp(trans_t *trans, ast_stmt_t *stmt)
{
    ast_exp_t *exp = stmt->u_exp.exp;

    if (is_null_exp(exp))
        return;

    if (is_tuple_exp(exp)) {
        int i;

        vector_foreach(exp->u_tup.elem_exps, i) {
            ast_exp_t *elem_exp = vector_get_exp(exp->u_tup.elem_exps, i);

            exp_trans(trans, elem_exp);

            /* Other expressions such as unary expressions have already been added, and only
             * the call expression needs to be added. */
            if (is_call_exp(elem_exp))
                bb_add_stmt(trans->bb, stmt_new_exp(elem_exp, &elem_exp->pos));
        }
    }
    else {
        exp_trans(trans, exp);

        /* same as above */
        if (is_call_exp(exp))
            bb_add_stmt(trans->bb, stmt);
    }
}

static void
resolve_var_meta(trans_t *trans, ast_exp_t *var_exp, ast_exp_t *val_exp)
{
    meta_t *meta = &var_exp->meta;

    /* Here we override the meta of the variable declared in the form
     * "interface variable = rvalue;" with the contract meta */

    /* If rvalue is the "null" literal, "val_exp->id" can be null */
    if (val_exp->id == NULL || !is_object_meta(meta) || !is_itf_id(meta->type_id))
        return;

    if (is_fn_id(val_exp->id))
        meta_set_object(meta, val_exp->id->up);
    else
        meta_set_object(meta, val_exp->id->meta.type_id);
}

static void
stmt_trans_assign(trans_t *trans, ast_stmt_t *stmt)
{
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_exp_t *r_exp = stmt->u_assign.r_exp;

    exp_trans(trans, l_exp);

    if (is_tuple_exp(l_exp)) {
        /* Make each expression a separate assignment statement */
        int i;
        vector_t *var_exps = l_exp->u_tup.elem_exps;
        vector_t *val_exps = r_exp->u_tup.elem_exps;

        ASSERT1(is_tuple_exp(r_exp), r_exp->kind);
        ASSERT2(vector_size(var_exps) == vector_size(val_exps), vector_size(var_exps),
                vector_size(val_exps));

        vector_foreach(val_exps, i) {
            ast_exp_t *var_exp = vector_get_exp(var_exps, i);
            ast_exp_t *val_exp = vector_get_exp(val_exps, i);

            resolve_var_meta(trans, var_exp, val_exp);

            if (var_exp->id != NULL && is_global_id(var_exp->id)) {
                trans->is_global = true;
                exp_trans(trans, val_exp);
                trans->is_global = false;
            }
            else {
                exp_trans(trans, val_exp);
            }

            bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, &stmt->pos));
        }
    }
    else {
        ASSERT(!is_tuple_exp(r_exp));

        resolve_var_meta(trans, l_exp, r_exp);

        if (l_exp->id != NULL && is_global_id(l_exp->id)) {
            trans->is_global = true;
            exp_trans(trans, r_exp);
            trans->is_global = false;
        }
        else {
            exp_trans(trans, r_exp);
        }

        bb_add_stmt(trans->bb, stmt);
    }
}

// TODO multiple return values
#if 0
static void
stmt_trans_assign(trans_t *trans, ast_stmt_t *stmt)
{
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_exp_t *r_exp = stmt->u_assign.r_exp;

    exp_trans(trans, l_exp);
    exp_trans(trans, r_exp);

    if (is_tuple_exp(l_exp) && is_tuple_exp(r_exp)) {
        /* Separate combinations of each expression made up of tuples into separate
         * assignment statements */
        int i, j;
        int var_idx = 0;
        src_pos_t *pos = &stmt->pos;
        vector_t *var_exps = l_exp->u_tup.elem_exps;
        vector_t *val_exps = r_exp->u_tup.elem_exps;
        ast_exp_t *var_exp, *val_exp;

        /* If rvalue has a function that returns multiple values, the number of left
         * and right expressions may be different */
        if (vector_size(var_exps) == vector_size(val_exps)) {
            vector_foreach(val_exps, i) {
                var_exp = vector_get_exp(var_exps, i);
                val_exp = vector_get_exp(val_exps, i);

                resolve_var_meta(trans, var_exp, val_exp);
                bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, pos));
            }
            return;
        }

        /* For a function that returns the multiple value mentioned above, an expression
         * is generated for each return value in the transformer and finally a tuple
         * expression is created. (see exp_trans_call()) */
        vector_foreach(val_exps, i) {
            ASSERT1(var_idx < vector_size(var_exps), var_idx);

            val_exp = vector_get_exp(val_exps, i);

            if (is_tuple_exp(val_exp)) {
                ast_exp_t *elem_exp;

                vector_foreach(val_exp->u_tup.elem_exps, j) {
                    var_exp = vector_get_exp(var_exps, var_idx++);
                    elem_exp = vector_get_exp(val_exp->u_tup.elem_exps, j);

                    resolve_var_meta(trans, var_exp, elem_exp);
                    bb_add_stmt(trans->bb, stmt_new_assign(var_exp, elem_exp, pos));
                }
            }
            else {
                var_exp = vector_get_exp(var_exps, var_idx++);

                resolve_var_meta(trans, var_exp, val_exp);
                bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, pos));
            }
        }
    }
    else {
        ASSERT(!is_tuple_exp(l_exp));

        resolve_var_meta(trans, l_exp, r_exp);
        bb_add_stmt(trans->bb, stmt);
    }
}
#endif

static void
stmt_trans_if(trans_t *trans, ast_stmt_t *stmt)
{
    int i;
    ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *next_bb = bb_new();
    vector_t *elif_stmts = &stmt->u_if.elif_stmts;

    /* The if statement is transformed to a combination of basic blocks, each condition is used as
     * a branch condition, and the else block is transformed by an unconditional branch.
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

    exp_trans(trans, stmt->u_if.cond_exp);

    fn_add_basic_blk(trans->fn, prev_bb);

    trans->bb = bb_new();

    bb_add_branch(prev_bb, stmt->u_if.cond_exp, trans->bb);

    if (stmt->u_if.if_blk != NULL)
        blk_trans(trans, stmt->u_if.if_blk);

    /* "trans->bb" can be null if the block ends with a return statement */
    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, next_bb);

        fn_add_basic_blk(trans->fn, trans->bb);
    }

    vector_foreach(elif_stmts, i) {
        ast_stmt_t *elif_stmt = vector_get_stmt(elif_stmts, i);

        trans->bb = bb_new();
        bb_add_branch(prev_bb, elif_stmt->u_if.cond_exp, trans->bb);

        exp_trans(trans, elif_stmt->u_if.cond_exp);

        if (elif_stmt->u_if.if_blk != NULL)
            blk_trans(trans, elif_stmt->u_if.if_blk);

        /* same as above */
        if (trans->bb != NULL) {
            bb_add_branch(trans->bb, NULL, next_bb);

            fn_add_basic_blk(trans->fn, trans->bb);
        }
    }

    if (stmt->u_if.else_blk != NULL) {
        trans->bb = bb_new();
        bb_add_branch(prev_bb, NULL, trans->bb);

        blk_trans(trans, stmt->u_if.else_blk);

        /* same as above */
        if (trans->bb != NULL) {
            bb_add_branch(trans->bb, NULL, next_bb);

            fn_add_basic_blk(trans->fn, trans->bb);
        }
    }
    else {
        bb_add_branch(prev_bb, NULL, next_bb);
    }

    trans->bb = next_bb;
}

static void
stmt_trans_loop(trans_t *trans, ast_stmt_t *stmt)
{
    //ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *loop_bb = bb_new();
    ir_bb_t *post_bb = bb_new();
    ir_bb_t *next_bb = bb_new();

    /* The initial expression is added to the end of prev_bb, the conditional expression is added
     * at the beginning of post_bb, and the afterthought expression is added at the end of the
     * loop block
     *         .---------------------.
     *         | prev_bb + init_stmt |
     *         '---------------------'
     *                    |
     *               .---------.
     *               | loop_bb |<--------.
     *               '---------'         |
     *                    |              |
     *               .----------.        |
     *               | loop blk |        |
     *               '----------'        |
     *               /          \        |
     *        .----------.  .---------.  |
     *        | break_bb |  | cont_bb |--'
     *        '----------'  '---------'
     */

    trans->loop_bb = loop_bb;
    trans->cont_bb = post_bb;
    trans->break_bb = next_bb;

    blk_trans(trans, stmt->u_loop.blk);

    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, post_bb);

        fn_add_basic_blk(trans->fn, trans->bb);
    }

    trans->bb = post_bb;

    if (stmt->u_loop.post_stmt != NULL)
        stmt_trans(trans, stmt->u_loop.post_stmt);

    /* Make loop using post block and loop block */
    bb_add_branch(post_bb, NULL, loop_bb);

    fn_add_basic_blk(trans->fn, post_bb);

    trans->loop_bb = NULL;
    trans->cont_bb = NULL;
    trans->break_bb = NULL;

    trans->bb = next_bb;
}

static void
stmt_trans_switch(trans_t *trans, ast_stmt_t *stmt)
{
    int i;
    ast_blk_t *blk = stmt->u_sw.blk;
    ir_bb_t *prev_bb = trans->bb;
    ir_bb_t *next_bb = bb_new();

    /* In a switch-case statement, each case block is transformed to a single basic block, and the
     * switch condition and the case value are compared and used as a branch condition.
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

    vector_foreach(&blk->stmts, i) {
        ast_stmt_t *case_stmt = vector_get_stmt(&blk->stmts, i);

        /* The case statement means the start of a case block or default block, and the remaining
         * statements are included in the corresponding block */
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
    ir_fn_t *fn = trans->fn;

    ASSERT(fn != NULL);

    if (stmt->u_ret.arg_exp != NULL) {
        ASSERT(stmt->u_ret.ret_id != NULL);
        ASSERT(!is_ctor_id(stmt->u_ret.ret_id->up));

        exp_trans(trans, stmt->u_ret.arg_exp);

        bb_add_stmt(trans->bb, stmt);
    }
    else {
        bb_add_branch(trans->bb, NULL, fn->exit_bb);
    }

    fn_add_basic_blk(fn, trans->bb);

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

static void
stmt_trans_pragma(trans_t *trans, ast_stmt_t *stmt)
{
    switch (stmt->u_pragma.kind) {
    case PRAGMA_ASSERT:
        exp_trans(trans, stmt->u_pragma.val_exp);

        if (stmt->u_pragma.desc_exp != NULL)
            exp_trans(trans, stmt->u_pragma.desc_exp);
        break;

    default:
        ASSERT1(!"invalid pragma", stmt->u_pragma.kind);
    }

    bb_add_stmt(trans->bb, stmt);
}

void
stmt_trans(trans_t *trans, ast_stmt_t *stmt)
{
    ASSERT(stmt != NULL);

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

    case STMT_ID:
        stmt_trans_id(trans, stmt);
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

    case STMT_PRAGMA:
        stmt_trans_pragma(trans, stmt);
        break;

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }
}

/* end of trans_stmt.c */
