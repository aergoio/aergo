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
    FLAG_TEST       = 0x20
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
#define TYPE_ALIGN(type)        type_aligns_[(type)]

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
    TYPE_PRIMITIVE  = TYPE_DOUBLE,

    TYPE_STRING     = 13,
    TYPE_COMPATIBLE = TYPE_STRING,

    TYPE_ACCOUNT    = 14,
    TYPE_STRUCT     = 15,
    TYPE_COMPARABLE = TYPE_STRUCT,

    TYPE_MAP        = 16,
    TYPE_OBJECT     = 17,           /* new contract() or null */
    TYPE_BUILTIN    = TYPE_OBJECT,

    TYPE_VOID       = 18,           /* for function */
    TYPE_TUPLE      = 19,           /* for tuple expression */
    TYPE_MAX
} type_t;

extern char *type_names_[TYPE_MAX];
extern int type_sizes_[TYPE_MAX];
extern int type_aligns_[TYPE_MAX];

#define ID_KIND(id)             id_kinds_[(id)->kind]

typedef enum id_kind_e {
    ID_VAR          = 0,
    ID_STRUCT,
    ID_ENUM,
    ID_RETURN,
    ID_FN,
    ID_CONTRACT,
    ID_LABEL,
    ID_TUPLE,
    ID_MAX
} id_kind_t;

extern char *id_kinds_[ID_MAX];

typedef enum scope_e {
    SCOPE_LOCAL     = 0,
    SCOPE_GLOBAL,
    SCOPE_MAX
} scope_t;

typedef enum exp_kind_e {
    EXP_NULL        = 0,
    EXP_LIT,
    EXP_ID_REF,
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
    EXP_LOCAL_REF,
    EXP_STACK_REF,
    EXP_MAX
} exp_kind_t;

#define STMT_KIND(stmt)         stmt_kinds_[(stmt)->kind]

typedef enum stmt_kind_e {
    STMT_NULL       = 0,
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

extern char *stmt_kinds_[STMT_MAX];

typedef enum blk_kind_e {
    BLK_NORMAL      = 0,
    BLK_ROOT,
    BLK_CONTRACT,
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

typedef enum sym_kind_e {
    SYM_NONE        = 0,
    SYM_VAR,
    SYM_REC,
    SYM_FUNC,
    SYM_MAX
} sym_kind_t;

#endif /* ! _ENUM_H */
