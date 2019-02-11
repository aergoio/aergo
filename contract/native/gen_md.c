/**
 * @file    gen_md.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_fn.h"
#include "gen_util.h"

#include "gen_md.h"

#define WASM_MEM_UNIT       65536
#define WASM_MAX_LEN        1024 * 1024

static void
table_gen(gen_t *gen, vector_t *fns)
{
    int i;
    char **names = xmalloc(sizeof(char *) * vector_size(fns));

    vector_foreach(fns, i) {
        names[i] = vector_get_fn(fns, i)->name;
    }

    BinaryenSetFunctionTable(gen->module, i, i, (const char **)names, i);
}

static void
sgmt_gen(gen_t *gen, ir_sgmt_t *sgmt)
{
    int i;
    BinaryenExpressionRef *addrs;

    if (sgmt->offset >= gen->flag.stack_size)
        FATAL(ERROR_STACK_OVERFLOW, gen->flag.stack_size / 1024, sgmt->offset);

    addrs = xmalloc(sizeof(BinaryenExpressionRef) * sgmt->size);

    for (i = 0; i < sgmt->size; i++) {
        addrs[i] = i32_gen(gen, sgmt->addrs[i]);
    }

    BinaryenSetMemory(gen->module, 1, sgmt->offset / WASM_MEM_UNIT + 1, "memory",
                      (const char **)sgmt->datas, addrs, sgmt->lens, sgmt->size, 0);

    BinaryenAddGlobal(gen->module, "stack$top", BinaryenTypeInt32(), 1,
                      i32_gen(gen, sgmt->offset));
    BinaryenAddGlobal(gen->module, "stack$max", BinaryenTypeInt32(), 0,
                      i32_gen(gen, gen->flag.stack_size));
    BinaryenAddGlobal(gen->module, "heap$offset", BinaryenTypeInt32(), 1,
                      i32_gen(gen, gen->flag.stack_size));
}

void
md_gen(gen_t *gen, ir_md_t *md)
{
    int i;

    gen->module = BinaryenModuleCreate();

    vector_foreach(&md->abis, i) {
        abi_gen(gen, vector_get_abi(&md->abis, i));
    }

    vector_foreach(&md->fns, i) {
        fn_gen(gen, vector_get_fn(&md->fns, i));
    }

    table_gen(gen, &md->fns);
    sgmt_gen(gen, &md->sgmt);

    if (is_flag_on(gen->flag, FLAG_DEBUG) || is_flag_on(gen->flag, FLAG_TEST)) {
        BinaryenSetDebugInfo(1);
    }
    else if (gen->flag.opt_lvl > 0) {
        ASSERT(BinaryenModuleValidate(gen->module));

        BinaryenSetOptimizeLevel(gen->flag.opt_lvl);
        BinaryenModuleOptimize(gen->module);
    }

    if (is_flag_on(gen->flag, FLAG_DUMP_WAT))
        BinaryenModulePrint(gen->module);

    ASSERT(BinaryenModuleValidate(gen->module));

    if (is_flag_off(gen->flag, FLAG_TEST)) {
        int n;
        int buf_size = WASM_MAX_LEN * 2;
        char *buf = xmalloc(buf_size);

        ASSERT(md->name != NULL);

        n = BinaryenModuleWrite(gen->module, buf, buf_size);
        if (n <= WASM_MAX_LEN) {
            char path[PATH_MAX_LEN + 1];

            snprintf(path, sizeof(path), "./%s.wasm", md->name);

            write_file(path, buf, n);
        }
        else {
            FATAL(ERROR_BINARY_OVERFLOW, 1, n);
        }
    }

    BinaryenModuleDispose(gen->module);

    gen->module = NULL;
}

/* end of gen_md.c */
