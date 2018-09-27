/**
 * @file    check.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"
#include "ast_var.h"
#include "ast_exp.h"
#include "ast_struct.h"

#include "check.h"

static void
check_init(check_t *cenv, ast_t *ast)
{
    ASSERT(ast != NULL);

    cenv->ast = ast;
    cenv->blk = NULL;
}

static void
check_type_exp_to_prim(check_t *cenv, ast_exp_t *exp)
{
    ASSERT(exp->u_type.name == NULL);
    ASSERT(exp->u_type.k_exp == NULL);
    ASSERT(exp->u_type.v_exp == NULL);
}

static void
check_type_exp_to_struct(check_t *cenv, ast_exp_t *exp)
{
    ast_struct_t *struc;

    ASSERT(exp->u_type.name != NULL);
    ASSERT(exp->u_type.k_exp == NULL);
    ASSERT(exp->u_type.v_exp == NULL);

    struc = ast_blk_search_struct(cenv->blk, exp->num, exp->u_type.name);
    if (struc == NULL)
        TRACE(ERROR_UNDEFINED_TYPE, &exp->pos, exp->u_type.name);
}

static void
exp_type_check(check_t *cenv, ast_exp_t *exp)
{
    ASSERT(exp->kind == EXP_TYPE);
    ASSERT(exp->u_type.type > TYPE_NONE);
    ASSERT(exp->u_type.type < TYPE_MAX);

    if (type_is_primitive(exp->u_type.type)) {
        check_type_exp_to_prim(cenv, exp);
    }
    else if (type_is_struct(exp->u_type.type)) {
        check_type_exp_to_struct(cenv, exp);
    }
    else if (type_is_map(exp->u_type.type)) {
        ASSERT(exp->u_type.name == NULL);
        ASSERT(exp->u_type.k_exp != NULL);
        ASSERT(exp->u_type.v_exp != NULL);

        //ast_exp_check(cenv, exp->u_type.k_exp);
    }
}

static void
ast_exp_check(check_t *cenv, ast_exp_t *exp)
{
    ASSERT(exp->kind < EXP_MAX);

    switch (exp->kind) {
    case EXP_ID:
        break;
    case EXP_LIT:
        break;
    case EXP_TYPE:
        exp_type_check(cenv, exp);
        break;
    case EXP_ARRAY:
        break;
    case EXP_OP:
        break;
    case EXP_ACCESS:
        break;
    case EXP_CALL:
        break;
    case EXP_SQL:
        break;
    case EXP_COND:
        break;
    case EXP_TUPLE:
        break;
    default:
        ASSERT1(!"invalid kind", exp->kind);
    }
}

static void
ast_var_check(check_t *cenv, ast_var_t *var)
{
    ASSERT(var->type_exp != NULL);
    ASSERT(var->id_exp != NULL);

    ast_exp_check(cenv, var->type_exp);
    ast_exp_check(cenv, var->id_exp);

    if (var->init_exp != NULL)
        ast_exp_check(cenv, var->init_exp);
}

static void
ast_blk_check(check_t *cenv, ast_blk_t *blk)
{
    list_node_t *node;

    ASSERT(blk != NULL);

    blk->up = cenv->blk;
    cenv->blk = blk;

    list_foreach(node, &blk->var_l) {
        ast_var_check(cenv, (ast_var_t *)node->item);
    }

    cenv->blk = blk->up;
}

void
check(ast_t *ast, flag_t flag)
{
    check_t cenv;
    list_node_t *node;

    check_init(&cenv, ast);

    list_foreach(node, &ast->blk_l) {
        ast_blk_check(&cenv, (ast_blk_t *)node->item);
    }
}

/* end of check.c */
