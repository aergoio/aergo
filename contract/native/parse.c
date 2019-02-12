/**
 * @file    parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "parse.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern void yyset_extra(parse_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(parse_t *, void *);
extern int yydebug;

static void
parse_init(parse_t *parse, flag_t flag, char *path)
{
    strbuf_t src;

    ASSERT(path != NULL);

    parse->path = path;
    parse->flag = flag;

    strbuf_init(&src);
    strbuf_load(&src, path);

    parse->src = strbuf_str(&src);
    parse->len = strbuf_size(&src);
    parse->pos = 0;

    parse->ast = NULL;

    parse->adj_token = 0;
    src_pos_init(&parse->adj_pos, parse->src, path);

    strbuf_init(&parse->buf);
    vector_init(&parse->labels);
}

void
parse(char *path, flag_t flag, ast_t **ast)
{
    parse_t parse;
    void *scanner;

    ASSERT(ast != NULL);

    parse_init(&parse, flag, path);
    yylex_init(&scanner);

    yyset_extra(&parse, scanner);

    if (is_flag_on(flag, FLAG_DUMP_LEX))
        yyset_debug(1, scanner);

    if (is_flag_on(flag, FLAG_DUMP_YACC))
        yydebug = 1;

    yyparse(&parse, scanner);

    yylex_destroy(scanner);

    *ast = parse.ast;
}

/* end of parser.c */
