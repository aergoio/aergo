#include <stdlib.h>
#include "lua.h"

#include "lualib.h"
#include "lauxlib.h"

// --- lua functions ---

static int get_contract_info_lua(lua_State *L) {
    const char* contract_id_hex = luaL_checkstring (L, 1);
    
    char* contract_id_base58 = (char *)CGetContractID(contract_id_hex);
    char* src_path = (char *)CGetSrc(contract_id_hex);
    
    lua_pushstring(L, contract_id_base58);
    lua_pushstring(L, src_path);
    
    free(contract_id_base58);
    free(src_path);

    return 2; //base58 encoded address, srcpath
}

static int set_breakpoint_lua(lua_State *L) {
    const char* contract_name = luaL_checkstring (L, 1);
    double line = luaL_checknumber (L, 2);

    CSetBreakPoint(contract_name, line);

    return 0;
}

static int delete_breakpoint_lua(lua_State *L) {
    const char* contract_name = luaL_checkstring (L, 1);
    double line = luaL_checknumber (L, 2);

    CDelBreakPoint(contract_name, line);

    return 0;
}


static int has_breakpoint_lua(lua_State *L) {
    const char* contract_id_hex = luaL_checkstring (L, 1);
    double line = luaL_checknumber (L, 2);

    int exist = CHasBreakPoint(contract_id_hex, line);

    lua_pushboolean(L, exist);
    
    return 1;
}

static int print_breakpoints_lua(lua_State *L) {
    PrintBreakPoints();

    return 0;
}

static int reset_breakpoints_lua(lua_State *L) {
    ResetBreakPoints();

    return 0;
}

const char* vm_set_debug_hook(lua_State *L)
{
    lua_pushcfunction(L, get_contract_info_lua);
    lua_setglobal(L, "__get_contract_info");
    lua_pushcfunction(L, set_breakpoint_lua);
    lua_setglobal(L, "__set_breakpoint");
    lua_pushcfunction(L, delete_breakpoint_lua);
    lua_setglobal(L, "__delete_breakpoint");
    lua_pushcfunction(L, has_breakpoint_lua);
    lua_setglobal(L, "__has_breakpoint");
    lua_pushcfunction(L, print_breakpoints_lua);
    lua_setglobal(L, "__print_breakpoints");
    lua_pushcfunction(L, reset_breakpoints_lua);
    lua_setglobal(L, "__reset_breakpoints");

    char* code = (char *)GetDebuggerCode();
    luaL_loadstring(L, code);
	int err = lua_pcall(L, 0, LUA_MULTRET, 0);
    free(code);
	if (err != 0) {
		return lua_tostring(L, -1);
	}
	return NULL;
}
