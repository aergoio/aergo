%{

/**
 *  @file   scanner.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "parser.h"

#define YY_DECL                                                                \
    int yylex(YYSTYPE *lval, YYLTYPE *lloc, yyscan_t yyscanner)

#undef  YY_INPUT
#define YY_INPUT(buf, result, max)                                             \
    do {                                                                       \
        (result) = min((max), yyextra->len - yyextra->pos);                    \
        if ((result) > 0) {                                                    \
            memcpy(buf, yyextra->src + yyextra->pos, (result));                \
            yyextra->pos += (result);                                          \
        }                                                                      \
    } while (0)

#define YY_EXTRA_TYPE           yyparam_t *

#define yy_update_last_line()                                                  \
    do {                                                                       \
        lloc->last.line++;                                                     \
        lloc->last.offset += lloc->last.col;                                   \
        lloc->last.col = 1;                                                    \
    } while (0)

#define yy_update_last_col()                                                   \
    do {                                                                       \
        lloc->last.col += yyleng;                                              \
    } while (0)

#define yy_update_first()       lloc->first = lloc->last

#define yy_text()               xstrndup(yytext, yyleng)
#define yy_literal()            xstrndup(yytext + 1, yyleng - 2)

#define yy_strcat()                                                            \
    do {                                                                       \
        if (yyextra->offset + yyleng >= STR_MAX_LEN) {                         \
            yy_error(ERROR_SCAN_STR_TOO_LONG);                                 \
            yy_update_last_col();                                              \
            yy_recover();                                                      \
        }                                                                      \
        strcat(yyextra->buf, yytext);                                          \
        yyextra->offset += yyleng;                                             \
    } while (0)

#define yy_error(ec)                                                           \
    do {                                                                       \
        yy_update_first();                                                     \
        ERROR(ec, yyextra->path, lloc->first.line, lloc->first.col,            \
              make_trace(yyextra->path, lloc));                                \
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
            yy_update_last_col();                                              \
        }                                                                      \
    } while (0)

#define yy_return(tok)                                                         \
    do {                                                                       \
        yy_update_first();                                                     \
        yy_update_last_col();                                                  \
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
%x  F_DIR

whitespace                  ([ ]|\t|(\xE3\x80\x80))
newline                     (\r|\n|\r\n)

D                           [0-9]
L                           [a-zA-Z\x80-\xFF_]
IS                          (u|U)

integer                     [+-]?{D}+{IS}?
float                       [+-]?({D}*"."{D}+|{D}+"."{D}*)

id                          {L}({L}|{D})*

%%

{whitespace}+               yy_update_last_col();
{newline}                   yy_update_last_line();

"//"[^\r\n]*                yy_update_last_col();

"/*"                        { BEGIN(C_COMMENT); yy_update_last_col(); }
<C_COMMENT>"*/"             { BEGIN(INITIAL); yy_update_last_col(); }
<C_COMMENT>[^*\r\n]+        yy_update_last_col();
<C_COMMENT>"*"              yy_update_last_col();
<C_COMMENT>{newline}        yy_update_last_line();
<C_COMMENT><<EOF>> {
    yy_error(ERROR_SCAN_UNTERMINATED_COMMENT);
    yyterminate();
}

\"(\\.|[^\\"\n\r])*\" {
    lval->str = yy_literal();
    yy_return(L_STR);
}
    /*
"\"" {
    BEGIN(D_QUOTE);
    yy_update_last_col();
    yyextra->buf[0] = '\0';
    yyextra->offset = 0;
}
<D_QUOTE>[^"\\\r\n]+        { yy_update_last_col(); yy_strcat(); }
<D_QUOTE>"\"" {
    BEGIN(INITIAL);
    yy_update_last_col();
    fprintf(stderr, "result: %s\n", yyextra->buf);
    //return L_STR;
}
<D_QUOTE>"\\"{whitespace}*{newline} {
    yy_update_last_col();
    yy_update_last_line();
}
<D_QUOTE>"\\"               { yy_update_last_col(); yy_strcat(); }
<D_QUOTE>{newline} {
    yy_error(ERROR_SCAN_UNTERMINATED_STRING);
    yy_update_last_line();
}
<D_QUOTE><<EOF>> {
    yy_error(ERROR_SCAN_UNTERMINATED_STRING);
    yyterminate();
}
    */

