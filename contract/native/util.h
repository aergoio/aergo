/**
 * @file    util.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _UTIL_H
#define _UTIL_H

#include "common.h"

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif /* ! _STRBUF_T */

#define MAX(x, y)           ((x) > (y) ? (x) : (y))
#define MIN(x, y)           ((x) > (y) ? (y) : (x))

FILE *open_file(char *path, char *mode);
void close_file(FILE *fp);

char *trim_str(char *str);

#endif /* ! _UTIL_H */
