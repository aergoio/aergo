/**
 *  @file   aergoscc.c
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "sc_common.h"

#include "sc_error.h"
#include "sc_parser.h"

static void
sc_print_help(void)
{
    printf("%s, Aergo smart contract compiler\n\n"
           "Usage: %s [options] file...\n"
           "Options:\n"
           "  --help        Display this information\n"
           "  --version     Display compiler version information\n\n"
           "Examples:\n"
           "  %s contract.sc\n",
           SC_EXECUTABLE, SC_EXECUTABLE, SC_EXECUTABLE);

    sc_exit(EXIT_SUCCESS);
}

static void
sc_print_version(void)
{
    printf("%s, Aergo smart contract compiler %d.%d.%d\n\n"
           "Copyright blah blah blah...\n",
           SC_EXECUTABLE, SC_VERSION_MAJOR, SC_VERSION_MINOR, SC_VERSION_PATCH);

    sc_exit(EXIT_SUCCESS);
}

static char *
sc_check_argv(int argc, char **argv)
{
    int i;

    if (argc <= 1)
        sc_print_help();

    for (i = 1; i < argc; i++) {
        if (*argv[i] != '-')
            break;

        if (strcmp(argv[i], "--help") == 0)
            sc_print_help();
        else if (strcmp(argv[i], "--version") == 0)
            sc_print_version();
        else
            sc_fatal(ERROR_INVALID_OPTION, argv[i]);
    }

    return argv[i];
}

int
main(int argc, char **argv)
{
    char *infile;

    infile = sc_check_argv(argc, argv);
    sc_assert(infile != NULL);

    sc_parse(infile);

    return EXIT_SUCCESS;
}
