#include <string.h>
#include <stdlib.h>
#include <ctype.h>
#include <time.h>
#include <sqlite3-binding.h>
#include "vm.h"
#include "sqlcheck.h"
#include "lgmp.h"
#include "util.h"
#include "_cgo_export.h"

#define LAST_ERROR(L,db,rc)                         \
    do {                                            \
        if ((rc) != SQLITE_OK) {                    \
            luaL_error((L), sqlite3_errmsg((db)));  \
        }                                           \
    } while(0)

#define RESOURCE_PSTMT_KEY "_RESOURCE_PSTMT_KEY_"
#define RESOURCE_RS_KEY "_RESOURCE_RS_KEY_"

extern int getLuaExecContext(lua_State *L);
static void get_column_meta(lua_State *L, sqlite3_stmt* stmt);

static int append_resource(lua_State *L, const char *key, void *data)
{
    int refno;
    if (luaL_findtable(L, LUA_REGISTRYINDEX, key, 0) != NULL) {
        luaL_error(L, "cannot find the environment of the db module");
    }
    /* tab */
    lua_pushlightuserdata(L, data);     /* tab pstmt */
    refno = luaL_ref(L, -2);            /* tab */
    lua_pop(L, 1);                      /* remove tab */
    return refno;
}

#define DB_PSTMT_ID "__db_pstmt__"

typedef struct {
    sqlite3 *db;
    sqlite3_stmt *s;
    int closed;
    int refno;
} db_pstmt_t;

#define DB_RS_ID "__db_rs__"

typedef struct {
    sqlite3 *db;
    sqlite3_stmt *s;
    int closed;
    int nc;
    int shared_stmt;
    char **decltypes;
    int refno;
} db_rs_t;

static db_rs_t *get_db_rs(lua_State *L, int pos)
{
    db_rs_t *rs = luaL_checkudata(L, pos, DB_RS_ID);
    if (rs->closed) {
        luaL_error(L, "resultset is closed");
    }
    return rs;
}

static int db_rs_tostr(lua_State *L)
{
    db_rs_t *rs = luaL_checkudata(L, 1, DB_RS_ID);
    if (rs->closed) {
        lua_pushfstring(L, "resultset is closed");
    } else {
        lua_pushfstring(L, "resultset{handle=%p}", rs->s);
    }
    return 1;
}

static char *dup_decltype(const char *decltype)
{
    int n;
    char *p;
    char *c;

    if (decltype == NULL) {
        return NULL;
    }

    p = c = malloc(strlen(decltype)+1);
    while ((*c++ = tolower(*decltype++))) ;

    if (strcmp(p, "date") == 0 || strcmp(p, "datetime") == 0 || strcmp(p, "timestamp") == 0 ||
        strcmp(p, "boolean") == 0) {
        return p;
    }
    free(p);
    return NULL;
}

static void free_decltypes(db_rs_t *rs)
{
    int i;
    for (i = 0; i < rs->nc; i++) {
        if (rs->decltypes[i] != NULL)
            free(rs->decltypes[i]);
    }
    free(rs->decltypes);
    rs->decltypes = NULL;
}

static int db_rs_get(lua_State *L)
{
    db_rs_t *rs = get_db_rs(L, 1);
    int i;
    sqlite3_int64 d;
    double f;
    int n;
    const unsigned char *s;

    if (rs->decltypes == NULL) {
        luaL_error(L, "`get' called without calling `next'");
    }
    for (i = 0; i < rs->nc; i++) {
        switch (sqlite3_column_type(rs->s, i)) {
        case SQLITE_INTEGER:
            d = sqlite3_column_int64(rs->s, i);
            if (rs->decltypes[i] == NULL)  {
                lua_pushinteger(L, d);
            } else if (strcmp(rs->decltypes[i], "boolean") == 0) {
                if (d != 0) {
                    lua_pushboolean(L, 1);
                } else {
                    lua_pushboolean(L, 0);
                }
            } else { // date, datetime, timestamp
                char buf[80];
                strftime(buf, 80, "%Y-%m-%d %H:%M:%S", gmtime((time_t *)&d));
                lua_pushlstring(L, (const char *)buf, strlen(buf));
            }
            break;
        case SQLITE_FLOAT:
            f = sqlite3_column_double(rs->s, i);
            lua_pushnumber(L, f);
            break;
        case SQLITE_TEXT:
            n = sqlite3_column_bytes(rs->s, i);
            s = sqlite3_column_text(rs->s, i);
            lua_pushlstring(L, (const char *)s, n);
            break;
        case SQLITE_NULL:
            /* fallthrough */
        default: /* unsupported types */
            lua_pushnil(L);
        }
    }
    return rs->nc;
}

