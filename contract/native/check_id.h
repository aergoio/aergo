/**
 * @file    check_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_ID_H
#define _CHECK_ID_H

#include "common.h"

#include "ast_id.h"
#include "check.h"

void id_check(check_t *check, ast_id_t *id);

void id_trycheck(check_t *check, ast_id_t *id);

#endif /* ! _CHECK_ID_H */
