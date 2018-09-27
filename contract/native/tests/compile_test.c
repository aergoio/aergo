/**
 * @file    compile_test.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>

#include "strbuf.h"
#include "util.h"
#include "stack.h"
#include "compile.h"

#define FILE_MAX_CNT    100

#define TAG_TITLE       "@desc"
#define TAG_ERROR       "@error"
#define TAG_EXPORT      "@export"

typedef struct env_s {
    char *needle;
    flag_t flag;

    char title[PATH_MAX_LEN];
    ec_t ec;

    stack_t exp;
} env_t;

bool is_failed = false;

static void
env_init(env_t *env)
{
    memset(env, 0x00, sizeof(env_t));

    env->flag = FLAG_SILENT;
    stack_init(&env->exp);
}

static void
env_reset(env_t *env)
{
    env->title[0] = '\0';
    env->ec = NO_ERROR;

    if (flag_off(env->flag, FLAG_VERBOSE)) {
        char *item;

        while ((item = stack_pop(&env->exp)) != NULL) {
            unlink(item);
        }
    }
}

static void
run_test(env_t *env, char *path)
{
    ec_t ac;

    if (env->needle != NULL && strstr(env->title, env->needle) == NULL) {
        unlink(path);
        env_reset(env);
        return;
    }

    printf("  + %-67s ", env->title);
    fflush(stdout);

    compile(path, env->flag);

    ac = error_first();
    if (ac == env->ec) {
        printf("  [ "ANSI_GREEN"ok"ANSI_NONE" ]\n");

        if (flag_on(env->flag, FLAG_VERBOSE))
            error_dump();

        unlink(path);
    }
    else {
        printf("[ "ANSI_RED"fail"ANSI_NONE" ]\n");

        if (ac != NO_ERROR)
            error_dump();

        printf("Expected: <%s>\nActually: <"ANSI_YELLOW"%s"ANSI_NONE">\n",
               error_to_string(env->ec), error_to_string(ac));

        is_failed = true;
    }

    error_clear();

    env_reset(env);
}

static void
read_test(env_t *env, char *path)
{
    int line = 1;
    int offset = 0;
    char buf[1024];
    char out_file[PATH_MAX_LEN];
    FILE *in_fp = open_file(path, "r");
    FILE *out_fp = NULL;
    FILE *exp_fp = NULL;

    printf("Checking %s...\n", FILENAME(path));

    while (fgets(buf, sizeof(buf), in_fp) != NULL) {
        if (strncasecmp(buf, TAG_TITLE, strlen(TAG_TITLE)) == 0) {
            if (exp_fp != NULL)
                close_file(exp_fp);

            if (out_fp != NULL) {
                close_file(out_fp);
                run_test(env, out_file);
            }

            exp_fp = NULL;

            offset += strlen(buf);
            strcpy(env->title, trim_str(buf + strlen(TAG_TITLE)));

            snprintf(out_file, sizeof(out_file), "%s", env->title);
            out_fp = open_file(out_file, "w");
        }
        else if (strncasecmp(buf, TAG_ERROR, strlen(TAG_ERROR)) == 0) {
            ASSERT(exp_fp == NULL);
            ASSERT(env->title[0] != '\0');
            ASSERT(env->ec == NO_ERROR);

            offset += strlen(buf);
            env->ec = error_to_code(trim_str(buf + strlen(TAG_ERROR)));
        }
        else if (strncasecmp(buf, TAG_EXPORT, strlen(TAG_EXPORT)) == 0) {
            char *exp_file;

            if (exp_fp != NULL)
                close_file(exp_fp);

            if (out_fp != NULL) {
                close_file(out_fp);
                run_test(env, out_file);
            }

            out_fp = NULL;

            exp_file = trim_str(xstrdup(buf) + strlen(TAG_EXPORT));
            exp_fp = open_file(exp_file, "w");

            offset += strlen(buf);
            stack_push(&env->exp, exp_file);
        }
        else {
            if (exp_fp != NULL)
                fwrite(buf, 1, strlen(buf), exp_fp);
            else
                fwrite(buf, 1, strlen(buf), out_fp);

            offset += strlen(buf);
        }
        line++;
    }

    if (out_fp != NULL) {
        close_file(out_fp);
        run_test(env, out_file);
    }
}

static void
get_opt(env_t *env, int argc, char **argv)
{
    int i;

    for (i = 1; i < argc; i++) {
        if (*argv[i] != '-') {
            env->needle = argv[i];
            continue;
        }

        if (strcmp(argv[i], "--verbose") == 0)
            flag_set(env->flag, FLAG_VERBOSE);
        else if (strcmp(argv[i], "--lex-dump") == 0)
            flag_set(env->flag, FLAG_LEX_DUMP);
        else if (strcmp(argv[i], "--yacc-dump") == 0)
            flag_set(env->flag, FLAG_YACC_DUMP);
        else
            FATAL(ERROR_INVALID_FLAG, argv[i]);
    }
}

int
main(int argc, char **argv)
{
    int i;
    int file_cnt = 0;
    char files[FILE_MAX_CNT][PATH_MAX_LEN];
    char delim[81];
    DIR *dir;
    struct dirent *entry;
    struct stat st;
    env_t env;
    flag_t flag = FLAG_NONE;
    
    dir = opendir(".");
    if (dir == NULL)
        FATAL(ERROR_DIR_IO, ".", strerror(errno));

    env_init(&env);
    get_opt(&env, argc, argv);

    memset(delim, '*', 80);
    delim[80] = '\0';

    printf("%s\n", delim);
    printf("* Starting %s...\n", FILENAME(argv[0]));
    printf("%s\n", delim);

    memset(files, 0x00, sizeof(files));

    while ((entry = readdir(dir)) != NULL) {
        stat(entry->d_name, &st);

        if (!S_ISREG(st.st_mode) || !isdigit(entry->d_name[0]))
            continue;

        ASSERT(file_cnt < FILE_MAX_CNT);
        strcpy(files[file_cnt++], entry->d_name);
    }

    qsort(files, file_cnt, PATH_MAX_LEN, 
          (int (*)(const void *, const void *))&strcmp);

    for (i = 0; i < file_cnt; i++) {
        read_test(&env, files[i]);
    }

    closedir(dir);

    printf("%s\n", delim);
    printf("* Finished %s %s\n", FILENAME(argv[0]),
           is_failed ? "with error(s)" : "successfully");
    printf("%s\n", delim);

    return EXIT_SUCCESS;
}

/* end of compile_test.c */
