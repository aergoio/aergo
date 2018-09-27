/**
 * @file    flag.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _FLAG_H
#define _FLAG_H

#include "common.h"

#define flag_set(x, y)      ((x) |= (y))
#define flag_on(x, y)       (((x) & (y)) == (y))
#define flag_off(x, y)      (((x) & (y)) != (y))

typedef enum flag_e {
    FLAG_NONE       = 0x00,
    FLAG_VERBOSE    = 0x01,
    FLAG_LEX_DUMP   = 0x02,
    FLAG_YACC_DUMP  = 0x04,
    FLAG_AST_DUMP   = 0x08,
    FLAG_TEST       = 0x10
} flag_t;

#endif /* ! _FLAG_H */
