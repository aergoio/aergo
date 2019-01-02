/**
 * @file    trans_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRANS_EXP_H
#define _TRANS_EXP_H

#include "common.h"

#include "ast_exp.h"
#include "trans.h"

#define exp_trans_to_lval(trans, exp)                                                    \
    do {                                                                                 \
        bool is_lval = (trans)->is_lval;                                                 \
        (trans)->is_lval = true;                                                         \
        exp_trans((trans), (exp));                                                       \
        (trans)->is_lval = is_lval;                                                      \
    } while (0)

#define exp_trans_to_rval(trans, exp)                                                    \
    do {                                                                                 \
        bool is_lval = (trans)->is_lval;                                                 \
        (trans)->is_lval = false;                                                        \
        exp_trans((trans), (exp));                                                       \
        (trans)->is_lval = is_lval;                                                      \
    } while (0)

void exp_trans(trans_t *trans, ast_exp_t *exp);

#endif /* ! _TRANS_EXP_H */
