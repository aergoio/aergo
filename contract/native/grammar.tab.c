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
    K_ASSERT = 286,
    K_BOOL = 287,
    K_BREAK = 288,
    K_BYTE = 289,
    K_CASE = 290,
    K_CONST = 291,
    K_CONTINUE = 292,
    K_CONTRACT = 293,
    K_CREATE = 294,
    K_CURSOR = 295,
    K_DEFAULT = 296,
    K_DELETE = 297,
    K_DOUBLE = 298,
    K_DROP = 299,
    K_ELSE = 300,
    K_ENUM = 301,
    K_FALSE = 302,
    K_FLOAT = 303,
    K_FOR = 304,
    K_FUNC = 305,
    K_GOTO = 306,
    K_IF = 307,
    K_IMPLEMENTS = 308,
    K_IMPORT = 309,
    K_IN = 310,
    K_INDEX = 311,
    K_INSERT = 312,
    K_INT = 313,
    K_INT8 = 314,
    K_INT16 = 315,
    K_INT32 = 316,
    K_INT64 = 317,
    K_INT128 = 318,
    K_INTERFACE = 319,
    K_LIBRARY = 320,
    K_MAP = 321,
    K_NEW = 322,
    K_NULL = 323,
    K_PAYABLE = 324,
    K_PRAGMA = 325,
    K_PUBLIC = 326,
    K_REPLACE = 327,
    K_RETURN = 328,
    K_SELECT = 329,
    K_STRING = 330,
    K_STRUCT = 331,
    K_SWITCH = 332,
    K_TABLE = 333,
    K_TRUE = 334,
    K_TYPE = 335,
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
    meta_t *meta;

#line 235 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:355  */
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

#line 265 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:358  */

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
#define YYFINAL  25
/* YYLAST -- Last index in YYTABLE.  */
#define YYLAST   1955

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  106
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  93
/* YYNRULES -- Number of rules.  */
#define YYNRULES  265
/* YYNSTATES -- Number of states.  */
#define YYNSTATES  450

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
       0,   255,   255,   256,   260,   261,   262,   263,   264,   265,
     269,   273,   282,   293,   294,   301,   306,   311,   316,   321,
     326,   334,   335,   340,   348,   349,   357,   358,   370,   382,
     386,   392,   402,   403,   404,   405,   406,   407,   408,   409,
     410,   413,   414,   418,   419,   434,   438,   450,   451,   455,
     459,   473,   474,   478,   482,   490,   499,   511,   515,   523,
     528,   536,   540,   547,   549,   553,   569,   570,   574,   581,
     582,   586,   591,   599,   608,   615,   616,   617,   618,   622,
     623,   624,   628,   629,   658,   664,   676,   677,   681,   686,
     693,   698,   703,   708,   716,   723,   728,   736,   743,   752,
     761,   770,   782,   783,   784,   785,   786,   787,   788,   789,
     790,   791,   792,   796,   803,   807,   815,   819,   826,   827,
     828,   829,   830,   831,   832,   833,   834,   835,   839,   844,
     848,   855,   859,   864,   869,   877,   881,   885,   889,   893,
     897,   901,   905,   909,   917,   918,   922,   923,   927,   931,
     935,   943,   944,   948,   953,   961,   965,   969,   973,   977,
     984,  1003,  1004,  1005,  1006,  1007,  1008,  1009,  1013,  1020,
    1027,  1037,  1038,  1052,  1053,  1072,  1073,  1074,  1075,  1076,
    1080,  1081,  1085,  1093,  1097,  1109,  1116,  1121,  1129,  1130,
    1131,  1138,  1139,  1146,  1147,  1154,  1155,  1162,  1163,  1170,
    1171,  1178,  1179,  1186,  1187,  1194,  1195,  1199,  1200,  1207,
    1208,  1209,  1210,  1214,  1215,  1222,  1223,  1227,  1228,  1235,
    1236,  1240,  1241,  1248,  1249,  1250,  1254,  1255,  1262,  1263,
    1267,  1274,  1275,  1276,  1277,  1281,  1282,  1286,  1290,  1294,
    1298,  1305,  1306,  1310,  1315,  1323,  1324,  1328,  1332,  1339,
    1343,  1347,  1351,  1356,  1361,  1375,  1382,  1383,  1384,  1385,
    1386,  1387,  1388,  1389,  1393,  1400
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
  "\"--\"", "\"account\"", "\"alter\"", "\"assert\"", "\"bool\"",
  "\"break\"", "\"byte\"", "\"case\"", "\"const\"", "\"continue\"",
  "\"contract\"", "\"create\"", "\"cursor\"", "\"default\"", "\"delete\"",
  "\"double\"", "\"drop\"", "\"else\"", "\"enum\"", "\"false\"",
  "\"float\"", "\"for\"", "\"func\"", "\"goto\"", "\"if\"",
  "\"implements\"", "\"import\"", "\"in\"", "\"index\"", "\"insert\"",
  "\"int\"", "\"int8\"", "\"int16\"", "\"int32\"", "\"int64\"",
  "\"int128\"", "\"interface\"", "\"library\"", "\"map\"", "\"new\"",
  "\"null\"", "\"payable\"", "\"pragma\"", "\"public\"", "\"replace\"",
  "\"return\"", "\"select\"", "\"string\"", "\"struct\"", "\"switch\"",
  "\"table\"", "\"true\"", "\"type\"", "\"update\"", "\"view\"", "'!'",
  "'|'", "'^'", "'&'", "'<'", "'>'", "'+'", "'-'", "'*'", "'/'", "'%'",
  "'('", "')'", "'.'", "'{'", "'}'", "';'", "'='", "','", "'['", "']'",
  "':'", "'?'", "$accept", "root", "component", "import", "contract",
  "impl_opt", "contract_body", "variable", "var_qual", "var_decl",
  "var_spec", "var_type", "prim_type", "declarator_list", "declarator",
  "size_opt", "var_init", "compound", "struct", "field_list",
  "enumeration", "enum_list", "enumerator", "comma_opt", "function",
  "func_spec", "ctor_spec", "param_list_opt", "param_list", "param_decl",
  "udf_spec", "modifier_opt", "return_opt", "return_list", "return_decl",
  "block", "blk_decl", "interface", "interface_body", "library",
  "library_body", "statement", "empty_stmt", "exp_stmt", "assign_stmt",
  "assign_op", "label_stmt", "if_stmt", "loop_stmt", "init_stmt",
  "cond_exp", "switch_stmt", "switch_blk", "case_blk", "jump_stmt",
  "ddl_stmt", "ddl_prefix", "blk_stmt", "pragma_stmt", "expression",
  "sql_exp", "sql_prefix", "new_exp", "alloc_exp", "initializer",
  "elem_list", "init_elem", "ternary_exp", "or_exp", "and_exp",
  "bit_or_exp", "bit_xor_exp", "bit_and_exp", "eq_exp", "eq_op", "cmp_exp",
  "cmp_op", "add_exp", "add_op", "shift_exp", "shift_op", "mul_exp",
  "mul_op", "cast_exp", "unary_exp", "unary_op", "post_exp",
  "arg_list_opt", "arg_list", "prim_exp", "literal", "non_reserved_token",
  "identifier", YY_NULLPTR
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

#define YYPACT_NINF -270

#define yypact_value_is_default(Yystate) \
  (!!((Yystate) == (-270)))

#define YYTABLE_NINF -31

#define yytable_value_is_error(Yytable_value) \
  0

  /* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
     STATE-NUM.  */
