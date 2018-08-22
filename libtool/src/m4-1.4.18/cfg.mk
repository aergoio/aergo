# Customize maint.mk.                           -*- makefile -*-
# Copyright (C) 2003-2014, 2016 Free Software Foundation, Inc.

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

# Used in maint.mk's web-manual rule
manual_title = GNU macro processor

# Always use shorthand copyrights.
update-copyright-env = \
  UPDATE_COPYRIGHT_USE_INTERVALS=1 \
  UPDATE_COPYRIGHT_MAX_LINE_LENGTH=72

# Tests not to run as part of "make syntax-check".
local-checks-to-skip =
# We haven't yet turned on i18n
local-checks-to-skip += sc_bindtextdomain
# M4 intentionally uses a coding style that compiles under C++.
local-checks-to-skip += sc_cast_of_x_alloc_return_value

# Our files include "m4.h", which in turn includes <config.h> first.
config_h_header = "m4\.h"

# Hash of NEWS contents, to ensure we don't add entries to wrong section.
old_NEWS_hash = ef92e5b599b23d9dfe915db5ce4194c5

# Indent only with spaces.
sc_prohibit_tab_based_indentation:
 @re='^ *    '                                               \
 msg='TAB in indentation; use only spaces'                   \
   $(_prohibit_regexp)

# List all syntax-check exemptions:
exclude_file_name_regexp--sc_prohibit_tab_based_indentation = \
  (^(GNU)?Makefile(\.am)?|\.mk|^HACKING|^ChangeLog.*)$$
exclude_file_name_regexp--update-copyright = ^m4/gnulib-cache.m4$$