static int db_rs_colcnt(lua_State *L)
{
    db_rs_t *rs = get_db_rs(L, 1);

    lua_pushinteger(L, rs->nc);
    return 1;
}

static void db_rs_close(lua_State *L, db_rs_t *rs, int remove)
{
    if (rs->closed) {
        return;
    }
    rs->closed = 1;
    if (rs->decltypes) {
        free_decltypes(rs);
    }
    if (rs->shared_stmt == 0) {
        sqlite3_finalize(rs->s);
    }
    if (remove) {
        if (luaL_findtable(L, LUA_REGISTRYINDEX, RESOURCE_RS_KEY, 0) != NULL) {
            luaL_error(L, "cannot find the environment of the db module");
        }
        luaL_unref(L, -1, rs->refno);
        lua_pop(L, 1);
    }
}

static int db_rs_next(lua_State *L)
{
    db_rs_t *rs = get_db_rs(L, 1);
    int rc;

    rc = sqlite3_step(rs->s);
    if (rc == SQLITE_DONE) {
        db_rs_close(L, rs, 1);
        lua_pushboolean(L, 0);
    } else if (rc != SQLITE_ROW) {
        rc = sqlite3_reset(rs->s);
        LAST_ERROR(L, rs->db, rc);
        db_rs_close(L, rs, 1);
        lua_pushboolean(L, 0);
    } else {
        if (rs->decltypes == NULL) {
            int i;
            rs->decltypes = malloc(sizeof(char *) * rs->nc);
            for (i = 0; i < rs->nc; i++) {
                rs->decltypes[i] = dup_decltype(sqlite3_column_decltype(rs->s, i));
            }
        }
        lua_pushboolean(L, 1);
    }
    return 1;
}

static int db_rs_gc(lua_State *L)
{
    db_rs_close(L, luaL_checkudata(L, 1, DB_RS_ID), 1);
    return 0;
}

static db_pstmt_t *get_db_pstmt(lua_State *L, int pos)
{
    db_pstmt_t *pstmt = luaL_checkudata(L, pos, DB_PSTMT_ID);
    if (pstmt->closed) {
        luaL_error(L, "prepared statement is closed");
    }
    return pstmt;
}

static int db_pstmt_tostr(lua_State *L)
{
    db_pstmt_t *pstmt = luaL_checkudata(L, 1, DB_PSTMT_ID);
    if (pstmt->closed) {
        lua_pushfstring(L, "prepared statement is closed");
    } else {
        lua_pushfstring(L, "prepared statement{handle=%p}", pstmt->s);
    }
    return 1;
}

