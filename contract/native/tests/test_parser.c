/**
 * @file    test_parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"
#include <unistd.h>
#include <fcntl.h>

#include "parser.h"

#include "CuTest.h"

static int
test_parse(char *file)
{
    return parse(file);
}

void TestScanNormalComment(CuTest *tc)
{
    CuAssertIntEquals(tc, RC_OK, test_parse("case_scan_normal_comment"));
}

void TestScanUnterminatedComment(CuTest *tc)
{
    CuAssertIntEquals(tc, RC_ERROR, test_parse("case_scan_unterminated_comment"));
}

void TestScanUnknownCharacter(CuTest *tc)
{
    CuAssertIntEquals(tc, RC_ERROR, test_parse("case_scan_unknown_char"));
}

/* end of test_parser.c */
