/**
 * @ib    iobuf.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "error.h"

#include "iobuf.h"

void
iobuf_init(iobuf_t *ib, char *path)
{
    ASSERT(ib != NULL);

    ib->path = path;

    ib->size = IOBUF_INIT_SIZE;
    ib->buf = xmalloc(ib->size + 1);

    iobuf_reset(ib);
}

void
iobuf_reset(iobuf_t *ib)
{
    ASSERT(ib != NULL);

    ib->offset = 0;
    ib->buf[0] = '\0';
}

void
iobuf_load(iobuf_t *ib)
{
    int n;
    char buf[IOBUF_INIT_SIZE];
    char *path = iobuf_path(ib);
    FILE *fp = open_file(path, "r");

    while ((n = fread(buf, 1, sizeof(buf), fp)) > 0) {
        if (ib->offset + n > ib->size) {
            ib->size += IOBUF_INIT_SIZE;
            ib->buf = xrealloc(ib->buf, ib->size + 1);
        }

        memcpy(ib->buf + ib->offset, buf, n);
        ib->offset += n;
    }

    if (!feof(fp))
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    ib->buf[ib->offset] = '\0';

    fclose(fp);
}

void
iobuf_print(iobuf_t *ib, char *path)
{
    int n;
    FILE *fp = open_file(path, "w");

    n = fwrite(iobuf_str(ib), 1, iobuf_size(ib), fp);
    if (n < 0)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    fclose(fp);
}

/* end of iobuf.c */
