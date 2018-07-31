#!/bin/bash


INPUT_DIR=$1

if [ "$#" -ne 1 ]; then
	    echo "./make.sh [input_dir]"
	    exit 
fi


elapsed=0
for file in $INPUT_DIR/*.trx; do
	echo $file
       start_time="$(date -u +%s.%N)"
       #./aergocli committx --jsontxpath $file  &> /dev/null &
       ./aergocli committx --jsontxpath $file  &> /dev/null 
       pids="$pids $!"
       end_time="$(date -u +%s.%N)"
       elapsed="$(bc <<<"$elapsed+($end_time-$start_time)")"
done


wait $pids
#echo "enter any key to stop all process"
#read -n1 kbd
#pkill -P $BASHPID



