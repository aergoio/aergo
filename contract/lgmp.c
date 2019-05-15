#include <stdlib.h>
#include <string.h>

#include "lua.h"
#include "lauxlib.h"
#include "lgmp.h"
#include "math.h"

#define lua_boxpointer(L,u) \
	(*(void **)(lua_newuserdata(L, sizeof(void *))) = (u))
#define MPZ(a) ((mpz_ptr)(a->mpptr))

static const char *mp_num_memory_error="bignum not enough memory";
static const char *mp_num_invalid_number="bignum invalid number string";
static const char *mp_num_divide_zero="bignum divide by zero";

static void Bnew(lua_State *L, mp_num x)
{
	lua_boxpointer(L,x);
	luaL_getmetatable(L,MYTYPE);
	lua_setmetatable(L,-2);
}

static mp_num bn_alloc (int type) {
	mp_num new = malloc(sizeof(bn_struct));
	if (new == NULL) 
		return NULL;
	new->type = type;
	mpz_ptr pz = malloc(sizeof(mpz_t));
	if (pz == NULL)
		return NULL;
	mpz_init(pz);
	new->mpptr = pz;
	return new;
}

static void mp_num_free(mp_num x) {
	mpz_clear(x->mpptr);
	free(x->mpptr);
	free(x);
}

const char *lua_set_bignum(lua_State *L, char *s)
{
	mp_num x;
	x = bn_alloc(BN_Integer);
	if (x == NULL) {
		return mp_num_memory_error;
	}
	if (mpz_init_set_str(x->mpptr, s, 0) != 0) {
		mp_num_free(x);
		return mp_num_invalid_number;
	}
	Bnew(L, x);
	return NULL;
}

mp_num Bgetbnum(lua_State *L, int i)
{
   return (mp_num)*((void**)luaL_checkudata(L,i,MYTYPE));
}

int lua_isbignumber(lua_State *L, int i)
{
	if (luaL_testudata(L, i, MYTYPE) != NULL)
		return 1;
	return 0;
}

int Bis(lua_State *L)
{
   lua_pushboolean(L, lua_isbignumber(L, 1) != 0);
   return 1;
}

static mp_num Bget(lua_State *L, int i)
{
	switch (lua_type(L, i)) 
	{
		case LUA_TNUMBER:
		{
			mp_num x;
			double d = lua_tonumber(L, i);
			x = bn_alloc(BN_Integer);
			if (x == NULL)
				luaL_error(L, mp_num_memory_error);
			if (isnan(d) || isinf(d)) {
			    luaL_error(L, "can't convert nan or infinity");
			}
			mpz_init_set_d(x->mpptr, d);
			Bnew(L, x);
			lua_replace(L, i);
			return x;

		}
		case LUA_TSTRING:
		{
			mp_num x;
			const char *s = lua_tostring(L, i);
			x = bn_alloc(BN_Integer);
			if (x == NULL)
				luaL_error(L, mp_num_memory_error);
			if (mpz_init_set_str(x->mpptr, s, 0) != 0) {
				mp_num_free(x);
				luaL_error(L, mp_num_invalid_number);
			}
			Bnew(L, x);
			lua_replace(L, i);
			return x;
		}
		default:
		return *((void**)luaL_checkudata(L,i,MYTYPE));
	}
	return NULL;
}

static void bn_copy (mp_num src) 
{
	mp_num new = bn_alloc(src->type);
	mpz_set(new->mpptr, src->mpptr);
}

static int Bdo1(lua_State *L, void (*f)(mpz_ptr a, mpz_srcptr b, mpz_srcptr c), char is_div)
{
	mp_num a = Bget(L, 1);
	mp_num b = Bget(L, 2);
	mp_num c;

	if (is_div == 1 && mpz_sgn(MPZ(b)) == 0)
		luaL_error(L, mp_num_divide_zero);

	c = bn_alloc(a->type);
	if (c == NULL)
		luaL_error(L, mp_num_memory_error);
	f(MPZ(c), MPZ(a), MPZ(b));
	Bnew(L, c);
	return 1;
}

