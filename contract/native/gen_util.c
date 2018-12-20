/**
 * @file    gen_util.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_stmt.h"

#include "gen_util.h"

uint32_t
gen_add_local(gen_t *gen, meta_t *meta)
{
    if (gen->locals == NULL)
        gen->locals = xmalloc(sizeof(BinaryenType));
    else
        gen->locals = xrealloc(gen->locals, sizeof(BinaryenType) * (gen->local_cnt + 1));

    gen->locals[gen->local_cnt++] = meta_gen(gen, meta);

    return gen->id_idx++;
}

void
gen_add_instr(gen_t *gen, BinaryenExpressionRef instr)
{
    if (instr == NULL)
        return;

    if (gen->instrs == NULL)
        gen->instrs = xmalloc(sizeof(BinaryenExpressionRef));
    else
        gen->instrs = xrealloc(gen->instrs,
                               sizeof(BinaryenExpressionRef) * (gen->instr_cnt + 1));

    gen->instrs[gen->instr_cnt++] = instr;
}

void
gen_stmt_array(gen_t *gen, array_t *stmts)
{
    int i;

    for (i = 0; i < array_size(stmts); i++) {
        gen_add_instr(gen, stmt_gen(gen, array_get(stmts, i, ast_stmt_t)));
    }
}

BinaryenType
type_gen(gen_t *gen, type_t type)
{
    switch (type) {
    case TYPE_NONE:
    case TYPE_VOID:
    case TYPE_OBJECT:
        return BinaryenTypeNone();

    case TYPE_BOOL:
    case TYPE_BYTE:
    case TYPE_INT8:
    case TYPE_INT16:
    case TYPE_INT32:
    case TYPE_UINT8:
    case TYPE_UINT16:
    case TYPE_UINT32:
        return BinaryenTypeInt32();

    case TYPE_INT64:
    case TYPE_UINT64:
        return BinaryenTypeInt64();

    case TYPE_FLOAT:
        return BinaryenTypeFloat32();

    case TYPE_DOUBLE:
        return BinaryenTypeFloat64();

    case TYPE_STRING:
    case TYPE_ACCOUNT:
    case TYPE_STRUCT:
    case TYPE_MAP:
        return BinaryenTypeInt32();

    case TYPE_TUPLE:
    default:
        ASSERT1(!"invalid type", type);
    }

    return BinaryenTypeUnreachable();
}

/* end of gen_util.c */
