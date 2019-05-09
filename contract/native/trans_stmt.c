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

static uint32_t make_array_initz(trans_t *trans, ast_exp_t *var_exp, uint32_t offset, int dim_idx);

void
stmt_trans_malloc(trans_t *trans, uint32_t reg_idx, bool is_heap, meta_t *meta)
{
    ast_exp_t *l_exp, *r_exp;
    ast_exp_t *stk_exp, *addr_exp;

    if (is_heap) {
        stmt_trans(trans, stmt_make_malloc(reg_idx, meta_memsz(meta), meta_align(meta), meta->pos));
        return;
    }

    l_exp = exp_new_reg(reg_idx);
    meta_set_int32(&l_exp->meta);

    fn_add_stack(trans->fn, meta_memsz(meta), meta);

    stk_exp = exp_new_reg(meta->base_idx);
    meta_set_int32(&stk_exp->meta);

    addr_exp = exp_new_lit_int(meta->rel_addr, meta->pos);
    meta_set_int32(&addr_exp->meta);

    r_exp = exp_new_binary(OP_ADD, stk_exp, addr_exp, meta->pos);
    meta_set_int32(&r_exp->meta);

    bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, meta->pos));
}

static uint32_t
make_map_initz(trans_t *trans, ast_exp_t *var_exp)
{
    fn_kind_t kind;
    ast_exp_t *val_exp;
    meta_t *meta = &var_exp->meta;

    ASSERT1(is_map_meta(meta), meta->type);
    ASSERT1(meta->elem_cnt == 2, meta->elem_cnt);

    if (is_int64_meta(meta->elems[0]) && is_int64_meta(meta->elems[1]))
        kind = FN_MAP_NEW_I64_I64;
    else if (is_int64_meta(meta->elems[0]))
        kind = FN_MAP_NEW_I64_I32;
    else if (is_int64_meta(meta->elems[1]))
        kind = FN_MAP_NEW_I32_I64;
    else
        kind = FN_MAP_NEW_I32_I32;

    val_exp = exp_new_call(kind, NULL, NULL, meta->pos);
    meta_set(&val_exp->meta, TYPE_MAP);

    exp_trans(trans, val_exp);

    bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, meta->pos));

    return meta_typsz(meta);
}

static uint32_t
make_struct_initz(trans_t *trans, ast_exp_t *var_exp, uint32_t offset)
{
    int i;
    meta_t *meta = &var_exp->meta;

    ASSERT1(meta->elem_cnt > 0, meta->elem_cnt);

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = meta->elems[i];

        if ((is_array_meta(elem_meta) && is_fixed_array(elem_meta)) ||
            (!is_array_meta(elem_meta) && is_struct_meta(elem_meta))) {
            ast_exp_t *mem_exp;

            mem_exp = exp_new_mem(meta->base_idx, meta->rel_addr, offset);
            meta_copy(&mem_exp->meta, elem_meta);

            if (is_struct_meta(elem_meta))
                make_struct_initz(trans, mem_exp, offset);
            else
                make_array_initz(trans, mem_exp, offset, 0);
        }

        offset += meta_memsz(elem_meta);
    }

    return meta_memsz(meta);
}

