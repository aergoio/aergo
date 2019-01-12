/**
 * @file    trans_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"
#include "ir_abi.h"
#include "ir_fn.h"
#include "ir_bb.h"
#include "trans_blk.h"
#include "trans_stmt.h"
#include "trans_exp.h"

#include "trans_id.h"

static void
id_trans_var(trans_t *trans, ast_id_t *id)
{
    if (is_global_id(id)) {
        /* Initialization of the global variable will be done in the constructor */
        ir_add_global(trans->ir, id);
        return;
    }

    ASSERT(trans->fn != NULL);

    if (is_stack_id(id))
        fn_add_stack(trans->fn, id);
    else
        fn_add_local(trans->fn, id);
}

static void
gen_init_stmt(array_t *stmts, ast_id_t *id)
{
    meta_t *meta = &id->meta;
    ast_exp_t *dflt_exp = id->u_var.dflt_exp;
    ast_exp_t *id_exp;

    if (is_const_id(id))
        return;

    ASSERT(id->up != NULL);
    ASSERT1(is_global_id(id), id->up->kind);

    if (dflt_exp == NULL) {
        dflt_exp = exp_new_lit(meta->pos);

        if (is_array_type(meta)) {
            value_set_ptr(&dflt_exp->u_lit.val, xcalloc(meta->arr_size), meta->arr_size);
        }
        else if (is_bool_type(meta)) {
            value_set_bool(&dflt_exp->u_lit.val, false);
        }
        else if (is_fpoint_type(meta)) {
            value_set_f64(&dflt_exp->u_lit.val, 0.0);
        }
        else if (is_integer_type(meta) || is_pointer_type(meta)) {
            value_set_i64(&dflt_exp->u_lit.val, 0);
        }
        else {
            ASSERT1(is_struct_type(meta), meta->type);
            value_set_ptr(&dflt_exp->u_lit.val, xcalloc(meta_size(meta)),
                          meta_size(meta));
        }

        meta_copy(&dflt_exp->meta, meta);
        meta_set_undef(&dflt_exp->meta);
    }

    id_exp = exp_new_id(id->name, &dflt_exp->pos);

    id_exp->id = id;
    meta_copy(&id_exp->meta, meta);

    ASSERT2(meta_cmp(&id_exp->meta, &dflt_exp->meta), id_exp->meta.type,
            dflt_exp->meta.type);

    stmt_add(stmts, stmt_new_assign(id_exp, dflt_exp, &dflt_exp->pos));
}

static void
id_trans_ctor(trans_t *trans, ast_id_t *id)
{
    int i, j;
    ast_id_t *addr_id;
    ast_exp_t *l_exp, *r_exp;
    ast_stmt_t *ret_stmt;
    array_t *stmts = array_new();
    ir_fn_t *fn = trans->fn;

    ASSERT(id->u_fn.ret_id != NULL);
    ASSERT1(is_return_id(id->u_fn.ret_id), id->u_fn.ret_id->kind);

    // 모든 전역 변수는 이 값을 기준으로 relative offset으로 접근한다.
    addr_id = id_new_tmp_var("cont$addr", TYPE_INT32);
    fn_add_local(fn, addr_id);

    l_exp = exp_new_local(TYPE_INT32, addr_id->idx);
    r_exp = exp_new_global(TYPE_INT32, "heap$offset");

    stmt_add(stmts, stmt_new_assign(l_exp, r_exp, &id->pos));

    fn->heap_id = addr_id;

    /* constructor initializes global variables */
    // 변수의 선언 순서를 지키기 위해 별도의 array_t에 담아서 join한다.
    array_foreach(&id->up->u_cont.blk->ids, i) {
        ast_id_t *var_id = array_get_id(&id->up->u_cont.blk->ids, i);

        if (is_var_id(var_id)) {
            gen_init_stmt(stmts, var_id);
        }
        else if (is_tuple_id(var_id)) {
            array_foreach(var_id->u_tup.elem_ids, j) {
                gen_init_stmt(stmts, array_get_id(var_id->u_tup.elem_ids, j));
            }
        }
    }

    array_join_first(&id->u_fn.blk->stmts, stmts);

    ret_stmt = stmt_new_return(l_exp, &id->pos);
    ret_stmt->u_ret.ret_id = id->u_fn.ret_id;

    stmt_add(&id->u_fn.blk->stmts, ret_stmt);
}

