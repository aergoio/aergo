#!/usr/bin/env bash

if [ "$#" != "4" ];then
	echo "Usage: $0 filepattern input_pat output_pat separator"
	exit 100
fi

rm -rf *.sedtmp

pattern=$1
input=$2
output=$3
separator=$4

echo "pattern=$pattern"

for file in $(ls *$pattern*); do
	echo $file

	tmpfile=$file.sedtmp
#	echo "sed s${separator}${input}${separator}$output${separator}g $file > $tmpfile"

	sed -e "s${separator}${input}${separator}$output${separator}g" $file > $tmpfile
	mv $tmpfile $file
done

