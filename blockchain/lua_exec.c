#include <string.h>
#include <stdlib.h>
#include "lua_exec.h"

#define MAX_SYSTEM_FNS 5

const char *luaExecContext= "__exec_context__";

static int systemPrint(lua_State *L);
static int setItem(lua_State *L);
static int getItem(lua_State *L);
static int getSender(lua_State *L);

static const struct luaL_Reg sys_lib[MAX_SYSTEM_FNS] = {
	{"print", systemPrint},
	{"setItem", setItem},
	{"getItem", getItem},
	{"getSender", getSender},
	{NULL, NULL}
};

lua_State *new_lstate(uint64_t max_inst_count)
{
	lua_State *L = luaL_newstate();
	return L;
}

void preloadSystem(lua_State *L)
{
	luaL_register(L, "system", sys_lib);
	luaL_openlibs(L);
}

void setLuaExecContext(lua_State *L, struct exec_context *exec)
{
	lua_pushlightuserdata(L, exec);
	lua_setglobal(L, luaExecContext);
}

static void* getLuaExecContext(lua_State *L)
{
	void *exec;
	lua_getglobal(L, luaExecContext);
	exec = (void *)lua_touserdata(L, 1);
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

static int getSender(lua_State *L)
{
	void *exec = getLuaExecContext(L);
	
	lua_pushstring(L, ((struct exec_context *)exec)->sender);
	return 1;
}

const char* vm_run(lua_State *L, const char *code, size_t sz, const char *name)
{
	int err;
	const char *errMsg = NULL;

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