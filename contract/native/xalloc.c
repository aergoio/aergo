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

    mem = malloc(size);
    if (mem == NULL)
        FATAL(ERROR_OUT_OF_MEM, strerror(errno));
        
    return mem;
}

void *
xcalloc(size_t size)
{
    void *mem;

    mem = calloc(1, size);
    if (mem == NULL)
        FATAL(ERROR_OUT_OF_MEM, strerror(errno));
        
    return mem;
}

void *
xrealloc(void *ptr, size_t size)
{
    void *mem;

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

/* end of xalloc.c */
