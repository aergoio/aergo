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
    K_BOOL = 285,
    K_BREAK = 286,
    K_BYTE = 287,
    K_CASE = 288,
    K_CONST = 289,
    K_CONTINUE = 290,
    K_CONTRACT = 291,
    K_CREATE = 292,
    K_DEFAULT = 293,
    K_DELETE = 294,
    K_DOUBLE = 295,
    K_DROP = 296,
    K_ELSE = 297,
    K_ENUM = 298,
    K_FALSE = 299,
    K_FLOAT = 300,
    K_FOR = 301,
    K_FUNC = 302,
    K_GOTO = 303,
    K_IF = 304,
    K_IMPLEMENTS = 305,
    K_IMPORT = 306,
    K_IN = 307,
    K_INDEX = 308,
    K_INSERT = 309,
    K_INT = 310,
    K_INT16 = 311,
    K_INT32 = 312,
    K_INT64 = 313,
    K_INT8 = 314,
    K_INTERFACE = 315,
    K_MAP = 316,
    K_NEW = 317,
    K_NULL = 318,
    K_PAYABLE = 319,
    K_PUBLIC = 320,
    K_RETURN = 321,
    K_SELECT = 322,
    K_STRING = 323,
    K_STRUCT = 324,
    K_SWITCH = 325,
    K_TABLE = 326,
    K_TRUE = 327,
    K_TYPE = 328,
    K_UINT = 329,
    K_UINT16 = 330,
    K_UINT32 = 331,
    K_UINT64 = 332,
    K_UINT8 = 333,
    K_UPDATE = 334
  };
#endif

/* Value type.  */
#if ! defined YYSTYPE && ! defined YYSTYPE_IS_DECLARED

union YYSTYPE
{
#line 145 "grammar.y" /* yacc.c:355  */

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

#line 233 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:355  */
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

#line 263 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:358  */

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
#define YYFINAL  18
/* YYLAST -- Last index in YYTABLE.  */
#define YYLAST   1736

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  103
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  89
/* YYNRULES -- Number of rules.  */
#define YYNRULES  252
/* YYNSTATES -- Number of states.  */
#define YYNSTATES  420

/* YYTRANSLATE[YYX] -- Symbol number corresponding to YYX as returned
   by yylex, with out-of-bounds checking.  */
#define YYUNDEFTOK  2
#define YYMAXUTOK   334

#define YYTRANSLATE(YYX)                                                \
  ((unsigned int) (YYX) <= YYMAXUTOK ? yytranslate[YYX] : YYUNDEFTOK)

/* YYTRANSLATE[TOKEN-NUM] -- Symbol number corresponding to TOKEN-NUM
   as returned by yylex, without out-of-bounds checking.  */
static const yytype_uint8 yytranslate[] =
{
       0,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    80,     2,     2,     2,    90,    83,     2,
      91,    92,    88,    86,    98,    87,    93,    89,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,   101,    96,
      84,    97,    85,   102,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,    99,     2,   100,    82,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    94,    81,    95,     2,     2,     2,     2,
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
      75,    76,    77,    78,    79
};

#if YYDEBUG
  /* YYRLINE[YYN] -- Source line where rule number YYN was defined.  */
static const yytype_uint16 yyrline[] =
{
       0,   254,   254,   259,   264,   269,   273,   277,   284,   291,
     300,   319,   320,   327,   332,   337,   342,   347,   352,   360,
     361,   366,   374,   375,   383,   384,   396,   408,   412,   418,
     428,   429,   430,   433,   434,   435,   436,   437,   438,   439,
     440,   441,   442,   443,   447,   448,   463,   467,   479,   480,
     484,   488,   502,   503,   507,   511,   519,   528,   540,   544,
     552,   557,   565,   569,   576,   578,   582,   598,   599,   603,
     610,   611,   615,   620,   628,   637,   638,   642,   647,   654,
     659,   664,   669,   677,   684,   685,   686,   687,   691,   692,
     693,   697,   698,   727,   733,   745,   752,   757,   765,   766,
     767,   768,   769,   770,   771,   772,   773,   774,   778,   785,
     789,   797,   801,   808,   809,   810,   811,   812,   813,   814,
     815,   816,   817,   821,   826,   830,   837,   841,   846,   851,
     859,   863,   867,   871,   875,   879,   883,   887,   891,   899,
     900,   904,   905,   909,   913,   917,   925,   926,   930,   935,
     943,   947,   951,   955,   959,   966,   985,   986,   987,   988,
     992,   999,  1000,  1014,  1015,  1034,  1035,  1036,  1037,  1041,
    1042,  1046,  1053,  1057,  1069,  1076,  1081,  1089,  1090,  1091,
    1098,  1099,  1106,  1107,  1114,  1115,  1122,  1123,  1130,  1131,
    1138,  1139,  1146,  1147,  1154,  1155,  1159,  1160,  1167,  1168,
    1169,  1170,  1174,  1175,  1182,  1183,  1187,  1188,  1195,  1196,
    1200,  1201,  1208,  1209,  1210,  1214,  1215,  1222,  1223,  1227,
    1234,  1235,  1236,  1237,  1241,  1242,  1246,  1250,  1254,  1258,
    1265,  1266,  1270,  1275,  1283,  1284,  1288,  1292,  1299,  1303,
    1307,  1311,  1318,  1325,  1332,  1339,  1346,  1347,  1348,  1349,
    1350,  1354,  1361
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
  "\"--\"", "\"account\"", "\"bool\"", "\"break\"", "\"byte\"", "\"case\"",
  "\"const\"", "\"continue\"", "\"contract\"", "\"create\"", "\"default\"",
  "\"delete\"", "\"double\"", "\"drop\"", "\"else\"", "\"enum\"",
  "\"false\"", "\"float\"", "\"for\"", "\"func\"", "\"goto\"", "\"if\"",
  "\"implements\"", "\"import\"", "\"in\"", "\"index\"", "\"insert\"",
  "\"int\"", "\"int16\"", "\"int32\"", "\"int64\"", "\"int8\"",
  "\"interface\"", "\"map\"", "\"new\"", "\"null\"", "\"payable\"",
  "\"public\"", "\"return\"", "\"select\"", "\"string\"", "\"struct\"",
  "\"switch\"", "\"table\"", "\"true\"", "\"type\"", "\"uint\"",
  "\"uint16\"", "\"uint32\"", "\"uint64\"", "\"uint8\"", "\"update\"",
  "'!'", "'|'", "'^'", "'&'", "'<'", "'>'", "'+'", "'-'", "'*'", "'/'",
  "'%'", "'('", "')'", "'.'", "'{'", "'}'", "';'", "'='", "','", "'['",
  "']'", "':'", "'?'", "$accept", "root", "import", "contract", "impl_opt",
  "contract_body", "variable", "var_qual", "var_decl", "var_spec",
  "var_type", "prim_type", "declarator_list", "declarator", "size_opt",
  "var_init", "compound", "struct", "field_list", "enumeration",
  "enum_list", "enumerator", "comma_opt", "function", "func_spec",
  "ctor_spec", "param_list_opt", "param_list", "param_decl", "block",
  "blk_decl", "udf_spec", "modifier_opt", "return_opt", "return_list",
  "return_decl", "interface", "interface_body", "statement", "empty_stmt",
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
      33,   124,    94,    38,    60,    62,    43,    45,    42,    47,
      37,    40,    41,    46,   123,   125,    59,    61,    44,    91,
      93,    58,    63
};
# endif

#define YYPACT_NINF -259

#define yypact_value_is_default(Yystate) \
  (!!((Yystate) == (-259)))

#define YYTABLE_NINF -29

#define yytable_value_is_error(Yytable_value) \
  0

