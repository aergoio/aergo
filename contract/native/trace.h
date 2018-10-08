/**
 * @file    trc.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRACE_H
#define _TRACE_H

#include "common.h"

#include "util.h"

#define src_pos_init(pos, path)         src_pos_set((pos), (path), 1, 1, 0)

#define trace_rel_path(trc)             (trc)->rel.path
#define trace_rel_line(trc)             (trc)->rel.first_line
#define trace_rel_offset(trc)           (trc)->rel.first_offset

typedef struct src_pos_s {
    char *path;
    int first_line;
    int first_col;
    int first_offset;
    int last_line;
    int last_col;
    int last_offset;
} src_pos_t;

typedef struct trace_s {
    char *src;
    src_pos_t abs;      /* absolute position */
    src_pos_t rel;      /* relative position */
} trace_t;

int trace_cmp(trace_t *trc1, trace_t *trc2);

void trace_dump(trace_t *trc);

static inline void
src_pos_set(src_pos_t *trc, char *path, int line, int col, int offset)
{
    trc->path = path;
    trc->first_line = line;
    trc->first_col = col;
    trc->first_offset = offset;
    trc->last_line = line;
    trc->last_col = col;
    trc->last_offset = offset;
}

static inline void
src_pos_update_first(src_pos_t *trc)
{
    trc->first_line = trc->last_line;
    trc->first_col = trc->last_col;
    trc->first_offset = trc->last_offset;
}

static inline void
src_pos_update_last_line(src_pos_t *trc)
{
    trc->last_line++;
    trc->last_offset += trc->last_col;
    trc->last_col = 1;
}

static inline void
src_pos_update_last_col(src_pos_t *trc, int cnt)
{
    trc->last_col += cnt;
}

static inline void
trace_init(trace_t *trc, char *src, char *path)
{
    ASSERT(trc != NULL);

    trc->src = src;
    src_pos_set(&trc->abs, path, 1, 1, 0);
    src_pos_set(&trc->rel, path, 1, 1, 0);
}

static inline void
trace_set_src(trace_t *trc, char *src)
{
    ASSERT(trc != NULL);

    trc->src = src;
}

static inline void
trace_update_first(trace_t *trc)
{
    ASSERT(trc != NULL);

    src_pos_update_first(&trc->abs);
    src_pos_update_first(&trc->rel);
}

static inline void
trace_update_last_line(trace_t *trc)
{
    ASSERT(trc != NULL);

    src_pos_update_last_line(&trc->abs);
    src_pos_update_last_line(&trc->rel);
}

static inline void
trace_update_last_col(trace_t *trc, int cnt)
{
    ASSERT(trc != NULL);

    src_pos_update_last_col(&trc->abs, cnt);
    src_pos_update_last_col(&trc->rel, cnt);
}

static inline void
trace_set_rel_path(trace_t *trc, char *path)
{
    ASSERT(trc != NULL);

    trc->rel.path = path;
}

static inline void
trace_set_rel_pos(trace_t *trc, int line, int col, int offset)
{
    ASSERT(trc != NULL);

    trc->rel.last_line = line;
    trc->rel.last_col = col;
    trc->rel.last_offset = offset;
}

#endif /* no _TRACE_H */