static void
id_trans_fn(trans_t *trans, ast_id_t *id)
{
    ast_id_t *addr_id;
    ast_exp_t *l_exp, *r_exp, *v_exp;
    ir_fn_t *fn = fn_new(id);

    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);

    trans->fn = fn;

    if (id->u_fn.blk == NULL)
        id->u_fn.blk = blk_new_fn(&id->pos);

    // binaryen 내부 용도로 사용
    fn_add_local(fn, id_new_tmp_var("relooper$helper", TYPE_INT32));

    // 모든 스택 변수는 이 값을 기준으로 relative offset으로 접근한다.
    addr_id = id_new_tmp_var("stack$addr", TYPE_INT32);

    fn_add_local(fn, addr_id);
    fn->stack_id = addr_id;

    if (is_ctor_id(id)) {
        id_trans_ctor(trans, id);
    }
    else {
        // constructor가 아닌 함수들은 모두 contract base address를 인자로 추가하고,
        // abi에도 반영하기 위해 param_ids에 추가한다. */
        addr_id = id_new_tmp_var("cont$addr", TYPE_OBJECT);

        addr_id->is_param = true;
        addr_id->up = id;

        meta_set_object(&addr_id->meta, id->up);

        if (id->u_fn.param_ids == NULL)
            id->u_fn.param_ids = array_new();

        array_add_first(id->u_fn.param_ids, addr_id);

        fn->heap_id = addr_id;
    }

    fn->abi = abi_lookup(&trans->ir->abis, id);

    trans->bb = fn->entry_bb;

    blk_trans(trans, id->u_fn.blk);

    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, fn->exit_bb);
        fn_add_basic_blk(fn, trans->bb);
    }

    fn_add_basic_blk(fn, fn->exit_bb);

    // 현재의 stack 사용량을 이용하여 stack address를 설정한다.
    l_exp = exp_new_local(TYPE_INT32, addr_id->idx);

    v_exp = exp_new_lit(&id->pos);
    value_set_i64(&v_exp->u_lit.val, ALIGN64(fn->usage));

    r_exp = exp_new_binary(OP_SUB, exp_new_global(TYPE_INT32, "stack$high"), v_exp,
                           &id->pos);

    array_add_first(&fn->entry_bb->stmts, stmt_new_assign(l_exp, r_exp, &id->pos));

    trans->fn = NULL;
    trans->bb = NULL;

    ir_add_fn(trans->ir, fn);
}

static void
id_trans_contract(trans_t *trans, ast_id_t *id)
{
    int i, j;
    ir_t *ir = trans->ir;
    ast_blk_t *blk = id->u_cont.blk;

    ASSERT(blk != NULL);
    ASSERT1(ir->offset == 0, ir->offset);

    if (id->u_cont.impl_exp != NULL) {
        /* rearrange functions according to the order in the interface */
        // 함수의 순서를 재조정하는 이유는 함수의 parameter로 interface를 사용하고,
        // 해당 interface를 implement한 contract가 여러개 있는 상황에서 argument로
        // 사용된 contract에 속한 함수의 위치를 알 수 있는 방법이 없기 때문에 함수
        // index를 동일하게 맞춤으로써 interface function의 index를 통해 함수를
        // 호출하기 위해서다.
        ast_id_t *itf_id = id->u_cont.impl_exp->id;

        ASSERT1(is_itf_id(itf_id), itf_id->kind);

        array_foreach(&itf_id->u_itf.blk->ids, i) {
            ast_id_t *spec_id = array_get_id(&itf_id->u_itf.blk->ids, i);

            array_foreach(&blk->ids, j) {
                ast_id_t *fn_id = array_get_id(&blk->ids, j);

                if (is_fn_id(fn_id) && strcmp(spec_id->name, fn_id->name) == 0) {
                    array_move(&blk->ids, j, i);
                    break;
                }
            }
        }
    }

    /* 이 값은 함수 인자가 interface이고, argument로 contract 변수가 전달될때 사용된다 */
    id->idx = array_size(&ir->fns);

    blk_trans(trans, id->u_cont.blk);

    ir->offset = 0;
}

static void
id_trans_label(trans_t *trans, ast_id_t *id)
{
    id->u_lab.stmt->label_bb = bb_new();
}

static void
id_trans_tuple(trans_t *trans, ast_id_t *id)
{
    int i;

    array_foreach(id->u_tup.elem_ids, i) {
        id_trans_var(trans, array_get_id(id->u_tup.elem_ids, i));
    }
}

void
id_trans(trans_t *trans, ast_id_t *id)
{
    switch (id->kind) {
    case ID_VAR:
        id_trans_var(trans, id);
        break;

    case ID_FN:
        id_trans_fn(trans, id);
        break;

    case ID_CONT:
        id_trans_contract(trans, id);
        break;

    case ID_LABEL:
        id_trans_label(trans, id);
        break;

    case ID_TUPLE:
        id_trans_tuple(trans, id);
        break;

    case ID_STRUCT:
    case ID_ENUM:
    case ID_ITF:
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of trans_id.c */
