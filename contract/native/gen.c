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

    gen->dsgmt = dsgmt_new();
    gen->id_idx = 0;

    gen->local_cnt = 0;
    gen->locals = NULL;

    gen->buf_size = WASM_MAX_LEN * 2;
    gen->buf = xmalloc(gen->buf_size);
}

static void
gen_reset(gen_t *gen)
{
    BinaryenModuleDispose(gen->module);
}

static void
mem_gen(gen_t *gen, dsgmt_t *dsgmt)
{
    int i;
    BinaryenExpressionRef *offsets;

    offsets = xmalloc(sizeof(BinaryenExpressionRef) * dsgmt->size);

    for (i = 0; i < dsgmt->size; i++) {
        offsets[i] = BinaryenConst(gen->module, BinaryenLiteralInt32(dsgmt->addrs[i]));
    }

    BinaryenSetMemory(gen->module, 1, dsgmt->offset, "memory",
                      (const char **)dsgmt->datas, offsets, dsgmt->lens, dsgmt->size, 0);
}

void
gen(ast_t *ast, flag_t flag, char *path)
{
    int i, n;
    gen_t gen;

    if (ast == NULL)
        return;

    gen_init(&gen, ast, flag, path);

    if (flag_on(flag, FLAG_TEST)) {
        // XXX: temporary
        //BinaryenModuleInterpret(gen.module);
    }
    else {
        for (i = 0; i < array_size(&gen.root->ids); i++) {
            id_gen(&gen, array_get(&gen.root->ids, i, ast_id_t));
        }

        mem_gen(&gen, gen.dsgmt);

        BinaryenModuleValidate(gen.module);
        BinaryenSetDebugInfo(1);

        n = BinaryenModuleWrite(gen.module, gen.buf, gen.buf_size);
        if (n <= WASM_MAX_LEN)
            write_file(gen.path, gen.buf, n);
        else
            FATAL(ERROR_BINARY_OVERFLOW, n);
    }

    gen_reset(&gen);
}

/* end of gen.c */
