#ifndef _VM_H
#define _VM_H

#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

typedef struct blockchain_ctx {
    char *sender;
    char *contractId;
    char *blockHash;
    char *txHash;
    unsigned long long blockHeight;
    long long timestamp;
    char *node;
    int32_t confirmed;
} bc_ctx_t;

void vm_getfield(lua_State *L, const char *name);
const char *vm_loadbuff(const char *code, size_t sz, const char *name, bc_ctx_t *bc_ctx, lua_State **p);
const char *vm_pcall(lua_State *L, int argc);
const bc_ctx_t *getLuaExecContext(lua_State *L);

#endif /* _VM_H */
