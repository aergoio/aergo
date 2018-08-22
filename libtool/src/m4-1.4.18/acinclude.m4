# Local additions to Autoconf macros.
# Copyright (C) 1992, 1994, 2004, 2006, 2008-2014, 2016 Free Software
# Foundation, Inc.
# Francois Pinard <pinard@iro.umontreal.ca>, 1992.
#
# This file is free software; the Free Software Foundation
# gives unlimited permission to copy and/or distribute it,
# with or without modifications, as long as this notice is preserved.

# Check if --with-dmalloc was given.

AC_DEFUN([M4_WITH_DMALLOC],
[AC_MSG_CHECKING([if malloc debugging is wanted])
AC_ARG_WITH([dmalloc],
[AS_HELP_STRING([--with-dmalloc],
   [use dmalloc, as in dmalloc.tar.gz from
 @/ftp.antaire.com:antaire/src/dmalloc.])],
[if test "$withval" = yes; then
  AC_MSG_RESULT([yes])
  AC_DEFINE([WITH_DMALLOC], [1], [Define to 1 if malloc debugging is enabled])
  LIBS="$LIBS -ldmalloc"
  LDFLAGS="$LDFLAGS -g"
else
  AC_MSG_RESULT([no])
fi], [AC_MSG_RESULT([no])])])
