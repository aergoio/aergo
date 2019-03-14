/**
 * @ib    iobuf.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IOBUF_H
#define _IOBUF_H

#include "common.h"

#define IOBUF_INIT_SIZE             8192

#define is_empty_iobuf(ib)          ((ib)->offset == 0)

#define iobuf_path(ib)              ((ib)->path)
#define iobuf_size(ib)              ((ib)->offset)
#define iobuf_str(ib)               ((ib)->buf)

#ifndef _IOBUF_T
#define _IOBUF_T
typedef struct iobuf_s iobuf_t;
#endif /* ! _IOBUF_T */

struct iobuf_s {
    char *path;
    int offset;
    int size;
    char *buf;
};

void iobuf_init(iobuf_t *ib, char *path);
void iobuf_reset(iobuf_t *ib);

void iobuf_load(iobuf_t *ib);
void iobuf_print(iobuf_t *ib, char *path);

static inline char
iobuf_char(iobuf_t *ib, int i)
{
    ASSERT(ib != NULL);

    return i > ib->offset ? EOF : ib->buf[i];
}

static inline void
iobuf_set(iobuf_t *ib, int size, char *buf)
{
    ib->offset = size;
    ib->size = size;
    ib->buf = buf;
}

#endif /* ! _IOBUF_H */
