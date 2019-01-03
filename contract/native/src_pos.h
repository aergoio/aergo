/**
 * @file    src_pos.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SRC_POS_H
#define _SRC_POS_H

#include "common.h"

#include "util.h"

typedef struct pos_s {
    char *path;
    int first_line;
    int first_col;
    int first_offset;
    int last_line;
    int last_col;
    int last_offset;
} pos_t;

typedef struct src_pos_s {
    char *src;
    pos_t abs;      /* absolute position */
    pos_t rel;      /* relative position */
} src_pos_t;

void src_pos_print(src_pos_t *pos, char *buf);

static inline void
pos_set(pos_t *pos, char *path, int line, int col, int offset)
{
    pos->path = path;
    pos->first_line = line;
    pos->first_col = col;
    pos->first_offset = offset;
    pos->last_line = line;
    pos->last_col = col;
    pos->last_offset = offset;
}

static inline void
pos_update_first(pos_t *pos)
{
    pos->first_line = pos->last_line;
    pos->first_col = pos->last_col;
    pos->first_offset = pos->last_offset;
}

static inline void
pos_update_line(pos_t *pos)
{
    pos->last_line++;
    pos->last_offset += pos->last_col;
    pos->last_col = 1;
}

static inline void
pos_update_col(pos_t *pos, int cnt)
{
    pos->last_col += cnt;
}

static inline void
src_pos_init(src_pos_t *pos, char *src, char *path)
{
    ASSERT(pos != NULL);

    pos->src = src;
    pos_set(&pos->abs, path, 1, 1, 0);
    pos_set(&pos->rel, path, 1, 1, 0);
}

static inline void
src_pos_update_first(src_pos_t *src_pos)
{
    pos_update_first(&src_pos->abs);
    pos_update_first(&src_pos->rel);
}

static inline void
src_pos_update_line(src_pos_t *src_pos)
{
    pos_update_line(&src_pos->abs);
    pos_update_line(&src_pos->rel);
}

static inline void
src_pos_update_col(src_pos_t *src_pos, int cnt)
{
    pos_update_col(&src_pos->abs, cnt);
    pos_update_col(&src_pos->rel, cnt);
}

static inline void
src_pos_set_src(src_pos_t *src_pos, char *src)
{
    src_pos->src = src;
}

static inline void
src_pos_set_path(src_pos_t *src_pos, char *path)
{
    src_pos->rel.path = path;
}

static inline void
src_pos_set_pos(src_pos_t *src_pos, int line, int col, int offset)
{
    src_pos->rel.last_line = line;
    src_pos->rel.last_col = col;
    src_pos->rel.last_offset = offset;
}

#endif /* no _SRC_POS_H */
