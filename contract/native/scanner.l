%{

/**
 *  @file   scanner.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "parse.h"

#define YY_DECL                                                                \
    int yyscan(YYSTYPE *lval, YYLTYPE *lloc, yyscan_t yyscanner)

#undef  YY_INPUT
#define YY_INPUT(buf, result, MAX)                                             \
    do {                                                                       \
        (result) = MIN((MAX), yyextra->len - yyextra->pos);                    \
        if ((result) > 0) {                                                    \
            memcpy(buf, yyextra->src + yyextra->pos, (result));                \
            yyextra->pos += (result);                                          \
        }                                                                      \
    } while (0)

#define YY_EXTRA_TYPE           yyparam_t *
#define YY_BUF                  &yyextra->buf

#define yy_update_last_col()    lloc->last_col += yyleng

#define yy_update_last_line()                                                  \
    do {                                                                       \
        lloc->last_line++;                                                     \
        lloc->last_offset += lloc->last_col;                                   \
        lloc->last_col = 1;                                                    \
    } while (0)

#define yy_update_first()                                                      \
    do {                                                                       \
        lloc->first_line = lloc->last_line;                                    \
        lloc->first_col = lloc->last_col;                                      \
        lloc->first_offset = lloc->last_offset;                                \
    } while (0)

#define yy_text()               xstrndup(yytext, yyleng)
#define yy_literal()            xstrndup(yytext + 1, yyleng - 2)

#define yy_error(ec)            TRACE(ec, lloc)

#define yy_recover()                                                           \
    do {                                                                       \
        register int c;                                                        \
        for ( ; ; ) {                                                          \
            c = input(yyscanner);                                              \
            if (c == '\r' || c == '\n') {                                      \
                unput(c);                                                      \
                break;                                                         \
            }                                                                  \
            else if (c == '\0') {                                              \
                yyterminate();                                                 \
            }                                                                  \
            lloc->last_col++;                                                  \
        }                                                                      \
    } while (0)

#define yy_return(tok)                                                         \
    do {                                                                       \
        yy_update_first();                                                     \
        yy_update_last_col();                                                  \
        return (tok);                                                          \
    } while (0)

#define yy_return_buf(tok)                                                     \
    do {                                                                       \
        lval->str = xstrndup(strbuf_text(YY_BUF), strbuf_length(YY_BUF));      \
        strbuf_reset(YY_BUF);                                                  \
        return (tok);                                                          \
    } while (0)

#define yy_scan_comment()                                                      \
    do {                                                                       \
        char c, prev = '\0';                                                   \
        while ((c = input(yyscanner)) != '\0') {                               \
            if (c == '/' && prev == '*') {                                     \
                lloc->last_col++;                                              \
                break;                                                         \
            }                                                                  \
            if (c == '\n')                                                     \
                yy_update_last_line();                                         \
            else                                                               \
                lloc->last_col++;                                              \
            prev = c;                                                          \
        }                                                                      \
        if (c == '\0') {                                                       \
            yy_error(ERROR_UNTERMINATED_COMMENT);                              \
            yyterminate();                                                     \
        }                                                                      \
    } while (0)

#define yy_scan_literal()                                                      \
    do {                                                                       \
        char c, prev = '\0';                                                   \
        while ((c = input(yyscanner)) != '\0') {                               \
            if (c == '"' && prev != '\\') {                                    \
                lloc->last_col++;                                              \
                break;                                                         \
            }                                                                  \
            if (c == '\n') {                                                   \
                strbuf_reset(YY_BUF);                                          \
                unput(c);                                                      \
                yy_error(ERROR_UNTERMINATED_STRING);                           \
                break;                                                         \
            }                                                                  \
            else {                                                             \
                lloc->last_col++;                                              \
                strbuf_append(YY_BUF, &c, 1);                                  \
            }                                                                  \
            prev = c;                                                          \
        }                                                                      \
        if (c == '\0') {                                                       \
            yy_error(ERROR_UNTERMINATED_STRING);                               \
            yyterminate();                                                     \
        }                                                                      \
    } while (0)

%}

%option case-sensitive
%option never-interactive
%option debug
%option reentrant
%option noyylineno
%option noyywrap

%x DIRECTIVE

ws                          [ \t\f\v]
nl                          ([\r\n]|\r\n)

D                           [0-9]
L                           [a-zA-Z\x80-\xFF_]
H                           [a-fA-F0-9]

integer                     [+-]?{D}+
float                       [+-]?({D}*"."{D}+|{D}+"."{D}*)
hexa                        0[xX]{H}+

id                          {L}({L}|{D})*

%%

{ws}+                       yy_update_last_col();
{nl}                        yy_update_last_line();

"//"[^\r\n]*                yy_update_last_col();

"/*" {
    yy_update_first();
    yy_update_last_col();
    yy_scan_comment();
}

"\"" {
    yy_update_first();
    yy_update_last_col();
    yy_scan_literal();
    yy_return_buf(L_STR);
}

{integer}                   yy_return(L_INT);
{float}                     yy_return(L_FLOAT);
{hexa}                      yy_return(L_HEXA);

"+="                        yy_return(ASSIGN_ADD);
"-="                        yy_return(ASSIGN_SUB);
"*="                        yy_return(ASSIGN_MUL);
"/="                        yy_return(ASSIGN_DIV);
"%="                        yy_return(ASSIGN_MOD);
"&="                        yy_return(ASSIGN_AND);
"^="                        yy_return(ASSIGN_XOR);
"|="                        yy_return(ASSIGN_OR);
">>="                       yy_return(ASSIGN_RS);
"<<="                       yy_return(ASSIGN_LS);
">>"                        yy_return(SHIFT_R);
"<<"                        yy_return(SHIFT_L);
"&&"                        yy_return(CMP_AND);
"||"                        yy_return(CMP_OR);
"<="                        yy_return(CMP_LE);
">="                        yy_return(CMP_GE);
"=="                        yy_return(CMP_EQ);
"!="                        yy_return(CMP_NE);
"++"                        yy_return(UNARY_INC);
"--"                        yy_return(UNARY_DEC);

"{"                         yy_return('{');
"}"                         yy_return('}');
"("                         yy_return('(');
")"                         yy_return(')');
"["                         yy_return('[');
"]"                         yy_return(']');
";"                         yy_return(';');
":"                         yy_return(':');
","                         yy_return(',');
"?"                         yy_return('?');
"."                         yy_return('.');
"+"                         yy_return('+');
"-"                         yy_return('-');
"*"                         yy_return('*');
"/"                         yy_return('/');
"%"                         yy_return('%');
"<"                         yy_return('<');
">"                         yy_return('>');
"&"                         yy_return('&');
"|"                         yy_return('|');
"="                         yy_return('=');
"^"                         yy_return('^');
"!"                         yy_return('!');
"'"                         yy_return('\'');

"account"                   yy_return(K_ACCOUNT);
"bool"                      yy_return(K_BOOL);
"break"                     yy_return(K_BREAK);
"byte"                      yy_return(K_BYTE);
"case"                      yy_return(K_CASE);
"const"                     yy_return(K_CONST);
"continue"                  yy_return(K_CONTINUE);
"contract"                  yy_return(K_CONTRACT);
"default"                   yy_return(K_DEFAULT);
"double"                    yy_return(K_DOUBLE);
"else"                      yy_return(K_ELSE);
"false"                     yy_return(K_FALSE);
"float"                     yy_return(K_FLOAT);
"for"                       yy_return(K_FOR);
"foreach"                   yy_return(K_FOREACH);
"func"                      yy_return(K_FUNC);
"if"                        yy_return(K_IF);
"in"                        yy_return(K_IN);
"index"                     yy_return(K_INDEX);
"int"                       yy_return(K_INT);
"int16"                     yy_return(K_INT16);
"int32"                     yy_return(K_INT32);
"int64"                     yy_return(K_INT64);
"local"                     yy_return(K_LOCAL);
"map"                       yy_return(K_MAP);
"new"                       yy_return(K_NEW);
"null"                      yy_return(K_NULL);
"readonly"                  yy_return(K_READONLY);
"return"                    yy_return(K_RETURN);
"shared"                    yy_return(K_SHARED);
"string"                    yy_return(K_STRING);
"struct"                    yy_return(K_STRUCT);
"switch"                    yy_return(K_SWITCH);
"table"                     yy_return(K_TABLE);
"transfer"                  yy_return(K_TRANSFER);
"true"                      yy_return(K_TRUE);
"uint"                      yy_return(K_UINT);
"uint16"                    yy_return(K_UINT16);
"uint32"                    yy_return(K_UINT32);
"uint64"                    yy_return(K_UINT64);

"create"                    yy_return(K_CREATE);
"delete"                    yy_return(K_DELETE);
"drop"                      yy_return(K_DROP);
"insert"                    yy_return(K_INSERT);
"select"                    yy_return(K_SELECT);
"update"                    yy_return(K_UPDATE);

"#file"                     BEGIN(DIRECTIVE);
<DIRECTIVE>{ws}+
<DIRECTIVE>\"[^\"\n\r]+\"   lloc->path = yy_literal();
<DIRECTIVE>{D}+" "{D}+ {
    lloc->last_line = atoi(strtok(yy_text(), " "));
    lloc->last_offset = atoi(strtok(NULL, " "));
}
<DIRECTIVE>{nl}             BEGIN(INITIAL);

{id} {
    lval->str = yy_text();
    yy_return(ID);
}

. {
    yy_update_first();
    yy_update_last_col();
    yy_error(ERROR_UNKNOWN_CHAR);
    yy_recover();
}

<<EOF>>                     yyterminate();

%%

int
yylex(YYSTYPE *lval, YYLTYPE *lloc, void *yyscanner)
{
    int token;
    yyparam_t *param = yyget_extra(yyscanner);

    if (param->adj_token != 0) {
        token = param->adj_token;
        *lloc = param->adj_pos;
        param->adj_token = 0;
        return token;
    }

    return yyscan(lval, lloc, yyscanner);
}

void
yylex_set_token(void *yyscanner, int token, YYLTYPE *lloc)
{
    yyparam_t *param = yyget_extra(yyscanner);

    param->adj_token = token;
    param->adj_pos = *lloc;
}

/* end of scanner.l */
