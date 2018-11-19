/**
 * @file    gen_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_ID_H
#define _GEN_ID_H

#include "common.h"

#include "ast_id.h"
#include "gen.h"

BinaryenExpressionRef id_gen(gen_t *gen, ast_id_t *id);

#endif /* ! _GEN_ID_H */
