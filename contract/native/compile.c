/**
 * @file    compile.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast.h"
#include "prep.h"
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
    ast_t *ast = NULL;

    ASSERT(path != NULL);

    parse(path, flag, &ast);

    /* empty contract can be null */
    if (ast != NULL) {
        ir_t *ir = NULL;

        prep(ast, flag, path);
        check(ast, flag);

        trans(ast, flag, &ir);

        gen(ir, flag, path);
    }

    if (is_flag_off(flag, FLAG_TEST))
        error_print();

    return has_error();
}

/* end of compile.c */
