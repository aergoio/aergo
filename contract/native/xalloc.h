/**
 * @file    xalloc.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _XALLOC_H
#define _XALLOC_H

#include "common.h"

#define xstrdup(s)          xstrndup((s), strlen(s))
#define xstrcat(s1, s2)     xstrncat((s1), strlen(s1), (s2), strlen(s2))

void *xmalloc(size_t size);
void *xcalloc(size_t size);
void *xrealloc(void *ptr, size_t size);
void xfree(void *ptr);

char *xstrndup(char *str, int len);
char *xstrncat(char *s1, int len1, char *s2, int len2);

#endif /* ! _XALLOC_H */
