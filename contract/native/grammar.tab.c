/* A Bison parser, made by GNU Bison 3.0.5.  */

/* Bison implementation for Yacc-like parsers in C

   Copyright (C) 1984, 1989-1990, 2000-2015, 2018 Free Software Foundation, Inc.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.  */

/* As a special exception, you may create a larger work that contains
   part or all of the Bison parser skeleton and distribute that work
   under terms of your choice, so long as that work isn't itself a
   parser generator using the skeleton or a modified version thereof
   as a parser skeleton.  Alternatively, if you modify or redistribute
   the parser skeleton itself, you may (at your option) remove this
   special exception, which will cause the skeleton and the resulting
   Bison output files to be licensed under the GNU General Public
   License without this special exception.

   This special exception was added by the Free Software Foundation in
   version 2.2 of Bison.  */

/* C LALR(1) parser skeleton written by Richard Stallman, by
   simplifying the original so-called "semantic" parser.  */

/* All symbols defined below should begin with yy or YY, to avoid
   infringing on user name space.  This should be done even for local
   variables, as they might otherwise be expanded by user macros.
   There are some unavoidable exceptions within include files to
   define necessary library symbols; they are noted "INFRINGES ON
   USER NAME SPACE" below.  */

/* Identify Bison output.  */
#define YYBISON 1

/* Bison version.  */
#define YYBISON_VERSION "3.0.5"

/* Skeleton name.  */
#define YYSKELETON_NAME "yacc.c"

/* Pure parsers.  */
#define YYPURE 2

/* Push parsers.  */
#define YYPUSH 0

/* Pull parsers.  */
#define YYPULL 1




/* Copy the first part of user declarations.  */
#line 1 "grammar.y" /* yacc.c:339  */


/**
 * @file    grammar.y
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "error.h"
#include "parse.h"

#define YYLLOC_DEFAULT(Current, Rhs, N)                                                  \
    (Current) = YYRHSLOC(Rhs, (N) > 0 ? 1 : 0)

#define AST             (parse->ast)
#define ROOT            AST->root
#define LABELS          (&parse->labels)

extern int yylex(YYSTYPE *yylval, YYLTYPE *yylloc, void *yyscanner);
extern void yylex_set_token(void *yyscanner, int token, YYLTYPE *yylloc);

static void yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner,
                    const char *msg);


#line 94 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:339  */

# ifndef YY_NULLPTR
#  if defined __cplusplus && 201103L <= __cplusplus
#   define YY_NULLPTR nullptr
#  else
#   define YY_NULLPTR 0
#  endif
# endif

/* Enabling verbose error messages.  */
#ifdef YYERROR_VERBOSE
# undef YYERROR_VERBOSE
# define YYERROR_VERBOSE 1
#else
# define YYERROR_VERBOSE 1
#endif

/* In a future release of Bison, this section will be replaced
   by #include "grammar.tab.h".  */
#ifndef YY_YY_HOME_WRPARK_BLOCKO_SRC_GITHUB_COM_AERGOIO_AERGO_CONTRACT_NATIVE_GRAMMAR_TAB_H_INCLUDED
# define YY_YY_HOME_WRPARK_BLOCKO_SRC_GITHUB_COM_AERGOIO_AERGO_CONTRACT_NATIVE_GRAMMAR_TAB_H_INCLUDED
/* Debug traces.  */
#ifndef YYDEBUG
# define YYDEBUG 1
#endif
#if YYDEBUG
extern int yydebug;
#endif

/* Token type.  */
#ifndef YYTOKENTYPE
# define YYTOKENTYPE
  enum yytokentype
  {
    END = 0,
    ID = 258,
    L_FLOAT = 259,
    L_HEX = 260,
    L_INT = 261,
    L_OCTAL = 262,
    L_STR = 263,
    ASSIGN_ADD = 264,
    ASSIGN_SUB = 265,
    ASSIGN_MUL = 266,
    ASSIGN_DIV = 267,
    ASSIGN_MOD = 268,
    ASSIGN_AND = 269,
    ASSIGN_XOR = 270,
    ASSIGN_OR = 271,
    ASSIGN_RS = 272,
    ASSIGN_LS = 273,
    SHIFT_R = 274,
    SHIFT_L = 275,
    CMP_AND = 276,
    CMP_OR = 277,
    CMP_LE = 278,
    CMP_GE = 279,
    CMP_EQ = 280,
    CMP_NE = 281,
    UNARY_INC = 282,
    UNARY_DEC = 283,
    K_ACCOUNT = 284,
    K_ALTER = 285,
    K_BOOL = 286,
    K_BREAK = 287,
    K_BYTE = 288,
    K_CASE = 289,
    K_CONST = 290,
    K_CONTINUE = 291,
    K_CONTRACT = 292,
    K_CREATE = 293,
    K_DEFAULT = 294,
    K_DELETE = 295,
    K_DOUBLE = 296,
    K_DROP = 297,
    K_ELSE = 298,
    K_ENUM = 299,
    K_FALSE = 300,
    K_FLOAT = 301,
    K_FOR = 302,
    K_FUNC = 303,
    K_GOTO = 304,
    K_IF = 305,
    K_IMPLEMENTS = 306,
    K_IMPORT = 307,
    K_IN = 308,
    K_INDEX = 309,
    K_INSERT = 310,
    K_INT = 311,
    K_INT16 = 312,
    K_INT32 = 313,
    K_INT64 = 314,
    K_INT8 = 315,
    K_INTERFACE = 316,
    K_MAP = 317,
    K_NEW = 318,
    K_NULL = 319,
    K_PAYABLE = 320,
    K_PUBLIC = 321,
    K_REPLACE = 322,
    K_RETURN = 323,
    K_SELECT = 324,
    K_STRING = 325,
    K_STRUCT = 326,
    K_SWITCH = 327,
    K_TABLE = 328,
    K_TRUE = 329,
    K_TYPE = 330,
    K_UINT = 331,
    K_UINT16 = 332,
    K_UINT32 = 333,
    K_UINT64 = 334,
    K_UINT8 = 335,
    K_UPDATE = 336,
    K_VIEW = 337
  };
#endif

/* Value type.  */
#if ! defined YYSTYPE && ! defined YYSTYPE_IS_DECLARED

union YYSTYPE
{
#line 148 "grammar.y" /* yacc.c:355  */

    bool flag;
    char *str;
    vector_t *vect;

    type_t type;
    op_kind_t op;
    sql_kind_t sql;
    modifier_t mod;

    ast_id_t *id;
    ast_blk_t *blk;
    ast_exp_t *exp;
    ast_stmt_t *stmt;
    ast_imp_t *imp;
    meta_t *meta;

#line 236 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:355  */
};

typedef union YYSTYPE YYSTYPE;
# define YYSTYPE_IS_TRIVIAL 1
# define YYSTYPE_IS_DECLARED 1
#endif

/* Location type.  */
#if ! defined YYLTYPE && ! defined YYLTYPE_IS_DECLARED
typedef struct YYLTYPE YYLTYPE;
struct YYLTYPE
{
  int first_line;
  int first_column;
  int last_line;
  int last_column;
};
# define YYLTYPE_IS_DECLARED 1
# define YYLTYPE_IS_TRIVIAL 1
#endif



int yyparse (parse_t *parse, void *yyscanner);

#endif /* !YY_YY_HOME_WRPARK_BLOCKO_SRC_GITHUB_COM_AERGOIO_AERGO_CONTRACT_NATIVE_GRAMMAR_TAB_H_INCLUDED  */

/* Copy the second part of user declarations.  */

#line 266 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:358  */

#ifdef short
# undef short
#endif

#ifdef YYTYPE_UINT8
typedef YYTYPE_UINT8 yytype_uint8;
#else
typedef unsigned char yytype_uint8;
#endif

#ifdef YYTYPE_INT8
typedef YYTYPE_INT8 yytype_int8;
#else
typedef signed char yytype_int8;
#endif

#ifdef YYTYPE_UINT16
typedef YYTYPE_UINT16 yytype_uint16;
#else
typedef unsigned short int yytype_uint16;
#endif

#ifdef YYTYPE_INT16
typedef YYTYPE_INT16 yytype_int16;
#else
typedef short int yytype_int16;
#endif

#ifndef YYSIZE_T
# ifdef __SIZE_TYPE__
#  define YYSIZE_T __SIZE_TYPE__
# elif defined size_t
#  define YYSIZE_T size_t
# elif ! defined YYSIZE_T
#  include <stddef.h> /* INFRINGES ON USER NAME SPACE */
#  define YYSIZE_T size_t
# else
#  define YYSIZE_T unsigned int
# endif
#endif

#define YYSIZE_MAXIMUM ((YYSIZE_T) -1)

#ifndef YY_
# if defined YYENABLE_NLS && YYENABLE_NLS
#  if ENABLE_NLS
#   include <libintl.h> /* INFRINGES ON USER NAME SPACE */
#   define YY_(Msgid) dgettext ("bison-runtime", Msgid)
#  endif
# endif
# ifndef YY_
#  define YY_(Msgid) Msgid
# endif
#endif

#ifndef YY_ATTRIBUTE
# if (defined __GNUC__                                               \
      && (2 < __GNUC__ || (__GNUC__ == 2 && 96 <= __GNUC_MINOR__)))  \
     || defined __SUNPRO_C && 0x5110 <= __SUNPRO_C
#  define YY_ATTRIBUTE(Spec) __attribute__(Spec)
# else
#  define YY_ATTRIBUTE(Spec) /* empty */
# endif
#endif

#ifndef YY_ATTRIBUTE_PURE
# define YY_ATTRIBUTE_PURE   YY_ATTRIBUTE ((__pure__))
#endif

#ifndef YY_ATTRIBUTE_UNUSED
# define YY_ATTRIBUTE_UNUSED YY_ATTRIBUTE ((__unused__))
#endif

#if !defined _Noreturn \
     && (!defined __STDC_VERSION__ || __STDC_VERSION__ < 201112)
# if defined _MSC_VER && 1200 <= _MSC_VER
#  define _Noreturn __declspec (noreturn)
# else
#  define _Noreturn YY_ATTRIBUTE ((__noreturn__))
# endif
#endif

/* Suppress unused-variable warnings by "using" E.  */
#if ! defined lint || defined __GNUC__
# define YYUSE(E) ((void) (E))
#else
# define YYUSE(E) /* empty */
#endif

#if defined __GNUC__ && 407 <= __GNUC__ * 100 + __GNUC_MINOR__
/* Suppress an incorrect diagnostic about yylval being uninitialized.  */
# define YY_IGNORE_MAYBE_UNINITIALIZED_BEGIN \
    _Pragma ("GCC diagnostic push") \
    _Pragma ("GCC diagnostic ignored \"-Wuninitialized\"")\
    _Pragma ("GCC diagnostic ignored \"-Wmaybe-uninitialized\"")
# define YY_IGNORE_MAYBE_UNINITIALIZED_END \
    _Pragma ("GCC diagnostic pop")
#else
# define YY_INITIAL_VALUE(Value) Value
#endif
#ifndef YY_IGNORE_MAYBE_UNINITIALIZED_BEGIN
# define YY_IGNORE_MAYBE_UNINITIALIZED_BEGIN
# define YY_IGNORE_MAYBE_UNINITIALIZED_END
#endif
#ifndef YY_INITIAL_VALUE
# define YY_INITIAL_VALUE(Value) /* Nothing. */
#endif


#if ! defined yyoverflow || YYERROR_VERBOSE

/* The parser invokes alloca or malloc; define the necessary symbols.  */

# ifdef YYSTACK_USE_ALLOCA
#  if YYSTACK_USE_ALLOCA
#   ifdef __GNUC__
#    define YYSTACK_ALLOC __builtin_alloca
#   elif defined __BUILTIN_VA_ARG_INCR
#    include <alloca.h> /* INFRINGES ON USER NAME SPACE */
#   elif defined _AIX
#    define YYSTACK_ALLOC __alloca
#   elif defined _MSC_VER
#    include <malloc.h> /* INFRINGES ON USER NAME SPACE */
#    define alloca _alloca
#   else
#    define YYSTACK_ALLOC alloca
#    if ! defined _ALLOCA_H && ! defined EXIT_SUCCESS
#     include <stdlib.h> /* INFRINGES ON USER NAME SPACE */
      /* Use EXIT_SUCCESS as a witness for stdlib.h.  */
#     ifndef EXIT_SUCCESS
#      define EXIT_SUCCESS 0
#     endif
#    endif
#   endif
#  endif
# endif

# ifdef YYSTACK_ALLOC
   /* Pacify GCC's 'empty if-body' warning.  */
#  define YYSTACK_FREE(Ptr) do { /* empty */; } while (0)
#  ifndef YYSTACK_ALLOC_MAXIMUM
    /* The OS might guarantee only one guard page at the bottom of the stack,
       and a page size can be as small as 4096 bytes.  So we cannot safely
       invoke alloca (N) if N exceeds 4096.  Use a slightly smaller number
       to allow for a few compiler-allocated temporary stack slots.  */
#   define YYSTACK_ALLOC_MAXIMUM 4032 /* reasonable circa 2006 */
#  endif
# else
#  define YYSTACK_ALLOC YYMALLOC
#  define YYSTACK_FREE YYFREE
#  ifndef YYSTACK_ALLOC_MAXIMUM
#   define YYSTACK_ALLOC_MAXIMUM YYSIZE_MAXIMUM
#  endif
#  if (defined __cplusplus && ! defined EXIT_SUCCESS \
       && ! ((defined YYMALLOC || defined malloc) \
             && (defined YYFREE || defined free)))
#   include <stdlib.h> /* INFRINGES ON USER NAME SPACE */
#   ifndef EXIT_SUCCESS
#    define EXIT_SUCCESS 0
#   endif
#  endif
#  ifndef YYMALLOC
#   define YYMALLOC malloc
#   if ! defined malloc && ! defined EXIT_SUCCESS
void *malloc (YYSIZE_T); /* INFRINGES ON USER NAME SPACE */
#   endif
#  endif
#  ifndef YYFREE
#   define YYFREE free
#   if ! defined free && ! defined EXIT_SUCCESS
void free (void *); /* INFRINGES ON USER NAME SPACE */
#   endif
#  endif
# endif
#endif /* ! defined yyoverflow || YYERROR_VERBOSE */


#if (! defined yyoverflow \
     && (! defined __cplusplus \
         || (defined YYLTYPE_IS_TRIVIAL && YYLTYPE_IS_TRIVIAL \
             && defined YYSTYPE_IS_TRIVIAL && YYSTYPE_IS_TRIVIAL)))

/* A type that is properly aligned for any stack member.  */
union yyalloc
{
  yytype_int16 yyss_alloc;
  YYSTYPE yyvs_alloc;
  YYLTYPE yyls_alloc;
};

/* The size of the maximum gap between one aligned stack and the next.  */
# define YYSTACK_GAP_MAXIMUM (sizeof (union yyalloc) - 1)

/* The size of an array large to enough to hold all stacks, each with
   N elements.  */
# define YYSTACK_BYTES(N) \
     ((N) * (sizeof (yytype_int16) + sizeof (YYSTYPE) + sizeof (YYLTYPE)) \
      + 2 * YYSTACK_GAP_MAXIMUM)

# define YYCOPY_NEEDED 1

