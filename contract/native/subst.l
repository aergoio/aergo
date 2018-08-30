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
        (result) = MIN((max), yyextra->len);                                   \
        if ((result) > 0) {                                                    \
            memcpy(buf, yyextra->src, (result));                               \
            yyextra->src += (result);                                          \
            yyextra->len -= (result);                                          \
        }                                                                      \
    } while (0)

#define YY_EXTRA_TYPE           subst_t *

#define YY_UPDATE_LINE()        yyextra->line++

#define YY_APPEND()             strbuf_append(yyextra->res, yytext, yyleng)

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

{whitespace}+               YY_APPEND();
{newline}                   { YY_UPDATE_LINE(); YY_APPEND(); }

"//"[^\r\n]*                YY_APPEND();

"/*"                        { BEGIN(C_COMMENT); YY_APPEND(); }
<C_COMMENT>"*/"             { BEGIN(INITIAL); YY_APPEND(); }
<C_COMMENT>[^*\r\n]+        YY_APPEND();
<C_COMMENT>"*"              YY_APPEND();
<C_COMMENT>{newline}        { YY_UPDATE_LINE(); YY_APPEND(); }
<C_COMMENT><<EOF>>          yyterminate();

"\"(\\.|[^\\\"\n\r])*\""    YY_APPEND();

"import"                    BEGIN(IMPORT);
<IMPORT>{whitespace}+
<IMPORT>{newline}           { BEGIN(INITIAL); YY_UPDATE_LINE(); }
<IMPORT>"\"[^\"\n\r]+\"" {
    append_directive(yyextra->res, yytext, 1);
    read_file(yytext, yyextra->res);
    append_directive(yyextra->res, yyextra->path, yyextra->line + 1);
}

.                           YY_APPEND(); 

<<EOF>>                     yyterminate();

%%

/* end of subst.l */
