/**
 * @file    parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"
#include "prep.h"
#include "util.h"

#include "parser.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern void yyset_in(FILE *, void *);
extern void yyset_extra(yyparam_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(yyparam_t *, void *);
extern int yydebug;

int
parse(char *path, opt_t opt)
{
    yyparam_t param;
    void *scanner;

    yyparam_init(&param, path);
    yylex_init(&scanner);

    yyset_in(param.fp, scanner);
    yyset_extra(&param, scanner);

    if (opt_enabled(opt, OPT_DEBUG)) {
        yyset_debug(1, scanner);
        yydebug = 1;
    }

    yyparse(&param, scanner);
    yylex_destroy(scanner);

    if (!opt_enabled(opt, OPT_TEST))
        error_dump();

    remove(path);

    return error_last();
}

/* end of parser.c */
