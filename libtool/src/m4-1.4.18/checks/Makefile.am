## Makefile.am - template for generating Makefile via Automake.
##
## Copyright (C) 1992-1994, 2006-2014, 2016 Free Software Foundation,
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

# Vern says that the first star is required around an Alpha make bug.
DOC_CHECKS = $(srcdir)/*[0-9][0-9][0-9].*
CHECKS = $(DOC_CHECKS) $(srcdir)/stackovf.test
EXTRA_DIST = get-them check-them stamp-checks stackovf.test $(DOC_CHECKS)

all-local: $(srcdir)/stamp-checks

$(srcdir)/stamp-checks: $(srcdir)/get-them $(top_srcdir)/doc/m4.texi
	rm -f $(DOC_CHECKS)
	cd $(srcdir) && AWK=$(AWK) ./get-them ../doc/m4.texi
	echo stamp >$@

MAINTAINERCLEANFILES = $(DOC_CHECKS) $(srcdir)/stamp-checks

check-local: $(srcdir)/stamp-checks
	PATH=`pwd`/../src"$(PATH_SEPARATOR)"$$PATH; export PATH; \
	$(srcdir)/check-them -I $(srcdir)/../examples $(CHECKS)

installcheck: $(srcdir)/stamp-checks
	PATH='$(bindir)'"$(PATH_SEPARATOR)"$$PATH; export PATH; \
	$(srcdir)/check-them -I $(srcdir)/../examples \
	-m "`echo m4 | sed '$(program_transform_name)'`" $(CHECKS)
