/**
 * @file    trace.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "trace.h"

int
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

    if (pos1->first_line < pos2->first_line)
        return -1;
    else if (pos1->first_line > pos2->first_line)
        return 1;

    if (pos1->first_col < pos2->first_col)
        return -1;
    else if (pos1->first_col == pos2->first_col)
        return 0;
    else
        return 1;
}

void
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

/* end of trace.c */
