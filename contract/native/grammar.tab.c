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
    K_INT256 = 318,
    K_INTERFACE = 319,
    K_LIBRARY = 320,
    K_MAP = 321,
    K_NEW = 322,
    K_NULL = 323,
    K_PAYABLE = 324,
    K_PUBLIC = 325,
    K_REPLACE = 326,
    K_RETURN = 327,
    K_SELECT = 328,
    K_STRING = 329,
    K_STRUCT = 330,
    K_SWITCH = 331,
    K_TABLE = 332,
    K_TRUE = 333,
    K_TYPE = 334,
    K_UINT = 335,
    K_UINT8 = 336,
    K_UINT16 = 337,
    K_UINT32 = 338,
    K_UINT64 = 339,
    K_UINT128 = 340,
    K_UINT256 = 341,
    K_UPDATE = 342,
    K_VIEW = 343
  };
#endif

/* Value type.  */
#if ! defined YYSTYPE && ! defined YYSTYPE_IS_DECLARED

union YYSTYPE
{
#line 154 "grammar.y" /* yacc.c:355  */

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

#line 241 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:355  */
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

#line 271 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:358  */

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
#define YYLAST   1935

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  112
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  92
/* YYNRULES -- Number of rules.  */
#define YYNRULES  270
/* YYNSTATES -- Number of states.  */
#define YYNSTATES  449

/* YYTRANSLATE[YYX] -- Symbol number corresponding to YYX as returned
   by yylex, with out-of-bounds checking.  */
#define YYUNDEFTOK  2
#define YYMAXUTOK   343

#define YYTRANSLATE(YYX)                                                \
  ((unsigned int) (YYX) <= YYMAXUTOK ? yytranslate[YYX] : YYUNDEFTOK)

/* YYTRANSLATE[TOKEN-NUM] -- Symbol number corresponding to TOKEN-NUM
   as returned by yylex, without out-of-bounds checking.  */
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
       0,   260,   260,   261,   265,   266,   267,   268,   269,   270,
     274,   278,   287,   306,   307,   314,   319,   324,   329,   334,
     339,   347,   348,   353,   361,   362,   370,   371,   383,   395,
     399,   405,   415,   416,   417,   418,   419,   420,   421,   422,
     423,   424,   425,   426,   427,   428,   429,   430,   431,   434,
     435,   439,   440,   455,   459,   471,   472,   476,   480,   494,
     495,   499,   503,   511,   520,   532,   536,   544,   549,   557,
     561,   568,   570,   574,   590,   591,   595,   602,   603,   607,
     612,   620,   629,   636,   637,   638,   639,   643,   644,   645,
     649,   650,   679,   685,   697,   698,   702,   707,   714,   719,
     724,   729,   737,   744,   749,   757,   764,   773,   782,   791,
     803,   804,   805,   806,   807,   808,   809,   810,   811,   812,
     816,   823,   827,   835,   839,   846,   847,   848,   849,   850,
     851,   852,   853,   854,   855,   859,   864,   868,   875,   879,
     884,   889,   897,   901,   905,   909,   913,   917,   921,   925,
     929,   937,   938,   942,   943,   947,   951,   955,   963,   964,
     968,   973,   981,   985,   989,   993,   997,  1004,  1023,  1024,
    1025,  1026,  1027,  1028,  1029,  1033,  1040,  1041,  1055,  1056,
    1075,  1076,  1077,  1078,  1079,  1083,  1084,  1088,  1096,  1100,
    1112,  1119,  1124,  1132,  1133,  1134,  1141,  1142,  1149,  1150,
    1157,  1158,  1165,  1166,  1173,  1174,  1181,  1182,  1189,  1190,
    1197,  1198,  1202,  1203,  1210,  1211,  1212,  1213,  1217,  1218,
    1225,  1226,  1230,  1231,  1238,  1239,  1243,  1244,  1251,  1252,
    1253,  1257,  1258,  1265,  1266,  1270,  1277,  1278,  1279,  1280,
    1284,  1285,  1289,  1293,  1297,  1301,  1308,  1309,  1313,  1318,
    1326,  1327,  1331,  1335,  1342,  1346,  1350,  1354,  1361,  1368,
    1375,  1382,  1389,  1390,  1391,  1392,  1393,  1394,  1395,  1399,
    1406
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
  "\"int32\"", "\"int64\"", "\"int128\"", "\"int256\"", "\"interface\"",
  "\"library\"", "\"map\"", "\"new\"", "\"null\"", "\"payable\"",
  "\"public\"", "\"replace\"", "\"return\"", "\"select\"", "\"string\"",
  "\"struct\"", "\"switch\"", "\"table\"", "\"true\"", "\"type\"",
  "\"uint\"", "\"uint8\"", "\"uint16\"", "\"uint32\"", "\"uint64\"",
  "\"uint128\"", "\"uint256\"", "\"update\"", "\"view\"", "'!'", "'|'",
  "'^'", "'&'", "'<'", "'>'", "'+'", "'-'", "'*'", "'/'", "'%'", "'('",
  "')'", "'.'", "'{'", "'}'", "';'", "'='", "','", "'['", "']'", "':'",
  "'?'", "$accept", "root", "component", "import", "contract", "impl_opt",
  "contract_body", "variable", "var_qual", "var_decl", "var_spec",
  "var_type", "prim_type", "declarator_list", "declarator", "size_opt",
  "var_init", "compound", "struct", "field_list", "enumeration",
  "enum_list", "enumerator", "comma_opt", "function", "func_spec",
  "ctor_spec", "param_list_opt", "param_list", "param_decl", "udf_spec",
  "modifier_opt", "return_opt", "return_list", "return_decl", "block",
  "blk_decl", "interface", "interface_body", "library", "library_body",
  "statement", "empty_stmt", "exp_stmt", "assign_stmt", "assign_op",
  "label_stmt", "if_stmt", "loop_stmt", "init_stmt", "cond_exp",
  "switch_stmt", "switch_blk", "case_blk", "jump_stmt", "ddl_stmt",
  "ddl_prefix", "blk_stmt", "expression", "sql_exp", "sql_prefix",
  "new_exp", "alloc_exp", "initializer", "elem_list", "init_elem",
  "ternary_exp", "or_exp", "and_exp", "bit_or_exp", "bit_xor_exp",
  "bit_and_exp", "eq_exp", "eq_op", "cmp_exp", "cmp_op", "add_exp",
  "add_op", "shift_exp", "shift_op", "mul_exp", "mul_op", "cast_exp",
  "unary_exp", "unary_op", "post_exp", "arg_list_opt", "arg_list",
  "prim_exp", "literal", "non_reserved_token", "identifier", YY_NULLPTR
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
     105,   313,    40,   313,   313,   112,  -273,  -273,  -273,  -273,
    -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,
       0,  -273,   -77,   -13,  -273,  -273,  -273,  -273,   313,     3,
      -8,    -8,  -273,   982,  -273,   116,     4,    30,   -36,   -80,
     -31,  -273,  -273,  -273,  1847,  -273,   237,  -273,  -273,  -273,
    -273,  -273,  -273,  -273,   121,  1808,  -273,   284,  -273,  -273,
    -273,  -273,  -273,  -273,  -273,  -273,  1046,  -273,  -273,  -273,
      92,   763,  -273,  -273,  -273,  -273,  -273,   119,  -273,  -273,
     123,  -273,  -273,   313,  -273,    87,   388,   216,  -273,  -273,
     -66,  -273,   313,  -273,   122,   124,  1847,  -273,   125,   153,
    -273,  -273,  -273,  -273,  -273,  1556,   127,   128,   133,  -273,
    -273,  1847,   130,  -273,   131,  -273,  -273,  -273,  -273,  -273,
    -273,  -273,   154,   144,  1611,   145,   -24,   141,  -273,    58,
    -273,    14,   313,    19,  -273,  1174,  -273,  -273,   240,  -273,
      18,  -273,  -273,  -273,  1655,  -273,  1420,  -273,  -273,  -273,
    -273,  -273,   493,  -273,  -273,  -273,  -273,  -273,   189,  -273,
    -273,  -273,  -273,   251,  -273,    89,  -273,   252,  -273,  -273,
      -6,   249,   179,   180,   181,   111,    35,    -9,   183,   109,
    -273,   248,  1655,    39,  -273,  -273,    90,   167,   267,  -273,
    -273,   313,   169,  -273,   175,    56,  -273,  -273,  -273,  -273,
     313,  1611,   313,   182,   173,  -273,  1847,  -273,  -273,  -273,
     313,    10,  -273,  -273,  -273,  -273,  -273,  -273,  -273,  -273,
     178,   806,  -273,   194,   190,  1556,  1215,  -273,   176,  -273,
     200,  -273,    76,   199,  1556,   895,  -273,  1556,  -273,   208,
     -61,  -273,  -273,  -273,  -273,   -27,   205,  -273,  1556,  1556,
     207,  1611,  1556,  1611,  1611,  1611,  1611,  -273,  -273,  1611,
    -273,  -273,  -273,  -273,  1611,  -273,  -273,  1611,  -273,  -273,
    1611,  -273,  -273,  -273,  1611,  -273,  -273,  -273,  -273,  -273,
    -273,  -273,  -273,  -273,  -273,  1556,  -273,  -273,  -273,  1710,
     313,  1611,   703,  -273,   210,   212,  -273,   214,  1847,  1847,
    -273,  1556,   133,   213,    -9,   133,  -273,  1847,   222,   200,
    -273,  -273,   909,    -3,  -273,  -273,   909,   -32,   223,   901,
    -273,  -273,   -29,  -273,   219,  -273,  -273,   220,  1611,  1710,
    -273,  -273,    31,  -273,  -273,   598,   221,  1655,  -273,   232,
    -273,  -273,    84,  -273,  -273,   249,   -82,   179,   180,   181,
     111,    35,    -9,   183,   109,  -273,   107,  -273,   241,   226,
    -273,   229,  -273,   221,  -273,   313,   239,  1611,   243,   236,
    1110,  -273,  -273,  -273,  1492,  -273,  1268,   246,  1754,  1344,
    1754,   119,   119,  1215,   250,  1215,   -15,   255,   254,  -273,
    -273,  -273,  1556,  -273,  1611,  -273,  -273,  1710,  -273,  -273,
    -273,   330,  -273,  -273,  -273,   253,  1847,  -273,  -273,   256,
     245,   119,    52,  -273,    22,   119,    61,    71,  -273,  -273,
    -273,  -273,  -273,  -273,  -273,  -273,    73,  -273,  -273,  -273,
      81,  1847,  1611,  -273,   119,   119,  -273,   119,   119,   119,
    -273,   245,   258,  -273,  -273,  -273,  -273,  -273,  -273
};

  /* YYDEFACT[STATE-NUM] -- Default reduction number in state STATE-NUM.
     Performed when YYTABLE does not specify something else to do.  Zero
     means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       0,     0,     0,     0,     0,     0,     2,     4,     5,     6,
       3,   269,   262,   263,   264,   265,   266,   267,   268,   270,
      13,    10,     0,     0,     1,     7,     8,     9,     0,     0,
      83,    83,    14,    83,    85,    84,     0,     0,    83,     0,
      83,    32,    33,    34,     0,    50,     0,    41,    35,    36,
      37,    38,    39,    40,     0,    84,    49,     0,    48,    42,
      43,    44,    45,    46,    47,    11,    83,    15,    21,    24,
       0,     0,    29,    16,    59,    60,    17,     0,    74,    75,
      30,    86,   103,     0,   102,     0,     0,     0,   107,   105,
       0,    25,     0,    30,     0,     0,     0,    22,     0,     0,
      12,    18,    19,    20,    26,     0,     0,    28,    51,    53,
      73,    77,     0,   104,     0,   260,   259,   257,   258,   261,
     236,   237,     0,     0,     0,     0,     0,     0,   180,     0,
     256,     0,     0,     0,   181,     0,   254,   182,     0,   183,
       0,   255,   184,   239,     0,   238,     0,    94,   120,    96,
      97,   175,     0,    98,   110,   111,   112,   113,   114,   115,
     116,   117,   118,     0,   119,     0,   176,     0,   178,   185,
     196,   198,   200,   202,   204,   206,   208,   212,   218,   222,
     226,   231,     0,   233,   240,   250,   251,     0,     0,   109,
      66,     0,     0,    62,     0,     0,    57,   231,   251,    23,
       0,    55,     0,     0,    78,    79,    77,   122,   169,   163,
       0,     0,   162,   171,   168,   173,   137,   172,   170,   174,
       0,     0,   142,     0,     0,     0,     0,   188,   186,   187,
      30,   164,     0,     0,     0,     0,   155,     0,   235,     0,
       0,    95,    99,   100,   101,     0,     0,   121,     0,     0,
       0,     0,     0,     0,     0,     0,     0,   210,   211,     0,
     216,   217,   214,   215,     0,   220,   221,     0,   224,   225,
       0,   228,   229,   230,     0,   125,   126,   127,   128,   129,
     130,   131,   132,   133,   134,     0,   234,   244,   245,   246,
       0,     0,     0,   106,     0,    71,    67,    69,     0,     0,
      27,     0,    52,     0,    56,    81,    76,     0,     0,     0,
     136,   150,     0,     0,   151,   152,     0,     0,   185,   251,
     166,   141,     0,   194,    71,   191,   193,   251,     0,   246,
     165,   157,     0,   158,   160,     0,     0,     0,   252,     0,
     140,   167,     0,   177,   179,   199,     0,   201,   203,   205,
     207,   209,   213,   219,   223,   227,     0,   248,     0,   247,
     243,     0,   135,   251,   108,    72,     0,     0,     0,     0,
       0,    58,    54,    80,    87,   153,     0,     0,     0,     0,
       0,     0,     0,    72,     0,     0,     0,     0,     0,   159,
     161,   232,     0,   123,     0,   124,   242,     0,   241,    68,
      65,    70,    31,    63,    61,     0,     0,    92,    82,    89,
      90,     0,     0,   154,     0,     0,     0,     0,   143,   138,
     192,   190,   195,   189,   253,   156,     0,   197,   249,    64,
       0,     0,    55,   146,     0,     0,   144,     0,     0,     0,
      88,    91,     0,   147,   149,   145,   148,   139,    93
};

  /* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -273,  -273,  -273,   354,   355,  -273,  -273,   296,   -26,   -37,
    -210,   -19,   224,  -273,    13,   -67,  -273,   -45,  -273,  -273,
    -273,  -273,     6,    50,   309,  -273,  -273,   177,  -273,    72,
     126,  -273,  -273,   -25,   -46,   -35,  -273,   381,  -273,  -273,
    -273,  -142,   185,  -273,   186,  -273,   162,  -273,  -273,  -273,
      82,  -273,    20,  -273,  -273,  -273,  -273,  -273,  -137,   -97,
    -273,  -272,  -273,   274,  -273,  -167,  -208,    21,   159,   158,
     160,   187,  -112,  -273,   198,  -273,  -195,  -273,   146,  -273,
     165,  -273,   163,   -81,  -273,  -161,   103,  -273,  -273,  -273,
    -273,    -1
};

  /* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     5,     6,     7,     8,    29,    66,    67,    68,    69,
      70,    92,    72,   107,   108,   303,   195,    73,    74,   370,
      75,   295,   296,   366,    76,    77,    78,   203,   204,   205,
      79,    37,   408,   409,   410,   151,   152,     9,    38,    10,
      40,   153,   154,   155,   156,   285,   157,   158,   159,   316,
     376,   160,   236,   335,   161,   162,   163,   164,   165,   166,
     167,   168,   228,   323,   324,   325,   169,   170,   171,   172,
     173,   174,   175,   259,   176,   264,   177,   267,   178,   270,
     179,   274,   180,   197,   182,   183,   358,   359,   184,   185,
      19,   198
};

  /* YYTABLE[YYPACT[STATE-NUM]] -- What to do in state STATE-NUM.  If
     positive, shift that token.  If negative, reduce the rule whose
     number is the opposite.  If YYTABLE_NINF, syntax error.  */
