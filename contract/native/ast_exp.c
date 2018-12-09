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
exp_new_ref(char *name, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_REF, pos);

    exp->u_ref.name = name;

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
exp_new_unary(op_kind_t kind, ast_exp_t *val_exp, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_UNARY, pos);

    exp->u_un.kind = kind;
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
exp_new_tuple(array_t *exps, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_TUPLE, pos);

    if (exps == NULL)
        exp->u_tup.exps = array_new();
    else
        exp->u_tup.exps = exps;

    return exp;
}

ast_exp_t *
exp_new_init(array_t *exps, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_INIT, pos);

    if (exps == NULL)
        exp->u_init.exps = array_new();
    else
        exp->u_init.exps = exps;

    return exp;
}

ast_exp_t *
exp_clone(ast_exp_t *exp)
{
    int i;
    ast_exp_t *res;
    array_t *exps;
    array_t *res_exps;

    if (exp == NULL)
        return NULL;

    switch (exp->kind) {
    case EXP_NULL:
        return exp_new_null(&exp->pos);

    case EXP_REF:
        return exp_new_ref(exp->u_ref.name, &exp->pos);

    case EXP_LIT:
        res = exp_new_lit(&exp->pos);
        res->u_lit.val = exp->u_lit.val;
        return res;

    case EXP_ARRAY:
        return exp_new_array(exp_clone(exp->u_arr.id_exp), exp_clone(exp->u_arr.idx_exp),
                             &exp->pos);

    case EXP_CAST:
        return exp_new_cast(exp->u_cast.to_meta.type, exp_clone(exp->u_cast.val_exp),
                            &exp->pos);

    case EXP_UNARY:
        return exp_new_unary(exp->u_un.kind, exp_clone(exp->u_un.val_exp), &exp->pos);

    case EXP_BINARY:
        return exp_new_binary(exp->u_bin.kind, exp_clone(exp->u_bin.l_exp),
                              exp_clone(exp->u_bin.r_exp), &exp->pos);

    case EXP_TERNARY:
        return exp_new_ternary(exp_clone(exp->u_tern.pre_exp),
                               exp_clone(exp->u_tern.in_exp),
                               exp_clone(exp->u_tern.post_exp), &exp->pos);

    case EXP_ACCESS:
        return exp_new_access(exp_clone(exp->u_acc.id_exp),
                              exp_clone(exp->u_acc.fld_exp), &exp->pos);

    case EXP_CALL:
        exps = exp->u_call.param_exps;
        res_exps = array_new();
        for (i = 0; i < array_size(exps); i++) {
            array_add_last(res_exps, exp_clone(array_get(exps, i, ast_exp_t)));
        }
        return exp_new_call(exp_clone(exp->u_call.id_exp), res_exps, &exp->pos);

    case EXP_SQL:
        return exp_new_sql(exp->u_sql.kind, exp->u_sql.sql, &exp->pos);

    case EXP_TUPLE:
        exps = exp->u_tup.exps;
        res_exps = array_new();
        for (i = 0; i < array_size(exps); i++) {
            array_add_last(res_exps, exp_clone(array_get(exps, i, ast_exp_t)));
        }
        return exp_new_tuple(res_exps, &exp->pos);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

void
ast_exp_dump(ast_exp_t *exp, int indent)
{
}

/* end of ast_exp.c */
