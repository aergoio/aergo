/**
 * @file    errors.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "stack.h"
#include "util.h"

#include "error.h"

char *err_lvls_[LVL_MAX] = {
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
    int errcnt = 0;
    stack_node_t *n;

    stack_foreach(n, &errstack_) {
        if (((error_t *)n->item)->level <= LVL_INFO)
            errcnt++;
    }

    return errcnt;
}

ec_t
error_first(void)
{
    stack_node_t *n;

    stack_foreach(n, &errstack_) {
        error_t *e = (error_t *)n->item;

        if (e->level <= LVL_INFO)
            return e->code;
    }

    return NO_ERROR;
}

static error_t *
error_new(ec_t ec, errlvl_t lvl, trace_t *trc, char *desc)
{
    error_t *error = xmalloc(sizeof(error_t));

    ASSERT(trc != NULL);

    error->code = ec;
    error->level = lvl;
    error->trc = *trc;
    strcpy(error->desc, desc);

    return error;
}

void
error_push(ec_t ec, errlvl_t lvl, trace_t *trc, ...)
{
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    va_start(vargs, trc);
    vsnprintf(errdesc, sizeof(errdesc), err_msgs_[ec], vargs);
    va_end(vargs);

    stack_push(&errstack_, error_new(ec, lvl, trc, errdesc));
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

void
error_dump(void)
{
    int i;
    array_t *array = stack_to_array(&errstack_, error_cmp);

    for (i = 0; i < array_size(array); i++) {
        error_t *e = array_item(array, i, error_t);

        ASSERT1(e->level >= LVL_FATAL && e->level < LVL_MAX, e->level); 
        ASSERT(trace_rel_path(&e->trc) != NULL);

        fprintf(stderr, "%s: "ANSI_NONE"%s:%d: %s\n", err_lvls_[e->level], 
                trace_rel_path(&e->trc), trace_rel_line(&e->trc), e->desc);

        trace_dump(&e->trc);
    }

    array_clear(array);
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

/* end of errors.c */
