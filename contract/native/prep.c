/**
 * @file    prep.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "error.h"
#include "util.h"
#include "stack.h"
#include "parse.h"
#include "iobuf.h"
#include "syslib.h"

#include "prep.h"

#define YY_LINE                 prep->pos.first_line
#define YY_OFFSET               prep->pos.first_offset

#define yy_update_line()        src_pos_update_line(&prep->pos)
#define yy_update_col()         src_pos_update_col(&prep->pos, 1)

#define yy_update_first()       src_pos_update_first(&prep->pos)

static void subst(iobuf_t *src, flag_t flag, char *work_dir, stack_t *imps, ast_t *ast);

static void
prep_init(prep_t *prep, flag_t flag, iobuf_t *src, char *work_dir, ast_t *ast)
{
    prep->flag = flag;

    prep->path = iobuf_path(src);
    prep->work_dir = work_dir;

    prep->offset = 0;
    prep->src = src;

    src_pos_init(&prep->pos, iobuf_str(prep->src), prep->path);

    prep->ast = ast;
}

static char
scan_next(prep_t *prep)
{
    char c = iobuf_char(prep->src, prep->offset++);

    if (c == EOF)
        return EOF;

    yy_update_col();

    if (c == '\n' || c == '\r') {
        yy_update_line();
        yy_update_first();
    }

    return c;
}

static char
scan_peek(prep_t *prep, int cnt)
{
    return iobuf_char(prep->src, prep->offset + cnt);
}

static bool
add_file(prep_t *prep, char *path, stack_t *imps)
{
    stack_node_t *node;

    stack_foreach(node, imps) {
        if (strcmp(node->item, path) == 0) {
            ERROR(ERROR_CYCLIC_IMPORT, &prep->pos);
            return false;
        }
    }

    stack_push(imps, xstrdup(path));

    return true;
}

static void
del_file(prep_t *prep, char *path, stack_t *imps)
{
    stack_pop(imps);
}

static void
skip_char(prep_t *prep, int n)
{
    prep->offset += n;
}

static void
skip_comment(prep_t *prep)
{
    char n;

    if (scan_peek(prep, 0) == '*') {
        while ((n = scan_next(prep)) != EOF) {
            if (n == '*' && scan_peek(prep, 0) == '/') {
                scan_next(prep);
                break;
            }
        }
    }
    else if (scan_peek(prep, 0) == '/') {
        while ((n = scan_next(prep)) != EOF) {
            if (n == '\n' || n == '\r')
                break;
        }
    }
}

static void
skip_literal(prep_t *prep)
{
    char n;

    while ((n = scan_next(prep)) != EOF) {
        if (n != '\\' && scan_peek(prep, 0) == '"') {
            scan_next(prep);
            break;
        }
    }
}

static void
subst_imp(prep_t *prep, stack_t *imps)
{
    int offset;
    char path[PATH_MAX_LEN + 1];
    char c, n;

    strcpy(path, prep->work_dir);
    offset = strlen(path);

    // TODO: need more error handling
    while ((c = scan_next(prep)) != EOF) {
        if (c == '"') {
            while ((n = scan_next(prep)) != EOF) {
                path[offset++] = n;

                if (n != '\\' && scan_peek(prep, 0) == '"') {
                    scan_next(prep);
                    path[offset] = '\0';

                    if (add_file(prep, path, imps)) {
                        iobuf_t src;

                        iobuf_init(&src, path);
                        iobuf_load(&src);

                        subst(&src, prep->flag, prep->work_dir, imps, prep->ast);
                        parse(&src, prep->flag, prep->ast);

                        del_file(prep, prep->path, imps);
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
subst(iobuf_t *src, flag_t flag, char *work_dir, stack_t *imps, ast_t *ast)
{
    bool is_first_ch = true;
    char c;
    prep_t prep;

    prep_init(&prep, flag, src, work_dir, ast);

    while ((c = scan_next(&prep)) != EOF) {
        if (c == '/') {
            skip_comment(&prep);
            is_first_ch = false;
        }
        else if (c == '"') {
            skip_literal(&prep);
            is_first_ch = false;
        }
        else if (c == '\n' || c == '\r') {
            is_first_ch = true;
        }
        else if (is_first_ch && c == 'i' &&
                 scan_peek(&prep, 0) == 'm' &&
                 scan_peek(&prep, 1) == 'p' &&
                 scan_peek(&prep, 2) == 'o' &&
                 scan_peek(&prep, 3) == 'r' &&
                 scan_peek(&prep, 4) == 't' &&
                 isblank(scan_peek(&prep, 5))) {
            skip_char(&prep, 6);
            subst_imp(&prep, imps);
            is_first_ch = false;
        }
        else if (!isblank(c)) {
            is_first_ch = false;
        }
    }
}

void
prep(iobuf_t *src, flag_t flag, ast_t *ast)
{
    stack_t imps;
    char *delim;
    char work_dir[PATH_MAX_LEN];

    ASSERT(src != NULL);
    ASSERT(ast != NULL);

    syslib_load(ast);

    stack_init(&imps);

    strcpy(work_dir, iobuf_path(src));

    delim = strrchr(work_dir, PATH_DELIM);
    if (delim == NULL)
        work_dir[0] = '\0';
    else
        *delim = '\0';

    stack_push(&imps, iobuf_path(src));

    subst(src, flag, work_dir, &imps, ast);

    stack_pop(&imps);
}

/* end of prep.c */