  /* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
     STATE-NUM.  */
static const yytype_int16 yypact[] =
{
     105,   255,    42,   255,    62,  -259,  -259,  -259,  -259,  -259,
    -259,  -259,  -259,  -259,  -259,    -4,  -259,   -39,  -259,  -259,
    -259,  -259,   255,   -19,   -36,  -259,   793,  -259,    18,     8,
      61,   -11,  -259,  -259,  -259,  1658,   102,  -259,  -259,  -259,
    -259,  -259,    23,  1607,  -259,   983,  -259,  -259,  -259,  -259,
    -259,  -259,   844,  -259,  -259,  -259,    63,   991,  -259,  -259,
    -259,  -259,  -259,    29,  -259,  -259,    37,  -259,  -259,   255,
    -259,    41,  -259,   255,  -259,    45,    39,  1658,  -259,    49,
      85,  -259,  -259,  -259,  -259,  -259,  1295,    53,    74,    82,
    -259,   359,  -259,  1658,    99,  -259,  -259,   255,   101,  -259,
     133,  -259,  -259,  -259,  -259,  -259,  -259,  -259,  -259,  -259,
    -259,   946,  -259,  -259,  -259,  -259,  -259,  1399,  -259,  1206,
     -31,  -259,   213,  -259,  -259,   -13,   214,   161,   163,   167,
     178,    17,   131,   228,    80,  -259,  -259,  1399,    33,  -259,
    -259,  -259,  -259,   255,  1460,   155,   156,  1460,   158,   -26,
     162,   -23,    15,   255,    16,   715,    21,  -259,  -259,  -259,
    -259,  -259,   455,  -259,  -259,  -259,  -259,  -259,   215,  -259,
    -259,  -259,  -259,   264,  -259,   124,  -259,   456,    28,   255,
     181,   185,  -259,  1658,   191,  -259,   179,  1658,  1658,  1005,
    -259,   193,  -259,   199,   255,  1295,  -259,   202,   -49,  -259,
    1295,   201,  1460,  1295,  1460,  1460,  1460,  1460,  -259,  -259,
    1460,  -259,  -259,  -259,  -259,  1460,  -259,  -259,  1460,  -259,
    -259,  1460,  -259,  -259,  -259,  1460,  -259,  -259,  -259,  1521,
     255,  1460,    82,   198,   131,  -259,  -259,    -7,  -259,  -259,
    -259,  -259,  -259,  -259,   206,   621,  -259,   211,   207,  1295,
    -259,   111,   219,  1295,   172,  -259,  -259,  -259,  -259,  -259,
     -25,   220,  -259,  1295,  1295,  -259,  -259,  -259,  -259,  -259,
    -259,  -259,  -259,  -259,  -259,  1295,   551,    82,  -259,  1658,
     217,   255,   222,  1460,   226,   227,   895,  -259,   224,  -259,
    -259,   229,  1460,  1521,   199,  1399,  -259,  -259,  -259,   214,
     -62,   161,   163,   167,   178,    17,   131,   228,    80,  -259,
    -259,   240,   236,  -259,   237,  -259,  -259,  -259,   728,    98,
    -259,  -259,   728,    19,   247,   285,  -259,  -259,    38,  -259,
    -259,    66,  -259,  -259,   233,   239,   250,  -259,  -259,   128,
    -259,   132,  -259,   239,  -259,  1340,  -259,  -259,   320,  -259,
    -259,  -259,   252,  1005,   248,  1005,     0,   261,  -259,  1460,
    -259,  1521,  -259,  -259,  1066,   259,  1582,  1136,  1582,    29,
      29,   265,  -259,  -259,  1295,  -259,  -259,  1658,  -259,  -259,
     260,   262,  -259,  -259,  -259,  -259,  -259,  -259,  -259,  -259,
      29,    79,  -259,    92,    29,   100,   120,  -259,  -259,  -259,
     104,   108,  1658,  1460,  -259,    29,    29,  -259,    29,    29,
      29,  -259,   262,   268,  -259,  -259,  -259,  -259,  -259,  -259
};

  /* YYDEFACT[STATE-NUM] -- Default reduction number in state STATE-NUM.
     Performed when YYTABLE does not specify something else to do.  Zero
     means the default is an error.  */
static const yytype_uint8 yydefact[] =
{
       0,     0,     0,     0,     0,     2,     3,     4,   251,   246,
     247,   248,   249,   250,   252,    11,     8,     0,     1,     5,
       6,     7,     0,     0,    84,    12,    84,    86,    85,     0,
       0,    84,    30,    31,    32,     0,     0,    33,    34,    35,
      36,    37,     0,    85,    38,     0,    39,    40,    41,    42,
      43,     9,    84,    13,    19,    22,     0,     0,    27,    14,
      52,    53,    15,     0,    67,    68,    28,    87,    96,     0,
      95,     0,    23,     0,    28,     0,     0,     0,    20,     0,
       0,    10,    16,    17,    18,    24,     0,     0,    26,    44,
      46,     0,    66,    70,     0,    97,    59,     0,     0,    55,
       0,   244,   243,   241,   242,   245,   220,   221,   165,   240,
     166,     0,   238,   167,   239,   168,   223,     0,   222,     0,
       0,    50,     0,   163,   169,   180,   182,   184,   186,   188,
     190,   192,   196,   202,   206,   210,   215,     0,   217,   224,
     234,   235,    21,     0,    48,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,    75,   108,    77,
      78,   160,     0,    79,    98,    99,   100,   101,   102,   103,
     104,   105,   106,     0,   107,     0,   161,   215,   235,     0,
       0,    71,    72,    70,    64,    60,    62,     0,     0,     0,
     172,   170,   171,    28,     0,     0,   219,     0,     0,    25,
       0,     0,     0,     0,     0,     0,     0,     0,   194,   195,
       0,   200,   201,   198,   199,     0,   204,   205,     0,   208,
     209,     0,   212,   213,   214,     0,   218,   228,   229,   230,
       0,     0,    45,     0,    49,   110,   151,     0,   150,   156,
     157,   125,   158,   159,     0,     0,   130,     0,     0,     0,
     152,     0,     0,     0,     0,   143,    76,    80,    81,    82,
       0,     0,   109,     0,     0,   113,   114,   115,   116,   117,
     118,   119,   120,   121,   122,     0,     0,    74,    69,     0,
       0,    65,     0,     0,     0,     0,     0,   178,    64,   175,
     177,   235,     0,   230,     0,     0,   236,    51,   164,   183,
       0,   185,   187,   189,   191,   193,   197,   203,   207,   211,
     232,     0,   231,   227,     0,    47,   124,   138,     0,     0,
     139,   140,     0,     0,   169,   235,   154,   129,     0,   153,
     145,     0,   146,   148,     0,     0,     0,   128,   155,     0,
     162,     0,   123,   235,    73,    88,    61,    58,    63,    29,
      56,    54,     0,    65,     0,     0,     0,     0,   216,     0,
     226,     0,   225,   141,     0,     0,     0,     0,     0,     0,
       0,     0,   147,   149,     0,   111,   112,     0,    93,    83,
      90,    91,    57,   176,   174,   179,   173,   237,   181,   233,
       0,     0,   142,     0,     0,     0,     0,   131,   126,   144,
       0,     0,     0,    48,   134,     0,     0,   132,     0,     0,
       0,    89,    92,     0,   135,   137,   133,   136,   127,    94
};

