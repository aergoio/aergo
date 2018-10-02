#include <ctype.h>
#include <string.h>
#include "_cgo_export.h"

#define KEYWORD_MINSIZE 4
#define KEYWORD_MAXSIZE 7
static char keyword[KEYWORD_MAXSIZE+1];

static int get_keyword(const char *sql)
{
    int in_bc = 0;
    int in_lc = 0;
    int spos = -1;
    int epos = 0;
    int l = strlen(sql);
    int i;

    for (i = 0; i < l; i++) {
        char c = sql[i];
        switch (c) {
        case ' ':
            /* fallthrough */
        case '\t':
            if (spos > -1) {
                epos = i;
                goto LOOP_END;
            }
            break;
        case '\n':
            if (in_lc) {
                in_lc = 0;
            }
            if (spos > -1) {
                epos = i;
                goto LOOP_END;
            }
            break;
        case '-':
            if (spos > -1) {
                epos = i;
                goto LOOP_END;
            }
            if (!in_bc && !in_lc && sql[i+1] == '-') {
                in_lc = 1;
                i++;
            }
            break;
        case '/':
            if (spos > -1) {
                epos = i;
                goto LOOP_END;
            }
            if (!in_lc && !in_bc && sql[i+1] == '*') {
                in_bc = 1;
                i++;
            }
            break;
        case '*':
            if (in_bc) {
                if (sql[i+1] == '/') {
                    in_bc = 0;
                    i++;
                }
            }
            break;
        default:
            if (!in_lc && !in_bc && spos == -1) {
                spos = i;
            }
        }
    }
    if (!in_bc && !in_lc) { /* EOF */
        epos = i;
    }

LOOP_END:
    if (spos > -1 && epos > spos) {
        int klen = epos - spos;
        if (klen >= KEYWORD_MINSIZE && klen <= KEYWORD_MAXSIZE) {
            strncpy(keyword, sql + spos, klen);
            keyword[klen] = '\0';
            i = 0;
            while (keyword[i]) {
                keyword[i] = toupper(keyword[i]);
                i++;
            }
            return epos;
        }
    }
    return -1;
}

int sqlcheck_is_permitted_sql(const char *sql)
{
    int end_offset = -1;

    end_offset = get_keyword(sql);
    if (end_offset > -1) {
        if (strncmp(keyword, "CREATE", 6) == 0) {
            end_offset = get_keyword(sql + end_offset);
            if (end_offset == -1) {
                return 0;
            }
            if (strncmp(keyword, "TABLE", 5) == 0 || strncmp(keyword, "INDEX", 5) == 0) {
                return 1;
            } else {
                return 0;
            }
        } else {
            return PermittedCmd(keyword);
        }
    }
    return 0;
}

int sqlcheck_is_readonly_sql(const char *sql)
{
    int end_offset = -1;

    end_offset = get_keyword(sql);
    if (end_offset > -1) {
        return strncmp(keyword, "SELECT", 6) == 0;
    }
    return 0;
}