static const yytype_int16 yytable[] =
{
      20,   232,    22,    23,    88,   181,   304,    91,   196,   240,
     244,   313,   211,   318,    71,   220,   251,   357,   326,   233,
     224,   102,   380,    86,   339,   249,    30,    32,   394,    97,
      87,   213,    80,    34,    35,   257,   258,    86,    34,    35,
     338,   150,   110,    93,   188,    95,   249,    71,    21,   287,
     288,   378,    28,   214,    93,   189,    99,   357,   260,   261,
     149,    34,    35,   238,   215,    80,   287,   288,    84,   352,
     109,   181,   382,    89,   248,   249,    86,   192,   249,    83,
     265,   266,   112,   361,   317,   186,   265,   266,   322,   369,
      31,   109,   202,   -30,   423,    93,   222,   332,   287,   288,
     240,   286,   104,   105,   377,   252,    33,   243,   377,    82,
      93,   342,    24,   217,   221,   346,   227,    86,   234,   225,
     310,   235,   289,   435,   290,   428,   242,   -30,   262,   263,
     291,   223,   388,   386,   230,   218,   257,   258,   249,   289,
     181,   290,     1,   -30,   350,   -30,   219,   291,   356,     1,
     362,   186,   343,   434,   -30,   -30,    36,    39,     2,   249,
     405,   300,   437,   301,    85,     2,    90,   -30,   249,     3,
       4,   289,   438,   290,   439,   326,     3,   326,   -30,   291,
     249,   330,   440,   249,   312,    81,   427,   202,   431,   393,
     297,   249,   113,   390,   247,   248,   249,   104,   105,   109,
     292,   109,   268,   269,   371,    93,   271,   272,   273,   309,
     340,   181,   395,   302,   249,   305,   420,   414,   422,   417,
     319,    96,    86,   111,   187,   327,   190,   191,   194,   193,
     206,   208,   199,   245,   336,   200,   207,   304,    94,   412,
      11,   201,   416,    11,   115,   116,   117,   118,   119,   209,
     212,   216,   246,   250,   181,   426,   391,   275,   276,   277,
     278,   279,   280,   281,   282,   283,   284,   120,   121,   254,
     253,   255,   293,   256,    12,   294,   298,    12,   299,   368,
     307,   128,   311,   306,   328,    98,   130,    11,   202,   360,
      13,   363,    14,    13,   321,    14,   134,    93,    93,   320,
     329,    15,    16,   331,    15,    16,    93,   135,   136,   337,
     341,   137,   344,   139,    17,   364,    11,    17,   141,   365,
     367,    12,   372,   374,   381,    18,   383,   142,    18,   143,
     385,   292,   392,   397,   363,   144,   145,    13,   398,    14,
     146,   403,   396,   400,   402,   231,   418,   419,    15,    16,
      12,   413,   251,   432,   421,   407,   424,   235,   429,    25,
      26,    17,   101,   431,   297,   442,    13,   448,    14,    93,
     239,   399,    18,    93,   384,   103,   433,    15,    16,   373,
     436,   430,   327,   308,   327,   441,    27,   407,   401,   114,
      17,    11,   115,   116,   117,   118,   119,   334,   379,   443,
     444,    18,   445,   446,   447,    93,   314,   315,   425,   229,
     345,   347,   407,   353,   348,   120,   121,    41,   122,    42,
     123,    43,   124,    44,   125,    12,   126,    45,   127,   128,
      93,   129,   387,    46,   130,   354,   131,   355,   132,   133,
       0,    13,   349,    14,   134,    47,    48,    49,    50,    51,
      52,    53,    15,    16,    54,   135,   136,   351,     0,   137,
     138,   139,    56,     0,   140,    17,   141,    57,    58,    59,
      60,    61,    62,    63,    64,   142,    18,   143,     0,     0,
       0,     0,     0,   144,   145,     0,     0,     0,   146,     0,
       0,    86,   147,   148,   114,     0,    11,   115,   116,   117,
     118,   119,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     120,   121,    41,   122,    42,   123,    43,   124,    44,   125,
      12,   126,    45,   127,   128,     0,   129,     0,    46,   130,
       0,   131,     0,   132,   133,     0,    13,     0,    14,   134,
      47,    48,    49,    50,    51,    52,    53,    15,    16,    54,
     135,   136,     0,     0,   137,   138,   139,    56,     0,   140,
      17,   141,    57,    58,    59,    60,    61,    62,    63,    64,
     142,    18,   143,     0,     0,     0,     0,     0,   144,   145,
       0,     0,     0,   146,     0,     0,    86,   241,   148,   114,
       0,    11,   115,   116,   117,   118,   119,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   120,   121,     0,   122,     0,
     123,     0,   124,     0,   125,    12,   126,     0,   127,   128,
       0,   129,     0,     0,   130,     0,   131,     0,   132,   133,
       0,    13,     0,    14,   134,     0,     0,     0,     0,     0,
       0,     0,    15,    16,     0,   135,   136,     0,     0,   137,
     138,   139,     0,     0,   140,    17,   141,     0,     0,     0,
       0,     0,     0,     0,     0,   142,    18,   143,     0,     0,
       0,     0,     0,   144,   145,     0,     0,     0,   146,     0,
       0,    86,   389,   148,   114,     0,    11,   115,   116,   117,
     118,   119,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     120,   121,     0,   122,     0,   123,     0,   124,     0,   125,
      12,   126,     0,   127,   128,     0,   129,     0,     0,   130,
       0,   131,     0,   132,   133,     0,    13,     0,    14,   134,
       0,     0,     0,     0,   106,     0,    11,    15,    16,     0,
     135,   136,     0,     0,   137,   138,   139,     0,     0,   140,
      17,   141,     0,     0,     0,     0,     0,     0,     0,     0,
     142,    18,   143,     0,     0,     0,     0,     0,   144,   145,
      12,     0,     0,   146,     0,     0,    86,     0,   148,    11,
     115,   116,   117,   118,   119,     0,    13,     0,    14,     0,
       0,     0,     0,     0,     0,     0,     0,    15,    16,     0,
       0,     0,     0,   120,   121,    41,     0,    42,     0,    43,
      17,     0,     0,    12,     0,    45,     0,   128,     0,     0,
       0,    18,   130,     0,     0,     0,     0,     0,     0,    13,
       0,    14,   134,    47,    48,    49,    50,    51,    52,    53,
      15,    16,    54,   135,   136,     0,     0,   137,     0,   139,
      56,     0,     0,    17,   141,     0,    58,    59,    60,    61,
      62,    63,    64,   142,    18,   143,     0,     0,    11,     0,
       0,   144,   145,     0,   -30,     0,   146,     0,     0,     0,
       0,   148,    11,   115,   116,   117,   118,   119,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,   124,
       0,     0,    12,     0,     0,   127,   120,   121,   -30,     0,
       0,     0,     0,     0,     0,     0,    12,     0,    13,     0,
      14,     0,     0,     0,   -30,   130,   -30,     0,     0,    15,
      16,     0,    13,     0,    14,   -30,   -30,     0,     0,     0,
       0,     0,    17,    15,    16,     0,   210,   136,   -30,     0,
       0,     0,     0,    18,     0,    11,    17,   141,     0,   -30,
       0,     0,     0,     0,     0,     0,     0,    18,   143,   333,
       0,     0,     0,     0,   144,   145,     0,     0,     0,   146,
       0,    41,     0,    42,   375,    43,     0,    44,     0,    12,
       0,    45,     0,     0,     0,     0,     0,    46,     0,     0,
       0,     0,     0,     0,     0,    13,     0,    14,     0,    47,
      48,    49,    50,    51,    52,    53,    15,    16,    54,    11,
       0,    34,    55,     0,     0,     0,    56,     0,     0,    17,
       0,    57,    58,    59,    60,    61,    62,    63,    64,     0,
      18,     0,     0,     0,     0,    41,     0,    42,     0,    43,
       0,    44,     0,    12,     0,    45,    65,     0,     0,     0,
       0,    46,     0,     0,     0,     0,     0,     0,     0,    13,
       0,    14,     0,    47,    48,    49,    50,    51,    52,    53,
      15,    16,    54,    11,     0,    34,    55,     0,     0,     0,
      56,     0,     0,    17,     0,    57,    58,    59,    60,    61,
      62,    63,    64,     0,    18,     0,     0,     0,     0,    41,
       0,    42,     0,    43,     0,     0,     0,    12,     0,    45,
     100,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,    13,     0,    14,     0,    47,    48,    49,
      50,    51,    52,    53,    15,    16,    54,    11,     0,     0,
       0,     0,     0,     0,    56,     0,     0,    17,     0,     0,
      58,    59,    60,    61,    62,    63,    64,     0,    18,     0,
       0,     0,     0,    41,     0,    42,     0,    43,     0,     0,
       0,    12,     0,    45,   404,     0,     0,     0,    11,   115,
     116,   117,   118,   119,     0,     0,     0,    13,     0,    14,
       0,    47,    48,    49,    50,    51,    52,    53,    15,    16,
      54,     0,   120,   121,     0,     0,     0,     0,    56,     0,
       0,    17,    12,     0,    58,    59,    60,    61,    62,    63,
      64,   130,    18,     0,     0,     0,     0,     0,    13,     0,
      14,    11,   115,   116,   117,   118,   119,   226,     0,    15,
      16,     0,   210,   136,     0,     0,     0,     0,     0,     0,
       0,     0,    17,   141,     0,   120,   121,     0,     0,     0,
       0,     0,     0,    18,   143,    12,     0,     0,     0,   128,
     144,   145,     0,     0,   130,   146,     0,     0,   226,     0,
       0,    13,     0,    14,   134,     0,     0,     0,     0,     0,
       0,     0,    15,    16,     0,   135,   136,     0,     0,   137,
       0,   139,     0,     0,     0,    17,   141,    11,   115,   116,
     117,   118,   119,     0,     0,   142,    18,   143,     0,     0,
       0,     0,     0,   144,   145,     0,     0,     0,   146,   411,
       0,   120,   121,     0,     0,     0,     0,     0,     0,     0,
       0,    12,     0,     0,     0,   128,     0,     0,     0,     0,
     130,     0,     0,     0,     0,     0,     0,    13,     0,    14,
     134,     0,     0,     0,     0,     0,     0,     0,    15,    16,
       0,   135,   136,     0,     0,   137,     0,   139,     0,     0,
       0,    17,   141,    11,   115,   116,   117,   118,   119,     0,
       0,   142,    18,   143,     0,     0,     0,     0,     0,   144,
     145,     0,     0,     0,   146,   415,     0,   120,   121,    41,
       0,    42,     0,    43,     0,     0,     0,    12,     0,    45,
       0,   128,     0,     0,     0,     0,   130,     0,     0,     0,
       0,     0,     0,    13,     0,    14,   134,    47,    48,    49,
      50,    51,    52,    53,    15,    16,     0,   135,   136,     0,
       0,   137,     0,   139,    56,    11,     0,    17,   141,     0,
      58,    59,    60,    61,    62,    63,    64,   142,    18,   143,
       0,     0,     0,     0,     0,   144,   145,     0,     0,     0,
     146,    41,     0,    42,     0,    43,     0,     0,     0,    12,
       0,    45,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,    13,     0,    14,     0,    47,
      48,    49,    50,    51,    52,    53,    15,    16,    54,    11,
     115,   116,   117,   118,   119,     0,    56,     0,     0,    17,
       0,     0,    58,    59,    60,    61,    62,    63,    64,     0,
      18,     0,     0,   120,   121,     0,     0,     0,     0,     0,
       0,     0,   406,    12,     0,     0,     0,   128,     0,     0,
       0,     0,   130,     0,     0,     0,     0,     0,     0,    13,
       0,    14,   134,     0,    11,   115,   116,   117,   118,   119,
      15,    16,     0,   135,   136,     0,     0,   137,     0,   139,
       0,     0,     0,    17,   141,     0,     0,     0,   120,   121,
       0,     0,     0,   142,    18,   143,     0,     0,    12,     0,
       0,   144,   145,     0,     0,     0,   146,   130,    11,   115,
     116,   117,   118,   119,    13,     0,    14,     0,     0,     0,
       0,     0,     0,     0,     0,    15,    16,     0,   210,   136,
       0,     0,   120,   121,     0,     0,     0,     0,    17,   141,
       0,     0,    12,     0,     0,     0,     0,     0,     0,    18,
     143,   130,     0,     0,     0,     0,   144,   145,    13,     0,
      14,   146,     0,    11,   115,   116,   117,   118,   119,    15,
      16,     0,   210,   136,     0,     0,     0,     0,     0,     0,
       0,     0,    17,   141,     0,     0,     0,   120,   121,     0,
       0,     0,     0,    18,   143,     0,     0,    12,     0,     0,
     144,   145,     0,     0,     0,   237,   130,    11,   115,   116,
     117,   118,   119,    13,     0,    14,     0,     0,     0,     0,
       0,     0,     0,     0,    15,    16,     0,   135,   136,     0,
       0,     0,     0,     0,     0,     0,     0,    17,   141,     0,
       0,    12,     0,     0,     0,     0,     0,     0,    18,   143,
     130,     0,     0,     0,     0,   144,   145,    13,     0,    14,
     146,    11,     0,     0,     0,     0,     0,     0,    15,    16,
       0,   210,   136,     0,     0,     0,     0,     0,     0,     0,
       0,    17,   141,     0,     0,     0,     0,    41,     0,    42,
       0,    43,    18,    44,     0,    12,     0,    45,     0,     0,
      11,     0,     0,     0,   237,     0,     0,     0,     0,     0,
       0,    13,     0,    14,     0,    47,    48,    49,    50,    51,
      52,    53,    15,    16,    54,     0,    41,    81,    42,     0,
      43,     0,    56,     0,    12,    17,    45,     0,    58,    59,
      60,    61,    62,    63,    64,     0,    18,     0,     0,     0,
      13,     0,    14,     0,    47,    48,    49,    50,    51,    52,
      53,    15,    16,    54,     0,     0,     0,     0,     0,     0,
       0,    56,     0,     0,    17,     0,     0,    58,    59,    60,
      61,    62,    63,    64,     0,    18
};

