/**
 * @file    option.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _OPTION_H
#define _OPTION_H

#include "common.h"

#define opt_set(x, y)       ((x) |= (y))
#define opt_enabled(x, y)   (((x) & (y)) == (y))

typedef enum opt_e {
    OPT_NORMAL      = 0x00,
    OPT_DEBUG       = 0x01,
    OPT_TEST        = 0x02
} opt_t;

#endif /* _OPTION_H */
