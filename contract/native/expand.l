%{

/**
 *  @file   expand.l
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
        (result) = fread(buf, 1, max, yyin);                                   \
        if ((result) == 0 && !feof(yyin))                                      \
            FATAL(ERROR_FILE_READ_FAILED, strerror(errno));                    \
    } while (0)

#define YY_EXTRA_TYPE           prep_param_t *

#define yy_update_line()        yyextra->line++

#define yy_append()             strbuf_append(yyextra->res, yytext, yyleng)

static void mark_file(char *file, int line, strbuf_t *res);

%}

%option case-sensitive
%option never-interactive
%option debug
%option reentrant
%option noyylineno
%option noyywrap
%option prefix="prep_yy"

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
    mark_file(yytext, 1, yyextra->res);
    read_file(yytext, yyextra->res);
    mark_file(yyextra->file, yyextra->line + 1, yyextra->res);
}

.                           yy_append(); 

<<EOF>>                     yyterminate();

%%

static void
mark_file(char *file, int line, strbuf_t *res)
{
    char buf[PATH_MAX_LEN + 32];

    snprintf(buf, sizeof(buf), "#file \"%s\" %d\n", file, line);

    strbuf_append(res, buf, strlen(buf));
}

/* end of expand.l */
