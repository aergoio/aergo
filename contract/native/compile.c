/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "parser.h"
#include "strbuf.h"

#include "compile.h"

int
compile(char *path, opt_t opt)
{
    strbuf_t res;

    strbuf_init(&res);

    preprocess(path, &res);

    return parse(path, opt, &res);
}

/* end of compile.c */
