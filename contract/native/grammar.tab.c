/* A Bison parser, made by GNU Bison 3.3.2.  */

/* Bison implementation for Yacc-like parsers in C

   Copyright (C) 1984, 1989-1990, 2000-2015, 2018-2019 Free Software Foundation,
   Inc.

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

/* Undocumented macros, especially those whose name start with YY_,
   are private implementation details.  Do not rely on them.  */

/* Identify Bison output.  */
#define YYBISON 1

/* Bison version.  */
#define YYBISON_VERSION "3.3.2"

/* Skeleton name.  */
#define YYSKELETON_NAME "yacc.c"

/* Pure parsers.  */
#define YYPURE 2

/* Push parsers.  */
#define YYPUSH 0

/* Pull parsers.  */
#define YYPULL 1




/* First part of user prologue.  */
#line 1 "grammar.y" /* yacc.c:337  */


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


#line 98 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:337  */
# ifndef YY_NULLPTR
#  if defined __cplusplus
#   if 201103L <= __cplusplus
#    define YY_NULLPTR nullptr
#   else
#    define YY_NULLPTR 0
#   endif
#  else
#   define YY_NULLPTR ((void*)0)
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
    K_UINT = 336,
    K_UINT8 = 337,
    K_UINT16 = 338,
    K_UINT32 = 339,
    K_UINT64 = 340,
    K_UINT128 = 341,
    K_UPDATE = 342,
    K_VIEW = 343
  };
#endif

/* Value type.  */
#if ! defined YYSTYPE && ! defined YYSTYPE_IS_DECLARED

union YYSTYPE
{
#line 154 "grammar.y" /* yacc.c:352  */

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

#line 248 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:352  */
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
typedef unsigned short yytype_uint16;
#endif

#ifdef YYTYPE_INT16
typedef YYTYPE_INT16 yytype_int16;
#else
typedef short yytype_int16;
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
#  define YYSIZE_T unsigned
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

/* Suppress unused-variable warnings by "using" E.  */
#if ! defined lint || defined __GNUC__
# define YYUSE(E) ((void) (E))
#else
# define YYUSE(E) /* empty */
#endif

#if defined __GNUC__ && ! defined __ICC && 407 <= __GNUC__ * 100 + __GNUC_MINOR__
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
#define YYLAST   1943

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  112
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  93
/* YYNRULES -- Number of rules.  */
#define YYNRULES  271
/* YYNSTATES -- Number of states.  */
#define YYNSTATES  456

#define YYUNDEFTOK  2
#define YYMAXUTOK   343

/* YYTRANSLATE(TOKEN-NUM) -- Symbol number corresponding to TOKEN-NUM
   as returned by yylex, with out-of-bounds checking.  */
#define YYTRANSLATE(YYX)                                                \
  ((unsigned) (YYX) <= YYMAXUTOK ? yytranslate[YYX] : YYUNDEFTOK)

/* YYTRANSLATE[TOKEN-NUM] -- Symbol number corresponding to TOKEN-NUM
   as returned by yylex.  */
static const yytype_uint8 yytranslate[] =
{
       0,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    89,     2,     2,     2,    99,    92,     2,
     100,   101,    97,    95,   107,    96,   102,    98,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,   110,   105,
      93,   106,    94,   111,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,   108,     2,   109,    91,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,   103,    90,   104,     2,     2,     2,     2,
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
      75,    76,    77,    78,    79,    80,    81,    82,    83,    84,
      85,    86,    87,    88
};

#if YYDEBUG
  /* YYRLINE[YYN] -- Source line where rule number YYN was defined.  */
static const yytype_uint16 yyrline[] =
{
       0,   261,   261,   262,   266,   267,   268,   269,   270,   271,
     275,   279,   288,   299,   300,   307,   312,   317,   322,   327,
     332,   340,   341,   346,   354,   355,   363,   364,   376,   388,
     392,   398,   408,   409,   410,   411,   412,   413,   414,   415,
     416,   417,   418,   419,   420,   421,   422,   425,   426,   430,
     431,   446,   450,   462,   463,   467,   471,   485,   486,   490,
     494,   502,   511,   523,   527,   535,   540,   548,   552,   559,
     561,   565,   581,   582,   586,   593,   594,   598,   603,   611,
     620,   627,   628,   629,   630,   634,   635,   636,   640,   641,
     670,   676,   688,   689,   693,   698,   705,   710,   715,   720,
     728,   735,   740,   748,   755,   764,   773,   782,   794,   795,
     796,   797,   798,   799,   800,   801,   802,   803,   804,   808,
     815,   819,   827,   831,   838,   839,   840,   841,   842,   843,
     844,   845,   846,   847,   851,   856,   860,   867,   871,   876,
     881,   889,   893,   897,   901,   905,   909,   913,   917,   921,
     929,   930,   934,   935,   939,   943,   947,   955,   956,   960,
     965,   973,   977,   981,   985,   989,   996,  1015,  1016,  1017,
    1018,  1019,  1020,  1021,  1025,  1032,  1039,  1049,  1050,  1064,
    1065,  1084,  1085,  1086,  1087,  1088,  1092,  1093,  1097,  1105,
    1109,  1121,  1128,  1133,  1141,  1142,  1143,  1150,  1151,  1158,
    1159,  1166,  1167,  1174,  1175,  1182,  1183,  1190,  1191,  1198,
    1199,  1206,  1207,  1211,  1212,  1219,  1220,  1221,  1222,  1226,
    1227,  1234,  1235,  1239,  1240,  1247,  1248,  1252,  1253,  1260,
    1261,  1262,  1266,  1267,  1274,  1275,  1279,  1286,  1287,  1288,
    1289,  1293,  1294,  1298,  1302,  1306,  1310,  1317,  1318,  1322,
    1327,  1335,  1336,  1340,  1344,  1351,  1355,  1359,  1363,  1368,
    1373,  1387,  1394,  1395,  1396,  1397,  1398,  1399,  1400,  1401,
    1405,  1412
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
  "\"table\"", "\"true\"", "\"type\"", "\"uint\"", "\"uint8\"",
  "\"uint16\"", "\"uint32\"", "\"uint64\"", "\"uint128\"", "\"update\"",
  "\"view\"", "'!'", "'|'", "'^'", "'&'", "'<'", "'>'", "'+'", "'-'",
  "'*'", "'/'", "'%'", "'('", "')'", "'.'", "'{'", "'}'", "';'", "'='",
  "','", "'['", "']'", "':'", "'?'", "$accept", "root", "component",
  "import", "contract", "impl_opt", "contract_body", "variable",
  "var_qual", "var_decl", "var_spec", "var_type", "prim_type",
  "declarator_list", "declarator", "size_opt", "var_init", "compound",
  "struct", "field_list", "enumeration", "enum_list", "enumerator",
  "comma_opt", "function", "func_spec", "ctor_spec", "param_list_opt",
  "param_list", "param_decl", "udf_spec", "modifier_opt", "return_opt",
  "return_list", "return_decl", "block", "blk_decl", "interface",
  "interface_body", "library", "library_body", "statement", "empty_stmt",
  "exp_stmt", "assign_stmt", "assign_op", "label_stmt", "if_stmt",
  "loop_stmt", "init_stmt", "cond_exp", "switch_stmt", "switch_blk",
  "case_blk", "jump_stmt", "ddl_stmt", "ddl_prefix", "blk_stmt",
  "pragma_stmt", "expression", "sql_exp", "sql_prefix", "new_exp",
  "alloc_exp", "initializer", "elem_list", "init_elem", "ternary_exp",
  "or_exp", "and_exp", "bit_or_exp", "bit_xor_exp", "bit_and_exp",
  "eq_exp", "eq_op", "cmp_exp", "cmp_op", "add_exp", "add_op", "shift_exp",
  "shift_op", "mul_exp", "mul_op", "cast_exp", "unary_exp", "unary_op",
  "post_exp", "arg_list_opt", "arg_list", "prim_exp", "literal",
  "non_reserved_token", "identifier", YY_NULLPTR
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
     335,   336,   337,   338,   339,   340,   341,   342,   343,    33,
     124,    94,    38,    60,    62,    43,    45,    42,    47,    37,
      40,    41,    46,   123,   125,    59,    61,    44,    91,    93,
      58,    63
};
# endif

#define YYPACT_NINF -273

#define yypact_value_is_default(Yystate) \
  (!!((Yystate) == (-273)))

#define YYTABLE_NINF -31

#define yytable_value_is_error(Yytable_value) \
  0

