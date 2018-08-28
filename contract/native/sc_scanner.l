%{
/**
 *  @file   sc_scanner.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "sc_common.h"

#include "sc_throw.h"
#include "sc_util.h"
#include "sc_parser.h"

#define YY_DECL                                                                \
    int sc_yylex(YYSTYPE *lval, YYLTYPE *lloc, yyscan_t yyscanner)

#undef  YY_INPUT
#define YY_INPUT(buf, result, max)                                             \
    do {                                                                       \
        result = sc_fread(yyin, max, buf);                                     \
    } while (0)

#define YY_EXTRA_TYPE       sc_lex_t *

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

#define yy_return(tok)                                                         \
    do {                                                                       \
        yy_update_column();                                                    \
        return (tok);                                                          \
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

lparen              "("
rparen              ")"
lbrace              "{"
rbrace              "}"
lbracket            "["
rbracket            "]"
dot                 "."
comma               ","
colon               ":"
semicolon           ";"
question            "?"
plus                "+"
minus               "-"
asterisk            "*"
slash               "/"
percent             "%"
equal               "="
dquote              "\""
sign                ({plus}|{minus})

cmp_eq              "=="
cmp_ne              "!="
cmp_lt              "<"
cmp_gt              ">"
cmp_le              "<="
cmp_ge              ">="

D                   [0-9]
L                   [a-zA-Z\x80-\xFF_]
IS                  (u|U)

integer             {D}+{IS}?
float               ({D}*"."{D}+|{D}+"."{D}*)

identifier          {L}({L}|{D})*

%%

{whitespace}+               yy_update_column();
{newline}                   yy_update_line();

"//"[^\r\n]*                yy_update_column();

"/*"                        { BEGIN(C_COMMENT); yy_update_column(); }
<C_COMMENT>"*/"             { BEGIN(INITIAL); yy_update_column(); }
<C_COMMENT>[^*\r\n]+        yy_update_column();
<C_COMMENT>"*"              yy_update_column();
<C_COMMENT>{newline}        yy_update_line();
<C_COMMENT><<EOF>> {
    yy_error(ERROR_SCAN_UNTERMINATED_COMMENT);
    yyterminate();
}

\"(\\.|[^\\"\n\r])*\"       yy_return(L_STR); 
    /*
"\"" {
    BEGIN(D_QUOTE);
    yy_update_column();
    yyextra->buf[0] = '\0';
    yyextra->offset = 0;
}
<D_QUOTE>[^"\\\r\n]+        { yy_update_column(); yy_strcat(); }
<D_QUOTE>"\"" {
    BEGIN(INITIAL);
    yy_update_column();
    fprintf(stderr, "result: %s\n", yyextra->buf);
    //return L_STR;
}
<D_QUOTE>"\\"{whitespace}*{newline} {
    yy_update_column();
    yy_update_line();
}
<D_QUOTE>"\\"               { yy_update_column(); yy_strcat(); }
<D_QUOTE>{newline} {
    yy_error(ERROR_SCAN_UNTERMINATED_STRING);
    yy_update_line();
}
<D_QUOTE><<EOF>> {
    yy_error(ERROR_SCAN_UNTERMINATED_STRING);
    yyterminate();
}
    */

{integer}           yy_return(L_INT);
{float}             yy_return(L_FLOAT);

"+="                yy_return(OP_ADD_ASSIGN);
"-="                yy_return(OP_SUB_ASSIGN);
"*="                yy_return(OP_MUL_ASSIGN);
"/="                yy_return(OP_DIV_ASSIGN);
"%="                yy_return(OP_MOD_ASSIGN);
"&="                yy_return(OP_AND_ASSIGN);
"^="                yy_return(OP_XOR_ASSIGN);
"|="                yy_return(OP_OR_ASSIGN);
">>"                yy_return(OP_RSHIFT);
"<<"                yy_return(OP_LSHIFT);
"++"                yy_return(OP_INC);
"--"                yy_return(OP_DEC);
"&&"                yy_return(OP_AND);
"||"                yy_return(OP_OR);
"<="                yy_return(OP_LE);
">="                yy_return(OP_GE);
"=="                yy_return(OP_EQ);
"!="                yy_return(OP_NE);

"{"                 yy_return('{');
"}"                 yy_return('}');
"("                 yy_return('(');
")"                 yy_return(')');
"["                 yy_return('[');
"]"                 yy_return(']');
";"                 yy_return(';');
":"                 yy_return(':');
","                 yy_return(',');
"?"                 yy_return('?');
"."                 yy_return('.');
"+"                 yy_return('+');
"-"                 yy_return('-');
"*"                 yy_return('*');
"/"                 yy_return('/');
"%"                 yy_return('%');
"<"                 yy_return('<');
">"                 yy_return('>');
"&"                 yy_return('&');
"|"                 yy_return('|');
"="                 yy_return('=');

"break"             yy_return(K_BREAK);
"byte"              yy_return(K_BYTE);
"case"              yy_return(K_CASE);
"const"             yy_return(K_CONST);
"continue"          yy_return(K_CONTINUE);
"contract"          yy_return(K_CONTRACT);
"double"            yy_return(K_DOUBLE);
"else"              yy_return(K_ELSE);
"float"             yy_return(K_FLOAT);
"for"               yy_return(K_FOR);
"if"                yy_return(K_IF);
"int"               yy_return(K_INT);
"int16"             yy_return(K_INT16);
"int32"             yy_return(K_INT32);
"int64"             yy_return(K_INT64);
"int8"              yy_return(K_INT8);
"new"               yy_return(K_NEW);
"private"           yy_return(K_PRIVATE);
"public"            yy_return(K_PUBLIC);
"return"            yy_return(K_RETURN);
"string"            yy_return(K_STRING);
"struct"            yy_return(K_STRUCT);
"switch"            yy_return(K_SWITCH);
"uint"              yy_return(K_UINT);
"uint16"            yy_return(K_UINT16);
"uint32"            yy_return(K_UINT32);
"uint64"            yy_return(K_UINT64);
"uint8"             yy_return(K_UINT8);
"while"             yy_return(K_WHILE);

    /*
"commit"            yy_return(K_COMMIT);
"delete"            yy_return(K_DELETE);
"insert"            yy_return(K_INSERT);
"rollback"          yy_return(K_ROLLBACK);
"select"            yy_return(K_SELECT);
    */

{identifier} {
    lval->str = malloc(yyleng + 1);
    sc_strcpy(lval->str, yytext);
    yy_return(ID);
}

. {
    yy_error(ERROR_SCAN_UNKNOWN_CHAR);
    yy_update_column();
    yy_recover();
}

<<EOF>>                         yyterminate();

%%

/* end of sc_scanner.l */