static const yytype_int16 yypact[] =
{
     156,   152,    33,   152,   152,    19,  -270,  -270,  -270,  -270,
    -270,  -270,  -270,  -270,  -270,  -270,  -270,  -270,  -270,  -270,
    -270,    -2,  -270,   -11,    14,  -270,  -270,  -270,  -270,   152,
      54,     8,     8,  -270,   932,  -270,    -8,    17,   118,   -48,
     -73,   -26,  -270,  -270,  -270,  1873,  -270,   613,  -270,  -270,
    -270,  -270,  -270,  -270,   102,  1828,  -270,   712,  -270,   977,
    -270,  -270,  -270,    30,   825,  -270,  -270,  -270,  -270,  -270,
     116,  -270,  -270,   133,  -270,  -270,   152,  -270,   140,   384,
     227,  -270,  -270,    -7,  -270,   152,  -270,   150,   157,  1873,
    -270,   162,   185,  -270,  -270,  -270,  -270,  -270,  1448,   168,
     169,   167,  -270,  -270,  1873,   178,  -270,   176,  -270,  -270,
    -270,  -270,  -270,  -270,   198,   180,  1538,   183,   -18,   179,
    -270,    92,  -270,     7,   152,    16,  -270,  1130,  -270,   254,
    -270,   866,  -270,    15,  -270,  -270,  -270,  1606,  -270,  1358,
    -270,  -270,  -270,  -270,  -270,   483,  -270,  -270,  -270,  -270,
    -270,   244,  -270,  -270,  -270,  -270,   289,  -270,  -270,    20,
    -270,   293,  -270,  -270,    10,   274,   212,   213,   214,   117,
     136,    87,   172,    74,  -270,   341,  1606,    38,  -270,  -270,
      93,   200,   294,  -270,  -270,   152,   202,  -270,   208,    53,
    -270,  -270,  -270,  -270,   152,  1538,   152,   215,   207,  -270,
    1873,  -270,  -270,  -270,   152,    23,  -270,  -270,  -270,  -270,
    -270,  -270,  -270,  -270,   211,   776,  -270,   217,   221,  1448,
    1063,  -270,   210,  -270,   219,   226,  -270,    88,   223,  1448,
     209,  -270,  1448,  -270,   216,   -65,  -270,  -270,  -270,  -270,
     -23,   224,  -270,  1448,  1448,   225,  1538,  1448,  1538,  1538,
    1538,  1538,  -270,  -270,  1538,  -270,  -270,  -270,  -270,  1538,
    -270,  -270,  1538,  -270,  -270,  1538,  -270,  -270,  -270,  1538,
    -270,  -270,  -270,  -270,  -270,  -270,  -270,  -270,  -270,  -270,
    1448,  -270,  -270,  -270,  1674,   152,  1538,   681,  -270,   228,
     230,  -270,   222,  1873,  1873,  -270,  1448,   167,   231,    87,
     167,  -270,  1873,   240,   219,  -270,  -270,   250,    36,  -270,
    -270,   250,   -20,   241,   846,  -270,  -270,    84,  -270,   236,
    -270,  -270,   234,  1538,  1674,  1538,  -270,  -270,   127,  -270,
    -270,   582,   239,  1606,  -270,   232,  -270,  -270,   110,  -270,
    -270,   274,   -42,   212,   213,   214,   117,   136,    87,   172,
      74,  -270,   137,  -270,   251,   246,  -270,   242,  -270,   239,
    -270,   152,   263,  1538,   267,   264,  1022,  -270,  -270,  -270,
    1740,  -270,  1172,   265,  1776,  1265,  1776,   116,   116,  1063,
     268,  1063,   -64,   272,    27,   271,  -270,  -270,  -270,  1448,
    -270,  1538,  -270,  -270,  1674,  -270,  -270,  -270,   303,  -270,
    -270,  -270,   275,  1873,  -270,  -270,   269,   273,   116,   130,
    -270,    43,   116,   131,    67,  -270,  -270,  -270,  -270,  -270,
    -270,  -270,  -270,  1538,  -270,   142,  -270,  -270,  -270,   151,
    1873,  1538,  -270,   116,   116,  -270,   116,   116,   112,   116,
    -270,   273,   270,  -270,  -270,  -270,  -270,  -270,  -270,  -270
};

  /* YYDEFACT[STATE-NUM] -- Default reduction number in state STATE-NUM.
     Performed when YYTABLE does not specify something else to do.  Zero
     means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       0,     0,     0,     0,     0,     0,     2,     4,     5,     6,
       3,   264,   256,   257,   258,   259,   260,   261,   262,   263,
     265,    13,    10,     0,     0,     1,     7,     8,     9,     0,
       0,    75,    75,    14,    75,    77,    76,     0,     0,    75,
       0,    75,    32,    33,    34,     0,    42,     0,    40,    35,
      36,    37,    38,    39,     0,    76,    41,     0,    11,    75,
      15,    21,    24,     0,     0,    29,    16,    51,    52,    17,
       0,    66,    67,    30,    78,    95,     0,    94,     0,     0,
       0,    99,    97,     0,    25,     0,    30,     0,     0,     0,
      22,     0,     0,    12,    18,    19,    20,    26,     0,     0,
      28,    43,    45,    65,    69,     0,    96,     0,   254,   252,
     253,   255,   231,   232,     0,     0,     0,     0,     0,     0,
     175,     0,   251,     0,     0,     0,   176,     0,   249,     0,
     177,     0,   178,     0,   250,   179,   234,     0,   233,     0,
      86,   113,    88,    89,   168,     0,    90,   102,   103,   104,
     105,   106,   107,   108,   109,   110,     0,   111,   112,     0,
     171,     0,   173,   180,   191,   193,   195,   197,   199,   201,
     203,   207,   213,   217,   221,   226,     0,   228,   235,   245,
     246,     0,     0,   101,    58,     0,     0,    54,     0,     0,
      49,   226,   246,    23,     0,    47,     0,     0,    70,    71,
      69,   115,   162,   156,     0,     0,   155,   164,   161,   166,
     130,   165,   163,   167,     0,     0,   135,     0,     0,     0,
       0,   183,   181,   182,    30,     0,   157,     0,     0,     0,
       0,   148,     0,   230,     0,     0,    87,    91,    92,    93,
       0,     0,   114,     0,     0,     0,     0,     0,     0,     0,
       0,     0,   205,   206,     0,   211,   212,   209,   210,     0,
     215,   216,     0,   219,   220,     0,   223,   224,   225,     0,
     118,   119,   120,   121,   122,   123,   124,   125,   126,   127,
       0,   229,   239,   240,   241,     0,     0,     0,    98,     0,
      63,    59,    61,     0,     0,    27,     0,    44,     0,    48,
      73,    68,     0,     0,     0,   129,   143,     0,     0,   144,
     145,     0,     0,   180,   246,   159,   134,     0,   189,    63,
     186,   188,   246,     0,   241,     0,   158,   150,     0,   151,
     153,     0,     0,     0,   247,     0,   133,   160,     0,   172,
     174,   194,     0,   196,   198,   200,   202,   204,   208,   214,
     218,   222,     0,   243,     0,   242,   238,     0,   128,   246,
     100,    64,     0,     0,     0,     0,     0,    50,    46,    72,
      79,   146,     0,     0,     0,     0,     0,     0,     0,    64,
       0,     0,     0,     0,     0,     0,   152,   154,   227,     0,
     116,     0,   117,   237,     0,   236,    60,    57,    62,    31,
      55,    53,     0,     0,    84,    74,    81,    82,     0,     0,
     147,     0,     0,     0,     0,   136,   131,   187,   185,   190,
     184,   248,   169,     0,   149,     0,   192,   244,    56,     0,
       0,    47,   139,     0,     0,   137,     0,     0,     0,     0,
      80,    83,     0,   140,   142,   138,   141,   170,   132,    85
};

  /* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -270,  -270,  -270,   366,   372,  -270,  -270,   320,   -37,   -34,
    -202,   -22,   243,  -270,   -96,   -47,  -270,   -39,  -270,  -270,
    -270,  -270,    22,    69,   327,  -270,  -270,   193,  -270,    94,
     141,  -270,  -270,    -9,   -35,   -36,  -270,   394,  -270,  -270,
    -270,  -131,   189,  -270,   190,  -270,   177,  -270,  -270,  -270,
      95,  -270,    24,  -270,  -270,  -270,  -270,  -270,  -270,  -130,
     -91,  -270,  -269,  -270,   283,  -270,  -113,  -193,    64,   186,
     191,   188,   184,  -110,  -270,   199,  -270,  -190,  -270,   206,
    -270,   195,  -270,   201,   -69,  -270,  -125,   145,  -270,  -270,
    -270,  -270,    -1
};

  /* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     5,     6,     7,     8,    30,    59,    60,    61,    62,
      63,    85,    65,   100,   101,   298,   189,    66,    67,   366,
      68,   290,   291,   362,    69,    70,    71,   197,   198,   199,
      72,    38,   405,   406,   407,   144,   145,     9,    39,    10,
      41,   146,   147,   148,   149,   280,   150,   151,   152,   311,
     372,   153,   231,   331,   154,   155,   156,   157,   158,   159,
     160,   161,   162,   222,   318,   319,   320,   163,   164,   165,
     166,   167,   168,   169,   254,   170,   259,   171,   262,   172,
     265,   173,   269,   174,   191,   176,   177,   354,   355,   178,
     179,    20,   192
};

  /* YYTABLE[YYPACT[STATE-NUM]] -- What to do in state STATE-NUM.  If
     positive, shift that token.  If negative, reduce the rule whose
     number is the opposite.  If YYTABLE_NINF, syntax error.  */
