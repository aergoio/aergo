#!/usr/bin/env bash
if [ $# != 1 ]; then
    echo "Usage: $0 <bpname>"
    exit 100
fi

bpname=$1
newpassword="1234"


if [[ $bpname =~ "BP*" ]]; then
	echo "Err: bpname($bpname) must have BP*"
	exit 100
fi

if [ ! -e genesis -o ! -e genesis.json ]; then 
	echo "no genesis file"
	clean.sh

	wallet=`aergocli account new --password 1234 --path genesis `

	echo "create genesis wallet : $wallet" 
	echo "$wallet" > genesis_wallet.txt

	wif=`aergocli account export --address ${wallet} --password 1234 --path genesis `

	echo "export wif : $wif"
	echo $wif > ./wif.txt

	sed  -e "s/_genesis_wallet_/${wallet}/g" _genesis.json > genesis.json
else
	echo "use prev genesis file"
	wallet=$(cat ./genesis_wallet.txt)
    wif=$(cat ./wif.txt)
fi


# make toml
#aergoconf-gen.sh $1 $2 $3


# genesis block patch

PWD=$(pwd)

file="${bpname}.toml"
echo "config=$file"
BP_NAME=$bpname
if [ "${BP_NAME}" != "tmpl" -a "${BP_NAME}" != "arglog" ]; then
	 N_NAME=${BP_NAME#BP}
	 N_PORT=$((${N_NAME}-1000))

	DATADIR="$PWD/data/$N_NAME"
	echo "datadir=$DATADIR"

	echo "init genesis block "
	echo "aergosvr init --genesis ./genesis.json --home ${PWD} --config ./$file"
	aergosvr init --genesis ./genesis.json --home ${PWD} --config ./$file

	echo "import wallet ${wallet} to ${N_NAME} from local"
	echo "aergocli account import --if ${wif} --password 1234 --path ${DATADIR}"
	aergocli account import --if ${wif} --password 1234 --path ${DATADIR} 

	echo "start server ${BP_NAME} "
	echo "aergosvr --home ${PWD} --config ./$file >> server_${BP_NAME}.log 2>&1 &"
#nohup aergosvr --home ${PWD} --config ./$file >> server_${BP_NAME}.log 2>&1 &
	nohup aergosvr --home ${PWD} --config ./$file >> server_${BP_NAME}.log 2>&1 &

	echo "sleep 3s to wait boot"
	sleep 3

	echo "unlock account ${wallet} "
	echo "aergocli -p ${N_PORT} getstate --address ${wallet}"
	aergocli -p ${N_PORT} getstate --address ${wallet}
	echo  "aergocli -p ${N_PORT} account unlock --address ${wallet} --password 1234"
	aergocli -p ${N_PORT} account unlock --address ${wallet} --password 1234 
	echo "aergocli -p ${N_PORT} account list"
	aergocli -p ${N_PORT} account list
fi