/* Relocate STACK from its old location to the new one.  The
   local variables YYSIZE and YYSTACKSIZE give the old and new number of
   elements in the stack, and YYPTR gives the new location of the
   stack.  Advance YYPTR to a properly aligned location for the next
   stack.  */
# define YYSTACK_RELOCATE(Stack_alloc, Stack)                           \
    do                                                                  \
      {                                                                 \
        YYSIZE_T yynewbytes;                                            \
        YYCOPY (&yyptr->Stack_alloc, Stack, yysize);                    \
        Stack = &yyptr->Stack_alloc;                                    \
        yynewbytes = yystacksize * sizeof (*Stack) + YYSTACK_GAP_MAXIMUM; \
        yyptr += yynewbytes / sizeof (*yyptr);                          \
      }                                                                 \
    while (0)

#endif

#if defined YYCOPY_NEEDED && YYCOPY_NEEDED
/* Copy COUNT objects from SRC to DST.  The source and destination do
   not overlap.  */
# ifndef YYCOPY
#  if defined __GNUC__ && 1 < __GNUC__
#   define YYCOPY(Dst, Src, Count) \
      __builtin_memcpy (Dst, Src, (Count) * sizeof (*(Src)))
#  else
#   define YYCOPY(Dst, Src, Count)              \
      do                                        \
        {                                       \
          YYSIZE_T yyi;                         \
          for (yyi = 0; yyi < (Count); yyi++)   \
            (Dst)[yyi] = (Src)[yyi];            \
        }                                       \
      while (0)
#  endif
# endif
#endif /* !YYCOPY_NEEDED */

/* YYFINAL -- State number of the termination state.  */
#define YYFINAL  19
/* YYLAST -- Last index in YYTABLE.  */
#define YYLAST   1831

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  106
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  89
/* YYNRULES -- Number of rules.  */
#define YYNRULES  257
/* YYNSTATES -- Number of states.  */
#define YYNSTATES  426

/* YYTRANSLATE[YYX] -- Symbol number corresponding to YYX as returned
   by yylex, with out-of-bounds checking.  */
#define YYUNDEFTOK  2
#define YYMAXUTOK   337

#define YYTRANSLATE(YYX)                                                \
  ((unsigned int) (YYX) <= YYMAXUTOK ? yytranslate[YYX] : YYUNDEFTOK)

/* YYTRANSLATE[TOKEN-NUM] -- Symbol number corresponding to TOKEN-NUM
   as returned by yylex, without out-of-bounds checking.  */
static const yytype_uint8 yytranslate[] =
{
       0,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    83,     2,     2,     2,    93,    86,     2,
      94,    95,    91,    89,   101,    90,    96,    92,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,   104,    99,
      87,   100,    88,   105,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,   102,     2,   103,    85,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    97,    84,    98,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     1,     2,     3,     4,
       5,     6,     7,     8,     9,    10,    11,    12,    13,    14,
      15,    16,    17,    18,    19,    20,    21,    22,    23,    24,
      25,    26,    27,    28,    29,    30,    31,    32,    33,    34,
      35,    36,    37,    38,    39,    40,    41,    42,    43,    44,
      45,    46,    47,    48,    49,    50,    51,    52,    53,    54,
      55,    56,    57,    58,    59,    60,    61,    62,    63,    64,
      65,    66,    67,    68,    69,    70,    71,    72,    73,    74,
      75,    76,    77,    78,    79,    80,    81,    82
};

#if YYDEBUG
  /* YYRLINE[YYN] -- Source line where rule number YYN was defined.  */
static const yytype_uint16 yyrline[] =
{
       0,   257,   257,   262,   267,   272,   276,   280,   287,   294,
     303,   322,   323,   330,   335,   340,   345,   350,   355,   363,
     364,   369,   377,   378,   386,   387,   399,   411,   415,   421,
     431,   432,   433,   436,   437,   438,   439,   440,   441,   442,
     443,   444,   445,   446,   450,   451,   466,   470,   482,   483,
     487,   491,   505,   506,   510,   514,   522,   531,   543,   547,
     555,   560,   568,   572,   579,   581,   585,   601,   602,   606,
     613,   614,   618,   623,   631,   640,   641,   645,   650,   657,
     662,   667,   672,   680,   687,   688,   689,   690,   694,   695,
     696,   700,   701,   730,   736,   748,   755,   760,   768,   769,
     770,   771,   772,   773,   774,   775,   776,   777,   781,   788,
     792,   800,   804,   811,   812,   813,   814,   815,   816,   817,
     818,   819,   820,   824,   829,   833,   840,   844,   849,   854,
     862,   866,   870,   874,   878,   882,   886,   890,   894,   902,
     903,   907,   908,   912,   916,   920,   928,   929,   933,   938,
     946,   950,   954,   958,   962,   969,   988,   989,   990,   991,
     992,   993,   994,   998,  1005,  1006,  1020,  1021,  1040,  1041,
    1042,  1043,  1044,  1048,  1049,  1053,  1061,  1065,  1077,  1084,
    1089,  1097,  1098,  1099,  1106,  1107,  1114,  1115,  1122,  1123,
    1130,  1131,  1138,  1139,  1146,  1147,  1154,  1155,  1162,  1163,
    1167,  1168,  1175,  1176,  1177,  1178,  1182,  1183,  1190,  1191,
    1195,  1196,  1203,  1204,  1208,  1209,  1216,  1217,  1218,  1222,
    1223,  1230,  1231,  1235,  1242,  1243,  1244,  1245,  1249,  1250,
    1254,  1258,  1262,  1266,  1273,  1274,  1278,  1283,  1291,  1292,
    1296,  1300,  1307,  1311,  1315,  1319,  1326,  1333,  1340,  1347,
    1354,  1355,  1356,  1357,  1358,  1359,  1363,  1370
};
#endif

#if YYDEBUG || YYERROR_VERBOSE || 1
/* YYTNAME[SYMBOL-NUM] -- String name of the symbol SYMBOL-NUM.
   First, the terminals, then, starting at YYNTOKENS, nonterminals.  */
static const char *const yytname[] =
{
  "\"EOF\"", "error", "$undefined", "\"identifier\"",
  "\"floating-point\"", "\"hexadecimal\"", "\"integer\"", "\"octal\"",
  "\"characters\"", "\"+=\"", "\"-=\"", "\"*=\"", "\"/=\"", "\"%=\"",
  "\"&=\"", "\"^=\"", "\"|=\"", "\">>=\"", "\"<<=\"", "\">>\"", "\"<<\"",
  "\"&&\"", "\"||\"", "\"<=\"", "\">=\"", "\"==\"", "\"!=\"", "\"++\"",
  "\"--\"", "\"account\"", "\"alter\"", "\"bool\"", "\"break\"",
  "\"byte\"", "\"case\"", "\"const\"", "\"continue\"", "\"contract\"",
  "\"create\"", "\"default\"", "\"delete\"", "\"double\"", "\"drop\"",
  "\"else\"", "\"enum\"", "\"false\"", "\"float\"", "\"for\"", "\"func\"",
  "\"goto\"", "\"if\"", "\"implements\"", "\"import\"", "\"in\"",
  "\"index\"", "\"insert\"", "\"int\"", "\"int16\"", "\"int32\"",
  "\"int64\"", "\"int8\"", "\"interface\"", "\"map\"", "\"new\"",
  "\"null\"", "\"payable\"", "\"public\"", "\"replace\"", "\"return\"",
  "\"select\"", "\"string\"", "\"struct\"", "\"switch\"", "\"table\"",
  "\"true\"", "\"type\"", "\"uint\"", "\"uint16\"", "\"uint32\"",
  "\"uint64\"", "\"uint8\"", "\"update\"", "\"view\"", "'!'", "'|'", "'^'",
  "'&'", "'<'", "'>'", "'+'", "'-'", "'*'", "'/'", "'%'", "'('", "')'",
  "'.'", "'{'", "'}'", "';'", "'='", "','", "'['", "']'", "':'", "'?'",
  "$accept", "root", "import", "contract", "impl_opt", "contract_body",
  "variable", "var_qual", "var_decl", "var_spec", "var_type", "prim_type",
  "declarator_list", "declarator", "size_opt", "var_init", "compound",
  "struct", "field_list", "enumeration", "enum_list", "enumerator",
  "comma_opt", "function", "func_spec", "ctor_spec", "param_list_opt",
  "param_list", "param_decl", "block", "blk_decl", "udf_spec",
  "modifier_opt", "return_opt", "return_list", "return_decl", "interface",
  "interface_body", "statement", "empty_stmt", "exp_stmt", "assign_stmt",
  "assign_op", "label_stmt", "if_stmt", "loop_stmt", "init_stmt",
  "cond_exp", "switch_stmt", "switch_blk", "case_blk", "jump_stmt",
  "ddl_stmt", "ddl_prefix", "blk_stmt", "expression", "sql_exp",
  "sql_prefix", "new_exp", "alloc_exp", "initializer", "elem_list",
  "init_elem", "ternary_exp", "or_exp", "and_exp", "bit_or_exp",
  "bit_xor_exp", "bit_and_exp", "eq_exp", "eq_op", "cmp_exp", "cmp_op",
  "add_exp", "add_op", "shift_exp", "shift_op", "mul_exp", "mul_op",
  "cast_exp", "unary_exp", "unary_op", "post_exp", "arg_list_opt",
  "arg_list", "prim_exp", "literal", "non_reserved_token", "identifier", YY_NULLPTR
};
#endif

# ifdef YYPRINT
/* YYTOKNUM[NUM] -- (External) token number corresponding to the
   (internal) symbol number NUM (which must be that of a token).  */
static const yytype_uint16 yytoknum[] =
{
       0,   256,   257,   258,   259,   260,   261,   262,   263,   264,
     265,   266,   267,   268,   269,   270,   271,   272,   273,   274,
     275,   276,   277,   278,   279,   280,   281,   282,   283,   284,
     285,   286,   287,   288,   289,   290,   291,   292,   293,   294,
     295,   296,   297,   298,   299,   300,   301,   302,   303,   304,
     305,   306,   307,   308,   309,   310,   311,   312,   313,   314,
     315,   316,   317,   318,   319,   320,   321,   322,   323,   324,
     325,   326,   327,   328,   329,   330,   331,   332,   333,   334,
     335,   336,   337,    33,   124,    94,    38,    60,    62,    43,
      45,    42,    47,    37,    40,    41,    46,   123,   125,    59,
      61,    44,    91,    93,    58,    63
};
# endif

#define YYPACT_NINF -220

#define yypact_value_is_default(Yystate) \
  (!!((Yystate) == (-220)))

#define YYTABLE_NINF -29

#define yytable_value_is_error(Yytable_value) \
  0

  /* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
     STATE-NUM.  */
static const yytype_int16 yypact[] =
{
      69,   147,    11,   147,    71,  -220,  -220,  -220,  -220,  -220,
    -220,  -220,  -220,  -220,  -220,  -220,   -23,  -220,   -51,  -220,
    -220,  -220,  -220,   147,   -39,   117,  -220,   919,  -220,    16,
     -13,    41,   -35,  -220,  -220,  -220,  1749,   714,  -220,  -220,
    -220,  -220,  -220,     9,  1714,  -220,   717,  -220,  -220,  -220,
    -220,  -220,  -220,   971,  -220,  -220,  -220,    88,   769,  -220,
    -220,  -220,  -220,  -220,    13,  -220,  -220,    55,  -220,  -220,
     147,  -220,    25,  -220,   147,  -220,   107,    57,  1749,  -220,
     111,    96,  -220,  -220,  -220,  -220,  -220,  1419,    98,   113,
     128,  -220,   361,  -220,  1749,   137,  -220,  -220,   147,   132,
    -220,   139,  -220,  -220,  -220,  -220,  -220,  -220,  -220,  -220,
    -220,  -220,  1058,  -220,  -220,  -220,  -220,  -220,  -220,  1518,
    -220,  1327,    73,  -220,   236,  -220,  -220,    -7,   223,   161,
     163,   165,   191,    17,   129,   204,   102,  -220,  -220,  1518,
      67,  -220,  -220,  -220,  -220,   147,  1561,   150,   179,   154,
    1561,   159,   -20,   162,    63,    15,   147,    20,   235,    19,
    -220,  -220,  -220,  -220,  -220,   460,  -220,  -220,  -220,  -220,
    -220,   222,  -220,  -220,  -220,  -220,   267,  -220,    59,  -220,
     461,    14,   147,   175,   168,  -220,  1749,   172,  -220,   176,
    1749,  1749,  1119,  -220,   177,  -220,   180,   147,  1419,  -220,
     182,   -57,  -220,  1419,   183,  1561,  1419,  1561,  1561,  1561,
    1561,  -220,  -220,  1561,  -220,  -220,  -220,  -220,  1561,  -220,
    -220,  1561,  -220,  -220,  1561,  -220,  -220,  -220,  1561,  -220,
    -220,  -220,  1625,   147,  1561,   128,   185,   129,  -220,  -220,
    -220,    -2,  -220,  -220,  -220,  -220,  -220,  -220,  -220,  -220,
     187,   755,  -220,   184,   194,  1419,  -220,    79,   195,  1419,
     173,  -220,  -220,  -220,  -220,  -220,   -25,   196,  -220,  1419,
    1419,  -220,  -220,  -220,  -220,  -220,  -220,  -220,  -220,  -220,
    -220,  1419,   658,   128,  -220,  1749,   199,   147,   202,  1561,
     206,   198,  1023,  -220,   209,  -220,  -220,   203,  1561,  1625,
     180,  1518,  -220,  -220,  -220,   223,   -54,   161,   163,   165,
     191,    17,   129,   204,   102,  -220,  -220,   208,   210,  -220,
     211,  -220,  -220,  -220,   852,   -16,  -220,  -220,   852,    26,
     220,   174,  -220,  -220,   -40,  -220,  -220,    36,  -220,  -220,
     559,   215,   226,  -220,  -220,    99,  -220,   105,  -220,   215,
    -220,  1458,  -220,  -220,   256,  -220,  -220,  -220,   224,  1119,
     228,  1119,    30,   227,  -220,  1561,  -220,  1625,  -220,  -220,
    1183,   229,  1668,  1255,  1668,    13,    13,   233,  -220,  -220,
    1419,  -220,  -220,  1749,  -220,  -220,   230,   219,  -220,  -220,
    -220,  -220,  -220,  -220,  -220,  -220,    13,    39,  -220,     5,
      13,    46,   278,  -220,  -220,  -220,    51,    56,  1749,  1561,
    -220,    13,    13,  -220,    13,    13,    13,  -220,   219,   232,
    -220,  -220,  -220,  -220,  -220,  -220
};

  /* YYDEFACT[STATE-NUM] -- Default reduction number in state STATE-NUM.
     Performed when YYTABLE does not specify something else to do.  Zero
     means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       0,     0,     0,     0,     0,     2,     3,     4,   256,   250,
     251,   252,   253,   254,   255,   257,    11,     8,     0,     1,
       5,     6,     7,     0,     0,    84,    12,    84,    86,    85,
       0,     0,    84,    30,    31,    32,     0,     0,    33,    34,
      35,    36,    37,     0,    85,    38,     0,    39,    40,    41,
      42,    43,     9,    84,    13,    19,    22,     0,     0,    27,
      14,    52,    53,    15,     0,    67,    68,    28,    87,    96,
       0,    95,     0,    23,     0,    28,     0,     0,     0,    20,
       0,     0,    10,    16,    17,    18,    24,     0,     0,    26,
      44,    46,     0,    66,    70,     0,    97,    59,     0,     0,
      55,     0,   248,   247,   245,   246,   249,   224,   225,   168,
     244,   169,     0,   242,   170,   171,   243,   172,   227,     0,
     226,     0,     0,    50,     0,   166,   173,   184,   186,   188,
     190,   192,   194,   196,   200,   206,   210,   214,   219,     0,
     221,   228,   238,   239,    21,     0,    48,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
      75,   108,    77,    78,   163,     0,    79,    98,    99,   100,
     101,   102,   103,   104,   105,   106,     0,   107,     0,   164,
     219,   239,     0,     0,    71,    72,    70,    64,    60,    62,
       0,     0,     0,   176,   174,   175,    28,     0,     0,   223,
       0,     0,    25,     0,     0,     0,     0,     0,     0,     0,
       0,   198,   199,     0,   204,   205,   202,   203,     0,   208,
     209,     0,   212,   213,     0,   216,   217,   218,     0,   222,
     232,   233,   234,     0,     0,    45,     0,    49,   110,   157,
     151,     0,   150,   159,   156,   161,   125,   160,   158,   162,
       0,     0,   130,     0,     0,     0,   152,     0,     0,     0,
       0,   143,    76,    80,    81,    82,     0,     0,   109,     0,
       0,   113,   114,   115,   116,   117,   118,   119,   120,   121,
     122,     0,     0,    74,    69,     0,     0,    65,     0,     0,
       0,     0,     0,   182,    64,   179,   181,   239,     0,   234,
       0,     0,   240,    51,   167,   187,     0,   189,   191,   193,
     195,   197,   201,   207,   211,   215,   236,     0,   235,   231,
       0,    47,   124,   138,     0,     0,   139,   140,     0,     0,
     173,   239,   154,   129,     0,   153,   145,     0,   146,   148,
       0,     0,     0,   128,   155,     0,   165,     0,   123,   239,
      73,    88,    61,    58,    63,    29,    56,    54,     0,    65,
       0,     0,     0,     0,   220,     0,   230,     0,   229,   141,
       0,     0,     0,     0,     0,     0,     0,     0,   147,   149,
       0,   111,   112,     0,    93,    83,    90,    91,    57,   180,
     178,   183,   177,   241,   185,   237,     0,     0,   142,     0,
       0,     0,     0,   131,   126,   144,     0,     0,     0,    48,
     134,     0,     0,   132,     0,     0,     0,    89,    92,     0,
     135,   137,   133,   136,   127,    94
};

  /* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -220,  -220,   328,   332,  -220,  -220,   285,   -36,   -29,  -177,
     -24,   221,  -220,  -118,   -69,  -220,   -43,  -220,  -220,  -220,
    -220,    54,    49,   291,  -220,  -220,   160,  -220,    60,   -63,
    -220,    35,  -220,  -220,   -28,   -61,   350,  -220,  -154,   106,
    -220,   110,  -220,   103,  -220,  -220,  -220,    28,  -220,    -6,
    -220,  -220,  -220,  -220,  -220,  -116,   -78,  -220,  -219,  -220,
     258,  -220,  -191,  -186,    86,   171,   170,   178,   169,  -132,
    -220,   166,  -220,  -142,  -220,   164,  -220,   157,  -220,   155,
     -80,  -220,  -159,   115,  -220,  -220,  -220,  -220,    -1
};

  /* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     4,     5,     6,    24,    53,    54,    55,    56,    57,
      74,    59,    89,    90,   236,   122,    60,    61,   292,    62,
     187,   188,   288,    63,    64,    65,   183,   184,   185,   164,
     165,    66,    31,   385,   386,   387,     7,    32,   166,   167,
     168,   169,   281,   170,   171,   172,   328,   370,   173,   261,
     340,   174,   175,   176,   177,   178,   179,   124,   125,   194,
     293,   294,   295,   126,   127,   128,   129,   130,   131,   132,
     213,   133,   218,   134,   221,   135,   224,   136,   228,   137,
     138,   139,   140,   317,   318,   141,   142,    15,   143
};

  /* YYTABLE[YYPACT[STATE-NUM]] -- What to do in state STATE-NUM.  If
     positive, shift that token.  If negative, reduce the rule whose
     number is the opposite.  If YYTABLE_NINF, syntax error.  */