static const yytype_int16 yytable[] =
{
      21,   227,    23,    24,    81,   299,   205,   190,   214,   235,
     175,    84,    64,   308,   239,   353,   228,   218,    90,    25,
      95,    35,   313,    36,    79,   260,   261,   321,    33,   335,
     334,    80,   246,    73,   103,   376,   244,    64,   207,   420,
     143,    22,   142,    35,    86,    36,    88,   183,   252,   253,
      77,    29,   252,   253,    86,   353,    92,     1,    73,   244,
     208,    74,   391,   102,   209,   282,   283,   186,   233,   348,
     282,   283,    82,     2,    79,   105,   175,    35,   180,    36,
     243,   244,   196,     3,   102,   312,    31,   216,    86,   317,
      79,   374,   365,   357,   282,   283,   -30,   182,   297,   328,
     300,   215,   235,    86,    79,   221,   238,   281,   237,   229,
     219,    32,   230,   338,   373,   247,    75,   342,   373,   242,
     243,   244,   422,   217,   -30,   427,   224,   305,   423,    97,
      98,   -30,   284,   382,   285,    97,    98,   284,   434,   285,
     286,   346,   252,   253,   180,   286,   175,   -30,   211,   -30,
     352,    34,   295,   339,   296,    11,   358,   -30,   -30,   255,
     256,   284,   437,   285,   402,   266,   267,   268,    76,   286,
     212,   -30,    37,    40,   213,   -30,   260,   261,   196,   378,
      78,   307,    83,    12,   292,   244,   321,   326,   321,   244,
      13,   263,   264,   102,     1,   102,    89,   287,   426,    86,
     387,   260,   261,   304,   336,   367,    14,   447,    15,   390,
       2,   244,    11,    79,   314,   384,    16,    17,   175,   322,
       3,     4,   385,   257,   258,   433,   436,   104,   244,   332,
      18,   244,   244,   438,    19,   181,   392,   439,   244,   106,
      12,   299,   409,   244,   116,   413,   440,    13,   184,   411,
     119,   414,   430,    11,   185,   108,   109,   110,   111,   425,
     187,   188,   175,    14,   388,    15,   417,   193,   419,   195,
     194,   364,   200,    16,    17,   201,   202,   112,   113,   203,
     196,    12,   206,   210,   356,   225,   359,    18,    13,   240,
     241,    19,    86,    86,   245,   248,   249,   122,   250,   288,
     251,    86,   289,   293,    14,   294,    15,   329,   302,   306,
     301,   333,   323,   324,    16,    17,   315,   204,   128,   316,
     325,   327,   363,   337,   340,   246,   389,   360,    18,   134,
     359,   361,    19,   136,   368,   370,   377,   379,   381,   137,
     138,   415,   416,   287,   139,   395,   393,   394,   404,   371,
     270,   271,   272,   273,   274,   275,   276,   277,   278,   279,
     292,   397,   399,   400,   410,    86,   418,   421,   230,    86,
     430,    26,   432,   449,   428,   431,   435,    27,   322,    94,
     322,   404,   234,   396,   442,   107,    96,    11,   380,   108,
     109,   110,   111,   303,   429,   441,   369,   443,   444,    28,
     445,   446,    86,   448,   309,   310,   375,   330,   404,   424,
     223,   112,   113,    42,   114,    12,    43,   115,    44,   116,
      45,   117,    13,   118,    46,   119,   120,   398,   121,    86,
      47,   122,   341,   123,   345,   124,   125,   344,    14,   343,
      15,   126,    48,    49,    50,    51,    52,    53,    16,    17,
      54,   127,   128,   347,   129,     0,   130,   131,   132,    56,
     350,   133,    18,   134,    57,   135,    19,   136,   349,   383,
     351,     0,     0,   137,   138,     0,     0,     0,   139,     0,
       0,    79,   140,   141,   107,     0,    11,     0,   108,   109,
     110,   111,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     112,   113,    42,   114,    12,    43,   115,    44,   116,    45,
     117,    13,   118,    46,   119,   120,     0,   121,     0,    47,
     122,     0,   123,     0,   124,   125,     0,    14,     0,    15,
     126,    48,    49,    50,    51,    52,    53,    16,    17,    54,
     127,   128,     0,   129,     0,   130,   131,   132,    56,     0,
     133,    18,   134,    57,   135,    19,   136,     0,     0,     0,
       0,     0,   137,   138,     0,     0,     0,   139,     0,     0,
      79,   236,   141,   107,     0,    11,     0,   108,   109,   110,
     111,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,   112,
     113,     0,   114,    12,    87,   115,    11,   116,     0,   117,
      13,   118,     0,   119,   120,     0,   121,     0,     0,   122,
       0,   123,     0,   124,   125,     0,    14,     0,    15,   126,
       0,     0,     0,     0,    12,     0,    16,    17,     0,   127,
     128,    13,   129,     0,   130,   131,   132,     0,     0,   133,
      18,   134,     0,   135,    19,   136,     0,    14,     0,    15,
       0,   137,   138,     0,     0,     0,   139,    16,    17,    79,
     386,   141,   107,     0,    11,     0,   108,   109,   110,   111,
       0,    18,     0,     0,     0,    19,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,   112,   113,
       0,   114,    12,    91,   115,    11,   116,     0,   117,    13,
     118,     0,   119,   120,     0,   121,     0,     0,   122,     0,
     123,     0,   124,   125,     0,    14,     0,    15,   126,     0,
       0,     0,     0,    12,     0,    16,    17,     0,   127,   128,
      13,   129,     0,   130,   131,   132,     0,     0,   133,    18,
     134,     0,   135,    19,   136,     0,    14,     0,    15,     0,
     137,   138,     0,     0,     0,   139,    16,    17,    79,    11,
     141,   108,   109,   110,   111,     0,     0,     0,     0,     0,
      18,     0,     0,     0,    19,     0,     0,     0,     0,     0,
       0,     0,     0,   112,   113,    42,     0,    12,    43,     0,
      44,     0,     0,     0,    13,     0,    46,     0,   120,     0,
       0,     0,     0,   122,     0,     0,    99,     0,    11,     0,
      14,     0,    15,   126,    48,    49,    50,    51,    52,    53,
      16,    17,    54,   127,   128,     0,     0,     0,   130,   -30,
     132,    56,     0,     0,    18,   134,    12,   135,    19,   136,
       0,     0,     0,    13,     0,   137,   138,     0,     0,    11,
     139,   108,   109,   110,   111,   141,     0,   -30,     0,    14,
       0,    15,     0,     0,   -30,     0,     0,     0,     0,    16,
      17,     0,     0,   112,   113,     0,     0,    12,     0,     0,
     -30,     0,   -30,    18,    13,     0,     0,    19,   120,     0,
     -30,   -30,     0,   122,     0,     0,     0,     0,     0,     0,
      14,     0,    15,   126,   -30,     0,     0,     0,   -30,     0,
      16,    17,     0,   127,   128,    11,     0,     0,   130,     0,
     132,     0,     0,     0,    18,   134,     0,   135,    19,   136,
       0,     0,     0,     0,     0,   137,   138,     0,     0,     0,
     139,    42,     0,    12,    43,   226,    44,     0,    45,     0,
      13,     0,    46,     0,     0,     0,     0,     0,    47,     0,
      11,     0,     0,     0,     0,     0,    14,     0,    15,     0,
      48,    49,    50,    51,    52,    53,    16,    17,    54,     0,
       0,    35,     0,    55,     0,     0,    42,    56,    12,    43,
      18,    44,    57,    45,    19,    13,     0,    46,     0,     0,
       0,     0,     0,    47,     0,    11,     0,     0,     0,     0,
      58,    14,     0,    15,     0,    48,    49,    50,    51,    52,
      53,    16,    17,    54,     0,     0,    35,     0,    55,     0,
       0,    42,    56,    12,    43,    18,    44,    57,     0,    19,
      13,     0,    46,     0,     0,     0,    11,     0,   108,   109,
     110,   111,     0,     0,     0,    93,    14,     0,    15,     0,
      48,    49,    50,    51,    52,    53,    16,    17,    54,     0,
     112,   113,     0,     0,    12,     0,     0,    56,     0,     0,
      18,    13,     0,     0,    19,     0,     0,     0,     0,     0,
     122,     0,     0,     0,     0,     0,     0,    14,     0,    15,
     401,     0,     0,     0,     0,     0,     0,    16,    17,     0,
     204,   128,     0,    11,     0,     0,     0,     0,     0,     0,
       0,    18,   134,     0,     0,    19,   136,     0,     0,     0,
       0,     0,   137,   138,     0,     0,     0,   139,     0,    42,
     220,    12,    43,     0,    44,     0,     0,     0,    13,     0,
      46,     0,     0,     0,     0,    11,     0,   108,   109,   110,
     111,     0,     0,     0,    14,     0,    15,     0,    48,    49,
      50,    51,    52,    53,    16,    17,    54,     0,     0,   112,
     113,     0,     0,    12,     0,    56,     0,     0,    18,     0,
      13,     0,    19,     0,   120,     0,     0,     0,     0,   122,
       0,     0,     0,     0,     0,     0,    14,   220,    15,   126,
       0,     0,     0,     0,     0,     0,    16,    17,     0,   127,
     128,     0,     0,     0,   130,     0,   132,     0,     0,     0,
      18,   134,     0,   135,    19,   136,     0,     0,     0,     0,
       0,   137,   138,     0,     0,     0,   139,   408,    11,     0,
     108,   109,   110,   111,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,   112,   113,     0,     0,    12,     0,     0,     0,
       0,     0,     0,    13,     0,     0,     0,   120,     0,     0,
       0,     0,   122,     0,     0,     0,     0,     0,     0,    14,
       0,    15,   126,     0,     0,     0,     0,     0,     0,    16,
      17,     0,   127,   128,     0,     0,     0,   130,     0,   132,
       0,     0,     0,    18,   134,     0,   135,    19,   136,     0,
       0,     0,     0,     0,   137,   138,     0,     0,     0,   139,
     412,    11,     0,   108,   109,   110,   111,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   112,   113,    42,     0,    12,
      43,     0,    44,     0,     0,     0,    13,     0,    46,     0,
     120,     0,     0,     0,     0,   122,     0,     0,     0,     0,
       0,     0,    14,     0,    15,   126,    48,    49,    50,    51,
      52,    53,    16,    17,     0,   127,   128,     0,     0,     0,
     130,     0,   132,    56,     0,     0,    18,   134,     0,   135,
      19,   136,     0,     0,     0,     0,     0,   137,   138,     0,
       0,    11,   139,   108,   109,   110,   111,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   112,   113,     0,     0,    12,
       0,     0,     0,     0,     0,     0,    13,     0,     0,     0,
     120,     0,     0,     0,     0,   122,     0,     0,     0,     0,
       0,     0,    14,     0,    15,   126,     0,     0,     0,     0,
       0,     0,    16,    17,     0,   127,   128,     0,     0,     0,
     130,     0,   132,     0,     0,     0,    18,   134,     0,   135,
      19,   136,     0,     0,     0,     0,     0,   137,   138,     0,
       0,    11,   139,   108,   109,   110,   111,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   112,   113,     0,     0,    12,
       0,     0,     0,     0,     0,     0,    13,     0,     0,     0,
       0,     0,     0,     0,     0,   122,     0,     0,     0,     0,
       0,     0,    14,     0,    15,     0,     0,     0,     0,     0,
       0,     0,    16,    17,     0,   204,   128,     0,     0,    11,
       0,   108,   109,   110,   111,     0,    18,   134,     0,     0,
      19,   136,     0,     0,     0,     0,     0,   137,   138,     0,
       0,     0,   139,   112,   113,     0,     0,    12,     0,     0,
       0,     0,     0,     0,    13,     0,     0,     0,     0,     0,
       0,     0,     0,   122,     0,     0,     0,     0,     0,     0,
      14,     0,    15,     0,     0,     0,     0,     0,     0,     0,
      16,    17,     0,   204,   128,     0,     0,    11,     0,   108,
     109,   110,   111,     0,    18,   134,     0,     0,    19,   136,
       0,     0,     0,     0,     0,   137,   138,     0,     0,     0,
     232,   112,   113,     0,     0,    12,     0,     0,     0,     0,
       0,     0,    13,     0,     0,     0,     0,     0,     0,     0,
       0,   122,     0,     0,     0,     0,     0,     0,    14,     0,
      15,     0,     0,     0,     0,     0,     0,     0,    16,    17,
       0,   127,   128,    11,     0,     0,     0,     0,     0,     0,
       0,     0,    18,   134,     0,     0,    19,   136,     0,     0,
       0,     0,     0,   137,   138,     0,     0,     0,   139,    42,
       0,    12,    43,     0,    44,     0,     0,     0,    13,    11,
      46,   108,   109,   110,   111,     0,     0,     0,     0,     0,
       0,     0,     0,     0,    14,     0,    15,     0,    48,    49,
      50,    51,    52,    53,    16,    17,    54,    12,     0,     0,
       0,     0,     0,     0,    13,    56,     0,     0,    18,     0,
       0,     0,    19,   122,     0,     0,     0,     0,     0,     0,
      14,    11,    15,     0,   403,     0,     0,     0,     0,     0,
      16,    17,     0,   204,   128,     0,     0,     0,     0,     0,
       0,     0,     0,     0,    18,   134,     0,    42,    19,    12,
      43,     0,    44,     0,    45,     0,    13,     0,    46,     0,
     232,     0,     0,     0,     0,     0,    11,     0,     0,     0,
       0,     0,    14,     0,    15,     0,    48,    49,    50,    51,
      52,    53,    16,    17,    54,     0,     0,    74,     0,     0,
       0,     0,    42,    56,    12,    43,    18,    44,     0,     0,
      19,    13,     0,    46,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,    14,     0,    15,
       0,    48,    49,    50,    51,    52,    53,    16,    17,    54,
       0,     0,     0,     0,     0,     0,     0,     0,    56,     0,
       0,    18,     0,     0,     0,    19
};

