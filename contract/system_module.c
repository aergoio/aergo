#include <string.h>
#include <stdlib.h>
#include <time.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern const bc_ctx_t *getLuaExecContext(lua_State *L);

static int systemPrint(lua_State *L)
{
    char *jsonValue;
	bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);

    jsonValue = lua_util_get_json_from_stack (L, 1, lua_gettop(L), true);
    if (jsonValue == NULL) {
		lua_error(L);
	}
    LuaPrint(exec->contractId, jsonValue);
    free(jsonValue);
	return 0;
}

int setItem(lua_State *L)
{
	const char *key;
	char *jsonValue;
	char *dbKey;
	bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}

	if (exec->isQuery) {
	    luaL_error(L, "set not permitted in query");
	}

	luaL_checkany(L, 2);
	key = luaL_checkstring(L, 1);

	jsonValue = lua_util_get_json (L, -1, false);
	if (jsonValue == NULL) {
		lua_error(L);
	}

	dbKey = lua_util_get_db_key(exec, key);
	if (LuaSetDB(L, exec->stateKey, dbKey, jsonValue) != 0) {
		free(jsonValue);
		free(dbKey);
		lua_error(L);
	}
	free(jsonValue);
	free(dbKey);

	return 0;
}

int getItem(lua_State *L)
{
	const char *key;
	char *dbKey;
	bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);
	char *jsonValue;
	int ret;

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	key = luaL_checkstring(L, 1);
	dbKey = lua_util_get_db_key(exec, key);

	ret = LuaGetDB(L, exec->stateKey, dbKey);

	free(dbKey);
	if (ret < 0) {
		lua_error(L);
	}
	if (ret == 0)
		return 0;
	jsonValue = (char *)luaL_checkstring(L, -1);
	lua_pop(L, 1);

	if (lua_util_json_to_lua(L, jsonValue, false) != 0) {
		luaL_error(L, "getItem error : can't convert %s", jsonValue);
	}
	return 1;
}

static int getSender(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	LuaAddressEncode(L, exec->sender);
	return 1;
}

static int getTxhash(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushstring(L, exec->txHash);
	return 1;
}

static int getBlockHeight(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushinteger(L, exec->blockHeight);
	return 1;
}

static int getTimestamp(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushinteger(L, exec->timestamp);
	return 1;
}

static int getContractID(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	LuaAddressEncode(L, exec->contractId);
	return 1;
}

static int getCreator(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	int ret;

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	ret = LuaGetDB(L, exec->stateKey, "Creator");
	if (ret < 0) {
		lua_error(L);
	}
	if (ret == 0)
		return 0;
	luaL_checkstring(L, -1);
	return 1;
}

static int getAmount(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushinteger(L, exec->amount);
	return 1;
}

static int getOrigin(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	LuaAddressEncode(L, exec->origin);
	return 1;
}

/* datetime-related functions from lib_os.c. time(NULL) is replaced by blocktime(L) */

static void setfield(lua_State *L, const char *key, int value)
{
    lua_pushinteger(L, value);
    lua_setfield(L, -2, key);
}

static void setboolfield(lua_State *L, const char *key, int value)
{
    if (value < 0)  /* undefined? */
        return;  /* does not set field */
    lua_pushboolean(L, value);
    lua_setfield(L, -2, key);
}

static int getboolfield(lua_State *L, const char *key)
{
    int res;
    lua_getfield(L, -1, key);
    res = lua_isnil(L, -1) ? -1 : lua_toboolean(L, -1);
    lua_pop(L, 1);
    return res;
}

static int getfield(lua_State *L, const char *key, int d)
{
    int res;
    lua_getfield(L, -1, key);
    if (lua_isnumber(L, -1)) {
        res = (int)lua_tointeger(L, -1);
    } else {
        if (d < 0)
            luaL_error(L, "field " LUA_QS " missing in date table", key);
        res = d;
    }
    lua_pop(L, 1);
    return res;
}