static const yytype_int16 yytable[] =
{
      16,    93,    18,    58,   237,   201,   296,    73,    79,   123,
      84,   265,   180,   316,   291,   205,   250,   -28,   241,    17,
     258,   254,    26,   211,   212,   342,    67,   235,    23,    58,
      28,    29,   230,   231,   243,    75,    77,   372,   302,   199,
     214,   215,   257,    75,   270,    81,    25,   270,   320,   163,
     365,   -28,    67,   244,    99,   376,   162,    91,    27,   229,
      30,   270,   245,    71,   283,   330,   -28,    72,   -28,    95,
     182,    19,    92,    91,   325,   -28,   312,    75,   310,   374,
     316,    68,   201,    86,    87,   180,    69,   -28,   193,    70,
     306,   181,   252,    75,   230,   231,   -28,   189,   206,   232,
     412,   233,   322,    78,   216,   217,     1,   234,     1,   251,
      92,   196,    92,   259,   255,   358,   260,   247,   282,   219,
     220,     2,   264,     2,    96,   303,   269,   270,   348,   263,
       3,   377,     3,   392,   411,   329,   248,   270,   371,   334,
     270,   414,   371,   337,    91,   249,   416,   270,   395,    94,
       8,   417,   270,   345,    98,   253,   362,   408,   268,   269,
     270,   232,   182,   233,   181,   347,   290,   101,   389,   234,
     391,   180,   202,   296,   203,   296,     8,   -28,   335,   394,
     270,    91,    28,    29,     9,    75,   379,    86,    87,    75,
      75,   297,   346,   225,   226,   227,   300,   144,   381,    10,
     270,    11,   180,   343,   382,    97,   270,   150,    12,   100,
       9,   -28,   153,   399,   145,   402,   211,   212,   219,   220,
      13,   364,   324,   222,   223,    10,   -28,    11,   -28,    14,
     146,   186,   319,   190,    12,   -28,   191,   204,     8,   102,
     103,   104,   105,   106,   207,   208,    13,   -28,   209,   238,
     331,   210,   239,   240,   397,    14,   -28,   401,   242,   341,
     180,   182,   107,   108,   406,   266,   246,   237,   267,   285,
     284,   338,     9,   287,   299,   109,   289,   301,   205,   298,
     110,   349,   304,   332,    75,   323,   189,    10,   321,    11,
     111,    75,   333,   336,   351,   344,    12,   356,   112,   113,
     353,   355,   114,   366,   115,   230,   231,   361,    13,   116,
     359,   367,   403,   404,   368,   375,   117,    14,   118,   282,
     380,   409,   393,   388,   119,   120,   390,   384,   398,   121,
     260,   408,    20,   410,   256,   425,    21,   413,    83,   349,
     419,   352,   200,   360,    85,   350,   286,   418,   420,   421,
      75,   422,   423,   424,    22,   407,   373,   326,   297,   384,
     297,   327,   147,   339,     8,   102,   103,   104,   105,   106,
     195,   405,   232,   415,   233,   354,   305,   307,   309,   311,
     234,   314,    75,   315,   384,   313,   308,     0,   107,   108,
      33,   148,    34,   149,    35,   150,    36,   151,     9,   152,
     153,   109,     0,   154,     0,    37,   110,    75,   155,     0,
     156,   157,     0,    10,   363,    11,   111,    38,    39,    40,
      41,    42,    12,    43,   112,   113,     0,     0,   114,   158,
     115,    45,     0,   159,    13,   116,    46,    47,    48,    49,
      50,    51,   117,    14,   118,     0,     0,     0,     0,     0,
     119,   120,     0,     0,     0,   121,     0,     0,    92,   160,
     161,   147,     0,     8,   102,   103,   104,   105,   106,     0,
     271,   272,   273,   274,   275,   276,   277,   278,   279,   280,
       0,     0,     0,     0,     0,     0,     0,   107,   108,    33,
     148,    34,   149,    35,   150,    36,   151,     9,   152,   153,
     109,     0,   154,     0,    37,   110,     0,   155,     0,   156,
     157,     0,    10,     0,    11,   111,    38,    39,    40,    41,
      42,    12,    43,   112,   113,     0,     0,   114,   158,   115,
      45,     0,   159,    13,   116,    46,    47,    48,    49,    50,
      51,   117,    14,   118,     0,     0,     0,     0,     0,   119,
     120,     0,     0,     0,   121,     0,     0,    92,   262,   161,
     147,     0,     8,   102,   103,   104,   105,   106,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,   107,   108,     0,   148,
       0,   149,     0,   150,     0,   151,     9,   152,   153,   109,
       0,   154,     0,     0,   110,     0,   155,     0,   156,   157,
       0,    10,     0,    11,   111,     0,     0,     0,     0,     0,
      12,     0,   112,   113,     0,     0,   114,   158,   115,     0,
       0,   159,    13,   116,     0,     0,     0,     0,     0,     0,
     117,    14,   118,     0,     0,     0,     0,     0,   119,   120,
       0,     0,     0,   121,     0,     0,    92,   378,   161,   147,
       0,     8,   102,   103,   104,   105,   106,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   107,   108,     0,   148,     0,
     149,     0,   150,     0,   151,     9,   152,   153,   109,     0,
     154,     0,     0,   110,     0,   155,     0,   156,   157,     0,
      10,     0,    11,   111,     0,    76,     0,     8,    80,    12,
       8,   112,   113,     0,     0,   114,   158,   115,     0,     0,
     159,    13,   116,     0,     0,     0,     0,     0,     0,   117,
      14,   118,     0,     0,     0,     0,     0,   119,   120,     0,
       0,     9,   121,     0,     9,    92,     0,   161,     8,   102,
     103,   104,   105,   106,     0,     0,    10,     0,    11,    10,
      88,    11,     8,     0,     0,    12,     0,     0,    12,     0,
       0,     0,   107,   108,    33,     0,    34,    13,    35,     0,
      13,     0,     9,     0,     0,   109,    14,     0,     0,    14,
     110,     0,     0,     0,     0,     0,     9,    10,     0,    11,
     111,    38,    39,    40,    41,    42,    12,    43,   112,   113,
       0,    10,   114,    11,   115,    45,     0,     0,    13,   116,
      12,    47,    48,    49,    50,    51,   117,    14,   118,     0,
       0,     0,    13,     0,   119,   120,     0,     0,     0,   121,
       0,    14,     0,     0,   161,     8,   102,   103,   104,   105,
     106,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,   107,
     108,     0,     0,     0,     0,     0,     0,     0,     0,     9,
       0,     0,     0,     0,     0,     0,     0,   110,     0,     0,
       0,     0,     0,     0,    10,     0,    11,     0,     0,     0,
       0,     0,     0,    12,     0,   197,   113,     0,     0,     0,
       0,     0,     8,     0,     0,    13,   116,     0,     0,     0,
       0,     0,     0,     0,    14,   118,     0,     0,     0,     0,
       0,   119,   120,     0,     0,     0,   121,     0,    33,     0,
      34,   369,    35,     0,    36,     0,     9,     0,     0,     0,
       0,     0,     0,    37,     0,     0,     0,     0,     0,     0,
       0,    10,     0,    11,     8,    38,    39,    40,    41,    42,
      12,    43,     0,     0,    28,    44,     0,     0,     0,    45,
       0,     0,    13,     0,    46,    47,    48,    49,    50,    51,
      33,    14,    34,     0,    35,     0,    36,     0,     9,     0,
       0,     0,     0,     0,     0,    37,     0,    52,     0,     0,
       0,     0,     0,    10,     0,    11,     8,    38,    39,    40,
      41,    42,    12,    43,     0,     0,    28,    44,     0,     0,
       0,    45,     0,     0,    13,     0,    46,    47,    48,    49,
      50,    51,    33,    14,    34,     0,    35,     0,     0,     0,
       9,     8,     0,     0,     0,     0,     0,     0,     0,    82,
       0,     0,     0,     0,     0,    10,     0,    11,     0,    38,
      39,    40,    41,    42,    12,    43,     0,    33,     0,    34,
       0,    35,     0,    45,     0,     9,    13,     0,     0,    47,
      48,    49,    50,    51,     0,    14,     0,     0,     0,     0,
      10,     0,    11,     0,    38,    39,    40,    41,    42,    12,
      43,   357,     8,   102,   103,   104,   105,   106,    45,     0,
       0,    13,     0,     0,    47,    48,    49,    50,    51,     0,
      14,     0,     0,     0,     0,     0,   107,   108,     0,     0,
       0,     0,     0,     0,     0,   192,     9,     0,     0,     0,
       0,     0,     0,     0,   110,     0,     0,     0,     0,     0,
       0,    10,     0,    11,     0,     0,     0,     0,     0,     0,
      12,     0,   197,   113,     0,     0,     8,   102,   103,   104,
     105,   106,    13,   116,     0,     0,     0,     0,     0,     0,
       0,    14,   118,     0,     0,     0,     0,     0,   119,   120,
     107,   108,     0,   121,     0,     0,   192,     0,     0,     0,
       9,     0,     0,   109,     0,     0,     0,     0,   110,     0,
       0,     0,     0,     0,     0,    10,     0,    11,   111,     0,
       0,     0,     0,     0,    12,     0,   112,   113,     0,     0,
     114,     0,   115,     0,     0,     0,    13,   116,     8,   102,
     103,   104,   105,   106,   117,    14,   118,     0,     0,     0,
       0,     0,   119,   120,     0,     0,     0,   121,   396,     0,
       0,     0,   107,   108,     0,     0,     0,     0,     0,     0,
       0,     0,     9,     0,     0,   109,     0,     0,     0,     0,
     110,     0,     0,     0,     0,     0,     0,    10,     0,    11,
     111,     0,     0,     0,     0,     0,    12,     0,   112,   113,
       0,     0,   114,     0,   115,     0,     0,     0,    13,   116,
       8,   102,   103,   104,   105,   106,   117,    14,   118,     0,
       0,     0,     0,     0,   119,   120,     0,     0,     0,   121,
     400,     0,     0,     0,   107,   108,    33,     0,    34,     0,
      35,     0,     0,     0,     9,     0,     0,   109,     0,     0,
       0,     0,   110,     0,     0,     0,     0,     0,     0,    10,
       0,    11,   111,    38,    39,    40,    41,    42,    12,     0,
     112,   113,     0,     0,   114,     0,   115,    45,     0,     0,
      13,   116,     0,    47,    48,    49,    50,    51,   117,    14,
     118,     0,     0,     0,     0,     0,   119,   120,     0,     0,
       0,   121,     8,   102,   103,   104,   105,   106,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,   107,   108,     0,     0,
       0,     0,     0,     0,     0,     0,     9,     0,     0,   109,
       0,     8,     0,     0,   110,     0,     0,     0,     0,     0,
       0,    10,     0,    11,   111,     0,     0,     0,     0,     0,
      12,     0,   112,   113,     0,     0,   114,    33,   115,    34,
       0,    35,    13,   116,     0,     9,     0,     0,     0,     0,
     117,    14,   118,     0,     0,     0,     0,     0,   119,   120,
      10,     0,    11,   121,    38,    39,    40,    41,    42,    12,
      43,     8,   102,   103,   104,   105,   106,     0,    45,     0,
       0,    13,     0,     0,    47,    48,    49,    50,    51,     0,
      14,     0,     0,     0,     0,   107,   108,     0,     0,     0,
       0,     0,   383,     0,     0,     9,     0,     0,     0,     0,
       0,     0,     0,   110,     8,   102,   103,   104,   105,   106,
      10,     0,    11,     0,     0,     0,     0,     0,     0,    12,
       0,   197,   113,     0,     0,     0,     0,     0,   107,   108,
       0,    13,   116,     0,     0,     0,     0,     0,     9,     0,
      14,   118,     0,     0,     0,     0,   110,   119,   120,     0,
       0,     0,   198,    10,     0,    11,     0,     0,     0,     0,
       0,     0,    12,     0,   197,   113,     0,     0,     8,   102,
     103,   104,   105,   106,    13,   116,     0,     0,     0,     0,
       0,     0,     0,    14,   118,     0,     0,     0,     0,     0,
     119,   120,   107,   108,     0,   121,     0,     0,     0,     0,
       0,     0,     9,     0,     0,     0,     0,     0,     0,     0,
     110,     8,   102,   103,   104,   105,   106,    10,     0,    11,
       0,     0,     0,     0,     0,     0,    12,     0,   112,   113,
       0,     0,     0,     0,     0,     0,     0,     0,    13,   116,
       0,     0,     0,     0,     0,     9,     0,    14,   118,     0,
       0,     0,     0,   110,   119,   120,     0,     8,     0,   121,
      10,     0,    11,     0,     0,     0,     0,     0,     0,    12,
       0,   197,   113,     0,     0,     0,     0,     0,     0,     0,
       0,    13,   116,    33,     0,    34,     0,    35,     0,    36,
      14,     9,     8,     0,     0,     0,     0,     0,     0,     0,
       0,     0,   198,     0,     0,     0,    10,     0,    11,     0,
      38,    39,    40,    41,    42,    12,    43,     0,    33,    68,
      34,     0,    35,     0,    45,     0,     9,    13,     0,     0,
      47,    48,    49,    50,    51,     0,    14,     0,     0,     0,
       0,    10,     0,    11,     0,    38,    39,    40,    41,    42,
      12,    43,     0,     0,     0,     0,     0,     0,     0,    45,
       0,     0,    13,     0,     0,    47,    48,    49,    50,    51,
       0,    14
};

