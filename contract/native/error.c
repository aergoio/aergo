/**
 * @file    errors.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "option.h"

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

error_t head_ = { NULL, NO_ERROR, LVL_FATAL, { '\0' } };

int
error_count(void)
{
    int count = 0;
    error_t *e = head_.next;

    while (e != NULL) {
        count++;
    }

    return count;
}

ec_t 
error_last(void)
{
    error_t *e = head_.next;

    if (e == NULL)
        return NO_ERROR;

    while (e->next != NULL) {
        e = e->next;
    }

    return e->code;
}

char *
error_text(ec_t ec)
{
    return errstrs_[ec];
}

static error_t *
error_new(ec_t ec, lvl_t lvl, char *desc)
{
    error_t *error;

    error = xmalloc(sizeof(error_t));

    error->next = NULL;
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
    error_t *e = &head_;

    va_start(vargs, lvl);
    vsnprintf(errdesc, sizeof(errdesc), errmsgs_[ec], vargs);
    va_end(vargs);

    while (e->next != NULL) {
        e = e->next;
    }

    e->next = error_new(ec, lvl, errdesc);
}

error_t *
error_pop(void)
{
    error_t *e = head_.next;

    if (e == NULL)
        return NULL;

    head_.next = e->next;

    return e;
}

void
error_clear(void)
{
    error_t *e;

    while ((e = error_pop()) != NULL) {
        xfree(e);
    }

    head_.next = NULL;
}

void
error_dump(void)
{
    error_t *e = head_.next;

    while (e != NULL) {
        fprintf(stderr, "%s: "ANSI_NONE"%s\n", errlvls_[e->level], e->desc);
        e = e->next;
    }
}

/* end of errors.c */
