/**
 * @file    preprocess.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "strbuf.h"
#include "stack.h"

#include "prep.h"

#define YY_LINE                 scan->pos.first_line
#define YY_COL                  scan->pos.first_col
#define YY_OFFSET               scan->pos.first_offset

#define yy_update_line()        scan->pos.first_line++
#define yy_update_col()         scan->pos.first_col++
#define yy_update_offset()      scan->pos.first_offset++

static void substitue(char *path, stack_t *imp, strbuf_t *out);

static void
scan_init(scan_t *scan, char *path, strbuf_t *out)
{
    scan->path = path;
    scan->fp = open_file(path, "r");

    errpos_init(&scan->pos, path);

    scan->buf_len = 0;
    scan->buf_pos = 0;
    scan->buf[0] = '\0';

    scan->out = out;
}

static char
scan_next(scan_t *scan)
{
    char c;

    if (scan->buf_pos >= scan->buf_len) {
        scan->buf_len = fread(scan->buf, 1, sizeof(scan->buf), scan->fp);
        if (scan->buf_len == 0)
            return EOF;

        scan->buf_pos = 0;
    }

    c = scan->buf[scan->buf_pos++];

    if (c == '\n' || c == '\r')
        yy_update_line();

    yy_update_offset();

    return c;
}

static char
scan_peek(scan_t *scan, int cnt)
{
    if (scan->buf_pos + cnt >= scan->buf_len) {
        scan->buf_len -= scan->buf_pos;
        memmove(scan->buf, scan->buf + scan->buf_pos, scan->buf_len);
        scan->buf_pos = 0;

        scan->buf_len +=
            fread(scan->buf + scan->buf_len, 1,
                  sizeof(scan->buf) - scan->buf_len, scan->fp);
        if (scan->buf_len <= cnt)
            return EOF;
    }

    return scan->buf[scan->buf_pos + cnt];
}

static bool
add_file(scan_t *scan, char *path, stack_t *imp)
{
    stack_node_t *node = stack_top(imp);

    if (node != NULL) {
        while (true) {
            if (strcmp(node->item, path) == 0) {
                TRACE(ERROR_CROSS_IMPORT, &scan->pos, path);
                return false;
            }
        }
    }

    stack_push(imp, xstrdup(path));
    scan->pos.path = path;

    return true;
}

static void
put_char(scan_t *scan, char c)
{
    strbuf_append(scan->out, &c, 1);
}

static void
skip_to_eol(scan_t *scan)
{
    char c;

    while ((c = scan_next(scan)) != EOF) {
        if (c == '\n' || c == '\r')
            break;
    }
}

static void
skip_comment(scan_t *scan, char c)
{
    char n;

    put_char(scan, c);

    if (scan_peek(scan, 0) == '*') {
        while ((n = scan_next(scan)) != EOF) {
            if (n == '*' && scan_peek(scan, 0) == '/') {
                scan_next(scan);
                break;
            }
        }
    }
    else if (scan_peek(scan, 0) == '/') {
        skip_to_eol(scan);
    }
}

static void
put_literal(scan_t *scan, char c)
{
    char n;

    put_char(scan, c);

    while ((n = scan_next(scan)) != EOF) {
        put_char(scan, n);

        if (n != '\\' && scan_peek(scan, 0) == '"') {
            put_char(scan, scan_next(scan));
            break;
        }
    }
}

/* WARNING: keep "void" not "static void" for tests */
void
mark_file(char *path, int line, int offset, strbuf_t *out)
{
    char buf[PATH_MAX_LEN + 16];

    snprintf(buf, sizeof(buf), "#file \"%s\" %d %d\n", path, line, offset);

    strbuf_append(out, buf, strlen(buf));
}

static void
put_import(scan_t *scan, stack_t *imp)
{
    int offset = 0;
    char path[PATH_MAX_LEN];
    char c, n;
    stack_node_t *node;

    // TODO: need more error handling
    while ((c = scan_next(scan)) != EOF) {
        if (c == '"') {
            while ((n = scan_next(scan)) != EOF) {
                path[offset++] = n;

                if (n != '\\' && scan_peek(scan, 0) == '"') {
                    path[offset] = '\0';

                    mark_file(path, 1, 0, scan->out);
                    substitue(path, imp, scan->out);
                    mark_file(scan->path, YY_LINE + 1, YY_OFFSET, scan->out);

                    stack_pop(imp);
                    offset = 0;
                    break;
                }
            }
        }
        else if (c == '\n' || c == '\r') {
            break;
        }
        /*
        else if (!isspace(c)) {
            TRACE(ERROR_UNKNOWN_CHAR, &scan->pos, scan->path);
            skip_to_eol(scan);
            break;
        }
        */
    }
}

static void
substitue(char *path, stack_t *imp, strbuf_t *out)
{
    bool is_first_ch = true;
    char c;
    scan_t scan;

    scan_init(&scan, path, out);

    if (!add_file(&scan, path, imp))
        return;

    while ((c = scan_next(&scan)) != EOF) {
        if (c == '/') {
            skip_comment(&scan, c);
            is_first_ch = false;
        }
        else if (c == '"') {
            put_literal(&scan, c);
            is_first_ch = false;
        }
        else if (c == '\n' || c == '\r') {
            put_char(&scan, c);
            is_first_ch = true;
        }
        else if (c == ' ' || c == '\t' || c == '\f') {
            put_char(&scan, c);
        }
        else if (is_first_ch && c == 'i' &&
                 scan_peek(&scan, 0) == 'm' &&
                 scan_peek(&scan, 0) == 'p' &&
                 scan_peek(&scan, 0) == 'o' &&
                 scan_peek(&scan, 0) == 'r' &&
                 scan_peek(&scan, 0) == 't' &&
                 isblank(scan_peek(&scan, 0))) {
            put_import(&scan, imp);
            is_first_ch = false;
        }
        else {
            put_char(&scan, c);
            is_first_ch = false;
        }
    }
}

void
preprocess(char *path, strbuf_t *out)
{
    stack_t imp;

    stack_init(&imp);

    substitue(path, &imp, out);
}

/* end of preprocess.c */