static uint32_t
make_array_initz(trans_t *trans, ast_exp_t *var_exp, uint32_t offset, int dim_idx)
{
    int i;
    ast_exp_t *l_exp, *r_exp;
    meta_t *meta = &var_exp->meta;

    ASSERT2(dim_idx < meta->max_dim, dim_idx, meta->max_dim);
    ASSERT2(meta->dim_sizes[dim_idx] > 0, dim_idx, meta->dim_sizes[dim_idx]);
    ASSERT2(meta_align(meta) > 0, meta->type, meta_align(meta));

    /* count of elements */
    l_exp = exp_new_mem(meta->base_idx, meta->rel_addr, offset);
    meta_set(&l_exp->meta, meta_align(meta) == 8 ? TYPE_INT64 : TYPE_INT32);

    r_exp = exp_new_lit_int(meta->dim_sizes[dim_idx], meta->pos);
    meta_copy(&r_exp->meta, &l_exp->meta);

    bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, meta->pos));
    offset += meta_align(meta);

    if (dim_idx == meta->max_dim - 1) {
        if (is_struct_meta(meta) || is_map_meta(meta)) {
            for (i = 0; i < meta->dim_sizes[dim_idx]; i++) {
                ast_exp_t *mem_exp;

                offset = ALIGN(offset, meta_align(meta));

                mem_exp = exp_new_mem(meta->base_idx, meta->rel_addr, offset);
                meta_copy(&mem_exp->meta, meta);

                if (is_struct_meta(meta))
                    offset += make_struct_initz(trans, mem_exp, offset);
                else
                    offset += make_map_initz(trans, mem_exp);
            }
        }
        else {
            offset += ALIGN(meta_typsz(meta), meta_align(meta)) * meta->dim_sizes[dim_idx];
        }
    }
    else {
        for (i = 0; i < meta->dim_sizes[dim_idx]; i++) {
            offset = make_array_initz(trans, var_exp, offset, dim_idx + 1);
        }
    }

    return offset;
}

void
stmt_trans_initz(trans_t *trans, ast_exp_t *var_exp)
{
    ast_exp_t *init_exp;
    meta_t *meta = &var_exp->meta;

    /* Since variables may be defined in the loop, they are explicitly initialized. */

    if (is_array_meta(meta)) {
        if (is_fixed_array(meta)) {
            make_array_initz(trans, var_exp, 0, 0);
        }
        else {
            init_exp = exp_new_lit_null(&var_exp->pos);
            meta_set(&init_exp->meta, TYPE_OBJECT);

            stmt_trans(trans, stmt_new_assign(var_exp, init_exp, &var_exp->pos));
        }
    }
    else if (is_struct_meta(meta)) {
        make_struct_initz(trans, var_exp, 0);
    }
    else if (is_map_meta(meta)) {
        make_map_initz(trans, var_exp);
    }
    else {
        if (is_bool_meta(meta))
            init_exp = exp_new_lit_bool(false, &var_exp->pos);
        else if (is_string_meta(meta) || is_object_meta(meta))
            init_exp = exp_new_lit_null(&var_exp->pos);
        else
            init_exp = exp_new_lit_int(0, &var_exp->pos);

        meta_copy(&init_exp->meta, meta);

        stmt_trans(trans, stmt_new_assign(var_exp, init_exp, &var_exp->pos));
    }
}

static void
make_mem_initz(trans_t *trans, ast_id_t *id)
{
    ast_exp_t *var_exp;
    meta_t *meta = &id->meta;

    if (is_global_id(id)) {
        var_exp = exp_new_mem(meta->base_idx, meta->rel_addr, 0);
        meta_copy(&var_exp->meta, meta);
    }
    else {
        var_exp = exp_new_reg(meta->base_idx);
        meta_copy(&var_exp->meta, meta);

        if ((is_array_meta(meta) && is_fixed_array(meta)) ||
            (!is_array_meta(meta) && is_struct_meta(meta))) {
            ast_exp_t *call_exp;
            ast_exp_t *arg_exp;
            vector_t *arg_exps = vector_new();

            exp_add(arg_exps, var_exp);

            arg_exp = exp_new_lit_int(0, meta->pos);
            meta_set_int32(&arg_exp->meta);
            exp_add(arg_exps, arg_exp);

            arg_exp = exp_new_lit_int(meta_memsz(meta), meta->pos);
            meta_set_int32(&arg_exp->meta);
            exp_add(arg_exps, arg_exp);

            call_exp = exp_new_call(FN_MEMSET, NULL, arg_exps, meta->pos);
            meta_set_void(&call_exp->meta);

            stmt_trans(trans, stmt_new_exp(call_exp, meta->pos));
        }
    }

    stmt_trans_initz(trans, var_exp);
}

