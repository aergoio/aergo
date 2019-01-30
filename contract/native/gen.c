/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_fn.h"
#include "gen_util.h"

#include "gen.h"

static void
gen_init(gen_t *gen, flag_t flag, ir_t *ir)
{
    gen->flag = flag;
    gen->ir = ir;

    gen->module = BinaryenModuleCreate();
    gen->relooper = NULL;

    array_init(&gen->instrs, BinaryenExpressionRef);

    gen->is_lval = false;
}

void
gen(ir_t *ir, flag_t flag, char *infile)
{
    int i;
    gen_t gen;

    if (has_error())
        return;

    gen_init(&gen, flag, ir);

    vector_foreach(&ir->abis, i) {
        abi_gen(&gen, vector_get_abi(&ir->abis, i));
    }

    vector_foreach(&ir->fns, i) {
        fn_gen(&gen, vector_get_fn(&ir->fns, i));
    }

    table_gen(&gen, &ir->fns);
    sgmt_gen(&gen, &ir->sgmt);

    ASSERT(BinaryenModuleValidate(gen.module));

    if (is_flag_on(flag, FLAG_DEBUG | FLAG_TEST)) {
        BinaryenSetDebugInfo(1);
    }
    else if (flag.opt_lvl > 0) {
        BinaryenSetOptimizeLevel(flag.opt_lvl);
        BinaryenModuleOptimize(gen.module);

        ASSERT(BinaryenModuleValidate(gen.module));
    }

    if (is_flag_on(flag, FLAG_WAT_DUMP))
        BinaryenModulePrint(gen.module);

    if (is_flag_off(flag, FLAG_TEST))
        wasm_gen(&gen, infile, flag.outfile);

    BinaryenModuleDispose(gen.module);
}

/* end of gen.c */