static const yytype_int16 yycheck[] =
{
       1,    64,     3,    27,   146,   121,   192,    36,    44,    87,
      53,   165,    92,   232,   191,    22,     1,     3,   150,     8,
       1,     1,    23,    25,    26,    50,    27,   145,    51,    53,
      65,    66,    27,    28,    54,    36,    37,    53,    95,   119,
      23,    24,   158,    44,   101,    46,    97,   101,   234,    92,
     104,    37,    53,    73,    78,    95,    92,    58,    97,   139,
      25,   101,    82,    98,   182,   251,    52,    32,    54,    70,
      94,     0,    97,    74,   251,    61,   218,    78,   210,    53,
     299,    65,   198,    99,   100,   165,    99,    73,   112,    48,
     206,    92,   155,    94,    27,    28,    82,    98,   105,    94,
      95,    96,   104,    94,    87,    88,    37,   102,    37,    94,
      97,   112,    97,    94,    94,   292,    97,    54,   104,    89,
      90,    52,   165,    52,    99,   203,   100,   101,   282,   165,
      61,    95,    61,   103,    95,   251,    73,   101,   324,   255,
     101,    95,   328,   259,   145,    82,    95,   101,   367,    94,
       3,    95,   101,   269,    97,   156,   298,   101,    99,   100,
     101,    94,   186,    96,   165,   281,   190,    71,   359,   102,
     361,   251,    99,   359,   101,   361,     3,     3,    99,   365,
     101,   182,    65,    66,    37,   186,   340,    99,   100,   190,
     191,   192,   270,    91,    92,    93,   197,    99,    99,    52,
     101,    54,   282,   266,    99,    98,   101,    34,    61,    98,
      37,    37,    39,   372,   101,   374,    25,    26,    89,    90,
      73,   301,   251,    19,    20,    52,    52,    54,    54,    82,
     102,    94,   233,   101,    61,    61,    97,     1,     3,     4,
       5,     6,     7,     8,    21,    84,    73,    73,    85,    99,
     251,    86,    73,    99,   370,    82,    82,   373,    99,   260,
     340,   285,    27,    28,   380,    43,   104,   409,     1,   101,
      95,    98,    37,   101,    94,    40,   100,    95,    22,   102,
      45,   282,    99,    99,   285,    98,   287,    52,   103,    54,
      55,   292,    98,    98,    95,    99,    61,    99,    63,    64,
      98,    95,    67,    95,    69,    27,    28,   104,    73,    74,
     101,   101,   375,   376,   103,    95,    81,    82,    83,   104,
      94,   102,    95,    99,    89,    90,    98,   351,    99,    94,
      97,   101,     4,   396,    99,   103,     4,   400,    53,   340,
     409,   287,   121,   294,    53,   285,   186,   408,   411,   412,
     351,   414,   415,   416,     4,   383,   328,   251,   359,   383,
     361,   251,     1,   260,     3,     4,     5,     6,     7,     8,
     112,   377,    94,    95,    96,   289,   205,   207,   209,   213,
     102,   224,   383,   228,   408,   221,   208,    -1,    27,    28,
      29,    30,    31,    32,    33,    34,    35,    36,    37,    38,
      39,    40,    -1,    42,    -1,    44,    45,   408,    47,    -1,
      49,    50,    -1,    52,   299,    54,    55,    56,    57,    58,
      59,    60,    61,    62,    63,    64,    -1,    -1,    67,    68,
      69,    70,    -1,    72,    73,    74,    75,    76,    77,    78,
      79,    80,    81,    82,    83,    -1,    -1,    -1,    -1,    -1,
      89,    90,    -1,    -1,    -1,    94,    -1,    -1,    97,    98,
      99,     1,    -1,     3,     4,     5,     6,     7,     8,    -1,
       9,    10,    11,    12,    13,    14,    15,    16,    17,    18,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,    28,    29,
      30,    31,    32,    33,    34,    35,    36,    37,    38,    39,
      40,    -1,    42,    -1,    44,    45,    -1,    47,    -1,    49,
      50,    -1,    52,    -1,    54,    55,    56,    57,    58,    59,
      60,    61,    62,    63,    64,    -1,    -1,    67,    68,    69,
      70,    -1,    72,    73,    74,    75,    76,    77,    78,    79,
      80,    81,    82,    83,    -1,    -1,    -1,    -1,    -1,    89,
      90,    -1,    -1,    -1,    94,    -1,    -1,    97,    98,    99,
       1,    -1,     3,     4,     5,     6,     7,     8,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    30,
      -1,    32,    -1,    34,    -1,    36,    37,    38,    39,    40,
      -1,    42,    -1,    -1,    45,    -1,    47,    -1,    49,    50,
      -1,    52,    -1,    54,    55,    -1,    -1,    -1,    -1,    -1,
      61,    -1,    63,    64,    -1,    -1,    67,    68,    69,    -1,
      -1,    72,    73,    74,    -1,    -1,    -1,    -1,    -1,    -1,
      81,    82,    83,    -1,    -1,    -1,    -1,    -1,    89,    90,
      -1,    -1,    -1,    94,    -1,    -1,    97,    98,    99,     1,
      -1,     3,     4,     5,     6,     7,     8,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    30,    -1,
      32,    -1,    34,    -1,    36,    37,    38,    39,    40,    -1,
      42,    -1,    -1,    45,    -1,    47,    -1,    49,    50,    -1,
      52,    -1,    54,    55,    -1,     1,    -1,     3,     1,    61,
       3,    63,    64,    -1,    -1,    67,    68,    69,    -1,    -1,
      72,    73,    74,    -1,    -1,    -1,    -1,    -1,    -1,    81,
      82,    83,    -1,    -1,    -1,    -1,    -1,    89,    90,    -1,
      -1,    37,    94,    -1,    37,    97,    -1,    99,     3,     4,
       5,     6,     7,     8,    -1,    -1,    52,    -1,    54,    52,
       1,    54,     3,    -1,    -1,    61,    -1,    -1,    61,    -1,
      -1,    -1,    27,    28,    29,    -1,    31,    73,    33,    -1,
      73,    -1,    37,    -1,    -1,    40,    82,    -1,    -1,    82,
      45,    -1,    -1,    -1,    -1,    -1,    37,    52,    -1,    54,
      55,    56,    57,    58,    59,    60,    61,    62,    63,    64,
      -1,    52,    67,    54,    69,    70,    -1,    -1,    73,    74,
      61,    76,    77,    78,    79,    80,    81,    82,    83,    -1,
      -1,    -1,    73,    -1,    89,    90,    -1,    -1,    -1,    94,
      -1,    82,    -1,    -1,    99,     3,     4,     5,     6,     7,
       8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,
      28,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    37,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    45,    -1,    -1,
      -1,    -1,    -1,    -1,    52,    -1,    54,    -1,    -1,    -1,
      -1,    -1,    -1,    61,    -1,    63,    64,    -1,    -1,    -1,
      -1,    -1,     3,    -1,    -1,    73,    74,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    82,    83,    -1,    -1,    -1,    -1,
      -1,    89,    90,    -1,    -1,    -1,    94,    -1,    29,    -1,
      31,    99,    33,    -1,    35,    -1,    37,    -1,    -1,    -1,
      -1,    -1,    -1,    44,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    52,    -1,    54,     3,    56,    57,    58,    59,    60,
      61,    62,    -1,    -1,    65,    66,    -1,    -1,    -1,    70,
      -1,    -1,    73,    -1,    75,    76,    77,    78,    79,    80,
      29,    82,    31,    -1,    33,    -1,    35,    -1,    37,    -1,
      -1,    -1,    -1,    -1,    -1,    44,    -1,    98,    -1,    -1,
      -1,    -1,    -1,    52,    -1,    54,     3,    56,    57,    58,
      59,    60,    61,    62,    -1,    -1,    65,    66,    -1,    -1,
      -1,    70,    -1,    -1,    73,    -1,    75,    76,    77,    78,
      79,    80,    29,    82,    31,    -1,    33,    -1,    -1,    -1,
      37,     3,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    98,
      -1,    -1,    -1,    -1,    -1,    52,    -1,    54,    -1,    56,
      57,    58,    59,    60,    61,    62,    -1,    29,    -1,    31,
      -1,    33,    -1,    70,    -1,    37,    73,    -1,    -1,    76,
      77,    78,    79,    80,    -1,    82,    -1,    -1,    -1,    -1,
      52,    -1,    54,    -1,    56,    57,    58,    59,    60,    61,
      62,    98,     3,     4,     5,     6,     7,     8,    70,    -1,
      -1,    73,    -1,    -1,    76,    77,    78,    79,    80,    -1,
      82,    -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    97,    37,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    45,    -1,    -1,    -1,    -1,    -1,
      -1,    52,    -1,    54,    -1,    -1,    -1,    -1,    -1,    -1,
      61,    -1,    63,    64,    -1,    -1,     3,     4,     5,     6,
       7,     8,    73,    74,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    82,    83,    -1,    -1,    -1,    -1,    -1,    89,    90,
      27,    28,    -1,    94,    -1,    -1,    97,    -1,    -1,    -1,
      37,    -1,    -1,    40,    -1,    -1,    -1,    -1,    45,    -1,
      -1,    -1,    -1,    -1,    -1,    52,    -1,    54,    55,    -1,
      -1,    -1,    -1,    -1,    61,    -1,    63,    64,    -1,    -1,
      67,    -1,    69,    -1,    -1,    -1,    73,    74,     3,     4,
       5,     6,     7,     8,    81,    82,    83,    -1,    -1,    -1,
      -1,    -1,    89,    90,    -1,    -1,    -1,    94,    95,    -1,
      -1,    -1,    27,    28,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    37,    -1,    -1,    40,    -1,    -1,    -1,    -1,
      45,    -1,    -1,    -1,    -1,    -1,    -1,    52,    -1,    54,
      55,    -1,    -1,    -1,    -1,    -1,    61,    -1,    63,    64,
      -1,    -1,    67,    -1,    69,    -1,    -1,    -1,    73,    74,
       3,     4,     5,     6,     7,     8,    81,    82,    83,    -1,
      -1,    -1,    -1,    -1,    89,    90,    -1,    -1,    -1,    94,
      95,    -1,    -1,    -1,    27,    28,    29,    -1,    31,    -1,
      33,    -1,    -1,    -1,    37,    -1,    -1,    40,    -1,    -1,
      -1,    -1,    45,    -1,    -1,    -1,    -1,    -1,    -1,    52,
      -1,    54,    55,    56,    57,    58,    59,    60,    61,    -1,
      63,    64,    -1,    -1,    67,    -1,    69,    70,    -1,    -1,
      73,    74,    -1,    76,    77,    78,    79,    80,    81,    82,
      83,    -1,    -1,    -1,    -1,    -1,    89,    90,    -1,    -1,
      -1,    94,     3,     4,     5,     6,     7,     8,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    37,    -1,    -1,    40,
      -1,     3,    -1,    -1,    45,    -1,    -1,    -1,    -1,    -1,
      -1,    52,    -1,    54,    55,    -1,    -1,    -1,    -1,    -1,
      61,    -1,    63,    64,    -1,    -1,    67,    29,    69,    31,
      -1,    33,    73,    74,    -1,    37,    -1,    -1,    -1,    -1,
      81,    82,    83,    -1,    -1,    -1,    -1,    -1,    89,    90,
      52,    -1,    54,    94,    56,    57,    58,    59,    60,    61,
      62,     3,     4,     5,     6,     7,     8,    -1,    70,    -1,
      -1,    73,    -1,    -1,    76,    77,    78,    79,    80,    -1,
      82,    -1,    -1,    -1,    -1,    27,    28,    -1,    -1,    -1,
      -1,    -1,    94,    -1,    -1,    37,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    45,     3,     4,     5,     6,     7,     8,
      52,    -1,    54,    -1,    -1,    -1,    -1,    -1,    -1,    61,
      -1,    63,    64,    -1,    -1,    -1,    -1,    -1,    27,    28,
      -1,    73,    74,    -1,    -1,    -1,    -1,    -1,    37,    -1,
      82,    83,    -1,    -1,    -1,    -1,    45,    89,    90,    -1,
      -1,    -1,    94,    52,    -1,    54,    -1,    -1,    -1,    -1,
      -1,    -1,    61,    -1,    63,    64,    -1,    -1,     3,     4,
       5,     6,     7,     8,    73,    74,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    82,    83,    -1,    -1,    -1,    -1,    -1,
      89,    90,    27,    28,    -1,    94,    -1,    -1,    -1,    -1,
      -1,    -1,    37,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      45,     3,     4,     5,     6,     7,     8,    52,    -1,    54,
      -1,    -1,    -1,    -1,    -1,    -1,    61,    -1,    63,    64,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    73,    74,
      -1,    -1,    -1,    -1,    -1,    37,    -1,    82,    83,    -1,
      -1,    -1,    -1,    45,    89,    90,    -1,     3,    -1,    94,
      52,    -1,    54,    -1,    -1,    -1,    -1,    -1,    -1,    61,
      -1,    63,    64,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    73,    74,    29,    -1,    31,    -1,    33,    -1,    35,
      82,    37,     3,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    94,    -1,    -1,    -1,    52,    -1,    54,    -1,
      56,    57,    58,    59,    60,    61,    62,    -1,    29,    65,
      31,    -1,    33,    -1,    70,    -1,    37,    73,    -1,    -1,
      76,    77,    78,    79,    80,    -1,    82,    -1,    -1,    -1,
      -1,    52,    -1,    54,    -1,    56,    57,    58,    59,    60,
      61,    62,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    70,
      -1,    -1,    73,    -1,    -1,    76,    77,    78,    79,    80,
      -1,    82
};

  /* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
     symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    37,    52,    61,   107,   108,   109,   142,     3,    37,
      52,    54,    61,    73,    82,   193,   194,     8,   194,     0,
     108,   109,   142,    51,   110,    97,   194,    97,    65,    66,
     137,   138,   143,    29,    31,    33,    35,    44,    56,    57,
      58,    59,    60,    62,    66,    70,    75,    76,    77,    78,
      79,    80,    98,   111,   112,   113,   114,   115,   116,   117,
     122,   123,   125,   129,   130,   131,   137,   194,    65,    99,
      48,    98,   137,   114,   116,   194,     1,   194,    94,   113,
       1,   194,    98,   112,   122,   129,    99,   100,     1,   118,
     119,   194,    97,   135,    94,   194,    99,    98,    97,   116,
      98,    71,     4,     5,     6,     7,     8,    27,    28,    40,
      45,    55,    63,    64,    67,    69,    74,    81,    83,    89,
      90,    94,   121,   162,   163,   164,   169,   170,   171,   172,
     173,   174,   175,   177,   179,   181,   183,   185,   186,   187,
     188,   191,   192,   194,    99,   101,   102,     1,    30,    32,
      34,    36,    38,    39,    42,    47,    49,    50,    68,    72,
      98,    99,   113,   122,   135,   136,   144,   145,   146,   147,
     149,   150,   151,   154,   157,   158,   159,   160,   161,   162,
     186,   194,   116,   132,   133,   134,    94,   126,   127,   194,
     101,    97,    97,   116,   165,   166,   194,    63,    94,   186,
     117,   161,    99,   101,     1,    22,   105,    21,    84,    85,
      86,    25,    26,   176,    23,    24,    87,    88,   178,    89,
      90,   180,    19,    20,   182,    91,    92,    93,   184,   186,
      27,    28,    94,    96,   102,   119,   120,   179,    99,    73,
      99,   175,    99,    54,    73,    82,   104,    54,    73,    82,
       1,    94,   135,   194,     1,    94,    99,   161,     1,    94,
      97,   155,    98,   113,   122,   144,    43,     1,    99,   100,
     101,     9,    10,    11,    12,    13,    14,    15,    16,    17,
      18,   148,   104,   119,    95,   101,   132,   101,   128,   100,
     116,   115,   124,   166,   167,   168,   169,   194,   102,    94,
     194,    95,    95,   162,    99,   171,   161,   172,   173,   174,
     175,   177,   179,   181,   183,   185,   164,   189,   190,   194,
     169,   103,   104,    98,   114,   115,   145,   147,   152,   161,
     169,   194,    99,    98,   161,    99,    98,   161,    98,   149,
     156,   194,    50,   135,    99,   161,   162,   161,   144,   194,
     134,    95,   127,    98,   170,    95,    99,    98,   115,   101,
     128,   104,   179,   189,   186,   104,    95,   101,   103,    99,
     153,   169,    53,   153,    53,    95,    95,    95,    98,   144,
      94,    99,    99,    94,   116,   139,   140,   141,    99,   168,
      98,   168,   103,    95,   169,   164,    95,   161,    99,   188,
      95,   161,   188,   135,   135,   155,   161,   140,   101,   102,
     135,    95,    95,   135,    95,    95,    95,    95,   141,   120,
     135,   135,   135,   135,   135,   103
};

  /* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint8 yyr1[] =
{
       0,   106,   107,   107,   107,   107,   107,   107,   108,   109,
     109,   110,   110,   111,   111,   111,   111,   111,   111,   112,
     112,   112,   113,   113,   114,   114,   115,   116,   116,   116,
     117,   117,   117,   117,   117,   117,   117,   117,   117,   117,
     117,   117,   117,   117,   118,   118,   119,   119,   120,   120,
     121,   121,   122,   122,   123,   123,   124,   124,   125,   125,
     126,   126,   127,   127,   128,   128,   129,   130,   130,   131,
     132,   132,   133,   133,   134,   135,   135,   136,   136,   136,
     136,   136,   136,   137,   138,   138,   138,   138,   139,   139,
     139,   140,   140,   141,   141,   142,   143,   143,   144,   144,
     144,   144,   144,   144,   144,   144,   144,   144,   145,   146,
     146,   147,   147,   148,   148,   148,   148,   148,   148,   148,
     148,   148,   148,   149,   149,   149,   150,   150,   150,   150,
     151,   151,   151,   151,   151,   151,   151,   151,   151,   152,
     152,   153,   153,   154,   154,   154,   155,   155,   156,   156,
     157,   157,   157,   157,   157,   158,   159,   159,   159,   159,
     159,   159,   159,   160,   161,   161,   162,   162,   163,   163,
     163,   163,   163,   164,   164,   164,   165,   165,   166,   167,
     167,   168,   168,   168,   169,   169,   170,   170,   171,   171,
     172,   172,   173,   173,   174,   174,   175,   175,   176,   176,
     177,   177,   178,   178,   178,   178,   179,   179,   180,   180,
     181,   181,   182,   182,   183,   183,   184,   184,   184,   185,
     185,   186,   186,   186,   187,   187,   187,   187,   188,   188,
     188,   188,   188,   188,   189,   189,   190,   190,   191,   191,
     191,   191,   192,   192,   192,   192,   192,   192,   192,   192,
     193,   193,   193,   193,   193,   193,   194,   194
};

  /* YYR2[YYN] -- Number of symbols on the right hand side of rule YYN.  */