  /* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
     STATE-NUM.  */
static const yytype_int16 yypact[] =
{
     138,   162,    19,   162,   162,   144,  -273,  -273,  -273,  -273,
    -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,
    -273,   -17,  -273,   -84,   -48,  -273,  -273,  -273,  -273,   162,
     -41,     3,     3,  -273,   972,  -273,    18,     7,    68,   -12,
     -73,    -3,  -273,  -273,  -273,  1855,  -273,   150,  -273,  -273,
    -273,  -273,  -273,  -273,    35,  1816,  -273,   275,  -273,  -273,
    -273,  -273,  -273,  -273,  -273,  1037,  -273,  -273,  -273,    -8,
     764,  -273,  -273,  -273,  -273,  -273,    46,  -273,  -273,    83,
    -273,  -273,   162,  -273,    57,   390,   170,  -273,  -273,   -58,
    -273,   162,  -273,   108,    91,  1855,  -273,   132,   165,  -273,
    -273,  -273,  -273,  -273,  1504,   169,   178,   171,  -273,  -273,
    1855,   193,  -273,   195,  -273,  -273,  -273,  -273,  -273,  -273,
     224,   203,  1620,   204,    48,   205,  -273,    59,  -273,    22,
     162,     0,  -273,  1167,  -273,   283,  -273,   902,  -273,    24,
    -273,  -273,  -273,  1663,  -273,  1408,  -273,  -273,  -273,  -273,
    -273,   495,  -273,  -273,  -273,  -273,  -273,   278,  -273,  -273,
    -273,  -273,   321,  -273,  -273,   126,  -273,   323,  -273,  -273,
       2,   304,   237,   239,   236,   227,    37,   172,   277,   163,
    -273,   496,  1663,   260,  -273,  -273,    11,   229,   324,  -273,
    -273,   162,   230,  -273,   235,    62,  -273,  -273,  -273,  -273,
     162,  1620,   162,   240,   238,  -273,  1855,  -273,  -273,  -273,
     162,    -4,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,
     242,   806,  -273,   244,   246,  1504,  1209,  -273,   234,  -273,
     251,   252,  -273,   151,   254,  1504,   107,  -273,  1504,  -273,
     255,   -60,  -273,  -273,  -273,  -273,   -20,   250,  -273,  1504,
    1504,   256,  1620,  1504,  1620,  1620,  1620,  1620,  -273,  -273,
    1620,  -273,  -273,  -273,  -273,  1620,  -273,  -273,  1620,  -273,
    -273,  1620,  -273,  -273,  -273,  1620,  -273,  -273,  -273,  -273,
    -273,  -273,  -273,  -273,  -273,  -273,  1504,  -273,  -273,  -273,
    1718,   162,  1620,   705,  -273,   259,   258,  -273,   253,  1855,
    1855,  -273,  1504,   171,   261,   172,   171,  -273,  1855,   268,
     251,  -273,  -273,   216,   -26,  -273,  -273,   216,   -29,   271,
     279,  -273,  -273,    32,  -273,   266,  -273,  -273,   264,  1620,
    1718,  1620,  -273,  -273,    39,  -273,  -273,   600,   267,  1663,
    -273,   276,  -273,  -273,   159,  -273,  -273,   304,    47,   237,
     239,   236,   227,    37,   172,   277,   163,  -273,   164,  -273,
     280,   272,  -273,   274,  -273,   267,  -273,   162,   281,  1620,
     287,   284,  1102,  -273,  -273,  -273,  1556,  -273,  1254,   285,
    1761,  1331,  1761,    46,    46,  1209,   288,  1209,   -61,   293,
      13,   296,  -273,  -273,  -273,  1504,  -273,  1620,  -273,  -273,
    1718,  -273,  -273,  -273,   358,  -273,  -273,  -273,   295,  1855,
    -273,  -273,   294,   297,    46,    63,  -273,    58,    46,    79,
     218,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  1620,
    -273,    90,  -273,  -273,  -273,   148,  1855,  1620,  -273,    46,
      46,  -273,    46,    46,    78,    46,  -273,   297,   301,  -273,
    -273,  -273,  -273,  -273,  -273,  -273
};

