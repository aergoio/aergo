/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ir.h"
#include "gen_fn.h"

#include "gen.h"

#define WASM_EXT        ".wasm"
#define WASM_MAX_LEN    1024 * 1024

static void
gen_init(gen_t *gen, BinaryenModuleRef module, flag_t flag, char *path)
{
    char *ptr;

    gen->flag = flag;

    gen->module = module;
    gen->relooper = NULL;

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

    gen->instr_cnt = 0;
    gen->instrs = NULL;

    gen->buf_size = WASM_MAX_LEN * 2;
    gen->buf = xmalloc(gen->buf_size);
}

void
gen(ir_t *ir, flag_t flag, char *path)
{
    int i, n;
    gen_t gen;
    BinaryenModuleRef module;

    if (ir == NULL)
        return;

    module = BinaryenModuleCreate();

    gen_init(&gen, module, flag, path);

    BinaryenSetDebugInfo(1);
    //BinaryenSetAPITracing(1);

    if (flag_on(flag, FLAG_TEST)) {
        // XXX: temporary
        //BinaryenModuleInterpret(gen.module);
    }
    else {
        // XXX: handle globals

        for (i = 0; i < array_size(&ir->fns); i++) {
            fn_gen(&gen, array_get(&ir->fns, i, ir_fn_t));
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
