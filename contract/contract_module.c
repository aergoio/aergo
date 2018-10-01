#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern const bc_ctx_t *getLuaExecContext(lua_State *L);

static int moduleCall(lua_State *L)
{
    char *contract;
    char *fname;
    char *json_args;
    int ret;
    bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);

    if (exec == NULL) {
        luaL_error(L, "cannot find execution context");
    }

    if (exec->isQuery) {
        luaL_error(L, "not permitted set in query");
    }

    contract = (char *)luaL_checkstring(L, 1);
    fname = (char *)luaL_checkstring(L, 2);
    json_args = lua_util_get_json_from_stack (L, 3, lua_gettop(L));
    if ((ret = LuaCallContract(L, exec, contract, fname, json_args)) < 0) {
        free(json_args);
        lua_error(L);
    }
    free(json_args);

	return ret;
}

static int moduleDelegateCall(lua_State *L)
{
    char *contract;
    char *fname;
    char *json_args;
    int ret;
    bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);

    if (exec == NULL) {
        luaL_error(L, "cannot find execution context");
    }

    if (exec->isQuery) {
        luaL_error(L, "not permitted set in query");
    }

    contract = (char *)luaL_checkstring(L, 1);
    fname = (char *)luaL_checkstring(L, 2);
    json_args = lua_util_get_json_from_stack (L, 3, lua_gettop(L));
    if ((ret = LuaDelegateCallContract(L, exec, contract, fname, json_args)) < 0) {
        free(json_args);
        lua_error(L);
    }
    free(json_args);

	return ret;
}

static const luaL_Reg contract_lib[] = {
	{"call", moduleCall},
	{"delegatecall", moduleDelegateCall},
	{NULL, NULL}
};

int luaopen_contract(lua_State *L)
{
	luaL_register(L, "contract", contract_lib);
	return 1;
}