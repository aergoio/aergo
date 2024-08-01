#include "_cgo_export.h"
#include "util.h"

extern void checkLuaExecContext(lua_State *L);

static int crypto_sha256(lua_State *L) {
	size_t len;
	char *arg;
	struct luaCryptoSha256_return ret;

	lua_gasuse(L, 500);

	luaL_checktype(L, 1, LUA_TSTRING);
	arg = (char *) lua_tolstring(L, 1, &len);

	ret = luaCryptoSha256(L, arg, len);
	if (ret.r1 < 0) {
		strPushAndRelease(L, ret.r1);
		lua_error(L);
	}
	strPushAndRelease(L, ret.r0);
	return 1;
}

static int crypto_ecverify(lua_State *L) {
	char *msg, *sig, *addr;
	struct luaECVerify_return ret;

	checkLuaExecContext(L);

	lua_gasuse(L, 5000);

	luaL_checktype(L, 1, LUA_TSTRING);
	luaL_checktype(L, 2, LUA_TSTRING);
	luaL_checktype(L, 3, LUA_TSTRING);
	msg  = (char *) lua_tostring(L, 1);
	sig  = (char *) lua_tostring(L, 2);
	addr = (char *) lua_tostring(L, 3);

	ret = luaECVerify(L, msg, sig, addr);
	if (ret.r1 != NULL) {
		strPushAndRelease(L, ret.r1);
		lua_error(L);
	}
	lua_pushboolean(L, ret.r0);
	return 1;
}

static void set_rlp_obj(struct rlp_obj *n, int type, void *data, size_t size) {
	n->rlp_obj_type = type;
	n->data = data;
	n->size = size;
}

static void set_str_rlp_obj(lua_State *L, int n, struct rlp_obj *o) {
	char *data;
	size_t size;
	data = (char *) lua_tolstring(L, n, &size);
	set_rlp_obj(o, RLP_TSTRING, (void *)data, size);
}

static struct rlp_obj *makeValue(lua_State *L, int n) {
	struct rlp_obj *o = (struct rlp_obj *) malloc(sizeof(struct rlp_obj));

	set_rlp_obj(o, RLP_TSTRING, NULL, 0);

	if (lua_isstring(L, n)) {
		set_str_rlp_obj(L, n, o);
	} else if (lua_istable(L, n)) {
		struct rlp_obj *list;
		int list_len, i;

		list_len = (int) lua_objlen(L, n);
		if (list_len > 20) {
			free(o);
			luaL_argerror(L, 2, "too many elements in the value");
		}

		list = (struct rlp_obj *) malloc(sizeof(struct rlp_obj) * list_len);
		set_rlp_obj(o, RLP_TLIST, list, list_len);

		for (i = 0; i < list_len; i++) {
			struct rlp_obj *elem = &list[i];
			lua_rawgeti(L, n, i+1);
			if (lua_isstring(L, -1)) {
				set_str_rlp_obj(L, -1, elem);
			} else {
				set_rlp_obj(elem, RLP_TSTRING, NULL, 0);
			}
			lua_pop(L, 1);
		}
	}

	return o;
}

static int crypto_verifyProof(lua_State *L) {
	struct luaCryptoVerifyProof_return ret;
	int argc = lua_gettop(L);
	char *k, *h;
	struct rlp_obj *v;
	struct proof *proof;
	size_t kLen, hLen, nProof;
	int i;
	const int proofIndex = 4;

	lua_gasuse(L, 5000);

	if (argc < proofIndex) {
		lua_pushboolean(L, 0);
		return 1;
	}

	nProof = argc - (proofIndex - 1);
	k = (char *) lua_tolstring(L, 1, &kLen);
	v = makeValue(L, 2);
	h = (char *) lua_tolstring(L, 3, &hLen);

	proof = (struct proof *) malloc(sizeof(struct proof) * nProof);
	for (i = proofIndex; i <= argc; ++i) {
		proof[i-proofIndex].data = (char *) lua_tolstring(L, i, &proof[i-proofIndex].len);
	}

	ret = luaCryptoVerifyProof(L, k, kLen, v, h, hLen, proof, nProof);
	if (ret.r1 != NULL) {
		strPushAndRelease(L, ret.r1);
		lua_error(L);
	}

	if (proof != NULL) {
		free(proof);
	}
	if (v != NULL) {
		if (v->rlp_obj_type == RLP_TLIST) {
			free(v->data);
		}
		free(v);
	}

	lua_pushboolean(L, ret.r0);
	return 1;
}

static int crypto_keccak256(lua_State *L) {
	size_t len;
	char *arg;
	struct luaCryptoKeccak256_return ret;

	lua_gasuse(L, 500);

	luaL_checktype(L, 1, LUA_TSTRING);
	arg = (char *) lua_tolstring(L, 1, &len);

	ret = luaCryptoKeccak256(L, arg, len);
	if (ret.r2 != NULL) {
		strPushAndRelease(L, ret.r2);
		lua_error(L);
	}
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

int luaopen_crypto(lua_State *L) {
	luaL_register(L, "crypto", crypto_lib);
	lua_pop(L, 1);
	return 1;
}
