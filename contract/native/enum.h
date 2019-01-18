/**
 * @file    enum.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ENUM_H
#define _ENUM_H

#include "common.h"

typedef enum flag_e {
    FLAG_NONE       = 0x00,
    FLAG_DEBUG      = 0x01,
    FLAG_OPT        = 0x02,
    FLAG_VERBOSE    = 0x04,
    FLAG_LEX_DUMP   = 0x08,
    FLAG_YACC_DUMP  = 0x10,
    FLAG_WAT_DUMP   = 0x20,
    FLAG_TEST       = 0x40
} flag_t;

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

#define TYPE_NAME(type)         type_names_[(type)]
#define TYPE_SIZE(type)         type_sizes_[(type)]
#define TYPE_BYTE(type)         type_bytes_[(type)]
#define TYPE_ALIGN              TYPE_SIZE

#define is_valid_type(type)     (type > TYPE_NONE && type < TYPE_MAX)

typedef enum type_e {
    TYPE_NONE       = 0,
    TYPE_BOOL       = 1,
    TYPE_BYTE       = 2,
    TYPE_INT8       = 3,
    TYPE_UINT8      = 4,
    TYPE_INT16      = 5,
    TYPE_UINT16     = 6,
    TYPE_INT32      = 7,
    TYPE_UINT32     = 8,
    TYPE_INT64      = 9,
    TYPE_UINT64     = 10,
    TYPE_FLOAT      = 11,
    TYPE_DOUBLE     = 12,
    TYPE_STRING     = 13,
    TYPE_COMPATIBLE = TYPE_STRING,

    TYPE_ACCOUNT    = 14,
    TYPE_STRUCT     = 15,
    TYPE_COMPARABLE = TYPE_STRUCT,

    TYPE_MAP        = 16,
    TYPE_OBJECT     = 17,           /* contract, interface or null */

    TYPE_VOID       = 18,           /* return type of function */
    TYPE_TUPLE      = 19,           /* tuple expression */
    TYPE_MAX
} type_t;

extern char *type_names_[TYPE_MAX];
extern int type_sizes_[TYPE_MAX];
extern int type_bytes_[TYPE_MAX];

typedef enum node_kind_e {
    ID_START        = 0,
    ID_VAR          = ID_START,
    ID_STRUCT       = 1,
    ID_ENUM         = 2,
    ID_FN           = 3,
    ID_CONT         = 4,
    ID_ITF          = 5,
    ID_LABEL        = 6,
    ID_TUPLE        = 7,
    ID_MAX          = ID_TUPLE,

    EXP_START       = 8,
    EXP_NULL        = EXP_START,
    EXP_LIT         = 9,
    EXP_ID          = 10,
    EXP_TYPE        = 11,
    EXP_ARRAY       = 12,
    EXP_CAST        = 13,
    EXP_UNARY       = 14,
    EXP_BINARY      = 15,
    EXP_TERNARY     = 16,
    EXP_ACCESS      = 17,
    EXP_CALL        = 18,
    EXP_SQL         = 19,
    EXP_TUPLE       = 20,
    EXP_INIT        = 21,
    EXP_ALLOC       = 22,
    EXP_GLOBAL      = 23,
    EXP_LOCAL       = 24,
    EXP_STACK       = 25,
    EXP_MAX         = EXP_STACK,

    STMT_START      = 26,
    STMT_NULL       = 27,
    STMT_EXP        = 28,
    STMT_ASSIGN     = 29,
    STMT_IF         = 30,
    STMT_LOOP       = 31,
    STMT_SWITCH     = 32,
    STMT_CASE       = 33,
    STMT_CONTINUE   = 34,
    STMT_BREAK      = 35,
    STMT_RETURN     = 36,
    STMT_GOTO       = 37,
    STMT_DDL        = 38,
    STMT_BLK        = 39,
    STMT_MAX        = STMT_BLK,

    NODE_MAX
} node_kind_t;

typedef enum blk_kind_e {
    BLK_NORMAL      = 0,
    BLK_ROOT        = 1,
    BLK_CONT        = 2,
    BLK_ITF         = 3,
    BLK_FN          = 4,
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
    SQL_MAX
} sql_kind_t;

typedef enum loop_kind_e {
    LOOP_FOR        = 0,
    LOOP_ARRAY,
    LOOP_MAX
} loop_kind_t;

typedef enum param_kind_e {
    PARAM_NONE      = 0,
    PARAM_IN,
    PARAM_OUT,
    PARAM_MAX
} param_kind_t;

#endif /* ! _ENUM_H */
