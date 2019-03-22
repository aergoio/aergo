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
    K_CURSOR = 294,
    K_DEFAULT = 295,
    K_DELETE = 296,
    K_DOUBLE = 297,
    K_DROP = 298,
    K_ELSE = 299,
    K_ENUM = 300,
    K_FALSE = 301,
    K_FLOAT = 302,
    K_FOR = 303,
    K_FUNC = 304,
    K_GOTO = 305,
    K_IF = 306,
    K_IMPLEMENTS = 307,
    K_IMPORT = 308,
    K_IN = 309,
    K_INDEX = 310,
    K_INSERT = 311,
    K_INT = 312,
    K_INT8 = 313,
    K_INT16 = 314,
    K_INT32 = 315,
    K_INT64 = 316,
    K_INT128 = 317,
    K_INTERFACE = 318,
    K_LIBRARY = 319,
    K_MAP = 320,
    K_NEW = 321,
    K_NULL = 322,
    K_PAYABLE = 323,
    K_PUBLIC = 324,
    K_REPLACE = 325,
    K_RETURN = 326,
    K_SELECT = 327,
    K_STRING = 328,
    K_STRUCT = 329,
    K_SWITCH = 330,
    K_TABLE = 331,
    K_TRUE = 332,
    K_TYPE = 333,
    K_UINT = 334,
    K_UINT8 = 335,
    K_UINT16 = 336,
    K_UINT32 = 337,
    K_UINT64 = 338,
    K_UINT128 = 339,
    K_UPDATE = 340,
    K_VIEW = 341
  };
#endif

/* Value type.  */
#if ! defined YYSTYPE && ! defined YYSTYPE_IS_DECLARED

union YYSTYPE
{
#line 152 "grammar.y" /* yacc.c:355  */

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

#line 239 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:355  */
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

#line 269 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:358  */

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
#define YYFINAL  24
/* YYLAST -- Last index in YYTABLE.  */
#define YYLAST   1883

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  110
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  92
/* YYNRULES -- Number of rules.  */
#define YYNRULES  268
/* YYNSTATES -- Number of states.  */
#define YYNSTATES  447

/* YYTRANSLATE[YYX] -- Symbol number corresponding to YYX as returned
   by yylex, with out-of-bounds checking.  */
#define YYUNDEFTOK  2
#define YYMAXUTOK   341

#define YYTRANSLATE(YYX)                                                \
  ((unsigned int) (YYX) <= YYMAXUTOK ? yytranslate[YYX] : YYUNDEFTOK)

/* YYTRANSLATE[TOKEN-NUM] -- Symbol number corresponding to TOKEN-NUM
   as returned by yylex, without out-of-bounds checking.  */
static const yytype_uint8 yytranslate[] =
{
       0,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    87,     2,     2,     2,    97,    90,     2,
      98,    99,    95,    93,   105,    94,   100,    96,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,   108,   103,
      91,   104,    92,   109,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,   106,     2,   107,    89,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,   101,    88,   102,     2,     2,     2,     2,
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
      85,    86
};

#if YYDEBUG
  /* YYRLINE[YYN] -- Source line where rule number YYN was defined.  */
static const yytype_uint16 yyrline[] =
{
       0,   258,   258,   259,   263,   264,   265,   266,   267,   268,
     272,   276,   285,   304,   305,   312,   317,   322,   327,   332,
     337,   345,   346,   351,   359,   360,   368,   369,   381,   393,
     397,   403,   413,   414,   415,   416,   417,   418,   419,   420,
     421,   422,   423,   424,   425,   426,   427,   430,   431,   435,
     436,   451,   455,   467,   468,   472,   476,   490,   491,   495,
     499,   507,   516,   528,   532,   540,   545,   553,   557,   564,
     566,   570,   586,   587,   591,   598,   599,   603,   608,   616,
     625,   632,   633,   634,   635,   639,   640,   641,   645,   646,
     675,   681,   693,   694,   698,   703,   710,   715,   720,   725,
     733,   740,   745,   753,   760,   769,   778,   787,   799,   800,
     801,   802,   803,   804,   805,   806,   807,   808,   812,   819,
     823,   831,   835,   842,   843,   844,   845,   846,   847,   848,
     849,   850,   851,   855,   860,   864,   871,   875,   880,   885,
     893,   897,   901,   905,   909,   913,   917,   921,   925,   933,
     934,   938,   939,   943,   947,   951,   959,   960,   964,   969,
     977,   981,   985,   989,   993,  1000,  1019,  1020,  1021,  1022,
    1023,  1024,  1025,  1029,  1036,  1037,  1051,  1052,  1071,  1072,
    1073,  1074,  1075,  1079,  1080,  1084,  1092,  1096,  1108,  1115,
    1120,  1128,  1129,  1130,  1137,  1138,  1145,  1146,  1153,  1154,
    1161,  1162,  1169,  1170,  1177,  1178,  1185,  1186,  1193,  1194,
    1198,  1199,  1206,  1207,  1208,  1209,  1213,  1214,  1221,  1222,
    1226,  1227,  1234,  1235,  1239,  1240,  1247,  1248,  1249,  1253,
    1254,  1261,  1262,  1266,  1273,  1274,  1275,  1276,  1280,  1281,
    1285,  1289,  1293,  1297,  1304,  1305,  1309,  1314,  1322,  1323,
    1327,  1331,  1338,  1342,  1346,  1350,  1355,  1360,  1365,  1372,
    1379,  1380,  1381,  1382,  1383,  1384,  1385,  1389,  1396
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
  "\"create\"", "\"cursor\"", "\"default\"", "\"delete\"", "\"double\"",
  "\"drop\"", "\"else\"", "\"enum\"", "\"false\"", "\"float\"", "\"for\"",
  "\"func\"", "\"goto\"", "\"if\"", "\"implements\"", "\"import\"",
  "\"in\"", "\"index\"", "\"insert\"", "\"int\"", "\"int8\"", "\"int16\"",
  "\"int32\"", "\"int64\"", "\"int128\"", "\"interface\"", "\"library\"",
  "\"map\"", "\"new\"", "\"null\"", "\"payable\"", "\"public\"",
  "\"replace\"", "\"return\"", "\"select\"", "\"string\"", "\"struct\"",
  "\"switch\"", "\"table\"", "\"true\"", "\"type\"", "\"uint\"",
  "\"uint8\"", "\"uint16\"", "\"uint32\"", "\"uint64\"", "\"uint128\"",
  "\"update\"", "\"view\"", "'!'", "'|'", "'^'", "'&'", "'<'", "'>'",
  "'+'", "'-'", "'*'", "'/'", "'%'", "'('", "')'", "'.'", "'{'", "'}'",
  "';'", "'='", "','", "'['", "']'", "':'", "'?'", "$accept", "root",
  "component", "import", "contract", "impl_opt", "contract_body",
  "variable", "var_qual", "var_decl", "var_spec", "var_type", "prim_type",
  "declarator_list", "declarator", "size_opt", "var_init", "compound",
  "struct", "field_list", "enumeration", "enum_list", "enumerator",
  "comma_opt", "function", "func_spec", "ctor_spec", "param_list_opt",
  "param_list", "param_decl", "udf_spec", "modifier_opt", "return_opt",
  "return_list", "return_decl", "block", "blk_decl", "interface",
  "interface_body", "library", "library_body", "statement", "empty_stmt",
  "exp_stmt", "assign_stmt", "assign_op", "label_stmt", "if_stmt",
  "loop_stmt", "init_stmt", "cond_exp", "switch_stmt", "switch_blk",
  "case_blk", "jump_stmt", "ddl_stmt", "ddl_prefix", "blk_stmt",
  "expression", "sql_exp", "sql_prefix", "new_exp", "alloc_exp",
  "initializer", "elem_list", "init_elem", "ternary_exp", "or_exp",
  "and_exp", "bit_or_exp", "bit_xor_exp", "bit_and_exp", "eq_exp", "eq_op",
  "cmp_exp", "cmp_op", "add_exp", "add_op", "shift_exp", "shift_op",
  "mul_exp", "mul_op", "cast_exp", "unary_exp", "unary_op", "post_exp",
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
     335,   336,   337,   338,   339,   340,   341,    33,   124,    94,
      38,    60,    62,    43,    45,    42,    47,    37,    40,    41,
      46,   123,   125,    59,    61,    44,    91,    93,    58,    63
};
# endif

#define YYPACT_NINF -278

#define yypact_value_is_default(Yystate) \
  (!!((Yystate) == (-278)))

#define YYTABLE_NINF -31

#define yytable_value_is_error(Yytable_value) \
  0