static const yytype_uint8 yyr2[] =
{
       0,     2,     1,     1,     1,     2,     2,     2,     2,     5,
       6,     0,     2,     1,     1,     1,     2,     2,     2,     1,
       2,     3,     1,     2,     2,     4,     2,     1,     1,     6,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     3,     1,     4,     0,     1,
       1,     3,     1,     1,     6,     3,     2,     3,     6,     3,
       1,     3,     1,     3,     0,     1,     2,     1,     1,     4,
       0,     1,     1,     3,     2,     2,     3,     1,     1,     1,
       2,     2,     2,     7,     0,     1,     1,     2,     0,     3,
       1,     1,     3,     1,     4,     5,     2,     3,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     2,
       2,     4,     4,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     3,     3,     2,     5,     7,     3,     3,
       2,     5,     6,     7,     6,     7,     7,     7,     3,     1,
       1,     1,     2,     2,     5,     3,     2,     3,     1,     2,
       2,     2,     2,     3,     3,     3,     2,     2,     2,     2,
       2,     2,     2,     1,     1,     3,     1,     3,     1,     1,
       1,     1,     1,     1,     2,     2,     1,     4,     4,     1,
       3,     1,     1,     3,     1,     5,     1,     3,     1,     3,
       1,     3,     1,     3,     1,     3,     1,     3,     1,     1,
       1,     3,     1,     1,     1,     1,     1,     3,     1,     1,
       1,     3,     1,     1,     1,     3,     1,     1,     1,     1,
       4,     1,     2,     2,     1,     1,     1,     1,     1,     4,
       4,     3,     2,     2,     0,     1,     1,     3,     1,     1,
       3,     5,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1
};


#define yyerrok         (yyerrstatus = 0)
#define yyclearin       (yychar = YYEMPTY)
#define YYEMPTY         (-2)
#define YYEOF           0

#define YYACCEPT        goto yyacceptlab
#define YYABORT         goto yyabortlab
#define YYERROR         goto yyerrorlab


#define YYRECOVERING()  (!!yyerrstatus)

#define YYBACKUP(Token, Value)                                  \
do                                                              \
  if (yychar == YYEMPTY)                                        \
    {                                                           \
      yychar = (Token);                                         \
      yylval = (Value);                                         \
      YYPOPSTACK (yylen);                                       \
      yystate = *yyssp;                                         \
      goto yybackup;                                            \
    }                                                           \
  else                                                          \
    {                                                           \
      yyerror (&yylloc, parse, yyscanner, YY_("syntax error: cannot back up")); \
      YYERROR;                                                  \
    }                                                           \
while (0)

/* Error token number */
#define YYTERROR        1
#define YYERRCODE       256


/* YYLLOC_DEFAULT -- Set CURRENT to span from RHS[1] to RHS[N].
   If N is 0, then set CURRENT to the empty location which ends
   the previous symbol: RHS[0] (always defined).  */

#ifndef YYLLOC_DEFAULT
# define YYLLOC_DEFAULT(Current, Rhs, N)                                \
    do                                                                  \
      if (N)                                                            \
        {                                                               \
          (Current).first_line   = YYRHSLOC (Rhs, 1).first_line;        \
          (Current).first_column = YYRHSLOC (Rhs, 1).first_column;      \
          (Current).last_line    = YYRHSLOC (Rhs, N).last_line;         \
          (Current).last_column  = YYRHSLOC (Rhs, N).last_column;       \
        }                                                               \
      else                                                              \
        {                                                               \
          (Current).first_line   = (Current).last_line   =              \
            YYRHSLOC (Rhs, 0).last_line;                                \
          (Current).first_column = (Current).last_column =              \
            YYRHSLOC (Rhs, 0).last_column;                              \
        }                                                               \
    while (0)
#endif

#define YYRHSLOC(Rhs, K) ((Rhs)[K])


/* Enable debugging if requested.  */
#if YYDEBUG

# ifndef YYFPRINTF
#  include <stdio.h> /* INFRINGES ON USER NAME SPACE */
#  define YYFPRINTF fprintf
# endif

# define YYDPRINTF(Args)                        \
do {                                            \
  if (yydebug)                                  \
    YYFPRINTF Args;                             \
} while (0)


/* YY_LOCATION_PRINT -- Print the location on the stream.
   This macro was not mandated originally: define only if we know
   we won't break user code: when these are the locations we know.  */

#ifndef YY_LOCATION_PRINT
# if defined YYLTYPE_IS_TRIVIAL && YYLTYPE_IS_TRIVIAL

/* Print *YYLOCP on YYO.  Private, do not rely on its existence. */

YY_ATTRIBUTE_UNUSED
static unsigned
yy_location_print_ (FILE *yyo, YYLTYPE const * const yylocp)
{
  unsigned res = 0;
  int end_col = 0 != yylocp->last_column ? yylocp->last_column - 1 : 0;
  if (0 <= yylocp->first_line)
    {
      res += YYFPRINTF (yyo, "%d", yylocp->first_line);
      if (0 <= yylocp->first_column)
        res += YYFPRINTF (yyo, ".%d", yylocp->first_column);
    }
  if (0 <= yylocp->last_line)
    {
      if (yylocp->first_line < yylocp->last_line)
        {
          res += YYFPRINTF (yyo, "-%d", yylocp->last_line);
          if (0 <= end_col)
            res += YYFPRINTF (yyo, ".%d", end_col);
        }
      else if (0 <= end_col && yylocp->first_column < end_col)
        res += YYFPRINTF (yyo, "-%d", end_col);
    }
  return res;
 }

#  define YY_LOCATION_PRINT(File, Loc)          \
  yy_location_print_ (File, &(Loc))

# else
#  define YY_LOCATION_PRINT(File, Loc) ((void) 0)
# endif
#endif


# define YY_SYMBOL_PRINT(Title, Type, Value, Location)                    \
do {                                                                      \
  if (yydebug)                                                            \
    {                                                                     \
      YYFPRINTF (stderr, "%s ", Title);                                   \
      yy_symbol_print (stderr,                                            \
                  Type, Value, Location, parse, yyscanner); \
      YYFPRINTF (stderr, "\n");                                           \
    }                                                                     \
} while (0)


/*----------------------------------------.
| Print this symbol's value on YYOUTPUT.  |
`----------------------------------------*/

static void
yy_symbol_value_print (FILE *yyoutput, int yytype, YYSTYPE const * const yyvaluep, YYLTYPE const * const yylocationp, parse_t *parse, void *yyscanner)
{
  FILE *yyo = yyoutput;
  YYUSE (yyo);
  YYUSE (yylocationp);
  YYUSE (parse);
  YYUSE (yyscanner);
  if (!yyvaluep)
    return;
# ifdef YYPRINT
  if (yytype < YYNTOKENS)
    YYPRINT (yyoutput, yytoknum[yytype], *yyvaluep);
# endif
  YYUSE (yytype);
}


/*--------------------------------.
| Print this symbol on YYOUTPUT.  |
`--------------------------------*/

static void
yy_symbol_print (FILE *yyoutput, int yytype, YYSTYPE const * const yyvaluep, YYLTYPE const * const yylocationp, parse_t *parse, void *yyscanner)
{
  YYFPRINTF (yyoutput, "%s %s (",
             yytype < YYNTOKENS ? "token" : "nterm", yytname[yytype]);

  YY_LOCATION_PRINT (yyoutput, *yylocationp);
  YYFPRINTF (yyoutput, ": ");
  yy_symbol_value_print (yyoutput, yytype, yyvaluep, yylocationp, parse, yyscanner);
  YYFPRINTF (yyoutput, ")");
}

/*------------------------------------------------------------------.
| yy_stack_print -- Print the state stack from its BOTTOM up to its |
| TOP (included).                                                   |
`------------------------------------------------------------------*/

static void
yy_stack_print (yytype_int16 *yybottom, yytype_int16 *yytop)
{
  YYFPRINTF (stderr, "Stack now");
  for (; yybottom <= yytop; yybottom++)
    {
      int yybot = *yybottom;
      YYFPRINTF (stderr, " %d", yybot);
    }
  YYFPRINTF (stderr, "\n");
}

# define YY_STACK_PRINT(Bottom, Top)                            \
do {                                                            \
  if (yydebug)                                                  \
    yy_stack_print ((Bottom), (Top));                           \
} while (0)


/*------------------------------------------------.
| Report that the YYRULE is going to be reduced.  |
`------------------------------------------------*/