char *lua_get_bignum_str(lua_State *L, int idx)
{
	char *res;
	mp_num a = Bget(L, idx);
	char *str = malloc(mpz_sizeinbase (a->mpptr, MPZ_BASE) + 2);
	if (str == NULL) 
		return NULL;
	
	res = mpz_get_str(str, MPZ_BASE, a->mpptr);
	return res;
}

long int lua_get_bignum_si(lua_State *L, int idx)
{
	mp_num a = Bget(L, idx);
	if (mpz_fits_slong_p(MPZ(a)) == 0)
		return 0;
	return mpz_get_si(MPZ(a));
}

int lua_bignum_is_zero(lua_State *L, int idx)
{
	mp_num a = Bget(L, idx);
	return mpz_sgn(MPZ(a));
}

static int Btostring(lua_State *L)
{
	char *res = lua_get_bignum_str(L, 1);
	if (res == NULL)
		luaL_error(L, mp_num_memory_error);
	lua_pushstring(L, res);
	free(res);
	return 1;
}

static int Btonumber(lua_State *L)
{
	mp_num a = Bget(L, 1);
	lua_pushnumber(L, mpz_get_d(a->mpptr));
	return 1;
}

static int Biszero(lua_State *L)
{
	mp_num a = Bget(L, 1);
	lua_pushboolean(L, mpz_sgn(MPZ(a)) == 0);
	
	return 1;
}

static int Bisneg(lua_State *L)
{
	mp_num a = Bget(L, 1);
	lua_pushboolean(L, (mpz_sgn(MPZ(a)) < 0));
	
	return 1;
}

static int Bnumber(lua_State *L) 
{
	Bget(L, 1);
	lua_settop(L, 1);
	return 1;
}

static int Bcompare(lua_State *L)
{
	mp_num a = Bget(L, 1);
	mp_num b = Bget(L, 2);
	lua_pushinteger(L, mpz_cmp(a->mpptr, b->mpptr));
	return 1;
}

static int Beq(lua_State *L)
{
	mp_num a = Bget(L, 1);
	mp_num b = Bget(L, 2);
	lua_pushboolean(L, mpz_cmp(a->mpptr, b->mpptr) == 0);
	return 1;
}

static int Blt(lua_State *L)
{
	mp_num a = Bget(L, 1);
	mp_num b = Bget(L, 2);
	lua_pushboolean(L, mpz_cmp(a->mpptr, b->mpptr) < 0);
	return 1;
}

static int Badd(lua_State *L)			/** add(x,y) */
{
	return Bdo1(L, mpz_add, 0);
}

static int Bsub(lua_State *L)			/** sub(x,y) */
{
	return Bdo1(L, mpz_sub, 0);
}

static int Bmul(lua_State *L)			/** mul(x,y) */
{
	return Bdo1(L, mpz_mul, 0);
}

static int Bpow(lua_State *L)			/** pow(x,y) */
{
	mp_num a = Bget(L, 1);
	mp_num b = Bget(L, 2);
	mp_num c;

	c = bn_alloc(a->type);
	if (c == NULL)
		luaL_error(L, mp_num_memory_error);
	
	mpz_pow_ui(c->mpptr, a->mpptr, mpz_get_ui(b->mpptr));
	Bnew(L, c);
	return 1;
}

static int Bdiv(lua_State *L)			/** div(x,y) */
{
	return Bdo1(L, mpz_tdiv_q, 1);
}

static int Bmod(lua_State *L)			/** mod(x,y) */
{
	return Bdo1(L, mpz_tdiv_r, 1);
}

