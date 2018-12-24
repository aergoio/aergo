/**
 * @file    trans_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRANS_EXP_H
#define _TRANS_EXP_H

#include "common.h"

#include "ast_exp.h"
#include "trans.h"

ast_exp_t *exp_trans(trans_t *trans, ast_exp_t *exp);

#endif /* ! _TRANS_EXP_H */
