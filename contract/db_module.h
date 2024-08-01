#ifndef _DB_MODULE_H
#define _DB_MODULE_H

#include "lua.h"
#include "sqlite3-binding.h"

extern int luaopen_db(lua_State *L);
extern int lua_db_release_resource(lua_State *L);

sqlite3 *vm_get_db(lua_State *L);

#endif /* _DB_MODULE_H */
