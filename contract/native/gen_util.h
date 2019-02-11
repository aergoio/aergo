/**
 * @file    gen_util.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_UTIL_H
#define _GEN_UTIL_H

#include "common.h"

#include "gen.h"
#include "binaryen-c.h"

#define i32_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralInt32(v))
#define i64_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralInt64(v))
#define f32_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralFloat32(v))
#define f64_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralFloat64(v))

#define meta_gen(meta)                                                                   \
    (is_array_meta(meta) ? BinaryenTypeInt32() : type_gen((meta)->type))

BinaryenType type_gen(type_t type);

void malloc_gen(gen_t *gen);

#endif /* ! _GEN_UTIL_H */