  /* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
     STATE-NUM.  */
static const yytype_int16 yypact[] =
{
     118,   628,    10,   628,   628,   109,  -278,  -278,  -278,  -278,
    -278,  -278,  -278,  -278,  -278,  -278,  -278,  -278,  -278,  -278,
      -4,  -278,   -35,     6,  -278,  -278,  -278,  -278,   628,    18,
     148,   148,  -278,   968,  -278,    65,   -29,    45,   -33,   -67,
       8,  -278,  -278,  -278,  1797,  -278,   302,  -278,  -278,  -278,
    -278,  -278,  -278,    25,  1759,  -278,   754,  -278,  -278,  -278,
    -278,  -278,  -278,  -278,  1032,  -278,  -278,  -278,   -46,  1282,
    -278,  -278,  -278,  -278,  -278,    46,  -278,  -278,    28,  -278,
    -278,   628,  -278,    48,   386,   153,  -278,  -278,   -59,  -278,
     628,  -278,    76,    91,  1797,  -278,    94,    95,  -278,  -278,
    -278,  -278,  -278,  1491,   102,   121,   127,  -278,  -278,  1797,
     136,  -278,   132,  -278,  -278,  -278,  -278,  -278,  -278,  -278,
     160,   134,  1596,   135,   -30,   133,  -278,   -22,  -278,    16,
     628,    15,  -278,  1160,  -278,  -278,   244,  -278,    23,  -278,
    -278,  -278,  1640,  -278,  1395,  -278,  -278,  -278,  -278,  -278,
     489,  -278,  -278,  -278,  -278,  -278,   196,  -278,  -278,  -278,
    -278,   243,  -278,    86,  -278,   245,  -278,  -278,    -7,   224,
     165,   168,   164,   203,     7,   137,    42,   124,  -278,   250,
    1640,   100,  -278,  -278,    67,   167,   265,  -278,  -278,   628,
     169,  -278,   175,    35,  -278,  -278,  -278,  -278,   628,  1596,
     628,   178,   174,  -278,  1797,  -278,  -278,  -278,   628,    26,
    -278,  -278,  -278,  -278,  -278,  -278,  -278,  -278,   180,   796,
    -278,   177,   181,  1491,  1201,  -278,   182,  -278,   186,  -278,
      81,   184,  1491,   272,  -278,  1491,  -278,   192,   -61,  -278,
    -278,  -278,  -278,   -11,   189,  -278,  1491,  1491,   190,  1596,
    1491,  1596,  1596,  1596,  1596,  -278,  -278,  1596,  -278,  -278,
    -278,  -278,  1596,  -278,  -278,  1596,  -278,  -278,  1596,  -278,
    -278,  -278,  1596,  -278,  -278,  -278,  -278,  -278,  -278,  -278,
    -278,  -278,  -278,  1491,  -278,  -278,  -278,  1693,   628,  1596,
     695,  -278,   191,   193,  -278,   197,  1797,  1797,  -278,  1491,
     127,   195,   137,   127,  -278,  1797,   214,   186,  -278,  -278,
     897,   -32,  -278,  -278,   897,   -26,   216,  1357,  -278,  -278,
     -18,  -278,   212,  -278,  -278,   210,  1596,  1693,  -278,  -278,
      36,  -278,  -278,   592,   211,  1640,  -278,   225,  -278,  -278,
      90,  -278,  -278,   224,   -85,   165,   168,   164,   203,     7,
     137,    42,   124,  -278,   105,  -278,   223,   219,  -278,   221,
    -278,   211,  -278,   628,   231,  1596,   227,   237,  1096,  -278,
    -278,  -278,  1533,  -278,  1245,   238,   912,  1320,   912,    46,
      46,  1201,   232,  1201,    43,   247,   242,  -278,  -278,  -278,
    1491,  -278,  1596,  -278,  -278,  1693,  -278,  -278,  -278,   322,
    -278,  -278,  -278,   246,  1797,  -278,  -278,   248,   253,    46,
      55,  -278,    58,    46,    60,    68,  -278,  -278,  -278,  -278,
    -278,  -278,  -278,  -278,    71,  -278,  -278,  -278,    74,  1797,
    1596,  -278,    46,    46,  -278,    46,    46,    46,  -278,   253,
     249,  -278,  -278,  -278,  -278,  -278,  -278
};

