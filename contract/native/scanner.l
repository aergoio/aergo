%{

/**
 *  @file   scanner.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "parse.h"

#define YY_DECL                                                                          \
    int yyscan(YYSTYPE *yylval, YYLTYPE *yylloc, yyscan_t yyscanner)

#undef  YY_INPUT
#define YY_INPUT(buf, result, max)                                                       \
    do {                                                                                 \
        (result) = MIN((max), yyextra->len - yyextra->pos);                              \
        if ((result) > 0) {                                                              \
            memcpy(buf, yyextra->src + yyextra->pos, (result));                          \
            yyextra->pos += (result);                                                    \
        }                                                                                \
    } while (0)

#define YY_EXTRA_TYPE           parse_t *
#define YY_BUF                  &yyextra->buf

#define yy_error(ec)            ERROR(ec, yylloc)

#define yy_recover()                                                                     \
    do {                                                                                 \
        register int c;                                                                  \
        for ( ; ; ) {                                                                    \
            c = input(yyscanner);                                                        \
            if (c == '\r' || c == '\n') {                                                \
                unput(c);                                                                \
                break;                                                                   \
            }                                                                            \
            else if (c == '\0') {                                                        \
                yyterminate();                                                           \
            }                                                                            \
            src_pos_update_col(yylloc, 1);                                               \
        }                                                                                \
    } while (0)

#define yy_scan_comment()                                                                \
    do {                                                                                 \
        char c, prev = '\0';                                                             \
        while ((c = input(yyscanner)) != '\0') {                                         \
            if (c == '/' && prev == '*') {                                               \
                src_pos_update_col(yylloc, 1);                                           \
                break;                                                                   \
            }                                                                            \
            if (c == '\n')                                                               \
                src_pos_update_line(yylloc);                                             \
            else                                                                         \
                src_pos_update_col(yylloc, 1);                                           \
            prev = c;                                                                    \
        }                                                                                \
        if (c == '\0') {                                                                 \
            yy_error(ERROR_UNTERMINATED_COMMENT);                                        \
            yyterminate();                                                               \
        }                                                                                \
    } while (0)

#define yy_scan_literal()                                                                \
    do {                                                                                 \
        char c, prev = '\0';                                                             \
        while ((c = input(yyscanner)) != '\0') {                                         \
            if (c == '"' && prev != '\\') {                                              \
                src_pos_update_col(yylloc, 1);                                           \
                break;                                                                   \
            }                                                                            \
            if (c == '\n') {                                                             \
                strbuf_reset(YY_BUF);                                                    \
                unput(c);                                                                \
                yy_error(ERROR_UNTERMINATED_STRING);                                     \
                break;                                                                   \
            }                                                                            \
            else {                                                                       \
                src_pos_update_col(yylloc, 1);                                           \
                strbuf_ncat(YY_BUF, &c, 1);                                              \
            }                                                                            \
            prev = c;                                                                    \
        }                                                                                \
        if (c == '\0') {                                                                 \
            yy_error(ERROR_UNTERMINATED_STRING);                                         \
            yyterminate();                                                               \
        }                                                                                \
    } while (0)

#define yy_return_strbuf(tok)                                                            \
    do {                                                                                 \
        yylval->str = xstrndup(strbuf_str(YY_BUF), strbuf_size(YY_BUF));                 \
        strbuf_reset(YY_BUF);                                                            \
        return (tok);                                                                    \
    } while (0)

#define yy_return_text(tok)                                                              \
    do {                                                                                 \
        src_pos_update_first(yylloc);                                                    \
        src_pos_update_col(yylloc, yyleng);                                              \
        yylval->str = xstrndup(yytext, yyleng);                                          \
        return (tok);                                                                    \
    } while (0)

#define yy_return(tok)                                                                   \
    do {                                                                                 \
        src_pos_update_first(yylloc);                                                    \
        src_pos_update_col(yylloc, yyleng);                                              \
        return (tok);                                                                    \
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
E                           ([eE][+-]?{D}+)

integer                     ([1-9]{D}*)
octal                       (0[0-7]*)
hexa                        (0[xX]{H}+)
float                       (({D}+{E})|({D}*"."{D}+|{D}+"."{D}*){E}?)

id                          {L}({L}|{D})*

%%

{ws}+                       src_pos_update_col(yylloc, yyleng);
{nl}                        src_pos_update_line(yylloc);

"//"[^\r\n]*                src_pos_update_col(yylloc, yyleng);

"/*" {
    src_pos_update_first(yylloc);
    src_pos_update_col(yylloc, yyleng);
    yy_scan_comment();
}

"\"" {
    src_pos_update_first(yylloc);
    src_pos_update_col(yylloc, yyleng);
    yy_scan_literal();
    yy_return_strbuf(L_STR);
}

{integer}                   yy_return_text(L_INT);
{octal}                     yy_return_text(L_OCTAL);
{hexa}                      yy_return_text(L_HEX);
{float}                     yy_return_text(L_FLOAT);

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
"enum"                      yy_return(K_ENUM);
"false"                     yy_return(K_FALSE);
"float"                     yy_return(K_FLOAT);
"for"                       yy_return(K_FOR);
"func"                      yy_return(K_FUNC);
"goto"                      yy_return(K_GOTO);
"if"                        yy_return(K_IF);
"implements"                yy_return(K_IMPLEMENTS);
"in"                        yy_return(K_IN);
"int"                       yy_return(K_INT);
"int16"                     yy_return(K_INT16);
"int32"                     yy_return(K_INT32);
"int64"                     yy_return(K_INT64);
"int8"                      yy_return(K_INT8);
"interface"                 yy_return(K_INTERFACE);
"map"                       yy_return(K_MAP);
"new"                       yy_return(K_NEW);
"null"                      yy_return(K_NULL);
"payable"                   yy_return(K_PAYABLE);
"public"                    yy_return(K_PUBLIC);
"readonly"                  yy_return(K_READONLY);
"return"                    yy_return(K_RETURN);
"string"                    yy_return(K_STRING);
"struct"                    yy_return(K_STRUCT);
"switch"                    yy_return(K_SWITCH);
"true"                      yy_return(K_TRUE);
"type"                      yy_return(K_TYPE);
"uint"                      yy_return(K_UINT);
"uint16"                    yy_return(K_UINT16);
"uint32"                    yy_return(K_UINT32);
"uint64"                    yy_return(K_UINT64);
"uint8"                     yy_return(K_UINT8);

"create"|"CREATE"           yy_return(K_CREATE);
"delete"|"DELETE"           yy_return(K_DELETE);
"drop"|"DROP"               yy_return(K_DROP);
"index"|"INDEX"             yy_return(K_INDEX);
"insert"|"INSERT"           yy_return(K_INSERT);
"select"|"SELECT"           yy_return(K_SELECT);
"table"|"TABLE"             yy_return(K_TABLE);
"update"|"UPDATE"           yy_return(K_UPDATE);

"#file"                     BEGIN(DIRECTIVE);
<DIRECTIVE>{ws}+
<DIRECTIVE>\"[^\"\n\r]+\"   src_pos_set_path(yylloc, xstrndup(yytext + 1, yyleng - 2));
<DIRECTIVE>{D}+" "{D}+ {
    int line = atoi(strtok(xstrndup(yytext, yyleng), " "));
    int offset = atoi(strtok(NULL, " "));

    src_pos_set_pos(yylloc, line, 1, offset);
    src_pos_update_first(yylloc);
}
<DIRECTIVE>{nl}             BEGIN(INITIAL);

{id}                        yy_return_text(ID);

. {
    src_pos_update_first(yylloc);
    src_pos_update_col(yylloc, yyleng);
    yy_error(ERROR_UNKNOWN_CHAR);
    yy_recover();
}

<<EOF>>                     yyterminate();

%%

int
yylex(YYSTYPE *yylval, YYLTYPE *yylloc, void *yyscanner)
{
    int token;
    parse_t *parse = yyget_extra(yyscanner);

    if (parse->adj_token != 0) {
        token = parse->adj_token;
        *yylloc = parse->adj_pos;
        parse->adj_token = 0;
        return token;
    }

    return yyscan(yylval, yylloc, yyscanner);
}

void
yylex_set_token(void *yyscanner, int token, YYLTYPE *yylloc)
{
    parse_t *parse = yyget_extra(yyscanner);

    parse->adj_token = token;
    parse->adj_pos = *yylloc;
}

/* end of scanner.l */
