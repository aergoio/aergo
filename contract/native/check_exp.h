/**
 * @file    check_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_EXP_H
#define _CHECK_EXP_H

#include "common.h"

#include "check.h"

int check_exp(check_t *check, ast_exp_t *exp);

int exp_id_check(check_t *check, ast_exp_t *exp);
int exp_lit_check(check_t *check, ast_exp_t *exp);
int exp_type_check(check_t *check, ast_exp_t *exp);
int exp_array_check(check_t *check, ast_exp_t *exp);
int exp_op_check(check_t *check, ast_exp_t *exp);
int exp_access_check(check_t *check, ast_exp_t *exp);
int exp_call_check(check_t *check, ast_exp_t *exp);
int exp_sql_check(check_t *check, ast_exp_t *exp);
int exp_ternary_check(check_t *check, ast_exp_t *exp);
int exp_tuple_check(check_t *check, ast_exp_t *exp);

#endif /* ! _CHECK_EXP_H */