  /* YYDEFACT[STATE-NUM] -- Default reduction number in state STATE-NUM.
     Performed when YYTABLE does not specify something else to do.  Zero
     means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       0,     0,     0,     0,     0,     0,     2,     4,     5,     6,
       3,   267,   260,   261,   262,   263,   264,   265,   266,   268,
      13,    10,     0,     0,     1,     7,     8,     9,     0,     0,
      81,    81,    14,    81,    83,    82,     0,     0,    81,     0,
      81,    32,    33,    34,     0,    48,     0,    40,    35,    36,
      37,    38,    39,     0,    82,    47,     0,    46,    41,    42,
      43,    44,    45,    11,    81,    15,    21,    24,     0,     0,
      29,    16,    57,    58,    17,     0,    72,    73,    30,    84,
     101,     0,   100,     0,     0,     0,   105,   103,     0,    25,
       0,    30,     0,     0,     0,    22,     0,     0,    12,    18,
      19,    20,    26,     0,     0,    28,    49,    51,    71,    75,
       0,   102,     0,   258,   257,   255,   256,   259,   234,   235,
       0,     0,     0,     0,     0,     0,   178,     0,   254,     0,
       0,     0,   179,     0,   252,   180,     0,   181,     0,   253,
     182,   237,     0,   236,     0,    92,   118,    94,    95,   173,
       0,    96,   108,   109,   110,   111,   112,   113,   114,   115,
     116,     0,   117,     0,   174,     0,   176,   183,   194,   196,
     198,   200,   202,   204,   206,   210,   216,   220,   224,   229,
       0,   231,   238,   248,   249,     0,     0,   107,    64,     0,
       0,    60,     0,     0,    55,   229,   249,    23,     0,    53,
       0,     0,    76,    77,    75,   120,   167,   161,     0,     0,
     160,   169,   166,   171,   135,   170,   168,   172,     0,     0,
     140,     0,     0,     0,     0,   186,   184,   185,    30,   162,
       0,     0,     0,     0,   153,     0,   233,     0,     0,    93,
      97,    98,    99,     0,     0,   119,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   208,   209,     0,   214,   215,
     212,   213,     0,   218,   219,     0,   222,   223,     0,   226,
     227,   228,     0,   123,   124,   125,   126,   127,   128,   129,
     130,   131,   132,     0,   232,   242,   243,   244,     0,     0,
       0,   104,     0,    69,    65,    67,     0,     0,    27,     0,
      50,     0,    54,    79,    74,     0,     0,     0,   134,   148,
       0,     0,   149,   150,     0,     0,   183,   249,   164,   139,
       0,   192,    69,   189,   191,   249,     0,   244,   163,   155,
       0,   156,   158,     0,     0,     0,   250,     0,   138,   165,
       0,   175,   177,   197,     0,   199,   201,   203,   205,   207,
     211,   217,   221,   225,     0,   246,     0,   245,   241,     0,
     133,   249,   106,    70,     0,     0,     0,     0,     0,    56,
      52,    78,    85,   151,     0,     0,     0,     0,     0,     0,
       0,    70,     0,     0,     0,     0,     0,   157,   159,   230,
       0,   121,     0,   122,   240,     0,   239,    66,    63,    68,
      31,    61,    59,     0,     0,    90,    80,    87,    88,     0,
       0,   152,     0,     0,     0,     0,   141,   136,   190,   188,
     193,   187,   251,   154,     0,   195,   247,    62,     0,     0,
      53,   144,     0,     0,   142,     0,     0,     0,    86,    89,
       0,   145,   147,   143,   146,   137,    91
};

  /* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -278,  -278,  -278,   345,   349,  -278,  -278,   296,   -25,   -39,
    -205,   -27,   217,  -278,    14,   -66,  -278,   -45,  -278,  -278,
    -278,  -278,     0,    47,   304,  -278,  -278,   166,  -278,    70,
     171,  -278,  -278,   -31,   -57,   -28,  -278,   371,  -278,  -278,
    -278,  -146,   176,  -278,   179,  -278,   146,  -278,  -278,  -278,
      69,  -278,    -2,  -278,  -278,  -278,  -278,  -278,  -135,   -95,
    -278,  -277,  -278,   263,  -278,  -159,  -198,    21,   150,   149,
     145,   157,  -109,  -278,   144,  -278,  -187,  -278,   141,  -278,
     143,  -278,   140,   -77,  -278,  -151,   103,  -278,  -278,  -278,
    -278,    -1
};

  /* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     5,     6,     7,     8,    29,    64,    65,    66,    67,
      68,    90,    70,   105,   106,   301,   193,    71,    72,   368,
      73,   293,   294,   364,    74,    75,    76,   201,   202,   203,
      77,    37,   406,   407,   408,   149,   150,     9,    38,    10,
      40,   151,   152,   153,   154,   283,   155,   156,   157,   314,
     374,   158,   234,   333,   159,   160,   161,   162,   163,   164,
     165,   166,   226,   321,   322,   323,   167,   168,   169,   170,
     171,   172,   173,   257,   174,   262,   175,   265,   176,   268,
     177,   272,   178,   195,   180,   181,   356,   357,   182,   183,
      19,   196
};

  /* YYTABLE[YYPACT[STATE-NUM]] -- What to do in state STATE-NUM.  If
     positive, shift that token.  If negative, reduce the rule whose
     number is the opposite.  If YYTABLE_NINF, syntax error.  */
static const yytype_int16 yytable[] =
{
      20,   230,    22,    23,   242,    89,    69,   179,   194,   238,
     355,    86,   302,   209,   311,   249,   222,   218,    21,   100,
     247,   316,   376,   392,   231,   211,   324,    32,   378,    95,
     258,   259,    78,   215,    84,    34,    35,    69,   336,   148,
     337,    85,    84,    91,   247,    93,   212,   108,    28,   186,
     355,   255,   256,    91,   216,    97,   213,   102,   103,   147,
     187,   266,   267,    78,   217,   236,    30,   190,   107,    82,
     -30,   102,   103,   179,    80,   350,    34,    35,   246,   247,
     110,   380,   200,   184,   315,   285,   286,   247,   320,   107,
      84,   359,   367,    91,    81,   285,   286,   330,   260,   261,
     238,   220,   250,   284,   -30,   241,   225,    31,    91,    24,
      87,   340,   375,   223,   219,   344,   375,    84,   426,    33,
     -30,   232,   -30,    94,   233,   240,   109,   285,   286,   221,
     -30,   -30,   228,    79,   308,   386,   263,   264,   298,   384,
     299,   247,   179,   -30,   360,   348,     1,    84,   354,   184,
     421,   111,   341,   -30,   432,     1,   287,   433,   288,   435,
     247,   185,     2,   403,   289,   247,   287,   436,   288,   192,
     437,     2,     3,   438,   289,   290,   247,   200,   188,   429,
     310,     3,     4,   324,   328,   324,   247,   388,   295,   245,
     246,   247,   189,   391,   425,   247,   191,   107,   287,   107,
     288,    36,    39,    91,   369,   197,   289,   307,   393,    83,
     247,    88,   300,   179,   303,   338,    34,    35,   317,   269,
     270,   271,   418,   325,   420,   412,   198,   415,   255,   256,
     263,   264,   334,   199,   204,   205,   206,   207,   210,   410,
     243,   214,   414,   302,   244,   251,   248,    11,   113,   114,
     115,   116,   117,   252,   254,   424,   179,   253,   389,   273,
     274,   275,   276,   277,   278,   279,   280,   281,   282,   366,
     291,   118,   119,   292,   296,    11,   297,   304,   200,   305,
     318,    12,   309,   319,   327,   126,   329,   358,   326,   361,
     128,   335,   339,   342,   362,    91,    91,    13,   363,    14,
     132,   365,   370,    92,    91,    11,   122,    15,    16,    12,
     133,   134,   125,   372,   135,   379,   137,   381,   383,   290,
      17,   139,   394,   390,   395,    13,   400,    14,   396,   140,
      18,   141,   361,   398,   419,    15,    16,   142,   143,    12,
     401,   411,   144,   233,   249,   405,   422,   229,    17,   427,
      25,   416,   417,   429,    26,    13,   446,    14,    18,   430,
      99,   237,   295,   397,   440,    15,    16,    91,   101,   382,
     306,    91,   439,   428,   331,   371,    27,   405,    17,   332,
     325,   431,   325,   377,   423,   434,   399,   112,    18,    11,
     113,   114,   115,   116,   117,   312,   227,   346,   313,   343,
     345,   349,   405,    91,   441,   442,   351,   443,   444,   445,
     347,   352,   353,   118,   119,    41,   120,    42,   121,    43,
     122,    44,   123,    12,   124,    45,   125,   126,    91,   127,
     385,    46,   128,     0,   129,     0,   130,   131,     0,    13,
       0,    14,   132,    47,    48,    49,    50,    51,    52,    15,
      16,    53,   133,   134,     0,     0,   135,   136,   137,    55,
       0,   138,    17,   139,    56,    57,    58,    59,    60,    61,
      62,   140,    18,   141,     0,     0,     0,     0,     0,   142,
     143,     0,     0,     0,   144,     0,     0,    84,   145,   146,
     112,     0,    11,   113,   114,   115,   116,   117,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,   118,   119,    41,   120,
      42,   121,    43,   122,    44,   123,    12,   124,    45,   125,
     126,     0,   127,     0,    46,   128,     0,   129,     0,   130,
     131,     0,    13,     0,    14,   132,    47,    48,    49,    50,
      51,    52,    15,    16,    53,   133,   134,     0,     0,   135,
     136,   137,    55,     0,   138,    17,   139,    56,    57,    58,
      59,    60,    61,    62,   140,    18,   141,     0,     0,     0,
       0,     0,   142,   143,     0,     0,     0,   144,     0,     0,
      84,   239,   146,   112,     0,    11,   113,   114,   115,   116,
     117,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,   118,
     119,     0,   120,     0,   121,     0,   122,     0,   123,    12,
     124,    11,   125,   126,     0,   127,     0,     0,   128,     0,
     129,     0,   130,   131,     0,    13,     0,    14,   132,     0,
       0,     0,     0,     0,     0,    15,    16,     0,   133,   134,
       0,     0,   135,   136,   137,    12,     0,   138,    17,   139,
       0,     0,     0,     0,     0,     0,     0,   140,    18,   141,
       0,    13,     0,    14,     0,   142,   143,     0,     0,     0,
     144,    15,    16,    84,   387,   146,   112,     0,    11,   113,
     114,   115,   116,   117,    17,     0,     0,     0,     0,     0,
       0,     0,     0,     0,    18,     0,     0,     0,     0,     0,
       0,     0,   118,   119,     0,   120,     0,   121,     0,   122,
       0,   123,    12,   124,     0,   125,   126,     0,   127,     0,
       0,   128,     0,   129,     0,   130,   131,     0,    13,     0,
      14,   132,     0,     0,     0,    96,     0,    11,    15,    16,
       0,   133,   134,     0,     0,   135,   136,   137,     0,     0,
     138,    17,   139,     0,     0,     0,     0,     0,     0,     0,
     140,    18,   141,     0,     0,     0,     0,     0,   142,   143,
       0,    12,     0,   144,     0,     0,    84,     0,   146,    11,
     113,   114,   115,   116,   117,     0,     0,    13,     0,    14,
       0,     0,     0,     0,     0,     0,     0,    15,    16,     0,
       0,     0,     0,   118,   119,    41,     0,    42,     0,    43,
      17,     0,     0,    12,     0,    45,     0,   126,     0,     0,
      18,     0,   128,     0,     0,     0,     0,     0,     0,    13,
       0,    14,   132,    47,    48,    49,    50,    51,    52,    15,
      16,    53,   133,   134,     0,     0,   135,     0,   137,    55,
       0,     0,    17,   139,     0,    57,    58,    59,    60,    61,
      62,   140,    18,   141,     0,     0,     0,     0,     0,   142,
     143,     0,     0,     0,   144,     0,     0,     0,     0,   146,
      11,   113,   114,   115,   116,   117,     0,     0,     0,     0,
       0,     0,     0,     0,     0,    11,   113,   114,   115,   116,
     117,     0,     0,     0,   118,   119,     0,     0,     0,     0,
       0,     0,     0,     0,    12,     0,     0,     0,     0,     0,
       0,     0,     0,   128,     0,     0,     0,     0,     0,    12,
      13,     0,    14,     0,     0,     0,     0,     0,   128,     0,
      15,    16,     0,   208,   134,    13,     0,    14,     0,     0,
       0,    11,     0,    17,   139,    15,    16,     0,   208,   134,
       0,     0,     0,    18,   141,     0,     0,     0,    17,   139,
     142,   143,     0,     0,     0,   144,     0,    41,    18,    42,
     373,    43,     0,    44,     0,    12,     0,    45,     0,     0,
     235,     0,     0,    46,     0,     0,     0,     0,     0,     0,
       0,    13,     0,    14,     0,    47,    48,    49,    50,    51,
      52,    15,    16,    53,     0,    11,    34,    54,     0,     0,
       0,    55,     0,     0,    17,     0,    56,    57,    58,    59,
      60,    61,    62,     0,    18,     0,     0,     0,     0,     0,
       0,    41,     0,    42,     0,    43,     0,    44,     0,    12,
      63,    45,     0,     0,     0,     0,     0,    46,     0,     0,
       0,     0,     0,     0,     0,    13,     0,    14,     0,    47,
      48,    49,    50,    51,    52,    15,    16,    53,     0,    11,
      34,    54,     0,     0,     0,    55,     0,     0,    17,     0,
      56,    57,    58,    59,    60,    61,    62,     0,    18,     0,
       0,     0,     0,     0,     0,    41,     0,    42,     0,    43,
       0,     0,     0,    12,    98,    45,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,    13,
       0,    14,     0,    47,    48,    49,    50,    51,    52,    15,
      16,    53,     0,    11,     0,     0,     0,     0,     0,    55,
       0,     0,    17,     0,     0,    57,    58,    59,    60,    61,
      62,     0,    18,     0,     0,     0,     0,     0,     0,    41,
       0,    42,     0,    43,     0,     0,     0,    12,   402,    45,
       0,     0,     0,     0,    11,   113,   114,   115,   116,   117,
       0,     0,     0,    13,     0,    14,     0,    47,    48,    49,
      50,    51,    52,    15,    16,    53,     0,     0,   118,   119,
       0,     0,     0,    55,     0,     0,    17,     0,    12,    57,
      58,    59,    60,    61,    62,     0,    18,   128,    11,   113,
     114,   115,   116,   117,    13,     0,    14,     0,     0,     0,
       0,   224,     0,     0,    15,    16,     0,   208,   134,     0,
       0,     0,   118,   119,     0,     0,     0,    17,   139,     0,
       0,     0,    12,   104,     0,    11,   126,    18,   141,     0,
       0,   128,     0,     0,   142,   143,     0,     0,    13,   144,
      14,   132,   224,     0,     0,     0,     0,     0,    15,    16,
       0,   133,   134,     0,     0,   135,     0,   137,     0,    12,
       0,    17,   139,    11,   113,   114,   115,   116,   117,     0,
     140,    18,   141,     0,     0,    13,     0,    14,   142,   143,
       0,     0,     0,   144,   409,    15,    16,   118,   119,     0,
       0,     0,     0,     0,     0,     0,     0,    12,    17,     0,
     -30,   126,     0,     0,     0,     0,   128,     0,    18,     0,
       0,     0,     0,    13,     0,    14,   132,     0,     0,     0,
       0,     0,     0,    15,    16,     0,   133,   134,     0,     0,
     135,     0,   137,     0,   -30,     0,    17,   139,    11,   113,
     114,   115,   116,   117,     0,   140,    18,   141,     0,     0,
     -30,     0,   -30,   142,   143,     0,     0,     0,   144,   413,
     -30,   -30,   118,   119,    41,     0,    42,     0,    43,     0,
       0,     0,    12,   -30,    45,     0,   126,     0,     0,     0,
       0,   128,     0,   -30,     0,     0,     0,     0,    13,     0,
      14,   132,    47,    48,    49,    50,    51,    52,    15,    16,
       0,   133,   134,     0,     0,   135,     0,   137,    55,     0,
       0,    17,   139,     0,    57,    58,    59,    60,    61,    62,
     140,    18,   141,     0,     0,     0,     0,     0,   142,   143,
       0,     0,     0,   144,    11,   113,   114,   115,   116,   117,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,   118,   119,
       0,     0,     0,     0,     0,     0,     0,     0,    12,     0,
       0,     0,   126,     0,     0,     0,    11,   128,     0,     0,
       0,     0,     0,     0,    13,     0,    14,   132,     0,     0,
       0,     0,     0,     0,    15,    16,     0,   133,   134,     0,
       0,   135,    41,   137,    42,     0,    43,    17,   139,     0,
      12,     0,    45,     0,     0,     0,   140,    18,   141,     0,
       0,     0,     0,     0,   142,   143,    13,     0,    14,   144,
      47,    48,    49,    50,    51,    52,    15,    16,    53,    11,
     113,   114,   115,   116,   117,     0,    55,     0,     0,    17,
       0,     0,    57,    58,    59,    60,    61,    62,     0,    18,
       0,     0,     0,   118,   119,     0,     0,     0,     0,     0,
       0,   404,     0,    12,     0,     0,     0,     0,     0,     0,
       0,     0,   128,    11,   113,   114,   115,   116,   117,    13,
       0,    14,     0,     0,     0,     0,     0,     0,     0,    15,
      16,     0,   208,   134,     0,     0,     0,   118,   119,     0,
       0,     0,    17,   139,     0,     0,     0,    12,     0,     0,
       0,     0,    18,   141,     0,     0,   128,     0,     0,   142,
     143,     0,     0,    13,   144,    14,    11,   113,   114,   115,
     116,   117,     0,    15,    16,     0,   208,   134,     0,     0,
       0,     0,     0,     0,     0,     0,    17,   139,     0,     0,
     118,   119,     0,     0,     0,     0,    18,   141,     0,     0,
      12,     0,     0,   142,   143,     0,     0,     0,   235,   128,
       0,     0,     0,     0,     0,     0,    13,     0,    14,     0,
       0,     0,     0,     0,     0,     0,    15,    16,     0,   133,
     134,     0,    11,     0,     0,     0,     0,     0,     0,    17,
     139,     0,     0,     0,     0,     0,     0,     0,     0,    18,
     141,     0,     0,     0,     0,     0,   142,   143,    41,     0,
      42,   144,    43,     0,    44,     0,    12,     0,    45,     0,
      11,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,    13,     0,    14,     0,    47,    48,    49,    50,
      51,    52,    15,    16,    53,     0,    41,    79,    42,     0,
      43,     0,    55,     0,    12,    17,    45,     0,    57,    58,
      59,    60,    61,    62,     0,    18,     0,     0,     0,     0,
      13,     0,    14,     0,    47,    48,    49,    50,    51,    52,
      15,    16,    53,     0,     0,     0,     0,     0,     0,     0,
      55,     0,     0,    17,     0,     0,    57,    58,    59,    60,
      61,    62,     0,    18
};

static const yytype_int16 yycheck[] =
{
       1,   136,     3,     4,   150,    44,    33,    84,   103,   144,
     287,    39,   199,   122,   219,    22,     1,     1,     8,    64,
     105,   219,    54,   108,     1,    55,   224,    28,    54,    54,
      23,    24,    33,    55,   101,    68,    69,    64,    99,    84,
      51,   108,   101,    44,   105,    46,    76,    75,    52,   108,
     327,    25,    26,    54,    76,    56,    86,   103,   104,    84,
      88,    19,    20,    64,    86,   142,   101,    94,    69,   102,
       3,   103,   104,   150,   103,   262,    68,    69,   104,   105,
      81,    99,   109,    84,   219,    27,    28,   105,   223,    90,
     101,   289,   297,    94,    49,    27,    28,   232,    91,    92,
     235,   129,   109,   180,    37,   150,   133,   101,   109,     0,
     102,   246,   310,    98,    98,   250,   314,   101,   395,   101,
      53,    98,    55,    98,   101,   150,    98,    27,    28,   130,
      63,    64,   133,    68,   108,    99,    93,    94,   103,   326,
     105,   105,   219,    76,   290,   254,    37,   101,   283,   150,
     107,   103,   247,    86,    99,    37,    98,    99,   100,    99,
     105,     8,    53,   368,   106,   105,    98,    99,   100,    74,
      99,    53,    63,    99,   106,   108,   105,   204,   102,   105,
     219,    63,    64,   381,   103,   383,   105,   333,   189,   103,
     104,   105,   101,   103,   392,   105,   102,   198,    98,   200,
     100,    30,    31,   204,   299,   103,   106,   208,   103,    38,
     105,    40,   198,   290,   200,   243,    68,    69,   219,    95,
      96,    97,   381,   224,   383,   376,   105,   378,    25,    26,
      93,    94,   233,   106,    98,   103,    76,   103,   103,   374,
      44,   108,   377,   430,     1,    21,     1,     3,     4,     5,
       6,     7,     8,    88,    90,   390,   333,    89,   335,     9,
      10,    11,    12,    13,    14,    15,    16,    17,    18,   296,
     103,    27,    28,     8,   105,     3,   101,    99,   305,   105,
     103,    37,   102,   102,    98,    41,   102,   288,   106,   290,
      46,    99,   103,   103,   103,   296,   297,    53,   105,    55,
      56,   104,   107,     1,   305,     3,    34,    63,    64,    37,
      66,    67,    40,    99,    70,    99,    72,   105,   108,   108,
      76,    77,    99,    98,   105,    53,    99,    55,   107,    85,
      86,    87,   333,   102,   102,    63,    64,    93,    94,    37,
     103,   103,    98,   101,    22,   372,    99,   103,    76,   103,
       5,   379,   380,   105,     5,    53,   107,    55,    86,   106,
      64,   144,   363,   363,   430,    63,    64,   368,    64,   322,
     204,   372,   429,   404,   102,   305,     5,   404,    76,   233,
     381,   409,   383,   314,   386,   413,   365,     1,    86,     3,
       4,     5,     6,     7,     8,   219,   133,   252,   219,   249,
     251,   257,   429,   404,   432,   433,   265,   435,   436,   437,
     253,   268,   272,    27,    28,    29,    30,    31,    32,    33,
      34,    35,    36,    37,    38,    39,    40,    41,   429,    43,
     327,    45,    46,    -1,    48,    -1,    50,    51,    -1,    53,
      -1,    55,    56,    57,    58,    59,    60,    61,    62,    63,
      64,    65,    66,    67,    -1,    -1,    70,    71,    72,    73,
      -1,    75,    76,    77,    78,    79,    80,    81,    82,    83,
      84,    85,    86,    87,    -1,    -1,    -1,    -1,    -1,    93,
      94,    -1,    -1,    -1,    98,    -1,    -1,   101,   102,   103,
       1,    -1,     3,     4,     5,     6,     7,     8,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    27,    28,    29,    30,
      31,    32,    33,    34,    35,    36,    37,    38,    39,    40,
      41,    -1,    43,    -1,    45,    46,    -1,    48,    -1,    50,
      51,    -1,    53,    -1,    55,    56,    57,    58,    59,    60,
      61,    62,    63,    64,    65,    66,    67,    -1,    -1,    70,
      71,    72,    73,    -1,    75,    76,    77,    78,    79,    80,
      81,    82,    83,    84,    85,    86,    87,    -1,    -1,    -1,
      -1,    -1,    93,    94,    -1,    -1,    -1,    98,    -1,    -1,
     101,   102,   103,     1,    -1,     3,     4,     5,     6,     7,
       8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,
      28,    -1,    30,    -1,    32,    -1,    34,    -1,    36,    37,
      38,     3,    40,    41,    -1,    43,    -1,    -1,    46,    -1,
      48,    -1,    50,    51,    -1,    53,    -1,    55,    56,    -1,
      -1,    -1,    -1,    -1,    -1,    63,    64,    -1,    66,    67,
      -1,    -1,    70,    71,    72,    37,    -1,    75,    76,    77,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    85,    86,    87,
      -1,    53,    -1,    55,    -1,    93,    94,    -1,    -1,    -1,
      98,    63,    64,   101,   102,   103,     1,    -1,     3,     4,
       5,     6,     7,     8,    76,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    86,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    27,    28,    -1,    30,    -1,    32,    -1,    34,
      -1,    36,    37,    38,    -1,    40,    41,    -1,    43,    -1,
      -1,    46,    -1,    48,    -1,    50,    51,    -1,    53,    -1,
      55,    56,    -1,    -1,    -1,     1,    -1,     3,    63,    64,
      -1,    66,    67,    -1,    -1,    70,    71,    72,    -1,    -1,
      75,    76,    77,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      85,    86,    87,    -1,    -1,    -1,    -1,    -1,    93,    94,
      -1,    37,    -1,    98,    -1,    -1,   101,    -1,   103,     3,
       4,     5,     6,     7,     8,    -1,    -1,    53,    -1,    55,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    63,    64,    -1,
      -1,    -1,    -1,    27,    28,    29,    -1,    31,    -1,    33,
      76,    -1,    -1,    37,    -1,    39,    -1,    41,    -1,    -1,
      86,    -1,    46,    -1,    -1,    -1,    -1,    -1,    -1,    53,
      -1,    55,    56,    57,    58,    59,    60,    61,    62,    63,
      64,    65,    66,    67,    -1,    -1,    70,    -1,    72,    73,
      -1,    -1,    76,    77,    -1,    79,    80,    81,    82,    83,
      84,    85,    86,    87,    -1,    -1,    -1,    -1,    -1,    93,
      94,    -1,    -1,    -1,    98,    -1,    -1,    -1,    -1,   103,
       3,     4,     5,     6,     7,     8,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,     3,     4,     5,     6,     7,
       8,    -1,    -1,    -1,    27,    28,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    37,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    46,    -1,    -1,    -1,    -1,    -1,    37,
      53,    -1,    55,    -1,    -1,    -1,    -1,    -1,    46,    -1,
      63,    64,    -1,    66,    67,    53,    -1,    55,    -1,    -1,
      -1,     3,    -1,    76,    77,    63,    64,    -1,    66,    67,
      -1,    -1,    -1,    86,    87,    -1,    -1,    -1,    76,    77,
      93,    94,    -1,    -1,    -1,    98,    -1,    29,    86,    31,
     103,    33,    -1,    35,    -1,    37,    -1,    39,    -1,    -1,
      98,    -1,    -1,    45,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    53,    -1,    55,    -1,    57,    58,    59,    60,    61,
      62,    63,    64,    65,    -1,     3,    68,    69,    -1,    -1,
      -1,    73,    -1,    -1,    76,    -1,    78,    79,    80,    81,
      82,    83,    84,    -1,    86,    -1,    -1,    -1,    -1,    -1,
      -1,    29,    -1,    31,    -1,    33,    -1,    35,    -1,    37,
     102,    39,    -1,    -1,    -1,    -1,    -1,    45,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    53,    -1,    55,    -1,    57,
      58,    59,    60,    61,    62,    63,    64,    65,    -1,     3,
      68,    69,    -1,    -1,    -1,    73,    -1,    -1,    76,    -1,
      78,    79,    80,    81,    82,    83,    84,    -1,    86,    -1,
      -1,    -1,    -1,    -1,    -1,    29,    -1,    31,    -1,    33,
      -1,    -1,    -1,    37,   102,    39,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    53,
      -1,    55,    -1,    57,    58,    59,    60,    61,    62,    63,
      64,    65,    -1,     3,    -1,    -1,    -1,    -1,    -1,    73,
      -1,    -1,    76,    -1,    -1,    79,    80,    81,    82,    83,
      84,    -1,    86,    -1,    -1,    -1,    -1,    -1,    -1,    29,
      -1,    31,    -1,    33,    -1,    -1,    -1,    37,   102,    39,
      -1,    -1,    -1,    -1,     3,     4,     5,     6,     7,     8,
      -1,    -1,    -1,    53,    -1,    55,    -1,    57,    58,    59,
      60,    61,    62,    63,    64,    65,    -1,    -1,    27,    28,
      -1,    -1,    -1,    73,    -1,    -1,    76,    -1,    37,    79,
      80,    81,    82,    83,    84,    -1,    86,    46,     3,     4,
       5,     6,     7,     8,    53,    -1,    55,    -1,    -1,    -1,
      -1,   101,    -1,    -1,    63,    64,    -1,    66,    67,    -1,
      -1,    -1,    27,    28,    -1,    -1,    -1,    76,    77,    -1,
      -1,    -1,    37,     1,    -1,     3,    41,    86,    87,    -1,
      -1,    46,    -1,    -1,    93,    94,    -1,    -1,    53,    98,
      55,    56,   101,    -1,    -1,    -1,    -1,    -1,    63,    64,
      -1,    66,    67,    -1,    -1,    70,    -1,    72,    -1,    37,
      -1,    76,    77,     3,     4,     5,     6,     7,     8,    -1,
      85,    86,    87,    -1,    -1,    53,    -1,    55,    93,    94,
      -1,    -1,    -1,    98,    99,    63,    64,    27,    28,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    37,    76,    -1,
       3,    41,    -1,    -1,    -1,    -1,    46,    -1,    86,    -1,
      -1,    -1,    -1,    53,    -1,    55,    56,    -1,    -1,    -1,
      -1,    -1,    -1,    63,    64,    -1,    66,    67,    -1,    -1,
      70,    -1,    72,    -1,    37,    -1,    76,    77,     3,     4,
       5,     6,     7,     8,    -1,    85,    86,    87,    -1,    -1,
      53,    -1,    55,    93,    94,    -1,    -1,    -1,    98,    99,
      63,    64,    27,    28,    29,    -1,    31,    -1,    33,    -1,
      -1,    -1,    37,    76,    39,    -1,    41,    -1,    -1,    -1,
      -1,    46,    -1,    86,    -1,    -1,    -1,    -1,    53,    -1,
      55,    56,    57,    58,    59,    60,    61,    62,    63,    64,
      -1,    66,    67,    -1,    -1,    70,    -1,    72,    73,    -1,
      -1,    76,    77,    -1,    79,    80,    81,    82,    83,    84,
      85,    86,    87,    -1,    -1,    -1,    -1,    -1,    93,    94,
      -1,    -1,    -1,    98,     3,     4,     5,     6,     7,     8,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,    28,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    37,    -1,
      -1,    -1,    41,    -1,    -1,    -1,     3,    46,    -1,    -1,
      -1,    -1,    -1,    -1,    53,    -1,    55,    56,    -1,    -1,
      -1,    -1,    -1,    -1,    63,    64,    -1,    66,    67,    -1,
      -1,    70,    29,    72,    31,    -1,    33,    76,    77,    -1,
      37,    -1,    39,    -1,    -1,    -1,    85,    86,    87,    -1,
      -1,    -1,    -1,    -1,    93,    94,    53,    -1,    55,    98,
      57,    58,    59,    60,    61,    62,    63,    64,    65,     3,
       4,     5,     6,     7,     8,    -1,    73,    -1,    -1,    76,
      -1,    -1,    79,    80,    81,    82,    83,    84,    -1,    86,
      -1,    -1,    -1,    27,    28,    -1,    -1,    -1,    -1,    -1,
      -1,    98,    -1,    37,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    46,     3,     4,     5,     6,     7,     8,    53,
      -1,    55,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    63,
      64,    -1,    66,    67,    -1,    -1,    -1,    27,    28,    -1,
      -1,    -1,    76,    77,    -1,    -1,    -1,    37,    -1,    -1,
      -1,    -1,    86,    87,    -1,    -1,    46,    -1,    -1,    93,
      94,    -1,    -1,    53,    98,    55,     3,     4,     5,     6,
       7,     8,    -1,    63,    64,    -1,    66,    67,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    76,    77,    -1,    -1,
      27,    28,    -1,    -1,    -1,    -1,    86,    87,    -1,    -1,
      37,    -1,    -1,    93,    94,    -1,    -1,    -1,    98,    46,
      -1,    -1,    -1,    -1,    -1,    -1,    53,    -1,    55,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    63,    64,    -1,    66,
      67,    -1,     3,    -1,    -1,    -1,    -1,    -1,    -1,    76,
      77,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    86,
      87,    -1,    -1,    -1,    -1,    -1,    93,    94,    29,    -1,
      31,    98,    33,    -1,    35,    -1,    37,    -1,    39,    -1,
       3,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    53,    -1,    55,    -1,    57,    58,    59,    60,
      61,    62,    63,    64,    65,    -1,    29,    68,    31,    -1,
      33,    -1,    73,    -1,    37,    76,    39,    -1,    79,    80,
      81,    82,    83,    84,    -1,    86,    -1,    -1,    -1,    -1,
      53,    -1,    55,    -1,    57,    58,    59,    60,    61,    62,
      63,    64,    65,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      73,    -1,    -1,    76,    -1,    -1,    79,    80,    81,    82,
      83,    84,    -1,    86
};

