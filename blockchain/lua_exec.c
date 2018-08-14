#include <string.h>
#include <stdlib.h>
#include "lua_exec.h"

#define MAX_SYSTEM_FNS 11

const char *luaExecContext= "__exec_context__";

static int systemPrint(lua_State *L);
static int setItem(lua_State *L);
static int getItem(lua_State *L);
static int getSender(lua_State *L);
static int getContractID(lua_State *L);
static int getBlockhash(lua_State *L);
static int getTxhash(lua_State *L);
static int getBlockHeight(lua_State *L);
static int getTimestamp(lua_State *L);

static const struct luaL_Reg sys_lib[MAX_SYSTEM_FNS] = {
	{"print", systemPrint},
	{"setItem", setItem},
	{"getItem", getItem},
	{"getSender", getSender},
	{"getCreator", getContractID},
	{"getBlockhash", getBlockhash},
	{"getTxhash", getTxhash},
	{"getBlockheight", getBlockHeight},
	{"getTimestamp", getTimestamp},
	{"getContractID", getContractID},
	{NULL, NULL}
};

lua_State *new_lstate(uint64_t max_inst_count)
{
	lua_State *L = luaL_newstate();
	return L;
}

void preloadSystem(lua_State *L)
{
	luaL_openlibs(L);
	luaL_register(L, "system", sys_lib);
}

static void setLuaExecContext(lua_State *L, bc_ctx_t *bc_ctx)
{
	lua_pushlightuserdata(L, bc_ctx);
	lua_setglobal(L, luaExecContext);
}

static bc_ctx_t* getLuaExecContext(lua_State *L)
{
	bc_ctx_t *exec;
	lua_getglobal(L, luaExecContext);
	exec = (bc_ctx_t *)lua_touserdata(L, 1);
	lua_pop(L, 1);

	return exec;
}
static int systemPrint(lua_State *L)
{
	printf ("systemPrinted");
	return 1;
}

static int setItem(lua_State *L)
{
	printf ("setItem");
	return 1;
}

static int getItem(lua_State *L)
{
	printf ("getItem");
	return 1;
}

const char *vm_loadbuff(const char *code, size_t sz, const char *name, bc_ctx_t *bc_ctx, lua_State **p)
{
	int err;
	const char *errMsg = NULL;

	preloadSystem(L);
	setLuaExecContext(L, bc_ctx);
	err = luaL_loadbuffer(L, code, sz, name);
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

static int getSender(lua_State *L)
{
	bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		return -1;
	}
	lua_pushstring(L, exec->sender);
	return 1;
}

static int getBlockhash(lua_State *L)
{
	bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		return -1;
	}

	lua_pushstring(L, exec->blockHash);
	return 1;
}

static int getTxhash(lua_State *L)
{
	bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		return -1;
	}

	lua_pushstring(L, exec->txHash);
	return 1;
}

static int getBlockHeight(lua_State *L)
{
	bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		return -1;
	}

	lua_pushinteger(L, exec->blockHeight);
	return 1;
}

static int getTimestamp(lua_State *L)
{
	bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		return -1;
	}
	lua_pushinteger(L, exec->timestamp);
	return 1;
}

static int getContractID(lua_State *L)
{
	bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		return -1;
	}
	
	lua_pushstring(L, exec->contractId);
	return 1;
}
