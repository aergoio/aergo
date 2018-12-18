#ifndef _LBC_H_
#define _LBC_H

#include "lua.h"
#include "number.h"

int luaopen_bc(lua_State *L);
void Bset(lua_State *L, char *s);
bc_num Bgetbnum(lua_State *L, int i);
int lua_isbignumber(lua_State *L, int i);

#endif /* _LBC_H */