  /* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -259,  -259,   365,   366,  -259,  -259,   319,   -28,   -30,  -175,
     -20,   253,  -259,  -120,   -29,  -259,   -44,  -259,  -259,  -259,
    -259,    94,    89,   321,  -259,  -259,   195,  -259,   106,   -59,
    -259,     2,  -259,  -259,     3,   -21,   375,  -259,  -155,   138,
    -259,   139,  -259,   145,  -259,  -259,  -259,    84,  -259,    40,
    -259,  -259,  -259,  -259,  -259,  -118,   -75,  -259,  -215,  -259,
     293,  -259,  -258,  -179,   126,   221,   238,   223,   218,  -127,
    -259,   230,  -259,  -141,  -259,   225,  -259,   231,  -259,   216,
     -79,  -259,  -135,   151,  -259,  -259,  -259,  -259,    -1
};

  /* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     4,     5,     6,    23,    52,    53,    54,    55,    56,
      73,    58,    88,    89,   233,   120,    59,    60,   286,    61,
     184,   185,   282,    62,    63,    64,   180,   181,   182,   161,
     162,    65,    30,   379,   380,   381,     7,    31,   163,   164,
     165,   166,   275,   167,   168,   169,   322,   364,   170,   255,
     334,   171,   172,   173,   174,   175,   176,   122,   123,   191,
     287,   288,   289,   124,   125,   126,   127,   128,   129,   130,
     210,   131,   215,   132,   218,   133,   221,   134,   225,   135,
     136,   137,   138,   311,   312,   139,   140,    14,   141
};

  /* YYTABLE[YYPACT[STATE-NUM]] -- What to do in state STATE-NUM.  If
     positive, shift that token.  If negative, reduce the rule whose
     number is the opposite.  If YYTABLE_NINF, syntax error.  */
static const yytype_int16 yytable[] =
{
      15,   198,    17,   234,    92,    72,    57,   259,    83,   202,
     290,   121,   177,   285,   310,    78,   244,   248,   208,   209,
     237,    25,   252,   232,   336,    66,    29,   239,    27,    28,
     242,   -28,    57,    71,    74,    76,   264,   251,   196,   359,
     211,   212,    74,   296,    80,   240,    22,   160,   243,   264,
      16,    66,   314,    27,    28,    24,    90,    98,   226,   277,
     227,   228,    18,   159,   -28,   199,   324,   200,    94,    91,
     319,   368,    90,   179,   306,    26,    74,   198,   310,   -28,
     304,   -28,    67,   177,    70,   300,   216,   217,   -28,   203,
     178,   190,    74,   246,   316,   383,   186,   385,     1,   -28,
     386,   213,   214,    75,    68,     8,   245,   249,    69,    91,
     193,   352,   253,     2,    77,   254,   263,   264,   258,   227,
     228,   342,     3,    91,   229,   297,   230,   323,    93,   276,
     370,   328,   231,    97,   257,   331,   264,    95,     9,   365,
      96,     1,    90,   365,    99,   339,   389,   227,   228,   142,
     366,   356,   247,    10,   100,    11,     2,   341,   371,    85,
      86,   178,    12,   179,   264,     3,   177,   284,   222,   223,
     224,   405,   143,    13,   290,     8,   290,   264,    90,   373,
     388,   144,    74,   229,   406,   230,    74,    74,   291,   340,
     183,   231,   408,   294,    85,    86,   410,   177,   264,   187,
     411,   337,   264,   208,   209,   147,   402,   329,     9,   264,
     150,   229,   409,   230,   201,   318,   358,   216,   217,   231,
     262,   263,   264,    10,   375,    11,   264,   188,   376,   313,
     264,   393,    12,   396,   145,   204,     8,   101,   102,   103,
     104,   105,   205,    13,   325,   206,   391,   219,   220,   395,
     207,   235,   236,   335,   238,   177,   400,   260,     8,   179,
     106,   107,   234,   241,   146,   261,   147,   332,   148,     9,
     149,   150,   108,   278,   151,   343,   283,   109,    74,   152,
     186,   153,   154,   279,    10,    74,    11,   110,   -28,   281,
     293,     9,   292,    12,   295,   111,   112,   298,   315,   155,
     113,   317,   327,   156,    13,   114,    10,   326,    11,   345,
     397,   398,   115,   116,   330,    12,   338,   347,   349,   117,
     118,   -28,   353,   350,   119,   378,    13,    91,   372,   158,
     355,   404,   360,   343,   361,   407,   -28,   362,   -28,   369,
     276,   374,   202,   384,    74,   -28,   414,   415,   382,   416,
     417,   418,   291,   387,   291,   392,   -28,   378,   402,   254,
     145,   403,     8,   101,   102,   103,   104,   105,   419,    19,
      20,    82,   197,    84,   413,   346,    74,   354,   280,    21,
     401,   412,   378,   320,   321,   344,   106,   107,    32,    33,
     146,    34,   147,    35,   148,     9,   149,   150,   108,   333,
     151,    74,    36,   109,   192,   152,   367,   153,   154,   348,
      10,   399,    11,   110,    37,    38,    39,    40,    41,    12,
      42,   111,   112,   299,   303,   155,   113,    44,   302,   156,
      13,   114,    45,    46,    47,    48,    49,    50,   115,   116,
     305,   309,   301,   307,   357,   117,   118,     0,     0,     0,
     119,     0,   308,    91,   157,   158,   145,     0,     8,   101,
     102,   103,   104,   105,     0,   265,   266,   267,   268,   269,
     270,   271,   272,   273,   274,     0,     0,     0,     0,     0,
       0,     0,   106,   107,    32,    33,   146,    34,   147,    35,
     148,     9,   149,   150,   108,     0,   151,     0,    36,   109,
       0,   152,     0,   153,   154,     0,    10,     0,    11,   110,
      37,    38,    39,    40,    41,    12,    42,   111,   112,     0,
       0,   155,   113,    44,     0,   156,    13,   114,    45,    46,
      47,    48,    49,    50,   115,   116,     0,     0,     0,     0,
       0,   117,   118,     0,     0,     0,   119,     0,     0,    91,
     256,   158,   145,     0,     8,   101,   102,   103,   104,   105,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,   106,   107,
       0,     0,   146,     0,   147,     0,   148,     9,   149,   150,
     108,     0,   151,     0,     0,   109,     0,   152,     0,   153,
     154,     0,    10,     0,    11,   110,     0,     0,     0,     0,
       0,    12,     0,   111,   112,     0,     0,   155,   113,     0,
       0,   156,    13,   114,     8,   101,   102,   103,   104,   105,
     115,   116,     0,     0,     0,     0,     0,   117,   118,     0,
       0,     0,   119,     0,     0,    91,     0,   158,   106,   107,
      32,    33,     0,    34,     0,     0,     0,     9,     0,     0,
     108,     0,     0,     0,     0,   109,     0,     0,     0,     0,
       0,     0,    10,     0,    11,   110,    37,    38,    39,    40,
      41,    12,    42,   111,   112,     0,     0,     0,   113,    44,
       0,     0,    13,   114,     0,    46,    47,    48,    49,    50,
     115,   116,     0,     0,     0,     0,     0,   117,   118,     0,
       0,     0,   119,     0,     0,     0,     0,   158,     8,   101,
     102,   103,   104,   105,     0,     0,     0,     0,     0,     0,
       0,     8,   101,   102,   103,   104,   105,     0,     0,     0,
       0,     0,   106,   107,     0,     0,     0,     0,     0,     0,
       0,     9,     0,     0,   108,   106,   107,     0,     0,   109,
       0,     0,     0,     0,     9,     0,    10,     0,    11,   110,
       0,     0,   109,     0,     0,    12,     0,   111,   112,    10,
       0,    11,   113,     0,     0,     0,    13,   114,    12,     0,
     194,   112,     0,     0,   115,   116,     8,     0,     0,    13,
     114,   117,   118,     0,     0,     0,   119,     0,   116,     0,
       0,   250,     0,     0,   117,   118,     0,     0,     0,   119,
       0,     0,    32,    33,   363,    34,     0,    35,     0,     9,
       0,     0,     0,     0,     0,     0,    36,     0,     0,     0,
       0,     0,     0,     0,    10,     0,    11,     8,    37,    38,
      39,    40,    41,    12,    42,     0,     0,    27,    43,     0,
       0,    44,     0,     0,    13,     0,    45,    46,    47,    48,
      49,    50,     0,    32,    33,     0,    34,     0,    35,     0,
       9,     0,     0,     0,     0,     0,     0,    36,    51,     0,
       0,     0,     0,     0,     0,    10,     0,    11,     8,    37,
      38,    39,    40,    41,    12,    42,     0,     0,    27,    43,
       0,     0,    44,     0,     0,    13,     0,    45,    46,    47,
      48,    49,    50,     0,    32,    33,     0,    34,     0,     0,
       0,     9,     0,     0,     0,     0,     0,     0,     0,    81,
       0,     0,     0,     0,     0,     0,    10,     0,    11,     8,
      37,    38,    39,    40,    41,    12,    42,     0,     0,     0,
       0,     0,     0,    44,     0,     0,    13,     0,     0,    46,
      47,    48,    49,    50,     0,    32,    33,     0,    34,     0,
       0,     0,     9,     0,    79,     0,     8,     0,     0,     0,
     351,     0,    87,     0,     8,     0,     0,    10,     0,    11,
       0,    37,    38,    39,    40,    41,    12,    42,     8,   101,
     102,   103,   104,   105,    44,     0,     0,    13,     0,     9,
      46,    47,    48,    49,    50,     0,     0,     9,     0,     0,
       0,     0,   106,   107,    10,     0,    11,     0,     0,     0,
     189,     9,    10,    12,    11,     0,     0,     0,     0,   109,
       0,    12,     0,     0,    13,     0,    10,     0,    11,     0,
       0,     0,    13,     0,     0,    12,     0,   194,   112,     8,
     101,   102,   103,   104,   105,     0,    13,   114,     0,     0,
       0,     0,     0,     0,     0,   116,     0,     0,     0,     0,
       0,   117,   118,   106,   107,     0,   119,     0,     0,   189,
       0,     0,     9,     0,     0,   108,     0,     0,     0,     0,
     109,     0,     0,     0,     0,     0,     0,    10,     0,    11,
     110,     0,     0,     0,     0,     0,    12,     0,   111,   112,
       0,     0,     0,   113,     0,     0,     0,    13,   114,     8,
     101,   102,   103,   104,   105,   115,   116,     0,     0,     0,
       0,     0,   117,   118,     0,     0,     0,   119,   390,     0,
       0,     0,     0,   106,   107,     0,     0,     0,     0,     0,
       0,     0,     9,     0,     0,   108,     0,     0,     0,     0,
     109,     0,     0,     0,     0,     0,     0,    10,     0,    11,
     110,     0,     0,     0,     0,     0,    12,     0,   111,   112,
       0,     0,     0,   113,     0,     0,     0,    13,   114,     8,
     101,   102,   103,   104,   105,   115,   116,     0,     0,     0,
       0,     0,   117,   118,     0,     0,     0,   119,   394,     0,
       0,     0,     0,   106,   107,    32,    33,     0,    34,     0,
       0,     0,     9,     0,     0,   108,     0,     0,     0,     0,
     109,     0,     0,     0,     0,     0,     0,    10,     0,    11,
     110,    37,    38,    39,    40,    41,    12,     0,   111,   112,
       0,     0,     0,   113,    44,     0,     0,    13,   114,     0,
      46,    47,    48,    49,    50,   115,   116,     0,     0,     0,
       0,     0,   117,   118,     0,     0,     0,   119,     8,   101,
     102,   103,   104,   105,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,   106,   107,     0,     0,     0,     0,     0,     0,
       0,     9,     0,     0,   108,     0,     0,     0,     0,   109,
       0,     0,     0,     8,     0,     0,    10,     0,    11,   110,
       0,     0,     0,     0,     0,    12,     0,   111,   112,     0,
       0,     0,   113,     0,     0,     0,    13,   114,     0,    32,
      33,     0,    34,     0,   115,   116,     9,     0,     0,     0,
       0,   117,   118,     0,     0,     0,   119,     0,     0,     0,
       0,    10,     0,    11,     0,    37,    38,    39,    40,    41,
      12,    42,     8,   101,   102,   103,   104,   105,    44,     0,
       0,    13,     0,     0,    46,    47,    48,    49,    50,     0,
       0,     0,     0,     0,     0,     0,   106,   107,     0,     0,
       0,   377,     0,     0,     0,     9,     0,     0,     0,     0,
       0,     0,     0,   109,     0,     0,     0,     0,     0,     0,
      10,     0,    11,     0,     0,     0,     0,     0,     0,    12,
       0,   194,   112,     8,   101,   102,   103,   104,   105,     0,
      13,   114,     0,     0,     0,     0,     0,     0,     0,   116,
       0,     0,     0,     0,     0,   117,   118,   106,   107,     0,
     195,     0,     0,     0,     0,     0,     9,     0,     0,     0,
       0,     0,     0,     0,   109,     0,     0,     0,     0,     0,
       0,    10,     0,    11,     0,     0,     0,     0,     0,     0,
      12,     0,   194,   112,     8,   101,   102,   103,   104,   105,
       0,    13,   114,     0,     0,     0,     0,     0,     0,     0,
     116,     0,     0,     0,     0,     0,   117,   118,   106,   107,
       0,   119,     0,     0,     0,     0,     0,     9,     0,     0,
       0,     0,     0,     0,     0,   109,     0,     0,     0,     0,
       0,     0,    10,     0,    11,     0,     0,     0,     0,     0,
       0,    12,     0,   111,   112,     8,   101,   102,   103,   104,
     105,     0,    13,   114,     0,     0,     0,     0,     0,     0,
       0,   116,     0,     0,     0,     0,     0,   117,   118,     0,
       8,     0,   119,     0,     0,     0,     0,     0,     9,     0,
       0,     0,     0,     0,     0,     0,   109,     0,     0,     0,
       0,     0,     0,    10,     0,    11,    32,    33,     0,    34,
       0,    35,    12,     9,   194,   112,     0,     0,     0,     0,
       0,     0,     0,    13,   114,     0,     0,     0,    10,     0,
      11,     8,    37,    38,    39,    40,    41,    12,    42,     0,
       0,    67,     0,   195,     0,    44,     0,     0,    13,     0,
       0,    46,    47,    48,    49,    50,     0,    32,    33,     0,
      34,     0,     0,     0,     9,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,    10,
       0,    11,     0,    37,    38,    39,    40,    41,    12,    42,
       0,     0,     0,     0,     0,     0,    44,     0,     0,    13,
       0,     0,    46,    47,    48,    49,    50
};

static const yytype_int16 yycheck[] =
{
       1,   119,     3,   144,    63,    35,    26,   162,    52,    22,
     189,    86,    91,   188,   229,    43,     1,     1,    25,    26,
     147,    22,     1,   143,    49,    26,    24,    53,    64,    65,
      53,     3,    52,    31,    35,    36,    98,   155,   117,   101,
      23,    24,    43,    92,    45,    71,    50,    91,    71,    98,
       8,    52,   231,    64,    65,    94,    57,    77,   137,   179,
      27,    28,     0,    91,    36,    96,   245,    98,    69,    94,
     245,    52,    73,    93,   215,    94,    77,   195,   293,    51,
     207,    53,    64,   162,    95,   203,    86,    87,    60,   102,
      91,   111,    93,   152,   101,   353,    97,   355,    36,    71,
     100,    84,    85,     1,    96,     3,    91,    91,    47,    94,
     111,   286,    91,    51,    91,    94,    97,    98,   162,    27,
      28,   276,    60,    94,    91,   200,    93,   245,    91,   101,
      92,   249,    99,    94,   162,   253,    98,    96,    36,   318,
      95,    36,   143,   322,    95,   263,   361,    27,    28,    96,
      52,   292,   153,    51,    69,    53,    51,   275,    92,    96,
      97,   162,    60,   183,    98,    60,   245,   187,    88,    89,
      90,    92,    98,    71,   353,     3,   355,    98,   179,   334,
     359,    99,   183,    91,    92,    93,   187,   188,   189,   264,
      91,    99,    92,   194,    96,    97,    92,   276,    98,    98,
      92,   260,    98,    25,    26,    33,    98,    96,    36,    98,
      38,    91,    92,    93,     1,   245,   295,    86,    87,    99,
      96,    97,    98,    51,    96,    53,    98,    94,    96,   230,
      98,   366,    60,   368,     1,    21,     3,     4,     5,     6,
       7,     8,    81,    71,   245,    82,   364,    19,    20,   367,
      83,    96,    96,   254,    96,   334,   374,    42,     3,   279,
      27,    28,   403,   101,    31,     1,    33,    95,    35,    36,
      37,    38,    39,    92,    41,   276,    97,    44,   279,    46,
     281,    48,    49,    98,    51,   286,    53,    54,     3,    98,
      91,    36,    99,    60,    92,    62,    63,    96,   100,    66,
      67,    95,    95,    70,    71,    72,    51,    96,    53,    92,
     369,   370,    79,    80,    95,    60,    96,    95,    92,    86,
      87,    36,    98,    96,    91,   345,    71,    94,    95,    96,
     101,   390,    92,   334,    98,   394,    51,   100,    53,    92,
     101,    91,    22,    95,   345,    60,   405,   406,    96,   408,
     409,   410,   353,    92,   355,    96,    71,   377,    98,    94,
       1,    99,     3,     4,     5,     6,     7,     8,   100,     4,
       4,    52,   119,    52,   403,   281,   377,   288,   183,     4,
     377,   402,   402,   245,   245,   279,    27,    28,    29,    30,
      31,    32,    33,    34,    35,    36,    37,    38,    39,   254,
      41,   402,    43,    44,   111,    46,   322,    48,    49,   283,
      51,   371,    53,    54,    55,    56,    57,    58,    59,    60,
      61,    62,    63,   202,   206,    66,    67,    68,   205,    70,
      71,    72,    73,    74,    75,    76,    77,    78,    79,    80,
     210,   225,   204,   218,   293,    86,    87,    -1,    -1,    -1,
      91,    -1,   221,    94,    95,    96,     1,    -1,     3,     4,
       5,     6,     7,     8,    -1,     9,    10,    11,    12,    13,
      14,    15,    16,    17,    18,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    27,    28,    29,    30,    31,    32,    33,    34,
      35,    36,    37,    38,    39,    -1,    41,    -1,    43,    44,
      -1,    46,    -1,    48,    49,    -1,    51,    -1,    53,    54,
      55,    56,    57,    58,    59,    60,    61,    62,    63,    -1,
      -1,    66,    67,    68,    -1,    70,    71,    72,    73,    74,
      75,    76,    77,    78,    79,    80,    -1,    -1,    -1,    -1,
      -1,    86,    87,    -1,    -1,    -1,    91,    -1,    -1,    94,
      95,    96,     1,    -1,     3,     4,     5,     6,     7,     8,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    27,    28,
      -1,    -1,    31,    -1,    33,    -1,    35,    36,    37,    38,
      39,    -1,    41,    -1,    -1,    44,    -1,    46,    -1,    48,
      49,    -1,    51,    -1,    53,    54,    -1,    -1,    -1,    -1,
      -1,    60,    -1,    62,    63,    -1,    -1,    66,    67,    -1,
      -1,    70,    71,    72,     3,     4,     5,     6,     7,     8,
      79,    80,    -1,    -1,    -1,    -1,    -1,    86,    87,    -1,
      -1,    -1,    91,    -1,    -1,    94,    -1,    96,    27,    28,
      29,    30,    -1,    32,    -1,    -1,    -1,    36,    -1,    -1,
      39,    -1,    -1,    -1,    -1,    44,    -1,    -1,    -1,    -1,
      -1,    -1,    51,    -1,    53,    54,    55,    56,    57,    58,
      59,    60,    61,    62,    63,    -1,    -1,    -1,    67,    68,
      -1,    -1,    71,    72,    -1,    74,    75,    76,    77,    78,
      79,    80,    -1,    -1,    -1,    -1,    -1,    86,    87,    -1,
      -1,    -1,    91,    -1,    -1,    -1,    -1,    96,     3,     4,
       5,     6,     7,     8,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,     3,     4,     5,     6,     7,     8,    -1,    -1,    -1,
      -1,    -1,    27,    28,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    36,    -1,    -1,    39,    27,    28,    -1,    -1,    44,
      -1,    -1,    -1,    -1,    36,    -1,    51,    -1,    53,    54,
      -1,    -1,    44,    -1,    -1,    60,    -1,    62,    63,    51,
      -1,    53,    67,    -1,    -1,    -1,    71,    72,    60,    -1,
      62,    63,    -1,    -1,    79,    80,     3,    -1,    -1,    71,
      72,    86,    87,    -1,    -1,    -1,    91,    -1,    80,    -1,
      -1,    96,    -1,    -1,    86,    87,    -1,    -1,    -1,    91,
      -1,    -1,    29,    30,    96,    32,    -1,    34,    -1,    36,
      -1,    -1,    -1,    -1,    -1,    -1,    43,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    51,    -1,    53,     3,    55,    56,
      57,    58,    59,    60,    61,    -1,    -1,    64,    65,    -1,
      -1,    68,    -1,    -1,    71,    -1,    73,    74,    75,    76,
      77,    78,    -1,    29,    30,    -1,    32,    -1,    34,    -1,
      36,    -1,    -1,    -1,    -1,    -1,    -1,    43,    95,    -1,
      -1,    -1,    -1,    -1,    -1,    51,    -1,    53,     3,    55,
      56,    57,    58,    59,    60,    61,    -1,    -1,    64,    65,
      -1,    -1,    68,    -1,    -1,    71,    -1,    73,    74,    75,
      76,    77,    78,    -1,    29,    30,    -1,    32,    -1,    -1,
      -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    95,
      -1,    -1,    -1,    -1,    -1,    -1,    51,    -1,    53,     3,
      55,    56,    57,    58,    59,    60,    61,    -1,    -1,    -1,
      -1,    -1,    -1,    68,    -1,    -1,    71,    -1,    -1,    74,
      75,    76,    77,    78,    -1,    29,    30,    -1,    32,    -1,
      -1,    -1,    36,    -1,     1,    -1,     3,    -1,    -1,    -1,
      95,    -1,     1,    -1,     3,    -1,    -1,    51,    -1,    53,
      -1,    55,    56,    57,    58,    59,    60,    61,     3,     4,
       5,     6,     7,     8,    68,    -1,    -1,    71,    -1,    36,
      74,    75,    76,    77,    78,    -1,    -1,    36,    -1,    -1,
      -1,    -1,    27,    28,    51,    -1,    53,    -1,    -1,    -1,
      94,    36,    51,    60,    53,    -1,    -1,    -1,    -1,    44,
      -1,    60,    -1,    -1,    71,    -1,    51,    -1,    53,    -1,
      -1,    -1,    71,    -1,    -1,    60,    -1,    62,    63,     3,
       4,     5,     6,     7,     8,    -1,    71,    72,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    80,    -1,    -1,    -1,    -1,
      -1,    86,    87,    27,    28,    -1,    91,    -1,    -1,    94,
      -1,    -1,    36,    -1,    -1,    39,    -1,    -1,    -1,    -1,
      44,    -1,    -1,    -1,    -1,    -1,    -1,    51,    -1,    53,
      54,    -1,    -1,    -1,    -1,    -1,    60,    -1,    62,    63,
      -1,    -1,    -1,    67,    -1,    -1,    -1,    71,    72,     3,
       4,     5,     6,     7,     8,    79,    80,    -1,    -1,    -1,
      -1,    -1,    86,    87,    -1,    -1,    -1,    91,    92,    -1,
      -1,    -1,    -1,    27,    28,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    36,    -1,    -1,    39,    -1,    -1,    -1,    -1,
      44,    -1,    -1,    -1,    -1,    -1,    -1,    51,    -1,    53,
      54,    -1,    -1,    -1,    -1,    -1,    60,    -1,    62,    63,
      -1,    -1,    -1,    67,    -1,    -1,    -1,    71,    72,     3,
       4,     5,     6,     7,     8,    79,    80,    -1,    -1,    -1,
      -1,    -1,    86,    87,    -1,    -1,    -1,    91,    92,    -1,
      -1,    -1,    -1,    27,    28,    29,    30,    -1,    32,    -1,
      -1,    -1,    36,    -1,    -1,    39,    -1,    -1,    -1,    -1,
      44,    -1,    -1,    -1,    -1,    -1,    -1,    51,    -1,    53,
      54,    55,    56,    57,    58,    59,    60,    -1,    62,    63,
      -1,    -1,    -1,    67,    68,    -1,    -1,    71,    72,    -1,
      74,    75,    76,    77,    78,    79,    80,    -1,    -1,    -1,
      -1,    -1,    86,    87,    -1,    -1,    -1,    91,     3,     4,
       5,     6,     7,     8,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    27,    28,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    36,    -1,    -1,    39,    -1,    -1,    -1,    -1,    44,
      -1,    -1,    -1,     3,    -1,    -1,    51,    -1,    53,    54,
      -1,    -1,    -1,    -1,    -1,    60,    -1,    62,    63,    -1,
      -1,    -1,    67,    -1,    -1,    -1,    71,    72,    -1,    29,
      30,    -1,    32,    -1,    79,    80,    36,    -1,    -1,    -1,
      -1,    86,    87,    -1,    -1,    -1,    91,    -1,    -1,    -1,
      -1,    51,    -1,    53,    -1,    55,    56,    57,    58,    59,
      60,    61,     3,     4,     5,     6,     7,     8,    68,    -1,
      -1,    71,    -1,    -1,    74,    75,    76,    77,    78,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    27,    28,    -1,    -1,
      -1,    91,    -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    44,    -1,    -1,    -1,    -1,    -1,    -1,
      51,    -1,    53,    -1,    -1,    -1,    -1,    -1,    -1,    60,
      -1,    62,    63,     3,     4,     5,     6,     7,     8,    -1,
      71,    72,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    80,
      -1,    -1,    -1,    -1,    -1,    86,    87,    27,    28,    -1,
      91,    -1,    -1,    -1,    -1,    -1,    36,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    44,    -1,    -1,    -1,    -1,    -1,
      -1,    51,    -1,    53,    -1,    -1,    -1,    -1,    -1,    -1,
      60,    -1,    62,    63,     3,     4,     5,     6,     7,     8,
      -1,    71,    72,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      80,    -1,    -1,    -1,    -1,    -1,    86,    87,    27,    28,
      -1,    91,    -1,    -1,    -1,    -1,    -1,    36,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    44,    -1,    -1,    -1,    -1,
      -1,    -1,    51,    -1,    53,    -1,    -1,    -1,    -1,    -1,
      -1,    60,    -1,    62,    63,     3,     4,     5,     6,     7,
       8,    -1,    71,    72,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    80,    -1,    -1,    -1,    -1,    -1,    86,    87,    -1,
       3,    -1,    91,    -1,    -1,    -1,    -1,    -1,    36,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    44,    -1,    -1,    -1,
      -1,    -1,    -1,    51,    -1,    53,    29,    30,    -1,    32,
      -1,    34,    60,    36,    62,    63,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    71,    72,    -1,    -1,    -1,    51,    -1,
      53,     3,    55,    56,    57,    58,    59,    60,    61,    -1,
      -1,    64,    -1,    91,    -1,    68,    -1,    -1,    71,    -1,
      -1,    74,    75,    76,    77,    78,    -1,    29,    30,    -1,
      32,    -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    51,
      -1,    53,    -1,    55,    56,    57,    58,    59,    60,    61,
      -1,    -1,    -1,    -1,    -1,    -1,    68,    -1,    -1,    71,
      -1,    -1,    74,    75,    76,    77,    78
};