static const yytype_int16 yycheck[] =
{
       1,   138,     3,     4,    39,    86,   201,    44,   105,   146,
     152,   221,   124,   221,    33,     1,    22,   289,   226,     1,
       1,    66,    54,   103,    51,   107,   103,    28,   110,    55,
     110,    55,    33,    69,    70,    25,    26,   103,    69,    70,
     101,    86,    77,    44,   110,    46,   107,    66,     8,    27,
      28,    54,    52,    77,    55,    90,    57,   329,    23,    24,
      86,    69,    70,   144,    88,    66,    27,    28,   104,   264,
      71,   152,   101,   104,   106,   107,   103,    96,   107,    49,
      95,    96,    83,   291,   221,    86,    95,    96,   225,   299,
     103,    92,   111,     3,   109,    96,   131,   234,    27,    28,
     237,   182,   105,   106,   312,   111,   103,   152,   316,   105,
     111,   248,     0,    55,   100,   252,   135,   103,   100,   100,
     110,   103,   100,   101,   102,   397,   152,    37,    93,    94,
     108,   132,   101,   328,   135,    77,    25,    26,   107,   100,
     221,   102,    37,    53,   256,    55,    88,   108,   285,    37,
     292,   152,   249,   101,    64,    65,    30,    31,    53,   107,
     370,   105,   101,   107,    38,    53,    40,    77,   107,    64,
      65,   100,   101,   102,   101,   383,    64,   385,    88,   108,
     107,   105,   101,   107,   221,    69,   394,   206,   107,   105,
     191,   107,   105,   335,   105,   106,   107,   105,   106,   200,
     110,   202,    19,    20,   301,   206,    97,    98,    99,   210,
     245,   292,   105,   200,   107,   202,   383,   378,   385,   380,
     221,   100,   103,   100,     8,   226,   104,   103,    75,   104,
     100,    77,   105,    44,   235,   107,   105,   432,     1,   376,
       3,   108,   379,     3,     4,     5,     6,     7,     8,   105,
     105,   110,     1,     1,   335,   392,   337,     9,    10,    11,
      12,    13,    14,    15,    16,    17,    18,    27,    28,    90,
      21,    91,   105,    92,    37,     8,   107,    37,   103,   298,
     107,    41,   104,   101,   108,     1,    46,     3,   307,   290,
      53,   292,    55,    53,   104,    55,    56,   298,   299,   105,
     100,    64,    65,   104,    64,    65,   307,    67,    68,   101,
     105,    71,   105,    73,    77,   105,     3,    77,    78,   107,
     106,    37,   109,   101,   101,    88,   107,    87,    88,    89,
     110,   110,   100,   107,   335,    95,    96,    53,   109,    55,
     100,   105,   101,   104,   101,   105,   381,   382,    64,    65,
      37,   105,    22,   108,   104,   374,   101,   103,   105,     5,
       5,    77,    66,   107,   365,   432,    53,   109,    55,   370,
     146,   365,    88,   374,   324,    66,   411,    64,    65,   307,
     415,   406,   383,   206,   385,   431,     5,   406,   367,     1,
      77,     3,     4,     5,     6,     7,     8,   235,   316,   434,
     435,    88,   437,   438,   439,   406,   221,   221,   388,   135,
     251,   253,   431,   267,   254,    27,    28,    29,    30,    31,
      32,    33,    34,    35,    36,    37,    38,    39,    40,    41,
     431,    43,   329,    45,    46,   270,    48,   274,    50,    51,
      -1,    53,   255,    55,    56,    57,    58,    59,    60,    61,
      62,    63,    64,    65,    66,    67,    68,   259,    -1,    71,
      72,    73,    74,    -1,    76,    77,    78,    79,    80,    81,
      82,    83,    84,    85,    86,    87,    88,    89,    -1,    -1,
      -1,    -1,    -1,    95,    96,    -1,    -1,    -1,   100,    -1,
      -1,   103,   104,   105,     1,    -1,     3,     4,     5,     6,
       7,     8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      27,    28,    29,    30,    31,    32,    33,    34,    35,    36,
      37,    38,    39,    40,    41,    -1,    43,    -1,    45,    46,
      -1,    48,    -1,    50,    51,    -1,    53,    -1,    55,    56,
      57,    58,    59,    60,    61,    62,    63,    64,    65,    66,
      67,    68,    -1,    -1,    71,    72,    73,    74,    -1,    76,
      77,    78,    79,    80,    81,    82,    83,    84,    85,    86,
      87,    88,    89,    -1,    -1,    -1,    -1,    -1,    95,    96,
      -1,    -1,    -1,   100,    -1,    -1,   103,   104,   105,     1,
      -1,     3,     4,     5,     6,     7,     8,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    30,    -1,
      32,    -1,    34,    -1,    36,    37,    38,    -1,    40,    41,
      -1,    43,    -1,    -1,    46,    -1,    48,    -1,    50,    51,
      -1,    53,    -1,    55,    56,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    64,    65,    -1,    67,    68,    -1,    -1,    71,
      72,    73,    -1,    -1,    76,    77,    78,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    87,    88,    89,    -1,    -1,
      -1,    -1,    -1,    95,    96,    -1,    -1,    -1,   100,    -1,
      -1,   103,   104,   105,     1,    -1,     3,     4,     5,     6,
       7,     8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      27,    28,    -1,    30,    -1,    32,    -1,    34,    -1,    36,
      37,    38,    -1,    40,    41,    -1,    43,    -1,    -1,    46,
      -1,    48,    -1,    50,    51,    -1,    53,    -1,    55,    56,
      -1,    -1,    -1,    -1,     1,    -1,     3,    64,    65,    -1,
      67,    68,    -1,    -1,    71,    72,    73,    -1,    -1,    76,
      77,    78,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      87,    88,    89,    -1,    -1,    -1,    -1,    -1,    95,    96,
      37,    -1,    -1,   100,    -1,    -1,   103,    -1,   105,     3,
       4,     5,     6,     7,     8,    -1,    53,    -1,    55,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    64,    65,    -1,
      -1,    -1,    -1,    27,    28,    29,    -1,    31,    -1,    33,
      77,    -1,    -1,    37,    -1,    39,    -1,    41,    -1,    -1,
      -1,    88,    46,    -1,    -1,    -1,    -1,    -1,    -1,    53,
      -1,    55,    56,    57,    58,    59,    60,    61,    62,    63,
      64,    65,    66,    67,    68,    -1,    -1,    71,    -1,    73,
      74,    -1,    -1,    77,    78,    -1,    80,    81,    82,    83,
      84,    85,    86,    87,    88,    89,    -1,    -1,     3,    -1,
      -1,    95,    96,    -1,     3,    -1,   100,    -1,    -1,    -1,
      -1,   105,     3,     4,     5,     6,     7,     8,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    34,
      -1,    -1,    37,    -1,    -1,    40,    27,    28,    37,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    37,    -1,    53,    -1,
      55,    -1,    -1,    -1,    53,    46,    55,    -1,    -1,    64,
      65,    -1,    53,    -1,    55,    64,    65,    -1,    -1,    -1,
      -1,    -1,    77,    64,    65,    -1,    67,    68,    77,    -1,
      -1,    -1,    -1,    88,    -1,     3,    77,    78,    -1,    88,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    88,    89,   104,
      -1,    -1,    -1,    -1,    95,    96,    -1,    -1,    -1,   100,
      -1,    29,    -1,    31,   105,    33,    -1,    35,    -1,    37,
      -1,    39,    -1,    -1,    -1,    -1,    -1,    45,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    53,    -1,    55,    -1,    57,
      58,    59,    60,    61,    62,    63,    64,    65,    66,     3,
      -1,    69,    70,    -1,    -1,    -1,    74,    -1,    -1,    77,
      -1,    79,    80,    81,    82,    83,    84,    85,    86,    -1,
      88,    -1,    -1,    -1,    -1,    29,    -1,    31,    -1,    33,
      -1,    35,    -1,    37,    -1,    39,   104,    -1,    -1,    -1,
      -1,    45,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    53,
      -1,    55,    -1,    57,    58,    59,    60,    61,    62,    63,
      64,    65,    66,     3,    -1,    69,    70,    -1,    -1,    -1,
      74,    -1,    -1,    77,    -1,    79,    80,    81,    82,    83,
      84,    85,    86,    -1,    88,    -1,    -1,    -1,    -1,    29,
      -1,    31,    -1,    33,    -1,    -1,    -1,    37,    -1,    39,
     104,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    53,    -1,    55,    -1,    57,    58,    59,
      60,    61,    62,    63,    64,    65,    66,     3,    -1,    -1,
      -1,    -1,    -1,    -1,    74,    -1,    -1,    77,    -1,    -1,
      80,    81,    82,    83,    84,    85,    86,    -1,    88,    -1,
      -1,    -1,    -1,    29,    -1,    31,    -1,    33,    -1,    -1,
      -1,    37,    -1,    39,   104,    -1,    -1,    -1,     3,     4,
       5,     6,     7,     8,    -1,    -1,    -1,    53,    -1,    55,
      -1,    57,    58,    59,    60,    61,    62,    63,    64,    65,
      66,    -1,    27,    28,    -1,    -1,    -1,    -1,    74,    -1,
      -1,    77,    37,    -1,    80,    81,    82,    83,    84,    85,
      86,    46,    88,    -1,    -1,    -1,    -1,    -1,    53,    -1,
      55,     3,     4,     5,     6,     7,     8,   103,    -1,    64,
      65,    -1,    67,    68,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    77,    78,    -1,    27,    28,    -1,    -1,    -1,
      -1,    -1,    -1,    88,    89,    37,    -1,    -1,    -1,    41,
      95,    96,    -1,    -1,    46,   100,    -1,    -1,   103,    -1,
      -1,    53,    -1,    55,    56,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    64,    65,    -1,    67,    68,    -1,    -1,    71,
      -1,    73,    -1,    -1,    -1,    77,    78,     3,     4,     5,
       6,     7,     8,    -1,    -1,    87,    88,    89,    -1,    -1,
      -1,    -1,    -1,    95,    96,    -1,    -1,    -1,   100,   101,
      -1,    27,    28,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    37,    -1,    -1,    -1,    41,    -1,    -1,    -1,    -1,
      46,    -1,    -1,    -1,    -1,    -1,    -1,    53,    -1,    55,
      56,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    64,    65,
      -1,    67,    68,    -1,    -1,    71,    -1,    73,    -1,    -1,
      -1,    77,    78,     3,     4,     5,     6,     7,     8,    -1,
      -1,    87,    88,    89,    -1,    -1,    -1,    -1,    -1,    95,
      96,    -1,    -1,    -1,   100,   101,    -1,    27,    28,    29,
      -1,    31,    -1,    33,    -1,    -1,    -1,    37,    -1,    39,
      -1,    41,    -1,    -1,    -1,    -1,    46,    -1,    -1,    -1,
      -1,    -1,    -1,    53,    -1,    55,    56,    57,    58,    59,
      60,    61,    62,    63,    64,    65,    -1,    67,    68,    -1,
      -1,    71,    -1,    73,    74,     3,    -1,    77,    78,    -1,
      80,    81,    82,    83,    84,    85,    86,    87,    88,    89,
      -1,    -1,    -1,    -1,    -1,    95,    96,    -1,    -1,    -1,
     100,    29,    -1,    31,    -1,    33,    -1,    -1,    -1,    37,
      -1,    39,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    53,    -1,    55,    -1,    57,
      58,    59,    60,    61,    62,    63,    64,    65,    66,     3,
       4,     5,     6,     7,     8,    -1,    74,    -1,    -1,    77,
      -1,    -1,    80,    81,    82,    83,    84,    85,    86,    -1,
      88,    -1,    -1,    27,    28,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,   100,    37,    -1,    -1,    -1,    41,    -1,    -1,
      -1,    -1,    46,    -1,    -1,    -1,    -1,    -1,    -1,    53,
      -1,    55,    56,    -1,     3,     4,     5,     6,     7,     8,
      64,    65,    -1,    67,    68,    -1,    -1,    71,    -1,    73,
      -1,    -1,    -1,    77,    78,    -1,    -1,    -1,    27,    28,
      -1,    -1,    -1,    87,    88,    89,    -1,    -1,    37,    -1,
      -1,    95,    96,    -1,    -1,    -1,   100,    46,     3,     4,
       5,     6,     7,     8,    53,    -1,    55,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    64,    65,    -1,    67,    68,
      -1,    -1,    27,    28,    -1,    -1,    -1,    -1,    77,    78,
      -1,    -1,    37,    -1,    -1,    -1,    -1,    -1,    -1,    88,
      89,    46,    -1,    -1,    -1,    -1,    95,    96,    53,    -1,
      55,   100,    -1,     3,     4,     5,     6,     7,     8,    64,
      65,    -1,    67,    68,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    77,    78,    -1,    -1,    -1,    27,    28,    -1,
      -1,    -1,    -1,    88,    89,    -1,    -1,    37,    -1,    -1,
      95,    96,    -1,    -1,    -1,   100,    46,     3,     4,     5,
       6,     7,     8,    53,    -1,    55,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    64,    65,    -1,    67,    68,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    77,    78,    -1,
      -1,    37,    -1,    -1,    -1,    -1,    -1,    -1,    88,    89,
      46,    -1,    -1,    -1,    -1,    95,    96,    53,    -1,    55,
     100,     3,    -1,    -1,    -1,    -1,    -1,    -1,    64,    65,
      -1,    67,    68,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    77,    78,    -1,    -1,    -1,    -1,    29,    -1,    31,
      -1,    33,    88,    35,    -1,    37,    -1,    39,    -1,    -1,
       3,    -1,    -1,    -1,   100,    -1,    -1,    -1,    -1,    -1,
      -1,    53,    -1,    55,    -1,    57,    58,    59,    60,    61,
      62,    63,    64,    65,    66,    -1,    29,    69,    31,    -1,
      33,    -1,    74,    -1,    37,    77,    39,    -1,    80,    81,
      82,    83,    84,    85,    86,    -1,    88,    -1,    -1,    -1,
      53,    -1,    55,    -1,    57,    58,    59,    60,    61,    62,
      63,    64,    65,    66,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    74,    -1,    -1,    77,    -1,    -1,    80,    81,    82,
      83,    84,    85,    86,    -1,    88
};

  /* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
     symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    37,    53,    64,    65,   113,   114,   115,   116,   149,
     151,     3,    37,    53,    55,    64,    65,    77,    88,   202,
     203,     8,   203,   203,     0,   115,   116,   149,    52,   117,
     103,   103,   203,   103,    69,    70,   142,   143,   150,   142,
     152,    29,    31,    33,    35,    39,    45,    57,    58,    59,
      60,    61,    62,    63,    66,    70,    74,    79,    80,    81,
      82,    83,    84,    85,    86,   104,   118,   119,   120,   121,
     122,   123,   124,   129,   130,   132,   136,   137,   138,   142,
     203,    69,   105,    49,   104,   142,   103,   110,   147,   104,
     142,   121,   123,   203,     1,   203,   100,   120,     1,   203,
     104,   119,   129,   136,   105,   106,     1,   125,   126,   203,
     147,   100,   203,   105,     1,     4,     5,     6,     7,     8,
      27,    28,    30,    32,    34,    36,    38,    40,    41,    43,
      46,    48,    50,    51,    56,    67,    68,    71,    72,    73,
      76,    78,    87,    89,    95,    96,   100,   104,   105,   120,
     129,   147,   148,   153,   154,   155,   156,   158,   159,   160,
     163,   166,   167,   168,   169,   170,   171,   172,   173,   178,
     179,   180,   181,   182,   183,   184,   186,   188,   190,   192,
     194,   195,   196,   197,   200,   201,   203,     8,   110,   147,
     104,   103,   123,   104,    75,   128,   171,   195,   203,   105,
     107,   108,   123,   139,   140,   141,   100,   105,    77,   105,
      67,   184,   105,    55,    77,    88,   110,    55,    77,    88,
       1,   100,   147,   203,     1,   100,   103,   123,   174,   175,
     203,   105,   170,     1,   100,   103,   164,   100,   195,   124,
     170,   104,   120,   129,   153,    44,     1,   105,   106,   107,
       1,    22,   111,    21,    90,    91,    92,    25,    26,   185,
      23,    24,    93,    94,   187,    95,    96,   189,    19,    20,
     191,    97,    98,    99,   193,     9,    10,    11,    12,    13,
      14,    15,    16,    17,    18,   157,   195,    27,    28,   100,
     102,   108,   110,   105,     8,   133,   134,   203,   107,   103,
     105,   107,   126,   127,   188,   126,   101,   107,   139,   203,
     110,   104,   121,   122,   154,   156,   161,   170,   178,   203,
     105,   104,   170,   175,   176,   177,   178,   203,   108,   100,
     105,   104,   170,   104,   158,   165,   203,   101,   101,    51,
     147,   105,   170,   171,   105,   180,   170,   181,   182,   183,
     184,   186,   188,   190,   192,   194,   170,   173,   198,   199,
     203,   178,   153,   203,   105,   107,   135,   106,   123,   122,
     131,   171,   109,   141,   101,   105,   162,   178,    54,   162,
      54,   101,   101,   107,   135,   110,   188,   198,   101,   104,
     153,   195,   100,   105,   110,   105,   101,   107,   109,   134,
     104,   179,   101,   105,   104,   122,   100,   123,   144,   145,
     146,   101,   170,   105,   197,   101,   170,   197,   147,   147,
     177,   104,   177,   109,   101,   164,   170,   178,   173,   105,
     145,   107,   108,   147,   101,   101,   147,   101,   101,   101,
     101,   146,   127,   147,   147,   147,   147,   147,   109
};

  /* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint8 yyr1[] =
{
       0,   112,   113,   113,   114,   114,   114,   114,   114,   114,
     115,   116,   116,   117,   117,   118,   118,   118,   118,   118,
     118,   119,   119,   119,   120,   120,   121,   121,   122,   123,
     123,   123,   124,   124,   124,   124,   124,   124,   124,   124,
     124,   124,   124,   124,   124,   124,   124,   124,   124,   124,
     124,   125,   125,   126,   126,   127,   127,   128,   128,   129,
     129,   130,   130,   131,   131,   132,   132,   133,   133,   134,
     134,   135,   135,   136,   137,   137,   138,   139,   139,   140,
     140,   141,   142,   143,   143,   143,   143,   144,   144,   144,
     145,   145,   146,   146,   147,   147,   148,   148,   148,   148,
     148,   148,   149,   150,   150,   151,   152,   152,   152,   152,
     153,   153,   153,   153,   153,   153,   153,   153,   153,   153,
     154,   155,   155,   156,   156,   157,   157,   157,   157,   157,
     157,   157,   157,   157,   157,   158,   158,   158,   159,   159,
     159,   159,   160,   160,   160,   160,   160,   160,   160,   160,
     160,   161,   161,   162,   162,   163,   163,   163,   164,   164,
     165,   165,   166,   166,   166,   166,   166,   167,   168,   168,
     168,   168,   168,   168,   168,   169,   170,   170,   171,   171,
     172,   172,   172,   172,   172,   173,   173,   173,   174,   174,
     175,   176,   176,   177,   177,   177,   178,   178,   179,   179,
     180,   180,   181,   181,   182,   182,   183,   183,   184,   184,
     185,   185,   186,   186,   187,   187,   187,   187,   188,   188,
     189,   189,   190,   190,   191,   191,   192,   192,   193,   193,
     193,   194,   194,   195,   195,   195,   196,   196,   196,   196,
     197,   197,   197,   197,   197,   197,   198,   198,   199,   199,
     200,   200,   200,   200,   201,   201,   201,   201,   201,   201,
     201,   201,   202,   202,   202,   202,   202,   202,   202,   203,
     203
};

  /* YYR2[YYN] -- Number of symbols on the right hand side of rule YYN.  */
