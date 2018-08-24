#!/bin/sh

rm cscope.* 2>/dev/null
find `pwd` \( -name '*.c' -o -name '*.cpp' -o -name '*.cc' -o -name '*.h' -o -name '*.s' -o -name '*.S' -o -name '*.list' \) -print > cscope.files
cscope -b

rm tags 2>/dev/null
ctags -R --c++-kinds=+p --fields=+iaS --extra=+q --exclude=build `pwd` 2>/dev/null