  /* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
     symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    37,    53,    63,    64,   111,   112,   113,   114,   147,
     149,     3,    37,    53,    55,    63,    64,    76,    86,   200,
     201,     8,   201,   201,     0,   113,   114,   147,    52,   115,
     101,   101,   201,   101,    68,    69,   140,   141,   148,   140,
     150,    29,    31,    33,    35,    39,    45,    57,    58,    59,
      60,    61,    62,    65,    69,    73,    78,    79,    80,    81,
      82,    83,    84,   102,   116,   117,   118,   119,   120,   121,
     122,   127,   128,   130,   134,   135,   136,   140,   201,    68,
     103,    49,   102,   140,   101,   108,   145,   102,   140,   119,
     121,   201,     1,   201,    98,   118,     1,   201,   102,   117,
     127,   134,   103,   104,     1,   123,   124,   201,   145,    98,
     201,   103,     1,     4,     5,     6,     7,     8,    27,    28,
      30,    32,    34,    36,    38,    40,    41,    43,    46,    48,
      50,    51,    56,    66,    67,    70,    71,    72,    75,    77,
      85,    87,    93,    94,    98,   102,   103,   118,   127,   145,
     146,   151,   152,   153,   154,   156,   157,   158,   161,   164,
     165,   166,   167,   168,   169,   170,   171,   176,   177,   178,
     179,   180,   181,   182,   184,   186,   188,   190,   192,   193,
     194,   195,   198,   199,   201,     8,   108,   145,   102,   101,
     121,   102,    74,   126,   169,   193,   201,   103,   105,   106,
     121,   137,   138,   139,    98,   103,    76,   103,    66,   182,
     103,    55,    76,    86,   108,    55,    76,    86,     1,    98,
     145,   201,     1,    98,   101,   121,   172,   173,   201,   103,
     168,     1,    98,   101,   162,    98,   193,   122,   168,   102,
     118,   127,   151,    44,     1,   103,   104,   105,     1,    22,
     109,    21,    88,    89,    90,    25,    26,   183,    23,    24,
      91,    92,   185,    93,    94,   187,    19,    20,   189,    95,
      96,    97,   191,     9,    10,    11,    12,    13,    14,    15,
      16,    17,    18,   155,   193,    27,    28,    98,   100,   106,
     108,   103,     8,   131,   132,   201,   105,   101,   103,   105,
     124,   125,   186,   124,    99,   105,   137,   201,   108,   102,
     119,   120,   152,   154,   159,   168,   176,   201,   103,   102,
     168,   173,   174,   175,   176,   201,   106,    98,   103,   102,
     168,   102,   156,   163,   201,    99,    99,    51,   145,   103,
     168,   169,   103,   178,   168,   179,   180,   181,   182,   184,
     186,   188,   190,   192,   168,   171,   196,   197,   201,   176,
     151,   201,   103,   105,   133,   104,   121,   120,   129,   169,
     107,   139,    99,   103,   160,   176,    54,   160,    54,    99,
      99,   105,   133,   108,   186,   196,    99,   102,   151,   193,
      98,   103,   108,   103,    99,   105,   107,   132,   102,   177,
      99,   103,   102,   120,    98,   121,   142,   143,   144,    99,
     168,   103,   195,    99,   168,   195,   145,   145,   175,   102,
     175,   107,    99,   162,   168,   176,   171,   103,   143,   105,
     106,   145,    99,    99,   145,    99,    99,    99,    99,   144,
     125,   145,   145,   145,   145,   145,   107
};

  /* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint8 yyr1[] =
{
       0,   110,   111,   111,   112,   112,   112,   112,   112,   112,
     113,   114,   114,   115,   115,   116,   116,   116,   116,   116,
     116,   117,   117,   117,   118,   118,   119,   119,   120,   121,
     121,   121,   122,   122,   122,   122,   122,   122,   122,   122,
     122,   122,   122,   122,   122,   122,   122,   122,   122,   123,
     123,   124,   124,   125,   125,   126,   126,   127,   127,   128,
     128,   129,   129,   130,   130,   131,   131,   132,   132,   133,
     133,   134,   135,   135,   136,   137,   137,   138,   138,   139,
     140,   141,   141,   141,   141,   142,   142,   142,   143,   143,
     144,   144,   145,   145,   146,   146,   146,   146,   146,   146,
     147,   148,   148,   149,   150,   150,   150,   150,   151,   151,
     151,   151,   151,   151,   151,   151,   151,   151,   152,   153,
     153,   154,   154,   155,   155,   155,   155,   155,   155,   155,
     155,   155,   155,   156,   156,   156,   157,   157,   157,   157,
     158,   158,   158,   158,   158,   158,   158,   158,   158,   159,
     159,   160,   160,   161,   161,   161,   162,   162,   163,   163,
     164,   164,   164,   164,   164,   165,   166,   166,   166,   166,
     166,   166,   166,   167,   168,   168,   169,   169,   170,   170,
     170,   170,   170,   171,   171,   171,   172,   172,   173,   174,
     174,   175,   175,   175,   176,   176,   177,   177,   178,   178,
     179,   179,   180,   180,   181,   181,   182,   182,   183,   183,
     184,   184,   185,   185,   185,   185,   186,   186,   187,   187,
     188,   188,   189,   189,   190,   190,   191,   191,   191,   192,
     192,   193,   193,   193,   194,   194,   194,   194,   195,   195,
     195,   195,   195,   195,   196,   196,   197,   197,   198,   198,
     198,   198,   199,   199,   199,   199,   199,   199,   199,   199,
     200,   200,   200,   200,   200,   200,   200,   201,   201
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
       1,     1,     1,     1,     1,     1,     1,     1,     1
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

#line 1917 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1430  */
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
#line 277 "grammar.y" /* yacc.c:1648  */
    {
        ast_blk_t *blk = blk_new_contract(&(yylsp[-1]));

        /* add default constructor */
        id_add(&blk->ids, id_new_ctor((yyvsp[-3].str), NULL, NULL, &(yylsp[-3])));

        id_add(&ROOT->ids, id_new_contract((yyvsp[-3].str), (yyvsp[-2].exp), blk, &(yyloc)));
    }
#line 2114 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 12:
#line 286 "grammar.y" /* yacc.c:1648  */
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

