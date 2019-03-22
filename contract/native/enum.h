/**
 * @file    enum.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ENUM_H
#define _ENUM_H

#include "common.h"

#define TYPE_NAME(type)         type_names_[(type)]
#define TYPE_SIZE(type)         type_sizes_[(type)]
#define TYPE_BYTE(type)         type_bytes_[(type)]
#define TYPE_ALIGN              TYPE_SIZE

#define FN_NAME(kind)           fn_names_[(kind)]

#define is_valid_type(type)     (type > TYPE_NONE && type < TYPE_MAX)

typedef enum flag_val_e {
    FLAG_NONE       = 0x0000,
    FLAG_DEBUG      = 0x0001,
    FLAG_VERBOSE    = 0x0002,
    FLAG_DUMP_LEX   = 0x0004,
    FLAG_DUMP_YACC  = 0x0008,
    FLAG_DUMP_WAT   = 0x0010,
    FLAG_TEST       = 0x0020
} flag_val_t;

typedef enum ec_e {
    NO_ERROR = 0,
#undef error
#define error(code, msg)    code,
#include "error.list"

    ERROR_MAX
} ec_t;

typedef enum errlvl_e {
    LVL_FATAL       = 0,
    LVL_ERROR,
    LVL_INFO,
    LVL_WARN,
    LVL_DEBUG,

    LVL_MAX
} errlvl_t;

typedef enum type_e {
    TYPE_NONE       = 0,
    TYPE_BOOL       = 1,
    TYPE_INT8       = 2,
    TYPE_UINT8      = 3,
    TYPE_INT16      = 4,
    TYPE_UINT16     = 5,
    TYPE_INT32      = 6,
    TYPE_UINT32     = 7,
    TYPE_INT64      = 8,
    TYPE_UINT64     = 9,
    TYPE_INT128     = 10,
    TYPE_UINT128    = 11,
    TYPE_FLOAT      = 12,
    TYPE_DOUBLE     = 13,
    TYPE_STRING     = 14,
    TYPE_COMPATIBLE = TYPE_STRING,

    TYPE_ACCOUNT    = 15,
    TYPE_STRUCT     = 16,
    TYPE_COMPARABLE = TYPE_STRUCT,

    TYPE_MAP        = 17,
    TYPE_OBJECT     = 18,           /* contract, interface or null */
    TYPE_CURSOR     = 19,

    TYPE_VOID       = 20,           /* return type of function */
    TYPE_TUPLE      = 21,           /* tuple expression */
    TYPE_MAX
} type_t;

typedef enum id_kind_e {
    ID_VAR          = 0,
    ID_STRUCT,
    ID_ENUM,
    ID_FN,
    ID_CONT,
    ID_ITF,
    ID_LIB,
    ID_LABEL,
    ID_TUPLE,
    ID_MAX
} id_kind_t;

typedef enum exp_kind_e {
    EXP_NULL        = 0,
    EXP_LIT,
    EXP_ID,
    EXP_TYPE,
    EXP_ARRAY,
    EXP_CAST,
    EXP_UNARY,
    EXP_BINARY,
    EXP_TERNARY,
    EXP_ACCESS,
    EXP_CALL,
    EXP_SQL,
    EXP_TUPLE,
    EXP_INIT,
    EXP_ALLOC,
    EXP_GLOBAL,
    EXP_REG,
    EXP_MEM,
    EXP_MAX
} exp_kind_t;

typedef enum stmt_kind_e {
    STMT_NULL       = 0,
    STMT_ID,
    STMT_EXP,
    STMT_ASSIGN,
    STMT_IF,
    STMT_LOOP,
    STMT_SWITCH,
    STMT_CASE,
    STMT_CONTINUE,
    STMT_BREAK,
    STMT_RETURN,
    STMT_GOTO,
    STMT_DDL,
    STMT_BLK,
    STMT_MAX
} stmt_kind_t;

typedef enum blk_kind_e {
    BLK_NORMAL      = 0,
    BLK_ROOT,
    BLK_CONT,
    BLK_ITF,
    BLK_LIB,
    BLK_FN,
    BLK_LOOP,
    BLK_SWITCH,
    BLK_MAX
} blk_kind_t;

typedef enum modifier_e {
    MOD_PRIVATE     = 0x00,
    MOD_PUBLIC      = 0x01,
    MOD_PAYABLE     = 0x02,
    MOD_READONLY    = 0x04,
    MOD_CONST       = 0x08,
    MOD_CTOR        = 0x10,
    MOD_SYSTEM      = 0x20
} modifier_t;

typedef enum op_kind_e {
    /* constant folding is possible up to OP_OR */
    OP_ADD          = 0,
    OP_SUB,
    OP_MUL,
    OP_DIV,
    OP_MOD,
    OP_EQ,
    OP_NE,
    OP_LT,
    OP_GT,
    OP_LE,
    OP_GE,
    OP_BIT_AND,
    OP_BIT_OR,
    OP_BIT_XOR,
    OP_RSHIFT,
    OP_LSHIFT,
    OP_NEG,
    OP_NOT,
    OP_AND,
    OP_OR,
    OP_CF_MAX      = OP_OR,

    OP_INC,
    OP_DEC,
    OP_MAX
} op_kind_t;

typedef enum sql_kind_e {
    SQL_QUERY       = 0,
    SQL_INSERT,
    SQL_UPDATE,
    SQL_DELETE,
    SQL_REPLACE,
    SQL_MAX
} sql_kind_t;

typedef enum loop_kind_e {
    LOOP_FOR        = 0,
    LOOP_ARRAY,
    LOOP_MAX
} loop_kind_t;

typedef enum fn_kind_e {
    FN_MALLOC       = 0,
    FN_MEMCPY,
    FN_STRCAT,
    FN_STRCMP,
    FN_ATOI32,
    FN_ATOI64,
    FN_ATOF32,
    FN_ATOF64,
    FN_ITOA32,
    FN_ITOA64,
    FN_FTOA32,
    FN_FTOA64,
    FN_MAP_NEW,
    FN_MAP_GET,
    FN_MAP_SET,
    FN_MAX
} fn_kind_t;

extern char *type_names_[TYPE_MAX];
extern int type_sizes_[TYPE_MAX];
extern int type_bytes_[TYPE_MAX];

#endif /* ! _ENUM_H */
