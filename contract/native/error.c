/**
 * @file    errors.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "stack.h"

#include "error.h"

char *errlvls_[LEVEL_MAX] = {
    ANSI_RED"fatal",
    ANSI_RED"error",
    ANSI_WHITE"info",
    ANSI_YELLOW"warning",
    ANSI_BLUE"debug"
};

char *errmsgs_[ERROR_MAX] = {
    "no error",
#undef error
#define error(code, msg)    msg,
#include "error.list"
};

char *errstrs_[ERROR_MAX] = {
    "NO_ERROR",
#undef error
#define error(code, msg)    #code,
#include "error.list"
};

//stack_t errstack_ = { 0, NULL };
ec_t top_ = NO_ERROR;

extern opt_t opt_;

char *
error_text(ec_t ec)
{
    return errstrs_[ec];
}

ec_t
error_top(void)
{
    return top_;
}

void
error_clear(void)
{
    top_ = NO_ERROR;
}

/*
int
error_count(void)
{
    return errstack_.size;
}

ec_t
error_first(void)
{
    stack_node_t *n;

    if (stack_empty(&errstack_))
        return NO_ERROR;

    n = stack_top(&errstack_);
    while (n->next != NULL) {
        n = n->next;
    }

    return ((error_t *)n->item)->code;
}

ec_t
error_last(void)
{
    if (stack_empty(&errstack_))
        return NO_ERROR;

    return ((error_t *)stack_top(&errstack_)->item)->code;
}

static error_t *
error_new(ec_t ec, lvl_t lvl, char *desc)
{
    error_t *error = xmalloc(sizeof(error_t));

    error->code = ec;
    error->level = lvl;
    strcpy(error->desc, desc);

    return error;
}

void
error_push(ec_t ec, lvl_t lvl, ...)
{
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    va_start(vargs, lvl);
    vsnprintf(errdesc, sizeof(errdesc), errmsgs_[ec], vargs);
    va_end(vargs);

    stack_push(&errstack_, error_new(ec, lvl, errdesc));
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
    stack_node_t *n = stack_top(&errstack_);

    while (n != NULL) {
        error_t *e = (error_t *)n->item;
        fprintf(stderr, "%s: "ANSI_NONE"%s\n", errlvls_[e->level], e->desc);
        n = n->next;
    }
}
*/

void
error_dump(ec_t ec, lvl_t lvl, ...)
{
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    if (top_ == NO_ERROR)
        top_ = ec;

    if (opt_enabled(opt_, OPT_SILENT))
        return;

    va_start(vargs, lvl);
    vsnprintf(errdesc, sizeof(errdesc), errmsgs_[ec], vargs);
    va_end(vargs);

    fprintf(stderr, "%s: "ANSI_NONE"%s\n", errlvls_[lvl], errdesc);
}

/* end of errors.c */
