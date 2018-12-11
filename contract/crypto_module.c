#include "_cgo_export.h"

static int crypto_sha256(lua_State *L)
{
    size_t len;
    char *arg;

    luaL_checktype(L, 1, LUA_TSTRING);
    arg = (char *)lua_tolstring(L, 1, &len);

    LuaCryptoSha256(L, arg, len);
	return 1;
}

static int crypto_ecverify(lua_State *L)
{
    char *msg, *sig, *addr;

    luaL_checktype(L, 1, LUA_TSTRING);
    luaL_checktype(L, 2, LUA_TSTRING);
    luaL_checktype(L, 3, LUA_TSTRING);
    msg = (char *)lua_tostring(L, 1);
    sig = (char *)lua_tostring(L, 2);
    addr = (char *)lua_tostring(L, 3);

    LuaECVerify(L, msg, sig, addr);
	return 1;
}

static const luaL_Reg crypto_lib[] = {
	{"sha256", crypto_sha256},
	{"ecverify", crypto_ecverify},
	{NULL, NULL}
};

int luaopen_crypto(lua_State *L)
{
	luaL_register(L, "crypto", crypto_lib);
	lua_pop(L, 1);
	return 1;
}
