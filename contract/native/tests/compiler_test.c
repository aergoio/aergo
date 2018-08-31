/**
 * @file    compiler_test.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"
#include <dirent.h>

#include "errors.h"
#include "strbuf.h"

#define TEST_DIR            "./TESTCASE"

static void
run_test(char *tc)
{
    strbuf_t sb;

    strbuf_init(&sb);
    read_file
}

int
main(int argc, char **argv)
{
    DIR *dir;
    struct dirent *entry;
    struct stat st;

    dir = opendir(TEST_DIR);
    if (dir == NULL)
        FATAL(ERROR_DIR_OPEN_FAILED, TEST_DIR, strerror(errno));

    printf("Starting compiler test...\n");

    while ((entry = readdir(dir)) != NULL) {
        stat(entry->d_name, &st);
        if (!S_ISREG(st.st_mode))
            continue;

        run_test(entry->d_name);
    }

    closedir(dir);

    return EXIT_SUCCESS;
}

/* end of compiler_test.c */
