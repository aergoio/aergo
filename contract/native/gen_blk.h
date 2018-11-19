/**
 * @file    gen_blk.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_BLK_H
#define _GEN_BLK_H

#include "common.h"

#include "ast_blk.h"
#include "gen.h"

BinaryenExpressionRef blk_gen(gen_t *gen, ast_blk_t *blk);

#endif /* ! _GEN_BLK_H */