static void
make_id_initz(trans_t *trans, ast_id_t *id)
{
    meta_t *meta = &id->meta;
    ast_exp_t *dflt_exp = id->u_var.dflt_exp;

    ASSERT1(is_var_id(id), id->kind);

    /* TODO If initializer is used as the default value of identifier, there is no need to
     *      allocate unnecessary memory. (see make_dynamic_init()) */

    if (!is_global_id(id)) {
        ASSERT(trans->fn != NULL);

        id->idx = fn_add_register(trans->fn, meta);

        if ((is_array_meta(meta) && is_fixed_array(meta)) ||
            (!is_array_meta(meta) && is_struct_meta(meta)))
            stmt_trans_malloc(trans, id->idx, dflt_exp != NULL && is_alloc_exp(dflt_exp), meta);

        meta->base_idx = id->idx;
        meta->rel_addr = 0;
    }

    if (dflt_exp == NULL)
        make_mem_initz(trans, id);
    else
        stmt_trans(trans, stmt_make_assign(id, dflt_exp));
}

static void
stmt_trans_id(trans_t *trans, ast_stmt_t *stmt)
{
    ast_id_t *id = stmt->u_id.id;

    ASSERT(id != NULL);

    if (is_var_id(id)) {
        make_id_initz(trans, id);
    }
    else if (is_tuple_id(id)) {
        int i;

        vector_foreach(id->u_tup.elem_ids, i) {
            make_id_initz(trans, vector_get_id(id->u_tup.elem_ids, i));
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
make_assign(trans_t *trans, ast_exp_t *var_exp, ast_exp_t *val_exp, src_pos_t *pos)
{
    ast_id_t *var_id = var_exp->id;
    meta_t *var_meta = &var_exp->meta;

    /* We override the meta of the variable declared in the form "interface variable = rvalue;"
     * with the actual contract meta */
    if (val_exp->id != NULL && is_object_meta(var_meta) && is_itf_id(var_meta->type_id)) {
        if (is_fn_id(val_exp->id))
            meta_set_object(var_meta, val_exp->id->up);
        else
            meta_set_object(var_meta, val_exp->id->meta.type_id);
    }

    if (var_id != NULL && is_global_id(var_id)) {
        trans->is_global = true;
        exp_trans(trans, val_exp);
        trans->is_global = false;
    }
    else {
        exp_trans(trans, val_exp);
    }

    if ((var_id != NULL &&
         /* when assigning to value of map */
         ((is_map_meta(&var_id->meta) && !is_map_meta(var_meta)) ||
         /* when assigning to byte value of string */
          (is_string_meta(&var_id->meta) && is_byte_meta(var_meta)))) ||
         /* when assigning to neither fixed-length arrays nor struct variable */
        ((!is_array_meta(var_meta) || !is_fixed_array(var_meta)) &&
         (is_array_meta(var_meta) || !is_struct_meta(var_meta))))
        bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, pos));
    else
        stmt_trans_memcpy(trans, var_exp, val_exp, meta_memsz(var_meta), pos);
}

static void
stmt_trans_assign(trans_t *trans, ast_stmt_t *stmt)
{
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_exp_t *r_exp = stmt->u_assign.r_exp;

    /* We should not use memcpy here because the left operand of an assignment may be a map
     * variable with a struct as its value type. In this case, we must create it with map_put(). */

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

            make_assign(trans, var_exp, val_exp, &stmt->pos);
        }
    }
    else {
        ASSERT(!is_tuple_exp(r_exp));

        make_assign(trans, l_exp, r_exp, &stmt->pos);
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

                resolve_meta(trans, var_exp, val_exp);
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

                    resolve_meta(trans, var_exp, elem_exp);
                    bb_add_stmt(trans->bb, stmt_new_assign(var_exp, elem_exp, pos));
                }
            }
            else {
                var_exp = vector_get_exp(var_exps, var_idx++);

                resolve_meta(trans, var_exp, val_exp);
                bb_add_stmt(trans->bb, stmt_new_assign(var_exp, val_exp, pos));
            }
        }
    }
    else {
        ASSERT(!is_tuple_exp(l_exp));

        resolve_meta(trans, l_exp, r_exp);
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
