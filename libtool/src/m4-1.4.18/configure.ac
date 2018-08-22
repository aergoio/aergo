# Configure template for GNU M4.           -*-Autoconf-*-
# Copyright (C) 1991, 1993-1994, 2004-2014, 2016 Free Software
# Foundation, Inc.
#
# This file is part of GNU M4.
#
# GNU M4 is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# GNU M4 is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

AC_PREREQ([2.62])
AC_INIT([GNU M4], m4_esyscmd([build-aux/git-version-gen .tarball-version]),
  [bug-m4@gnu.org])
AC_CONFIG_AUX_DIR([build-aux])

AM_INIT_AUTOMAKE([1.11.6 dist-bzip2 dist-xz color-tests parallel-tests
silent-rules subdir-objects gnu])
AM_SILENT_RULES([yes]) # make --enable-silent-rules the default.

m4_pattern_forbid([^M4_[A-Z]])

AC_CONFIG_SRCDIR([src/m4.h])
AC_CONFIG_HEADERS([lib/config.h:lib/config.hin])

AC_PROG_CC
AC_DEFUN([gl_CXX_CHOICE_DEFAULT_NO])
M4_EARLY

# Maintainer note - comment this line out if you plan to rerun
# GNULIB_POSIXCHECK testing to see if M4 should be using more modules.
# Leave it uncommented for normal releases, for faster ./configure.
gl_ASSERT_NO_GNULIB_POSIXCHECK

# M4 is single-threaded; so we can optimize gnulib code by using this:
gl_DISABLE_THREADS

AC_ARG_ENABLE([gcc-warnings],
  [AS_HELP_STRING([--enable-gcc-warnings],
                  [turn on lots of GCC warnings (for developers)])],
  [case $enableval in
     yes|no) ;;
     *)      AC_MSG_ERROR([bad value $enableval for gcc-warnings option]) ;;
   esac
   gl_gcc_warnings=$enableval],
  [if test -f "$srcdir"/.tarball-version; then
     gl_gcc_warnings=no
   else
     gl_gcc_warnings=yes
   fi]
)

if test "$gl_gcc_warnings" = yes; then
  gl_WARN_ADD([-Werror], [WERROR_CFLAGS])
  AC_SUBST([WERROR_CFLAGS])

  # We use '#pragma GCC diagnostic push' to silence some
  # warnings, but older gcc doesn't support this.
  AC_CACHE_CHECK([whether pragma GCC diagnostic push works],
    [M4_cv_gcc_pragma_push_works], [
    save_CFLAGS=$CFLAGS
    CFLAGS='-Wunknown-pragmas -Werror'
    AC_COMPILE_IFELSE([AC_LANG_PROGRAM([[
      #pragma GCC diagnostic push
      #pragma GCC diagnostic pop
    ]])],
    [M4_cv_gcc_pragma_push_works=yes],
    [M4_cv_gcc_pragma_push_works=no])
    CFLAGS=$save_CFLAGS])

  # This, $nw, is the list of warnings we disable.
  nw=
  nw="$nw -Waggregate-return"       # K&R is anachronistic
  nw="$nw -Wtraditional-conversion" # K&R is anachronistic
  nw="$nw -Wundef"                  # K&R is anachronistic
  nw="$nw -Wtraditional"            # K&R is anachronistic
  nw="$nw -Wlong-long"              # C90 is anachronistic
  nw="$nw -Wsystem-headers"         # Don't let system headers trigger warnings
  nw="$nw -Wpadded"                 # Our structs are not packed
  nw="$nw -Wconversion"             # Too many warnings for now
  nw="$nw -Wsign-conversion"        # Too many warnings for now
  nw="$nw -Wcast-qual"              # Too many warnings for now
  nw="$nw -Wswitch-enum"            # Too many warnings for now
  # gcc 4.4.6 complains enum-compare is C++ only; gcc 4.7.0 implies it in -Wall
  nw="$nw -Wenum-compare"

  # Gnulib uses '#pragma GCC diagnostic push' to silence some
  # warnings, but older gcc doesn't support this.
  AC_CACHE_CHECK([whether pragma GCC diagnostic push works],
    [M4_cv_gcc_pragma_push_works], [
    save_CFLAGS=$CFLAGS
    CFLAGS='-Wunknown-pragmas -Werror'
    AC_COMPILE_IFELSE([AC_LANG_PROGRAM([[
      #pragma GCC diagnostic push
      #pragma GCC diagnostic pop
    ]])],
    [M4_cv_gcc_pragma_push_works=yes],
    [M4_cv_gcc_pragma_push_works=no])
    CFLAGS=$save_CFLAGS])
  if test $M4_cv_gcc_pragma_push_works = no; then
    nw="$nw -Wmissing-prototypes"
    nw="$nw -Wmissing-declarations"
    nw="$nw -Wunreachable-code"
  fi

  gl_MANYWARN_ALL_GCC([ws])
  gl_MANYWARN_COMPLEMENT([ws], [$ws], [$nw])
  for w in $ws; do
    gl_WARN_ADD([$w])
  done

  # Gnulib turns on -Wformat=2, which implies -Wformat-nonliteral, but
  # we temporarily override it in format.c if possible.
  if test $M4_cv_gcc_pragma_push_works = no; then
    gl_WARN_ADD([-Wno-format-nonliteral])
  fi

  AC_SUBST([WARN_CFLAGS])

  AH_VERBATIM([FORTIFY_SOURCE],
  [/* Enable compile-time and run-time bounds-checking, and some warnings,
      without upsetting newer glibc. */
   #if defined __OPTIMIZE__ && __OPTIMIZE__
   # define _FORTIFY_SOURCE 2
   #endif
  ])
