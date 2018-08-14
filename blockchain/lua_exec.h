#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

#define uint64_t unsigned long long
#define int64_t long long

typedef struct blockchain_ctx {
    char *sender;
    char *contractId;
    char *blockHash;
    char *txHash;
    uint64_t blockHeight;
    int64_t timestamp;
    char *node;
    int32_t confirmed;
} bc_ctx_t;

void vm_getfield(lua_State *L, const char *name);
const char *vm_loadbuff(const char *code, size_t sz, const char *name, bc_ctx_t *bc_ctx, lua_State **p);
const char *vm_pcall(lua_State *L, int argc);
