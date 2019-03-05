/**
 * @file    gen_md.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "gen_fn.h"
#include "gen_util.h"

#include "gen_md.h"

#define WASM_MEM_UNIT       65536
#define WASM_MAX_LEN        1024 * 1024

static void
import_gen(gen_t *gen, vector_t *abis)
{
    int i;
    char qname[NAME_MAX_LEN * 2 + 2];

    vector_foreach(abis, i) {
        ir_abi_t *abi = vector_get_abi(abis, i);

        snprintf(qname, sizeof(qname), "%s.%s", abi->module, abi->name);

        BinaryenAddFunctionImport(gen->module, qname, abi->module, abi->name,
                                  abi_gen(gen, abi));
    }
}

static void
sgmt_gen(gen_t *gen, ir_sgmt_t *sgmt)
{
    int i;
    BinaryenExpressionRef *addrs;

    if (sgmt->offset >= gen->flag.stack_size)
        FATAL(ERROR_STACK_OVERFLOW, gen->flag.stack_size, sgmt->offset);

    addrs = xmalloc(sizeof(BinaryenExpressionRef) * sgmt->size);

    for (i = 0; i < sgmt->size; i++) {
        addrs[i] = i32_gen(gen, sgmt->addrs[i]);
    }

    BinaryenSetMemory(gen->module, 1, sgmt->offset / WASM_MEM_UNIT + 1, "memory",
                      (const char **)sgmt->datas, addrs, sgmt->lens, sgmt->size, 0);

    BinaryenAddGlobal(gen->module, "stack_top", BinaryenTypeInt32(), 1,
                      i32_gen(gen, ALIGN64(sgmt->offset)));
    BinaryenAddGlobal(gen->module, "stack_max", BinaryenTypeInt32(), 0,
                      i32_gen(gen, gen->flag.stack_size));
}

void
md_gen(gen_t *gen, ir_md_t *md)
{
    int i;

    gen->module = BinaryenModuleCreate();

    import_gen(gen, &md->abis);

    vector_foreach(&md->fns, i) {
        fn_gen(gen, vector_get_fn(&md->fns, i));
    }

    sgmt_gen(gen, &md->sgmt);

    if (is_flag_on(gen->flag, FLAG_DEBUG)) {
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
            FATAL(ERROR_BINARY_OVERFLOW, WASM_MAX_LEN, n);
        }
    }

    BinaryenModuleDispose(gen->module);

    gen->module = NULL;
}

/* end of gen_md.c */
