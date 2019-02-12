/**
 * @file    src_pos.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "src_pos.h"

src_pos_t null_pos_ = { NULL, NULL, 1, 1, 0, 1, 1, 0 };

void
src_pos_print(src_pos_t *pos, char *buf, int buf_size)
{
    int i;
    int idx = 0;
    int tok_len;
    int line_max = (buf_size - 19) / 2;
    int adj_col, adj_offset, adj_len;

    adj_offset = pos->first_offset;
    adj_col = pos->first_col;

    ASSERT(adj_offset >= 0);
    ASSERT(adj_col > 0);

    tok_len = MIN(pos->last_offset - pos->first_offset, line_max - 1);
    ASSERT(tok_len >= 0);

    if (adj_col + tok_len > line_max) {
        adj_col = line_max - tok_len;
        adj_offset += pos->first_col - adj_col;
    }

    adj_len = MIN(strlen(pos->src + adj_offset), line_max);

    for (i = 0; i < adj_len; i++) {
        char c = pos->src[i + adj_offset];

        if (c == '\0' || c == '\n' || c == '\r')
            break;

        buf[idx++] = c;
    }

    buf[idx++] = '\n';

    for (i = 0; i < adj_col - 1; i++) {
        char c = pos->src[i + adj_offset];

        if (c == '\t' || c == '\f')
            buf[idx++] = c;
        else
            buf[idx++] = ' ';
    }

    strcpy(&buf[idx++], ANSI_GREEN"^"ANSI_NONE);
}

/* end of src_pos.c */
