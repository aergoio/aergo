/**
 * @file    gen_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_STMT_H
#define _GEN_STMT_H

#include "common.h"

#include "ast_stmt.h"
#include "gen.h"

void stmt_gen(gen_t *gen, ast_stmt_t *stmt);

#endif /* ! _GEN_STMT_H */