static int bind(lua_State *L, sqlite3 *db, sqlite3_stmt *pstmt)
{
    int rc, i;
    int argc = lua_gettop(L) - 1;
    int param_count;

    param_count = sqlite3_bind_parameter_count(pstmt);
    if (argc != param_count) {
        lua_pushfstring(L, "parameter count mismatch: want %d got %d", param_count, argc);
        return -1;
    }

    rc = sqlite3_reset(pstmt);
    sqlite3_clear_bindings(pstmt);
    if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
        lua_pushfstring(L, sqlite3_errmsg(db));
        return -1;
    }

    for (i = 1; i <= argc; i++) {
        int t, b, n = i + 1;
        const char *s;
        size_t l;

        luaL_checkany(L, n);
        t = lua_type(L, n);

        switch (t) {
        case LUA_TNUMBER:
            if (luaL_isinteger(L, n)) {
                lua_Integer d = lua_tointeger(L, n);
                rc = sqlite3_bind_int64(pstmt, i, (sqlite3_int64)d);
            } else {
                lua_Number d = lua_tonumber(L, n);
                rc = sqlite3_bind_double(pstmt, i, (double)d);
            }
            break;
        case LUA_TSTRING:
            s = lua_tolstring(L, n, &l);
            rc = sqlite3_bind_text(pstmt, i, s, l, SQLITE_TRANSIENT);
            break;
        case LUA_TBOOLEAN:
            b = lua_toboolean(L, i+1);
            if (b) {
                rc = sqlite3_bind_int(pstmt, i, 1);
            } else {
                rc = sqlite3_bind_int(pstmt, i, 0);
            }
            break;
        case LUA_TNIL:
            rc = sqlite3_bind_null(pstmt, i);
            break;
        case LUA_TUSERDATA:
        {
            if (lua_isbignumber(L, n)) {
                long int d = lua_get_bignum_si(L, n);
                if (d == 0 && lua_bignum_is_zero(L, n) != 0) {
                    char *s = lua_get_bignum_str(L, n);
                    if (s != NULL) {
                        lua_pushfstring(L, "bignum value overflow for binding %s", s);
                        free(s);
                    }
                    return -1;
                }
                rc = sqlite3_bind_int64(pstmt, i, (sqlite3_int64)d);
                break;
            }
        }
        default:
            lua_pushfstring(L, "unsupported type: %s", lua_typename(L, n));
            return -1;
        }
        if (rc != SQLITE_OK) {
            lua_pushfstring(L, sqlite3_errmsg(db));
            return -1;
        }
    }
    
    return 0;
}

static int db_pstmt_exec(lua_State *L)
{
    int rc, n;
    db_pstmt_t *pstmt = get_db_pstmt(L, 1);

    /*check for exec in function */
	if (luaCheckView(getLuaExecContext(L)) > 0) {
        luaL_error(L, "not permitted in view function");
    }
    rc = bind(L, pstmt->db, pstmt->s);
    if (rc == -1) {
        sqlite3_reset(pstmt->s);
        sqlite3_clear_bindings(pstmt->s);
        luaL_error(L, lua_tostring(L, -1));
    }
    rc = sqlite3_step(pstmt->s);
    if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
        sqlite3_reset(pstmt->s);
        sqlite3_clear_bindings(pstmt->s);
        luaL_error(L, sqlite3_errmsg(pstmt->db));
    }
    n = sqlite3_changes(pstmt->db);
    lua_pushinteger(L, n);
    return 1;
}

static int db_pstmt_query(lua_State *L)
{
    int rc;
    db_pstmt_t *pstmt = get_db_pstmt(L, 1);
    db_rs_t *rs;

	getLuaExecContext(L);
    if (!sqlite3_stmt_readonly(pstmt->s)) {
        luaL_error(L, "invalid sql command(permitted readonly)");
    }
    rc = bind(L, pstmt->db, pstmt->s);
    if (rc != 0) {
        sqlite3_reset(pstmt->s);
        sqlite3_clear_bindings(pstmt->s);
        luaL_error(L, lua_tostring(L, -1));
    }

    rs = (db_rs_t *)lua_newuserdata(L, sizeof(db_rs_t));
    luaL_getmetatable(L, DB_RS_ID);
    lua_setmetatable(L, -2);
    rs->db = pstmt->db;
    rs->s = pstmt->s;
    rs->closed = 0;
    rs->nc = sqlite3_column_count(pstmt->s);
    rs->shared_stmt = 1;
    rs->decltypes = NULL;
    rs->refno = append_resource(L, RESOURCE_RS_KEY, (void *)rs);

    return 1;
}

static void get_column_meta(lua_State *L, sqlite3_stmt* stmt)
{
    const char *name, *decltype;
    int type;
    int colcnt = sqlite3_column_count(stmt);
    int i;

    lua_createtable(L, 0, 2);
    lua_pushinteger(L, colcnt);
    lua_setfield(L, -2, "colcnt");
    if (colcnt > 0) {
        lua_createtable(L, colcnt, 0);  /* colinfos names */
        lua_createtable(L, colcnt, 0);  /* colinfos names decltypes */
    }
    else {
        lua_pushnil(L);
        lua_pushnil(L);
    }
    for (i = 0; i < colcnt; i++) {
        name = sqlite3_column_name(stmt, i);
        if (name == NULL)
            lua_pushstring(L, "");
        else
            lua_pushstring(L, name);
        lua_rawseti(L, -3, i+1);

        decltype = sqlite3_column_decltype(stmt, i);
        if (decltype == NULL)
            lua_pushstring(L, "");
         else
            lua_pushstring(L, decltype);
        lua_rawseti(L, -2, i+1);
    }
    lua_setfield(L, -3, "decltypes");
    lua_setfield(L, -2, "names");
}

