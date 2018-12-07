/**
 * @file    gen_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_EXP_H
#define _GEN_EXP_H

#include "common.h"

#include "ast_exp.h"
#include "gen.h"

BinaryenExpressionRef exp_gen(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref);

#endif /* ! _GEN_EXP_H */