  /* YYDEFACT[STATE-NUM] -- Default reduction number in state STATE-NUM.
     Performed when YYTABLE does not specify something else to do.  Zero
     means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       0,     0,     0,     0,     0,     0,     2,     4,     5,     6,
       3,   270,   262,   263,   264,   265,   266,   267,   268,   269,
     271,    13,    10,     0,     0,     1,     7,     8,     9,     0,
       0,    81,    81,    14,    81,    83,    82,     0,     0,    81,
       0,    81,    32,    33,    34,     0,    48,     0,    40,    35,
      36,    37,    38,    39,     0,    82,    47,     0,    46,    41,
      42,    43,    44,    45,    11,    81,    15,    21,    24,     0,
       0,    29,    16,    57,    58,    17,     0,    72,    73,    30,
      84,   101,     0,   100,     0,     0,     0,   105,   103,     0,
      25,     0,    30,     0,     0,     0,    22,     0,     0,    12,
      18,    19,    20,    26,     0,     0,    28,    49,    51,    71,
      75,     0,   102,     0,   260,   258,   259,   261,   237,   238,
       0,     0,     0,     0,     0,     0,   181,     0,   257,     0,
       0,     0,   182,     0,   255,     0,   183,     0,   184,     0,
     256,   185,   240,     0,   239,     0,    92,   119,    94,    95,
     174,     0,    96,   108,   109,   110,   111,   112,   113,   114,
     115,   116,     0,   117,   118,     0,   177,     0,   179,   186,
     197,   199,   201,   203,   205,   207,   209,   213,   219,   223,
     227,   232,     0,   234,   241,   251,   252,     0,     0,   107,
      64,     0,     0,    60,     0,     0,    55,   232,   252,    23,
       0,    53,     0,     0,    76,    77,    75,   121,   168,   162,
       0,     0,   161,   170,   167,   172,   136,   171,   169,   173,
       0,     0,   141,     0,     0,     0,     0,   189,   187,   188,
      30,     0,   163,     0,     0,     0,     0,   154,     0,   236,
       0,     0,    93,    97,    98,    99,     0,     0,   120,     0,
       0,     0,     0,     0,     0,     0,     0,     0,   211,   212,
       0,   217,   218,   215,   216,     0,   221,   222,     0,   225,
     226,     0,   229,   230,   231,     0,   124,   125,   126,   127,
     128,   129,   130,   131,   132,   133,     0,   235,   245,   246,
     247,     0,     0,     0,   104,     0,    69,    65,    67,     0,
       0,    27,     0,    50,     0,    54,    79,    74,     0,     0,
       0,   135,   149,     0,     0,   150,   151,     0,     0,   186,
     252,   165,   140,     0,   195,    69,   192,   194,   252,     0,
     247,     0,   164,   156,     0,   157,   159,     0,     0,     0,
     253,     0,   139,   166,     0,   178,   180,   200,     0,   202,
     204,   206,   208,   210,   214,   220,   224,   228,     0,   249,
       0,   248,   244,     0,   134,   252,   106,    70,     0,     0,
       0,     0,     0,    56,    52,    78,    85,   152,     0,     0,
       0,     0,     0,     0,     0,    70,     0,     0,     0,     0,
       0,     0,   158,   160,   233,     0,   122,     0,   123,   243,
       0,   242,    66,    63,    68,    31,    61,    59,     0,     0,
      90,    80,    87,    88,     0,     0,   153,     0,     0,     0,
       0,   142,   137,   193,   191,   196,   190,   254,   175,     0,
     155,     0,   198,   250,    62,     0,     0,    53,   145,     0,
       0,   143,     0,     0,     0,     0,    86,    89,     0,   146,
     148,   144,   147,   176,   138,    91
};

  /* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -273,  -273,  -273,   397,   406,  -273,  -273,   347,   -35,   -32,
    -204,   -22,   270,  -273,    73,   -24,  -273,   -34,  -273,  -273,
    -273,  -273,    49,   113,   368,  -273,  -273,   282,  -273,   135,
     198,  -273,  -273,    31,     9,   -36,  -273,   454,  -273,  -273,
    -273,  -141,   245,  -273,   262,  -273,   225,  -273,  -273,  -273,
     167,  -273,    89,  -273,  -273,  -273,  -273,  -273,  -273,  -130,
     -95,  -273,  -272,  -273,   348,  -273,   -96,  -210,   118,   247,
     228,   249,   233,  -114,  -273,   231,  -273,  -195,  -273,   248,
    -273,   221,  -273,   222,   -80,  -273,   -79,   185,  -273,  -273,
    -273,  -273,    -1
};

  /* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     5,     6,     7,     8,    30,    65,    66,    67,    68,
      69,    91,    71,   106,   107,   304,   195,    72,    73,   372,
      74,   296,   297,   368,    75,    76,    77,   203,   204,   205,
      78,    38,   411,   412,   413,   150,   151,     9,    39,    10,
      41,   152,   153,   154,   155,   286,   156,   157,   158,   317,
     378,   159,   237,   337,   160,   161,   162,   163,   164,   165,
     166,   167,   168,   228,   324,   325,   326,   169,   170,   171,
     172,   173,   174,   175,   260,   176,   265,   177,   268,   178,
     271,   179,   275,   180,   197,   182,   183,   360,   361,   184,
     185,    20,   198
};

  /* YYTABLE[YYPACT[STATE-NUM]] -- What to do in state STATE-NUM.  If
     positive, shift that token.  If negative, reduce the rule whose
     number is the opposite.  If YYTABLE_NINF, syntax error.  */
static const yytype_int16 yytable[] =
{
      21,   224,    23,    24,    87,   181,   305,   233,   211,   196,
     245,   319,    70,    90,   -30,   241,   327,   314,   359,    31,
      96,   258,   259,   220,   252,   234,   382,    22,    33,   380,
      85,   101,   341,    79,   266,   267,    29,    86,   258,   259,
     109,   340,   -30,    70,    92,    85,    94,   250,   426,   -30,
     148,   149,   188,   189,    92,    32,    98,    35,   359,    36,
     261,   262,    34,   239,    79,   -30,    35,   -30,    36,   108,
     354,   181,    35,   192,    36,   -30,   -30,   249,   250,   103,
     104,   111,   363,    85,   186,   288,   289,    80,   202,   -30,
     108,   318,    83,   222,    92,   323,   371,   103,   104,   -30,
     225,    88,   287,   379,   213,   334,   311,   379,   241,    92,
      11,   227,    81,   253,   428,   217,   243,   244,    82,   344,
     429,   293,   221,   348,   235,    85,   214,   236,   433,   223,
     263,   264,   230,   384,   388,    95,   215,   218,    12,   250,
     391,   181,   122,   352,    25,    13,   250,   219,   125,    85,
     186,    93,   364,    11,   250,   345,   358,   397,   290,   440,
     291,    14,   112,    15,   439,    11,   292,   301,   408,   302,
     250,    16,    17,   266,   267,   327,     1,   327,   187,   453,
     442,    12,     1,   110,   202,    18,   250,   432,    13,   313,
     298,   445,     2,    12,   191,    19,   393,   250,     2,   108,
      13,   108,     3,     4,    14,    92,    15,   373,     3,   310,
     342,   335,   190,   181,    16,    17,    14,   390,    15,    11,
     320,   114,   115,   116,   117,   328,    16,    17,    18,    37,
      40,   248,   249,   250,   444,   338,   193,    84,    19,    89,
      18,   194,   305,   118,   119,   288,   289,    12,   415,   446,
      19,   419,   258,   259,    13,   436,   332,   181,   250,   394,
     272,   273,   274,   128,   396,   431,   250,   266,   267,   398,
      14,   250,    15,   303,   199,   306,    97,   370,    11,   201,
      16,    17,   -30,   210,   134,   200,   202,   288,   289,   423,
     362,   425,   365,   206,    18,   140,   269,   270,    92,    92,
     207,   417,   208,   420,    19,   142,    12,    92,   209,   212,
     -30,   143,   144,    13,   231,   216,   145,   -30,   290,   443,
     291,   377,   247,   246,   251,   254,   292,   255,   257,    14,
     256,    15,   295,   -30,   294,   -30,   365,   299,   300,    16,
      17,   307,   329,   -30,   -30,   308,   312,   421,   422,   321,
     322,   330,   331,    18,   410,   343,   339,   -30,   333,   369,
     290,   346,   291,    19,   366,   367,   298,   -30,   292,   376,
     374,    92,   383,   385,   387,    92,   395,   293,   438,   400,
     252,   399,   441,   401,   328,   403,   328,   410,   405,   406,
     416,   113,   424,    11,   427,   114,   115,   116,   117,   236,
     434,   436,    26,   449,   450,   437,   451,   452,    92,   454,
     455,    27,   100,   448,   410,   240,   402,   118,   119,    42,
     120,    12,    43,   121,    44,   122,    45,   123,    13,   124,
      46,   125,   126,   102,   127,    92,    47,   128,   386,   129,
     435,   130,   131,   375,    14,   447,    15,   132,    48,    49,
      50,    51,    52,    53,    16,    17,    54,   133,   134,    28,
     135,   336,   136,   137,   138,    56,   315,   139,    18,   140,
      57,    58,    59,    60,    61,    62,    63,   141,    19,   142,
     430,   229,   349,   316,   381,   143,   144,   404,   309,   351,
     145,   353,   356,    85,   146,   147,   113,   357,    11,   347,
     114,   115,   116,   117,   350,   276,   277,   278,   279,   280,
     281,   282,   283,   284,   285,   389,   355,     0,     0,     0,
       0,     0,   118,   119,    42,   120,    12,    43,   121,    44,
     122,    45,   123,    13,   124,    46,   125,   126,     0,   127,
       0,    47,   128,     0,   129,     0,   130,   131,     0,    14,
       0,    15,   132,    48,    49,    50,    51,    52,    53,    16,
      17,    54,   133,   134,     0,   135,     0,   136,   137,   138,
      56,     0,   139,    18,   140,    57,    58,    59,    60,    61,
      62,    63,   141,    19,   142,     0,     0,     0,     0,     0,
     143,   144,     0,     0,     0,   145,     0,     0,    85,   242,
     147,   113,     0,    11,     0,   114,   115,   116,   117,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,   118,   119,     0,
     120,    12,     0,   121,     0,   122,     0,   123,    13,   124,
       0,   125,   126,     0,   127,     0,     0,   128,     0,   129,
       0,   130,   131,     0,    14,     0,    15,   132,     0,     0,
       0,     0,     0,     0,    16,    17,     0,   133,   134,     0,
     135,     0,   136,   137,   138,     0,     0,   139,    18,   140,
       0,     0,     0,     0,     0,     0,     0,   141,    19,   142,
       0,     0,     0,     0,     0,   143,   144,     0,     0,     0,
     145,     0,     0,    85,   392,   147,   113,     0,    11,     0,
     114,   115,   116,   117,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,   118,   119,     0,   120,    12,     0,   121,     0,
     122,     0,   123,    13,   124,     0,   125,   126,     0,   127,
       0,     0,   128,     0,   129,     0,   130,   131,     0,    14,
       0,    15,   132,     0,     0,   105,     0,    11,     0,    16,
      17,     0,   133,   134,     0,   135,     0,   136,   137,   138,
       0,     0,   139,    18,   140,     0,     0,     0,     0,     0,
       0,     0,   141,    19,   142,    12,     0,     0,     0,     0,
     143,   144,    13,     0,     0,   145,     0,     0,    85,    11,
     147,   114,   115,   116,   117,     0,     0,     0,    14,     0,
      15,     0,     0,     0,     0,     0,     0,     0,    16,    17,
       0,     0,     0,   118,   119,    42,     0,    12,    43,     0,
      44,     0,    18,     0,    13,     0,    46,     0,   126,     0,
       0,     0,    19,   128,     0,     0,     0,     0,     0,     0,
      14,     0,    15,   132,    48,    49,    50,    51,    52,    53,
      16,    17,    54,   133,   134,     0,     0,     0,   136,     0,
     138,    56,     0,     0,    18,   140,     0,    58,    59,    60,
      61,    62,    63,   141,    19,   142,     0,     0,     0,     0,
       0,   143,   144,     0,     0,    11,   145,   114,   115,   116,
     117,   147,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,   118,
     119,     0,     0,    12,     0,     0,     0,     0,     0,     0,
      13,     0,     0,     0,   126,     0,     0,     0,     0,   128,
       0,     0,     0,     0,     0,     0,    14,     0,    15,   132,
       0,     0,     0,     0,     0,     0,    16,    17,     0,   133,
     134,     0,     0,     0,   136,    11,   138,     0,     0,     0,
      18,   140,     0,     0,     0,     0,     0,     0,     0,   141,
      19,   142,     0,     0,     0,     0,     0,   143,   144,     0,
       0,    42,   145,    12,    43,     0,    44,   232,    45,     0,
      13,     0,    46,     0,     0,     0,     0,     0,    47,     0,
       0,     0,     0,     0,     0,     0,    14,     0,    15,     0,
      48,    49,    50,    51,    52,    53,    16,    17,    54,     0,
      11,    35,     0,    55,     0,     0,     0,    56,     0,     0,
      18,     0,    57,    58,    59,    60,    61,    62,    63,     0,
      19,     0,     0,     0,     0,     0,    42,     0,    12,    43,
       0,    44,     0,    45,     0,    13,    64,    46,     0,     0,
       0,     0,     0,    47,     0,     0,     0,     0,     0,     0,
       0,    14,     0,    15,     0,    48,    49,    50,    51,    52,
      53,    16,    17,    54,     0,    11,    35,     0,    55,     0,
       0,     0,    56,     0,     0,    18,     0,    57,    58,    59,
      60,    61,    62,    63,     0,    19,     0,     0,     0,     0,
       0,    42,     0,    12,    43,     0,    44,     0,     0,     0,
      13,    99,    46,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,    14,     0,    15,     0,
      48,    49,    50,    51,    52,    53,    16,    17,    54,     0,
      11,     0,     0,     0,     0,     0,     0,    56,     0,     0,
      18,     0,     0,    58,    59,    60,    61,    62,    63,     0,
      19,     0,     0,     0,     0,     0,    42,     0,    12,    43,
       0,    44,     0,     0,     0,    13,   407,    46,     0,     0,
       0,     0,    11,     0,   114,   115,   116,   117,     0,     0,
       0,    14,     0,    15,     0,    48,    49,    50,    51,    52,
      53,    16,    17,    54,     0,     0,   118,   119,     0,     0,
      12,     0,    56,     0,     0,    18,     0,    13,    58,    59,
      60,    61,    62,    63,     0,    19,   128,    11,     0,   114,
     115,   116,   117,    14,     0,    15,     0,     0,     0,     0,
     226,     0,     0,    16,    17,     0,   210,   134,     0,     0,
       0,   118,   119,     0,     0,    12,     0,    18,   140,     0,
       0,     0,    13,     0,     0,     0,   126,    19,   142,     0,
       0,   128,     0,     0,   143,   144,     0,     0,    14,   145,
      15,   132,   226,     0,     0,     0,     0,     0,    16,    17,
       0,   133,   134,     0,     0,     0,   136,     0,   138,     0,
       0,     0,    18,   140,    11,     0,   114,   115,   116,   117,
       0,   141,    19,   142,     0,     0,     0,     0,     0,   143,
     144,     0,     0,     0,   145,   414,     0,     0,   118,   119,
       0,     0,    12,     0,     0,     0,     0,     0,     0,    13,
       0,     0,     0,   126,     0,     0,     0,     0,   128,     0,
       0,     0,     0,     0,     0,    14,     0,    15,   132,     0,
       0,     0,     0,     0,     0,    16,    17,     0,   133,   134,
       0,     0,     0,   136,     0,   138,     0,     0,     0,    18,
     140,    11,     0,   114,   115,   116,   117,     0,   141,    19,
     142,     0,     0,     0,     0,     0,   143,   144,     0,     0,
       0,   145,   418,     0,     0,   118,   119,    42,     0,    12,
      43,     0,    44,     0,     0,     0,    13,     0,    46,     0,
     126,     0,     0,     0,     0,   128,     0,     0,     0,     0,
       0,     0,    14,     0,    15,   132,    48,    49,    50,    51,
      52,    53,    16,    17,     0,   133,   134,     0,     0,     0,
     136,     0,   138,    56,     0,     0,    18,   140,     0,    58,
      59,    60,    61,    62,    63,   141,    19,   142,     0,     0,
       0,     0,     0,   143,   144,     0,     0,    11,   145,   114,
     115,   116,   117,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,   118,   119,     0,     0,    12,     0,     0,     0,     0,
       0,     0,    13,     0,     0,     0,   126,     0,     0,     0,
       0,   128,     0,     0,     0,     0,     0,     0,    14,    11,
      15,   132,     0,     0,     0,     0,     0,     0,    16,    17,
       0,   133,   134,     0,     0,     0,   136,     0,   138,     0,
       0,     0,    18,   140,     0,    42,     0,    12,    43,     0,
      44,   141,    19,   142,    13,     0,    46,     0,     0,   143,
     144,     0,     0,     0,   145,     0,     0,     0,     0,     0,
      14,     0,    15,     0,    48,    49,    50,    51,    52,    53,
      16,    17,    54,    11,     0,   114,   115,   116,   117,     0,
       0,    56,     0,     0,    18,     0,     0,    58,    59,    60,
      61,    62,    63,     0,    19,     0,     0,   118,   119,     0,
       0,    12,     0,     0,     0,     0,   409,     0,    13,     0,
       0,     0,     0,     0,     0,     0,    11,   128,   114,   115,
     116,   117,     0,     0,    14,     0,    15,     0,     0,     0,
       0,     0,     0,     0,    16,    17,     0,   210,   134,     0,
     118,   119,     0,     0,    12,     0,     0,     0,    18,   140,
       0,    13,     0,     0,     0,     0,     0,     0,    19,   142,
     128,     0,     0,     0,     0,   143,   144,    14,     0,    15,
     145,    11,     0,   114,   115,   116,   117,    16,    17,     0,
     210,   134,     0,     0,     0,     0,     0,     0,     0,     0,
       0,    18,   140,     0,     0,   118,   119,     0,     0,    12,
       0,    19,   142,     0,     0,     0,    13,     0,   143,   144,
       0,     0,     0,   238,    11,   128,   114,   115,   116,   117,
       0,     0,    14,     0,    15,     0,     0,     0,     0,     0,
       0,     0,    16,    17,     0,   133,   134,     0,     0,     0,
       0,     0,    12,     0,     0,     0,    18,   140,     0,    13,
       0,     0,     0,     0,     0,     0,    19,   142,   128,     0,
       0,     0,     0,   143,   144,    14,     0,    15,   145,    11,
       0,     0,     0,     0,     0,    16,    17,     0,   210,   134,
       0,     0,     0,     0,     0,     0,     0,     0,     0,    18,
     140,     0,     0,     0,     0,    42,     0,    12,    43,    19,
      44,     0,    45,     0,    13,     0,    46,     0,    11,     0,
       0,   238,     0,     0,     0,     0,     0,     0,     0,     0,
      14,     0,    15,     0,    48,    49,    50,    51,    52,    53,
      16,    17,    54,     0,    42,    80,    12,    43,     0,    44,
       0,    56,     0,    13,    18,    46,     0,    58,    59,    60,
      61,    62,    63,     0,    19,     0,     0,     0,     0,    14,
       0,    15,     0,    48,    49,    50,    51,    52,    53,    16,
      17,    54,     0,     0,     0,     0,     0,     0,     0,     0,
      56,     0,     0,    18,     0,     0,    58,    59,    60,    61,
      62,    63,     0,    19
};

static const yytype_int16 yycheck[] =
{
       1,     1,     3,     4,    40,    85,   201,   137,   122,   104,
     151,   221,    34,    45,     3,   145,   226,   221,   290,   103,
      55,    25,    26,     1,    22,     1,    55,     8,    29,    55,
     103,    65,    52,    34,    95,    96,    53,   110,    25,    26,
      76,   101,    31,    65,    45,   103,    47,   107,   109,    38,
      85,    85,   110,    89,    55,   103,    57,    69,   330,    71,
      23,    24,   103,   143,    65,    54,    69,    56,    71,    70,
     265,   151,    69,    95,    71,    64,    65,   106,   107,   105,
     106,    82,   292,   103,    85,    27,    28,    69,   110,    78,
      91,   221,   104,   129,    95,   225,   300,   105,   106,    88,
     100,   104,   182,   313,    56,   235,   110,   317,   238,   110,
       3,   133,   105,   111,   101,    56,   151,   151,    50,   249,
     107,   110,   100,   253,   100,   103,    78,   103,   400,   130,
      93,    94,   133,   101,   329,   100,    88,    78,    31,   107,
     101,   221,    35,   257,     0,    38,   107,    88,    41,   103,
     151,     1,   293,     3,   107,   250,   286,   110,   100,   101,
     102,    54,   105,    56,   101,     3,   108,   105,   372,   107,
     107,    64,    65,    95,    96,   385,    38,   387,     8,   101,
     101,    31,    38,   100,   206,    78,   107,   397,    38,   221,
     191,   101,    54,    31,   103,    88,   337,   107,    54,   200,
      38,   202,    64,    65,    54,   206,    56,   302,    64,   210,
     246,   104,   104,   293,    64,    65,    54,   331,    56,     3,
     221,     5,     6,     7,     8,   226,    64,    65,    78,    31,
      32,   105,   106,   107,   429,   236,   104,    39,    88,    41,
      78,    76,   437,    27,    28,    27,    28,    31,   378,   101,
      88,   381,    25,    26,    38,   107,   105,   337,   107,   339,
      97,    98,    99,    47,   105,   395,   107,    95,    96,   105,
      54,   107,    56,   200,   105,   202,     1,   299,     3,   108,
      64,    65,     3,    67,    68,   107,   308,    27,    28,   385,
     291,   387,   293,   100,    78,    79,    19,    20,   299,   300,
     105,   380,    78,   382,    88,    89,    31,   308,   105,   105,
      31,    95,    96,    38,    31,   110,   100,    38,   100,   101,
     102,   105,     1,    45,     1,    21,   108,    90,    92,    54,
      91,    56,     8,    54,   105,    56,   337,   107,   103,    64,
      65,   101,   108,    64,    65,   107,   104,   383,   384,   105,
     104,   100,   100,    78,   376,   105,   101,    78,   104,   106,
     100,   105,   102,    88,   105,   107,   367,    88,   108,   101,
     109,   372,   101,   107,   110,   376,   100,   110,   414,   107,
      22,   101,   418,   109,   385,   104,   387,   409,   101,   105,
     105,     1,   104,     3,   101,     5,     6,     7,     8,   103,
     105,   107,     5,   439,   440,   108,   442,   443,   409,   445,
     109,     5,    65,   437,   436,   145,   367,    27,    28,    29,
      30,    31,    32,    33,    34,    35,    36,    37,    38,    39,
      40,    41,    42,    65,    44,   436,    46,    47,   325,    49,
     409,    51,    52,   308,    54,   436,    56,    57,    58,    59,
      60,    61,    62,    63,    64,    65,    66,    67,    68,     5,
      70,   236,    72,    73,    74,    75,   221,    77,    78,    79,
      80,    81,    82,    83,    84,    85,    86,    87,    88,    89,
     391,   133,   254,   221,   317,    95,    96,   369,   206,   256,
     100,   260,   271,   103,   104,   105,     1,   275,     3,   252,
       5,     6,     7,     8,   255,     9,    10,    11,    12,    13,
      14,    15,    16,    17,    18,   330,   268,    -1,    -1,    -1,
      -1,    -1,    27,    28,    29,    30,    31,    32,    33,    34,
      35,    36,    37,    38,    39,    40,    41,    42,    -1,    44,
      -1,    46,    47,    -1,    49,    -1,    51,    52,    -1,    54,
      -1,    56,    57,    58,    59,    60,    61,    62,    63,    64,
      65,    66,    67,    68,    -1,    70,    -1,    72,    73,    74,
      75,    -1,    77,    78,    79,    80,    81,    82,    83,    84,
      85,    86,    87,    88,    89,    -1,    -1,    -1,    -1,    -1,
      95,    96,    -1,    -1,    -1,   100,    -1,    -1,   103,   104,
     105,     1,    -1,     3,    -1,     5,     6,     7,     8,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,    28,    -1,
      30,    31,    -1,    33,    -1,    35,    -1,    37,    38,    39,
      -1,    41,    42,    -1,    44,    -1,    -1,    47,    -1,    49,
      -1,    51,    52,    -1,    54,    -1,    56,    57,    -1,    -1,
      -1,    -1,    -1,    -1,    64,    65,    -1,    67,    68,    -1,
      70,    -1,    72,    73,    74,    -1,    -1,    77,    78,    79,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    87,    88,    89,
      -1,    -1,    -1,    -1,    -1,    95,    96,    -1,    -1,    -1,
     100,    -1,    -1,   103,   104,   105,     1,    -1,     3,    -1,
       5,     6,     7,     8,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    27,    28,    -1,    30,    31,    -1,    33,    -1,
      35,    -1,    37,    38,    39,    -1,    41,    42,    -1,    44,
      -1,    -1,    47,    -1,    49,    -1,    51,    52,    -1,    54,
      -1,    56,    57,    -1,    -1,     1,    -1,     3,    -1,    64,
      65,    -1,    67,    68,    -1,    70,    -1,    72,    73,    74,
      -1,    -1,    77,    78,    79,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    87,    88,    89,    31,    -1,    -1,    -1,    -1,
      95,    96,    38,    -1,    -1,   100,    -1,    -1,   103,     3,
     105,     5,     6,     7,     8,    -1,    -1,    -1,    54,    -1,
      56,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    64,    65,
      -1,    -1,    -1,    27,    28,    29,    -1,    31,    32,    -1,
      34,    -1,    78,    -1,    38,    -1,    40,    -1,    42,    -1,
      -1,    -1,    88,    47,    -1,    -1,    -1,    -1,    -1,    -1,
      54,    -1,    56,    57,    58,    59,    60,    61,    62,    63,
      64,    65,    66,    67,    68,    -1,    -1,    -1,    72,    -1,
      74,    75,    -1,    -1,    78,    79,    -1,    81,    82,    83,
      84,    85,    86,    87,    88,    89,    -1,    -1,    -1,    -1,
      -1,    95,    96,    -1,    -1,     3,   100,     5,     6,     7,
       8,   105,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,
      28,    -1,    -1,    31,    -1,    -1,    -1,    -1,    -1,    -1,
      38,    -1,    -1,    -1,    42,    -1,    -1,    -1,    -1,    47,
      -1,    -1,    -1,    -1,    -1,    -1,    54,    -1,    56,    57,
      -1,    -1,    -1,    -1,    -1,    -1,    64,    65,    -1,    67,
      68,    -1,    -1,    -1,    72,     3,    74,    -1,    -1,    -1,
      78,    79,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    87,
      88,    89,    -1,    -1,    -1,    -1,    -1,    95,    96,    -1,
      -1,    29,   100,    31,    32,    -1,    34,   105,    36,    -1,
      38,    -1,    40,    -1,    -1,    -1,    -1,    -1,    46,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    54,    -1,    56,    -1,
      58,    59,    60,    61,    62,    63,    64,    65,    66,    -1,
       3,    69,    -1,    71,    -1,    -1,    -1,    75,    -1,    -1,
      78,    -1,    80,    81,    82,    83,    84,    85,    86,    -1,
      88,    -1,    -1,    -1,    -1,    -1,    29,    -1,    31,    32,
      -1,    34,    -1,    36,    -1,    38,   104,    40,    -1,    -1,
      -1,    -1,    -1,    46,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    54,    -1,    56,    -1,    58,    59,    60,    61,    62,
      63,    64,    65,    66,    -1,     3,    69,    -1,    71,    -1,
      -1,    -1,    75,    -1,    -1,    78,    -1,    80,    81,    82,
      83,    84,    85,    86,    -1,    88,    -1,    -1,    -1,    -1,
      -1,    29,    -1,    31,    32,    -1,    34,    -1,    -1,    -1,
      38,   104,    40,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    54,    -1,    56,    -1,
      58,    59,    60,    61,    62,    63,    64,    65,    66,    -1,
       3,    -1,    -1,    -1,    -1,    -1,    -1,    75,    -1,    -1,
      78,    -1,    -1,    81,    82,    83,    84,    85,    86,    -1,
      88,    -1,    -1,    -1,    -1,    -1,    29,    -1,    31,    32,
      -1,    34,    -1,    -1,    -1,    38,   104,    40,    -1,    -1,
      -1,    -1,     3,    -1,     5,     6,     7,     8,    -1,    -1,
      -1,    54,    -1,    56,    -1,    58,    59,    60,    61,    62,
      63,    64,    65,    66,    -1,    -1,    27,    28,    -1,    -1,
      31,    -1,    75,    -1,    -1,    78,    -1,    38,    81,    82,
      83,    84,    85,    86,    -1,    88,    47,     3,    -1,     5,
       6,     7,     8,    54,    -1,    56,    -1,    -1,    -1,    -1,
     103,    -1,    -1,    64,    65,    -1,    67,    68,    -1,    -1,
      -1,    27,    28,    -1,    -1,    31,    -1,    78,    79,    -1,
      -1,    -1,    38,    -1,    -1,    -1,    42,    88,    89,    -1,
      -1,    47,    -1,    -1,    95,    96,    -1,    -1,    54,   100,
      56,    57,   103,    -1,    -1,    -1,    -1,    -1,    64,    65,
      -1,    67,    68,    -1,    -1,    -1,    72,    -1,    74,    -1,
      -1,    -1,    78,    79,     3,    -1,     5,     6,     7,     8,
      -1,    87,    88,    89,    -1,    -1,    -1,    -1,    -1,    95,
      96,    -1,    -1,    -1,   100,   101,    -1,    -1,    27,    28,
      -1,    -1,    31,    -1,    -1,    -1,    -1,    -1,    -1,    38,
      -1,    -1,    -1,    42,    -1,    -1,    -1,    -1,    47,    -1,
      -1,    -1,    -1,    -1,    -1,    54,    -1,    56,    57,    -1,
      -1,    -1,    -1,    -1,    -1,    64,    65,    -1,    67,    68,
      -1,    -1,    -1,    72,    -1,    74,    -1,    -1,    -1,    78,
      79,     3,    -1,     5,     6,     7,     8,    -1,    87,    88,
      89,    -1,    -1,    -1,    -1,    -1,    95,    96,    -1,    -1,
      -1,   100,   101,    -1,    -1,    27,    28,    29,    -1,    31,
      32,    -1,    34,    -1,    -1,    -1,    38,    -1,    40,    -1,
      42,    -1,    -1,    -1,    -1,    47,    -1,    -1,    -1,    -1,
      -1,    -1,    54,    -1,    56,    57,    58,    59,    60,    61,
      62,    63,    64,    65,    -1,    67,    68,    -1,    -1,    -1,
      72,    -1,    74,    75,    -1,    -1,    78,    79,    -1,    81,
      82,    83,    84,    85,    86,    87,    88,    89,    -1,    -1,
      -1,    -1,    -1,    95,    96,    -1,    -1,     3,   100,     5,
       6,     7,     8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    27,    28,    -1,    -1,    31,    -1,    -1,    -1,    -1,
      -1,    -1,    38,    -1,    -1,    -1,    42,    -1,    -1,    -1,
      -1,    47,    -1,    -1,    -1,    -1,    -1,    -1,    54,     3,
      56,    57,    -1,    -1,    -1,    -1,    -1,    -1,    64,    65,
      -1,    67,    68,    -1,    -1,    -1,    72,    -1,    74,    -1,
      -1,    -1,    78,    79,    -1,    29,    -1,    31,    32,    -1,
      34,    87,    88,    89,    38,    -1,    40,    -1,    -1,    95,
      96,    -1,    -1,    -1,   100,    -1,    -1,    -1,    -1,    -1,
      54,    -1,    56,    -1,    58,    59,    60,    61,    62,    63,
      64,    65,    66,     3,    -1,     5,     6,     7,     8,    -1,
      -1,    75,    -1,    -1,    78,    -1,    -1,    81,    82,    83,
      84,    85,    86,    -1,    88,    -1,    -1,    27,    28,    -1,
      -1,    31,    -1,    -1,    -1,    -1,   100,    -1,    38,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,     3,    47,     5,     6,
       7,     8,    -1,    -1,    54,    -1,    56,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    64,    65,    -1,    67,    68,    -1,
      27,    28,    -1,    -1,    31,    -1,    -1,    -1,    78,    79,
      -1,    38,    -1,    -1,    -1,    -1,    -1,    -1,    88,    89,
      47,    -1,    -1,    -1,    -1,    95,    96,    54,    -1,    56,
     100,     3,    -1,     5,     6,     7,     8,    64,    65,    -1,
      67,    68,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    78,    79,    -1,    -1,    27,    28,    -1,    -1,    31,
      -1,    88,    89,    -1,    -1,    -1,    38,    -1,    95,    96,
      -1,    -1,    -1,   100,     3,    47,     5,     6,     7,     8,
      -1,    -1,    54,    -1,    56,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    64,    65,    -1,    67,    68,    -1,    -1,    -1,
      -1,    -1,    31,    -1,    -1,    -1,    78,    79,    -1,    38,
      -1,    -1,    -1,    -1,    -1,    -1,    88,    89,    47,    -1,
      -1,    -1,    -1,    95,    96,    54,    -1,    56,   100,     3,
      -1,    -1,    -1,    -1,    -1,    64,    65,    -1,    67,    68,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    78,
      79,    -1,    -1,    -1,    -1,    29,    -1,    31,    32,    88,
      34,    -1,    36,    -1,    38,    -1,    40,    -1,     3,    -1,
      -1,   100,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      54,    -1,    56,    -1,    58,    59,    60,    61,    62,    63,
      64,    65,    66,    -1,    29,    69,    31,    32,    -1,    34,
      -1,    75,    -1,    38,    78,    40,    -1,    81,    82,    83,
      84,    85,    86,    -1,    88,    -1,    -1,    -1,    -1,    54,
      -1,    56,    -1,    58,    59,    60,    61,    62,    63,    64,
      65,    66,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      75,    -1,    -1,    78,    -1,    -1,    81,    82,    83,    84,
      85,    86,    -1,    88
};

