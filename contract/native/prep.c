/**
 * @file    preprocess.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "strbuf.h"

#include "prep.h"

extern int prep_yylex_init(void **);
extern int prep_yylex_destroy(void *);

extern int prep_yylex(void *);

extern void prep_yyset_in(FILE *, void *);
extern void prep_yyset_extra(yyparam_t *, void *);
extern void prep_yyset_debug(int, void *);

int
preprocess(char *infile, char *outfile, opt_t opt)
{
    yyparam_t param;
    void *scanner;

    yyparam_init(&param, infile);
    prep_yylex_init(&scanner);

    mark_fpos(infile, 1, &param.buf);

    prep_yyset_in(param.fp, scanner);
    prep_yyset_extra(&param, scanner);

    if (opt_enabled(opt, OPT_DEBUG))
        prep_yyset_debug(1, scanner);

    prep_yylex(scanner);
    prep_yylex_destroy(scanner);

    if (!opt_enabled(opt, OPT_TEST))
        error_dump();

    if (error_empty()) {
        write_file(outfile, &param.buf);

        return NO_ERROR;
    }   

    return error_last();
}   

void
mark_fpos(char *path, int line, strbuf_t *sb)
{
    char buf[PATH_MAX_LEN + 32];

    snprintf(buf, sizeof(buf), "#file \"%s\" %d\n", path, line);

    strbuf_append(sb, buf, strlen(buf));
}

/* end of preprocess.c */
