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
    ID_STRUCT       = 1,
    ID_ENUM         = 2,
    ID_RETURN       = 3,
    ID_FN           = 4,
    ID_CONT         = 5,
    ID_ITF          = 6,
    ID_LABEL        = 7,
    ID_TUPLE        = 8,
    ID_MAX
} id_kind_t;

extern char *id_kinds_[ID_MAX];

typedef enum exp_kind_e {
    EXP_NULL        = 0,
    EXP_LIT         = 1,
    EXP_ID          = 2,
    EXP_ARRAY       = 3,
    EXP_CAST        = 4,
    EXP_UNARY       = 5,
    EXP_BINARY      = 6,
    EXP_TERNARY     = 7,
    EXP_ACCESS      = 8,
    EXP_CALL        = 9,
    EXP_SQL         = 10,
    EXP_TUPLE       = 11,
    EXP_INIT        = 12,
    EXP_GLOBAL      = 13,
    EXP_LOCAL       = 14,
    EXP_STACK       = 15,
    EXP_MAX
} exp_kind_t;

#define STMT_KIND(stmt)         stmt_kinds_[(stmt)->kind]

typedef enum stmt_kind_e {
    STMT_NULL       = 0,
    STMT_EXP        = 1,
    STMT_ASSIGN     = 2,
    STMT_IF         = 3,
    STMT_LOOP       = 4,
    STMT_SWITCH     = 5,
    STMT_CASE       = 6,
    STMT_CONTINUE   = 7,
    STMT_BREAK      = 8,
    STMT_RETURN     = 9,
    STMT_GOTO       = 10,
    STMT_DDL        = 11,
    STMT_BLK        = 12,
    STMT_MAX
} stmt_kind_t;

extern char *stmt_kinds_[STMT_MAX];

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

typedef enum sym_kind_e {
    SYM_NONE        = 0,
    SYM_VAR,
    SYM_REC,
    SYM_FUNC,
    SYM_MAX
} sym_kind_t;

#endif /* ! _ENUM_H */
