/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "parser.h"
#include "ast.h"
#include "strbuf.h"

#include "compile.h"

int
compile(char *path, opt_t opt)
{
    strbuf_t src;

    strbuf_init(&src);

    preprocess(path, &src);

    return parse(path, opt, &src);
}

/* end of compile.c */
