/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "ast.h"
#include "parse.h"
#include "check.h"
#include "ir.h"
#include "trans.h"
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

    /* empty contract can be null */
    if (ast != NULL) {
        ir_t *ir = NULL;

        check(ast, flag);
        trans(ast, flag, &ir);

        gen(ir, flag, path);
    }

    if (flag_off(flag, FLAG_TEST))
        error_print();

    return has_error();
}

/* end of compile.c */