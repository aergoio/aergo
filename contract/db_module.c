#include <string.h>
#include <stdlib.h>
#include <ctype.h>
#include <sqlite3-binding.h>
#include "vm.h"
#include "sqlcheck.h"

extern const bc_ctx_t *getLuaExecContext(lua_State *L);

#define LAST_ERROR(L,db,rc)                         \
    do {                                            \
        if ((rc) != SQLITE_OK) {                    \
            luaL_error((L), sqlite3_errmsg((db)));  \
        }                                           \
    } while(0)

#define DB_PSTMT_ID "__db_pstmt__"

typedef struct {
    sqlite3 *db;
    sqlite3_stmt *s;
    int closed;
    int ref;
    int refcnt;
} db_pstmt_t;

#define DB_RS_ID "__db_rs__"

typedef struct {
    sqlite3 *db;
    sqlite3_stmt *s;
    int closed;
    int nc;
    int ref;
    int *is_boolean;
} db_rs_t;

static void db_pstmt_close(lua_State *L, int ref);

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

static int is_booleantype(const char *tname)
{
    if (tname == NULL || strlen(tname) != 7)
        return 0;
    if (tname[0] != 'B' && tname[0] != 'b')
        return 0;
    if (tname[1] != 'O' && tname[0] != 'o')
        return 0;
    if (tname[2] != 'O' && tname[2] != 'o')
        return 0;
    if (tname[3] != 'L' && tname[3] != 'l')
        return 0;
    if (tname[4] != 'E' && tname[4] != 'e')
        return 0;
    if (tname[5] != 'A' && tname[5] != 'a')
        return 0;
    if (tname[6] != 'N' && tname[6] != 'n')
        return 0;
    return 1;
}