  /* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
     symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    38,    54,    64,    65,   113,   114,   115,   116,   149,
     151,     3,    31,    38,    54,    56,    64,    65,    78,    88,
     203,   204,     8,   204,   204,     0,   115,   116,   149,    53,
     117,   103,   103,   204,   103,    69,    71,   142,   143,   150,
     142,   152,    29,    32,    34,    36,    40,    46,    58,    59,
      60,    61,    62,    63,    66,    71,    75,    80,    81,    82,
      83,    84,    85,    86,   104,   118,   119,   120,   121,   122,
     123,   124,   129,   130,   132,   136,   137,   138,   142,   204,
      69,   105,    50,   104,   142,   103,   110,   147,   104,   142,
     121,   123,   204,     1,   204,   100,   120,     1,   204,   104,
     119,   129,   136,   105,   106,     1,   125,   126,   204,   147,
     100,   204,   105,     1,     5,     6,     7,     8,    27,    28,
      30,    33,    35,    37,    39,    41,    42,    44,    47,    49,
      51,    52,    57,    67,    68,    70,    72,    73,    74,    77,
      79,    87,    89,    95,    96,   100,   104,   105,   120,   129,
     147,   148,   153,   154,   155,   156,   158,   159,   160,   163,
     166,   167,   168,   169,   170,   171,   172,   173,   174,   179,
     180,   181,   182,   183,   184,   185,   187,   189,   191,   193,
     195,   196,   197,   198,   201,   202,   204,     8,   110,   147,
     104,   103,   123,   104,    76,   128,   172,   196,   204,   105,
     107,   108,   123,   139,   140,   141,   100,   105,    78,   105,
      67,   185,   105,    56,    78,    88,   110,    56,    78,    88,
       1,   100,   147,   204,     1,   100,   103,   123,   175,   176,
     204,    31,   105,   171,     1,   100,   103,   164,   100,   196,
     124,   171,   104,   120,   129,   153,    45,     1,   105,   106,
     107,     1,    22,   111,    21,    90,    91,    92,    25,    26,
     186,    23,    24,    93,    94,   188,    95,    96,   190,    19,
      20,   192,    97,    98,    99,   194,     9,    10,    11,    12,
      13,    14,    15,    16,    17,    18,   157,   196,    27,    28,
     100,   102,   108,   110,   105,     8,   133,   134,   204,   107,
     103,   105,   107,   126,   127,   189,   126,   101,   107,   139,
     204,   110,   104,   121,   122,   154,   156,   161,   171,   179,
     204,   105,   104,   171,   176,   177,   178,   179,   204,   108,
     100,   100,   105,   104,   171,   104,   158,   165,   204,   101,
     101,    52,   147,   105,   171,   172,   105,   181,   171,   182,
     183,   184,   185,   187,   189,   191,   193,   195,   171,   174,
     199,   200,   204,   179,   153,   204,   105,   107,   135,   106,
     123,   122,   131,   172,   109,   141,   101,   105,   162,   179,
      55,   162,    55,   101,   101,   107,   135,   110,   189,   199,
     185,   101,   104,   153,   196,   100,   105,   110,   105,   101,
     107,   109,   134,   104,   180,   101,   105,   104,   122,   100,
     123,   144,   145,   146,   101,   171,   105,   198,   101,   171,
     198,   147,   147,   178,   104,   178,   109,   101,   101,   107,
     164,   171,   179,   174,   105,   145,   107,   108,   147,   101,
     101,   147,   101,   101,   189,   101,   101,   146,   127,   147,
     147,   147,   147,   101,   147,   109
};

  /* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint8 yyr1[] =
{
       0,   112,   113,   113,   114,   114,   114,   114,   114,   114,
     115,   116,   116,   117,   117,   118,   118,   118,   118,   118,
     118,   119,   119,   119,   120,   120,   121,   121,   122,   123,
     123,   123,   124,   124,   124,   124,   124,   124,   124,   124,
     124,   124,   124,   124,   124,   124,   124,   124,   124,   125,
     125,   126,   126,   127,   127,   128,   128,   129,   129,   130,
     130,   131,   131,   132,   132,   133,   133,   134,   134,   135,
     135,   136,   137,   137,   138,   139,   139,   140,   140,   141,
     142,   143,   143,   143,   143,   144,   144,   144,   145,   145,
     146,   146,   147,   147,   148,   148,   148,   148,   148,   148,
     149,   150,   150,   151,   152,   152,   152,   152,   153,   153,
     153,   153,   153,   153,   153,   153,   153,   153,   153,   154,
     155,   155,   156,   156,   157,   157,   157,   157,   157,   157,
     157,   157,   157,   157,   158,   158,   158,   159,   159,   159,
     159,   160,   160,   160,   160,   160,   160,   160,   160,   160,
     161,   161,   162,   162,   163,   163,   163,   164,   164,   165,
     165,   166,   166,   166,   166,   166,   167,   168,   168,   168,
     168,   168,   168,   168,   169,   170,   170,   171,   171,   172,
     172,   173,   173,   173,   173,   173,   174,   174,   174,   175,
     175,   176,   177,   177,   178,   178,   178,   179,   179,   180,
     180,   181,   181,   182,   182,   183,   183,   184,   184,   185,
     185,   186,   186,   187,   187,   188,   188,   188,   188,   189,
     189,   190,   190,   191,   191,   192,   192,   193,   193,   194,
     194,   194,   195,   195,   196,   196,   196,   197,   197,   197,
     197,   198,   198,   198,   198,   198,   198,   199,   199,   200,
     200,   201,   201,   201,   201,   202,   202,   202,   202,   202,
     202,   202,   203,   203,   203,   203,   203,   203,   203,   203,
     204,   204
};

  /* YYR2[YYN] -- Number of symbols on the right hand side of rule YYN.  */
static const yytype_uint8 yyr2[] =
{
       0,     2,     1,     1,     1,     1,     1,     2,     2,     2,
       2,     5,     6,     0,     2,     1,     1,     1,     2,     2,
       2,     1,     2,     3,     1,     2,     2,     4,     2,     1,
       1,     6,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       3,     1,     4,     0,     1,     1,     3,     1,     1,     6,
       3,     2,     3,     6,     3,     1,     3,     1,     3,     0,
       1,     2,     1,     1,     4,     0,     1,     1,     3,     2,
       7,     0,     1,     1,     2,     0,     3,     1,     1,     3,
       1,     4,     2,     3,     1,     1,     1,     2,     2,     2,
       5,     2,     3,     5,     4,     2,     5,     3,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       2,     2,     4,     4,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     3,     3,     2,     5,     7,     3,
       3,     2,     5,     6,     7,     6,     7,     7,     7,     3,
       1,     1,     1,     2,     2,     5,     3,     2,     3,     1,
       2,     2,     2,     2,     3,     3,     3,     2,     2,     2,
       2,     2,     2,     2,     1,     5,     7,     1,     3,     1,
       3,     1,     1,     1,     1,     1,     1,     2,     2,     1,
       4,     4,     1,     3,     1,     1,     3,     1,     5,     1,
       3,     1,     3,     1,     3,     1,     3,     1,     3,     1,
       3,     1,     1,     1,     3,     1,     1,     1,     1,     1,
       3,     1,     1,     1,     3,     1,     1,     1,     3,     1,
       1,     1,     1,     4,     1,     2,     2,     1,     1,     1,
       1,     1,     4,     4,     3,     2,     2,     0,     1,     1,
       3,     1,     1,     3,     5,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1
};


#define yyerrok         (yyerrstatus = 0)
#define yyclearin       (yychar = YYEMPTY)
#define YYEMPTY         (-2)
#define YYEOF           0

#define YYACCEPT        goto yyacceptlab
#define YYABORT         goto yyabortlab
#define YYERROR         goto yyerrorlab


#define YYRECOVERING()  (!!yyerrstatus)

#define YYBACKUP(Token, Value)                                    \
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
static int
yy_location_print_ (FILE *yyo, YYLTYPE const * const yylocp)
{
  int res = 0;
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


/*-----------------------------------.
| Print this symbol's value on YYO.  |
`-----------------------------------*/

static void
yy_symbol_value_print (FILE *yyo, int yytype, YYSTYPE const * const yyvaluep, YYLTYPE const * const yylocationp, parse_t *parse, void *yyscanner)
{
  FILE *yyoutput = yyo;
  YYUSE (yyoutput);
  YYUSE (yylocationp);
  YYUSE (parse);
  YYUSE (yyscanner);
  if (!yyvaluep)
    return;
# ifdef YYPRINT
  if (yytype < YYNTOKENS)
    YYPRINT (yyo, yytoknum[yytype], *yyvaluep);
# endif
  YYUSE (yytype);
}


/*---------------------------.
| Print this symbol on YYO.  |
`---------------------------*/

static void
yy_symbol_print (FILE *yyo, int yytype, YYSTYPE const * const yyvaluep, YYLTYPE const * const yylocationp, parse_t *parse, void *yyscanner)
{
  YYFPRINTF (yyo, "%s %s (",
             yytype < YYNTOKENS ? "token" : "nterm", yytname[yytype]);

  YY_LOCATION_PRINT (yyo, *yylocationp);
  YYFPRINTF (yyo, ": ");
  yy_symbol_value_print (yyo, yytype, yyvaluep, yylocationp, parse, yyscanner);
  YYFPRINTF (yyo, ")");
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
  unsigned long yylno = yyrline[yyrule];
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
                       &yyvsp[(yyi + 1) - (yynrhs)]
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
            else
              goto append;

          append:
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

  return (YYSIZE_T) (yystpcpy (yyres, yystr) - yyres);
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
                  if (yysize <= yysize1 && yysize1 <= YYSTACK_ALLOC_MAXIMUM)
                    yysize = yysize1;
                  else
                    return 2;
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
    if (yysize <= yysize1 && yysize1 <= YYSTACK_ALLOC_MAXIMUM)
      yysize = yysize1;
    else
      return 2;
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
#line 36 "grammar.y" /* yacc.c:1431  */
{
    src_pos_init(&yylloc, parse->src, parse->path);
}

#line 1939 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1431  */
  yylsp[0] = yylloc;
  goto yysetstate;


/*------------------------------------------------------------.
| yynewstate -- push a new state, which is found in yystate.  |
`------------------------------------------------------------*/
yynewstate:
  /* In all cases, when you get here, the value and location stacks
     have just been pushed.  So pushing a state here evens the stacks.  */
  yyssp++;


/*--------------------------------------------------------------------.
| yynewstate -- set current state (the top of the stack) to yystate.  |
`--------------------------------------------------------------------*/
yysetstate:
  *yyssp = (yytype_int16) yystate;

  if (yyss + yystacksize - 1 <= yyssp)
#if !defined yyoverflow && !defined YYSTACK_RELOCATE
    goto yyexhaustedlab;
#else
    {
      /* Get the current used size of the three stacks, in elements.  */
      YYSIZE_T yysize = (YYSIZE_T) (yyssp - yyss + 1);

# if defined yyoverflow
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
        yyss = yyss1;
        yyvs = yyvs1;
        yyls = yyls1;
      }
# else /* defined YYSTACK_RELOCATE */
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
# undef YYSTACK_RELOCATE
        if (yyss1 != yyssa)
          YYSTACK_FREE (yyss1);
      }
# endif

      yyssp = yyss + yysize - 1;
      yyvsp = yyvs + yysize - 1;
      yylsp = yyls + yysize - 1;

      YYDPRINTF ((stderr, "Stack size increased to %lu\n",
                  (unsigned long) yystacksize));

      if (yyss + yystacksize - 1 <= yyssp)
        YYABORT;
    }
#endif /* !defined yyoverflow && !defined YYSTACK_RELOCATE */

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
| yyreduce -- do a reduction.  |
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
#line 280 "grammar.y" /* yacc.c:1652  */
    {
        ast_blk_t *blk = blk_new_contract(&(yylsp[-1]));

        /* add default constructor */
        id_add(&blk->ids, id_new_ctor((yyvsp[-3].str), NULL, NULL, &(yylsp[-3])));

        id_add(&ROOT->ids, id_new_contract((yyvsp[-3].str), (yyvsp[-2].exp), blk, &(yyloc)));
    }
#line 2140 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 12:
#line 289 "grammar.y" /* yacc.c:1652  */
    {
        if (is_empty_vector(&(yyvsp[-1].blk)->ids) || !is_ctor_id(vector_get_id(&(yyvsp[-1].blk)->ids, 0)))
            /* add default constructor */
            id_add(&(yyvsp[-1].blk)->ids, id_new_ctor((yyvsp[-4].str), NULL, NULL, &(yylsp[-4])));

        id_add(&ROOT->ids, id_new_contract((yyvsp[-4].str), (yyvsp[-3].exp), (yyvsp[-1].blk), &(yyloc)));
    }
#line 2152 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 13:
#line 299 "grammar.y" /* yacc.c:1652  */
    { (yyval.exp) = NULL; }
#line 2158 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 14:
#line 301 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yylsp[0]));
    }
