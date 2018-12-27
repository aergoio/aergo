/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ir.h"
#include "gen_id.h"
#include "gen_fn.h"

#include "gen.h"

#define WASM_EXT        ".wasm"
#define WASM_MAX_LEN    1024 * 1024

static void
gen_init(gen_t *gen, flag_t flag, char *path)
{
    char *ptr;

    gen->flag = flag;

    gen->module = BinaryenModuleCreate();
    gen->relooper = NULL;

    strcpy(gen->path, path);

    ptr = strrchr(gen->path, '.');
    if (ptr == NULL)
        strcat(gen->path, WASM_EXT);
    else
        strcpy(ptr, WASM_EXT);

    gen->dsgmt = dsgmt_new();

    gen->local_cnt = 0;
    gen->locals = NULL;

    gen->instr_cnt = 0;
    gen->instrs = NULL;
}

void
gen(ir_t *ir, flag_t flag, char *path)
{
    int i, n;
    gen_t gen;

    if (ir == NULL || has_error())
        return;

    gen_init(&gen, flag, path);

    BinaryenSetDebugInfo(1);
    //BinaryenSetAPITracing(1);

    for (i = 0; i < array_size(&ir->globals); i++) {
        id_gen(&gen, array_get_id(&ir->globals, i));
    }

    for (i = 0; i < array_size(&ir->fns); i++) {
        fn_gen(&gen, array_get_fn(&ir->fns, i));
    }

    BinaryenSetMemory(gen.module, 1, gen.dsgmt->offset / UINT16_MAX + 1, "memory",
                      (const char **)gen.dsgmt->datas, gen.dsgmt->addrs,
                      gen.dsgmt->lens, gen.dsgmt->size, 0);

    ASSERT(BinaryenModuleValidate(gen.module));

    if (flag_on(flag, FLAG_TEST)) {
        // XXX: temporary
        //BinaryenModuleInterpret(gen.module);
    }
    else {
        int buf_size = WASM_MAX_LEN * 2;
        char *buf = xmalloc(buf_size);

        n = BinaryenModuleWrite(gen.module, buf, buf_size);
        if (n <= WASM_MAX_LEN)
            write_file(path, buf, n);
        else
            FATAL(ERROR_BINARY_OVERFLOW, n);
    }

    BinaryenModuleDispose(gen.module);
}

/* end of gen.c */
