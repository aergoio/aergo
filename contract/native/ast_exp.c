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
exp_new_val(src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_VAL, pos);

    value_init(&exp->u_val.val);

    return exp;
}

ast_exp_t *
exp_new_type(type_t type, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_TYPE, pos);

    exp->u_type.type = type;

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
exp_new_op(op_kind_t kind, ast_exp_t *l_exp, ast_exp_t *r_exp, src_pos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_OP, pos);

    exp->u_op.kind = kind;
    exp->u_op.l_exp = l_exp;
    exp->u_op.r_exp = r_exp;

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

void
ast_exp_dump(ast_exp_t *exp, int indent)
{
}

/* end of ast_exp.c */
