/**
 * @file    gen_util.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_fn.h"
#include "ir_sgmt.h"

#include "gen_util.h"

BinaryenType
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

void
table_gen(gen_t *gen, vector_t *fns)
{
    int i;
    char **names = xmalloc(sizeof(char *) * vector_size(fns));

    vector_foreach(fns, i) {
        names[i] = vector_get_fn(fns, i)->name;
    }

    BinaryenSetFunctionTable(gen->module, i, i, (const char **)names, i);
}

void
sgmt_gen(gen_t *gen, ir_sgmt_t *sgmt)
{
    int i;
    BinaryenExpressionRef *addrs = xmalloc(sizeof(BinaryenExpressionRef) * sgmt->size);

    for (i = 0; i < sgmt->size; i++) {
        addrs[i] = i32_gen(gen, sgmt->addrs[i]);
    }

    BinaryenSetMemory(gen->module, 1, sgmt->offset / UINT16_MAX + 1, "memory",
                      (const char **)sgmt->datas, addrs, sgmt->lens, sgmt->size, 0);

    BinaryenAddGlobal(gen->module, "stack$offset", BinaryenTypeInt32(), 1,
                      i32_gen(gen, STACK_SIZE - 1));
    BinaryenAddGlobal(gen->module, "heap$offset", BinaryenTypeInt32(), 1,
                      i32_gen(gen, STACK_SIZE));
}

void
wasm_gen(gen_t *gen, char *infile, char *outfile)
{
    int n;
    int buf_size = WASM_MAX_LEN * 2;
    char *buf = xmalloc(buf_size);

    n = BinaryenModuleWrite(gen->module, buf, buf_size);
    if (n <= WASM_MAX_LEN) {
        char *ptr;
        char path[PATH_MAX_LEN + 5];

        if (outfile == NULL || outfile[0] == '\0') {
            strcpy(path, infile);

            ptr = strrchr(path, '.');
            if (ptr == NULL)
                strcat(path, WASM_EXT);
            else
                strcpy(ptr, WASM_EXT);

            outfile = path;
        }

        write_file(outfile, buf, n);
    }
    else {
        FATAL(ERROR_BINARY_OVERFLOW, n);
    }
}

void
malloc_gen(gen_t *gen)
{
    BinaryenType type = BinaryenTypeInt32();
    BinaryenType params[] = { BinaryenTypeInt32() };
    BinaryenType locals[] = { BinaryenTypeInt32() };
    BinaryenFunctionTypeRef spec;
    BinaryenExpressionRef instrs[3];
    BinaryenModuleRef module = gen->module;

    BinaryenAddGlobal(module, "heap$offset", type, 1, i32_gen(gen, STACK_SIZE));

    spec = BinaryenAddFunctionType(module, "system$malloc", type, params, 1);

    instrs[0] = BinaryenSetLocal(module, 1,
                                 BinaryenGetGlobal(module, "heap$offset", type));

    instrs[1] = BinaryenSetGlobal(module, "heap$offset",
                                  BinaryenBinary(module, BinaryenAddInt32(),
                                                 BinaryenGetLocal(module, 1, type),
                                                 BinaryenGetLocal(module, 0, type)));

    instrs[2] = BinaryenReturn(module, BinaryenGetLocal(module, 1, type));

    BinaryenAddFunction(module, "system$malloc", spec, locals, 1,
                        BinaryenBlock(module, NULL, instrs, 3, BinaryenTypeInt32()));
}

/* end of gen_util.c */