        id_add(&ROOT->ids, id_new_contract((yyvsp[-4].str), (yyvsp[-3].exp), (yyvsp[-1].blk), &(yyloc)));
    }
#line 2134 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 13:
#line 304 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 2140 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 14:
#line 306 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yylsp[0]));
    }
#line 2148 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 15:
#line 313 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2157 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 16:
#line 318 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2166 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 17:
#line 323 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2175 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 18:
#line 328 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2184 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 19:
#line 333 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2193 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 20:
#line 338 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2202 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 22:
#line 347 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_PUBLIC;
    }
#line 2211 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 23:
#line 352 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2220 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 25:
#line 361 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_CONST;
    }
#line 2229 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 27:
#line 370 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.dflt_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.dflt_exp = (yyvsp[-1].exp);
    }
#line 2242 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 28:
#line 382 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.type_exp = (yyvsp[-1].exp);
    }
#line 2255 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 29:
#line 394 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type((yyvsp[0].type), &(yylsp[0]));
    }
#line 2263 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 30:
#line 398 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_NONE, &(yylsp[0]));

        (yyval.exp)->u_type.name = (yyvsp[0].str);
    }
#line 2273 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 31:
#line 404 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_MAP, &(yylsp[-5]));

        (yyval.exp)->u_type.k_exp = (yyvsp[-3].exp);
        (yyval.exp)->u_type.v_exp = (yyvsp[-1].exp);
    }
