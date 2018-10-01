/**
 * @file    errors.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "stack.h"
#include "util.h"

#include "error.h"

char *err_lvls_[LEVEL_MAX] = {
    ANSI_RED"fatal",
    ANSI_RED"error",
    ANSI_WHITE"info",
    ANSI_YELLOW"warning",
    ANSI_BLUE"debug"
};

char *err_msgs_[ERROR_MAX] = {
    "no error",
#undef error
#define error(code, msg)    msg,
#include "error.list"
};

char *err_codes_[ERROR_MAX] = {
    "NO_ERROR",
#undef error
#define error(code, msg)    #code,
#include "error.list"
};

stack_t errstack_ = { 0, NULL };

char *
error_to_string(ec_t ec)
{
    ASSERT(ec >= 0 && ec < ERROR_MAX);
    return err_codes_[ec];
}

ec_t 
error_to_code(char *str)
{
    int i;

    for (i = 0; i < ERROR_MAX; i++) {
        if (strcmp(err_codes_[i], str) == 0)
            return i;
    }
    ASSERT(!"invalid errcode");
}

int
error_count(void)
{
    return errstack_.size;
}

ec_t
error_first(void)
{
    if (stack_empty(&errstack_))
        return NO_ERROR;

    return ((error_t *)stack_head(&errstack_)->item)->code;
}

ec_t
error_last(void)
{
    if (stack_empty(&errstack_))
        return NO_ERROR;

    return ((error_t *)stack_tail(&errstack_)->item)->code;
}

static error_t *
error_new(ec_t ec, errlvl_t lvl, errpos_t *pos, char *desc)
{
    error_t *error = xmalloc(sizeof(error_t));

    error->code = ec;
    error->level = lvl;

    if (pos == NULL)
        errpos_init(&error->pos, NULL);
    else
        error->pos = *pos;

    strcpy(error->desc, desc);

    return error;
}

void
error_push(ec_t ec, errlvl_t lvl, errpos_t *pos, ...)
{
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    va_start(vargs, pos);
    vsnprintf(errdesc, sizeof(errdesc), err_msgs_[ec], vargs);
    va_end(vargs);

    stack_push(&errstack_, error_new(ec, lvl, pos, errdesc));
}

error_t *
error_pop(void)
{
    return (error_t *)stack_pop(&errstack_);
}

void
error_clear(void)
{
    void *item;

    while ((item = stack_pop(&errstack_)) != NULL) {
        xfree(item);
    }
}

static char *
make_trace(errpos_t *pos)
{
#define TRACE_LINE_MAX      80
    int i, j;
    int nread;
    int tok_len;
    int adj_offset = pos->first_offset;
    int adj_col = pos->first_col;
    FILE *fp = open_file(pos->path, "r");
    char *buf;

    ASSERT(adj_offset >= 0);
    ASSERT(adj_col > 0);

    tok_len = MIN(pos->last_offset - pos->first_offset, TRACE_LINE_MAX - 1);
    ASSERT(tok_len >= 0);

    if (adj_col + tok_len > TRACE_LINE_MAX) {
        adj_col = TRACE_LINE_MAX - tok_len;
        adj_offset += pos->first_col - adj_col;
    }

    if (fseek(fp, adj_offset, SEEK_SET) < 0)
        FATAL(ERROR_FILE_IO, pos->path, strerror(errno));

    buf = xmalloc(TRACE_LINE_MAX * 3);

    nread = fread(buf, 1, TRACE_LINE_MAX, fp);
    if (nread <= 0 && !feof(fp))
        FATAL(ERROR_FILE_IO, pos->path, strerror(errno));

    for (i = 0; i < nread; i++) {
        if (buf[i] == '\n' || buf[i] == '\r')
            break;
    }
    buf[i++] = '\n';

    for (j = 0; j < adj_col - 1; j++) {
        buf[i++] = ' ';
    }

    strcpy(buf + i, ANSI_GREEN"^"ANSI_NONE);

    close_file(fp);

    return buf;
}

void
error_dump(void)
{
    stack_node_t *n;

    stack_foreach(n, &errstack_) {
        error_t *e = (error_t *)n->item;

        fprintf(stderr, "%s: "ANSI_NONE"%s:%d: %s\n%s\n", err_lvls_[e->level], 
                e->pos.path, e->pos.first_line, e->desc, make_trace(&e->pos));
    }
}

void
error_exit(ec_t ec, errlvl_t lvl, ...)
{
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    va_start(vargs, lvl);
    vsnprintf(errdesc, sizeof(errdesc), err_msgs_[ec], vargs);
    va_end(vargs);

    fprintf(stderr, "%s: "ANSI_NONE"%s\n", err_lvls_[lvl], errdesc);

    exit(EXIT_FAILURE);
}

void
assert_exit(char *cond, char *file, int line, int argc, ...)
{
    int i;
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    snprintf(errdesc, sizeof(errdesc), 
             "%s:%d: internal error with condition '%s'", file, line, cond);

    fprintf(stderr, "%s: "ANSI_NONE"%s\n", err_lvls_[LVL_FATAL], errdesc);

    va_start(vargs, argc);

    for (i = 0; i < argc; i++) {
        int size;
        char *name;
        char c;
        uint32_t i32;
        uint64_t i64;

        name = va_arg(vargs, char *);
        size = va_arg(vargs, int);

        fprintf(stderr, "    %s = ", name);

        switch (size) {
        case 1:
            c = (char)va_arg(vargs, int);
            fprintf(stderr, "'%c' = 0x%x\n", c, c);
            break;
        case 2:
        case 4:
            i32 = va_arg(vargs, uint32_t);
            fprintf(stderr, "%d = %u = 0x%x\n", i32, i32, i32);
            break;
        case 8:
            i64 = va_arg(vargs, uint64_t);
            fprintf(stderr, "%"PRId64" = %"PRIu64" = 0x%"PRIx64"\n", 
                    (int64_t)i64, i64, i64);
            break;
        default:
            break;
        }
    }

    va_end(vargs);

    exit(EXIT_FAILURE);
}

/* end of errors.c */
