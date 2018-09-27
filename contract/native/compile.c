/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "parse.h"
#include "check.h"
#include "ast.h"
#include "strbuf.h"

#include "compile.h"

void
compile(char *path, flag_t flag)
{
    strbuf_t src;
    list_t blk_l;

    strbuf_init(&src);
    list_init(&blk_l);

    preprocess(path, flag, &src);

    parse(path, flag, &src, &blk_l);

    check(&blk_l, flag);

    if (flag_off(flag, FLAG_SILENT))
        error_dump();
}

/* end of compile.c */
