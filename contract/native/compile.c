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
    ast_t *ast = ast_new();
    ir_t *ir = ir_new();

    ASSERT(path != NULL);

    prep(path, flag, ast);
    parse(path, flag, ast);

    check(ast, flag);
    trans(ast, flag, ir);

    gen(ir, flag, path);

    if (is_flag_off(flag, FLAG_TEST))
        error_print();

    return has_error();
}

/* end of compile.c */
