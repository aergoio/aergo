/**
 * @file    ast_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_exp.h"

static ast_exp_t *
ast_exp_new(exp_kind_t kind, src_pos_t *pos)
{
    ast_exp_t *exp = xcalloc(sizeof(ast_exp_t));

    ast_node_init(exp, pos);

    exp->kind = kind;

    meta_init(&exp->meta, &exp->pos);

    return exp;
}

ast_exp_t *
exp_new_null(src_pos_t *pos)
{
    return ast_exp_new(EXP_NULL, pos);
}

ast_exp_t *
exp_new_lit(src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_LIT, pos);

    value_init(&exp->u_lit.val);

    return exp;
}

ast_exp_t *
exp_new_id(char *name, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_ID, pos);

    exp->u_id.name = name;

    return exp;
}

ast_exp_t *
exp_new_array(ast_exp_t *id_exp, ast_exp_t *idx_exp, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_ARRAY, pos);

    exp->u_arr.id_exp = id_exp;
    exp->u_arr.idx_exp = idx_exp;

    return exp;
}

ast_exp_t *
exp_new_cast(type_t type, ast_exp_t *val_exp, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_CAST, pos);

    exp->u_cast.val_exp = val_exp;

    meta_set(&exp->u_cast.to_meta, type);

    return exp;
}

ast_exp_t *
exp_new_call(ast_exp_t *id_exp, array_t *param_exps, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_CALL, pos);

    exp->u_call.id_exp = id_exp;
    exp->u_call.param_exps = param_exps;

    return exp;
}

ast_exp_t *
exp_new_access(ast_exp_t *id_exp, ast_exp_t *fld_exp, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_ACCESS, pos);

    exp->u_acc.id_exp = id_exp;
    exp->u_acc.fld_exp = fld_exp;

    return exp;
}

ast_exp_t *
exp_new_unary(op_kind_t kind, bool is_prefix, ast_exp_t *val_exp, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_UNARY, pos);

    exp->u_un.kind = kind;
    exp->u_un.is_prefix = is_prefix;
    exp->u_un.val_exp = val_exp;

    return exp;
}

ast_exp_t *
exp_new_binary(op_kind_t kind, ast_exp_t *l_exp, ast_exp_t *r_exp, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_BINARY, pos);

    exp->u_bin.kind = kind;
    exp->u_bin.l_exp = l_exp;
    exp->u_bin.r_exp = r_exp;

    return exp;
}

ast_exp_t *
exp_new_ternary(ast_exp_t *pre_exp, ast_exp_t *in_exp, ast_exp_t *post_exp,
                src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_TERNARY, pos);

    exp->u_tern.pre_exp = pre_exp;
    exp->u_tern.in_exp = in_exp;
    exp->u_tern.post_exp = post_exp;

    return exp;
}

ast_exp_t *
exp_new_sql(sql_kind_t kind, char *sql, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_SQL, pos);

    exp->u_sql.kind = kind;
    exp->u_sql.sql = sql;

    return exp;
}

ast_exp_t *
exp_new_tuple(array_t *elem_exps, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_TUPLE, pos);

    exp->u_tup.elem_exps = elem_exps;

    return exp;
}

ast_exp_t *
exp_new_init(array_t *elem_exps, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_INIT, pos);

    if (elem_exps == NULL)
        exp->u_init.elem_exps = array_new();
    else
        exp->u_init.elem_exps = elem_exps;

    return exp;
}

ast_exp_t *
exp_new_global(char *name)
{
    ast_exp_t *exp = ast_exp_new(EXP_GLOBAL, &null_src_pos_);

    ASSERT(name != NULL);

    exp->u_glob.name = name;

    return exp;
}

ast_exp_t *
exp_new_local(int idx)
{
    ast_exp_t *exp = ast_exp_new(EXP_LOCAL, &null_src_pos_);

    ASSERT(idx >= 0);

    exp->u_local.idx = idx;

    return exp;
}

ast_exp_t *
exp_new_stack(int addr, int offset)
{
    ast_exp_t *exp = ast_exp_new(EXP_STACK, &null_src_pos_);

    ASSERT(addr >= 0);
    ASSERT(offset >= 0);

    exp->u_stk.addr = addr;
    exp->u_stk.offset = offset;

    return exp;
}

void
exp_set_lit(ast_exp_t *exp, value_t *val)
{
    exp->kind = EXP_LIT;

    if (val != NULL)
        exp->u_lit.val = *val;
    else
        value_init(&exp->u_lit.val);
}

void
exp_set_local(ast_exp_t *exp, int idx)
{
    ASSERT(idx >= 0);

    exp->kind = EXP_LOCAL;
    exp->u_local.idx = idx;
}

void
exp_set_stack(ast_exp_t *exp, int addr, int offset)
{
    ASSERT(addr >= 0);
    ASSERT(offset >= 0);

    exp->kind = EXP_STACK;
    exp->u_stk.addr = addr;
    exp->u_stk.offset = offset;
}

ast_exp_t *
exp_clone(ast_exp_t *exp)
{
    int i;
    ast_exp_t *res = NULL;
    array_t *elem_exps;
    array_t *res_exps;

    if (exp == NULL)
        return NULL;

    switch (exp->kind) {
    case EXP_NULL:
        res = exp_new_null(&exp->pos);
        break;

    case EXP_ID:
        res = exp_new_id(exp->u_id.name, &exp->pos);
        break;

    case EXP_LIT:
        res = exp_new_lit(&exp->pos);
        res->u_lit.val = exp->u_lit.val;
        break;

    case EXP_ARRAY:
        res = exp_new_array(exp_clone(exp->u_arr.id_exp), exp_clone(exp->u_arr.idx_exp),
                            &exp->pos);
        break;

    case EXP_CAST:
        res = exp_new_cast(exp->u_cast.to_meta.type, exp_clone(exp->u_cast.val_exp),
                           &exp->pos);
        break;

    case EXP_UNARY:
        res = exp_new_unary(exp->u_un.kind, exp->u_un.is_prefix,
                            exp_clone(exp->u_un.val_exp), &exp->pos);
        break;

    case EXP_BINARY:
        res = exp_new_binary(exp->u_bin.kind, exp_clone(exp->u_bin.l_exp),
                             exp_clone(exp->u_bin.r_exp), &exp->pos);
        break;

    case EXP_TERNARY:
        res = exp_new_ternary(exp_clone(exp->u_tern.pre_exp),
                              exp_clone(exp->u_tern.in_exp),
                              exp_clone(exp->u_tern.post_exp), &exp->pos);
        break;

    case EXP_ACCESS:
        res = exp_new_access(exp_clone(exp->u_acc.id_exp),
                             exp_clone(exp->u_acc.fld_exp), &exp->pos);
        break;

    case EXP_CALL:
        elem_exps = exp->u_call.param_exps;
        res_exps = array_new();
        array_foreach(elem_exps, i) {
            array_add_last(res_exps, exp_clone(array_get_exp(elem_exps, i)));
        }
        res = exp_new_call(exp_clone(exp->u_call.id_exp), res_exps, &exp->pos);
        break;

    case EXP_SQL:
        res = exp_new_sql(exp->u_sql.kind, exp->u_sql.sql, &exp->pos);
        break;

    case EXP_TUPLE:
        elem_exps = exp->u_tup.elem_exps;
        res_exps = array_new();
        array_foreach(elem_exps, i) {
            array_add_last(res_exps, exp_clone(array_get_exp(elem_exps, i)));
        }
        res = exp_new_tuple(res_exps, &exp->pos);
        break;

    case EXP_GLOBAL:
        res = exp_new_global(exp->u_glob.name);
        break;

    case EXP_LOCAL:
        res = exp_new_local(exp->u_local.idx);
        break;

    case EXP_STACK:
        res = exp_new_stack(exp->u_stk.addr, exp->u_stk.offset);
        break;

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    res->id = exp->id;
    meta_copy(&res->meta, &exp->meta);

    return res;
}

bool
exp_equals(ast_exp_t *x, ast_exp_t *y)
{
    int i;

    if (x == NULL && y == NULL)
        return true;

    if (x == NULL || y == NULL || x->kind != y->kind)
        return false;

    switch (x->kind) {
    case EXP_NULL:
        return true;

    case EXP_ID:
        return strcmp(x->u_id.name, y->u_id.name) == 0;

    case EXP_LIT:
        return x->u_lit.val.type == y->u_lit.val.type &&
            value_cmp(&x->u_lit.val, &y->u_lit.val) == 0;

    case EXP_ARRAY:
        return exp_equals(x->u_arr.id_exp, y->u_arr.id_exp) &&
            exp_equals(x->u_arr.idx_exp, y->u_arr.idx_exp);

    case EXP_CAST:
        return x->u_cast.to_meta.type == y->u_cast.to_meta.type &&
            exp_equals(x->u_cast.val_exp, y->u_cast.val_exp);

    case EXP_UNARY:
        return x->u_un.kind == y->u_un.kind &&
            exp_equals(x->u_un.val_exp, y->u_un.val_exp);

    case EXP_BINARY:
        return x->u_bin.kind == y->u_bin.kind &&
            exp_equals(x->u_bin.l_exp, y->u_bin.l_exp) &&
            exp_equals(x->u_bin.r_exp, y->u_bin.r_exp);

    case EXP_TERNARY:
        return exp_equals(x->u_tern.pre_exp, y->u_tern.pre_exp) &&
            exp_equals(x->u_tern.in_exp, y->u_tern.in_exp) &&
            exp_equals(x->u_tern.post_exp, y->u_tern.post_exp);

    case EXP_ACCESS:
        return exp_equals(x->u_acc.id_exp, y->u_acc.id_exp) &&
            exp_equals(x->u_acc.fld_exp, y->u_acc.fld_exp);

    case EXP_CALL:
        if (array_size(x->u_call.param_exps) != array_size(y->u_call.param_exps))
            return false;

        array_foreach(x->u_call.param_exps, i) {
            if (!exp_equals(array_get_exp(x->u_call.param_exps, i),
                            array_get_exp(y->u_call.param_exps, i)))
                return false;
        }
        return exp_equals(x->u_acc.id_exp, y->u_acc.id_exp);

    case EXP_SQL:
        return x->u_sql.kind == y->u_sql.kind &&
            strcmp(x->u_sql.sql, y->u_sql.sql) == 0;

    case EXP_TUPLE:
        if (array_size(x->u_tup.elem_exps) != array_size(y->u_tup.elem_exps))
            return false;

        array_foreach(x->u_tup.elem_exps, i) {
            if (!exp_equals(array_get_exp(x->u_tup.elem_exps, i),
                            array_get_exp(y->u_tup.elem_exps, i)))
                return false;
        }
        return true;

    default:
        ASSERT1(!"invalid expression", x->kind);
    }

    return false;
}

/* end of ast_exp.c */
