#include <string.h>
#include <stdlib.h>
#include <time.h>
#include <stdint.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

#define STATE_DB_KEY_PREFIX "_"

extern int getLuaExecContext(lua_State *L);

static int systemPrint(lua_State *L)
{
    char *jsonValue;
	int service = getLuaExecContext(L);

    lua_gasuse(L, 100);
    jsonValue = lua_util_get_json_from_stack (L, 1, lua_gettop(L), true);
    if (jsonValue == NULL) {
		luaL_throwerror(L);
	}
    luaPrint(L, service, jsonValue);
    free(jsonValue);
	return 0;
}

static char *getDbKey(lua_State *L, int *len)
{
    size_t size;
    char *key;

	lua_pushvalue(L, 1);    /* prefix key */
	lua_concat(L, 2);       /* dbKey(prefix..key) */
	key = (char *)lua_tolstring(L, -1, &size);
	*len = size;
	return key;
}

int setItemWithPrefix(lua_State *L)
{
	char *dbKey;
	char *jsonValue;
	int service = getLuaExecContext(L);
	char *errStr;
	int keylen;

	lua_gasuse(L, 100);

	luaL_checkstring(L, 1);
	luaL_checkany(L, 2);
	luaL_checkstring(L, 3);
	dbKey = getDbKey(L, &keylen);

	jsonValue = lua_util_get_json (L, 2, false);
	if (jsonValue == NULL) {
		luaL_throwerror(L);
	}

	lua_gasuse_mul(L, GAS_SDATA, strlen(jsonValue));
	if ((errStr = luaSetDB(L, service, dbKey, keylen, jsonValue)) != NULL) {
		free(jsonValue);
		strPushAndRelease(L, errStr);
		luaL_throwerror(L);
	}
	free(jsonValue);
	return 0;
}

int setItem(lua_State *L)
{
	luaL_checkstring(L, 1);
	luaL_checkany(L, 2);
	lua_pushstring(L, STATE_DB_KEY_PREFIX);
	return setItemWithPrefix(L);
}

