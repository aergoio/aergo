#!/usr/bin/env bash
# usage: clean [all]
if [ $# -gt 1 ]; then
	echo "usage: clean [all]"
	exit 100
fi

function rmall() {
	echo "rmall"
	#rm -rf *.id *.key *.pub BP*.toml
	rm -rf *.log
}

killall -9 aergosvr
rm -rf data* genesis*
#rm *.log

# make empty args to string
if [ "$1" = "all" ]; then
	rmall
fi