static const yytype_uint8 yyr2[] =
{
       0,     2,     1,     1,     1,     1,     1,     2,     2,     2,
       2,     5,     6,     0,     2,     1,     1,     1,     2,     2,
       2,     1,     2,     3,     1,     2,     2,     4,     2,     1,
       1,     6,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     3,     1,     4,     0,     1,     1,     3,     1,
       1,     6,     3,     2,     3,     6,     3,     1,     3,     1,
       3,     0,     1,     2,     1,     1,     4,     0,     1,     1,
       3,     2,     7,     0,     1,     1,     2,     0,     3,     1,
       1,     3,     1,     4,     2,     3,     1,     1,     1,     2,
       2,     2,     5,     2,     3,     5,     4,     2,     5,     3,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     2,     2,     4,     4,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     3,     3,     2,     5,     7,
       3,     3,     2,     5,     6,     7,     6,     7,     7,     7,
       3,     1,     1,     1,     2,     2,     5,     3,     2,     3,
       1,     2,     2,     2,     2,     3,     3,     3,     2,     2,
       2,     2,     2,     2,     2,     1,     1,     3,     1,     3,
       1,     1,     1,     1,     1,     1,     2,     2,     1,     4,
       4,     1,     3,     1,     1,     3,     1,     5,     1,     3,
       1,     3,     1,     3,     1,     3,     1,     3,     1,     3,
       1,     1,     1,     3,     1,     1,     1,     1,     1,     3,
       1,     1,     1,     3,     1,     1,     1,     3,     1,     1,
       1,     1,     4,     1,     2,     2,     1,     1,     1,     1,
       1,     4,     4,     3,     2,     2,     0,     1,     1,     3,
       1,     1,     3,     5,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       1
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

#line 1933 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1430  */
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
#line 279 "grammar.y" /* yacc.c:1648  */
    {
        ast_blk_t *blk = blk_new_contract(&(yylsp[-1]));

        /* add default constructor */
        id_add(&blk->ids, id_new_ctor((yyvsp[-3].str), NULL, NULL, &(yylsp[-3])));

        id_add(&ROOT->ids, id_new_contract((yyvsp[-3].str), (yyvsp[-2].exp), blk, &(yyloc)));
    }
#line 2130 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 12:
#line 288 "grammar.y" /* yacc.c:1648  */
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
#line 2150 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 13:
#line 306 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 2156 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 14:
#line 308 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yylsp[0]));
    }
