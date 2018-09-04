/**
 * @file    parser.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSER_H
#define _PARSER_H

#include "common.h"

#include "compile.h"

#define YYLTYPE             yylloc_t

#include "grammar.tab.h"

int parse(char *path, opt_t opt);

static inline void
yyget_trace(char *src, int len, yylloc_t *lloc, char *buf)
{
    int i, j;

    for (i = lloc->first.offset, j = 0; i < len; i++) {
        buf[j++] = src[i];
        if (src[i] == '\n' || src[i] == '\r')
            break;
    }

    for (i = 0; i < lloc->first.col - 1; i++) {
        buf[j++] = ' ';
    }

    strcpy(buf + j, ANSI_GREEN"^"ANSI_NONE);
}

#endif /*_PARSER_H */