  /* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
     symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    36,    51,    60,   104,   105,   106,   139,     3,    36,
      51,    53,    60,    71,   190,   191,     8,   191,     0,   105,
     106,   139,    50,   107,    94,   191,    94,    64,    65,   134,
     135,   140,    29,    30,    32,    34,    43,    55,    56,    57,
      58,    59,    61,    65,    68,    73,    74,    75,    76,    77,
      78,    95,   108,   109,   110,   111,   112,   113,   114,   119,
     120,   122,   126,   127,   128,   134,   191,    64,    96,    47,
      95,   134,   111,   113,   191,     1,   191,    91,   110,     1,
     191,    95,   109,   119,   126,    96,    97,     1,   115,   116,
     191,    94,   132,    91,   191,    96,    95,    94,   113,    95,
      69,     4,     5,     6,     7,     8,    27,    28,    39,    44,
      54,    62,    63,    67,    72,    79,    80,    86,    87,    91,
     118,   159,   160,   161,   166,   167,   168,   169,   170,   171,
     172,   174,   176,   178,   180,   182,   183,   184,   185,   188,
     189,   191,    96,    98,    99,     1,    31,    33,    35,    37,
      38,    41,    46,    48,    49,    66,    70,    95,    96,   110,
     119,   132,   133,   141,   142,   143,   144,   146,   147,   148,
     151,   154,   155,   156,   157,   158,   159,   183,   191,   113,
     129,   130,   131,    91,   123,   124,   191,    98,    94,    94,
     113,   162,   163,   191,    62,    91,   183,   114,   158,    96,
      98,     1,    22,   102,    21,    81,    82,    83,    25,    26,
     173,    23,    24,    84,    85,   175,    86,    87,   177,    19,
      20,   179,    88,    89,    90,   181,   183,    27,    28,    91,
      93,    99,   116,   117,   176,    96,    96,   172,    96,    53,
      71,   101,    53,    71,     1,    91,   132,   191,     1,    91,
      96,   158,     1,    91,    94,   152,    95,   110,   119,   141,
      42,     1,    96,    97,    98,     9,    10,    11,    12,    13,
      14,    15,    16,    17,    18,   145,   101,   116,    92,    98,
     129,    98,   125,    97,   113,   112,   121,   163,   164,   165,
     166,   191,    99,    91,   191,    92,    92,   159,    96,   168,
     158,   169,   170,   171,   172,   174,   176,   178,   180,   182,
     161,   186,   187,   191,   166,   100,   101,    95,   111,   112,
     142,   144,   149,   158,   166,   191,    96,    95,   158,    96,
      95,   158,    95,   146,   153,   191,    49,   132,    96,   158,
     159,   158,   141,   191,   131,    92,   124,    95,   167,    92,
      96,    95,   112,    98,   125,   101,   176,   186,   183,   101,
      92,    98,   100,    96,   150,   166,    52,   150,    52,    92,
      92,    92,    95,   141,    91,    96,    96,    91,   113,   136,
     137,   138,    96,   165,    95,   165,   100,    92,   166,   161,
      92,   158,    96,   185,    92,   158,   185,   132,   132,   152,
     158,   137,    98,    99,   132,    92,    92,   132,    92,    92,
      92,    92,   138,   117,   132,   132,   132,   132,   132,   100
};

  /* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint8 yyr1[] =
{
       0,   103,   104,   104,   104,   104,   104,   104,   105,   106,
     106,   107,   107,   108,   108,   108,   108,   108,   108,   109,
     109,   109,   110,   110,   111,   111,   112,   113,   113,   113,
     114,   114,   114,   114,   114,   114,   114,   114,   114,   114,
     114,   114,   114,   114,   115,   115,   116,   116,   117,   117,
     118,   118,   119,   119,   120,   120,   121,   121,   122,   122,
     123,   123,   124,   124,   125,   125,   126,   127,   127,   128,
     129,   129,   130,   130,   131,   132,   132,   133,   133,   133,
     133,   133,   133,   134,   135,   135,   135,   135,   136,   136,
     136,   137,   137,   138,   138,   139,   140,   140,   141,   141,
     141,   141,   141,   141,   141,   141,   141,   141,   142,   143,
     143,   144,   144,   145,   145,   145,   145,   145,   145,   145,
     145,   145,   145,   146,   146,   146,   147,   147,   147,   147,
     148,   148,   148,   148,   148,   148,   148,   148,   148,   149,
     149,   150,   150,   151,   151,   151,   152,   152,   153,   153,
     154,   154,   154,   154,   154,   155,   156,   156,   156,   156,
     157,   158,   158,   159,   159,   160,   160,   160,   160,   161,
     161,   161,   162,   162,   163,   164,   164,   165,   165,   165,
     166,   166,   167,   167,   168,   168,   169,   169,   170,   170,
     171,   171,   172,   172,   173,   173,   174,   174,   175,   175,
     175,   175,   176,   176,   177,   177,   178,   178,   179,   179,
     180,   180,   181,   181,   181,   182,   182,   183,   183,   183,
     184,   184,   184,   184,   185,   185,   185,   185,   185,   185,
     186,   186,   187,   187,   188,   188,   188,   188,   189,   189,
     189,   189,   189,   189,   189,   189,   190,   190,   190,   190,
     190,   191,   191
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
       1,     1,     3,     1,     3,     1,     1,     1,     1,     1,
       2,     2,     1,     4,     4,     1,     3,     1,     1,     3,
       1,     5,     1,     3,     1,     3,     1,     3,     1,     3,
       1,     3,     1,     3,     1,     1,     1,     3,     1,     1,
       1,     1,     1,     3,     1,     1,     1,     3,     1,     1,
       1,     3,     1,     1,     1,     1,     4,     1,     2,     2,
       1,     1,     1,     1,     1,     4,     4,     3,     2,     2,
       0,     1,     1,     3,     1,     1,     3,     5,     1,     1,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     1
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

#line 1864 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1430  */
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
#line 255 "grammar.y" /* yacc.c:1648  */
    {
        AST = ast_new();
        imp_add(&AST->imps, (yyvsp[0].imp));
    }
#line 2057 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 3:
#line 260 "grammar.y" /* yacc.c:1648  */
    {
        AST = ast_new();
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2066 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 4:
#line 265 "grammar.y" /* yacc.c:1648  */
    {
        AST = ast_new();
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2075 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 5:
#line 270 "grammar.y" /* yacc.c:1648  */
    {
        imp_add(&AST->imps, (yyvsp[0].imp));
    }
#line 2083 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 6:
#line 274 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2091 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 7:
#line 278 "grammar.y" /* yacc.c:1648  */
    {
        id_add(&ROOT->ids, (yyvsp[0].id));
    }
#line 2099 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 8:
#line 285 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.imp) = imp_new((yyvsp[0].str), &(yylsp[0]));
    }
#line 2107 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 9:
#line 292 "grammar.y" /* yacc.c:1648  */
    {
        ast_blk_t *blk = blk_new_contract(&(yylsp[-1]));

        /* add default constructor */
        id_add(&blk->ids, id_new_ctor((yyvsp[-3].str), NULL, NULL, &(yylsp[-3])));

        (yyval.id) = id_new_contract((yyvsp[-3].str), (yyvsp[-2].exp), blk, &(yyloc));
    }
#line 2120 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 10:
#line 301 "grammar.y" /* yacc.c:1648  */
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
#line 2140 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 11:
#line 319 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 2146 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 12:
#line 321 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yylsp[0]));
    }
#line 2154 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 13:
#line 328 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2163 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 14:
#line 333 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2172 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 15:
#line 338 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_contract(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2181 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 16:
#line 343 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2190 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 17:
#line 348 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2199 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 18:
#line 353 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[0].id));
    }