#line 2164 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 15:
#line 315 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2173 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 16:
#line 320 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2182 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 17:
#line 325 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2191 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 18:
#line 330 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2200 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 19:
#line 335 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2209 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 20:
#line 340 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2218 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 22:
#line 349 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_PUBLIC;
    }
#line 2227 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 23:
#line 354 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2236 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 25:
#line 363 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_CONST;
    }
#line 2245 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 27:
#line 372 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.dflt_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.dflt_exp = (yyvsp[-1].exp);
    }
#line 2258 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 28:
#line 384 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.type_exp = (yyvsp[-1].exp);
    }
#line 2271 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 29:
#line 396 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type((yyvsp[0].type), &(yylsp[0]));
    }
#line 2279 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 30:
#line 400 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_NONE, &(yylsp[0]));

        (yyval.exp)->u_type.name = (yyvsp[0].str);
    }
#line 2289 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 31:
#line 406 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_MAP, &(yylsp[-5]));

        (yyval.exp)->u_type.k_exp = (yyvsp[-3].exp);
        (yyval.exp)->u_type.v_exp = (yyvsp[-1].exp);
    }
#line 2300 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 32:
#line 415 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_ACCOUNT; }
#line 2306 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 33:
#line 416 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BOOL; }
#line 2312 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 34:
#line 417 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT8; }
#line 2318 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 35:
#line 418 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT8; }
#line 2324 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 36:
#line 419 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT16; }
#line 2330 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 37:
#line 420 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2336 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 38:
#line 421 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT64; }
#line 2342 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 39:
#line 422 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT128; }
#line 2348 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 40:
#line 423 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT256; }
#line 2354 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 41:
#line 424 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2360 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 42:
#line 425 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT8; }
#line 2366 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 43:
#line 426 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT16; }
#line 2372 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 44:
#line 427 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2378 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 45:
#line 428 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT64; }
#line 2384 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 46:
#line 429 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT128; }
#line 2390 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 47:
#line 430 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT256; }
#line 2396 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 48:
#line 431 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2402 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 49:
#line 434 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_STRING; }
#line 2408 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 50:
#line 435 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_CURSOR; }
#line 2414 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 52:
#line 441 "grammar.y" /* yacc.c:1648  */
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
#line 2430 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 53:
#line 456 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PRIVATE, &(yylsp[0]));
    }
