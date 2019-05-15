#include "lua.h"
#include "vm.h"
#include <stdbool.h>

char *lua_util_get_json (lua_State *L, int idx, bool json_form);
int lua_util_json_to_lua (lua_State *L, char *json, bool check);
char *lua_util_get_json_from_stack (lua_State *L, int start, int end, bool json_form);
void minus_inst_count(lua_State *L, int count);

int luaopen_json(lua_State *L);

#define strPushAndRelease(L,s) \
    do { \
        lua_pushstring((L), (s)); \
        free((s)); \
    } while(0)