static int db_rs_get(lua_State *L)
{
    db_rs_t *rs = get_db_rs(L, 1);
    int i;
    sqlite3_int64 d;
    double f;
    int n;
    const unsigned char *s;

    if (rs->is_boolean == NULL) {
        luaL_error(L, "`get' called without calling `next'");
    }
    for (i = 0; i < rs->nc; i++) {
        switch (sqlite3_column_type(rs->s, i)) {
        case SQLITE_INTEGER:
            d = sqlite3_column_int64(rs->s, i);
            if (rs->is_boolean[i])  {
                if (d != 0) {
                    lua_pushboolean(L, 1);
                } else {
                    lua_pushboolean(L, 0);
                }
            } else {
                lua_pushinteger(L, d);
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

static void db_rs_close(lua_State *L, db_rs_t *rs)
{
    if (rs->closed) {
        return;
    }
    rs->closed = 1;
    if (rs->is_boolean) {
        free(rs->is_boolean);
        rs->is_boolean = NULL;
    }
    if (rs->ref == -1) {
        sqlite3_finalize(rs->s);
    } else {
        db_pstmt_t *pstmt;
        lua_rawgeti(L, LUA_REGISTRYINDEX, rs->ref);
        pstmt = (db_pstmt_t *)luaL_checkudata(L, -1, DB_PSTMT_ID);
        pstmt->refcnt--;
        if (pstmt->refcnt == 0) {
            db_pstmt_close(L, rs->ref);
        }
    }
}

static int db_rs_next(lua_State *L)
{
    db_rs_t *rs = get_db_rs(L, 1);
    int rc;

    rc = sqlite3_step(rs->s);
    if (rc == SQLITE_DONE) {
        db_rs_close(L, rs);
        lua_pushboolean(L, 0);
    } else if (rc != SQLITE_ROW) {
        rc = sqlite3_reset(rs->s);
        LAST_ERROR(L, rs->db, rc);
        db_rs_close(L, rs);
        lua_pushboolean(L, 0);
    } else {
        if (rs->is_boolean == NULL) {
            int i;
            rs->is_boolean = malloc(sizeof(int) * rs->nc);
            for (i = 0; i < rs->nc; i++) {
                rs->is_boolean[i] = is_booleantype(sqlite3_column_decltype(rs->s, i));
            }
        }
        lua_pushboolean(L, 1);
    }
    return 1;
}

static int db_rs_gc(lua_State *L)
{
    db_rs_t *rs = luaL_checkudata(L, 1, DB_RS_ID);

    if (rs->closed) {
        return 0;
    }
    rs->closed = 1;
    if (rs->is_boolean) {
        free(rs->is_boolean);
        rs->is_boolean = NULL;
    }
    if (rs->ref == -1) {
        sqlite3_finalize(rs->s);
    } else {
        db_pstmt_t *pstmt;
        lua_rawgeti(L, LUA_REGISTRYINDEX, rs->ref);
        pstmt = (db_pstmt_t *)luaL_checkudata(L, -1, DB_PSTMT_ID);
        pstmt->refcnt--;
        if (pstmt->refcnt == 0) {
            db_pstmt_close(L, rs->ref);
        }
    }
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

static int bind(lua_State *L, db_pstmt_t *pstmt)
{
    int rc, i;
    int argc = lua_gettop(L) - 1;

    rc = sqlite3_reset(pstmt->s);
    if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
        lua_pushfstring(L, sqlite3_errmsg(pstmt->db));
        return -1;
    }

    for (i = 1; i <= argc; i++) {
        int t, b, n = i + 1;
        const char *s;
        size_t l;
        lua_Number d;

        luaL_checkany(L, n);
        t = lua_type(L, n);

        switch (t) {
        case LUA_TNUMBER:
            d = lua_tonumber(L, n);
            if ((double)d == (double)((lua_Integer)d)) {
                rc = sqlite3_bind_int64(pstmt->s, i, (sqlite3_int64)d);
            } else {
                rc = sqlite3_bind_double(pstmt->s, i, (sqlite3_int64)d);
            }
            break;
        case LUA_TSTRING:
            s = lua_tolstring(L, n, &l);
            rc = sqlite3_bind_text(pstmt->s, i, s, l, SQLITE_TRANSIENT);
            break;
        case LUA_TBOOLEAN:
            b = lua_toboolean(L, i+1);
            if (b) {
                rc = sqlite3_bind_int(pstmt->s, i, 1);
            } else {
                rc = sqlite3_bind_int(pstmt->s, i, 0);
            }
            break;
        case LUA_TNIL:
            rc = sqlite3_bind_null(pstmt->s, i);
            break;
        default:
            lua_pushfstring(L, "unsupported type: %s", lua_typename(L, n));
            return -1;
        }
        if (rc != SQLITE_OK) {
            lua_pushfstring(L, sqlite3_errmsg(pstmt->db));
            return -1;
        }
    }
    
    return 0;
}

static int db_pstmt_exec(lua_State *L)
{
    int rc, n;
    db_pstmt_t *pstmt = get_db_pstmt(L, 1);

    rc = bind(L, pstmt);
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

    rc = bind(L, pstmt);
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
    rs->ref = pstmt->ref;
    rs->is_boolean = NULL;
    pstmt->refcnt++;
    return 1;
}

static void db_pstmt_close(lua_State *L, int ref)
{
    db_pstmt_t *pstmt = luaL_checkudata(L, -1, DB_PSTMT_ID);

    if (!pstmt->closed) {
        pstmt->closed = 1;
        sqlite3_finalize(pstmt->s);
        luaL_unref(L, LUA_REGISTRYINDEX, ref);
    }
}

static int db_pstmt_gc(lua_State *L)
{
    db_pstmt_t *pstmt = luaL_checkudata(L, 1, DB_PSTMT_ID);

    if (!pstmt->closed) {
        pstmt->closed = 1;
        sqlite3_finalize(pstmt->s);
    }
    return 0;
}

static int db_exec(lua_State *L)
{
    const bc_ctx_t *ctx;
    const char *cmd;
    int rc, n;

    ctx = getLuaExecContext(L);
    cmd = luaL_checkstring(L, 1);
    if (!sqlcheck_is_permitted_sql(cmd)) {
        luaL_error(L, "invalid sql command");
    }
    rc = sqlite3_exec(ctx->db, cmd, 0, 0, 0);
    LAST_ERROR(L, ctx->db, rc);
    n = sqlite3_changes(ctx->db);
    lua_pushinteger(L, n);
    return 1;
}

static int db_query(lua_State *L)
{
    const bc_ctx_t *ctx;
    const char *query;
    int rc;
    sqlite3_stmt *s;
    db_rs_t *rs;

    ctx = getLuaExecContext(L);
    query = luaL_checkstring(L, 1);
    if (!sqlcheck_is_permitted_sql(query)) {
        luaL_error(L, "invalid sql command");
    }
    rc = sqlite3_prepare_v2(ctx->db, query, -1, &s, NULL);
    LAST_ERROR(L, ctx->db, rc);

    rs = (db_rs_t *)lua_newuserdata(L, sizeof(db_rs_t));
    luaL_getmetatable(L, DB_RS_ID);
    lua_setmetatable(L, -2);
    rs->db = ctx->db;
    rs->s = s;
    rs->closed = 0;
    rs->nc = sqlite3_column_count(s);
    rs->ref = -1;
    rs->is_boolean = NULL;
    return 1;
}

static int db_prepare(lua_State *L)
{
    const bc_ctx_t *ctx;
    const char *sql;
    int rc;
    int ref;
    sqlite3_stmt *s;
    db_pstmt_t *pstmt;

    ctx = getLuaExecContext(L);
    sql = luaL_checkstring(L, 1);
    if (!sqlcheck_is_permitted_sql(sql)) {
        luaL_error(L, "invalid sql command");
    }
    rc = sqlite3_prepare_v2(ctx->db, sql, -1, &s, NULL);
    LAST_ERROR(L, ctx->db, rc);

    pstmt = (db_pstmt_t *)lua_newuserdata(L, sizeof(db_pstmt_t));
    luaL_getmetatable(L, DB_PSTMT_ID);
    lua_setmetatable(L, -2);
    lua_pushvalue(L, -1);
    ref = luaL_ref(L, LUA_REGISTRYINDEX);
    pstmt->db = ctx->db;
    pstmt->s = s;
    pstmt->closed = 0;
    pstmt->refcnt = 0;
    pstmt->ref = ref;

    return 1;
}

int luaopen_db(lua_State *L)
{
    static const luaL_Reg rs_methods[] = {
        {"next",  db_rs_next},
        {"get", db_rs_get},
        {"__tostring", db_rs_tostr},
        {"__gc", db_rs_gc},
        {NULL, NULL}
    };

    static const luaL_Reg pstmt_methods[] = {
        {"exec",  db_pstmt_exec},
        {"query", db_pstmt_query},
        {"__tostring", db_pstmt_tostr},
        {"__gc", db_pstmt_gc},
        {NULL, NULL}
    };

    static const luaL_Reg db_lib[] = {
        {"exec", db_exec},
        {"query", db_query},
        {"prepare", db_prepare},
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
	return 1;
}
