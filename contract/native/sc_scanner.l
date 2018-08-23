%{
/**
 *  @file   sc_scanner.l
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "sc_common.h"

#include "sc_error.h"
#include "sc_util.h"
#include "sc_parser.h"

#define YY_EXTRA_TYPE       sc_yyextra_t *

#undef  YY_INPUT
#define YY_INPUT(buf, result, max)                                             \
    do {                                                                       \
        errno = 0;                                                             \
        while ((result = fread(buf, 1, max, yyin)) == 0 && ferror(yyin)) {     \
            if (errno != EINTR) {                                              \
                sc_fatal(ERROR_FILE_READ_FAILED, ferror(yyin));                \
                break;                                                         \
            }                                                                  \
            errno = 0;                                                         \
            clearerr(yyin);                                                    \
        }                                                                      \
    } while (0)

#define yy_update_line()                                                       \
    do {                                                                       \
        yyextra->line++;                                                       \
        yyextra->offset += yyextra->column;                                    \
        yyextra->column = 1;                                                   \
    } while (0)

#define yy_update_column()          (yyextra->column += yyleng)

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
        sc_fseek(fp, yyextra->offset);                                         \
        sc_fgets(fp, sizeof(line), line);                                      \
        snprintf(errloc, sizeof(errloc), "%s:%d:%d", yyextra->file,            \
                 yyextra->line, yyextra->column);                              \
        snprintf(errdesc, sizeof(errdesc), "%s\n%s", errmsg, line);            \
        for (i = 0; i < yyextra->column - 1; i++) {                            \
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

whitespace          ([ ]|\t|(\xE3\x80\x80))
newline             (\r|\n|\r\n)

cpp_comment         "//"
c_comment_start     "/*"
c_comment_end       "*/"

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
<C_COMMENT>"*"                  yy_update_column();
<C_COMMENT>{newline}            yy_update_line();
<C_COMMENT><<EOF>> {
    yy_error("unterminated comment");
    yyterminate();
}             

. {
    yy_error("unknown character");
    yy_update_column();
    yy_recover();
}

<<EOF>>                         yyterminate();

%%