{integer}                   yy_return(L_INT);
{float}                     yy_return(L_FLOAT);

"+="                        yy_return(OP_ADD_ASSIGN);
"-="                        yy_return(OP_SUB_ASSIGN);
"*="                        yy_return(OP_MUL_ASSIGN);
"/="                        yy_return(OP_DIV_ASSIGN);
"%="                        yy_return(OP_MOD_ASSIGN);
"&="                        yy_return(OP_AND_ASSIGN);
"^="                        yy_return(OP_XOR_ASSIGN);
"|="                        yy_return(OP_OR_ASSIGN);
">>"                        yy_return(OP_RSHIFT);
"<<"                        yy_return(OP_LSHIFT);
"++"                        yy_return(OP_INC);
"--"                        yy_return(OP_DEC);
"&&"                        yy_return(OP_AND);
"||"                        yy_return(OP_OR);
"<="                        yy_return(OP_LE);
">="                        yy_return(OP_GE);
"=="                        yy_return(OP_EQ);
"!="                        yy_return(OP_NE);

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
"#"                         yy_return('#');

"account"                   yy_return(K_ACCOUNT);
"bool"                      yy_return(K_BOOL);
"break"                     yy_return(K_BREAK);
"byte"                      yy_return(K_BYTE);
"case"                      yy_return(K_CASE);
"const"                     yy_return(K_CONST);
"constructor"               yy_return(K_CONSTRUCTOR);
"continue"                  yy_return(K_CONTINUE);
"contract"                  yy_return(K_CONTRACT);
"default"                   yy_return(K_DEFAULT);
"double"                    yy_return(K_DOUBLE);
"else"                      yy_return(K_ELSE);
"false"                     yy_return(K_FALSE);
"file"                      yy_return(K_FILE);
"float"                     yy_return(K_FLOAT);
"for"                       yy_return(K_FOR);
"func"                      yy_return(K_FUNC);
"if"                        yy_return(K_IF);
"int"                       yy_return(K_INT);
"int16"                     yy_return(K_INT16);
"int32"                     yy_return(K_INT32);
"int64"                     yy_return(K_INT64);
"int8"                      yy_return(K_INT8);
"map"                       yy_return(K_MAP);
"new"                       yy_return(K_NEW);
"null"                      yy_return(K_NULL);
"payable"                   yy_return(K_PAYABLE);
"private"                   yy_return(K_PRIVATE);
"public"                    yy_return(K_PUBLIC);
"return"                    yy_return(K_RETURN);
"string"                    yy_return(K_STRING);
"struct"                    yy_return(K_STRUCT);
"switch"                    yy_return(K_SWITCH);
"true"                      yy_return(K_TRUE);
"uint"                      yy_return(K_UINT);
"uint16"                    yy_return(K_UINT16);
"uint32"                    yy_return(K_UINT32);
"uint64"                    yy_return(K_UINT64);
"uint8"                     yy_return(K_UINT8);
"while"                     yy_return(K_WHILE);

    /*
"commit"                    yy_return(K_COMMIT);
"delete"                    yy_return(K_DELETE);
"insert"                    yy_return(K_INSERT);
"rollback"                  yy_return(K_ROLLBACK);
"select"                    yy_return(K_SELECT);
    */

"#file"                     BEGIN(F_DIR);
<F_DIR>{whitespace}+
<F_DIR>\"[^\"\n\r]+\"       yyextra->path = yy_literal();
<F_DIR>{D}+" "{D}+ {
    lloc->last.line = atoi(strtok(yy_text(), " "));
    lloc->last.offset = atoi(strtok(NULL, " "));
}
<F_DIR>{newline}            BEGIN(INITIAL);

{id} {
    lval->str = yy_text();
    yy_return(ID);
}

. {
    yy_error(ERROR_SCAN_UNKNOWN_CHAR);
    yy_recover();
}

<<EOF>>                     yyterminate();

%%

/* end of scanner.l */