static void
yy_reduce_print (yytype_int16 *yyssp, YYSTYPE *yyvsp, YYLTYPE *yylsp, int yyrule, parse_t *parse, void *yyscanner)
{
  unsigned long int yylno = yyrline[yyrule];
  int yynrhs = yyr2[yyrule];
  int yyi;
  YYFPRINTF (stderr, "Reducing stack by rule %d (line %lu):\n",
             yyrule - 1, yylno);
  /* The symbols being reduced.  */
  for (yyi = 0; yyi < yynrhs; yyi++)
    {
      YYFPRINTF (stderr, "   $%d = ", yyi + 1);
      yy_symbol_print (stderr,
                       yystos[yyssp[yyi + 1 - yynrhs]],
                       &(yyvsp[(yyi + 1) - (yynrhs)])
                       , &(yylsp[(yyi + 1) - (yynrhs)])                       , parse, yyscanner);
      YYFPRINTF (stderr, "\n");
    }
}

# define YY_REDUCE_PRINT(Rule)          \
do {                                    \
  if (yydebug)                          \
    yy_reduce_print (yyssp, yyvsp, yylsp, Rule, parse, yyscanner); \
} while (0)

/* Nonzero means print parse trace.  It is left uninitialized so that
   multiple parsers can coexist.  */
int yydebug;
#else /* !YYDEBUG */
# define YYDPRINTF(Args)
# define YY_SYMBOL_PRINT(Title, Type, Value, Location)
# define YY_STACK_PRINT(Bottom, Top)
# define YY_REDUCE_PRINT(Rule)
#endif /* !YYDEBUG */


/* YYINITDEPTH -- initial size of the parser's stacks.  */
#ifndef YYINITDEPTH
# define YYINITDEPTH 200
#endif

/* YYMAXDEPTH -- maximum size the stacks can grow to (effective only
   if the built-in stack extension method is used).

   Do not make this value too large; the results are undefined if
   YYSTACK_ALLOC_MAXIMUM < YYSTACK_BYTES (YYMAXDEPTH)
   evaluated with infinite-precision integer arithmetic.  */

#ifndef YYMAXDEPTH
# define YYMAXDEPTH 10000
#endif


#if YYERROR_VERBOSE

# ifndef yystrlen
#  if defined __GLIBC__ && defined _STRING_H
#   define yystrlen strlen
#  else
/* Return the length of YYSTR.  */
static YYSIZE_T
yystrlen (const char *yystr)
{
  YYSIZE_T yylen;
  for (yylen = 0; yystr[yylen]; yylen++)
    continue;
  return yylen;
}
#  endif
# endif

# ifndef yystpcpy
#  if defined __GLIBC__ && defined _STRING_H && defined _GNU_SOURCE
#   define yystpcpy stpcpy
#  else
/* Copy YYSRC to YYDEST, returning the address of the terminating '\0' in
   YYDEST.  */
static char *
yystpcpy (char *yydest, const char *yysrc)
{
  char *yyd = yydest;
  const char *yys = yysrc;

  while ((*yyd++ = *yys++) != '\0')
    continue;

  return yyd - 1;
}
#  endif
# endif

# ifndef yytnamerr
/* Copy to YYRES the contents of YYSTR after stripping away unnecessary
   quotes and backslashes, so that it's suitable for yyerror.  The
   heuristic is that double-quoting is unnecessary unless the string
   contains an apostrophe, a comma, or backslash (other than
   backslash-backslash).  YYSTR is taken from yytname.  If YYRES is
   null, do not copy; instead, return the length of what the result
   would have been.  */
static YYSIZE_T
yytnamerr (char *yyres, const char *yystr)
{
  if (*yystr == '"')
    {
      YYSIZE_T yyn = 0;
      char const *yyp = yystr;

      for (;;)
        switch (*++yyp)
          {
          case '\'':
          case ',':
            goto do_not_strip_quotes;

          case '\\':
            if (*++yyp != '\\')
              goto do_not_strip_quotes;
            /* Fall through.  */
          default:
            if (yyres)
              yyres[yyn] = *yyp;
            yyn++;
            break;

          case '"':
            if (yyres)
              yyres[yyn] = '\0';
            return yyn;
          }
    do_not_strip_quotes: ;
    }

  if (! yyres)
    return yystrlen (yystr);

  return yystpcpy (yyres, yystr) - yyres;
}
# endif

/* Copy into *YYMSG, which is of size *YYMSG_ALLOC, an error message
   about the unexpected token YYTOKEN for the state stack whose top is
   YYSSP.

   Return 0 if *YYMSG was successfully written.  Return 1 if *YYMSG is
   not large enough to hold the message.  In that case, also set
   *YYMSG_ALLOC to the required number of bytes.  Return 2 if the
   required number of bytes is too large to store.  */
static int
yysyntax_error (YYSIZE_T *yymsg_alloc, char **yymsg,
                yytype_int16 *yyssp, int yytoken)
{
  YYSIZE_T yysize0 = yytnamerr (YY_NULLPTR, yytname[yytoken]);
  YYSIZE_T yysize = yysize0;
  enum { YYERROR_VERBOSE_ARGS_MAXIMUM = 5 };
  /* Internationalized format string. */
  const char *yyformat = YY_NULLPTR;
  /* Arguments of yyformat. */
  char const *yyarg[YYERROR_VERBOSE_ARGS_MAXIMUM];
  /* Number of reported tokens (one for the "unexpected", one per
     "expected"). */
  int yycount = 0;

  /* There are many possibilities here to consider:
     - If this state is a consistent state with a default action, then
       the only way this function was invoked is if the default action
       is an error action.  In that case, don't check for expected
       tokens because there are none.
     - The only way there can be no lookahead present (in yychar) is if
       this state is a consistent state with a default action.  Thus,
       detecting the absence of a lookahead is sufficient to determine
       that there is no unexpected or expected token to report.  In that
       case, just report a simple "syntax error".
     - Don't assume there isn't a lookahead just because this state is a
       consistent state with a default action.  There might have been a
       previous inconsistent state, consistent state with a non-default
       action, or user semantic action that manipulated yychar.
     - Of course, the expected token list depends on states to have
       correct lookahead information, and it depends on the parser not
       to perform extra reductions after fetching a lookahead from the
       scanner and before detecting a syntax error.  Thus, state merging
       (from LALR or IELR) and default reductions corrupt the expected
       token list.  However, the list is correct for canonical LR with
       one exception: it will still contain any token that will not be
       accepted due to an error action in a later state.
  */
  if (yytoken != YYEMPTY)
    {
      int yyn = yypact[*yyssp];
      yyarg[yycount++] = yytname[yytoken];
      if (!yypact_value_is_default (yyn))
        {
          /* Start YYX at -YYN if negative to avoid negative indexes in
             YYCHECK.  In other words, skip the first -YYN actions for
             this state because they are default actions.  */
          int yyxbegin = yyn < 0 ? -yyn : 0;
          /* Stay within bounds of both yycheck and yytname.  */
          int yychecklim = YYLAST - yyn + 1;
          int yyxend = yychecklim < YYNTOKENS ? yychecklim : YYNTOKENS;
          int yyx;

          for (yyx = yyxbegin; yyx < yyxend; ++yyx)
            if (yycheck[yyx + yyn] == yyx && yyx != YYTERROR
                && !yytable_value_is_error (yytable[yyx + yyn]))
              {
                if (yycount == YYERROR_VERBOSE_ARGS_MAXIMUM)
                  {
                    yycount = 1;
                    yysize = yysize0;
                    break;
                  }
                yyarg[yycount++] = yytname[yyx];
                {
                  YYSIZE_T yysize1 = yysize + yytnamerr (YY_NULLPTR, yytname[yyx]);
                  if (! (yysize <= yysize1
                         && yysize1 <= YYSTACK_ALLOC_MAXIMUM))
                    return 2;
                  yysize = yysize1;
                }
              }
        }
    }

  switch (yycount)
    {
# define YYCASE_(N, S)                      \
      case N:                               \
        yyformat = S;                       \
      break
    default: /* Avoid compiler warnings. */
      YYCASE_(0, YY_("syntax error"));
      YYCASE_(1, YY_("syntax error, unexpected %s"));
      YYCASE_(2, YY_("syntax error, unexpected %s, expecting %s"));
      YYCASE_(3, YY_("syntax error, unexpected %s, expecting %s or %s"));
      YYCASE_(4, YY_("syntax error, unexpected %s, expecting %s or %s or %s"));
      YYCASE_(5, YY_("syntax error, unexpected %s, expecting %s or %s or %s or %s"));
# undef YYCASE_
    }

  {
    YYSIZE_T yysize1 = yysize + yystrlen (yyformat);
    if (! (yysize <= yysize1 && yysize1 <= YYSTACK_ALLOC_MAXIMUM))
      return 2;
    yysize = yysize1;
  }

  if (*yymsg_alloc < yysize)
    {
      *yymsg_alloc = 2 * yysize;
      if (! (yysize <= *yymsg_alloc
             && *yymsg_alloc <= YYSTACK_ALLOC_MAXIMUM))
        *yymsg_alloc = YYSTACK_ALLOC_MAXIMUM;
      return 1;
    }

  /* Avoid sprintf, as that infringes on the user's name space.
     Don't have undefined behavior even if the translation
     produced a string with the wrong number of "%s"s.  */
  {
    char *yyp = *yymsg;
    int yyi = 0;
    while ((*yyp = *yyformat) != '\0')
      if (*yyp == '%' && yyformat[1] == 's' && yyi < yycount)
        {
          yyp += yytnamerr (yyp, yyarg[yyi++]);
          yyformat += 2;
        }
      else
        {
          yyp++;
          yyformat++;
        }
  }
  return 0;
}
#endif /* YYERROR_VERBOSE */

/*-----------------------------------------------.
| Release the memory associated to this symbol.  |
`-----------------------------------------------*/

static void
yydestruct (const char *yymsg, int yytype, YYSTYPE *yyvaluep, YYLTYPE *yylocationp, parse_t *parse, void *yyscanner)
{
  YYUSE (yyvaluep);
  YYUSE (yylocationp);
  YYUSE (parse);
  YYUSE (yyscanner);
  if (!yymsg)
    yymsg = "Deleting";
  YY_SYMBOL_PRINT (yymsg, yytype, yyvaluep, yylocationp);

  YY_IGNORE_MAYBE_UNINITIALIZED_BEGIN
  YYUSE (yytype);
  YY_IGNORE_MAYBE_UNINITIALIZED_END
}




/*----------.
| yyparse.  |
`----------*/

