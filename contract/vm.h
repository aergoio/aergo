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

lua_State *vm_newstate();
void vm_getfield(lua_State *L, const char *name);
const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, const char *name, bc_ctx_t *bc_ctx);
const char *vm_pcall(lua_State *L, int argc, int* nresult);
const char *vm_get_json_ret(lua_State *L, int nresult);
const char *vm_compile(lua_State *L, const char *lua_filename, const char *bc_filename, const char *abi_filename);

#endif /* _VM_H */
