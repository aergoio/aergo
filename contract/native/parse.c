/**
 * @file    parse.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "iobuf.h"

#include "parse.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern void yyset_extra(parse_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(parse_t *, void *);
extern int yydebug;

static void
parse_init(parse_t *parse, flag_t flag, iobuf_t *src, ast_t *ast)
{
    char *path = iobuf_path(src);

    parse->path = path;
    parse->flag = flag;

    parse->src = iobuf_str(src);
    parse->len = iobuf_size(src);
    parse->pos = 0;

    parse->ast = ast;

    parse->adj_token = 0;
    src_pos_init(&parse->adj_pos, parse->src, path);

    strbuf_init(&parse->buf);
    vector_init(&parse->labels);
}

void
parse(iobuf_t *src, flag_t flag, ast_t *ast)
{
    parse_t parse;
    void *scanner;

    parse_init(&parse, flag, src, ast);
    yylex_init(&scanner);

    yyset_extra(&parse, scanner);

    if (is_flag_on(flag, FLAG_DUMP_LEX))
        yyset_debug(1, scanner);

    if (is_flag_on(flag, FLAG_DUMP_YACC))
        yydebug = 1;

    yyparse(&parse, scanner);

    yylex_destroy(scanner);
}

/* end of parse.c */