int
yyparse (parse_t *parse, void *yyscanner)
{
/* The lookahead symbol.  */
int yychar;


/* The semantic value of the lookahead symbol.  */
/* Default value used for initialization, for pacifying older GCCs
   or non-GCC compilers.  */
YY_INITIAL_VALUE (static YYSTYPE yyval_default;)
YYSTYPE yylval YY_INITIAL_VALUE (= yyval_default);

/* Location data for the lookahead symbol.  */
static YYLTYPE yyloc_default
# if defined YYLTYPE_IS_TRIVIAL && YYLTYPE_IS_TRIVIAL
  = { 1, 1, 1, 1 }
# endif
;
YYLTYPE yylloc = yyloc_default;

    /* Number of syntax errors so far.  */
    int yynerrs;

    int yystate;
    /* Number of tokens to shift before error messages enabled.  */
    int yyerrstatus;

    /* The stacks and their tools:
       'yyss': related to states.
       'yyvs': related to semantic values.
       'yyls': related to locations.

       Refer to the stacks through separate pointers, to allow yyoverflow
       to reallocate them elsewhere.  */

    /* The state stack.  */
    yytype_int16 yyssa[YYINITDEPTH];
    yytype_int16 *yyss;
    yytype_int16 *yyssp;

    /* The semantic value stack.  */
    YYSTYPE yyvsa[YYINITDEPTH];
    YYSTYPE *yyvs;
    YYSTYPE *yyvsp;

    /* The location stack.  */
    YYLTYPE yylsa[YYINITDEPTH];
    YYLTYPE *yyls;
    YYLTYPE *yylsp;

    /* The locations where the error started and ended.  */
    YYLTYPE yyerror_range[3];

    YYSIZE_T yystacksize;

  int yyn;
  int yyresult;
  /* Lookahead token as an internal (translated) token number.  */
  int yytoken = 0;
  /* The variables used to return semantic value and location from the
     action routines.  */
  YYSTYPE yyval;
  YYLTYPE yyloc;

#if YYERROR_VERBOSE
  /* Buffer for error messages, and its allocated size.  */
  char yymsgbuf[128];
  char *yymsg = yymsgbuf;
  YYSIZE_T yymsg_alloc = sizeof yymsgbuf;
#endif

#define YYPOPSTACK(N)   (yyvsp -= (N), yyssp -= (N), yylsp -= (N))

  /* The number of symbols on the RHS of the reduced rule.
     Keep to zero when no symbol should be popped.  */
  int yylen = 0;

  yyssp = yyss = yyssa;
  yyvsp = yyvs = yyvsa;
  yylsp = yyls = yylsa;
  yystacksize = YYINITDEPTH;

  YYDPRINTF ((stderr, "Starting parse\n"));

  yystate = 0;
  yyerrstatus = 0;
  yynerrs = 0;
  yychar = YYEMPTY; /* Cause a token to be read.  */

/* User initialization code.  */
#line 36 "grammar.y" /* yacc.c:1430  */
{
    src_pos_init(&yylloc, parse->src, parse->path);
}

#line 1890 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1430  */
  yylsp[0] = yylloc;
  goto yysetstate;

/*------------------------------------------------------------.
| yynewstate -- Push a new state, which is found in yystate.  |
`------------------------------------------------------------*/
 yynewstate:
  /* In all cases, when you get here, the value and location stacks
     have just been pushed.  So pushing a state here evens the stacks.  */
  yyssp++;

 yysetstate:
  *yyssp = yystate;

  if (yyss + yystacksize - 1 <= yyssp)
    {
      /* Get the current used size of the three stacks, in elements.  */
      YYSIZE_T yysize = yyssp - yyss + 1;

#ifdef yyoverflow
      {
        /* Give user a chance to reallocate the stack.  Use copies of
           these so that the &'s don't force the real ones into
           memory.  */
        YYSTYPE *yyvs1 = yyvs;
        yytype_int16 *yyss1 = yyss;
        YYLTYPE *yyls1 = yyls;

        /* Each stack pointer address is followed by the size of the
           data in use in that stack, in bytes.  This used to be a
           conditional around just the two extra args, but that might
           be undefined if yyoverflow is a macro.  */
        yyoverflow (YY_("memory exhausted"),
                    &yyss1, yysize * sizeof (*yyssp),
                    &yyvs1, yysize * sizeof (*yyvsp),
                    &yyls1, yysize * sizeof (*yylsp),
                    &yystacksize);

        yyls = yyls1;
        yyss = yyss1;
        yyvs = yyvs1;
      }
#else /* no yyoverflow */
# ifndef YYSTACK_RELOCATE
      goto yyexhaustedlab;
# else
      /* Extend the stack our own way.  */
      if (YYMAXDEPTH <= yystacksize)
        goto yyexhaustedlab;
      yystacksize *= 2;
      if (YYMAXDEPTH < yystacksize)
        yystacksize = YYMAXDEPTH;

      {
        yytype_int16 *yyss1 = yyss;
        union yyalloc *yyptr =
          (union yyalloc *) YYSTACK_ALLOC (YYSTACK_BYTES (yystacksize));
        if (! yyptr)
          goto yyexhaustedlab;
        YYSTACK_RELOCATE (yyss_alloc, yyss);
        YYSTACK_RELOCATE (yyvs_alloc, yyvs);
        YYSTACK_RELOCATE (yyls_alloc, yyls);
#  undef YYSTACK_RELOCATE
        if (yyss1 != yyssa)
          YYSTACK_FREE (yyss1);
      }
# endif
#endif /* no yyoverflow */

      yyssp = yyss + yysize - 1;
      yyvsp = yyvs + yysize - 1;
      yylsp = yyls + yysize - 1;

      YYDPRINTF ((stderr, "Stack size increased to %lu\n",
                  (unsigned long int) yystacksize));

      if (yyss + yystacksize - 1 <= yyssp)
        YYABORT;
    }

  YYDPRINTF ((stderr, "Entering state %d\n", yystate));

  if (yystate == YYFINAL)
    YYACCEPT;

  goto yybackup;

/*-----------.
| yybackup.  |
`-----------*/
yybackup:

  /* Do appropriate processing given the current state.  Read a
     lookahead token if we need one and don't already have one.  */

  /* First try to decide what to do without reference to lookahead token.  */
  yyn = yypact[yystate];
  if (yypact_value_is_default (yyn))
    goto yydefault;

  /* Not known => get a lookahead token if don't already have one.  */

  /* YYCHAR is either YYEMPTY or YYEOF or a valid lookahead symbol.  */
  if (yychar == YYEMPTY)
    {
      YYDPRINTF ((stderr, "Reading a token: "));
      yychar = yylex (&yylval, &yylloc, yyscanner);
    }

  if (yychar <= YYEOF)
    {
      yychar = yytoken = YYEOF;
      YYDPRINTF ((stderr, "Now at end of input.\n"));
    }
  else
    {
      yytoken = YYTRANSLATE (yychar);
      YY_SYMBOL_PRINT ("Next token is", yytoken, &yylval, &yylloc);
    }

  /* If the proper action on seeing token YYTOKEN is to reduce or to
     detect an error, take that action.  */
  yyn += yytoken;
  if (yyn < 0 || YYLAST < yyn || yycheck[yyn] != yytoken)
    goto yydefault;
  yyn = yytable[yyn];
  if (yyn <= 0)
    {
      if (yytable_value_is_error (yyn))
        goto yyerrlab;
      yyn = -yyn;
      goto yyreduce;
    }

  /* Count tokens shifted since error; after three, turn off error
     status.  */
  if (yyerrstatus)
    yyerrstatus--;

  /* Shift the lookahead token.  */
  YY_SYMBOL_PRINT ("Shifting", yytoken, &yylval, &yylloc);

  /* Discard the shifted token.  */
  yychar = YYEMPTY;

  yystate = yyn;
  YY_IGNORE_MAYBE_UNINITIALIZED_BEGIN
  *++yyvsp = yylval;
  YY_IGNORE_MAYBE_UNINITIALIZED_END
  *++yylsp = yylloc;
  goto yynewstate;


/*-----------------------------------------------------------.
| yydefault -- do the default action for the current state.  |
`-----------------------------------------------------------*/
yydefault:
  yyn = yydefact[yystate];
  if (yyn == 0)
    goto yyerrlab;
  goto yyreduce;


/*-----------------------------.
| yyreduce -- Do a reduction.  |
`-----------------------------*/
yyreduce:
  /* yyn is the number of a rule to reduce with.  */
  yylen = yyr2[yyn];

  /* If YYLEN is nonzero, implement the default value of the action:
     '$$ = $1'.

     Otherwise, the following line sets YYVAL to garbage.
     This behavior is undocumented and Bison
     users should not rely upon it.  Assigning to YYVAL
     unconditionally makes the parser a bit smaller, and it avoids a
     GCC warning that YYVAL may be used uninitialized.  */
  yyval = yyvsp[1-yylen];

  /* Default location. */
  YYLLOC_DEFAULT (yyloc, (yylsp - yylen), yylen);
  yyerror_range[1] = yyloc;
  YY_REDUCE_PRINT (yyn);
  switch (yyn)
    {
        case 2:
#line 258 "grammar.y" /* yacc.c:1648  */
    {
        AST = ast_new();
        imp_add(&AST->imps, (yyvsp[0].imp));
    }
#line 2083 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 3:
#line 263 "grammar.y" /* yacc.c:1648  */
    {
        AST = ast_new();
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2092 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 4:
#line 268 "grammar.y" /* yacc.c:1648  */
    {
        AST = ast_new();
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2101 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 5:
#line 273 "grammar.y" /* yacc.c:1648  */
    {
        imp_add(&AST->imps, (yyvsp[0].imp));
    }
#line 2109 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 6:
#line 277 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2117 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 7:
#line 281 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2125 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 8:
#line 288 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.imp) = imp_new((yyvsp[0].str), &(yylsp[0]));
    }
#line 2133 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 9:
#line 295 "grammar.y" /* yacc.c:1648  */
    {
        ast_blk_t *blk = blk_new_contract(&(yylsp[-1]));

        /* add default constructor */
        id_add(&blk->ids, id_new_ctor((yyvsp[-3].str), NULL, NULL, &(yylsp[-3])));

        (yyval.id) = id_new_contract((yyvsp[-3].str), (yyvsp[-2].exp), blk, &(yyloc));
    }
#line 2146 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 10:
#line 304 "grammar.y" /* yacc.c:1648  */
    {
        int i;
        bool has_ctor = false;

        vector_foreach(&(yyvsp[-1].blk)->ids, i) {
            if (is_ctor_id(vector_get_id(&(yyvsp[-1].blk)->ids, i)))
                has_ctor = true;
        }

        if (!has_ctor)
            /* add default constructor */
            id_add(&(yyvsp[-1].blk)->ids, id_new_ctor((yyvsp[-4].str), NULL, NULL, &(yylsp[-4])));

        (yyval.id) = id_new_contract((yyvsp[-4].str), (yyvsp[-3].exp), (yyvsp[-1].blk), &(yyloc));
    }
#line 2166 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 11:
#line 322 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 2172 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 12:
#line 324 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yylsp[0]));
    }
#line 2180 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 13:
#line 331 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2189 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 14:
#line 336 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2198 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 15:
#line 341 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2207 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 16:
#line 346 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2216 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 17:
#line 351 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2225 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 18:
#line 356 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2234 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 20:
#line 365 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_PUBLIC;
    }
#line 2243 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 21:
#line 370 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2252 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 23:
#line 379 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_CONST;
    }
#line 2261 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 25:
#line 388 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.dflt_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.dflt_exp = (yyvsp[-1].exp);
    }
#line 2274 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 26:
#line 400 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.type_exp = (yyvsp[-1].exp);
    }
#line 2287 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 27:
#line 412 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type((yyvsp[0].type), &(yylsp[0]));
    }
#line 2295 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 28:
#line 416 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_NONE, &(yylsp[0]));

        (yyval.exp)->u_type.name = (yyvsp[0].str);
    }
#line 2305 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 29:
#line 422 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_MAP, &(yylsp[-5]));

        (yyval.exp)->u_type.k_exp = (yyvsp[-3].exp);
        (yyval.exp)->u_type.v_exp = (yyvsp[-1].exp);
    }
#line 2316 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 30:
#line 431 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_ACCOUNT; }
#line 2322 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 31:
#line 432 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BOOL; }
#line 2328 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 32:
#line 433 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BYTE; }
#line 2334 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 33:
#line 436 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2340 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 34:
#line 437 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT16; }
#line 2346 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 35:
#line 438 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2352 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 36:
#line 439 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT64; }
#line 2358 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 37:
#line 440 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT8; }
#line 2364 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 38:
#line 441 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_STRING; }
#line 2370 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 39:
#line 442 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2376 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 40:
#line 443 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT16; }
#line 2382 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 41:
#line 444 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2388 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 42:
#line 445 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT64; }
#line 2394 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 43:
#line 446 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT8; }
#line 2400 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 45:
#line 452 "grammar.y" /* yacc.c:1648  */
    {
        if (is_tuple_id((yyvsp[-2].id))) {
            (yyval.id) = (yyvsp[-2].id);
        }
        else {
            (yyval.id) = id_new_tuple(&(yylsp[-2]));
            id_add((yyval.id)->u_tup.elem_ids, (yyvsp[-2].id));
        }

        id_add((yyval.id)->u_tup.elem_ids, (yyvsp[0].id));
    }
#line 2416 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 46:
#line 467 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PRIVATE, &(yylsp[0]));
    }
#line 2424 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 47:
#line 471 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2437 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 48:
#line 482 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = exp_new_null(&(yyloc)); }
#line 2443 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 50:
#line 488 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 2451 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 51:
#line 492 "grammar.y" /* yacc.c:1648  */
    {
        if (is_tuple_exp((yyvsp[-2].exp))) {
            (yyval.exp) = (yyvsp[-2].exp);
        }
        else {
            (yyval.exp) = exp_new_tuple(vector_new(), &(yylsp[-2]));
            exp_add((yyval.exp)->u_tup.elem_exps, (yyvsp[-2].exp));
        }
        exp_add((yyval.exp)->u_tup.elem_exps, (yyvsp[0].exp));
    }
#line 2466 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 54:
#line 511 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_struct((yyvsp[-4].str), (yyvsp[-1].vect), &(yyloc));
    }
#line 2474 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 55:
#line 515 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2483 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 56:
#line 523 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2496 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 57:
#line 532 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2509 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 58:
#line 544 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_enum((yyvsp[-4].str), (yyvsp[-2].vect), &(yyloc));
    }
#line 2517 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 59:
#line 548 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2526 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 60:
#line 556 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2535 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 61:
#line 561 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2544 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 62:
#line 569 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PUBLIC | MOD_CONST, &(yylsp[0]));
    }
#line 2552 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 63:
#line 573 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[-2].str), MOD_PUBLIC | MOD_CONST, &(yylsp[-2]));
        (yyval.id)->u_var.dflt_exp = (yyvsp[0].exp);
    }
#line 2561 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 66:
#line 586 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-1].id);
        (yyval.id)->u_fn.blk = (yyvsp[0].blk);

        /* The label is added to the topmost block because it can be referenced
         * regardless of the order of declaration. */
        if (!is_empty_vector(LABELS)) {
            ASSERT((yyvsp[0].blk) != NULL);
            id_join(&(yyvsp[0].blk)->ids, LABELS);
            vector_reset(LABELS);
        }
    }
#line 2578 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 69:
#line 607 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_ctor((yyvsp[-3].str), (yyvsp[-1].vect), NULL, &(yylsp[-3]));
    }
#line 2586 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 70:
#line 613 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 2592 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 72:
#line 619 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2601 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 73:
#line 624 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2610 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 74:
#line 632 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->u_var.is_param = true;
        (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
    }
#line 2620 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 75:
#line 640 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 2626 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 76:
#line 641 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 2632 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 77:
#line 646 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2641 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 78:
#line 651 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        /* Unlike state variables, local variables are referenced according to their
         * order of declaration. */
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2652 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 79:
#line 658 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2661 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 80:
#line 663 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2670 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 81:
#line 668 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2679 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 82:
#line 673 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2688 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 83:
#line 681 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_func((yyvsp[-4].str), (yyvsp[-6].mod), (yyvsp[-2].vect), (yyvsp[0].id), NULL, &(yylsp[-4]));
    }
#line 2696 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 84:
#line 687 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PRIVATE; }
#line 2702 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 85:
#line 688 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC; }
#line 2708 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 86:
#line 689 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PAYABLE; }
#line 2714 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 87:
#line 690 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC | MOD_PAYABLE; }
#line 2720 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 88:
#line 694 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = NULL; }
#line 2726 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 89:
#line 695 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = (yyvsp[-1].id); }
#line 2732 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 92:
#line 702 "grammar.y" /* yacc.c:1648  */
    {
        yyerror(&(yylsp[0]), parse, yyscanner, "not supported yet");

        /* TODO multiple return values
         *
         * Since WebAssembly does not support this syntax at present, it is better to
         * implement it when it is formally supported than implementing arbitarily
         */
#if 0
        /* The reason for making the return list as tuple is because of
         * the convenience of meta comparison. If vector_t is used,
         * it must be looped for each id and compared directly,
         * but for tuples, meta_cmp() is sufficient */

        if (is_tuple_id((yyvsp[-2].id))) {
            (yyval.id) = (yyvsp[-2].id);
        }
        else {
            (yyval.id) = id_new_tuple(&(yylsp[-2]));
            id_add((yyval.id)->u_tup.elem_ids, (yyvsp[-2].id));
        }

        id_add((yyval.id)->u_tup.elem_ids, (yyvsp[0].id));
#endif
    }
#line 2762 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 93:
#line 731 "grammar.y" /* yacc.c:1648  */
    {
        /* We wanted to use a type expression, but we can not store size expressions
         * and declare it as an identifier. */
        (yyval.id) = id_new_param(NULL, (yyvsp[0].exp), &(yylsp[0]));
    }
#line 2772 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 94:
#line 737 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2785 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 95:
#line 749 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_interface((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3]));
    }
#line 2793 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 96:
#line 756 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_interface(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2802 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 97:
#line 761 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2811 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 108:
#line 782 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_null(&(yyloc));
    }
#line 2819 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 109:
#line 789 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_exp((yyvsp[-1].exp), &(yyloc));
    }
#line 2827 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 110:
#line 793 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2836 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 111:
#line 801 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2]));
    }
#line 2844 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 112:
#line 805 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), exp_new_binary((yyvsp[-2].op), (yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2])), &(yylsp[-2]));
    }
#line 2852 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 113:
#line 811 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 2858 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 114:
#line 812 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 2864 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 115:
#line 813 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 2870 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 116:
#line 814 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 2876 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 117:
#line 815 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 2882 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 118:
#line 816 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_AND; }
#line 2888 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 119:
#line 817 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_XOR; }
#line 2894 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 120:
#line 818 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_OR; }
#line 2900 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 121:
#line 819 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 2906 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 122:
#line 820 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 2912 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 123:
#line 825 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[0].stmt);
        id_add(LABELS, id_new_label((yyvsp[-2].str), (yyvsp[0].stmt), &(yylsp[-2])));
    }
#line 2921 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 124:
#line 830 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case((yyvsp[-1].exp), &(yyloc));
    }
#line 2929 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 125:
#line 834 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case(NULL, &(yyloc));
    }
#line 2937 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 126:
#line 841 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 2945 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 127:
#line 845 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-6].stmt);
        stmt_add(&(yyval.stmt)->u_if.elif_stmts, stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yylsp[-5])));
    }
#line 2954 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 128:
#line 850 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-2].stmt);
        (yyval.stmt)->u_if.else_blk = (yyvsp[0].blk);
    }
#line 2963 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 129:
#line 855 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2972 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 130:
#line 863 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, NULL, NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2980 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 131:
#line 867 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2988 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 132:
#line 871 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-3].stmt), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2996 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 133:
#line 875 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-4].stmt), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3004 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 134:
#line 879 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-3].id), &(yylsp[-3])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3012 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 135:
#line 883 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3020 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 136:
#line 887 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_exp((yyvsp[-4].exp), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3028 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 137:
#line 891 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3036 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 138:
#line 895 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3045 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 141:
#line 907 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 3051 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 143:
#line 913 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch(NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3059 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 144:
#line 917 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3067 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 145:
#line 921 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3076 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 146:
#line 928 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 3082 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 147:
#line 929 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 3088 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 148:
#line 934 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_switch(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3097 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 149:
#line 939 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3106 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 150:
#line 947 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_CONTINUE, NULL, &(yyloc));
    }
