/**
 * @file    src_pos.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SRC_POS_H
#define _SRC_POS_H

#include "common.h"

typedef struct src_pos_s {
    char *path;
    char *src;

    int first_line;
    int first_col;
    int first_offset;

    int last_line;
    int last_col;
    int last_offset;
} src_pos_t;

extern src_pos_t null_pos_;

void src_pos_print(src_pos_t *pos, char *buf, int buf_sz);

static inline void
src_pos_init(src_pos_t *pos, char *src, char *path)
{
    ASSERT(pos != NULL);

    pos->src = src;
    pos->path = path;

    pos->first_line = 1;
    pos->first_col = 1;
    pos->first_offset = 0;

    pos->last_line = 1;
    pos->last_col = 1;
    pos->last_offset = 0;
}

static inline void
src_pos_update_first(src_pos_t *pos)
{
    pos->first_line = pos->last_line;
    pos->first_col = pos->last_col;
    pos->first_offset = pos->last_offset;
}

static inline void
src_pos_update_line(src_pos_t *pos)
{
    pos->last_line++;
    pos->last_offset += pos->last_col;
    pos->last_col = 1;
}

static inline void
src_pos_update_col(src_pos_t *pos, int cnt)
{
    pos->last_col += cnt;
}

#endif /* no _SRC_POS_H */