int getItemWithPrefix(lua_State *L)
{
	char *dbKey;
	int service = getLuaExecContext(L);
	char *jsonValue;
	char *blkno = NULL;
	struct luaGetDB_return ret;
	int keylen;

	lua_gasuse(L, 100);

	luaL_checkstring(L, 1);
	if(lua_gettop(L) == 2) {
	    luaL_checkstring(L, 2);
	}
	else if (lua_gettop(L) == 3) {
	    if (!lua_isnil(L, 2)) {
	        int type = lua_type(L,2);
	        if (type != LUA_TNUMBER && type != LUA_TSTRING)
	            luaL_error(L, "snap height permitted number or string type");
	        blkno = (char *)lua_tostring(L, 2);
	    }
	    luaL_checkstring(L, 3);
	}
	dbKey = getDbKey(L, &keylen);

	ret = luaGetDB(L, service, dbKey, keylen, blkno);
	if (ret.r1 != NULL) {
        strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	if (ret.r0 == NULL)
		return 0;

    minus_inst_count(L, strlen(ret.r0));
	if (lua_util_json_to_lua(L, ret.r0, false) != 0) {
	    strPushAndRelease(L, ret.r0);
		luaL_error(L, "getItem error : can't convert %s", lua_tostring(L, -1));
	}
	return 1;
}

int getItem(lua_State *L)
{
	luaL_checkstring(L, 1);
	lua_pushstring(L, STATE_DB_KEY_PREFIX);
	if (lua_gettop(L) == 3) {
	    if (!lua_isnil(L, 2))
	        luaL_checknumber(L, 2);
	}
	return getItemWithPrefix(L);
}

int delItemWithPrefix(lua_State *L)
{
	char *dbKey;
	int service = getLuaExecContext(L);
	char *jsonValue;
	char *ret;
	int keylen;

	lua_gasuse(L, 100);

	luaL_checkstring(L, 1);
	luaL_checkstring(L, 2);
	dbKey = getDbKey(L, &keylen);
	ret = luaDelDB(L, service, dbKey, keylen);
	if (ret != NULL) {
	    strPushAndRelease(L, ret);
		luaL_throwerror(L);
	}
    return 0;
}

static int getSender(lua_State *L)
{
	int service = getLuaExecContext(L);
	char *sender;

	lua_gasuse(L, 1000);

	sender = luaGetSender(L, service);
	strPushAndRelease(L, sender);
	return 1;
}

static int getTxhash(lua_State *L)
{
	int service = getLuaExecContext(L);
	char *hash;

	lua_gasuse(L, 500);

	hash = luaGetHash(L, service);
	strPushAndRelease(L, hash);
	return 1;
}

static int getBlockHeight(lua_State *L)
{
	int service = getLuaExecContext(L);

	lua_gasuse(L, 300);

	lua_pushinteger(L, luaGetBlockNo(L, service));
	return 1;
}

static int getTimestamp(lua_State *L)
{
	int service = getLuaExecContext(L);

	lua_gasuse(L, 300);

	lua_pushinteger(L, luaGetTimeStamp(L, service));
	return 1;
}

static int getContractID(lua_State *L)
{
	int service = getLuaExecContext(L);
	char *id;

	lua_gasuse(L, 1000);

	id = luaGetContractId(L, service);
	strPushAndRelease(L, id);
	return 1;
}

static int getCreator(lua_State *L)
{
	int service = getLuaExecContext(L);
	struct luaGetDB_return ret;
	int keylen = 7;

	lua_gasuse(L, 500);

	ret = luaGetDB(L, service, "Creator", keylen, 0);
	if (ret.r1 != NULL) {
	    strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	if (ret.r0 == NULL)
		return 0;
	strPushAndRelease(L, ret.r0);
	return 1;
}

static int getAmount(lua_State *L)
{
	int service = getLuaExecContext(L);
	char *amount;

	lua_gasuse(L, 300);

	amount = luaGetAmount(L, service);
	strPushAndRelease(L, amount);
	return 1;
}

static int getOrigin(lua_State *L)
{
	int service = getLuaExecContext(L);
	char *origin;

	lua_gasuse(L, 1000);

	origin = luaGetOrigin(L, service);
	strPushAndRelease(L, origin);
	return 1;
}

static int getPrevBlockHash(lua_State *L)
{
	int service = getLuaExecContext(L);
	char *hash;

	lua_gasuse(L, 500);

	hash = luaGetPrevBlockHash(L, service);
	strPushAndRelease(L, hash);
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
    lua_gasuse(L, 100);
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
                if (cc[1] == 'c') {
                    reslen = strftime(buff, sizeof(buff), "%Y-%m-%d %H:%M:%S", stm);
                } else {
                    reslen = strftime(buff, sizeof(buff), cc, stm);
                }
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
    lua_gasuse(L, 100);
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
#if LJ_TARGET_POSIX
        t = timegm(&ts);
#else
        t = _mkgmtime(&ts);
#endif
    }
    if (t == (time_t)(-1))
        lua_pushnil(L);
    else
        lua_pushnumber(L, (lua_Number)t);
    return 1;
}

static int os_difftime(lua_State *L)
{
    lua_gasuse(L, 100);
    lua_pushnumber(L, difftime((time_t)(luaL_checknumber(L, 1)),
                (time_t)(luaL_optnumber(L, 2, (lua_Number)0))));
    return 1;
}

/* end of datetime functions */

static int lua_random(lua_State *L)
{
	int service = getLuaExecContext(L);
	int min, max;

    lua_gasuse(L, 100);

	switch (lua_gettop(L)) {
	case 1:
        max = luaL_checkint(L, 1);
        if (max < 1) {
            luaL_error(L, "system.random: the maximum value must be greater than zero");
        }
        lua_pushinteger(L, luaRandomInt(1, max, service));
        break;
	case 2:
		min = luaL_checkint(L, 1);
		max = luaL_checkint(L, 2);
		if (min < 1) {
			luaL_error(L, "system.random: the minimum value must be greater than zero");
		}
		if (min > max) {
			luaL_error(L, "system.random: the maximum value must be greater than the minimum value");
		}
        lua_pushinteger(L, luaRandomInt(min, max, service));
        break;
	default:
        luaL_error(L, "system.random: 1 or 2 arguments required");
        break;
	}
    return 1;
}

static int is_contract(lua_State *L)
{
    char *contract;
	int service = getLuaExecContext(L);
	struct luaIsContract_return ret;

    lua_gasuse(L, 100);

	contract = (char *)luaL_checkstring(L, 1);
    ret = luaIsContract(L, service, contract);
	if (ret.r1 != NULL) {
	    strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	if (ret.r0 == 0)
	    lua_pushboolean(L, false);
	else
	    lua_pushboolean(L, true);

    return 1;
}

static int is_fee_delegation(lua_State *L)
{
	int service = getLuaExecContext(L);
	struct luaIsFeeDelegation_return ret;

    ret = luaIsFeeDelegation(L, service);
	if (ret.r1 != NULL) {
	    strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	if (ret.r0 == 0) {
	    lua_pushboolean(L, false);
    } else {
	    lua_pushboolean(L, true);
    }
    return 1;
}

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
	{"getPrevBlockHash", getPrevBlockHash},
	{"date", os_date},
	{"time", os_time},
	{"difftime", os_difftime},
	{"random", lua_random},
	{"isContract", is_contract},
	{"isFeeDelegation", is_fee_delegation},
	{NULL, NULL}
};

int luaopen_system(lua_State *L)
{
	luaL_register(L, "system", sys_lib);
	lua_pop(L, 1);
	return 1;
}