static const yytype_int16 yycheck[] =
{
       1,   131,     3,     4,    40,   195,   116,    98,     1,   139,
      79,    45,    34,   215,   145,   284,     1,     1,    55,     0,
      59,    69,   215,    71,    97,    89,    90,   220,    29,    52,
      95,   104,    22,    34,    70,    55,   101,    59,    56,   103,
      79,     8,    79,    69,    45,    71,    47,    83,    25,    26,
      98,    53,    25,    26,    55,   324,    57,    38,    59,   101,
      78,    69,   104,    64,    82,    27,    28,    89,   137,   259,
      27,    28,    98,    54,    97,    76,   145,    69,    79,    71,
     100,   101,   104,    64,    85,   215,    97,   123,    89,   219,
      97,    55,   294,   286,    27,    28,     3,   104,   194,   229,
     196,    94,   232,   104,    97,   127,   145,   176,   145,    94,
      94,    97,    97,   243,   307,   105,    99,   247,   311,    99,
     100,   101,    95,   124,    31,   394,   127,   104,   101,    99,
     100,    38,    94,   323,    96,    99,   100,    94,    95,    96,
     102,   251,    25,    26,   145,   102,   215,    54,    56,    56,
     280,    97,    99,   244,   101,     3,   287,    64,    65,    23,
      24,    94,    95,    96,   366,    91,    92,    93,    50,   102,
      78,    78,    31,    32,    82,    82,    89,    90,   200,    95,
      39,   215,    41,    31,   185,   101,   379,    99,   381,   101,
      38,    19,    20,   194,    38,   196,    94,   104,   391,   200,
     331,    89,    90,   204,   240,   296,    54,    95,    56,    99,
      54,   101,     3,    97,   215,   325,    64,    65,   287,   220,
      64,    65,    95,    87,    88,    95,    95,    94,   101,   230,
      78,   101,   101,   423,    82,     8,    99,    95,   101,    99,
      31,   431,   372,   101,    35,   375,    95,    38,    98,   374,
      41,   376,   101,     3,    97,     5,     6,     7,     8,   389,
      98,    76,   331,    54,   333,    56,   379,    99,   381,   102,
     101,   293,    94,    64,    65,    99,    78,    27,    28,    99,
     302,    31,    99,   104,   285,    31,   287,    78,    38,    45,
       1,    82,   293,   294,     1,    21,    84,    47,    85,    99,
      86,   302,     8,   101,    54,    97,    56,    98,   101,    98,
      95,    95,   102,    94,    64,    65,    99,    67,    68,    98,
      94,    98,   100,    99,    99,    22,    94,    99,    78,    79,
     331,   101,    82,    83,   103,    95,    95,   101,   104,    89,
      90,   377,   378,   104,    94,   103,    95,   101,   370,    99,
       9,    10,    11,    12,    13,    14,    15,    16,    17,    18,
     361,    98,    95,    99,    99,   366,    98,    95,    97,   370,
     101,     5,   408,   103,    99,   102,   412,     5,   379,    59,
     381,   403,   139,   361,   431,     1,    59,     3,   319,     5,
       6,     7,     8,   200,   403,   430,   302,   433,   434,     5,
     436,   437,   403,   439,   215,   215,   311,   230,   430,   385,
     127,    27,    28,    29,    30,    31,    32,    33,    34,    35,
      36,    37,    38,    39,    40,    41,    42,   363,    44,   430,
      46,    47,   246,    49,   250,    51,    52,   249,    54,   248,
      56,    57,    58,    59,    60,    61,    62,    63,    64,    65,
      66,    67,    68,   254,    70,    -1,    72,    73,    74,    75,
     265,    77,    78,    79,    80,    81,    82,    83,   262,   324,
     269,    -1,    -1,    89,    90,    -1,    -1,    -1,    94,    -1,
      -1,    97,    98,    99,     1,    -1,     3,    -1,     5,     6,
       7,     8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      27,    28,    29,    30,    31,    32,    33,    34,    35,    36,
      37,    38,    39,    40,    41,    42,    -1,    44,    -1,    46,
      47,    -1,    49,    -1,    51,    52,    -1,    54,    -1,    56,
      57,    58,    59,    60,    61,    62,    63,    64,    65,    66,
      67,    68,    -1,    70,    -1,    72,    73,    74,    75,    -1,
      77,    78,    79,    80,    81,    82,    83,    -1,    -1,    -1,
      -1,    -1,    89,    90,    -1,    -1,    -1,    94,    -1,    -1,
      97,    98,    99,     1,    -1,     3,    -1,     5,     6,     7,
       8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,
      28,    -1,    30,    31,     1,    33,     3,    35,    -1,    37,
      38,    39,    -1,    41,    42,    -1,    44,    -1,    -1,    47,
      -1,    49,    -1,    51,    52,    -1,    54,    -1,    56,    57,
      -1,    -1,    -1,    -1,    31,    -1,    64,    65,    -1,    67,
      68,    38,    70,    -1,    72,    73,    74,    -1,    -1,    77,
      78,    79,    -1,    81,    82,    83,    -1,    54,    -1,    56,
      -1,    89,    90,    -1,    -1,    -1,    94,    64,    65,    97,
      98,    99,     1,    -1,     3,    -1,     5,     6,     7,     8,
      -1,    78,    -1,    -1,    -1,    82,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,    28,
      -1,    30,    31,     1,    33,     3,    35,    -1,    37,    38,
      39,    -1,    41,    42,    -1,    44,    -1,    -1,    47,    -1,
      49,    -1,    51,    52,    -1,    54,    -1,    56,    57,    -1,
      -1,    -1,    -1,    31,    -1,    64,    65,    -1,    67,    68,
      38,    70,    -1,    72,    73,    74,    -1,    -1,    77,    78,
      79,    -1,    81,    82,    83,    -1,    54,    -1,    56,    -1,
      89,    90,    -1,    -1,    -1,    94,    64,    65,    97,     3,
      99,     5,     6,     7,     8,    -1,    -1,    -1,    -1,    -1,
      78,    -1,    -1,    -1,    82,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    27,    28,    29,    -1,    31,    32,    -1,
      34,    -1,    -1,    -1,    38,    -1,    40,    -1,    42,    -1,
      -1,    -1,    -1,    47,    -1,    -1,     1,    -1,     3,    -1,
      54,    -1,    56,    57,    58,    59,    60,    61,    62,    63,
      64,    65,    66,    67,    68,    -1,    -1,    -1,    72,     3,
      74,    75,    -1,    -1,    78,    79,    31,    81,    82,    83,
      -1,    -1,    -1,    38,    -1,    89,    90,    -1,    -1,     3,
      94,     5,     6,     7,     8,    99,    -1,    31,    -1,    54,
      -1,    56,    -1,    -1,    38,    -1,    -1,    -1,    -1,    64,
      65,    -1,    -1,    27,    28,    -1,    -1,    31,    -1,    -1,
      54,    -1,    56,    78,    38,    -1,    -1,    82,    42,    -1,
      64,    65,    -1,    47,    -1,    -1,    -1,    -1,    -1,    -1,
      54,    -1,    56,    57,    78,    -1,    -1,    -1,    82,    -1,
      64,    65,    -1,    67,    68,     3,    -1,    -1,    72,    -1,
      74,    -1,    -1,    -1,    78,    79,    -1,    81,    82,    83,
      -1,    -1,    -1,    -1,    -1,    89,    90,    -1,    -1,    -1,
      94,    29,    -1,    31,    32,    99,    34,    -1,    36,    -1,
      38,    -1,    40,    -1,    -1,    -1,    -1,    -1,    46,    -1,
       3,    -1,    -1,    -1,    -1,    -1,    54,    -1,    56,    -1,
      58,    59,    60,    61,    62,    63,    64,    65,    66,    -1,
      -1,    69,    -1,    71,    -1,    -1,    29,    75,    31,    32,
      78,    34,    80,    36,    82,    38,    -1,    40,    -1,    -1,
      -1,    -1,    -1,    46,    -1,     3,    -1,    -1,    -1,    -1,
      98,    54,    -1,    56,    -1,    58,    59,    60,    61,    62,
      63,    64,    65,    66,    -1,    -1,    69,    -1,    71,    -1,
      -1,    29,    75,    31,    32,    78,    34,    80,    -1,    82,
      38,    -1,    40,    -1,    -1,    -1,     3,    -1,     5,     6,
       7,     8,    -1,    -1,    -1,    98,    54,    -1,    56,    -1,
      58,    59,    60,    61,    62,    63,    64,    65,    66,    -1,
      27,    28,    -1,    -1,    31,    -1,    -1,    75,    -1,    -1,
      78,    38,    -1,    -1,    82,    -1,    -1,    -1,    -1,    -1,
      47,    -1,    -1,    -1,    -1,    -1,    -1,    54,    -1,    56,
      98,    -1,    -1,    -1,    -1,    -1,    -1,    64,    65,    -1,
      67,    68,    -1,     3,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    78,    79,    -1,    -1,    82,    83,    -1,    -1,    -1,
      -1,    -1,    89,    90,    -1,    -1,    -1,    94,    -1,    29,
      97,    31,    32,    -1,    34,    -1,    -1,    -1,    38,    -1,
      40,    -1,    -1,    -1,    -1,     3,    -1,     5,     6,     7,
       8,    -1,    -1,    -1,    54,    -1,    56,    -1,    58,    59,
      60,    61,    62,    63,    64,    65,    66,    -1,    -1,    27,
      28,    -1,    -1,    31,    -1,    75,    -1,    -1,    78,    -1,
      38,    -1,    82,    -1,    42,    -1,    -1,    -1,    -1,    47,
      -1,    -1,    -1,    -1,    -1,    -1,    54,    97,    56,    57,
      -1,    -1,    -1,    -1,    -1,    -1,    64,    65,    -1,    67,
      68,    -1,    -1,    -1,    72,    -1,    74,    -1,    -1,    -1,
      78,    79,    -1,    81,    82,    83,    -1,    -1,    -1,    -1,
      -1,    89,    90,    -1,    -1,    -1,    94,    95,     3,    -1,
       5,     6,     7,     8,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    27,    28,    -1,    -1,    31,    -1,    -1,    -1,
      -1,    -1,    -1,    38,    -1,    -1,    -1,    42,    -1,    -1,
      -1,    -1,    47,    -1,    -1,    -1,    -1,    -1,    -1,    54,
      -1,    56,    57,    -1,    -1,    -1,    -1,    -1,    -1,    64,
      65,    -1,    67,    68,    -1,    -1,    -1,    72,    -1,    74,
      -1,    -1,    -1,    78,    79,    -1,    81,    82,    83,    -1,
      -1,    -1,    -1,    -1,    89,    90,    -1,    -1,    -1,    94,
      95,     3,    -1,     5,     6,     7,     8,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    27,    28,    29,    -1,    31,
      32,    -1,    34,    -1,    -1,    -1,    38,    -1,    40,    -1,
      42,    -1,    -1,    -1,    -1,    47,    -1,    -1,    -1,    -1,
      -1,    -1,    54,    -1,    56,    57,    58,    59,    60,    61,
      62,    63,    64,    65,    -1,    67,    68,    -1,    -1,    -1,
      72,    -1,    74,    75,    -1,    -1,    78,    79,    -1,    81,
      82,    83,    -1,    -1,    -1,    -1,    -1,    89,    90,    -1,
      -1,     3,    94,     5,     6,     7,     8,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    -1,    31,
      -1,    -1,    -1,    -1,    -1,    -1,    38,    -1,    -1,    -1,
      42,    -1,    -1,    -1,    -1,    47,    -1,    -1,    -1,    -1,
      -1,    -1,    54,    -1,    56,    57,    -1,    -1,    -1,    -1,
      -1,    -1,    64,    65,    -1,    67,    68,    -1,    -1,    -1,
      72,    -1,    74,    -1,    -1,    -1,    78,    79,    -1,    81,
      82,    83,    -1,    -1,    -1,    -1,    -1,    89,    90,    -1,
      -1,     3,    94,     5,     6,     7,     8,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    -1,    31,
      -1,    -1,    -1,    -1,    -1,    -1,    38,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    47,    -1,    -1,    -1,    -1,
      -1,    -1,    54,    -1,    56,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    64,    65,    -1,    67,    68,    -1,    -1,     3,
      -1,     5,     6,     7,     8,    -1,    78,    79,    -1,    -1,
      82,    83,    -1,    -1,    -1,    -1,    -1,    89,    90,    -1,
      -1,    -1,    94,    27,    28,    -1,    -1,    31,    -1,    -1,
      -1,    -1,    -1,    -1,    38,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    47,    -1,    -1,    -1,    -1,    -1,    -1,
      54,    -1,    56,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      64,    65,    -1,    67,    68,    -1,    -1,     3,    -1,     5,
       6,     7,     8,    -1,    78,    79,    -1,    -1,    82,    83,
      -1,    -1,    -1,    -1,    -1,    89,    90,    -1,    -1,    -1,
      94,    27,    28,    -1,    -1,    31,    -1,    -1,    -1,    -1,
      -1,    -1,    38,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    47,    -1,    -1,    -1,    -1,    -1,    -1,    54,    -1,
      56,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    64,    65,
      -1,    67,    68,     3,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    78,    79,    -1,    -1,    82,    83,    -1,    -1,
      -1,    -1,    -1,    89,    90,    -1,    -1,    -1,    94,    29,
      -1,    31,    32,    -1,    34,    -1,    -1,    -1,    38,     3,
      40,     5,     6,     7,     8,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    54,    -1,    56,    -1,    58,    59,
      60,    61,    62,    63,    64,    65,    66,    31,    -1,    -1,
      -1,    -1,    -1,    -1,    38,    75,    -1,    -1,    78,    -1,
      -1,    -1,    82,    47,    -1,    -1,    -1,    -1,    -1,    -1,
      54,     3,    56,    -1,    94,    -1,    -1,    -1,    -1,    -1,
      64,    65,    -1,    67,    68,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    78,    79,    -1,    29,    82,    31,
      32,    -1,    34,    -1,    36,    -1,    38,    -1,    40,    -1,
      94,    -1,    -1,    -1,    -1,    -1,     3,    -1,    -1,    -1,
      -1,    -1,    54,    -1,    56,    -1,    58,    59,    60,    61,
      62,    63,    64,    65,    66,    -1,    -1,    69,    -1,    -1,
      -1,    -1,    29,    75,    31,    32,    78,    34,    -1,    -1,
      82,    38,    -1,    40,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    54,    -1,    56,
      -1,    58,    59,    60,    61,    62,    63,    64,    65,    66,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    75,    -1,
      -1,    78,    -1,    -1,    -1,    82
};

  /* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
     symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    38,    54,    64,    65,   107,   108,   109,   110,   143,
     145,     3,    31,    38,    54,    56,    64,    65,    78,    82,
     197,   198,     8,   198,   198,     0,   109,   110,   143,    53,
     111,    97,    97,   198,    97,    69,    71,   136,   137,   144,
     136,   146,    29,    32,    34,    36,    40,    46,    58,    59,
      60,    61,    62,    63,    66,    71,    75,    80,    98,   112,
     113,   114,   115,   116,   117,   118,   123,   124,   126,   130,
     131,   132,   136,   198,    69,    99,    50,    98,   136,    97,
     104,   141,    98,   136,   115,   117,   198,     1,   198,    94,
     114,     1,   198,    98,   113,   123,   130,    99,   100,     1,
     119,   120,   198,   141,    94,   198,    99,     1,     5,     6,
       7,     8,    27,    28,    30,    33,    35,    37,    39,    41,
      42,    44,    47,    49,    51,    52,    57,    67,    68,    70,
      72,    73,    74,    77,    79,    81,    83,    89,    90,    94,
      98,    99,   114,   123,   141,   142,   147,   148,   149,   150,
     152,   153,   154,   157,   160,   161,   162,   163,   164,   165,
     166,   167,   168,   173,   174,   175,   176,   177,   178,   179,
     181,   183,   185,   187,   189,   190,   191,   192,   195,   196,
     198,     8,   104,   141,    98,    97,   117,    98,    76,   122,
     166,   190,   198,    99,   101,   102,   117,   133,   134,   135,
      94,    99,    78,    99,    67,   179,    99,    56,    78,    82,
     104,    56,    78,    82,     1,    94,   141,   198,     1,    94,
      97,   117,   169,   170,   198,    31,    99,   165,     1,    94,
      97,   158,    94,   190,   118,   165,    98,   114,   123,   147,
      45,     1,    99,   100,   101,     1,    22,   105,    21,    84,
      85,    86,    25,    26,   180,    23,    24,    87,    88,   182,
      89,    90,   184,    19,    20,   186,    91,    92,    93,   188,
       9,    10,    11,    12,    13,    14,    15,    16,    17,    18,
     151,   190,    27,    28,    94,    96,   102,   104,    99,     8,
     127,   128,   198,   101,    97,    99,   101,   120,   121,   183,
     120,    95,   101,   133,   198,   104,    98,   115,   116,   148,
     150,   155,   165,   173,   198,    99,    98,   165,   170,   171,
     172,   173,   198,   102,    94,    94,    99,    98,   165,    98,
     152,   159,   198,    95,    95,    52,   141,    99,   165,   166,
      99,   175,   165,   176,   177,   178,   179,   181,   183,   185,
     187,   189,   165,   168,   193,   194,   198,   173,   147,   198,
      99,   101,   129,   100,   117,   116,   125,   166,   103,   135,
      95,    99,   156,   173,    55,   156,    55,    95,    95,   101,
     129,   104,   183,   193,   179,    95,    98,   147,   190,    94,
      99,   104,    99,    95,   101,   103,   128,    98,   174,    95,
      99,    98,   116,    94,   117,   138,   139,   140,    95,   165,
      99,   192,    95,   165,   192,   141,   141,   172,    98,   172,
     103,    95,    95,   101,   158,   165,   173,   168,    99,   139,
     101,   102,   141,    95,    95,   141,    95,    95,   183,    95,
      95,   140,   121,   141,   141,   141,   141,    95,   141,   103
};

  /* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint8 yyr1[] =
{
       0,   106,   107,   107,   108,   108,   108,   108,   108,   108,
     109,   110,   110,   111,   111,   112,   112,   112,   112,   112,
     112,   113,   113,   113,   114,   114,   115,   115,   116,   117,
     117,   117,   118,   118,   118,   118,   118,   118,   118,   118,
     118,   118,   118,   119,   119,   120,   120,   121,   121,   122,
     122,   123,   123,   124,   124,   125,   125,   126,   126,   127,
     127,   128,   128,   129,   129,   130,   131,   131,   132,   133,
     133,   134,   134,   135,   136,   137,   137,   137,   137,   138,
     138,   138,   139,   139,   140,   140,   141,   141,   142,   142,
     142,   142,   142,   142,   143,   144,   144,   145,   146,   146,
     146,   146,   147,   147,   147,   147,   147,   147,   147,   147,
     147,   147,   147,   148,   149,   149,   150,   150,   151,   151,
     151,   151,   151,   151,   151,   151,   151,   151,   152,   152,
     152,   153,   153,   153,   153,   154,   154,   154,   154,   154,
     154,   154,   154,   154,   155,   155,   156,   156,   157,   157,
     157,   158,   158,   159,   159,   160,   160,   160,   160,   160,
     161,   162,   162,   162,   162,   162,   162,   162,   163,   164,
     164,   165,   165,   166,   166,   167,   167,   167,   167,   167,
     168,   168,   168,   169,   169,   170,   171,   171,   172,   172,
     172,   173,   173,   174,   174,   175,   175,   176,   176,   177,
     177,   178,   178,   179,   179,   180,   180,   181,   181,   182,
     182,   182,   182,   183,   183,   184,   184,   185,   185,   186,
     186,   187,   187,   188,   188,   188,   189,   189,   190,   190,
     190,   191,   191,   191,   191,   192,   192,   192,   192,   192,
     192,   193,   193,   194,   194,   195,   195,   195,   195,   196,
     196,   196,   196,   196,   196,   196,   197,   197,   197,   197,
     197,   197,   197,   197,   198,   198
};

  /* YYR2[YYN] -- Number of symbols on the right hand side of rule YYN.  */
