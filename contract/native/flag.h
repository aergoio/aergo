/**
 * @file    flag.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _FLAG_H
#define _FLAG_H

#include "common.h"

#define flag_set(x, y)      ((x) |= (y))
#define flag_enabled(x, y)  (((x) & (y)) == (y))

typedef enum flag_e {
    FLAG_NONE       = 0x00,
    FLAG_SILENT     = 0x01,
    FLAG_LEX_DUMP   = 0x02,
    FLAG_YACC_DUMP  = 0x04
} flag_t;

#endif /* ! _FLAG_H */
