/**
 * @file    preprocess.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "xutil.h"

#include "preprocess.h"

FILE *
preprocess(char *path)
{
    FILE *fp;

    fp = xfopen(path, "r");
    ASSERT(fp != NULL);

    // TODO: preprocess import

    return fp;
}

/* end of preprocess.c */
