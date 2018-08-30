%{
/**
 *  @file   scanner.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "throw.h"
#include "xutil.h"
#include "parser.h"

#define YY_DECL                                                                \
    int yylex(YYSTYPE *lval, YYLTYPE *lloc, yyscan_t yyscanner)

#undef  YY_INPUT
#define YY_INPUT(buf, result, max)                                             \
    do {                                                                       \
        result = xfread(yyin, max, buf);                                       \
    } while (0)

#define YY_EXTRA_TYPE       scan_t *

#define YY_UPDATE_LINE()                                                       \
    do {                                                                       \
        yyextra->lloc.line++;                                                  \
        yyextra->lloc.offset += yyextra->lloc.column;                          \
        yyextra->lloc.column = 1;                                              \
    } while (0)

#define YY_UPDATE_COL()          (yyextra->lloc.column += yyleng)

#define YY_STRCAT()                                                            \
    do {                                                                       \
        if (yyextra->offset + yyleng >= STR_MAX_LEN) {                         \
            YY_ERROR(ERROR_SCAN_STR_TOO_LONG);                                 \
            YY_UPDATE_COL();                                                   \
            YY_RECOVER();                                                      \
        }                                                                      \
        strcat(yyextra->buf, yytext);                                          \
        yyextra->offset += yyleng;                                             \
    } while (0)

#define YY_ERROR(errmsg)                                                       \
    do {                                                                       \
        int i;                                                                 \
        char line[80 + 1];                                                     \
        char errloc[512];                                                      \
        char errdesc[ERROR_MAX_DESC_LEN];                                      \
        FILE *fp;                                                              \
        if (yyextra->errcnt > 10)                                              \
            yyterminate();                                                     \
        fp = xfopen(yyextra->path, "r");                                       \
        xfseek(fp, yyextra->lloc.offset);                                      \
        xfgets(fp, sizeof(line), line);                                        \
        snprintf(errloc, sizeof(errloc), "%s:%d:%d", yyextra->file,            \
                 yyextra->lloc.line, yyextra->lloc.column);                    \
        snprintf(errdesc, sizeof(errdesc), "%s\n%s", errmsg, line);            \
        for (i = 0; i < yyextra->lloc.column - 1; i++) {                       \
            strcat(errdesc, " ");                                              \
        }                                                                      \
        strcat(errdesc, ANSI_GREEN"^"ANSI_DEFAULT);                            \
        ERROR(errloc, errdesc);                                                \
        xfclose(fp);                                                           \
        yyextra->errcnt++;                                                     \
    } while (0)

#define YY_RECOVER()                                                           \
    do {                                                                       \
        register int c;                                                        \
        yyleng = 1;                                                            \
        for ( ; ; ) {                                                          \
            c = input(yyscanner);                                              \
            if (c == '\r' || c == '\n') {                                      \
                unput(c);                                                      \
                break;                                                         \
            }                                                                  \
            YY_UPDATE_COL();                                                   \
        }                                                                      \
    } while (0)

#define YY_RETURN(tok)                                                         \
    do {                                                                       \
        YY_UPDATE_COL();                                                       \
        return (tok);                                                          \
    } while (0)

%}

%option case-sensitive
%option never-interactive
%option debug
%option reentrant
%option noyylineno
%option noyywrap

%x  C_COMMENT
%x  D_QUOTE

whitespace      ([ ]|\t|(\xE3\x80\x80))
newline         (\r|\n|\r\n)

D               [0-9]
L               [a-zA-Z\x80-\xFF_]
IS              (u|U)

integer         [+-]?{D}+{IS}?
float           [+-]?({D}*"."{D}+|{D}+"."{D}*)

id              {L}({L}|{D})*

%%

{whitespace}+               YY_UPDATE_COL();
{newline}                   YY_UPDATE_LINE();

"//"[^\r\n]*                YY_UPDATE_COL();

"/*"                        { BEGIN(C_COMMENT); YY_UPDATE_COL(); }
<C_COMMENT>"*/"             { BEGIN(INITIAL); YY_UPDATE_COL(); }
<C_COMMENT>[^*\r\n]+        YY_UPDATE_COL();
<C_COMMENT>"*"              YY_UPDATE_COL();
<C_COMMENT>{newline}        YY_UPDATE_LINE();
<C_COMMENT><<EOF>> {
    YY_ERROR(ERROR_SCAN_UNTERMINATED_COMMENT);
    yyterminate();
}

\"(\\.|[^\\"\n\r])*\"       YY_RETURN(L_STR); 
    /*
"\"" {
    BEGIN(D_QUOTE);
    YY_UPDATE_COL();
    yyextra->buf[0] = '\0';
    yyextra->offset = 0;
}
<D_QUOTE>[^"\\\r\n]+        { YY_UPDATE_COL(); YY_STRCAT(); }
<D_QUOTE>"\"" {
    BEGIN(INITIAL);
    YY_UPDATE_COL();
    fprintf(stderr, "result: %s\n", yyextra->buf);
    //return L_STR;
}
<D_QUOTE>"\\"{whitespace}*{newline} {
    YY_UPDATE_COL();
    YY_UPDATE_LINE();
}
<D_QUOTE>"\\"               { YY_UPDATE_COL(); YY_STRCAT(); }
<D_QUOTE>{newline} {
    YY_ERROR(ERROR_SCAN_UNTERMINATED_STRING);
    YY_UPDATE_LINE();
}
<D_QUOTE><<EOF>> {
    YY_ERROR(ERROR_SCAN_UNTERMINATED_STRING);
    yyterminate();
}
    */

