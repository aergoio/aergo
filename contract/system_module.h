#ifndef _SYSTEM_MODULE_H
#define _SYSTEM_MODULE_H

#include "lua.h"

extern int luaopen_system(lua_State *L);

extern int setItem(lua_State *L);
extern int setItemWithPrefix(lua_State *L);
extern int getItem(lua_State *L);
extern int getItemWithPrefix(lua_State *L);
extern int delItemWithPrefix(lua_State *L);

#endif /* _SYSTEM_MODULE_H */