#line 2284 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 32:
#line 413 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_ACCOUNT; }
#line 2290 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 33:
#line 414 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BOOL; }
#line 2296 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 34:
#line 415 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT8; }
#line 2302 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 35:
#line 416 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT8; }
#line 2308 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 36:
#line 417 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT16; }
#line 2314 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 37:
#line 418 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2320 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 38:
#line 419 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT64; }
#line 2326 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 39:
#line 420 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT128; }
#line 2332 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 40:
#line 421 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2338 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 41:
#line 422 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT8; }
#line 2344 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 42:
#line 423 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT16; }
#line 2350 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 43:
#line 424 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2356 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 44:
#line 425 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT64; }
#line 2362 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 45:
#line 426 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT128; }
#line 2368 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 46:
#line 427 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2374 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 47:
#line 430 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_STRING; }
#line 2380 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 48:
#line 431 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_CURSOR; }
#line 2386 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 50:
#line 437 "grammar.y" /* yacc.c:1648  */
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
#line 2402 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 51:
#line 452 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PRIVATE, &(yylsp[0]));
    }
#line 2410 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 52:
#line 456 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2423 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 53:
#line 467 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = exp_new_null(&(yyloc)); }
#line 2429 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 55:
#line 473 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 2437 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 56:
#line 477 "grammar.y" /* yacc.c:1648  */
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
#line 2452 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 59:
#line 496 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_struct((yyvsp[-4].str), (yyvsp[-1].vect), &(yyloc));
    }