#line 2166 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 15:
#line 308 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2175 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 16:
#line 313 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2184 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 17:
#line 318 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2193 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 18:
#line 323 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2202 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 19:
#line 328 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2211 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 20:
#line 333 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2220 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 22:
#line 342 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_PUBLIC;
    }
#line 2229 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 23:
#line 347 "grammar.y" /* yacc.c:1652  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2238 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 25:
#line 356 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_CONST;
    }
#line 2247 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 27:
#line 365 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.dflt_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.dflt_exp = (yyvsp[-1].exp);
    }
#line 2260 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 28:
#line 377 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = (yyvsp[0].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.type_exp = (yyvsp[-1].exp);
    }
#line 2273 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 29:
#line 389 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_type((yyvsp[0].type), &(yylsp[0]));
    }
#line 2281 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 30:
#line 393 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_type(TYPE_NONE, &(yylsp[0]));

        (yyval.exp)->u_type.name = (yyvsp[0].str);
    }
#line 2291 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 31:
#line 399 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_type(TYPE_MAP, &(yylsp[-5]));

        (yyval.exp)->u_type.k_exp = (yyvsp[-3].exp);
        (yyval.exp)->u_type.v_exp = (yyvsp[-1].exp);
    }
#line 2302 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 32:
#line 408 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_ACCOUNT; }
#line 2308 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 33:
#line 409 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_BOOL; }
#line 2314 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 34:
#line 410 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_UINT8; }
#line 2320 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 35:
#line 411 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_INT8; }
#line 2326 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 36:
#line 412 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_INT16; }
#line 2332 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 37:
#line 413 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_INT32; }
#line 2338 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 38:
#line 414 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_INT64; }
#line 2344 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 39:
#line 415 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_INT128; }
#line 2350 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 40:
#line 416 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_INT32; }
#line 2356 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 41:
#line 417 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_UINT8; }
#line 2362 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 42:
#line 418 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_UINT16; }
#line 2368 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 43:
#line 419 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_UINT32; }
#line 2374 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 44:
#line 420 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_UINT64; }
#line 2380 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 45:
#line 421 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_UINT128; }
#line 2386 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 46:
#line 422 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_UINT32; }
#line 2392 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 47:
#line 425 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_STRING; }
#line 2398 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 48:
#line 426 "grammar.y" /* yacc.c:1652  */
    { (yyval.type) = TYPE_CURSOR; }
