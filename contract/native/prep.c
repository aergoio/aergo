/**
 * @file    prep.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "strbuf.h"
#include "stack.h"

#include "prep.h"

#define YY_LINE                 scan->pos.rel.first_line
#define YY_OFFSET               scan->pos.rel.first_offset

#define yy_update_line()        src_pos_update_line(&scan->pos)
#define yy_update_col()         src_pos_update_col(&scan->pos, 1)

#define yy_update_first()       src_pos_update_first(&scan->pos)

static void substitue(char *path, char *work_dir, stack_t *imp, strbuf_t *out);

static void
scan_init(scan_t *scan, char *path, char *work_dir, strbuf_t *out)
{
    scan->path = path;
    scan->work_dir = work_dir;

    scan->offset = 0;

    strbuf_init(&scan->in);
    strbuf_load(&scan->in, path);

    src_pos_init(&scan->pos, strbuf_str(&scan->in), xstrdup(path));

    scan->out = out;
}

static char
scan_next(scan_t *scan)
{
    char c;

    if (scan->offset >= strbuf_size(&scan->in))
        return EOF;

    c = strbuf_char(&scan->in, scan->offset++);

    yy_update_col();

    if (c == '\n' || c == '\r') {
        yy_update_line();
        yy_update_first();
    }

    return c;
}

static char
scan_peek(scan_t *scan, int cnt)
{
    if (scan->offset >= strbuf_size(&scan->in))
        return EOF;

    return strbuf_char(&scan->in, scan->offset + cnt);
}

static bool
add_file(scan_t *scan, char *path, stack_t *imp)
{
    stack_node_t *node;

    stack_foreach(node, imp) {
        if (strcmp(node->item, path) == 0) {
            ERROR(ERROR_CROSS_IMPORT, &scan->pos, FILENAME(path));
            return false;
        }
    }

    stack_push(imp, xstrdup(path));

    return true;
}

static void
del_file(scan_t *scan, char *path, stack_t *imp)
{
    stack_pop(imp);
}

static void
put_char(scan_t *scan, char c)
{
    strbuf_ncat(scan->out, &c, 1);
}

static void
put_comment(scan_t *scan, char c)
{
    char n;

    put_char(scan, c);

    if (scan_peek(scan, 0) == '*') {
        while ((n = scan_next(scan)) != EOF) {
            put_char(scan, n);

            if (n == '*' && scan_peek(scan, 0) == '/') {
                put_char(scan, scan_next(scan));
                break;
            }
        }
    }
    else if (scan_peek(scan, 0) == '/') {
        while ((n = scan_next(scan)) != EOF) {
            put_char(scan, n);
            
            if (n == '\n' || n == '\r')
                break;
        }
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

static void
mark_file(char *path, int line, int offset, strbuf_t *out)
{
    char buf[PATH_MAX_LEN + 16];

    snprintf(buf, sizeof(buf), "#file \"%s\" %d %d\n", path, line, offset);

    strbuf_cat(out, buf);
}

static void
put_import(scan_t *scan, stack_t *imp)
{
    int offset;
    char path[PATH_MAX_LEN];
    char c, n;

    strcpy(path, scan->work_dir);
    offset = strlen(path);

    // TODO: need more error handling
    while ((c = scan_next(scan)) != EOF) {
        if (c == '"') {
            while ((n = scan_next(scan)) != EOF) {
                path[offset++] = n;

                if (n != '\\' && scan_peek(scan, 0) == '"') {
                    scan_next(scan);
                    path[offset] = '\0';

                    if (add_file(scan, path, imp)) {
                        mark_file(path, 1, 0, scan->out);
                        substitue(path, scan->work_dir, imp, scan->out);
                        mark_file(scan->path, YY_LINE + 1, YY_OFFSET, scan->out);
                        del_file(scan, scan->path, imp);
                    }
                    offset = 0;
                    break;
                }
            }
        }
        else if (c == '\n' || c == '\r') {
            break;
        }
    }
}

static void
substitue(char *path, char *work_dir, stack_t *imp, strbuf_t *out)
{
    bool is_first_ch = true;
    char c;
    scan_t scan;

    scan_init(&scan, path, work_dir, out);

    while ((c = scan_next(&scan)) != EOF) {
        if (c == '/') {
            put_comment(&scan, c);
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
                 scan_peek(&scan, 1) == 'p' &&
                 scan_peek(&scan, 2) == 'o' &&
                 scan_peek(&scan, 3) == 'r' &&
                 scan_peek(&scan, 4) == 't' &&
                 isblank(scan_peek(&scan, 5))) {
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
preprocess(char *path, flag_t flag, strbuf_t *out)
{
    stack_t imp;
    char *delim;
    char work_dir[PATH_MAX_LEN];

    stack_init(&imp);

    strcpy(work_dir, path);
    delim = strrchr(work_dir, PATH_DELIM);
    if (delim == NULL)
        work_dir[0] = '\0';
    else
        *delim = '\0';

    stack_push(&imp, path);

    substitue(path, work_dir, &imp, out);

    stack_pop(&imp);

    if (flag_on(flag, FLAG_VERBOSE)) {
        char file[PATH_MAX_LEN];

        snprintf(file, sizeof(file), "%s.i", path);
        strbuf_dump(out, file);
    }
}

/* end of prep.c */
