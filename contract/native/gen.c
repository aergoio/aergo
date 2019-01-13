/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ir.h"
#include "ir_abi.h"
#include "ir_sgmt.h"
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

    gen->local_cnt = 0;
    gen->locals = NULL;

    gen->instr_cnt = 0;
    gen->instrs = NULL;
}

static void
table_gen(gen_t *gen, array_t *fns)
{
    int i;
    char **names = xmalloc(sizeof(char *) * array_size(fns));

    array_foreach(fns, i) {
        names[i] = array_get_fn(fns, i)->name;
    }

    BinaryenSetFunctionTable(gen->module, i, i, (const char **)names, i);
}

static void
sgmt_gen(gen_t *gen, ir_sgmt_t *sgmt)
{
    int i;
    int stack_size = UINT16_MAX;
    BinaryenExpressionRef *addrs = xmalloc(sizeof(BinaryenExpressionRef) * sgmt->size);

    for (i = 0; i < sgmt->size; i++) {
        addrs[i] = i32_gen(gen, sgmt->addrs[i]);
    }

    BinaryenSetMemory(gen->module, 1, sgmt->offset / UINT16_MAX + 1, "memory",
                      (const char **)sgmt->datas, addrs, sgmt->lens, sgmt->size, 0);

    BinaryenAddGlobal(gen->module, "stack$high", BinaryenTypeInt32(), 1,
                      i32_gen(gen, stack_size));

    BinaryenAddGlobal(gen->module, "stack$low", BinaryenTypeInt32(), 0,
                      i32_gen(gen, ALIGN64(sgmt->offset)));

    BinaryenAddGlobal(gen->module, "heap$offset", BinaryenTypeInt32(), 1,
                      i32_gen(gen, stack_size + 1));
}

void
gen(ir_t *ir, flag_t flag, char *path)
{
    int i, n;
    gen_t gen;

    if (has_error())
        return;

    gen_init(&gen, flag, path);

    BinaryenSetDebugInfo(1);

    array_foreach(&ir->abis, i) {
        abi_gen(&gen, array_get_abi(&ir->abis, i));
    }

    array_foreach(&ir->fns, i) {
        fn_gen(&gen, array_get_fn(&ir->fns, i));
    }

    table_gen(&gen, &ir->fns);

    sgmt_gen(&gen, &ir->sgmt);

    if (flag_on(flag, FLAG_WAT_DUMP))
        BinaryenModulePrint(gen.module);

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

static BinaryenType
type_gen(type_t type)
{
    switch (type) {
    case TYPE_NONE:
    case TYPE_VOID:
        return BinaryenTypeNone();

    case TYPE_BOOL:
    case TYPE_BYTE:
    case TYPE_INT8:
    case TYPE_INT16:
    case TYPE_INT32:
    case TYPE_UINT8:
    case TYPE_UINT16:
    case TYPE_UINT32:
        return BinaryenTypeInt32();

    case TYPE_INT64:
    case TYPE_UINT64:
        return BinaryenTypeInt64();

    case TYPE_FLOAT:
        return BinaryenTypeFloat32();

    case TYPE_DOUBLE:
        return BinaryenTypeFloat64();

    case TYPE_STRING:
    case TYPE_ACCOUNT:
    case TYPE_STRUCT:
    case TYPE_MAP:
    case TYPE_OBJECT:
        return BinaryenTypeInt32();

    case TYPE_TUPLE:
    default:
        ASSERT1(!"invalid type", type);
    }

    return BinaryenTypeUnreachable();
}

BinaryenType
meta_gen(meta_t *meta)
{
    if (is_array_type(meta))
        return BinaryenTypeInt32();

    return type_gen(meta->type);
}

void
local_add(gen_t *gen, type_t type)
{
    /* TODO: need to improve */
    if (gen->locals == NULL)
        gen->locals = xmalloc(sizeof(BinaryenType));
    else
        gen->locals = xrealloc(gen->locals, sizeof(BinaryenType) * (gen->local_cnt + 1));

    gen->locals[gen->local_cnt++] = type_gen(type);
}

void
instr_add(gen_t *gen, BinaryenExpressionRef instr)
{
    if (instr == NULL)
        return;

    /* TODO: need to improve */
    if (gen->instrs == NULL)
        gen->instrs = xmalloc(sizeof(BinaryenExpressionRef));
    else
        gen->instrs = xrealloc(gen->instrs,
                               sizeof(BinaryenExpressionRef) * (gen->instr_cnt + 1));

    gen->instrs[gen->instr_cnt++] = instr;
}

/* end of gen.c */
