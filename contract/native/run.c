/**
 * @file    run.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "run.h"

int
run(char *path, flag_t flag, char **argv)
{
    return vm_run(path, "main", argv);
}

/* end of run.c */