static time_t blocktime(lua_State *L)
{
    time_t t;
    getTimestamp(L);
    t = (time_t)lua_tointeger(L, -1);
    lua_pop(L, 1);
    return t;
}

static int os_date(lua_State *L)
{
    const char *s = luaL_optstring(L, 1, "%c");
    time_t t = luaL_opt(L, (time_t)luaL_checknumber, 2, blocktime(L));
    struct tm *stm;
#if LJ_TARGET_POSIX
    struct tm rtm;
#endif
    if (*s == '!') {  /* UTC? */
        s++;  /* Skip '!' */
    }
#if LJ_TARGET_POSIX
    stm = gmtime_r(&t, &rtm);
#else
    stm = gmtime(&t);
#endif
    if (stm == NULL) {  /* Invalid date? */
        lua_pushnil(L);
    } else if (strcmp(s, "*t") == 0) {
        lua_createtable(L, 0, 9);  /* 9 = number of fields */
        setfield(L, "sec", stm->tm_sec);
        setfield(L, "min", stm->tm_min);
        setfield(L, "hour", stm->tm_hour);
        setfield(L, "day", stm->tm_mday);
        setfield(L, "month", stm->tm_mon+1);
        setfield(L, "year", stm->tm_year+1900);
        setfield(L, "wday", stm->tm_wday+1);
        setfield(L, "yday", stm->tm_yday+1);
        setboolfield(L, "isdst", stm->tm_isdst);
    } else {
        char cc[3];
        luaL_Buffer b;
        cc[0] = '%'; cc[2] = '\0';
        luaL_buffinit(L, &b);
        for (; *s; s++) {
            if (*s != '%' || *(s + 1) == '\0') {  /* No conversion specifier? */
                luaL_addchar(&b, *s);
            } else {
                size_t reslen;
                char buff[200];  /* Should be big enough for any conversion result. */
                cc[1] = *(++s);
                reslen = strftime(buff, sizeof(buff), cc, stm);
                luaL_addlstring(&b, buff, reslen);
            }
        }
        luaL_pushresult(&b);
    }
    return 1;
}

static int os_time(lua_State *L)
{
    time_t t;
    if (lua_isnoneornil(L, 1)) {
        t = blocktime(L);
    } else {
        struct tm ts;
        luaL_checktype(L, 1, LUA_TTABLE);
        lua_settop(L, 1);  /* make sure table is at the top */
        ts.tm_sec = getfield(L, "sec", 0);
        ts.tm_min = getfield(L, "min", 0);
        ts.tm_hour = getfield(L, "hour", 12);
        ts.tm_mday = getfield(L, "day", -1);
        ts.tm_mon = getfield(L, "month", -1) - 1;
        ts.tm_year = getfield(L, "year", -1) - 1900;
        ts.tm_isdst = getboolfield(L, "isdst");
        t = timegm(&ts);
    }
    if (t == (time_t)(-1))
        lua_pushnil(L);
    else
        lua_pushnumber(L, (lua_Number)t);
    return 1;
}

static int os_difftime(lua_State *L)
{
    lua_pushnumber(L, difftime((time_t)(luaL_checknumber(L, 1)),
                (time_t)(luaL_optnumber(L, 2, (lua_Number)0))));
    return 1;
}

/* end of datetime functions */

static const luaL_Reg sys_lib[] = {
	{"print", systemPrint},
	{"setItem", setItem},
	{"getItem", getItem},
	{"getSender", getSender},
	{"getCreator", getCreator},
	{"getTxhash", getTxhash},
	{"getBlockheight", getBlockHeight},
	{"getTimestamp", getTimestamp},
	{"getContractID", getContractID},
	{"getOrigin", getOrigin},
	{"getAmount", getAmount},
	{"date", os_date},
	{"time", os_time},
	{"difftime", os_difftime},
	{NULL, NULL}
};

int luaopen_system(lua_State *L)
{
	luaL_register(L, "system", sys_lib);
	lua_pop(L, 1);
	return 1;
}