#line 2438 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 54:
#line 460 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2451 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 55:
#line 471 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = exp_new_null(&(yyloc)); }
#line 2457 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 57:
#line 477 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 2465 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 58:
#line 481 "grammar.y" /* yacc.c:1648  */
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
#line 2480 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 61:
#line 500 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_struct((yyvsp[-4].str), (yyvsp[-1].vect), &(yyloc));
    }
#line 2488 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 62:
#line 504 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2497 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 63:
#line 512 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2510 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 64:
#line 521 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2523 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 65:
#line 533 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_enum((yyvsp[-4].str), (yyvsp[-2].vect), &(yyloc));
    }
#line 2531 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 66:
#line 537 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2540 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 67:
#line 545 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2549 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 68:
#line 550 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2558 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 69:
#line 558 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PUBLIC | MOD_CONST, &(yylsp[0]));
    }
#line 2566 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 70:
#line 562 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[-2].str), MOD_PUBLIC | MOD_CONST, &(yylsp[-2]));
        (yyval.id)->u_var.dflt_exp = (yyvsp[0].exp);
    }
#line 2575 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 73:
#line 575 "grammar.y" /* yacc.c:1648  */
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
#line 2592 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 76:
#line 596 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_ctor((yyvsp[-3].str), (yyvsp[-1].vect), NULL, &(yylsp[-3]));
    }