static const yytype_uint8 yyr2[] =
{
       0,     2,     1,     1,     1,     1,     1,     2,     2,     2,
       2,     5,     6,     0,     2,     1,     1,     1,     2,     2,
       2,     1,     2,     3,     1,     2,     2,     4,     2,     1,
       1,     6,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     3,     1,     4,     0,     1,     1,
       3,     1,     1,     6,     3,     2,     3,     6,     3,     1,
       3,     1,     3,     0,     1,     2,     1,     1,     4,     0,
       1,     1,     3,     2,     7,     0,     1,     1,     2,     0,
       3,     1,     1,     3,     1,     4,     2,     3,     1,     1,
       1,     2,     2,     2,     5,     2,     3,     5,     4,     2,
       5,     3,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     2,     2,     4,     4,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     3,     3,
       2,     5,     7,     3,     3,     2,     5,     6,     7,     6,
       7,     7,     7,     3,     1,     1,     1,     2,     2,     5,
       3,     2,     3,     1,     2,     2,     2,     2,     3,     3,
       3,     2,     2,     2,     2,     2,     2,     2,     1,     5,
       7,     1,     3,     1,     3,     1,     1,     1,     1,     1,
       1,     2,     2,     1,     4,     4,     1,     3,     1,     1,
       3,     1,     5,     1,     3,     1,     3,     1,     3,     1,
       3,     1,     3,     1,     3,     1,     1,     1,     3,     1,
       1,     1,     1,     1,     3,     1,     1,     1,     3,     1,
       1,     1,     3,     1,     1,     1,     1,     4,     1,     2,
       2,     1,     1,     1,     1,     1,     4,     4,     3,     2,
       2,     0,     1,     1,     3,     1,     1,     3,     5,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     1
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

#line 1926 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1430  */
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
        case 11:
#line 274 "grammar.y" /* yacc.c:1648  */
    {
        ast_blk_t *blk = blk_new_contract(&(yylsp[-1]));

        /* add default constructor */
        id_add(&blk->ids, id_new_ctor((yyvsp[-3].str), NULL, NULL, &(yylsp[-3])));

        id_add(&ROOT->ids, id_new_contract((yyvsp[-3].str), (yyvsp[-2].exp), blk, &(yyloc)));
    }
#line 2123 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 12:
#line 283 "grammar.y" /* yacc.c:1648  */
    {
        if (is_empty_vector(&(yyvsp[-1].blk)->ids) || !is_ctor_id(vector_get_id(&(yyvsp[-1].blk)->ids, 0)))
            /* add default constructor */
            id_add(&(yyvsp[-1].blk)->ids, id_new_ctor((yyvsp[-4].str), NULL, NULL, &(yylsp[-4])));

        id_add(&ROOT->ids, id_new_contract((yyvsp[-4].str), (yyvsp[-3].exp), (yyvsp[-1].blk), &(yyloc)));
    }
#line 2135 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 13:
#line 293 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 2141 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 14:
#line 295 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yylsp[0]));
    }
