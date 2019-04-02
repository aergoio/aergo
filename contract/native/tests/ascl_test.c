/**
 * @file    ascl_test.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>

#include "WAVM/ASCL/ASCLVM.h"

#include "strbuf.h"
#include "util.h"
#include "stack.h"
#include "prep.h"
#include "parse.h"
#include "check.h"
#include "trans.h"
#include "gen.h"

#define FILE_MAX_CNT    100

#define TAG_TITLE       "@test"
#define TAG_ERROR       "@error"
#define TAG_EXPORT      "@export"
#define TAG_RUN         "@run"

typedef struct env_s {
    char *path;

    bool fits_l;
    bool fits_r;
    char *needle;

    flag_t flag;

    char title[PATH_MAX_LEN + 1];
    char module[PATH_MAX_LEN + 1];
    char func[PATH_MAX_LEN + 1];
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

    env->flag.stack_size = UINT16_MAX + 1;

    stack_init(&env->exp);
}

static void
env_reset(env_t *env)
{
    env->title[0] = '\0';
    env->module[0] = '\0';
    env->func[0] = '\0';
    env->ec = NO_ERROR;
    env->ec_cnt = 0;

    if (is_flag_off(env->flag, FLAG_VERBOSE)) {
        char *item;

        while ((item = stack_pop(&env->exp)) != NULL) {
            unlink(item);
        }
    }

    error_clear();
}

static bool
match_title(env_t *env)
{
    int size;
    char *needle = env->needle;

    if (needle == NULL || needle[0] == '\0')
        return true;

    size = strlen(needle);

    if (env->fits_l) {
        if (env->fits_r) {
            if (strcmp(env->path, needle) == 0 || strcmp(env->title, needle) == 0)
                return true;
        }
        else if (strncmp(env->path, needle, size) == 0 ||
                 strncmp(env->title, needle, size) == 0) {
            return true;
        }
    }
    else if (env->fits_r) {
        if (((int)strlen(env->path) >= size &&
             strcmp(env->path + strlen(env->path) - size, needle) == 0) ||
            ((int)strlen(env->title) >= size &&
             strcmp(env->title + strlen(env->title) - size, needle) == 0))
            return true;
    }
    else if (strstr(env->path, needle) != NULL || strstr(env->title, needle) != NULL) {
        return true;
    }

    return false;
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

    if (is_flag_on(env->flag, FLAG_VERBOSE))
        error_print();
}

static void
run_test(env_t *env, char *path)
{
    iobuf_t src;
    ast_t *ast = ast_new();
    ir_t *ir = ir_new();

    env->total_cnt++;

    printf("  + %-67s ", env->title);
    fflush(stdout);

    iobuf_init(&src, path);
    iobuf_load(&src);

    prep(&src, env->flag, ast);
    parse(&src, env->flag, ast);

    check(ast, env->flag);
    trans(ast, env->flag, ir);

    gen(ir, env->flag, path);

    if (!has_error() && env->module[0] != '\0') {
        char *argv[2] = { "0", NULL };
        char wasm[PATH_MAX_LEN + 6];

        snprintf(wasm, sizeof(wasm), "%s.wasm", env->module);

        vm_run(wasm, env->func, argv);
    }

    print_results(env);

    unlink(path);
    env_reset(env);
}

static void
read_test(env_t *env, char *path)
{
    char buf[1024];
    char out_file[PATH_MAX_LEN + 1];
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

            while (true) {
                strcpy(env->title, strtrim(buf + strlen(TAG_TITLE), "() \t\n\r"));

                if (match_title(env))
                    break;

                while (fgets(buf, sizeof(buf), in_fp) != NULL) {
                    if (strncasecmp(buf, TAG_TITLE, strlen(TAG_TITLE)) == 0)
                        break;
                }

                if (feof(in_fp))
                    return;
            }

            snprintf(out_file, sizeof(out_file), "%s", env->title);
            out_fp = open_file(out_file, "w");
        }
        else if (strncasecmp(buf, TAG_ERROR, strlen(TAG_ERROR)) == 0) {
            char *args;

            ASSERT(exp_fp == NULL);
            ASSERT(env->title[0] != '\0');
            ASSERT(env->ec == NO_ERROR);

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

            stack_push(&env->exp, exp_file);
        }
        else if (strncasecmp(buf, TAG_RUN, strlen(TAG_RUN)) == 0) {
            char *ptr;
            char *qname = strtrim(buf + strlen(TAG_RUN), "() \t\n\r");

            ptr = strchr(qname, '.');
            if (ptr != NULL) {
                *ptr = '\0';
                strcpy(env->func, ptr + 1);
            }
            else {
                strcpy(env->func, qname);
            }
            strcpy(env->module, qname);
        }
        else {
            if (exp_fp != NULL)
                fwrite(buf, 1, strlen(buf), exp_fp);
            else
                fwrite(buf, 1, strlen(buf), out_fp);
        }
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
            if (argv[i][0] == '^') {
                env->needle = argv[i] + 1;
                env->fits_l = true;
            }
            else {
                env->needle = argv[i];
            }

            if (env->needle[(int)strlen(env->needle) - 1] == '$') {
                env->needle[(int)strlen(env->needle) - 1] = '\0';
                env->fits_r = true;
            }
            continue;
        }

        if (strcmp(argv[i], "-v") == 0 || strcmp(argv[i], "--verbose") == 0)
            flag_set(env->flag, FLAG_VERBOSE);
        else if (strcmp(argv[i], "-l") == 0 || strcmp(argv[i], "--print-lex") == 0)
            flag_set(env->flag, FLAG_DUMP_LEX);
        else if (strcmp(argv[i], "-y") == 0 || strcmp(argv[i], "--print-yacc") == 0)
            flag_set(env->flag, FLAG_DUMP_YACC);
        else if (strcmp(argv[i], "-w") == 0 || strcmp(argv[i], "--print-wat") == 0)
            flag_set(env->flag, FLAG_DUMP_WAT);
        else if (strcmp(argv[i], "-o") == 0 || strcmp(argv[i], "--optimize") == 0)
            env->flag.opt_lvl = 2;
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
    if (env.total_cnt == 0) {
        printf("%s\n", "* No tests found!!!");
    }
    else if (env.failed_cnt > 0) {
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
