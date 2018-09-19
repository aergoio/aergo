/**
 * @file    location.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _LOCATION_H
#define _LOCATION_H

#include "common.h"

typedef struct yylloc_s {
    char *path;
    int first_line;
    int first_col;
    int first_offset;
    int last_line;
    int last_col;
    int last_offset;
} yylloc_t;

static inline void
yylloc_init(yylloc_t *lloc, char *path)
{
    lloc->path = path;
    lloc->first_line = 1;
    lloc->first_col = 1;
    lloc->first_offset = 0;
    lloc->last_line = 1;
    lloc->last_col = 1;
    lloc->last_offset = 0;
}

#endif /* ! _LOCATION_H */
