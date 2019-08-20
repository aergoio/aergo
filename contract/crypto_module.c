#include "_cgo_export.h"
#include "util.h"

extern const int *getLuaExecContext(lua_State *L);

static int crypto_sha256(lua_State *L)
{
    size_t len;
    char *arg;
    struct luaCryptoSha256_return ret;

    luaL_checktype(L, 1, LUA_TSTRING);
    arg = (char *)lua_tolstring(L, 1, &len);

    ret = luaCryptoSha256(L, arg, len);
    if (ret.r1 < 0) {
        strPushAndRelease(L, ret.r1);
        lua_error(L);
    }
    strPushAndRelease(L, ret.r0);
	return 1;
}

static int crypto_ecverify(lua_State *L)
{
    char *msg, *sig, *addr;
    struct luaECVerify_return ret;
	int *service = (int *)getLuaExecContext(L);

    luaL_checktype(L, 1, LUA_TSTRING);
    luaL_checktype(L, 2, LUA_TSTRING);
    luaL_checktype(L, 3, LUA_TSTRING);
    msg = (char *)lua_tostring(L, 1);
    sig = (char *)lua_tostring(L, 2);
    addr = (char *)lua_tostring(L, 3);

    ret = luaECVerify(L, *service, msg, sig, addr);
    if (ret.r1 != NULL) {
        strPushAndRelease(L, ret.r1);
        lua_error(L);
    }

    lua_pushboolean(L, ret.r0);

	return 1;
}

static int crypto_verifyProof(lua_State *L)
{
    int argc = lua_gettop(L);
    char *k, *v, *h;
    struct proof *proof;
    size_t kLen, vLen, hLen, nProof;
    int i, b;
    const int proofIndex = 4;
    if (argc < proofIndex) {
        lua_pushboolean(L, 0);
        return 1;
    }
    nProof = argc - (proofIndex - 1);
    k = (char *)lua_tolstring(L, 1, &kLen);
    v = (char *)lua_tolstring(L, 2, &vLen);
    h = (char *)lua_tolstring(L, 3, &hLen);
    proof = (struct proof *)malloc(sizeof(struct proof) * nProof);
    for (i = proofIndex; i <= argc; ++i) {
        proof[i-proofIndex].data = (char *)lua_tolstring(L, i, &proof[i-proofIndex].len);
    }
    b = luaCryptoVerifyProof(k, kLen, v, vLen, h, hLen, proof, nProof);
    free(proof);
    lua_pushboolean(L, b);
    return 1;
}

static int crypto_keccak256(lua_State *L)
{
    size_t len;
    char *arg;
    struct luaCryptoKeccak256_return ret;

    luaL_checktype(L, 1, LUA_TSTRING);
    arg = (char *)lua_tolstring(L, 1, &len);

    ret = luaCryptoKeccak256(arg, len);
    lua_pushlstring(L, ret.r0, ret.r1);
    free(ret.r0);
	return 1;
}

static const luaL_Reg crypto_lib[] = {
	{"sha256", crypto_sha256},
	{"ecverify", crypto_ecverify},
	{"verifyProof", crypto_verifyProof},
	{"keccak256", crypto_keccak256},
	{NULL, NULL}
};

int luaopen_crypto(lua_State *L)
{
	luaL_register(L, "crypto", crypto_lib);
	lua_pop(L, 1);
	return 1;
}
