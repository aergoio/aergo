#!/usr/bin/env bash
# usage: $0 [bpprefix]
# help :  kill.sh BP11

pattern=$1

if [ $# -gt 1 ]; then
	echo "Usage: $0 pattern(ex:BP11001)"
fi

if [ $# -eq 0 ]; then
    pattern="BP"
fi

echo "kill $pattern"
for i in $(ps -ef| grep aergosvr | grep "BP"| grep "$pattern" | awk '{ print $2 }')
do
    if [ $i -gt 0 ]; then
        echo "kill -9 $i"
        kill -9 $i
    fi
done

sleep 3
echo "remain processes..."
echo "$(ps -ef| grep aergosvr | grep BP | grep -v grep)"
echo "done"