#line 2404 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 50:
#line 432 "grammar.y" /* yacc.c:1652  */
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
#line 2420 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 51:
#line 447 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PRIVATE, &(yylsp[0]));
    }
#line 2428 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 52:
#line 451 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2441 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 53:
#line 462 "grammar.y" /* yacc.c:1652  */
    { (yyval.exp) = exp_new_null(&(yyloc)); }
#line 2447 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 55:
#line 468 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 2455 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 56:
#line 472 "grammar.y" /* yacc.c:1652  */
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
#line 2470 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 59:
#line 491 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = id_new_struct((yyvsp[-4].str), (yyvsp[-1].vect), &(yyloc));
    }
#line 2478 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 60:
#line 495 "grammar.y" /* yacc.c:1652  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2487 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 61:
#line 503 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = vector_new();

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2500 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 62:
#line 512 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = (yyvsp[-2].vect);

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2513 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 63:
#line 524 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = id_new_enum((yyvsp[-4].str), (yyvsp[-2].vect), &(yyloc));
    }
#line 2521 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 64:
#line 528 "grammar.y" /* yacc.c:1652  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2530 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 65:
#line 536 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = vector_new();
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2539 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 66:
#line 541 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2548 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 67:
#line 549 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PUBLIC | MOD_CONST, &(yylsp[0]));
    }
