/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ast_blk.h"
#include "gen_id.h"
#include "gen_meta.h"
#include "gen_util.h"

#include "gen.h"

#define WASM_EXT        ".wasm"
#define WASM_MAX_LEN    1024 * 1024

static void
gen_init(gen_t *gen, BinaryenModuleRef module, ast_t *ast, flag_t flag, char *path)
{
    char *ptr;

    gen->flag = flag;
    gen->root = ast->root;
    gen->module = module;

    strcpy(gen->path, path);

    ptr = strrchr(gen->path, '.');
    if (ptr == NULL)
        strcat(gen->path, WASM_EXT);
    else
        strcpy(ptr, WASM_EXT);

    gen->dsgmt = dsgmt_new();

    gen->id_idx = 0;
    gen->ret_idx = 0;

    gen->local_cnt = 0;
    gen->locals = NULL;

    gen->instr_cnt = 0;
    gen->instrs = NULL;

    gen->buf_size = WASM_MAX_LEN * 2;
    gen->buf = xmalloc(gen->buf_size);
}

void
gen(ast_t *ast, flag_t flag, char *path)
{
    int i, n;
    gen_t gen;
    BinaryenModuleRef module;

    if (ast == NULL)
        return;

    module = BinaryenModuleCreate();

    gen_init(&gen, module, ast, flag, path);

    BinaryenSetDebugInfo(1);
    //BinaryenSetAPITracing(1);

    if (flag_on(flag, FLAG_TEST)) {
        // XXX: temporary
        //BinaryenModuleInterpret(gen.module);
    }
    else {
        for (i = 0; i < array_size(&gen.root->ids); i++) {
            id_gen(&gen, array_get(&gen.root->ids, i, ast_id_t));
        }

        BinaryenSetMemory(module, 1, gen.dsgmt->offset / UINT16_MAX + 1, "memory",
                          (const char **)gen.dsgmt->datas, gen.dsgmt->addrs, 
                          gen.dsgmt->lens, gen.dsgmt->size, 0);

        BinaryenModuleValidate(module);

        n = BinaryenModuleWrite(module, gen.buf, gen.buf_size);
        if (n <= WASM_MAX_LEN)
            write_file(gen.path, gen.buf, n);
        else
            FATAL(ERROR_BINARY_OVERFLOW, n);
    }

    BinaryenModuleDispose(module);
}

/* end of gen.c */