#line 2600 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 77:
#line 602 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 2606 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 79:
#line 608 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2615 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 80:
#line 613 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2624 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 81:
#line 621 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->u_var.is_param = true;
        (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
    }
#line 2634 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 82:
#line 630 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_func((yyvsp[-4].str), (yyvsp[-6].mod), (yyvsp[-2].vect), (yyvsp[0].id), NULL, &(yylsp[-4]));
    }
#line 2642 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 83:
#line 636 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PRIVATE; }
#line 2648 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 84:
#line 637 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC; }
#line 2654 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 85:
#line 638 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PAYABLE; }
#line 2660 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 86:
#line 639 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC | MOD_PAYABLE; }
#line 2666 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 87:
#line 643 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = NULL; }
#line 2672 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 88:
#line 644 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = (yyvsp[-1].id); }
#line 2678 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 91:
#line 651 "grammar.y" /* yacc.c:1648  */
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
#line 2708 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 92:
#line 680 "grammar.y" /* yacc.c:1648  */
    {
        /* We wanted to use a type expression, but we can not store size expressions
         * and declare it as an identifier. */
        (yyval.id) = id_new_param(NULL, (yyvsp[0].exp), &(yylsp[0]));
    }
#line 2718 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 93:
#line 686 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2731 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 94:
#line 697 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 2737 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 95:
#line 698 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 2743 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 96:
#line 703 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2752 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 97:
#line 708 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        /* Unlike state variables, local variables are referenced according to their
         * order of declaration. */
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2763 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 98:
#line 715 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2772 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 99:
#line 720 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2781 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 100:
#line 725 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2790 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 101:
#line 730 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2799 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 102:
#line 738 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, id_new_interface((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2807 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 103:
#line 745 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_interface(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2816 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 104:
#line 750 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2825 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 105:
#line 758 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, id_new_library((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3])));
    }
#line 2833 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 106:
#line 765 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-3].id)->mod = MOD_SYSTEM;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2846 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 107:
#line 774 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_library(&(yyloc));

        (yyvsp[-1].id)->mod = MOD_SYSTEM;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2859 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 108:
#line 783 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-4].blk);

        (yyvsp[-3].id)->mod = MOD_SYSTEM;
        (yyvsp[-3].id)->u_fn.alias = (yyvsp[-1].str);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-3].id));
    }
#line 2872 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 109:
#line 792 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);

        (yyvsp[-1].id)->mod = MOD_SYSTEM;
        (yyvsp[-1].id)->u_fn.blk = (yyvsp[0].blk);

        vector_add_last(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2885 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 120:
#line 817 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_null(&(yyloc));
    }
#line 2893 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 121:
#line 824 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_exp((yyvsp[-1].exp), &(yyloc));
    }
#line 2901 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 122:
#line 828 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2910 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 123:
#line 836 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2]));
    }