static int db_pstmt_column_info(lua_State *L)
{
    int colcnt;
    db_pstmt_t *pstmt = get_db_pstmt(L, 1);
	getLuaExecContext(L);

    get_column_meta(L, pstmt->s);
    return 1;
}

static int db_pstmt_bind_param_cnt(lua_State *L)
{
    db_pstmt_t *pstmt = get_db_pstmt(L, 1);
	getLuaExecContext(L);

	lua_pushinteger(L, sqlite3_bind_parameter_count(pstmt->s));

	return 1;
}

static void db_pstmt_close(lua_State *L, db_pstmt_t *pstmt, int remove)
{
    if (pstmt->closed)
        return;
    pstmt->closed = 1;
    sqlite3_finalize(pstmt->s);
    if (remove) {
        if (luaL_findtable(L, LUA_REGISTRYINDEX, RESOURCE_PSTMT_KEY, 0) != NULL) {
            luaL_error(L, "cannot find the environment of the db module");
        }
        luaL_unref(L, -1, pstmt->refno);
        lua_pop(L, 1);
    }
}

static int db_pstmt_gc(lua_State *L)
{
    db_pstmt_close(L, luaL_checkudata(L, 1, DB_PSTMT_ID), 1);
    return 0;
}

static int db_exec(lua_State *L)
{
    const char *cmd;
    sqlite3 *db;
    sqlite3_stmt *s;
    int rc;

    /*check for exec in function */
    if (luaCheckView(getLuaExecContext(L))> 0) {
        luaL_error(L, "not permitted in view function");
    }
    cmd = luaL_checkstring(L, 1);
    if (!sqlcheck_is_permitted_sql(cmd)) {
    	lua_pushfstring(L, "invalid sql commond:" LUA_QS, cmd);
        lua_error(L);
    }
    db = vm_get_db(L);
    rc = sqlite3_prepare_v2(db, cmd, -1, &s, NULL);
    LAST_ERROR(L, db, rc);

    rc = bind(L, db, s);
    if (rc == -1) {
        sqlite3_finalize(s);
        luaL_error(L, lua_tostring(L, -1));
    }

    rc = sqlite3_step(s);
    if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
        sqlite3_finalize(s);
        luaL_error(L, sqlite3_errmsg(db));
    }
    sqlite3_finalize(s);

    lua_pushinteger(L, sqlite3_changes(db));
    return 1;
}

static int db_query(lua_State *L)
{
    const char *query;
    int rc;
    sqlite3 *db;
    sqlite3_stmt *s;
    db_rs_t *rs;

	getLuaExecContext(L);
    query = luaL_checkstring(L, 1);
    if (!sqlcheck_is_readonly_sql(query)) {
        luaL_error(L, "invalid sql command(permitted readonly)");
    }
    db = vm_get_db(L);
    rc = sqlite3_prepare_v2(db, query, -1, &s, NULL);
    LAST_ERROR(L, db, rc);

    rc = bind(L, db, s);
    if (rc == -1) {
        sqlite3_finalize(s);
        luaL_error(L, lua_tostring(L, -1));
    }

    rs = (db_rs_t *)lua_newuserdata(L, sizeof(db_rs_t));
    luaL_getmetatable(L, DB_RS_ID);
    lua_setmetatable(L, -2);
    rs->db = db;
    rs->s = s;
    rs->closed = 0;
    rs->nc = sqlite3_column_count(s);
    rs->shared_stmt = 0;
    rs->decltypes = NULL;
    rs->refno = append_resource(L, RESOURCE_RS_KEY, (void *)rs);

    return 1;
}

