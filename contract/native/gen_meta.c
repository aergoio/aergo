/**
 * @file    gen_meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_meta.h"

BinaryenType
meta_gen(gen_t *gen, meta_t *meta)
{
    switch (meta->type) {
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
    case TYPE_TUPLE:
    default:
        ASSERT1(!"invalid type", meta->type);
    }

    return BinaryenTypeUnreachable();
}

/* end of gen_meta.c */
