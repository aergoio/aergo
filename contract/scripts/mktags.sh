#!/bin/sh

rm cscope.* 2>/dev/null
rm tags 2>/dev/null

find `pwd` \( -name '*.c' -o -name '*.cpp' -o -name '*.h' -o -name '*.hpp' -o -name '*.l' -o -name '*.y' -o -name '*.list' \) ! \( -path "*/build/*" \) -print > cscope.files
cscope -b

ctags -R --c++-kinds=+p --fields=+iaS --extra=+f --exclude=build --sort=no `pwd` 2>/dev/null

egrep "\.list$" cscope.files | xargs perl -ne '
next if /.*define.*/;
@def_formats = ( "%s" ) if $ARGV ne $last_fname;
$last_fname = $ARGV;
if (/.*tag-format\s*:\s*(.*)$/) {
    @def_formats = split /, */, $1;
    next;
}
if (/^\s*\w+\(\s*(\w+)/gc ) {
    my @def = ( $1 );
    while (/\s*,\s*(\w+)/gc) { push @def, $1; }
    chomp;
    s#/#\\/#g;
    my $line = $_;
    map {
        printf 
            "$_\t%" . ($#def + 2) . "\$s\t" .
            "/^%" . ($#def + 3) . "\$s\$/;\"\td\n", 
            @def, $ARGV, $line
    } @def_formats;
}' >> tags
