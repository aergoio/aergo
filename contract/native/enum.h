/**
 * @file    enum.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ENUM_H
#define _ENUM_H

#include "common.h"

#define TYPE_NAME(type)             type_names_[(type)]
#define TYPE_SIZE(type)             type_sizes_[(type)]
#define TYPE_C_SIZE(type)           type_c_sizes_[(type)]
#define TYPE_ALIGN                  TYPE_SIZE

#define ADDR_SIZE                   type_sizes_[TYPE_INT32]

#define is_valid_type(type)         (type > TYPE_NONE && type < TYPE_MAX)

typedef enum flag_val_e {
    FLAG_NONE       = 0x0000,
    FLAG_DEBUG      = 0x0001,
    FLAG_VERBOSE    = 0x0002,
    FLAG_DUMP_LEX   = 0x0004,
    FLAG_DUMP_YACC  = 0x0008,
    FLAG_DUMP_WAT   = 0x0010,
    FLAG_MAX        = 0XFFFF
} flag_val_t;

typedef enum ec_e {
    NO_ERROR = 0,
#undef error
#define error(code, msg)            code,
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
    TYPE_BYTE       = 2,
    TYPE_INT8       = 3,
    TYPE_INT16      = 4,
    TYPE_INT32      = 5,
    TYPE_INT64      = 6,
    TYPE_INT128     = 7,
    TYPE_INT256     = 8,
    TYPE_FLOAT      = 9,
    TYPE_DOUBLE     = 10,
    TYPE_STRING     = 11,
    TYPE_COMPATIBLE = TYPE_STRING,

    TYPE_ACCOUNT    = 12,
    TYPE_STRUCT     = 13,
    TYPE_COMPARABLE = TYPE_STRUCT,

    TYPE_MAP        = 14,
    TYPE_OBJECT     = 15,           /* contract, interface or null */
    TYPE_CURSOR     = 16,

    TYPE_VOID       = 17,           /* return type of function */
    TYPE_TUPLE      = 18,           /* tuple expression */
    TYPE_MAX
} type_t;

typedef enum id_kind_e {
    ID_VAR          = 0,
    ID_STRUCT       = 1,
    ID_ENUM         = 2,
    ID_FN           = 3,
    ID_CONT         = 4,
    ID_ITF          = 5,
    ID_LIB          = 6,
    ID_LABEL        = 7,
    ID_TUPLE        = 8,
    ID_MAX
} id_kind_t;

typedef enum exp_kind_e {
    EXP_NULL        = 0,
    EXP_LIT         = 1,
    EXP_ID          = 2,
    EXP_TYPE        = 3,
    EXP_ARRAY       = 4,
    EXP_CAST        = 5,
    EXP_UNARY       = 6,
    EXP_BINARY      = 7,
    EXP_TERNARY     = 8,
    EXP_ACCESS      = 9,
    EXP_CALL        = 10,
    EXP_SQL         = 11,
    EXP_TUPLE       = 12,
    EXP_INIT        = 13,
    EXP_ALLOC       = 14,
    EXP_GLOBAL      = 15,
    EXP_REG         = 16,
    EXP_MEM         = 17,
    EXP_MAX
} exp_kind_t;

typedef enum stmt_kind_e {
    STMT_NULL       = 0,
    STMT_ID         = 1,
    STMT_EXP        = 2,
    STMT_ASSIGN     = 3,
    STMT_IF         = 4,
    STMT_LOOP       = 5,
    STMT_SWITCH     = 6,
    STMT_CASE       = 7,
    STMT_CONTINUE   = 8,
    STMT_BREAK      = 9,
    STMT_RETURN     = 10,
    STMT_GOTO       = 11,
    STMT_DDL        = 12,
    STMT_BLK        = 13,
    STMT_PRAGMA     = 14,
    STMT_MAX
} stmt_kind_t;

typedef enum blk_kind_e {
    BLK_NORMAL      = 0,
    BLK_ROOT        = 1,
    BLK_CONT        = 2,
    BLK_ITF         = 3,
    BLK_LIB         = 4,
    BLK_LOOP        = 5,
    BLK_SWITCH      = 6,
    BLK_MAX
} blk_kind_t;

typedef enum modifier_e {
    MOD_PRIVATE     = 0x00,
    MOD_PUBLIC      = 0x01,
    MOD_PAYABLE     = 0x02,
    MOD_READONLY    = 0x04,
    MOD_CONST       = 0x08,
    MOD_CTOR        = 0x10
} modifier_t;

typedef enum op_kind_e {
    /* constant folding is possible up to OP_OR */
    OP_ADD          = 0,
    OP_SUB          = 1,
    OP_MUL          = 2,
    OP_DIV          = 3,
    OP_MOD          = 4,
    OP_EQ           = 5,
    OP_NE           = 6,
    OP_LT           = 7,
    OP_GT           = 8,
    OP_LE           = 9,
    OP_GE           = 10,
    OP_BIT_AND      = 11,
    OP_BIT_OR       = 12,
    OP_BIT_XOR      = 13,
    OP_RSHIFT       = 14,
    OP_LSHIFT       = 15,
    OP_NEG          = 16,
    OP_NOT          = 17,
    OP_AND          = 18,
    OP_OR           = 19,
    OP_CF_MAX       = OP_OR,

    OP_INC          = 20,
    OP_DEC          = 21,
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

typedef enum pragma_kind_e {
    PRAGMA_ASSERT   = 0,
    PRAGMA_MAX
} pragma_kind_t;

typedef enum fn_kind_e {
    FN_UDF          = 0,
    FN_CTOR,
#undef fn_def
#define fn_def(kind, name, ...)		kind,
#include "syslib.list"
    FN_MAX
} fn_kind_t;

extern char *type_names_[TYPE_MAX];
extern int type_sizes_[TYPE_MAX];
extern int type_c_sizes_[TYPE_MAX];

#endif /* ! _ENUM_H */
