/**
 * @file    src_pos.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "src_pos.h"

void
src_pos_dump(src_pos_t *src_pos, char *buf)
{
#define SRC_POS_LINE_MAX      80
    int i;
    int idx = 0;
    int tok_len;
    int adj_col, adj_offset, adj_len;
    pos_t *pos = &src_pos->abs;

    adj_offset = pos->first_offset;
    adj_col = pos->first_col;

    ASSERT(adj_offset >= 0);
    ASSERT(adj_col > 0);

    tok_len = MIN(pos->last_offset - pos->first_offset, SRC_POS_LINE_MAX - 1);
    ASSERT(tok_len >= 0);

    if (adj_col + tok_len > SRC_POS_LINE_MAX) {
        adj_col = SRC_POS_LINE_MAX - tok_len;
        adj_offset += pos->first_col - adj_col;
    }

    adj_len = MIN(strlen(src_pos->src + adj_offset), SRC_POS_LINE_MAX);

    for (i = 0; i < adj_len; i++) {
        char c = src_pos->src[i + adj_offset];

        if (c == '\0' || c == '\n' || c == '\r')
            break;

        buf[idx++] = c;
    }

    buf[idx++] = '\n';

    for (i = 0; i < adj_col - 1; i++) {
        buf[idx++] = ' ';
    }

    strcpy(&buf[idx++], ANSI_GREEN"^"ANSI_NONE);
}

/* end of src_pos.c */