#line 2149 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 15:
#line 302 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2158 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 16:
#line 307 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2167 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 17:
#line 312 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2176 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 18:
#line 317 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2185 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 19:
#line 322 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2194 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 20:
#line 327 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2203 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 22:
#line 336 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_PUBLIC;
    }
#line 2212 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 23:
#line 341 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2221 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 25:
#line 350 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_CONST;
    }
#line 2230 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 27:
#line 359 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.dflt_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.dflt_exp = (yyvsp[-1].exp);
    }
#line 2243 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 28:
#line 371 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.type_exp = (yyvsp[-1].exp);
    }
#line 2256 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 29:
#line 383 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type((yyvsp[0].type), &(yylsp[0]));
    }
#line 2264 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 30:
#line 387 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_NONE, &(yylsp[0]));

        (yyval.exp)->u_type.name = (yyvsp[0].str);
    }
#line 2274 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 31:
#line 393 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_MAP, &(yylsp[-5]));

        (yyval.exp)->u_type.k_exp = (yyvsp[-3].exp);
        (yyval.exp)->u_type.v_exp = (yyvsp[-1].exp);
    }
#line 2285 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 32:
#line 402 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_ACCOUNT; }
#line 2291 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 33:
#line 403 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BOOL; }
#line 2297 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 34:
#line 404 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BYTE; }
#line 2303 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 35:
#line 405 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT8; }
#line 2309 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 36:
#line 406 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT16; }
#line 2315 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 37:
#line 407 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2321 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 38:
#line 408 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT64; }
#line 2327 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 39:
#line 409 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT128; }
#line 2333 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 40:
#line 410 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2339 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 41:
#line 413 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_STRING; }
#line 2345 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 42:
#line 414 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_CURSOR; }
#line 2351 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 44:
#line 420 "grammar.y" /* yacc.c:1648  */
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
#line 2367 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 45:
#line 435 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PRIVATE, &(yylsp[0]));
    }
