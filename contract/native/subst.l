%{

/**
 *  @file   subst.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "strbuf.h"

#include "prep.h"

#define YY_NO_INPUT
#define YY_NO_UNPUT

#undef  YY_INPUT
#define YY_INPUT(buf, result, max)                                             \
    do {                                                                       \
        (result) = min((max), yyextra->len);                                   \
        if ((result) > 0) {                                                    \
            memcpy(buf, yyextra->src, (result));                               \
            yyextra->src += (result);                                          \
            yyextra->len -= (result);                                          \
        }                                                                      \
    } while (0)

#define YY_EXTRA_TYPE           subst_t *

#define yy_update_line()        yyextra->line++

#define yy_append()             strbuf_append(yyextra->res, yytext, yyleng)

%}

%option case-sensitive
%option never-interactive
%option debug
%option reentrant
%option noyylineno
%option noyywrap

%x  C_COMMENT
%x  IMPORT

whitespace      ([ ]|\t|(\xE3\x80\x80))
newline         (\r|\n|\r\n)

%%

{whitespace}+               yy_append();
{newline}                   { yy_update_line(); yy_append(); }

"//"[^\r\n]*                yy_append();

"/*"                        { BEGIN(C_COMMENT); yy_append(); }
<C_COMMENT>"*/"             { BEGIN(INITIAL); yy_append(); }
<C_COMMENT>[^*\r\n]+        yy_append();
<C_COMMENT>"*"              yy_append();
<C_COMMENT>{newline}        { yy_update_line(); yy_append(); }
<C_COMMENT><<EOF>>          yyterminate();

"\"(\\.|[^\\\"\n\r])*\""    yy_append();

"import"                    BEGIN(IMPORT);
<IMPORT>{whitespace}+
<IMPORT>{newline}           { BEGIN(INITIAL); yy_update_line(); }
<IMPORT>"\"[^\"\n\r]+\"" {
    append_directive(yytext, 1, yyextra->res);
    read_file(yytext, yyextra->res);
    append_directive(yyextra->path, yyextra->line + 1, yyextra->res);
}

.                           yy_append(); 

<<EOF>>                     yyterminate();

%%

/* end of subst.l */
