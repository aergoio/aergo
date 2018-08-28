/**
 * @file    sc_parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "sc_common.h"

#include "sc_throw.h"
#include "sc_util.h"

#include "sc_parser.h"

extern int sc_yylex_init(void **);
extern int sc_yylex_destroy(void *);

extern void sc_yyset_in(FILE *, void *);
extern void sc_yyset_extra(void *, void *);
extern void sc_yyset_debug(int, void *);

extern int sc_yyparse(void *);
extern int sc_yydebug;

static void
sc_lex_init(sc_lex_t *lex, char *path)
{
    char *delim;

    lex->path = path;

    delim = strrchr(path, SC_PATH_DELIM);
    strcpy(lex->file, delim == NULL ? path : delim + 1);

    lex->errcnt = 0;

    lex->lloc.line = 1;
    lex->lloc.column = 1;
    lex->lloc.offset = 0;

    lex->offset = 0;
    lex->buf = malloc(SC_STR_MAX_LEN + 1);
    lex->buf[0] = '\0';
}

static void
sc_yacc_init(sc_yacc_t *yacc)
{
    sc_yylex_init(&yacc->scanner);
}

int
sc_parse(char *path)
{
    FILE *fp;
    sc_lex_t lex;
    sc_yacc_t yacc;

    fp = sc_fopen(path, "r");
    sc_assert(fp != NULL);

    sc_lex_init(&lex, path);
    sc_yacc_init(&yacc);

    sc_yyset_in(fp, yacc.scanner);
    sc_yyset_extra(&lex, yacc.scanner);

    // sc_yyset_debug(1, yacc.scanner);
    // sc_yydebug = 1;

    sc_yyparse(yacc.scanner);
    sc_yylex_destroy(yacc.scanner);

    sc_fclose(fp);

    if (lex.errcnt > 0) {
        sc_warn(ERROR_PARSE_FAILED);
        return RC_ERROR;
    }

    return RC_OK;
}

/* end of sc_parser.c */
