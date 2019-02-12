/**
 * @file    gen_fn.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_FN_H
#define _GEN_FN_H

#include "common.h"

#include "ir_abi.h"
#include "ir_fn.h"
#include "gen.h"

BinaryenFunctionTypeRef abi_gen(gen_t *gen, ir_abi_t *abi);

void fn_gen(gen_t *gen, ir_fn_t *fn);

#endif /* ! _GEN_FN_H */
