/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "ast.h"
#include "parse.h"
#include "check.h"
#include "gen.h"
#include "strbuf.h"

#include "compile.h"

int
compile(char *path, flag_t flag)
{
    strbuf_t src;
    ast_t *ast = NULL;

    strbuf_init(&src);
    preprocess(path, flag, &src);

    parse(path, flag, &src, &ast);

    if (ast != NULL) {
        /* empty contract can be null */
        check(ast, flag);

        if (is_no_error())
            gen(ast, flag, path);
    }

    if (flag_off(flag, FLAG_TEST))
        error_dump();

    return is_no_error() ? EXIT_SUCCESS : EXIT_FAILURE;
}

/* end of compile.c */
