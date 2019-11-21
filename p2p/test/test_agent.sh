#!/bin/bash

if [ $# -ne 1 ]; then
  echo "illegal number of parameters"
  kill -INT $$
fi

confFile="$1"
echo "using config file ${confFile} "

my_dir="$(dirname "$0")"

# importing config
. "$my_dir/$confFile"

# check config
if [ ${#machines[@]} != ${#m_dirs[@]} ] ; then
  echo "machine info and dir info size are differ machine ${#machines[@]} , but dirs ${#m_dirs[@]} "
  kill -INT $$
fi

function generateConf() {
  nodeVal=$1
  count=${#nodeVal[@]}
    for (( j=1; j <= count; ++j ));
    do
      echo "$nodeName - $j = ${nodeVal[$j]}"
    done
}

generateConf "${nodes1}"

echo "Finished"
