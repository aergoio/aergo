/**
 *  @file   aergoscc.c
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "version.h"
#include "compile.h"
#include "util.h"

static void
print_help(void)
{
    printf("%s, Aergo smart contract compiler\n\n"
           "Usage: %s [options] file...\n"
           "Options:\n"
           "  --help        Display this information\n"
           "  --version     Display compiler version information\n\n"
           "Examples:\n"
           "  %s contract.sc\n",
           EXECUTABLE, EXECUTABLE, EXECUTABLE);

    exit(EXIT_SUCCESS);
}

static void
print_version(void)
{
    printf("%s, Aergo smart contract compiler %d.%d\n\n"
           "Copyright blah blah blah...\n",
           EXECUTABLE, MAJOR_VER, MINOR_VER);

    exit(EXIT_SUCCESS);
}

static char *
check_argv(int argc, char **argv, opt_t *opt)
{
    int i;

    if (argc <= 1)
        print_help();

    for (i = 1; i < argc; i++) {
        if (*argv[i] != '-')
            break;

        if (strcmp(argv[i], "--help") == 0)
            print_help();
        else if (strcmp(argv[i], "--version") == 0)
            print_version();
        else if (strcmp(argv[i], "--debug") == 0)
            opt_set(*opt, OPT_DEBUG);
        else
            FATAL(ERROR_INVALID_OPTION, argv[i]);
    }

    return argv[i];
}

int
main(int argc, char **argv)
{
    ec_t ec;
    char *infile;
    opt_t opt = OPT_NORMAL;

    infile = check_argv(argc, argv, &opt);
    ASSERT(infile != NULL);

    ec = compile(infile, opt);
    if (ec != NO_ERROR)
        return EXIT_FAILURE;

    return EXIT_SUCCESS;
}
