#!/bin/sh

set -x
sed -i -e "s/$1/$2/g" `find . -name '*.[chly]' -o -name '*.list' -o -name '0*' | grep -v '\.build' | xargs egrep -n $1 | cut -d: -f1 | uniq`