static int Bdivmod(lua_State *L)		/** divmod(x,y) */
{
	mp_num a=Bget(L,1);
	mp_num b=Bget(L,2);
	mp_num q;
	mp_num r;

	if (mpz_sgn(MPZ(b)) == 0)
		luaL_error(L, mp_num_divide_zero);

	q = bn_alloc(a->type);
	r = bn_alloc(a->type);
	if (q == NULL || r == NULL)
		luaL_error(L, mp_num_memory_error);
	
	mpz_tdiv_qr(q->mpptr, r->mpptr, a->mpptr, b->mpptr);
	Bnew(L, q);
	Bnew(L, r);
	return 2;
}

static int Bgc(lua_State *L)
{
	mp_num x=Bget(L,1);
	mp_num_free(x);
	lua_pushnil(L);
	lua_setmetatable(L,1);
	return 0;
}

static int Bneg(lua_State *L)			/** neg(x) */
{
	mp_num a=Bget(L,1);
	mp_num res;

	res = bn_alloc(a->type);
	if (res == NULL)
		luaL_error(L, mp_num_memory_error);
	
	mpz_neg (res->mpptr, a->mpptr);
	Bnew(L, res);
	return 1;
}

static int Bpowmod(lua_State *L)		/** powmod(x,y,m) */
{
	mp_num a=Bget(L,1);
	mp_num k=Bget(L,2);
	mp_num m=Bget(L,3);
	mp_num r;

	if (mpz_sgn(MPZ(m)) == 0)
		luaL_error(L, mp_num_divide_zero);

	r = bn_alloc(a->type);
	if (r == NULL)
		luaL_error(L, mp_num_memory_error);
	
	mpz_powm(r->mpptr, a->mpptr, k->mpptr, m->mpptr);
	Bnew(L, r);
	return 1;
}

static int Bsqrt(lua_State *L)			/** sqrt(x) */
{
	mp_num a=Bget(L,1);
	mp_num res;

	res = bn_alloc(a->type);
	if (res == NULL)
		luaL_error(L, mp_num_memory_error);
	
	mpz_sqrt (res->mpptr, a->mpptr);
	Bnew(L, res);
	return 1;
}

static const luaL_Reg R[] =
{
	{ "__add",	Badd },		/** __add(x,y) */
	{ "__div",	Bdiv	},		/** __div(x,y) */
	{ "__eq",	Beq	},		/** __eq(x,y) */
	{ "__gc",	Bgc	},
	{ "__lt",	Blt	},		/** __lt(x,y) */
	{ "__mod",	Bmod	},		/** __mod(x,y) */
	{ "__mul",	Bmul	},		/** __mul(x,y) */
	{ "__pow",	Bpow	},		/** __pow(x,y) */
	{ "__sub",	Bsub	},		/** __sub(x,y) */
	{ "__tostring",	Btostring},		/** __tostring(x) */
	{ "__unm",	Bneg	},		/** __unm(x) */
	{ "add",	Badd    },
	{ "compare",	Bcompare},
	{ "div",	Bdiv	},
	{ "divmod",	Bdivmod	},
	{ "isneg",	Bisneg	},
	{ "iszero",	Biszero	},
	{ "mod",	Bmod	},
	{ "mul",	Bmul	},
	{ "neg",	Bneg	},
	{ "number",	Bnumber	},
	{ "pow",	Bpow	},
	{ "powmod",	Bpowmod	},
	{ "sqrt",	Bsqrt	},
	{ "sub",	Bsub	},
	{ "tonumber",	Btonumber},
	{ "tostring",	Btostring},
	{ "isbignum",	Bis},
	{ NULL,		NULL	}
};

LUALIB_API int luaopen_gmp(lua_State *L)
{
	luaL_newmetatable(L,MYTYPE);
	lua_setglobal(L,MYNAME);
	luaL_register(L,MYNAME,R);
	lua_pushliteral(L,"version");			/** version */
	lua_pushliteral(L,MYVERSION);
	lua_settable(L,-3);
	lua_pushliteral(L,"__index");
	lua_pushvalue(L,-2);
	lua_settable(L,-3);
	return 1;
}
