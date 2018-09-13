/**
 * @file    location.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _LOCATION_H
#define _LOCATION_H

#include "common.h"

typedef struct yypos_s {
    int line;
    int col;
    int offset;
} yypos_t;

typedef struct yylloc_s {
    yypos_t first;
    yypos_t last;
} yylloc_t;

static inline void
yypos_init(yypos_t *pos)
{
    pos->line = 1;
    pos->col = 1;
    pos->offset = 0;
}

#endif /* _LOCATION_H */