static int db_prepare(lua_State *L)
{
    const char *sql;
    int rc;
    int ref;
    sqlite3 *db;
    sqlite3_stmt *s;
    db_pstmt_t *pstmt;

    sql = luaL_checkstring(L, 1);
    if (!sqlcheck_is_permitted_sql(sql)) {
    	lua_pushfstring(L, "invalid sql commond:" LUA_QS, sql);
        lua_error(L);
    }
    db = vm_get_db(L);
    rc = sqlite3_prepare_v2(db, sql, -1, &s, NULL);
    LAST_ERROR(L, db, rc);

    pstmt = (db_pstmt_t *)lua_newuserdata(L, sizeof(db_pstmt_t));
    luaL_getmetatable(L, DB_PSTMT_ID);
    lua_setmetatable(L, -2);
    pstmt->db = db;
    pstmt->s = s;
    pstmt->closed = 0;
    pstmt->refno = append_resource(L, RESOURCE_PSTMT_KEY, (void *)pstmt);

    return 1;
}

static int db_get_snapshot(lua_State *L)
{
    char *snapshot;
    int service = getLuaExecContext(L);

    snapshot = LuaGetDbSnapshot(service);
    strPushAndRelease(L, snapshot);

    return 1;
}

static int db_open_with_snapshot(lua_State *L)
{
    char *snapshot = (char *)luaL_checkstring(L, 1);
    char *errStr;
    int service = getLuaExecContext(L);

    errStr = LuaGetDbHandleSnap(service, snapshot);
    if (errStr != NULL) {
        strPushAndRelease(L, errStr);
        luaL_throwerror(L);
    }
    return 1;
}

static int db_last_insert_rowid(lua_State *L)
{
    sqlite3 *db;
    sqlite3_int64 id;
    db = vm_get_db(L);

    id = sqlite3_last_insert_rowid(db);
    lua_pushinteger(L, id);
    return 1;
}

int lua_db_release_resource(lua_State *L)
{
    lua_getfield(L, LUA_REGISTRYINDEX, RESOURCE_RS_KEY);
    if (lua_istable(L, -1)) {
        /* T */
        lua_pushnil(L); /* T nil(key) */
        while (lua_next(L, -2)) {
            if (lua_islightuserdata(L, -1))
                db_rs_close(L, (db_rs_t *)lua_topointer(L, -1), 0);
            lua_pop(L, 1);
        }
        lua_pop(L, 1);
    }
    lua_getfield(L, LUA_REGISTRYINDEX, RESOURCE_PSTMT_KEY);
    if (lua_istable(L, -1)) {
        /* T */
        lua_pushnil(L); /* T nil(key) */
        while (lua_next(L, -2)) {
            if (lua_islightuserdata(L, -1))
                db_pstmt_close(L, (db_pstmt_t *)lua_topointer(L, -1), 0);
            lua_pop(L, 1);
        }
        lua_pop(L, 1);
    }
    return 0;
}

int luaopen_db(lua_State *L)
{
    static const luaL_Reg rs_methods[] = {
        {"next",  db_rs_next},
        {"get", db_rs_get},
        {"colcnt", db_rs_colcnt},
        {"__tostring", db_rs_tostr},
        {"__gc", db_rs_gc},
        {NULL, NULL}
    };

    static const luaL_Reg pstmt_methods[] = {
        {"exec",  db_pstmt_exec},
        {"query", db_pstmt_query},
        {"column_info", db_pstmt_column_info},
        {"bind_param_cnt", db_pstmt_bind_param_cnt},
        {"__tostring", db_pstmt_tostr},
        {"__gc", db_pstmt_gc},
        {NULL, NULL}
    };

    static const luaL_Reg db_lib[] = {
        {"exec", db_exec},
        {"query", db_query},
        {"prepare", db_prepare},
        {"getsnap", db_get_snapshot},
        {"open_with_snapshot", db_open_with_snapshot},
        {"last_insert_rowid", db_last_insert_rowid},
        {NULL, NULL}
    };

    luaL_newmetatable(L, DB_RS_ID);
    lua_pushvalue(L, -1);
    lua_setfield(L, -2, "__index");
    luaL_register(L, NULL, rs_methods);

    luaL_newmetatable(L, DB_PSTMT_ID);
    lua_pushvalue(L, -1);
    lua_setfield(L, -2, "__index");
    luaL_register(L, NULL, pstmt_methods);

	luaL_register(L, "db", db_lib);
	lua_pop(L, 3);
	return 1;
}