#line 2460 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 60:
#line 500 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2469 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 61:
#line 508 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2482 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 62:
#line 517 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2495 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 63:
#line 529 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_enum((yyvsp[-4].str), (yyvsp[-2].vect), &(yyloc));
    }
#line 2503 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 64:
#line 533 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2512 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 65:
#line 541 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2521 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 66:
#line 546 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2530 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 67:
#line 554 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PUBLIC | MOD_CONST, &(yylsp[0]));
    }
#line 2538 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 68:
#line 558 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[-2].str), MOD_PUBLIC | MOD_CONST, &(yylsp[-2]));
        (yyval.id)->u_var.dflt_exp = (yyvsp[0].exp);
    }
#line 2547 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 71:
#line 571 "grammar.y" /* yacc.c:1648  */
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
#line 2564 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 74:
#line 592 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_ctor((yyvsp[-3].str), (yyvsp[-1].vect), NULL, &(yylsp[-3]));
    }
#line 2572 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 75:
#line 598 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 2578 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 77:
#line 604 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2587 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 78:
#line 609 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2596 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 79:
#line 617 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->u_var.is_param = true;
        (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
    }
#line 2606 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 80:
#line 626 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_func((yyvsp[-4].str), (yyvsp[-6].mod), (yyvsp[-2].vect), (yyvsp[0].id), NULL, &(yylsp[-4]));
    }
#line 2614 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 81:
#line 632 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PRIVATE; }
#line 2620 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 82:
#line 633 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC; }
#line 2626 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 83:
#line 634 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PAYABLE; }
#line 2632 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 84:
#line 635 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC | MOD_PAYABLE; }
#line 2638 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 85:
#line 639 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = NULL; }
#line 2644 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 86:
#line 640 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = (yyvsp[-1].id); }
#line 2650 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 89:
#line 647 "grammar.y" /* yacc.c:1648  */
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
#line 2680 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 90:
#line 676 "grammar.y" /* yacc.c:1648  */
    {
        /* We wanted to use a type expression, but we can not store size expressions
         * and declare it as an identifier. */
        (yyval.id) = id_new_param(NULL, (yyvsp[0].exp), &(yylsp[0]));
    }
#line 2690 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 91:
#line 682 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2703 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 92:
#line 693 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 2709 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 93:
#line 694 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 2715 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 94:
#line 699 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2724 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 95:
#line 704 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        /* Unlike state variables, local variables are referenced according to their
         * order of declaration. */
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2735 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 96:
#line 711 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2744 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 97:
#line 716 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2753 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 98:
#line 721 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2762 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 99:
#line 726 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2771 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 100:
#line 734 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, id_new_interface((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2779 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 101:
#line 741 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_interface(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2788 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 102:
#line 746 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2797 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 103:
#line 754 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, id_new_library((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2805 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 104:
#line 761 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-3].id)->mod = MOD_SYSTEM;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2818 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 105:
#line 770 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-1].id)->mod = MOD_SYSTEM;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2831 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 106:
#line 779 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-4].blk);

        (yyvsp[-3].id)->mod = MOD_SYSTEM;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2844 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 107:
#line 788 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);

        (yyvsp[-1].id)->mod = MOD_SYSTEM;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2857 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 118:
#line 813 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_null(&(yyloc));
    }
#line 2865 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 119:
#line 820 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_exp((yyvsp[-1].exp), &(yyloc));
    }
#line 2873 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 120:
#line 824 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2882 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 121:
#line 832 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2]));
    }
#line 2890 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 122:
#line 836 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), exp_new_binary((yyvsp[-2].op), (yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2])), &(yylsp[-2]));
    }
