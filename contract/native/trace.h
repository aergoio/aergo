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
#define trace_rel_col(trc)              (trc)->rel.first_col
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
src_pos_update_last_offset(src_pos_t *trc, int cnt)
{
    trc->last_offset += cnt;
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
trace_update_last_offset(trace_t *trc, int cnt)
{
    ASSERT(trc != NULL);

    src_pos_update_last_offset(&trc->abs, cnt);
    src_pos_update_last_offset(&trc->rel, cnt);
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

static inline int
trace_cmp(trace_t *trc1, trace_t *trc2)
{
    int res;
    src_pos_t *pos1 = &trc1->rel;
    src_pos_t *pos2 = &trc2->rel;

    ASSERT(pos1->path != NULL);
    ASSERT(pos2->path != NULL);

    res = strcmp(pos1->path, pos2->path);
    if (res != 0)
        return res;

    if (pos1->first_line == pos2->first_line)
        return 0;
    else if (pos1->first_line < pos2->first_line)
        return -1;
    else
        return 1;
}

static inline void
trace_dump(trace_t *trc)
{
#define TRACE_LINE_MAX      80
    int i;
    int tok_len;
    int adj_col, adj_offset, adj_len;
    char buf[TRACE_LINE_MAX * 3 + 4];
    src_pos_t *pos = &trc->abs;

    adj_offset = pos->first_offset;
    ASSERT(adj_offset >= 0);

    adj_col = pos->first_col;
    ASSERT(adj_col > 0);

    tok_len = MIN(pos->last_offset - pos->first_offset, TRACE_LINE_MAX - 1);
    ASSERT(tok_len >= 0);

    if (adj_col + tok_len > TRACE_LINE_MAX) {
        adj_col = TRACE_LINE_MAX - tok_len;
        adj_offset += pos->first_col - adj_col;
    }

    adj_len = MIN(strlen(trc->src + adj_offset), TRACE_LINE_MAX);

    for (i = 0; i < adj_len; i++) {
        char c = trc->src[i + adj_offset];

        if (c == '\0' || c == '\n' || c == '\r')
            break;

        buf[i] = c;
    }
    buf[i] = '\0';

    fprintf(stderr, "%s\n", buf);

    for (i = 0; i < adj_col - 1; i++) {
        buf[i] = ' ';
    }
    strcpy(&buf[i], ANSI_GREEN"^"ANSI_NONE);

    fprintf(stderr, "%s\n", buf);
}

#endif /* no _TRACE_H */