#line 2556 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 68:
#line 553 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = id_new_var((yyvsp[-2].str), MOD_PUBLIC | MOD_CONST, &(yylsp[-2]));
        (yyval.id)->u_var.dflt_exp = (yyvsp[0].exp);
    }
#line 2565 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 71:
#line 566 "grammar.y" /* yacc.c:1652  */
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
#line 2582 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 74:
#line 587 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = id_new_ctor((yyvsp[-3].str), (yyvsp[-1].vect), NULL, &(yylsp[-3]));
    }
#line 2590 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 75:
#line 593 "grammar.y" /* yacc.c:1652  */
    { (yyval.vect) = NULL; }
#line 2596 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 77:
#line 599 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2605 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 78:
#line 604 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2614 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 79:
#line 612 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->u_var.is_param = true;
        (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
    }
#line 2624 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 80:
#line 621 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = id_new_func((yyvsp[-4].str), (yyvsp[-6].mod), (yyvsp[-2].vect), (yyvsp[0].id), NULL, &(yylsp[-4]));
    }
#line 2632 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 81:
#line 627 "grammar.y" /* yacc.c:1652  */
    { (yyval.mod) = MOD_PRIVATE; }
#line 2638 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 82:
#line 628 "grammar.y" /* yacc.c:1652  */
    { (yyval.mod) = MOD_PUBLIC; }
#line 2644 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 83:
#line 629 "grammar.y" /* yacc.c:1652  */
    { (yyval.mod) = MOD_PAYABLE; }
#line 2650 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 84:
#line 630 "grammar.y" /* yacc.c:1652  */
    { (yyval.mod) = MOD_PUBLIC | MOD_PAYABLE; }
#line 2656 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 85:
#line 634 "grammar.y" /* yacc.c:1652  */
    { (yyval.id) = NULL; }
#line 2662 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 86:
#line 635 "grammar.y" /* yacc.c:1652  */
    { (yyval.id) = (yyvsp[-1].id); }
#line 2668 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 89:
#line 642 "grammar.y" /* yacc.c:1652  */
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
#line 2698 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 90:
#line 671 "grammar.y" /* yacc.c:1652  */
    {
        /* We wanted to use a type expression, but we can not store size expressions
         * and declare it as an identifier. */
        (yyval.id) = id_new_param(NULL, (yyvsp[0].exp), &(yylsp[0]));
    }
#line 2708 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 91:
#line 677 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2721 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 92:
#line 688 "grammar.y" /* yacc.c:1652  */
    { (yyval.blk) = NULL; }
#line 2727 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 93:
#line 689 "grammar.y" /* yacc.c:1652  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 2733 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 94:
#line 694 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2742 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 95:
#line 699 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        /* Unlike state variables, local variables are referenced according to their
         * order of declaration. */
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2753 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 96:
#line 706 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2762 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 97:
#line 711 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2771 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 98:
#line 716 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2780 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 99:
#line 721 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2789 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 100:
#line 729 "grammar.y" /* yacc.c:1652  */
    {
        id_add(&ROOT->ids, id_new_interface((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2797 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 101:
#line 736 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_interface(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2806 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 102:
#line 741 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-2].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2815 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 103:
#line 749 "grammar.y" /* yacc.c:1652  */
    {
        id_add(&ROOT->ids, id_new_library((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2823 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 104:
#line 756 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-3].id)->mod = MOD_SYSTEM;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2836 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 105:
#line 765 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-1].id)->mod = MOD_SYSTEM;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2849 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 106:
#line 774 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-4].blk);

        (yyvsp[-3].id)->mod = MOD_SYSTEM;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2862 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 107:
#line 783 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-2].blk);

        (yyvsp[-1].id)->mod = MOD_SYSTEM;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2875 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 119:
#line 809 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_null(&(yyloc));
    }
#line 2883 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 120:
#line 816 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_exp((yyvsp[-1].exp), &(yyloc));
    }
#line 2891 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 121:
#line 820 "grammar.y" /* yacc.c:1652  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2900 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 122:
#line 828 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2]));
    }
#line 2908 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 123:
#line 832 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), exp_new_binary((yyvsp[-2].op), (yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2])), &(yylsp[-2]));
    }
#line 2916 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 124:
#line 838 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_ADD; }
#line 2922 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 125:
#line 839 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_SUB; }
#line 2928 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 126:
#line 840 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_MUL; }
#line 2934 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 127:
#line 841 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_DIV; }
#line 2940 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 128:
#line 842 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_MOD; }
#line 2946 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 129:
#line 843 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_BIT_AND; }
#line 2952 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 130:
#line 844 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_BIT_XOR; }
#line 2958 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 131:
#line 845 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_BIT_OR; }
#line 2964 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 132:
#line 846 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_RSHIFT; }
#line 2970 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 133:
#line 847 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_LSHIFT; }
#line 2976 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 134:
#line 852 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = (yyvsp[0].stmt);
        id_add(LABELS, id_new_label((yyvsp[-2].str), (yyvsp[0].stmt), &(yylsp[-2])));
    }
#line 2985 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 135:
#line 857 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_case((yyvsp[-1].exp), &(yyloc));
    }
#line 2993 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 136:
#line 861 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_case(NULL, &(yyloc));
    }
#line 3001 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 137:
#line 868 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3009 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 138:
#line 872 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = (yyvsp[-6].stmt);
        stmt_add(&(yyval.stmt)->u_if.elif_stmts, stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yylsp[-5])));
    }
#line 3018 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 139:
#line 877 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = (yyvsp[-2].stmt);
        (yyval.stmt)->u_if.else_blk = (yyvsp[0].blk);
    }
