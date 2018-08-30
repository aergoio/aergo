/**
 * @file    errors.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "errors.h"

char *errmsgs_[ERROR_MAX] = {
    "no error",
#undef error
#define error(code, msg)    msg,
#include "errors.list"
}; 

/* end of errors.c */
