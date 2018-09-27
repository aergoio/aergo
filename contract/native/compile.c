/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "parse.h"
#include "check.h"
#include "gen.h"
#include "ast.h"
#include "strbuf.h"

#include "compile.h"

void
compile(char *path, flag_t flag)
{
    strbuf_t src;
    ast_t *ast = NULL;

    strbuf_init(&src);

    preprocess(path, flag, &src);
    parse(path, flag, &src, &ast);

    if (flag_on(flag, FLAG_AST_DUMP))
        ast_dump(ast);

    if (ast != NULL) {
        check(ast, flag);
        gen(ast, flag);
    }

    if (flag_off(flag, FLAG_TEST))
        error_dump();
}

/* end of compile.c */
