/**
 * @file    ascl_test.c
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

#define TAG_TITLE       "@test"
#define TAG_ERROR       "@error"
#define TAG_EXPORT      "@export"

typedef struct env_s {
    char *path;
    char *needle;

    flag_t flag;

    char title[PATH_MAX_LEN];
    ec_t ec;
    int ec_cnt;

    int total_cnt;
    int failed_cnt;

    stack_t exp;
} env_t;

static void
env_init(env_t *env)
{
    memset(env, 0x00, sizeof(env_t));

    env->flag = FLAG_TEST;
    env->ec_cnt = 0;
    stack_init(&env->exp);
}

static void
env_reset(env_t *env)
{
    env->title[0] = '\0';
    env->ec = NO_ERROR;
    env->ec_cnt = 0;

    if (flag_off(env->flag, FLAG_VERBOSE)) {
        char *item;

        while ((item = stack_pop(&env->exp)) != NULL) {
            unlink(item);
        }
    }
}

static void
print_results(env_t *env)
{
    int i;
    int ac_cnt = env->ec_cnt > 0 ? error_size() : 1;

    if (env->ec_cnt > 0 && env->ec_cnt != ac_cnt) {
        printf("[ "ANSI_RED"fail"ANSI_NONE" ]\n");

        if (ac_cnt > 0)
            error_print();

        printf("Expected: <%d errors>\n"
               "Actually: <"ANSI_YELLOW"%d errors"ANSI_NONE">\n",
               env->ec_cnt, ac_cnt);

        env->failed_cnt++;
        return;
    }

    for (i = 0; i < ac_cnt; i++) {
        ec_t ac = error_item(i);

        if (ac != env->ec) {
            printf("[ "ANSI_RED"fail"ANSI_NONE" ]\n");

            if (ac != NO_ERROR)
                error_print();

            printf("Expected: <%s>\n"
                   "Actually: <"ANSI_YELLOW"%s"ANSI_NONE">\n",
                   error_to_str(env->ec), error_to_str(ac));

            env->failed_cnt++;
            return;
        }
    }

    printf("  [ "ANSI_GREEN"ok"ANSI_NONE" ]\n");

    if (flag_on(env->flag, FLAG_VERBOSE))
        error_print();
}

static void
run_test(env_t *env, char *path)
{
    int i = 0;

    if (env->needle != NULL) {
        if ((env->needle[0] == '^' &&
             strncmp(env->path, env->needle + 1, strlen(env->needle + 1)) != 0 &&
             strncmp(env->title, env->needle + 1, strlen(env->needle + 1)) != 0) ||
            (env->needle[0] != '^' &&
             strstr(env->path, env->needle) == NULL &&
             strstr(env->title, env->needle) == NULL)) {
            unlink(path);
            env_reset(env);
            return;
        }
    }

    env->total_cnt++;

    printf("  + %-67s ", env->title);
    fflush(stdout);

    compile(path, env->flag);

    print_results(env);

    error_clear();

    unlink(path);
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
    env->path = path;

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
            strcpy(env->title, strtrim(buf + strlen(TAG_TITLE), "() \t\n\r"));

            snprintf(out_file, sizeof(out_file), "%s", env->title);
            out_fp = open_file(out_file, "w");
        }
        else if (strncasecmp(buf, TAG_ERROR, strlen(TAG_ERROR)) == 0) {
            char *args;

            ASSERT(exp_fp == NULL);
            ASSERT(env->title[0] != '\0');
            ASSERT(env->ec == NO_ERROR);

            offset += strlen(buf);
            args = strtrim(buf + strlen(TAG_ERROR), "() \t\n\r");

            if (strchr(args, ',') != NULL) {
                env->ec = error_to_code(strtrim(strtok(args, ","), " \t\n\r"));
                env->ec_cnt = atoi(strtrim(strtok(NULL, ","), " \t\n\r"));
            }
            else {
                env->ec = error_to_code(args);
            }
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

            exp_file = strtrim(xstrdup(buf) + strlen(TAG_EXPORT), "() \t\n\r");
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

        if (strcmp(argv[i], "-v") == 0 || strcmp(argv[i], "--verbose") == 0)
            flag_set(env->flag, FLAG_VERBOSE);
        else if (strcmp(argv[i], "-l") == 0 || strcmp(argv[i], "--lex-dump") == 0)
            flag_set(env->flag, FLAG_LEX_DUMP);
        else if (strcmp(argv[i], "-y") == 0 || strcmp(argv[i], "--yacc-dump") == 0)
            flag_set(env->flag, FLAG_YACC_DUMP);
        else if (strcmp(argv[i], "-w") == 0 || strcmp(argv[i], "--wat-dump") == 0)
            flag_set(env->flag, FLAG_WAT_DUMP);
        else
            FATAL(ERROR_INVALID_FLAG, argv[i]);
    }
}

int
main(int argc, char **argv)
{
#define LINE_MAX_SIZE   80
    int i;
    int file_cnt = 0;
    char *gopath;
    char path[PATH_MAX_LEN];
    char files[FILE_MAX_CNT][PATH_MAX_LEN];
    char delim[LINE_MAX_SIZE + 1];
    char buf[LINE_MAX_SIZE + 1];
    DIR *dir;
    struct dirent *entry;
    struct stat st;
    env_t env;

    gopath = getenv("GOPATH");
    ASSERT(gopath != NULL);

    strcpy(path, gopath);
    strtrim(path, "/");
    strcat(path, "/src/github.com/aergoio/aergo/contract/native/tests");

    dir = opendir(path);
    if (dir == NULL)
        FATAL(ERROR_DIR_IO, path, strerror(errno));

    env_init(&env);
    get_opt(&env, argc, argv);

    strset(delim, '*', sizeof(delim) - 1);

    printf("%s\n", delim);
    printf("* Starting %s...\n", FILENAME(argv[0]));
    printf("%s\n", delim);

    memset(files, 0x00, sizeof(files));

    while ((entry = readdir(dir)) != NULL) {
        if (!isdigit(entry->d_name[0]))
            continue;

        ASSERT(file_cnt < FILE_MAX_CNT);

        strcpy(files[file_cnt], path);
        strcat(files[file_cnt], "/");
        strcat(files[file_cnt], entry->d_name);

        stat(files[file_cnt], &st);
        if (!S_ISREG(st.st_mode))
            continue;

        file_cnt++;
    }

    qsort(files, file_cnt, PATH_MAX_LEN, (int (*)(const void *, const void *))&strcmp);

    for (i = 0; i < file_cnt; i++) {
        read_test(&env, files[i]);
    }

    closedir(dir);

    printf("%s\n", delim);
    if (env.failed_cnt > 0) {
        sprintf(buf, "[ "ANSI_RED"%d"ANSI_NONE" / "ANSI_RED"%d"ANSI_NONE" ]",
                env.total_cnt - env.failed_cnt, env.total_cnt);
        printf("%-66s %31s\n", "* Some tests failed with errors!!!", buf);
    }
    else {
        sprintf(buf, "[ "ANSI_GREEN"%d"ANSI_NONE" / "ANSI_GREEN"%d"ANSI_NONE" ]",
                env.total_cnt - env.failed_cnt, env.total_cnt);
        printf("%-66s %31s\n", "* All tests passed!!!", buf);
    }
    printf("%s\n", delim);

    return EXIT_SUCCESS;
}

/* end of ascl_test.c */
