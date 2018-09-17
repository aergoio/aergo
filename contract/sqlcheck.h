#ifndef _SQLCHECK_H
#define _SQLCHECK_H

int sqlcheck_is_permitted_sql(const char *sql);
int sqlcheck_is_readonly_sql(const char *sql);

#endif /* _SQLCHECK_H */