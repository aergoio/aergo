/**
 * @file    preprocess.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "xutil.h"

#include "prep.h"

FILE *
preprocess(char *ipath)
{
    FILE *infp, *outfp;
    char opath[PATH_MAX_LEN + 3];

    infp = xfopen(ipath, "r");
    ASSERT(infp != NULL);

    snprintf(opath, sizeof(opath), "%s.p", ipath);

    outfp = xfopen(opath, "w+");
    ASSERT(outfp != NULL);

    // TODO: preprocess import

    return outfp;
}

/* end of preprocess.c */