#line 2375 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 46:
#line 439 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2388 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 47:
#line 450 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = exp_new_null(&(yyloc)); }
#line 2394 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 49:
#line 456 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 2402 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 50:
#line 460 "grammar.y" /* yacc.c:1648  */
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
#line 2417 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 53:
#line 479 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_struct((yyvsp[-4].str), (yyvsp[-1].vect), &(yyloc));
    }
#line 2425 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 54:
#line 483 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2434 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 55:
#line 491 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2447 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 56:
#line 500 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2460 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 57:
#line 512 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_enum((yyvsp[-4].str), (yyvsp[-2].vect), &(yyloc));
    }
#line 2468 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 58:
#line 516 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2477 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 59:
#line 524 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2486 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 60:
#line 529 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2495 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 61:
#line 537 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PUBLIC | MOD_CONST, &(yylsp[0]));
    }
#line 2503 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 62:
#line 541 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[-2].str), MOD_PUBLIC | MOD_CONST, &(yylsp[-2]));
        (yyval.id)->u_var.dflt_exp = (yyvsp[0].exp);
    }
#line 2512 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 65:
#line 554 "grammar.y" /* yacc.c:1648  */
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
#line 2529 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 68:
#line 575 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_ctor((yyvsp[-3].str), (yyvsp[-1].vect), NULL, &(yylsp[-3]));
    }
#line 2537 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 69:
#line 581 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 2543 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 71:
#line 587 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2552 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 72:
#line 592 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2561 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 73:
#line 600 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->u_var.is_param = true;
        (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
    }
#line 2571 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 74:
#line 609 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_func((yyvsp[-4].str), (yyvsp[-6].mod), (yyvsp[-2].vect), (yyvsp[0].id), NULL, &(yylsp[-4]));
    }
#line 2579 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 75:
#line 615 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PRIVATE; }
#line 2585 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 76:
#line 616 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC; }
#line 2591 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 77:
#line 617 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PAYABLE; }
#line 2597 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 78:
#line 618 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC | MOD_PAYABLE; }
#line 2603 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 79:
#line 622 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = NULL; }
#line 2609 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 80:
#line 623 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = (yyvsp[-1].id); }
#line 2615 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 83:
#line 630 "grammar.y" /* yacc.c:1648  */
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
#line 2645 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 84:
#line 659 "grammar.y" /* yacc.c:1648  */
    {
        /* We wanted to use a type expression, but we can not store size expressions
         * and declare it as an identifier. */
        (yyval.id) = id_new_param(NULL, (yyvsp[0].exp), &(yylsp[0]));
    }
#line 2655 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 85:
#line 665 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2668 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 86:
#line 676 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 2674 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 87:
#line 677 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 2680 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 88:
#line 682 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2689 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 89:
#line 687 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        /* Unlike state variables, local variables are referenced according to their
         * order of declaration. */
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2700 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 90:
#line 694 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2709 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 91:
#line 699 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2718 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 92:
#line 704 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2727 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 93:
#line 709 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2736 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 94:
#line 717 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, id_new_interface((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2744 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 95:
#line 724 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_interface(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2753 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 96:
#line 729 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2762 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 97:
#line 737 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, id_new_library((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2770 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 98:
#line 744 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-3].id)->mod = MOD_PUBLIC;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2783 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 99:
#line 753 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-1].id)->mod = MOD_PUBLIC;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2796 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 100:
#line 762 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-4].blk);

        (yyvsp[-3].id)->mod = MOD_PUBLIC;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2809 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 101:
#line 771 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);

        (yyvsp[-1].id)->mod = MOD_PUBLIC;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2822 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 113:
#line 797 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_null(&(yyloc));
    }
#line 2830 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 114:
#line 804 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_exp((yyvsp[-1].exp), &(yyloc));
    }
#line 2838 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 115:
#line 808 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2847 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 116:
#line 816 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2]));
    }
#line 2855 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 117:
#line 820 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), exp_new_binary((yyvsp[-2].op), (yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2])), &(yylsp[-2]));
    }
#line 2863 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 118:
#line 826 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 2869 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 119:
#line 827 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 2875 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 120:
#line 828 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 2881 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 121:
#line 829 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 2887 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 122:
#line 830 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 2893 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 123:
#line 831 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_AND; }
#line 2899 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 124:
#line 832 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_XOR; }
#line 2905 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 125:
#line 833 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_OR; }
#line 2911 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 126:
#line 834 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 2917 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 127:
#line 835 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 2923 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 128:
#line 840 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[0].stmt);
        id_add(LABELS, id_new_label((yyvsp[-2].str), (yyvsp[0].stmt), &(yylsp[-2])));
    }
