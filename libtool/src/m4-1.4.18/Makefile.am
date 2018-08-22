## Makefile.am - template for generating Makefile via Automake
##
## Copyright (C) 2000-2001, 2003-2014, 2016 Free Software Foundation,
## Inc.
##
## This file is part of GNU M4.
##
## GNU M4 is free software: you can redistribute it and/or modify
## it under the terms of the GNU General Public License as published by
## the Free Software Foundation, either version 3 of the License, or
## (at your option) any later version.
##
## GNU M4 is distributed in the hope that it will be useful,
## but WITHOUT ANY WARRANTY; without even the implied warranty of
## MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
## GNU General Public License for more details.
##
## You should have received a copy of the GNU General Public License
## along with this program.  If not, see <http://www.gnu.org/licenses/>.
##
## Written by Gary V. Vaughan <gary@gnu.org>

SUBDIRS = . examples lib src doc checks tests

EXTRA_DIST = bootstrap c-boxes.el cfg.mk maint.mk \
	.prev-version .version m4/gnulib-cache.m4 ChangeLog-2014
DISTCLEANFILES = stamp-h
## maintainer-clean should remove as much as possible that ./bootstrap can
## recreate.  In the m4 directory, keep only gnulib-cache.m4.
MAINTAINERCLEANFILES = INSTALL Makefile.in aclocal.m4 \
	config-h.in configure depcomp doc/fdl.texi gendocs.sh install-sh \
	lib/[a-z]* m4/[a-fh-z]* m4/g[a-mo-z]* m4/gnulib-comp.m4 \
	m4/gnulib-tool.m4 missing stamp-h.in

ACLOCAL_AMFLAGS = -I m4
## Enough users install GNU M4 as gm4 that we make sure 'make installcheck'
## will handle that, prior to making a release.
DISTCHECK_CONFIGURE_FLAGS = --enable-changeword --program-prefix=g \
	--enable-gcc-warnings --enable-silent-rules --enable-cxx

# Generate the ChangeLog from git history.
gen_start_date = 2015-01-01
.PHONY: gen-ChangeLog
gen-ChangeLog:
	$(AM_V_GEN)if test -d $(top_srcdir)/.git; then			\
	  log_fix="$(srcdir)/build-aux/git-log-fix";			\
	  test -e "$$log_fix"						\
	    && amend_git_log="--amend=$$log_fix"			\
	    || amend_git_log=;						\
	  $(top_srcdir)/build-aux/gitlog-to-changelog $$amend_git_log	\
	   --since='$(gen_start_date)' > $(distdir)/cl-t		\
	  && rm -f $(distdir)/ChangeLog					\
	  && mv $(distdir)/cl-t $(distdir)/ChangeLog;			\
	fi

# Dummy rule, to pacify automake "gnu" strictness and allow a clean
# bootstrap.  Creates a dummy ChangeLog (unless one is already present,
# as might be the case for users of vc-dwim) that will be overridden
# with a proper one at distribution time.
$(srcdir)/ChangeLog:
	@echo dummy > $@

BUILT_SOURCES = $(top_srcdir)/.version
$(top_srcdir)/.version:
	echo $(VERSION) > $@-t && mv $@-t $@

# Arrange so that .tarball-version appears only in the distribution
# tarball, and never in a checked-out repository.
dist-hook: gen-ChangeLog
	$(AM_V_GEN)echo $(VERSION) > $(distdir)/.tarball-version
