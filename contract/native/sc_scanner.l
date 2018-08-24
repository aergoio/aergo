%{
/**
 *  @file   sc_scanner.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "sc_common.h"

#include "sc_throw.h"
#include "sc_util.h"
#include "sc_parser.h"

#define YY_EXTRA_TYPE       sc_yyextra_t *

#undef  YY_INPUT
#define YY_INPUT(buf, result, max)                                             \
    do {                                                                       \
        result = sc_fread(yyin, max, buf);                                     \
    } while (0)

#define yy_update_line()                                                       \
    do {                                                                       \
        yyextra->lloc.line++;                                                  \
        yyextra->lloc.offset += yyextra->lloc.column;                          \
        yyextra->lloc.column = 1;                                              \
    } while (0)

#define yy_update_column()          (yyextra->lloc.column += yyleng)

#define yy_strcat()                                                            \
    do {                                                                       \
        if (yyextra->offset + yyleng >= SC_STR_MAX_LEN) {                      \
            yy_error(ERROR_SCAN_STR_TOO_LONG);                                 \
            yy_update_column();                                                \
            yy_recover();                                                      \
        }                                                                      \
        strcat(yyextra->buf, yytext);                                          \
        yyextra->offset += yyleng;                                             \
    } while (0)

#define yy_error(errmsg)                                                       \
    do {                                                                       \
        int i;                                                                 \
        char line[80 + 1];                                                     \
        char errloc[512];                                                      \
        char errdesc[ERROR_MAX_DESC_LEN];                                      \
        FILE *fp;                                                              \
        if (yyextra->errcnt > 10)                                              \
            yyterminate();                                                     \
        fp = sc_fopen(yyextra->path, "r");                                     \
        sc_fseek(fp, yyextra->lloc.offset);                                    \
        sc_fgets(fp, sizeof(line), line);                                      \
        snprintf(errloc, sizeof(errloc), "%s:%d:%d", yyextra->file,            \
                 yyextra->lloc.line, yyextra->lloc.column);                    \
        snprintf(errdesc, sizeof(errdesc), "%s\n%s", errmsg, line);            \
        for (i = 0; i < yyextra->lloc.column - 1; i++) {                       \
            strcat(errdesc, " ");                                              \
        }                                                                      \
        strcat(errdesc, ANSI_GREEN"^"ANSI_DEFAULT);                            \
        sc_error(errloc, errdesc);                                             \
        sc_fclose(fp);                                                         \
        yyextra->errcnt++;                                                     \
    } while (0)

#define yy_recover()                                                           \
    do {                                                                       \
        register int c;                                                        \
        yyleng = 1;                                                            \
        for ( ; ; ) {                                                          \
            c = input(yyscanner);                                              \
            if (c == '\r' || c == '\n') {                                      \
                unput(c);                                                      \
                break;                                                         \
            }                                                                  \
            yy_update_column();                                                \
        }                                                                      \
    } while (0)

%}

%option case-insensitive
%option never-interactive
%option debug
%option reentrant
%option noyylineno
%option noyywrap
%option prefix="sc_yy"

%x  C_COMMENT
%x  D_QUOTE

whitespace          ([ ]|\t|(\xE3\x80\x80))
newline             (\r|\n|\r\n)

cpp_comment         "//"
c_comment_start     "/*"
c_comment_end       "*/"

lparen          "("
rparen          ")"
lbrace          "{"
rbrace          "}"
lbracket        "["
rbracket        "]"
dot             "."
comma           ","
colon           ":"
semicolon       ";"
question        "?"
plus            "+"
minus           "-"
asterisk        "*"
slash           "/"
percent         "%"
equal           "="
dquote          "\""
sign            ({plus}|{minus})

cmp_eq          "=="
cmp_ne          "!="
cmp_lt          "<"
cmp_gt          ">"
cmp_le          "<="
cmp_ge          ">="

digit           [:digit:]
alpha           [a-zA-Z\x80-\xFF_]

integer         {digit}+
float           ({digit}*{dot}{digit}+|{digit}+{dot}{digit}*)
exponent        ({integer}|{float})[eE]{sign}?{digit}+

identifier      {alpha}({alpha}|{digit})*

%%

{whitespace}+                   yy_update_column();
{newline}                       yy_update_line();

{cpp_comment}[^\r\n]*           yy_update_column();

{c_comment_start} {
    BEGIN(C_COMMENT);
    yy_update_column();
}
<C_COMMENT>{c_comment_end} {
    BEGIN(INITIAL);
    yy_update_column();
}
<C_COMMENT>[^*\r\n]+            yy_update_column();
<C_COMMENT>{asterisk}           yy_update_column();
<C_COMMENT>{newline}            yy_update_line();
<C_COMMENT><<EOF>> {
    yy_error(ERROR_SCAN_UNTERMINATED_COMMENT);
}             

{dquote} {
    BEGIN(D_QUOTE);
    yy_update_column();
    yyextra->buf[0] = '\0';
    yyextra->offset = 0;
}
<D_QUOTE>[^"\\\r\n]+ {
    yy_update_column();
    yy_strcat();
}
<D_QUOTE>{dquote} {
    BEGIN(INITIAL);
    yy_update_column();
    fprintf(stderr, "result: %s\n", yyextra->buf);
    //return L_STR;
}
<D_QUOTE>"\\"{whitespace}*{newline} {
    yy_update_column();
    yy_update_line();
}
<D_QUOTE>"\\" {
    yy_update_column();
    yy_strcat();
}
<D_QUOTE>{newline} {
    yy_error(ERROR_SCAN_UNTERMINATED_STRING);
    yy_update_line();
}
<D_QUOTE><<EOF>> {
    yy_error(ERROR_SCAN_UNTERMINATED_STRING);
    yyterminate();
}

{integer} {
    yy_update_column();
    // return L_INT;
}
{float} {
    yy_update_column();
    // return L_FLOAT;
}
{exponent} {
    yy_update_column();
    // return L_EXPN;
}

. {
    yy_error(ERROR_SCAN_UNKNOWN_CHAR);
    yy_update_column();
    yy_recover();
}

<<EOF>>                         yyterminate();

%%
