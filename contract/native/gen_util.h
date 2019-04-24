/**
 * @file    gen_util.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_UTIL_H
#define _GEN_UTIL_H

#include "common.h"

#include "enum.h"
#include "binaryen-c.h"

#define i32_gen(gen, v)             BinaryenConst((gen)->module, BinaryenLiteralInt32(v))
#define i64_gen(gen, v)             BinaryenConst((gen)->module, BinaryenLiteralInt64(v))
#define f32_gen(gen, v)             BinaryenConst((gen)->module, BinaryenLiteralFloat32(v))
#define f64_gen(gen, v)             BinaryenConst((gen)->module, BinaryenLiteralFloat64(v))

#define meta_gen(meta)                                                                             \
    (is_array_meta(meta) ? BinaryenTypeInt32() : type_gen((meta)->type))

static inline BinaryenType
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
    case TYPE_INT128:
        return BinaryenTypeInt32();

    case TYPE_INT64:
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
    case TYPE_TUPLE:
        return BinaryenTypeInt32();

    default:
        ASSERT1(!"invalid type", type);
    }

    return BinaryenTypeUnreachable();
}

#endif /* ! _GEN_UTIL_H */
