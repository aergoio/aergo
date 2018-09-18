/**
 * @file    xalloc.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _XALLOC_H
#define _XALLOC_H

#include "common.h"

#define xstrdup(s)          xstrndup(s, strlen(s))

void *xmalloc(size_t size);
void *xcalloc(size_t size);
void *xrealloc(void *ptr, size_t size);
void xfree(void *ptr);

char *xstrndup(char *str, int len);

#endif /* ! _XALLOC_H */