#line 3027 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 140:
#line 882 "grammar.y" /* yacc.c:1652  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3036 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 141:
#line 890 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, NULL, NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3044 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 142:
#line 894 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3052 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 143:
#line 898 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-3].stmt), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3060 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 144:
#line 902 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-4].stmt), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3068 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 145:
#line 906 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-3].id), &(yylsp[-3])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3076 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 146:
#line 910 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3084 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 147:
#line 914 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_exp((yyvsp[-4].exp), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3092 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 148:
#line 918 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3100 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 149:
#line 922 "grammar.y" /* yacc.c:1652  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3109 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 152:
#line 934 "grammar.y" /* yacc.c:1652  */
    { (yyval.exp) = NULL; }
#line 3115 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 154:
#line 940 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_switch(NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3123 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 155:
#line 944 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_switch((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3131 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 156:
#line 948 "grammar.y" /* yacc.c:1652  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3140 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 157:
#line 955 "grammar.y" /* yacc.c:1652  */
    { (yyval.blk) = NULL; }
#line 3146 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 158:
#line 956 "grammar.y" /* yacc.c:1652  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 3152 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 159:
#line 961 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = blk_new_switch(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3161 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 160:
#line 966 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3170 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 161:
#line 974 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_CONTINUE, NULL, &(yyloc));
    }
#line 3178 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 162:
#line 978 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_BREAK, NULL, &(yyloc));
    }
#line 3186 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 163:
#line 982 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_return(NULL, &(yyloc));
    }
#line 3194 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 164:
#line 986 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_return((yyvsp[-1].exp), &(yyloc));
    }
#line 3202 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 165:
#line 990 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_goto((yyvsp[-1].str), &(yylsp[-1]));
    }
#line 3210 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 166:
#line 997 "grammar.y" /* yacc.c:1652  */
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
#line 3230 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 174:
#line 1026 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.stmt) = stmt_new_blk((yyvsp[0].blk), &(yyloc));
    }
#line 3238 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 175:
#line 1033 "grammar.y" /* yacc.c:1652  */
    {
        char *cond_str =
            xstrndup(parse->src + (yyloc).first_offset + (yylsp[-2]).last_col - 1, (yylsp[0]).first_col - (yylsp[-2]).last_col);

        (yyval.stmt) = stmt_new_pragma(PRAGMA_ASSERT, (yyvsp[-1].exp), cond_str, NULL, &(yylsp[-4]));
    }
#line 3249 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 176:
#line 1040 "grammar.y" /* yacc.c:1652  */
    {
        char *cond_str =
            xstrndup(parse->src + (yyloc).first_offset + (yylsp[-4]).last_col - 1, (yylsp[-2]).first_col - (yylsp[-4]).last_col);

        (yyval.stmt) = stmt_new_pragma(PRAGMA_ASSERT, (yyvsp[-3].exp), cond_str, (yyvsp[-1].exp), &(yylsp[-6]));
    }
#line 3260 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 178:
#line 1051 "grammar.y" /* yacc.c:1652  */
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
#line 3275 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 180:
#line 1066 "grammar.y" /* yacc.c:1652  */
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
#line 3295 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 181:
#line 1084 "grammar.y" /* yacc.c:1652  */
    { (yyval.sql) = SQL_DELETE; }
#line 3301 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 182:
#line 1085 "grammar.y" /* yacc.c:1652  */
    { (yyval.sql) = SQL_INSERT; }
#line 3307 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 183:
#line 1086 "grammar.y" /* yacc.c:1652  */
    { (yyval.sql) = SQL_REPLACE; }
#line 3313 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 184:
#line 1087 "grammar.y" /* yacc.c:1652  */
    { (yyval.sql) = SQL_QUERY; }
#line 3319 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 185:
#line 1088 "grammar.y" /* yacc.c:1652  */
    { (yyval.sql) = SQL_UPDATE; }
#line 3325 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 187:
#line 1094 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3333 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 188:
#line 1098 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = (yyvsp[0].exp);
        (yyval.exp)->u_init.is_outmost = true;
    }
#line 3342 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 189:
#line 1106 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_alloc((yyvsp[0].exp), &(yylsp[0]));
    }
#line 3350 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 190:
#line 1110 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = (yyvsp[-3].exp);

        if ((yyval.exp)->u_alloc.size_exps == NULL)
            (yyval.exp)->u_alloc.size_exps = vector_new();

        exp_add((yyval.exp)->u_alloc.size_exps, (yyvsp[-1].exp));
    }
#line 3363 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 191:
#line 1122 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_init((yyvsp[-2].vect), &(yyloc));
    }
#line 3371 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 192:
#line 1129 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3380 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 193:
#line 1134 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3389 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 196:
#line 1144 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3397 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 198:
#line 1152 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_ternary((yyvsp[-4].exp), (yyvsp[-2].exp), (yyvsp[0].exp), &(yyloc));
    }
#line 3405 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 200:
#line 1160 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary(OP_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3413 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 202:
#line 1168 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary(OP_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3421 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 204:
#line 1176 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3429 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 206:
#line 1184 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_XOR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3437 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 208:
#line 1192 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3445 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 210:
#line 1200 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3453 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 211:
#line 1206 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_EQ; }
#line 3459 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 212:
#line 1207 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_NE; }
#line 3465 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 214:
#line 1213 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3473 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 215:
#line 1219 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_LT; }
#line 3479 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 216:
#line 1220 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_GT; }
#line 3485 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 217:
#line 1221 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_LE; }
#line 3491 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 218:
#line 1222 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_GE; }
#line 3497 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 220:
#line 1228 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3505 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 221:
#line 1234 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_ADD; }
#line 3511 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 222:
#line 1235 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_SUB; }
#line 3517 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 224:
#line 1241 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3525 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 225:
#line 1247 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_RSHIFT; }
#line 3531 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 226:
#line 1248 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_LSHIFT; }
#line 3537 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 228:
#line 1254 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3545 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 229:
#line 1260 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_MUL; }
#line 3551 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 230:
#line 1261 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_DIV; }
#line 3557 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 231:
#line 1262 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_MOD; }
#line 3563 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 233:
#line 1268 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_cast((yyvsp[-2].type), (yyvsp[0].exp), &(yylsp[-2]));
    }
#line 3571 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 235:
#line 1276 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_unary((yyvsp[-1].op), true, (yyvsp[0].exp), &(yyloc));
    }
#line 3579 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 236:
#line 1280 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3587 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 237:
#line 1286 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_INC; }
#line 3593 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 238:
#line 1287 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_DEC; }
#line 3599 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 239:
#line 1288 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_NEG; }
#line 3605 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 240:
#line 1289 "grammar.y" /* yacc.c:1652  */
    { (yyval.op) = OP_NOT; }
#line 3611 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 242:
#line 1295 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_array((yyvsp[-3].exp), (yyvsp[-1].exp), &(yyloc));
    }
#line 3619 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 243:
#line 1299 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_call(false, (yyvsp[-3].exp), (yyvsp[-1].vect), &(yyloc));
    }
#line 3627 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 244:
#line 1303 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_access((yyvsp[-2].exp), exp_new_id((yyvsp[0].str), &(yylsp[0])), &(yyloc));
    }
#line 3635 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 245:
#line 1307 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_unary(OP_INC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3643 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 246:
#line 1311 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_unary(OP_DEC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3651 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 247:
#line 1317 "grammar.y" /* yacc.c:1652  */
    { (yyval.vect) = NULL; }
#line 3657 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 249:
#line 1323 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3666 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 250:
#line 1328 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3675 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 252:
#line 1337 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yyloc));
    }
#line 3683 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 253:
#line 1341 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = (yyvsp[-1].exp);
    }
#line 3691 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 254:
#line 1345 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_call(true, exp_new_id((yyvsp[-3].str), &(yylsp[-3])), (yyvsp[-1].vect), &(yylsp[-3]));
    }
#line 3699 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 255:
#line 1352 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_lit_null(&(yyloc));
    }
#line 3707 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 256:
#line 1356 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_lit_bool(true, &(yyloc));
    }
#line 3715 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 257:
#line 1360 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_lit_bool(false, &(yyloc));
    }
#line 3723 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 258:
#line 1364 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 10);
    }
#line 3732 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 259:
#line 1369 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 8);
    }
#line 3741 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 260:
#line 1374 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 0);
    }
#line 3750 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 261:
#line 1388 "grammar.y" /* yacc.c:1652  */
    {
        (yyval.exp) = exp_new_lit_str((yyvsp[0].str), &(yyloc));
    }
#line 3758 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 262:
#line 1394 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("assert"); }
#line 3764 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 263:
#line 1395 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("contract"); }
#line 3770 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 264:
#line 1396 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("import"); }
#line 3776 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 265:
#line 1397 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("index"); }
#line 3782 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 266:
#line 1398 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("interface"); }
#line 3788 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 267:
#line 1399 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("library"); }
#line 3794 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 268:
#line 1400 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("table"); }
#line 3800 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 269:
#line 1401 "grammar.y" /* yacc.c:1652  */
    { (yyval.str) = xstrdup("view"); }
#line 3806 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;

  case 270:
#line 1406 "grammar.y" /* yacc.c:1652  */
    {
        if (strlen((yyvsp[0].str)) > NAME_MAX_LEN)
            ERROR(ERROR_TOO_LONG_ID, &(yylsp[0]), NAME_MAX_LEN, strlen((yyvsp[0].str)));

        (yyval.str) = (yyvsp[0].str);
    }
#line 3817 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
    break;


#line 3821 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1652  */
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
  {
    const int yylhs = yyr1[yyn] - YYNTOKENS;
    const int yyi = yypgoto[yylhs] + *yyssp;
    yystate = (0 <= yyi && yyi <= YYLAST && yycheck[yyi] == *yyssp
               ? yytable[yyi]
               : yydefgoto[yylhs]);
  }

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
  /* Pacify compilers when the user code never invokes YYERROR and the
     label yyerrorlab therefore never appears in user code.  */
  if (0)
    YYERROR;

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


/*-----------------------------------------------------.
| yyreturn -- parsing is finished, return the result.  |
`-----------------------------------------------------*/
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
#line 1415 "grammar.y" /* yacc.c:1918  */


static void
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
