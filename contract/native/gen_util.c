/**
 * @file    gen_util.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_fn.h"
#include "ir_sgmt.h"

#include "gen_util.h"

BinaryenType
type_gen(type_t type)
{
    switch (type) {
    case TYPE_NONE:
    case TYPE_VOID:
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
    case TYPE_OBJECT:
        return BinaryenTypeInt32();

    case TYPE_TUPLE:
    default:
        ASSERT1(!"invalid type", type);
    }

    return BinaryenTypeUnreachable();
}

void
malloc_gen(gen_t *gen)
{
    BinaryenType type = BinaryenTypeInt32();
    BinaryenType params[] = { BinaryenTypeInt32() };
    BinaryenType locals[] = { BinaryenTypeInt32() };
    BinaryenFunctionTypeRef spec;
    BinaryenExpressionRef instrs[3];
    BinaryenModuleRef module = gen->module;

    BinaryenAddGlobal(module, "heap$offset", type, 1, i32_gen(gen, gen->flag.stack_size));

    spec = BinaryenAddFunctionType(module, "system$alloc", type, params, 1);

    instrs[0] = BinaryenSetLocal(module, 1,
                                 BinaryenGetGlobal(module, "heap$offset", type));

    instrs[1] = BinaryenSetGlobal(module, "heap$offset",
                                  BinaryenBinary(module, BinaryenAddInt32(),
                                                 BinaryenGetLocal(module, 1, type),
                                                 BinaryenGetLocal(module, 0, type)));

    instrs[2] = BinaryenReturn(module, BinaryenGetLocal(module, 1, type));

    BinaryenAddFunction(module, "system$alloc", spec, locals, 1,
                        BinaryenBlock(module, NULL, instrs, 3, BinaryenTypeInt32()));
}

/* end of gen_util.c */
