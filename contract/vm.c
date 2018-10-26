#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "system_module.h"
#include "contract_module.h"
#include "db_module.h"
#include "util.h"

const char *luaExecContext= "__exec_context__";

static void preloadModules(lua_State *L)
{
	luaopen_system(L);
	luaopen_contract(L);
    luaopen_db(L);
}

static void setLuaExecContext(lua_State *L, bc_ctx_t *bc_ctx)
{
	lua_pushlightuserdata(L, bc_ctx);
	lua_setglobal(L, luaExecContext);
}

const bc_ctx_t *getLuaExecContext(lua_State *L)
{
	bc_ctx_t *exec;
	lua_getglobal(L, luaExecContext);
	exec = (bc_ctx_t *)lua_touserdata(L, -1);
	lua_pop(L, 1);

	return exec;
}

void bc_ctx_delete(bc_ctx_t *bc_ctx) {
	if (bc_ctx == NULL)
		return;
	free(bc_ctx->stateKey);
	free(bc_ctx->sender);
	free(bc_ctx->txHash);
	free(bc_ctx->contractId);
	free(bc_ctx->node);
}

lua_State *vm_newstate()
{
	lua_State *L = luaL_newstate();
	if (L == NULL)
	    return NULL;
	luaL_openlibs(L);
	preloadModules(L);
	return L;
}

const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, bc_ctx_t *bc_ctx)
{
	int err;
	const char *errMsg = NULL;

	setLuaExecContext(L, bc_ctx);

	err = luaL_loadbuffer(L, code, sz, bc_ctx->contractId);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	err = lua_pcall(L, 0, 0, 0);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	return NULL;
}

void vm_getfield(lua_State *L, const char *name)
{
	lua_getfield(L, LUA_GLOBALSINDEX, name);
}

int vm_isnil(lua_State *L, int idx)
{
	return lua_isnil(L, idx);
}

void vm_remove_construct(lua_State *L, const char *construct_name)
{
    lua_pushnil(L);
	lua_setfield(L, LUA_GLOBALSINDEX, construct_name);
}

void count_hook(lua_State *L, lua_Debug *ar)
{
    lua_pushstring(L, "exceeded the maximum instruction count");
    lua_error(L);
}

const char *vm_pcall(lua_State *L, int argc, int *nresult)
{
	int err;
	const char *errMsg = NULL;
	int nr = lua_gettop(L) - argc - 1;

    lua_sethook (L, count_hook, LUA_MASKCOUNT, 500000);

	err = lua_pcall(L, argc, LUA_MULTRET, 0);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	*nresult = lua_gettop(L) - nr;
	return NULL;
}

const char *vm_get_json_ret(lua_State *L, int nresult)
{
	int top = lua_gettop(L);
	char *json_ret = lua_util_get_json_from_stack(L, top - nresult + 1, top, true);

    if (json_ret == NULL)
        return lua_tostring(L, -1);

	lua_pushstring(L, json_ret);
	free(json_ret);
	
	return lua_tostring(L, -1);
}

const char *vm_tostring(lua_State *L, int idx)
{
    return lua_tolstring(L, idx, NULL);
}

const char *vm_copy_result(lua_State *L, lua_State *target, int cnt)
{
	int i;
	int top = lua_gettop(L);
	char *json;

	for (i = top - cnt + 1; i <= top; ++i) {
		json = lua_util_get_json (L, i, false);
		if (json == NULL)
		    return lua_tostring(L, -1);

		lua_util_json_to_lua(target, json);
		free (json);
	}
	return NULL;
}
