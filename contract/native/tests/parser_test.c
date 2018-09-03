/**
 * @file    parser_test.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"
#include <dirent.h>
#include <ctype.h>
#include <sys/stat.h>

#include "strbuf.h"
#include "parser.h"

#define PREFIX              "tc_parser_"

#define TAG_TITLE           "@Test"
#define TAG_ERROR           "@Error"

static char *
trim_str(char *str)
{
    int i;
    int str_len = strlen(str);
    char *ptr = str;

    for (i = 0; i < str_len; i++) {
        if (!isspace(str[i]))
            break;

        ptr++;
    }

    str_len = strlen(ptr);

    for (i = str_len - 1; i >= 0; i--) {
        if (!isspace(ptr[i]))
            break;

        ptr[i] = '\0';
    }

    return ptr;
}

static int
get_errcode(char *str)
{
    int i;

    for (i = 0; i < ERROR_MAX; i++) {
        if (strcmp(error_text(i), str) == 0)
            return i;
    }
    ASSERT(!"invalid errcode");
}

static void
run_test(char *title, ec_t ex, char *src, int len)
{
    ec_t ac;

    printf("    * "ANSI_WHITE"%s"ANSI_NONE"... ", title);

    ac = parse(src, len, OPT_TEST);
    if (ex == ac) {
        printf(ANSI_GREEN"ok\n"ANSI_NONE);
    }
    else {
        printf(ANSI_RED"fail\n"ANSI_NONE);

        if (ex == NO_ERROR)
            error_dump();

        printf(ANSI_RED"fatal:"ANSI_NONE" expected <%s> but, actually <%s>\n",
               error_text(ex), error_text(ac));

        exit(EXIT_FAILURE);
    }

    error_clear();
}

static void
read_test(char *file)
{
    int line = 1;
    FILE *fp;
    char title[128];
    ec_t ec = NO_ERROR;
    strbuf_t sb;
    char buf[8192];

    strbuf_init(&sb);

    fp = fopen(file, "r");
    if (fp == NULL)
        FATAL(ERROR_FILE_OPEN_FAILED, file, strerror(errno));

    printf("  Checking %s...\n", file);

    while (fgets(buf, sizeof(buf), fp) != NULL) {
        if (strncasecmp(buf, TAG_TITLE, strlen(TAG_TITLE)) == 0) {
            if (!strbuf_empty(&sb)) {
                run_test(title, ec, strbuf_text(&sb), strbuf_length(&sb));
                strbuf_reset(&sb);
                title[0] = '\0';
                ec = NO_ERROR;
            }

            strcpy(title, trim_str(buf + strlen(TAG_TITLE)));
        }
        else if (strncasecmp(buf, TAG_ERROR, strlen(TAG_ERROR)) == 0) {
            ec = get_errcode(trim_str(buf + strlen(TAG_ERROR)));
        }
        else {
            strbuf_append(&sb, buf, strlen(buf));
        }
        line++;
    }

    if (!strbuf_empty(&sb))
        run_test(title, ec, strbuf_text(&sb), strbuf_length(&sb));
}

int
main(int argc, char **argv)
{
    char delim[81];
    DIR *dir;
    struct dirent *entry;
    struct stat st;

    dir = opendir(".");
    if (dir == NULL)
        FATAL(ERROR_DIR_OPEN_FAILED, ".", strerror(errno));

    memset(delim, '*', 80);
    delim[80] = '\0';

    printf("%s\n", delim);
    printf("Starting %s...\n", argv[0]);
    printf("%s\n", delim);

    while ((entry = readdir(dir)) != NULL) {
        stat(entry->d_name, &st);

        if (!S_ISREG(st.st_mode) ||
            strncmp(entry->d_name, PREFIX, strlen(PREFIX)) != 0)
            continue;

        read_test(entry->d_name);
    }

    closedir(dir);

    printf("%s\n", delim);
    printf("%s finished successfully\n", argv[0]);
    printf("%s\n", delim);

    return EXIT_SUCCESS;
}

/* end of parser_test.c */
