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
    OPT_NONE        = 0x00,
    OPT_LEX_DUMP    = 0x01,
    OPT_YACC_DUMP   = 0x02,
    OPT_SILENT      = 0x04
} opt_t;

#endif /* _OPTION_H */
