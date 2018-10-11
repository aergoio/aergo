/**
 * @file    enum.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ENUM_H
#define _ENUM_H

#include "common.h"

#define TYPE_NAME(type)         type_names_[(type)]

typedef enum flag_e {
    FLAG_NONE       = 0x00,
    FLAG_VERBOSE    = 0x01,
    FLAG_LEX_DUMP   = 0x02,
    FLAG_YACC_DUMP  = 0x04,
    FLAG_AST_DUMP   = 0x08,
    FLAG_TEST       = 0x10
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

typedef enum type_e {
    TYPE_NONE       = 0,
    TYPE_BOOL,
    TYPE_BYTE,
    TYPE_INT8,
    TYPE_UINT8,
    TYPE_INT16,
    TYPE_UINT16,
    TYPE_INT32,
    TYPE_UINT32,
    TYPE_INT64,
    TYPE_UINT64,
    TYPE_FLOAT,
    TYPE_DOUBLE,
    TYPE_STRING,
    TYPE_STRUCT,
    TYPE_REF,
    TYPE_ACCOUNT,
    TYPE_COMPARABLE = TYPE_ACCOUNT,

    TYPE_MAP,
    TYPE_PRIMITIVE  = TYPE_MAP,

    TYPE_VOID,                      /* for return type of function */
    TYPE_TUPLE,                     /* for tuple expression */
    TYPE_MAX
} type_t;

typedef enum id_kind_e {
    ID_VAR          = 0,
    ID_STRUCT,
    ID_FUNC,
    ID_CONTRACT,
    ID_MAX
} id_kind_t;

typedef enum exp_kind_e {
    EXP_NULL        = 0,
    EXP_VAL,
    EXP_TYPE,
    EXP_ID,
    EXP_ARRAY,
    EXP_OP,
    EXP_ACCESS,
    EXP_CALL,
    EXP_SQL,
    EXP_TERNARY,
    EXP_TUPLE,
    EXP_MAX
} exp_kind_t;

typedef enum stmt_kind_e {
    STMT_NULL       = 0,
    STMT_EXP,
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
    BLK_ANON        = 0,
    BLK_ROOT,
    BLK_LOOP,
    BLK_MAX
} blk_kind_t;

typedef enum modifier_e {
    MOD_GLOBAL      = 0x00,
    MOD_LOCAL       = 0x01,
    MOD_PAYABLE     = 0x02,
    MOD_READONLY    = 0x04,
    MOD_CTOR        = 0x08
} modifier_t;

typedef enum op_kind_e {
    /* constant folding is possible up to OP_NOT */
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
    OP_NOT,
    OP_CF_MAX,                      

    /* short-circuit evaluation is possible up to OP_OR */
    OP_AND          = OP_CF_MAX,
    OP_OR,
    OP_SCE_MAX,

    OP_INC          = OP_SCE_MAX,
    OP_DEC,
    OP_ASSIGN,
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
    LOOP_EACH,
    LOOP_MAX
} loop_kind_t;

typedef enum val_kind_e {
    VAL_NULL        = 0,
    VAL_BOOL,
    VAL_INT,
    VAL_FP,
    VAL_STR,
    VAL_MAX
} val_kind_t;

extern char *type_names_[TYPE_MAX];

#endif /* ! _ENUM_H */
