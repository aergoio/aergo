/**
 * @file    parser_test.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"
#include <dirent.h>
#include <ctype.h>
#include <sys/stat.h>
#include <unistd.h>

#include "strbuf.h"
#include "util.h"
#include "prep.h"
#include "parser.h"

#define PREFIX              "tc_parser_"

#define TAG_TITLE           "@desc"
#define TAG_ERROR           "@return"

extern void mark_file(char *path, int line, int offset, strbuf_t *out);

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
run_test(char *title, ec_t ex, char *path, opt_t opt, strbuf_t *sb)
{
    ec_t ac;

    printf("    * %s... ", title);

    ac = parse(path, opt, sb);
    if (ex == ac) {
        printf(ANSI_GREEN"ok"ANSI_NONE"\n");
    }
    else {
        printf(ANSI_RED"fail"ANSI_NONE"\n");

        if (ex == NO_ERROR)
            error_dump();

        printf("Expected: <%s>\nActually: <"ANSI_YELLOW"%s"ANSI_NONE">\n",
               error_text(ex), error_text(ac));

        exit(EXIT_FAILURE);
    }

    error_clear();
}

static void
read_test(char *path, opt_t opt)
{
    int line = 1;
    int offset = 0;
    char title[128];
    ec_t ec = NO_ERROR;
    strbuf_t sb;
    char buf[1024];
    FILE *fp;

    fp = open_file(path, "r");

    strbuf_init(&sb);

    printf("  Checking %s...\n", path);

    while (fgets(buf, sizeof(buf), fp) != NULL) {
        if (strncasecmp(buf, TAG_TITLE, strlen(TAG_TITLE)) == 0) {
            if (!strbuf_empty(&sb)) {
                run_test(title, ec, path, opt, &sb);
                strbuf_reset(&sb);
                title[0] = '\0';
                ec = NO_ERROR;
            }

            offset += strlen(buf);
            strcpy(title, trim_str(buf + strlen(TAG_TITLE)));
        }
        else if (strncasecmp(buf, TAG_ERROR, strlen(TAG_ERROR)) == 0) {
            offset += strlen(buf);
            ec = get_errcode(trim_str(buf + strlen(TAG_ERROR)));
        }
        else {
            mark_file(path, line, offset, &sb);
            strbuf_append_str(&sb, buf, strlen(buf));
            offset += strlen(buf);
        }

        line++;
    }

    if (!strbuf_empty(&sb))
        run_test(title, ec, path, opt, &sb);
}

int
main(int argc, char **argv)
{
    char delim[81];
    opt_t opt = OPT_TEST;
    DIR *dir;
    struct dirent *entry;
    struct stat st;

    if (argc >= 2 && strcmp(argv[1], "--debug") == 0)
        opt_set(opt, OPT_DEBUG);

    dir = opendir(".");
    if (dir == NULL)
        FATAL(ERROR_DIRECTORY_IO_FAILED, ".", strerror(errno));

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

        read_test(entry->d_name, opt);
    }

    closedir(dir);

    printf("%s\n", delim);
    printf("%s finished successfully\n", argv[0]);
    printf("%s\n", delim);

    return EXIT_SUCCESS;
}

/* end of parser_test.c */
