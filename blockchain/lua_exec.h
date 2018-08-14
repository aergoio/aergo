#include <stdint.h>
#include <sys/time.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

struct exec_context {
    char *sender;
    char *blockHash;
    char *txHash;
    int32_t blockHeight;
    struct timeval timestamp;
    char *node;
    int32_t confirmed;
};

lua_State *new_lstate(uint64_t max_inst_count);
void preloadSystem(lua_State *);
const char* vm_run(lua_State *L, const char *code, size_t sz, const char *name);
void setLuaExecContext(lua_State *L, struct exec_context *exec);