#line 2932 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 129:
#line 845 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case((yyvsp[-1].exp), &(yyloc));
    }
#line 2940 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 130:
#line 849 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case(NULL, &(yyloc));
    }
#line 2948 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 131:
#line 856 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 2956 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 132:
#line 860 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-6].stmt);
        stmt_add(&(yyval.stmt)->u_if.elif_stmts, stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yylsp[-5])));
    }
#line 2965 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 133:
#line 865 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-2].stmt);
        (yyval.stmt)->u_if.else_blk = (yyvsp[0].blk);
    }
#line 2974 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 134:
#line 870 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2983 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 135:
#line 878 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, NULL, NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2991 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 136:
#line 882 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2999 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 137:
#line 886 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-3].stmt), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3007 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 138:
#line 890 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-4].stmt), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3015 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 139:
#line 894 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-3].id), &(yylsp[-3])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3023 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 140:
#line 898 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3031 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 141:
#line 902 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_exp((yyvsp[-4].exp), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3039 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 142:
#line 906 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3047 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 143:
#line 910 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3056 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 146:
#line 922 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 3062 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 148:
#line 928 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch(NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3070 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 149:
#line 932 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3078 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 150:
#line 936 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3087 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 151:
#line 943 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 3093 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 152:
#line 944 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 3099 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 153:
#line 949 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_switch(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3108 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 154:
#line 954 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3117 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 155:
#line 962 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_CONTINUE, NULL, &(yyloc));
    }
#line 3125 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 156:
#line 966 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_BREAK, NULL, &(yyloc));
    }
#line 3133 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 157:
#line 970 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return(NULL, &(yyloc));
    }
#line 3141 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 158:
#line 974 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return((yyvsp[-1].exp), &(yyloc));
    }
#line 3149 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 159:
#line 978 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_goto((yyvsp[-1].str), &(yylsp[-1]));
    }
#line 3157 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 160:
#line 985 "grammar.y" /* yacc.c:1648  */
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
#line 3177 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 168:
#line 1014 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_blk((yyvsp[0].blk), &(yyloc));
    }
#line 3185 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 169:
#line 1021 "grammar.y" /* yacc.c:1648  */
    {
        char *cond_str =
            xstrndup(parse->src + (yyloc).first_offset + (yylsp[-2]).last_col - 1, (yylsp[0]).first_col - (yylsp[-2]).last_col);

        (yyval.stmt) = stmt_new_pragma(PRAGMA_ASSERT, (yyvsp[-1].exp), cond_str, NULL, &(yylsp[-4]));
    }
#line 3196 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 170:
#line 1028 "grammar.y" /* yacc.c:1648  */
    {
        char *cond_str =
            xstrndup(parse->src + (yyloc).first_offset + (yylsp[-4]).last_col - 1, (yylsp[-2]).first_col - (yylsp[-4]).last_col);

        (yyval.stmt) = stmt_new_pragma(PRAGMA_ASSERT, (yyvsp[-3].exp), cond_str, (yyvsp[-1].exp), &(yylsp[-6]));
    }
#line 3207 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 172:
#line 1039 "grammar.y" /* yacc.c:1648  */
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
#line 3222 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 174:
#line 1054 "grammar.y" /* yacc.c:1648  */
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
#line 3242 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 175:
#line 1072 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_DELETE; }
#line 3248 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 176:
#line 1073 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_INSERT; }
#line 3254 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 177:
#line 1074 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_REPLACE; }
#line 3260 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 178:
#line 1075 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_QUERY; }
#line 3266 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 179:
#line 1076 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_UPDATE; }
#line 3272 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 181:
#line 1082 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3280 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 182:
#line 1086 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
        (yyval.exp)->u_init.is_outmost = true;
    }
#line 3289 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 183:
#line 1094 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_alloc((yyvsp[0].exp), &(yylsp[0]));
    }
#line 3297 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 184:
#line 1098 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-3].exp);

        if ((yyval.exp)->u_alloc.size_exps == NULL)
            (yyval.exp)->u_alloc.size_exps = vector_new();

        exp_add((yyval.exp)->u_alloc.size_exps, (yyvsp[-1].exp));
    }
#line 3310 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 185:
#line 1110 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_init((yyvsp[-2].vect), &(yyloc));
    }
#line 3318 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 186:
#line 1117 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3327 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 187:
#line 1122 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3336 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 190:
#line 1132 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3344 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 192:
#line 1140 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_ternary((yyvsp[-4].exp), (yyvsp[-2].exp), (yyvsp[0].exp), &(yyloc));
    }
#line 3352 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 194:
#line 1148 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3360 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 196:
#line 1156 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3368 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 198:
#line 1164 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3376 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 200:
#line 1172 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_XOR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3384 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 202:
#line 1180 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3392 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 204:
#line 1188 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3400 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 205:
#line 1194 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_EQ; }
#line 3406 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 206:
#line 1195 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NE; }
#line 3412 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 208:
#line 1201 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3420 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 209:
#line 1207 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LT; }
#line 3426 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 210:
#line 1208 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GT; }
#line 3432 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 211:
#line 1209 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LE; }
#line 3438 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 212:
#line 1210 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GE; }
#line 3444 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 214:
#line 1216 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3452 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 215:
#line 1222 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 3458 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 216:
#line 1223 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 3464 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 218:
#line 1229 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3472 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 219:
#line 1235 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 3478 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 220:
#line 1236 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 3484 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 222:
#line 1242 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3492 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 223:
#line 1248 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 3498 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 224:
#line 1249 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 3504 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 225:
#line 1250 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 3510 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 227:
#line 1256 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_cast((yyvsp[-2].type), (yyvsp[0].exp), &(yylsp[-2]));
    }
#line 3518 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 229:
#line 1264 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary((yyvsp[-1].op), true, (yyvsp[0].exp), &(yyloc));
    }
#line 3526 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 230:
#line 1268 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3534 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 231:
#line 1274 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_INC; }
#line 3540 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 232:
#line 1275 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DEC; }
#line 3546 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 233:
#line 1276 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NEG; }
#line 3552 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 234:
#line 1277 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NOT; }
#line 3558 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 236:
#line 1283 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_array((yyvsp[-3].exp), (yyvsp[-1].exp), &(yyloc));
    }
#line 3566 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 237:
#line 1287 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(FN_UDF, (yyvsp[-3].exp), (yyvsp[-1].vect), &(yyloc));
    }
#line 3574 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 238:
#line 1291 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_access((yyvsp[-2].exp), exp_new_id((yyvsp[0].str), &(yylsp[0])), &(yyloc));
    }
#line 3582 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 239:
#line 1295 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_INC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3590 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 240:
#line 1299 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_DEC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3598 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 241:
#line 1305 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 3604 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 243:
#line 1311 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3613 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 244:
#line 1316 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3622 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 246:
#line 1325 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yyloc));
    }
#line 3630 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 247:
#line 1329 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-1].exp);
    }
#line 3638 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 248:
#line 1333 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(FN_CTOR, exp_new_id((yyvsp[-3].str), &(yylsp[-3])), (yyvsp[-1].vect), &(yylsp[-3]));
    }
#line 3646 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 249:
#line 1340 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_null(&(yyloc));
    }
#line 3654 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 250:
#line 1344 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(true, &(yyloc));
    }
#line 3662 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 251:
#line 1348 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(false, &(yyloc));
    }
#line 3670 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 252:
#line 1352 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 10);
    }
#line 3679 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 253:
#line 1357 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 8);
    }
#line 3688 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 254:
#line 1362 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 0);
    }
#line 3697 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 255:
#line 1376 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_str((yyvsp[0].str), &(yyloc));
    }
#line 3705 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 256:
#line 1382 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("assert"); }
#line 3711 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 257:
#line 1383 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("contract"); }
#line 3717 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 258:
#line 1384 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("import"); }
#line 3723 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 259:
#line 1385 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("index"); }
#line 3729 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 260:
#line 1386 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("interface"); }
#line 3735 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 261:
#line 1387 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("library"); }
#line 3741 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 262:
#line 1388 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("table"); }
#line 3747 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 263:
#line 1389 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("view"); }
#line 3753 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 264:
#line 1394 "grammar.y" /* yacc.c:1648  */
    {
        if (strlen((yyvsp[0].str)) > NAME_MAX_LEN)
            ERROR(ERROR_TOO_LONG_ID, &(yylsp[0]), NAME_MAX_LEN, strlen((yyvsp[0].str)));

        (yyval.str) = (yyvsp[0].str);
    }
#line 3764 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;


#line 3768 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
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
#line 1403 "grammar.y" /* yacc.c:1907  */


static void
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
