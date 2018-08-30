/**
 * @file    test_parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"
#include <unistd.h>
#include <fcntl.h>

#include "preprocess.h"
#include "parser.h"

#include "CuTest.h"

static int
test_parse(CuTest *tc, char *file)
{
    FILE *fp;

    dup2(open("/dev/null", O_WRONLY), STDERR_FILENO);

    fp = preprocess(file);
    CuAssertTrue(tc, fp != NULL);

    return parse(fp);
}

void TestScanNormalComment(CuTest *tc)
{
    CuAssertIntEquals(tc, RC_OK, test_parse(tc, "case_scan_normal_comment"));
}

void TestScanUnterminatedComment(CuTest *tc)
{
    CuAssertIntEquals(tc, RC_ERROR, test_parse(tc, "case_scan_unterminated_comment"));
}

void TestScanUnknownCharacter(CuTest *tc)
{
    CuAssertIntEquals(tc, RC_ERROR, test_parse(tc, "case_scan_unknown_char"));
}

/* end of test_parser.c */
