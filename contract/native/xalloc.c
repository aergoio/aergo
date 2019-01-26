/**
 * @file    xalloc.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "error.h"

#include "xalloc.h"

// TODO: gather memory stats

void *
xmalloc(size_t size)
{
    void *mem;

    if (size == 0)
        return NULL;

    mem = malloc(size);
    if (mem == NULL)
        FATAL(ERROR_OUT_OF_MEM, strerror(errno));
        
    return mem;
}

void *
xcalloc(size_t size)
{
    void *mem;

    if (size == 0)
        return NULL;

    mem = calloc(1, size);
    if (mem == NULL)
        FATAL(ERROR_OUT_OF_MEM, strerror(errno));
        
    return mem;
}

void *
xrealloc(void *ptr, size_t size)
{
    void *mem;

    if (size == 0)
        return NULL;

    mem = realloc(ptr, size);
    if (mem == NULL)
        FATAL(ERROR_OUT_OF_MEM, strerror(errno));
        
    return mem;
}

void
xfree(void *ptr)
{
    free(ptr);
}

char *
xstrndup(char *str, int len)
{
    char *buf = xmalloc(len + 1);

    memcpy(buf, str, len);
    buf[len] = '\0';

    return buf;
}

char *
xstrncat(char *s1, int len1, char *s2, int len2)
{
    char *buf = xmalloc(len1 + len2 + 1);

    memcpy(buf, s1, len1);
    memcpy(buf + len1, s2, len2);

    buf[len1 + len2] = '\0';

    return buf;
}

/* end of xalloc.c */
