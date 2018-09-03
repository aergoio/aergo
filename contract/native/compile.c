/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"
#include "prep.h"
#include "parser.h"

#include "compile.h"

int
compile(char *path, opt_t opt)
{
    volatile ec_t ec = NO_ERROR;
    strbuf_t sb;

    strbuf_init(&sb);

    ec = preprocess(path, opt, &sb);
    if (ec == NO_ERROR)
        ec = parse(strbuf_text(&sb), strbuf_length(&sb), opt);

    return ec;
}

/* end of compile.c */
