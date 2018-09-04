/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "parser.h"
#include "util.h"

#include "compile.h"

int
compile(char *path, opt_t opt)
{
    ec_t ec = NO_ERROR;
    char opath[PATH_MAX_LEN + 3];

    snprintf(opath, sizeof(opath), "%s.i", path);

    ec = preprocess(path, opath, opt);
    if (ec == NO_ERROR)
        ec = parse(opath, opt);

    return ec;
}

static void
yypos_init(yypos_t *pos)
{
    pos->line = 1;
    pos->col = 1;
    pos->offset = 0;
}

void
yyparam_init(yyparam_t *param, char *path)
{
    ASSERT(path != NULL);

    param->path = path;
    param->fp = open_file(path, "r");

    yypos_init(&param->lloc.first);
    yypos_init(&param->lloc.last);

    strbuf_init(&param->buf);
}

char *
yyparam_trace(yyparam_t *param)
{
    int i, j;
    int nread;
    int buf_size;
    char *buf;
    FILE *fp = open_file(param->path, "r");

    if (fseek(fp, param->lloc.first.offset, SEEK_SET) < 0)
        FATAL(ERROR_FILE_IO_FAILED, strerror(errno));

    buf_size = max(param->lloc.first.col * 2, STRBUF_INIT_SIZE);
    buf = malloc(buf_size);

    nread = fread(buf, buf_size, 1, fp);
    if (nread <= 0 && !feof(fp))
        FATAL(ERROR_FILE_IO_FAILED, strerror(errno));

    for (i = 0; i < nread; i++) {
        if (buf[i] == '\n' || buf[i] == '\r')
            break;
    }

    for (j = 0; j < param->lloc.first.col - 1; j++) {
        buf[i + j] = ' ';
    }

    strcpy(buf + i + j, ANSI_GREEN"^"ANSI_NONE);

    return buf;
}

/* end of compile.c */
