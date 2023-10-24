#ifndef _LGMP_H_
#define _LGMP_H_

#include "gmp.h"

typedef struct bn_struct *mp_num;

typedef struct bn_struct {
	int type;
	void *mpptr;
} bn_struct;

#define MYNAME		"bignum"
#define MYVERSION	MYNAME " library for " LUA_VERSION " / Apr 2010 / "\
			"based on GNU bc-1.06"
#define MYTYPE		MYNAME " bignumber"
#define MPZ_BASE 10

enum bn_type {
	BN_Integer = 0,
	BN_Float
};

int luaopen_bignum(lua_State *L);
const char *lua_set_bignum(lua_State *L, char *s);
mp_num Bgetbnum(lua_State *L, int i);
int lua_isbignumber(lua_State *L, int i);
char *lua_get_bignum_str(lua_State *L, int idx);
long int lua_get_bignum_si(lua_State *L, int idx);
int lua_bignum_is_zero(lua_State *L, int idx);
const char *init_bignum();

#endif /*_LGMP_H_*/