#line 3114 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 151:
#line 951 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_BREAK, NULL, &(yyloc));
    }
#line 3122 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 152:
#line 955 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return(NULL, &(yyloc));
    }
#line 3130 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 153:
#line 959 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return((yyvsp[-1].exp), &(yyloc));
    }
#line 3138 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 154:
#line 963 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_goto((yyvsp[-1].str), &(yylsp[-1]));
    }
#line 3146 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 155:
#line 970 "grammar.y" /* yacc.c:1648  */
    {
        int len;
        char *ddl;

        yyerrok;
        error_pop();

        len = (yyloc).last_offset - (yyloc).first_offset;
        ddl = xstrndup(parse->src + (yyloc).first_offset, len);

        (yyval.stmt) = stmt_new_ddl(ddl, &(yyloc));

        yylex_set_token(yyscanner, ';', &(yylsp[0]));
        yyclearin;
    }
#line 3166 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 163:
#line 999 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_blk((yyvsp[0].blk), &(yyloc));
    }
#line 3174 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 165:
#line 1007 "grammar.y" /* yacc.c:1648  */
    {
        if (is_tuple_exp((yyvsp[-2].exp))) {
            (yyval.exp) = (yyvsp[-2].exp);
        }
        else {
            (yyval.exp) = exp_new_tuple(vector_new(), &(yylsp[-2]));
            exp_add((yyval.exp)->u_tup.elem_exps, (yyvsp[-2].exp));
        }
        exp_add((yyval.exp)->u_tup.elem_exps, (yyvsp[0].exp));
    }
#line 3189 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 167:
#line 1022 "grammar.y" /* yacc.c:1648  */
    {
        int len;
        char *sql;

        yyerrok;
        error_pop();

        len = (yyloc).last_offset - (yyloc).first_offset;
        sql = xstrndup(parse->src + (yyloc).first_offset, len);

        (yyval.exp) = exp_new_sql((yyvsp[-2].sql), sql, &(yyloc));

        yylex_set_token(yyscanner, ';', &(yylsp[0]));
        yyclearin;
    }
#line 3209 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 168:
#line 1040 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_DELETE; }
#line 3215 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 169:
#line 1041 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_INSERT; }
#line 3221 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 170:
#line 1042 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_REPLACE; }
#line 3227 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 171:
#line 1043 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_QUERY; }
#line 3233 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 172:
#line 1044 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_UPDATE; }
#line 3239 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 174:
#line 1050 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3247 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 175:
#line 1054 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
        (yyval.exp)->u_init.is_outmost = true;
    }
#line 3256 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 176:
#line 1062 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_alloc((yyvsp[0].exp), &(yylsp[0]));
    }
#line 3264 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 177:
#line 1066 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-3].exp);

        if ((yyval.exp)->u_alloc.size_exps == NULL)
            (yyval.exp)->u_alloc.size_exps = vector_new();

        exp_add((yyval.exp)->u_alloc.size_exps, (yyvsp[-1].exp));
    }
#line 3277 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 178:
#line 1078 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_init((yyvsp[-2].vect), &(yyloc));
    }
#line 3285 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 179:
#line 1085 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3294 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 180:
#line 1090 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3303 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 183:
#line 1100 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3311 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 185:
#line 1108 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_ternary((yyvsp[-4].exp), (yyvsp[-2].exp), (yyvsp[0].exp), &(yyloc));
    }
#line 3319 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 187:
#line 1116 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3327 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 189:
#line 1124 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3335 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 191:
#line 1132 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3343 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 193:
#line 1140 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_XOR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3351 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 195:
#line 1148 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3359 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 197:
#line 1156 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3367 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 198:
#line 1162 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_EQ; }
#line 3373 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 199:
#line 1163 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NE; }
#line 3379 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 201:
#line 1169 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3387 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 202:
#line 1175 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LT; }
#line 3393 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 203:
#line 1176 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GT; }
#line 3399 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 204:
#line 1177 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LE; }
#line 3405 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 205:
#line 1178 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GE; }
#line 3411 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 207:
#line 1184 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3419 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 208:
#line 1190 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 3425 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 209:
#line 1191 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 3431 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 211:
#line 1197 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3439 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 212:
#line 1203 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 3445 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 213:
#line 1204 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 3451 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 215:
#line 1210 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3459 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 216:
#line 1216 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 3465 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 217:
#line 1217 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 3471 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 218:
#line 1218 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 3477 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 220:
#line 1224 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_cast((yyvsp[-2].type), (yyvsp[0].exp), &(yylsp[-2]));
    }
#line 3485 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 222:
#line 1232 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary((yyvsp[-1].op), true, (yyvsp[0].exp), &(yyloc));
    }
#line 3493 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 223:
#line 1236 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3501 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 224:
#line 1242 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_INC; }
#line 3507 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 225:
#line 1243 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DEC; }
#line 3513 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 226:
#line 1244 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NEG; }
#line 3519 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 227:
#line 1245 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NOT; }
#line 3525 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 229:
#line 1251 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_array((yyvsp[-3].exp), (yyvsp[-1].exp), &(yyloc));
    }
#line 3533 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 230:
#line 1255 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(false, (yyvsp[-3].exp), (yyvsp[-1].vect), &(yyloc));
    }
#line 3541 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 231:
#line 1259 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_access((yyvsp[-2].exp), exp_new_id((yyvsp[0].str), &(yylsp[0])), &(yyloc));
    }
#line 3549 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 232:
#line 1263 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_INC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3557 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 233:
#line 1267 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_DEC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3565 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 234:
#line 1273 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 3571 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 236:
#line 1279 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3580 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 237:
#line 1284 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3589 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 239:
#line 1293 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yyloc));
    }
#line 3597 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 240:
#line 1297 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-1].exp);
    }
#line 3605 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 241:
#line 1301 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(true, exp_new_id((yyvsp[-3].str), &(yylsp[-3])), (yyvsp[-1].vect), &(yylsp[-3]));
    }
#line 3613 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 242:
#line 1308 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_null(&(yyloc));
    }
#line 3621 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 243:
#line 1312 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(true, &(yyloc));
    }
#line 3629 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 244:
#line 1316 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(false, &(yyloc));
    }
#line 3637 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 245:
#line 1320 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNu64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3648 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 246:
#line 1327 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNo64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3659 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 247:
#line 1334 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNx64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3670 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 248:
#line 1341 "grammar.y" /* yacc.c:1648  */
    {
        double v;

        sscanf((yyvsp[0].str), "%lf", &v);
        (yyval.exp) = exp_new_lit_f64(v, &(yyloc));
    }
#line 3681 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 249:
#line 1348 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_str((yyvsp[0].str), &(yyloc));
    }
#line 3689 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 250:
#line 1354 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("contract"); }
#line 3695 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 251:
#line 1355 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("import"); }
#line 3701 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 252:
#line 1356 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("index"); }
#line 3707 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 253:
#line 1357 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("interface"); }
#line 3713 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 254:
#line 1358 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("table"); }
#line 3719 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 255:
#line 1359 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("view"); }
#line 3725 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 256:
#line 1364 "grammar.y" /* yacc.c:1648  */
    {
        if (strlen((yyvsp[0].str)) > NAME_MAX_LEN)
            ERROR(ERROR_TOO_LONG_ID, &(yylsp[0]), NAME_MAX_LEN, strlen((yyvsp[0].str)));

        (yyval.str) = (yyvsp[0].str);
    }
#line 3736 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;


#line 3740 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
      default: break;
    }
  /* User semantic actions sometimes alter yychar, and that requires
     that yytoken be updated with the new translation.  We take the
     approach of translating immediately before every use of yytoken.
     One alternative is translating here after every semantic action,
     but that translation would be missed if the semantic action invokes
     YYABORT, YYACCEPT, or YYERROR immediately after altering yychar or
     if it invokes YYBACKUP.  In the case of YYABORT or YYACCEPT, an
     incorrect destructor might then be invoked immediately.  In the
     case of YYERROR or YYBACKUP, subsequent parser actions might lead
     to an incorrect destructor call or verbose syntax error message
     before the lookahead is translated.  */
  YY_SYMBOL_PRINT ("-> $$ =", yyr1[yyn], &yyval, &yyloc);

  YYPOPSTACK (yylen);
  yylen = 0;
  YY_STACK_PRINT (yyss, yyssp);

  *++yyvsp = yyval;
  *++yylsp = yyloc;

  /* Now 'shift' the result of the reduction.  Determine what state
     that goes to, based on the state we popped back to and the rule
     number reduced by.  */

  yyn = yyr1[yyn];

  yystate = yypgoto[yyn - YYNTOKENS] + *yyssp;
  if (0 <= yystate && yystate <= YYLAST && yycheck[yystate] == *yyssp)
    yystate = yytable[yystate];
  else
    yystate = yydefgoto[yyn - YYNTOKENS];

  goto yynewstate;


/*--------------------------------------.
| yyerrlab -- here on detecting error.  |
`--------------------------------------*/
yyerrlab:
  /* Make sure we have latest lookahead translation.  See comments at
     user semantic actions for why this is necessary.  */
  yytoken = yychar == YYEMPTY ? YYEMPTY : YYTRANSLATE (yychar);

  /* If not already recovering from an error, report this error.  */
  if (!yyerrstatus)
    {
      ++yynerrs;
#if ! YYERROR_VERBOSE
      yyerror (&yylloc, parse, yyscanner, YY_("syntax error"));
#else
# define YYSYNTAX_ERROR yysyntax_error (&yymsg_alloc, &yymsg, \
                                        yyssp, yytoken)
      {
        char const *yymsgp = YY_("syntax error");
        int yysyntax_error_status;
        yysyntax_error_status = YYSYNTAX_ERROR;
        if (yysyntax_error_status == 0)
          yymsgp = yymsg;
        else if (yysyntax_error_status == 1)
          {
            if (yymsg != yymsgbuf)
              YYSTACK_FREE (yymsg);
            yymsg = (char *) YYSTACK_ALLOC (yymsg_alloc);
            if (!yymsg)
              {
                yymsg = yymsgbuf;
                yymsg_alloc = sizeof yymsgbuf;
                yysyntax_error_status = 2;
              }
            else
              {
                yysyntax_error_status = YYSYNTAX_ERROR;
                yymsgp = yymsg;
              }
          }
        yyerror (&yylloc, parse, yyscanner, yymsgp);
        if (yysyntax_error_status == 2)
          goto yyexhaustedlab;
      }
# undef YYSYNTAX_ERROR
#endif
    }

  yyerror_range[1] = yylloc;

  if (yyerrstatus == 3)
    {
      /* If just tried and failed to reuse lookahead token after an
         error, discard it.  */

      if (yychar <= YYEOF)
        {
          /* Return failure if at end of input.  */
          if (yychar == YYEOF)
            YYABORT;
        }
      else
        {
          yydestruct ("Error: discarding",
                      yytoken, &yylval, &yylloc, parse, yyscanner);
          yychar = YYEMPTY;
        }
    }

  /* Else will try to reuse lookahead token after shifting the error
     token.  */
  goto yyerrlab1;


/*---------------------------------------------------.
| yyerrorlab -- error raised explicitly by YYERROR.  |
`---------------------------------------------------*/
yyerrorlab:

  /* Pacify compilers like GCC when the user code never invokes
     YYERROR and the label yyerrorlab therefore never appears in user
     code.  */
  if (/*CONSTCOND*/ 0)
     goto yyerrorlab;

  /* Do not reclaim the symbols of the rule whose action triggered
     this YYERROR.  */
  YYPOPSTACK (yylen);
  yylen = 0;
  YY_STACK_PRINT (yyss, yyssp);
  yystate = *yyssp;
  goto yyerrlab1;


/*-------------------------------------------------------------.
| yyerrlab1 -- common code for both syntax error and YYERROR.  |
`-------------------------------------------------------------*/
yyerrlab1:
  yyerrstatus = 3;      /* Each real token shifted decrements this.  */

  for (;;)
    {
      yyn = yypact[yystate];
      if (!yypact_value_is_default (yyn))
        {
          yyn += YYTERROR;
          if (0 <= yyn && yyn <= YYLAST && yycheck[yyn] == YYTERROR)
            {
              yyn = yytable[yyn];
              if (0 < yyn)
                break;
            }
        }

      /* Pop the current state because it cannot handle the error token.  */
      if (yyssp == yyss)
        YYABORT;

      yyerror_range[1] = *yylsp;
      yydestruct ("Error: popping",
                  yystos[yystate], yyvsp, yylsp, parse, yyscanner);
      YYPOPSTACK (1);
      yystate = *yyssp;
      YY_STACK_PRINT (yyss, yyssp);
    }

  YY_IGNORE_MAYBE_UNINITIALIZED_BEGIN
  *++yyvsp = yylval;
  YY_IGNORE_MAYBE_UNINITIALIZED_END

  yyerror_range[2] = yylloc;
  /* Using YYLLOC is tempting, but would change the location of
     the lookahead.  YYLOC is available though.  */
  YYLLOC_DEFAULT (yyloc, yyerror_range, 2);
  *++yylsp = yyloc;

  /* Shift the error token.  */
  YY_SYMBOL_PRINT ("Shifting", yystos[yyn], yyvsp, yylsp);

  yystate = yyn;
  goto yynewstate;


/*-------------------------------------.
| yyacceptlab -- YYACCEPT comes here.  |
`-------------------------------------*/
yyacceptlab:
  yyresult = 0;
  goto yyreturn;

/*-----------------------------------.
| yyabortlab -- YYABORT comes here.  |
`-----------------------------------*/
yyabortlab:
  yyresult = 1;
  goto yyreturn;

#if !defined yyoverflow || YYERROR_VERBOSE
/*-------------------------------------------------.
| yyexhaustedlab -- memory exhaustion comes here.  |
`-------------------------------------------------*/
yyexhaustedlab:
  yyerror (&yylloc, parse, yyscanner, YY_("memory exhausted"));
  yyresult = 2;
  /* Fall through.  */
#endif

yyreturn:
  if (yychar != YYEMPTY)
    {
      /* Make sure we have latest lookahead translation.  See comments at
         user semantic actions for why this is necessary.  */
      yytoken = YYTRANSLATE (yychar);
      yydestruct ("Cleanup: discarding lookahead",
                  yytoken, &yylval, &yylloc, parse, yyscanner);
    }
  /* Do not reclaim the symbols of the rule whose action triggered
     this YYABORT or YYACCEPT.  */
  YYPOPSTACK (yylen);
  YY_STACK_PRINT (yyss, yyssp);
  while (yyssp != yyss)
    {
      yydestruct ("Cleanup: popping",
                  yystos[*yyssp], yyvsp, yylsp, parse, yyscanner);
      YYPOPSTACK (1);
    }
#ifndef yyoverflow
  if (yyss != yyssa)
    YYSTACK_FREE (yyss);
#endif
#if YYERROR_VERBOSE
  if (yymsg != yymsgbuf)
    YYSTACK_FREE (yymsg);
#endif
  return yyresult;
}
#line 1373 "grammar.y" /* yacc.c:1907  */


static void
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
