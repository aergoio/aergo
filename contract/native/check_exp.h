/**
 * @file    check_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_EXP_H
#define _CHECK_EXP_H

#include "common.h"

#include "check.h"
#include "ast_exp.h"

int check_exp(check_t *ctx, ast_exp_t *exp);

#endif /* ! _CHECK_EXP_H */