fi

# Tandem/NSK is broken - it has 'long long int' but not
# 'unsigned long long int', which confuses assumptions made by gnulib.
# Simply pretend that neither type exists if both do not work.
AC_TYPE_LONG_LONG_INT
AC_TYPE_UNSIGNED_LONG_LONG_INT
if test $ac_cv_type_long_long_int:$ac_cv_type_unsigned_long_long_int = yes:no
then
  ac_cv_type_long_long_int=no
  AC_DEFINE([HAVE_LONG_LONG_INT], 0,
    [Define to 1 if the system has the type `long long int'.])
fi

M4_INIT

AC_CACHE_CHECK([whether an open file can be renamed],
  [M4_cv_func_rename_open_file_works],
  [AC_RUN_IFELSE([AC_LANG_PROGRAM([AC_INCLUDES_DEFAULT],
      [FILE *f = fopen ("conftest.1", "w+");
       int result = rename ("conftest.1", "conftest.2");
       fclose (f); remove ("conftest.1"); remove ("conftest.2");
       return result;])],
    [M4_cv_func_rename_open_file_works=yes],
    [M4_cv_func_rename_open_file_works=no],
    [M4_cv_func_rename_open_file_works='guessing no'])])
if test "$M4_cv_func_rename_open_file_works" = yes ; then
  M4_rename_open_works=1
else
  M4_rename_open_works=0
fi
AC_DEFINE_UNQUOTED([RENAME_OPEN_FILE_WORKS], [$M4_rename_open_works],
  [Define to 1 if a file can be renamed while open, or to 0 if not.])

dnl Don't let changeword get in our way, if bootstrapping with a version of
dnl m4 that already turned the feature on.
m4_ifdef([changeword], [m4_undefine([changeword])])dnl

AC_MSG_CHECKING([[if changeword is wanted]])
AC_ARG_ENABLE([changeword],
  [AS_HELP_STRING([--enable-changeword],
     [enable -W and changeword() builtin])],
  [if test "$enableval" = yes; then
    AC_MSG_RESULT([yes])
    AC_DEFINE([ENABLE_CHANGEWORD], [1],
      [Define to 1 if the changeword(REGEXP) functionality is wanted])
  else
    AC_MSG_RESULT([no])
  fi], [AC_MSG_RESULT([no])])

AC_MSG_CHECKING([[which shell to use for syscmd]])
AC_ARG_WITH([syscmd-shell],
  [AS_HELP_STRING([--with-syscmd-shell], [shell used by syscmd [/bin/sh]])],
  [case $withval in
    yes[)] with_syscmd_shell=no;;
   esac], [with_syscmd_shell=no])
if test "$with_syscmd_shell" = no ; then
  with_syscmd_shell=/bin/sh
  if test "$cross_compiling" != yes ; then
dnl Give mingw a default that is more likely to be available.
    AS_IF([AS_EXECUTABLE_P([/bin/sh$EXEEXT])], [],
      [if (cmd /c) 2>/dev/null; then with_syscmd_shell=cmd; fi])
dnl Too bad _AS_PATH_WALK is not public.
    M4_save_IFS=$IFS; IFS=$PATH_SEPARATOR
    for M4_dir in `if (command -p getconf PATH) 2>/dev/null ; then
        command -p getconf PATH
      else
        echo "/bin$PATH_SEPARATOR$PATH"
      fi`
    do
      IFS=$M4_save_IFS
      test -z "$M4_dir" && continue
      AS_EXECUTABLE_P(["$M4_dir/sh"]) \
        && { with_syscmd_shell=$M4_dir/sh; break; }
    done
    IFS=$M4_save_IFS
  fi
fi
AC_MSG_RESULT([$with_syscmd_shell])
AC_DEFINE_UNQUOTED([SYSCMD_SHELL], ["$with_syscmd_shell"],
  [Shell used by syscmd and esyscmd, must accept -c argument.])

M4_WITH_DMALLOC

AC_CONFIG_COMMANDS([stamp-h], [[test -z "$CONFIG_HEADERS" || date > stamp-h]])

AS_CASE([$host], [*-*-os2*], [OS2_LDFLAGS="-Zargs-resp"])
AC_SUBST([OS2_LDFLAGS])

AC_CONFIG_FILES([Makefile
                 doc/Makefile
                 lib/Makefile
                 src/Makefile
                 tests/Makefile
                 checks/Makefile
                 examples/Makefile
])

AC_OUTPUT
