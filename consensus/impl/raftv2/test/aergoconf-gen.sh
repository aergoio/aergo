#!/bin/bash 
# Usage: aergoconf-gen.sh 10001 tmpl.toml 5
#
# gen_start, gen_end: XXXX.toml 을 생성할 범위
# peer_start, peer_end : peer로 들어갈 node의 범위

# pkgen 명령 위치(path)를 적어줄 것
pkgen="aergocli keygen"

# pkey는 항상 0~22까지를 생성. 존재하면 생성 skip

if [ $# != 3 ]; then
    echo "Usage: $0 <starting port:10001~> <template> <max>"
    exit 100
fi


port0=$1
tmpl_file=$2
max=$3
gen_start=0
gen_end=$(($max - 1))

echo "Generate aergo config file from $tmpl_file, from $port0 to ($port0 + $max - 1)"

function gen_ids() {
	for ((i=1; i <= $max; i++))
	do
		out[i]=${!i}
	done

	echo ${out[*]} | sed -e 's= =\,\\n=g' -e 's=\/=\\/=g'
}

# gen_peers selfno startno endno ( 0 <= *no < $max)
function gen_peers() {
	self=$1
	start=$2
	end=$3

	j=0
	for ((i=$start; i <= $end; i++))
	do
		if [ $i != $self ]; then
			out[j]=${peer[$i]}
			j=$(($j + 1))
		fi

	done

    echo ${out[*]} | sed -e 's= =\,=g' -e 's=\/=\\/=g'
}

    

# generate id[]
for((i=0; i < $max; ++i))
do
    rpcport=$(($port0 + $i))
    rpc[i]=$rpcport

    profileport=$(($port0 + $i + 100))
    prof[i]=$profileport

    restport=$(($port0 + $i + 200))
    rest[i]=$restport
	p2pport=$(($port0 + $i + 1000))
    p2p[i]=$p2pport

	httpport=$(($port0 + $i + 3000))
    http[i]=$httpport

    pk[i]=${p2pport}.key
    pk_id[i]=${p2pport}.id

    ofile[i]=BP${p2pport}.toml

	# check if file exist
	if [ ! -e ${pk[i]} ]; then
		echo "${pk[i]}: $pkgen $p2pport"
		$pkgen $p2pport
	fi

    tmpid=$(cat ${p2pport}.id)
    id[i]="\"${tmpid}\""
	peer[i]="\"/ip4/127.0.0.1/tcp/${p2pport}/p2p/${tmpid}\""
	#echo "peer[ $i ]= ${peer[i]}"

	raftname[i]="aergo$(($i + 1))"
	raftbp[i]="{name=\"${raftname[i]}\",url=\"http://127.0.0.1:$httpport\",p2pid=\"${tmpid}\"}"
	echo "raftbp[$i]=${raftbp[i]}"
done

#ids=$(gen_ids ${id[*]})
#echo $ids
raftbps=$(gen_ids ${raftbp[*]})
echo "raftbps=$raftbps"

#for((i=0; i < $max; ++i))

for((i=$gen_start; i <= $gen_end; ++i))
do
    peers=$(gen_peers ${i} $gen_start $gen_end)

#	echo  "${peers}"

    echo "s=_home_=$PWD/${p2p[i]}=g" 
    echo "s/_rpc_/${rpc[i]}/g" 
    echo "s/_p2p_/${p2p[i]}/g"
    echo "s/_peer_/${peers}/g" 
    echo "s/_pk_/${pk[i]}/g" 
    echo "s/_http_/${http[i]}/g" 
	echo "s/_raftbps_/$raftbps/g" 


    sed -e "s=_home_=$PWD/${p2p[i]}=g" \
        -e "s=_data_=$PWD/data/${p2p[i]}=g" \
        -e "s/_rest_/${rest[i]}/g" \
		-e "s/_prof_/${prof[i]}/g" \
		-e "s/_rpc_/${rpc[i]}/g" \
        -e "s/_p2p_/${p2p[i]}/g" \
        -e "s/_peer_/${peers}/g" \
        -e "s/_http_/${http[i]}/g" \
        -e "s/_raftname_/${raftname[i]}/g" \
        -e "s/_raftbps_/${raftbps}/g" \
        -e "s/_pk_/${pk[i]}/g"  $tmpl_file >${ofile[i]}
done
echo $(pwd)