#line 2208 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 20:
#line 362 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_PUBLIC;
    }
#line 2217 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 21:
#line 367 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2226 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 23:
#line 376 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->mod |= MOD_CONST;
    }
#line 2235 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 25:
#line 385 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.dflt_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.dflt_exp = (yyvsp[-1].exp);
    }
#line 2248 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 26:
#line 397 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);

        if (is_var_id((yyval.id)))
            (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
        else
            (yyval.id)->u_tup.type_exp = (yyvsp[-1].exp);
    }
#line 2261 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 27:
#line 409 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type((yyvsp[0].type), &(yylsp[0]));
    }
#line 2269 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 28:
#line 413 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_NONE, &(yylsp[0]));

        (yyval.exp)->u_type.name = (yyvsp[0].str);
    }
#line 2279 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 29:
#line 419 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_type(TYPE_MAP, &(yylsp[-5]));

        (yyval.exp)->u_type.k_exp = (yyvsp[-3].exp);
        (yyval.exp)->u_type.v_exp = (yyvsp[-1].exp);
    }
#line 2290 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 30:
#line 428 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_ACCOUNT; }
#line 2296 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 31:
#line 429 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BOOL; }
#line 2302 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 32:
#line 430 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_BYTE; }
#line 2308 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 33:
#line 433 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2314 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 34:
#line 434 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT16; }
#line 2320 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 35:
#line 435 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT32; }
#line 2326 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 36:
#line 436 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT64; }
#line 2332 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 37:
#line 437 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_INT8; }
#line 2338 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 38:
#line 438 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_STRING; }
#line 2344 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 39:
#line 439 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2350 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 40:
#line 440 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT16; }
#line 2356 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 41:
#line 441 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT32; }
#line 2362 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 42:
#line 442 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT64; }
#line 2368 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 43:
#line 443 "grammar.y" /* yacc.c:1648  */
    { (yyval.type) = TYPE_UINT8; }
