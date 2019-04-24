/**
 * @file    src_pos.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "src_pos.h"

src_pos_t null_pos_ = { NULL, NULL, 1, 1, 0, 1, 1, 0 };

void
src_pos_print(src_pos_t *pos, char *buf, int buf_sz)
{
    int i;
    int idx = 0;
    int src_len;

    ASSERT(pos->src != NULL);
    ASSERT1(pos->first_offset >= 0, pos->first_offset);
    ASSERT1(pos->first_col > 0, pos->first_col);

    /* 18 = '\n' + strlen(ANSI_GREEN) + '^' + strlen(ANSI_NONE) + '\0' */
    src_len = MIN((int)strlen(pos->src + pos->first_offset), buf_sz - pos->first_col - 18);

    for (i = 0; i < src_len; i++) {
        char c = pos->src[i + pos->first_offset];

        if (c == '\0' || c == '\n' || c == '\r')
            break;

        buf[idx++] = c;
    }

    buf[idx++] = '\n';

    for (i = 0; i < pos->first_col - 1; i++) {
        char c = pos->src[i + pos->first_offset];

        if (c == '\t' || c == '\f')
            buf[idx++] = c;
        else
            buf[idx++] = ' ';
    }

    strcpy(&buf[idx++], ANSI_GREEN"^"ANSI_NONE);
}

/* end of src_pos.c */
