/**
 * @file    gen_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_exp.h"

static BinaryenExpressionRef
exp_gen_id(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_val(gen_t *gen, ast_exp_t *exp)
{
    ASSERT1(is_lit_exp(exp), exp->kind);

    switch (exp->u_lit.val.kind) {
    case VAL_BOOL:
        return BinaryenConst(gen->module, BinaryenLiteralInt32(bool_val(&exp->u_lit.val)));
    case VAL_INT:
        return BinaryenConst(gen->module, BinaryenLiteralInt64(int_val(&exp->u_lit.val)));
    case VAL_FP:
        return BinaryenConst(gen->module, BinaryenLiteralFloat64(fp_val(&exp->u_lit.val)));
    case VAL_STR:
        // XXX
        break;
    case VAL_OBJ:
        // XXX
        break;
    default:
        ASSERT1(!"invalid value", exp->u_lit.val.kind);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_array(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_cast(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_arith(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_bit(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_cmp(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_unary(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_bool_cmp(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op(gen_t *gen, ast_exp_t *exp)
{
    ASSERT1(is_op_exp(exp), exp->kind);

    switch (exp->u_op.kind) {
    case OP_ADD:
    case OP_SUB:
    case OP_MUL:
    case OP_DIV:
    case OP_MOD:
        return exp_gen_op_arith(gen, exp);

    case OP_BIT_AND:
    case OP_BIT_OR:
    case OP_BIT_XOR:
    case OP_RSHIFT:
    case OP_LSHIFT:
        return exp_gen_op_bit(gen, exp);

    case OP_EQ:
    case OP_NE:
    case OP_LT:
    case OP_GT:
    case OP_LE:
    case OP_GE:
        return exp_gen_op_cmp(gen, exp);

    case OP_INC:
    case OP_DEC:
    case OP_NOT:
    case OP_NEG:
        return exp_gen_op_unary(gen, exp);

    case OP_AND:
    case OP_OR:
        return exp_gen_op_bool_cmp(gen, exp);

    default:
        ASSERT1(!"invalid operator", exp->u_op.kind);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_access(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_call(gen_t *gen, ast_exp_t *exp)
{
    return BinaryenNop(gen->module);
}

static BinaryenExpressionRef
exp_gen_sql(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_ternary(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_tuple(gen_t *gen, ast_exp_t *exp)
{
    return NULL;
}

BinaryenExpressionRef
exp_gen(gen_t *gen, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        return NULL;

    case EXP_ID:
        return exp_gen_id(gen, exp);

    case EXP_LIT:
        return exp_gen_val(gen, exp);

    case EXP_ARRAY:
        return exp_gen_array(gen, exp);

    case EXP_CAST:
        return exp_gen_cast(gen, exp);

    case EXP_OP:
        return exp_gen_op(gen, exp);

    case EXP_ACCESS:
        return exp_gen_access(gen, exp);

    case EXP_CALL:
        return exp_gen_call(gen, exp);

    case EXP_SQL:
        return exp_gen_sql(gen, exp);

    case EXP_TERNARY:
        return exp_gen_ternary(gen, exp);

    case EXP_TUPLE:
        return exp_gen_tuple(gen, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

/* end of gen_exp.c */