{integer}           YY_RETURN(L_INT);
{float}             YY_RETURN(L_FLOAT);

"+="                YY_RETURN(OP_ADD_ASSIGN);
"-="                YY_RETURN(OP_SUB_ASSIGN);
"*="                YY_RETURN(OP_MUL_ASSIGN);
"/="                YY_RETURN(OP_DIV_ASSIGN);
"%="                YY_RETURN(OP_MOD_ASSIGN);
"&="                YY_RETURN(OP_AND_ASSIGN);
"^="                YY_RETURN(OP_XOR_ASSIGN);
"|="                YY_RETURN(OP_OR_ASSIGN);
">>"                YY_RETURN(OP_RSHIFT);
"<<"                YY_RETURN(OP_LSHIFT);
"++"                YY_RETURN(OP_INC);
"--"                YY_RETURN(OP_DEC);
"&&"                YY_RETURN(OP_AND);
"||"                YY_RETURN(OP_OR);
"<="                YY_RETURN(OP_LE);
">="                YY_RETURN(OP_GE);
"=="                YY_RETURN(OP_EQ);
"!="                YY_RETURN(OP_NE);

"{"                 YY_RETURN('{');
"}"                 YY_RETURN('}');
"("                 YY_RETURN('(');
")"                 YY_RETURN(')');
"["                 YY_RETURN('[');
"]"                 YY_RETURN(']');
";"                 YY_RETURN(';');
":"                 YY_RETURN(':');
","                 YY_RETURN(',');
"?"                 YY_RETURN('?');
"."                 YY_RETURN('.');
"+"                 YY_RETURN('+');
"-"                 YY_RETURN('-');
"*"                 YY_RETURN('*');
"/"                 YY_RETURN('/');
"%"                 YY_RETURN('%');
"<"                 YY_RETURN('<');
">"                 YY_RETURN('>');
"&"                 YY_RETURN('&');
"|"                 YY_RETURN('|');
"="                 YY_RETURN('=');

"account"           YY_RETURN(K_ACCOUNT);
"bool"              YY_RETURN(K_BOOL);
"break"             YY_RETURN(K_BREAK);
"byte"              YY_RETURN(K_BYTE);
"case"              YY_RETURN(K_CASE);
"const"             YY_RETURN(K_CONST);
"constructor"       YY_RETURN(K_CONSTRUCTOR);
"continue"          YY_RETURN(K_CONTINUE);
"contract"          YY_RETURN(K_CONTRACT);
"default"           YY_RETURN(K_DEFAULT);
"double"            YY_RETURN(K_DOUBLE);
"else"              YY_RETURN(K_ELSE);
"false"             YY_RETURN(K_FALSE);
"file"              YY_RETURN(K_FILE);
"float"             YY_RETURN(K_FLOAT);
"for"               YY_RETURN(K_FOR);
"func"              YY_RETURN(K_FUNC);
"if"                YY_RETURN(K_IF);
"int"               YY_RETURN(K_INT);
"int16"             YY_RETURN(K_INT16);
"int32"             YY_RETURN(K_INT32);
"int64"             YY_RETURN(K_INT64);
"int8"              YY_RETURN(K_INT8);
"map"               YY_RETURN(K_MAP);
"new"               YY_RETURN(K_NEW);
"null"              YY_RETURN(K_NULL);
"payable"           YY_RETURN(K_PAYABLE);
"pragma"            YY_RETURN(K_PRAGMA);
"private"           YY_RETURN(K_PRIVATE);
"public"            YY_RETURN(K_PUBLIC);
"return"            YY_RETURN(K_RETURN);
"string"            YY_RETURN(K_STRING);
"struct"            YY_RETURN(K_STRUCT);
"switch"            YY_RETURN(K_SWITCH);
"true"              YY_RETURN(K_TRUE);
"uint"              YY_RETURN(K_UINT);
"uint16"            YY_RETURN(K_UINT16);
"uint32"            YY_RETURN(K_UINT32);
"uint64"            YY_RETURN(K_UINT64);
"uint8"             YY_RETURN(K_UINT8);
"version"           YY_RETURN(K_VERSION);
"while"             YY_RETURN(K_WHILE);

    /*
"commit"            YY_RETURN(K_COMMIT);
"delete"            YY_RETURN(K_DELETE);
"insert"            YY_RETURN(K_INSERT);
"rollback"          YY_RETURN(K_ROLLBACK);
"select"            YY_RETURN(K_SELECT);
    */

{id} {
    lval->str = malloc(yyleng + 1);
    xstrcpy(lval->str, yytext);
    YY_RETURN(ID);
}

. {
    YY_ERROR(ERROR_SCAN_UNKNOWN_CHAR);
    YY_UPDATE_COL();
    YY_RECOVER();
}

<<EOF>>                         yyterminate();

%%

/* end of scanner.l */
