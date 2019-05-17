/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "error.h"
#include "ast.h"
#include "prep.h"
#include "parse.h"
#include "check.h"
#include "ir.h"
#include "trans.h"
#include "gen.h"
#include "iobuf.h"

#include "compile.h"

int
compile(flag_t flag)
{
    iobuf_t src;
    ast_t *ast = ast_new();
    ir_t *ir = ir_new();

    ASSERT(flag.path != NULL);

    iobuf_init(&src, flag.path);
    iobuf_load(&src);

    prep(&src, flag, ast);
    parse(&src, flag, ast);

    check(ast, flag);
    trans(ast, flag, ir);

    gen(ir, flag);

    if (has_error()) {
        error_print();
        return EXIT_FAILURE;
    }

    return EXIT_SUCCESS;
}

/* end of compile.c */