#line 2374 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 45:
#line 449 "grammar.y" /* yacc.c:1648  */
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
#line 2390 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 46:
#line 464 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PRIVATE, &(yylsp[0]));
    }
#line 2398 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 47:
#line 468 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2411 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 48:
#line 479 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = exp_new_null(&(yyloc)); }
#line 2417 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 50:
#line 485 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 2425 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 51:
#line 489 "grammar.y" /* yacc.c:1648  */
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
#line 2440 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 54:
#line 508 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_struct((yyvsp[-4].str), (yyvsp[-1].vect), &(yyloc));
    }
#line 2448 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 55:
#line 512 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2457 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 56:
#line 520 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2470 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 57:
#line 529 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);

        if (is_var_id((yyvsp[-1].id)))
            id_add((yyval.vect), (yyvsp[-1].id));
        else
            id_join((yyval.vect), id_strip((yyvsp[-1].id)));
    }
#line 2483 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 58:
#line 541 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_enum((yyvsp[-4].str), (yyvsp[-2].vect), &(yyloc));
    }
#line 2491 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 59:
#line 545 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.id) = NULL;
    }
#line 2500 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 60:
#line 553 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2509 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 61:
#line 558 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        id_add((yyval.vect), (yyvsp[0].id));
    }
#line 2518 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 62:
#line 566 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[0].str), MOD_PUBLIC | MOD_CONST, &(yylsp[0]));
    }
