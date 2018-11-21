/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ast_blk.h"
#include "gen_id.h"

#include "gen.h"

#define WASM_EXT        ".wasm"
#define WASM_MAX_LEN    1024 * 1024

static void
gen_init(gen_t *gen, ast_t *ast, flag_t flag, char *path)
{
    char *ptr;

    gen->flag = flag;
    gen->root = ast->root;
    gen->module = BinaryenModuleCreate();

    strcpy(gen->path, path);

    ptr = strrchr(gen->path, '.');
    if (ptr == NULL)
        strcat(gen->path, WASM_EXT);
    else
        strcpy(ptr, WASM_EXT);

    gen->buf_size = WASM_MAX_LEN + 1;
    gen->buf = xmalloc(gen->buf_size);
}

static void
gen_reset(gen_t *gen)
{
    BinaryenModuleDispose(gen->module);
}

void
gen(ast_t *ast, flag_t flag, char *path)
{
    //int i, n;
    gen_t gen;

    if (ast == NULL)
        return;

    gen_init(&gen, ast, flag, path);

    /*
    for (i = 0; i < array_size(&gen.root->ids); i++) {
        id_gen(&gen, array_item(&gen.root->ids, i, ast_id_t));
    }

    BinaryenModuleValidate(gen.module);

    if (flag_on(flag, FLAG_TEST)) {
        BinaryenModuleInterpret(gen.module);
    }
    else {
        BinaryenSetDebugInfo(1);
        n = BinaryenModuleWrite(gen.module, gen.buf, gen.buf_size);
        if (n < gen.buf_size)
            write_file(gen.path, gen.buf, n);
        else
            FATAL(ERROR_BINARY_OVERFLOW);
    }
    */

    gen_reset(&gen);
}

/* end of gen.c */
