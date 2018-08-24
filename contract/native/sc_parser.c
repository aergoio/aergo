/**
 * @file    sc_parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "sc_common.h"

#include "sc_error.h"
#include "sc_util.h"
#include "sc_scanner.yy.h"

#include "sc_parser.h"

extern int sc_yylex_init(void **);
extern int sc_yylex_destroy(void *);
extern int sc_yylex(void *);

extern void sc_yyset_in(FILE *, void *);
extern void sc_yyset_extra(void *, void *);

static void
sc_yyextra_init(sc_yyextra_t *data, char *path)
{
    char *delim = strrchr(path, SC_PATH_DELIM);

    data->path = path;
    strcpy(data->file, delim == NULL ? path : delim + 1);
    data->errcnt = 0;
    data->line = 1;
    data->column = 1;
    data->offset = 0;
}

int
sc_parse(char *path)
{
    FILE *fp;
    yyscan_t scanner;
    sc_yyextra_t data;

    fp = sc_fopen(path, "r");

    // TODO: check extension

    sc_yylex_init(&scanner);
    sc_yyextra_init(&data, path);

    sc_yyset_in(fp, scanner);
    sc_yyset_extra(&data, scanner);

    sc_yylex(scanner);
    sc_yylex_destroy(scanner);

    sc_fclose(fp);

    if (data.errcnt > 0) {
        sc_warn(ERROR_PARSE_FAILED);
        return RC_ERROR;
    }

    return RC_OK;
}

/* end of sc_parser.c */