#line 2526 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 63:
#line 570 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_var((yyvsp[-2].str), MOD_PUBLIC | MOD_CONST, &(yylsp[-2]));
        (yyval.id)->u_var.dflt_exp = (yyvsp[0].exp);
    }
#line 2535 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 66:
#line 583 "grammar.y" /* yacc.c:1648  */
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
#line 2552 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 69:
#line 604 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_ctor((yyvsp[-3].str), (yyvsp[-1].vect), NULL, &(yylsp[-3]));
    }
#line 2560 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 70:
#line 610 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 2566 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 72:
#line 616 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2575 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 73:
#line 621 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].id));
    }
#line 2584 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 74:
#line 629 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[0].id);
        (yyval.id)->u_var.is_param = true;
        (yyval.id)->u_var.type_exp = (yyvsp[-1].exp);
    }
#line 2594 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 75:
#line 637 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 2600 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 76:
#line 638 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 2606 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 77:
#line 643 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2615 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 78:
#line 648 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        /* Unlike state variables, local variables are referenced according to their
         * order of declaration. */
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2626 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 79:
#line 655 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_normal(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2635 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 80:
#line 660 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2644 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 81:
#line 665 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, stmt_new_id((yyvsp[0].id), &(yylsp[0])));
    }
#line 2653 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 82:
#line 670 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 2662 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 83:
#line 678 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_func((yyvsp[-4].str), (yyvsp[-6].mod), (yyvsp[-2].vect), (yyvsp[0].id), NULL, &(yylsp[-4]));
    }
#line 2670 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 84:
#line 684 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PRIVATE; }
#line 2676 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 85:
#line 685 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC; }
#line 2682 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 86:
#line 686 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PAYABLE; }
#line 2688 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 87:
#line 687 "grammar.y" /* yacc.c:1648  */
    { (yyval.mod) = MOD_PUBLIC | MOD_PAYABLE; }
#line 2694 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 88:
#line 691 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = NULL; }
#line 2700 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 89:
#line 692 "grammar.y" /* yacc.c:1648  */
    { (yyval.id) = (yyvsp[-1].id); }
#line 2706 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 92:
#line 699 "grammar.y" /* yacc.c:1648  */
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
#line 2736 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 93:
#line 728 "grammar.y" /* yacc.c:1648  */
    {
        /* We wanted to use a type expression, but we can not store size expressions
         * and declare it as an identifier. */
        (yyval.id) = id_new_param(NULL, (yyvsp[0].exp), &(yylsp[0]));
    }
#line 2746 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 94:
#line 734 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = (yyvsp[-3].id);

        if ((yyval.id)->u_var.size_exps == NULL)
            (yyval.id)->u_var.size_exps = vector_new();

        exp_add((yyval.id)->u_var.size_exps, (yyvsp[-1].exp));
    }
#line 2759 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 95:
#line 746 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.id) = id_new_interface((yyvsp[-3].str), (yyvsp[-1].blk), &(yylsp[-3]));
    }
#line 2767 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 96:
#line 753 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_interface(&(yyloc));
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2776 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 97:
#line 758 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-2].blk);
        id_add(&(yyval.blk)->ids, (yyvsp[-1].id));
    }
#line 2785 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 108:
#line 779 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_null(&(yyloc));
    }
#line 2793 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 109:
#line 786 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_exp((yyvsp[-1].exp), &(yyloc));
    }
#line 2801 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 110:
#line 790 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2810 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 111:
#line 798 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2]));
    }
#line 2818 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 112:
#line 802 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_assign((yyvsp[-3].exp), exp_new_binary((yyvsp[-2].op), (yyvsp[-3].exp), (yyvsp[-1].exp), &(yylsp[-2])), &(yylsp[-2]));
    }
#line 2826 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 113:
#line 808 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 2832 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 114:
#line 809 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 2838 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 115:
#line 810 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 2844 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 116:
#line 811 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 2850 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 117:
#line 812 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 2856 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 118:
#line 813 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_AND; }
#line 2862 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 119:
#line 814 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_XOR; }
#line 2868 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 120:
#line 815 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_BIT_OR; }
#line 2874 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 121:
#line 816 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 2880 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 122:
#line 817 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 2886 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 123:
#line 822 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[0].stmt);
        id_add(LABELS, id_new_label((yyvsp[-2].str), (yyvsp[0].stmt), &(yylsp[-2])));
    }
#line 2895 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 124:
#line 827 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case((yyvsp[-1].exp), &(yyloc));
    }
#line 2903 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 125:
#line 831 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_case(NULL, &(yyloc));
    }
#line 2911 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 126:
#line 838 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 2919 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 127:
#line 842 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-6].stmt);
        stmt_add(&(yyval.stmt)->u_if.elif_stmts, stmt_new_if((yyvsp[-2].exp), (yyvsp[0].blk), &(yylsp[-5])));
    }
#line 2928 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 128:
#line 847 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = (yyvsp[-2].stmt);
        (yyval.stmt)->u_if.else_blk = (yyvsp[0].blk);
    }
#line 2937 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 129:
#line 852 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 2946 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 130:
#line 860 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, NULL, NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2954 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 131:
#line 864 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, NULL, (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2962 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 132:
#line 868 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-3].stmt), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2970 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 133:
#line 872 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, (yyvsp[-4].stmt), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 2978 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 134:
#line 876 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-3].id), &(yylsp[-3])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 2986 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 135:
#line 880 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_FOR, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-3].exp), (yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 2994 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 136:
#line 884 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_exp((yyvsp[-4].exp), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3002 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 137:
#line 888 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_loop(LOOP_ARRAY, stmt_new_id((yyvsp[-4].id), &(yylsp[-4])), (yyvsp[-2].exp), NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3010 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 138:
#line 892 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3019 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 141:
#line 904 "grammar.y" /* yacc.c:1648  */
    { (yyval.exp) = NULL; }
#line 3025 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 143:
#line 910 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch(NULL, (yyvsp[0].blk), &(yyloc));
    }
#line 3033 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 144:
#line 914 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_switch((yyvsp[-2].exp), (yyvsp[0].blk), &(yyloc));
    }