#line 2898 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 123:
#line 842 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 2904 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 124:
#line 843 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 2910 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 125:
#line 844 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 2916 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 126:
#line 845 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 2922 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 127:
#line 846 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 2928 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 128:
#line 847 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_AND; }
#line 2934 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 129:
#line 848 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_XOR; }
#line 2940 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 130:
#line 849 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_OR; }
#line 2946 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 131:
#line 850 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 2952 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 132:
#line 851 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 2958 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 133:
#line 856 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[0].stmt);
        id_add(LABELS, id_new_label((yyvsp[-2].str), (yyvsp[0].stmt), &(yylsp[-2])));
    }
#line 2967 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 134:
#line 861 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case((yyvsp[-1].exp), &(yyloc));
    }
#line 2975 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 135:
#line 865 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case(NULL, &(yyloc));
    }
#line 2983 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 136:
#line 872 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 2991 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 137:
#line 876 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-6].stmt);
        stmt_add(&(yyval.stmt)->u_if.elif_stmts, stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yylsp[-5])));
    }
#line 3000 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 138:
#line 881 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-2].stmt);
        (yyval.stmt)->u_if.else_blk = (yyvsp[0].blk);
    }
#line 3009 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 139:
#line 886 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3018 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 140:
#line 894 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, NULL, NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3026 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 141:
#line 898 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3034 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 142:
#line 902 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-3].stmt), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3042 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 143:
#line 906 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-4].stmt), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3050 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 144:
#line 910 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-3].id), &(yylsp[-3])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3058 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 145:
#line 914 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3066 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 146:
#line 918 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_exp((yyvsp[-4].exp), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3074 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 147:
#line 922 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3082 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 148:
#line 926 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3091 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 151:
#line 938 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 3097 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 153:
#line 944 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch(NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3105 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 154:
#line 948 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3113 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 155:
#line 952 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3122 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 156:
#line 959 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 3128 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 157:
#line 960 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 3134 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 158:
#line 965 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_switch(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3143 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 159:
#line 970 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3152 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 160:
#line 978 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_CONTINUE, NULL, &(yyloc));
    }
#line 3160 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 161:
#line 982 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_BREAK, NULL, &(yyloc));
    }
#line 3168 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 162:
#line 986 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return(NULL, &(yyloc));
    }
#line 3176 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 163:
#line 990 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return((yyvsp[-1].exp), &(yyloc));
    }
#line 3184 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 164:
#line 994 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_goto((yyvsp[-1].str), &(yylsp[-1]));
    }
#line 3192 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 165:
#line 1001 "grammar.y" /* yacc.c:1648  */
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
#line 3212 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 173:
#line 1030 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_blk((yyvsp[0].blk), &(yyloc));
    }
#line 3220 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 175:
#line 1038 "grammar.y" /* yacc.c:1648  */
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
#line 3235 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 177:
#line 1053 "grammar.y" /* yacc.c:1648  */
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
#line 3255 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 178:
#line 1071 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_DELETE; }
#line 3261 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 179:
#line 1072 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_INSERT; }
#line 3267 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 180:
#line 1073 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_REPLACE; }
#line 3273 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 181:
#line 1074 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_QUERY; }
#line 3279 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 182:
#line 1075 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_UPDATE; }
#line 3285 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 184:
#line 1081 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3293 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 185:
#line 1085 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
        (yyval.exp)->u_init.is_outmost = true;
    }
#line 3302 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 186:
#line 1093 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_alloc((yyvsp[0].exp), &(yylsp[0]));
    }
#line 3310 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 187:
#line 1097 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-3].exp);

        if ((yyval.exp)->u_alloc.size_exps == NULL)
            (yyval.exp)->u_alloc.size_exps = vector_new();

        exp_add((yyval.exp)->u_alloc.size_exps, (yyvsp[-1].exp));
    }
#line 3323 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 188:
#line 1109 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_init((yyvsp[-2].vect), &(yyloc));
    }
#line 3331 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 189:
#line 1116 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3340 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 190:
#line 1121 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3349 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 193:
#line 1131 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3357 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 195:
#line 1139 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_ternary((yyvsp[-4].exp), (yyvsp[-2].exp), (yyvsp[0].exp), &(yyloc));
    }
#line 3365 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 197:
#line 1147 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3373 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 199:
#line 1155 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3381 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 201:
#line 1163 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3389 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 203:
#line 1171 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_XOR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3397 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 205:
#line 1179 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3405 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 207:
#line 1187 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3413 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 208:
#line 1193 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_EQ; }
#line 3419 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 209:
#line 1194 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NE; }
#line 3425 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 211:
#line 1200 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3433 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 212:
#line 1206 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LT; }
#line 3439 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 213:
#line 1207 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GT; }
#line 3445 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 214:
#line 1208 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LE; }
#line 3451 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 215:
#line 1209 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GE; }
#line 3457 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 217:
#line 1215 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3465 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 218:
#line 1221 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 3471 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 219:
#line 1222 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 3477 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 221:
#line 1228 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3485 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 222:
#line 1234 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 3491 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 223:
#line 1235 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 3497 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 225:
#line 1241 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3505 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 226:
#line 1247 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 3511 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 227:
#line 1248 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 3517 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 228:
#line 1249 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 3523 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 230:
#line 1255 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_cast((yyvsp[-2].type), (yyvsp[0].exp), &(yylsp[-2]));
    }
#line 3531 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 232:
#line 1263 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary((yyvsp[-1].op), true, (yyvsp[0].exp), &(yyloc));
    }
#line 3539 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 233:
#line 1267 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3547 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 234:
#line 1273 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_INC; }
#line 3553 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 235:
#line 1274 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DEC; }
#line 3559 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 236:
#line 1275 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NEG; }
#line 3565 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 237:
#line 1276 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NOT; }
#line 3571 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 239:
#line 1282 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_array((yyvsp[-3].exp), (yyvsp[-1].exp), &(yyloc));
    }
#line 3579 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 240:
#line 1286 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(false, (yyvsp[-3].exp), (yyvsp[-1].vect), &(yyloc));
    }
#line 3587 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 241:
#line 1290 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_access((yyvsp[-2].exp), exp_new_id((yyvsp[0].str), &(yylsp[0])), &(yyloc));
    }
#line 3595 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 242:
#line 1294 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_INC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3603 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 243:
#line 1298 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_DEC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3611 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 244:
#line 1304 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 3617 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 246:
#line 1310 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3626 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 247:
#line 1315 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3635 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 249:
#line 1324 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yyloc));
    }
#line 3643 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 250:
#line 1328 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-1].exp);
    }
#line 3651 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 251:
#line 1332 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(true, exp_new_id((yyvsp[-3].str), &(yylsp[-3])), (yyvsp[-1].vect), &(yylsp[-3]));
    }
#line 3659 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 252:
#line 1339 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_null(&(yyloc));
    }
#line 3667 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 253:
#line 1343 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(true, &(yyloc));
    }
#line 3675 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 254:
#line 1347 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(false, &(yyloc));
    }
#line 3683 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 255:
#line 1351 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 10);
    }
#line 3692 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 256:
#line 1356 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 8);
    }
#line 3701 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 257:
#line 1361 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_int(0, &(yyloc));
        mpz_set_str(val_mpz(&(yyval.exp)->u_lit.val), (yyvsp[0].str), 16);
    }
#line 3710 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 258:
#line 1366 "grammar.y" /* yacc.c:1648  */
    {
        double v;

        sscanf((yyvsp[0].str), "%lf", &v);
        (yyval.exp) = exp_new_lit_f64(v, &(yyloc));
    }
#line 3721 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 259:
#line 1373 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_str((yyvsp[0].str), &(yyloc));
    }
#line 3729 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 260:
#line 1379 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("contract"); }
#line 3735 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 261:
#line 1380 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("import"); }
#line 3741 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 262:
#line 1381 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("index"); }
#line 3747 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 263:
#line 1382 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("interface"); }
#line 3753 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 264:
#line 1383 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("library"); }
#line 3759 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 265:
#line 1384 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("table"); }
#line 3765 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 266:
#line 1385 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("view"); }
#line 3771 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 267:
#line 1390 "grammar.y" /* yacc.c:1648  */
    {
        if (strlen((yyvsp[0].str)) > NAME_MAX_LEN)
            ERROR(ERROR_TOO_LONG_ID, &(yylsp[0]), NAME_MAX_LEN, strlen((yyvsp[0].str)));

        (yyval.str) = (yyvsp[0].str);
    }
#line 3782 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;


#line 3786 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
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
#line 1399 "grammar.y" /* yacc.c:1907  */


static void
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
