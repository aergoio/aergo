/**
 * @file    util.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _UTIL_H
#define _UTIL_H

#include "common.h"

#include "errors.h"

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif  /* _STRBUF_T */

void read_file(char *path, strbuf_t *sb);

#endif /* no _UTIL_H */