#line 3041 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 145:
#line 918 "grammar.y" /* yacc.c:1648  */
    {
        yyerrok;
        (yyval.stmt) = NULL;
    }
#line 3050 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 146:
#line 925 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = NULL; }
#line 3056 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 147:
#line 926 "grammar.y" /* yacc.c:1648  */
    { (yyval.blk) = (yyvsp[-1].blk); }
#line 3062 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 148:
#line 931 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = blk_new_switch(&(yyloc));
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3071 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 149:
#line 936 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.blk) = (yyvsp[-1].blk);
        stmt_add(&(yyval.blk)->stmts, (yyvsp[0].stmt));
    }
#line 3080 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 150:
#line 944 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_CONTINUE, NULL, &(yyloc));
    }
#line 3088 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 151:
#line 948 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_jump(STMT_BREAK, NULL, &(yyloc));
    }
#line 3096 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 152:
#line 952 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return(NULL, &(yyloc));
    }
#line 3104 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 153:
#line 956 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_return((yyvsp[-1].exp), &(yyloc));
    }
#line 3112 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 154:
#line 960 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_goto((yyvsp[-1].str), &(yylsp[-1]));
    }
#line 3120 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 155:
#line 967 "grammar.y" /* yacc.c:1648  */
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
#line 3140 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 160:
#line 993 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.stmt) = stmt_new_blk((yyvsp[0].blk), &(yyloc));
    }
#line 3148 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 162:
#line 1001 "grammar.y" /* yacc.c:1648  */
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
#line 3163 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 164:
#line 1016 "grammar.y" /* yacc.c:1648  */
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
#line 3183 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 165:
#line 1034 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_DELETE; }
#line 3189 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 166:
#line 1035 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_INSERT; }
#line 3195 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 167:
#line 1036 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_QUERY; }
#line 3201 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 168:
#line 1037 "grammar.y" /* yacc.c:1648  */
    { (yyval.sql) = SQL_UPDATE; }
#line 3207 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 170:
#line 1043 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3215 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 171:
#line 1047 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3223 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 172:
#line 1054 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_alloc((yyvsp[0].exp), &(yylsp[0]));
    }
#line 3231 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 173:
#line 1058 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-3].exp);

        if ((yyval.exp)->u_alloc.size_exps == NULL)
            (yyval.exp)->u_alloc.size_exps = vector_new();

        exp_add((yyval.exp)->u_alloc.size_exps, (yyvsp[-1].exp));
    }
#line 3244 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 174:
#line 1070 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_init((yyvsp[-2].vect), &(yyloc));
    }
#line 3252 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 175:
#line 1077 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3261 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 176:
#line 1082 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3270 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 179:
#line 1092 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3278 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 181:
#line 1100 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_ternary((yyvsp[-4].exp), (yyvsp[-2].exp), (yyvsp[0].exp), &(yyloc));
    }
#line 3286 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 183:
#line 1108 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3294 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 185:
#line 1116 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3302 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 187:
#line 1124 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_OR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3310 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 189:
#line 1132 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_XOR, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3318 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 191:
#line 1140 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary(OP_BIT_AND, (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3326 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 193:
#line 1148 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3334 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 194:
#line 1154 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_EQ; }
#line 3340 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 195:
#line 1155 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NE; }
#line 3346 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 197:
#line 1161 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3354 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 198:
#line 1167 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LT; }
#line 3360 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 199:
#line 1168 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GT; }
#line 3366 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 200:
#line 1169 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LE; }
#line 3372 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 201:
#line 1170 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_GE; }
#line 3378 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 203:
#line 1176 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3386 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 204:
#line 1182 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_ADD; }
#line 3392 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 205:
#line 1183 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_SUB; }
#line 3398 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 207:
#line 1189 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3406 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 208:
#line 1195 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_RSHIFT; }
#line 3412 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 209:
#line 1196 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_LSHIFT; }
#line 3418 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 211:
#line 1202 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_binary((yyvsp[-1].op), (yyvsp[-2].exp), (yyvsp[0].exp), &(yylsp[-1]));
    }
#line 3426 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 212:
#line 1208 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MUL; }
#line 3432 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 213:
#line 1209 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DIV; }
#line 3438 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 214:
#line 1210 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_MOD; }
#line 3444 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 216:
#line 1216 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_cast((yyvsp[-2].type), (yyvsp[0].exp), &(yylsp[-2]));
    }
#line 3452 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 218:
#line 1224 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary((yyvsp[-1].op), true, (yyvsp[0].exp), &(yyloc));
    }
#line 3460 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 219:
#line 1228 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[0].exp);
    }
#line 3468 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 220:
#line 1234 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_INC; }
#line 3474 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 221:
#line 1235 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_DEC; }
#line 3480 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 222:
#line 1236 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NEG; }
#line 3486 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 223:
#line 1237 "grammar.y" /* yacc.c:1648  */
    { (yyval.op) = OP_NOT; }
#line 3492 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 225:
#line 1243 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_array((yyvsp[-3].exp), (yyvsp[-1].exp), &(yyloc));
    }
#line 3500 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 226:
#line 1247 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(false, (yyvsp[-3].exp), (yyvsp[-1].vect), &(yyloc));
    }
#line 3508 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 227:
#line 1251 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_access((yyvsp[-2].exp), exp_new_id((yyvsp[0].str), &(yylsp[0])), &(yyloc));
    }
#line 3516 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 228:
#line 1255 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_INC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3524 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 229:
#line 1259 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_unary(OP_DEC, false, (yyvsp[-1].exp), &(yyloc));
    }
#line 3532 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 230:
#line 1265 "grammar.y" /* yacc.c:1648  */
    { (yyval.vect) = NULL; }
#line 3538 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 232:
#line 1271 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = vector_new();
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3547 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 233:
#line 1276 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.vect) = (yyvsp[-2].vect);
        exp_add((yyval.vect), (yyvsp[0].exp));
    }
#line 3556 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 235:
#line 1285 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_id((yyvsp[0].str), &(yyloc));
    }
#line 3564 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 236:
#line 1289 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = (yyvsp[-1].exp);
    }
#line 3572 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 237:
#line 1293 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_call(true, exp_new_id((yyvsp[-3].str), &(yylsp[-3])), (yyvsp[-1].vect), &(yylsp[-3]));
    }
#line 3580 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 238:
#line 1300 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_null(&(yyloc));
    }
#line 3588 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 239:
#line 1304 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(true, &(yyloc));
    }
#line 3596 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 240:
#line 1308 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_bool(false, &(yyloc));
    }
#line 3604 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 241:
#line 1312 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNu64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3615 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 242:
#line 1319 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNo64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3626 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 243:
#line 1326 "grammar.y" /* yacc.c:1648  */
    {
        uint64_t v;

        sscanf((yyvsp[0].str), "%"SCNx64, &v);
        (yyval.exp) = exp_new_lit_i64(v, &(yyloc));
    }
#line 3637 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 244:
#line 1333 "grammar.y" /* yacc.c:1648  */
    {
        double v;

        sscanf((yyvsp[0].str), "%lf", &v);
        (yyval.exp) = exp_new_lit_f64(v, &(yyloc));
    }
#line 3648 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 245:
#line 1340 "grammar.y" /* yacc.c:1648  */
    {
        (yyval.exp) = exp_new_lit_str((yyvsp[0].str), &(yyloc));
    }
#line 3656 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 246:
#line 1346 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("contract"); }
#line 3662 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 247:
#line 1347 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("import"); }
#line 3668 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 248:
#line 1348 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("index"); }
#line 3674 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 249:
#line 1349 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("interface"); }
#line 3680 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 250:
#line 1350 "grammar.y" /* yacc.c:1648  */
    { (yyval.str) = xstrdup("table"); }
#line 3686 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;

  case 251:
#line 1355 "grammar.y" /* yacc.c:1648  */
    {
        if (strlen((yyvsp[0].str)) > NAME_MAX_LEN)
            ERROR(ERROR_TOO_LONG_ID, &(yylsp[0]), NAME_MAX_LEN, strlen((yyvsp[0].str)));

        (yyval.str) = (yyvsp[0].str);
    }
#line 3697 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
    break;


#line 3701 "/home/wrpark/blocko/src/github.com/aergoio/aergo/contract/native/grammar.tab.c" /* yacc.c:1648  */
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
#line 1364 "grammar.y" /* yacc.c:1907  */


static void
yyerror(YYLTYPE *yylloc, parse_t *parse, void *scanner, const char *msg)
{
    ERROR(ERROR_SYNTAX, yylloc, msg);
}

/* end of grammar.y */