#line 2918 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 124:
#line 840 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), exp_new_binary((yyvsp[-2].op), (yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2])), &(yylsp[-2]));
    }
#line 2926 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 125:
#line 846 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 2932 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 126:
#line 847 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 2938 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 127:
#line 848 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 2944 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 128:
#line 849 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 2950 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 129:
#line 850 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 2956 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 130:
#line 851 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_AND; }
#line 2962 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 131:
#line 852 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_XOR; }
#line 2968 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 132:
#line 853 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_OR; }
#line 2974 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 133:
#line 854 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 2980 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 134:
#line 855 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 2986 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 135:
#line 860 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[0].stmt);
        id_add(LABELS, id_new_label((yyvsp[-2].str), (yyvsp[0].stmt), &(yylsp[-2])));
    }
#line 2995 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 136:
#line 865 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case((yyvsp[-1].exp), &(yyloc));
    }
#line 3003 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 137:
#line 869 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case(NULL, &(yyloc));
    }
#line 3011 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 138:
#line 876 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3019 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 139:
#line 880 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-6].stmt);
        stmt_add(&(yyval.stmt)->u_if.elif_stmts, stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yylsp[-5])));
    }
#line 3028 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 140:
#line 885 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-2].stmt);
        (yyval.stmt)->u_if.else_blk = (yyvsp[0].blk);
    }
#line 3037 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 141:
#line 890 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3046 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 142:
#line 898 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, NULL, NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3054 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 143:
#line 902 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3062 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 144:
#line 906 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-3].stmt), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3070 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 145:
#line 910 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-4].stmt), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3078 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 146:
#line 914 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-3].id), &(yylsp[-3])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3086 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 147:
#line 918 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3094 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 148:
#line 922 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_exp((yyvsp[-4].exp), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3102 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 149:
#line 926 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3110 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 150:
#line 930 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3119 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 153:
#line 942 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 3125 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 155:
#line 948 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch(NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3133 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 156:
#line 952 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3141 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 157:
#line 956 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3150 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 158:
#line 963 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 3156 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 159:
#line 964 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 3162 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 160:
#line 969 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_switch(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3171 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 161:
#line 974 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3180 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 162:
#line 982 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_CONTINUE, NULL, &(yyloc));
    }
#line 3188 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 163:
#line 986 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_BREAK, NULL, &(yyloc));
    }
#line 3196 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 164:
#line 990 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return(NULL, &(yyloc));
    }
#line 3204 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 165:
#line 994 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return((yyvsp[-1].exp), &(yyloc));
    }
#line 3212 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 166:
#line 998 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_goto((yyvsp[-1].str), &(yylsp[-1]));
    }
#line 3220 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 167:
#line 1005 "grammar.y" /* yacc.c:1648  */
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
#line 3240 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 175:
#line 1034 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_blk((yyvsp[0].blk), &(yyloc));
    }
#line 3248 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 177:
#line 1042 "grammar.y" /* yacc.c:1648  */
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
#line 3263 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 179:
#line 1057 "grammar.y" /* yacc.c:1648  */
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
#line 3283 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 180:
#line 1075 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_DELETE; }
#line 3289 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 181:
#line 1076 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_INSERT; }
#line 3295 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 182:
#line 1077 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_REPLACE; }
#line 3301 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 183:
#line 1078 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_QUERY; }
#line 3307 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 184:
#line 1079 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_UPDATE; }
#line 3313 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 186:
#line 1085 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3321 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 187:
#line 1089 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
        (yyval.exp)->u_init.is_outmost = true;
    }
#line 3330 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 188:
#line 1097 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_alloc((yyvsp[0].exp), &(yylsp[0]));
    }
#line 3338 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 189:
#line 1101 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-3].exp);

        if ((yyval.exp)->u_alloc.size_exps == NULL)
            (yyval.exp)->u_alloc.size_exps = vector_new();

        exp_add((yyval.exp)->u_alloc.size_exps, (yyvsp[-1].exp));
    }
#line 3351 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 190:
#line 1113 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_init((yyvsp[-2].vect), &(yyloc));
    }
#line 3359 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 191:
#line 1120 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3368 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 192:
#line 1125 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3377 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 195:
#line 1135 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3385 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 197:
#line 1143 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_ternary((yyvsp[-4].exp), (yyvsp[-2].exp), (yyvsp[0].exp), &(yyloc));
    }
#line 3393 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 199:
#line 1151 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3401 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 201:
#line 1159 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3409 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 203:
#line 1167 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3417 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 205:
#line 1175 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_XOR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3425 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 207:
#line 1183 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3433 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 209:
#line 1191 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3441 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 210:
#line 1197 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_EQ; }
#line 3447 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 211:
#line 1198 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NE; }
#line 3453 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 213:
#line 1204 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3461 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 214:
#line 1210 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LT; }
#line 3467 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 215:
#line 1211 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GT; }
#line 3473 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 216:
#line 1212 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LE; }
#line 3479 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 217:
#line 1213 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GE; }
#line 3485 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 219:
#line 1219 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3493 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 220:
#line 1225 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 3499 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 221:
#line 1226 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 3505 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 223:
#line 1232 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3513 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 224:
#line 1238 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 3519 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 225:
#line 1239 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 3525 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 227:
#line 1245 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3533 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 228:
#line 1251 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 3539 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 229:
#line 1252 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 3545 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 230:
#line 1253 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 3551 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 232:
#line 1259 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_cast((yyvsp[-2].type), (yyvsp[0].exp), &(yylsp[-2]));
    }
#line 3559 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 234:
#line 1267 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary((yyvsp[-1].op), true, (yyvsp[0].exp), &(yyloc));
    }
#line 3567 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 235:
#line 1271 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3575 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 236:
#line 1277 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_INC; }
#line 3581 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 237:
#line 1278 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DEC; }
#line 3587 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 238:
#line 1279 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NEG; }
#line 3593 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 239:
#line 1280 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NOT; }
#line 3599 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 241:
#line 1286 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_array((yyvsp[-3].exp), (yyvsp[-1].exp), &(yyloc));
    }
#line 3607 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 242:
#line 1290 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(false, (yyvsp[-3].exp), (yyvsp[-1].vect), &(yyloc));
    }
#line 3615 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 243:
#line 1294 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_access((yyvsp[-2].exp), exp_new_id((yyvsp[0].str), &(yylsp[0])), &(yyloc));
    }
#line 3623 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 244:
#line 1298 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_INC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3631 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 245:
#line 1302 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_DEC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3639 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 246:
#line 1308 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 3645 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 248:
#line 1314 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3654 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 249:
#line 1319 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3663 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 251:
#line 1328 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yyloc));
    }
#line 3671 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 252:
#line 1332 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-1].exp);
    }
#line 3679 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 253:
#line 1336 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(true, exp_new_id((yyvsp[-3].str), &(yylsp[-3])), (yyvsp[-1].vect), &(yylsp[-3]));
    }
#line 3687 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 254:
#line 1343 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_null(&(yyloc));
    }
#line 3695 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 255:
#line 1347 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(true, &(yyloc));
    }
#line 3703 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 256:
#line 1351 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(false, &(yyloc));
    }
#line 3711 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 257:
#line 1355 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNu64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3722 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 258:
#line 1362 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNo64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3733 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 259:
#line 1369 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNx64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3744 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 260:
#line 1376 "grammar.y" /* yacc.c:1648  */
    {
        double v;

        sscanf((yyvsp[0].str), "%lf", &v);
        (yyval.exp) = exp_new_lit_f64(v, &(yyloc));
    }
#line 3755 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 261:
#line 1383 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_str((yyvsp[0].str), &(yyloc));
    }
#line 3763 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 262:
#line 1389 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("contract"); }
#line 3769 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 263:
#line 1390 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("import"); }
#line 3775 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 264:
#line 1391 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("index"); }
#line 3781 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 265:
#line 1392 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("interface"); }
#line 3787 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 266:
#line 1393 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("library"); }
#line 3793 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 267:
#line 1394 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("table"); }
#line 3799 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 268:
#line 1395 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("view"); }
#line 3805 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 269:
#line 1400 "grammar.y" /* yacc.c:1648  */
    {
        if (strlen((yyvsp[0].str)) > NAME_MAX_LEN)
            ERROR(ERROR_TOO_LONG_ID, &(yylsp[0]), NAME_MAX_LEN, strlen((yyvsp[0].str)));

        (yyval.str) = (yyvsp[0].str);
    }
#line 3816 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;


#line 3820 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
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
#line 1409 "grammar.y" /* yacc.c:1907  */


static void
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
