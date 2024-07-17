set -e
source common.sh

fork_version=$1


echo "-- deploy --"

deploy ../contract/vm_dummy/test_files/feature_multicall.lua
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')
assert_equals "$status" "CREATED"

deploy ../contract/vm_dummy/test_files/feature_multicall.lua
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
address1=$(cat receipt.json | jq .contractAddress | sed 's/"//g')
assert_equals "$status" "CREATED"

deploy ../contract/vm_dummy/test_files/feature_multicall.lua
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
address2=$(cat receipt.json | jq .contractAddress | sed 's/"//g')
assert_equals "$status" "CREATED"

deploy ../contract/vm_dummy/test_files/feature_multicall.lua
get_receipt $txhash
status=$(cat receipt.json | jq .status | sed 's/"//g')
address3=$(cat receipt.json | jq .contractAddress | sed 's/"//g')
assert_equals "$status" "CREATED"



echo "-- create accounts --"


# AmNLhiVVdLxbPW5NxpLibqoLdobc2TaKVk8bwrPn5VXz6gUSLvke
# AmLhv6rvLEMoL5MLws1tNTyiYYyzTx4JGjaPAugzqabpjxxWyP34
# AmPfKfrbSjbHD77JjpEvdktgH3w6sbWRhhwFmFAyi1ZSsiePs7XP
# AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw

../bin/aergocli account import --keystore . --password bmttest --if 47FPLReaeJi6BNRgxYURjR1yvtWoCY9vs2zEU7NFCUEyQKKMWeZe9SKaU41YR6aVMAHpyHSDw
../bin/aergocli account import --keystore . --password bmttest --if 477rUib5imztPLhszhdKrqarWpxwxgP8hwPDTbxFJBzdSG1rfssBQeZv3q3FqSzbQF1ioMjMJ
../bin/aergocli account import --keystore . --password bmttest --if 47djKu2o5NshYhMtPsu62GektW7C1nKaxVNgnJUJWMKDr77gjv2aHSom1aZjrwwqQNKGKRaFv
../bin/aergocli account import --keystore . --password bmttest --if 47G3YUtkcLQGWX2wQMTTFTvyf2ZRTQz1ZdEMaX26FcXqMRhNF2J8KVSrhY1MrZuVvTgaMWmDC

account0=AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R

account1=AmNLhiVVdLxbPW5NxpLibqoLdobc2TaKVk8bwrPn5VXz6gUSLvke
account2=AmLhv6rvLEMoL5MLws1tNTyiYYyzTx4JGjaPAugzqabpjxxWyP34
account3=AmPfKfrbSjbHD77JjpEvdktgH3w6sbWRhhwFmFAyi1ZSsiePs7XP
account4=AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw

if [ "$consensus" != "sbp" ]; then
	# send 20 aergo to each address
	accounts=($account1 $account2 $account3 $account4)
	for account in "${accounts[@]}"; do
		echo "sending 20 aergo to $account ..."

		txhash=$(../bin/aergocli --keystore . --password bmttest \
			sendtx --from $account0 --to $account --amount 20aergo \
			| jq .hash | sed 's/"//g')

		get_receipt $txhash

		status=$(cat receipt.json | jq .status | sed 's/"//g')
		assert_equals "$status"   "SUCCESS"
	done
fi


function multicall() {
	account=$1
	script=$2
	expected_error=$3
	expected_result=$4

	echo " "
	echo "-- multicall --"

	echo "script=$script"
	echo "expected_result=$expected_result"

	if [ "$account" == "ac0" ]; then
		account=AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R
	elif [ "$account" == "ac1" ]; then
		account=AmNLhiVVdLxbPW5NxpLibqoLdobc2TaKVk8bwrPn5VXz6gUSLvke
	elif [ "$account" == "ac2" ]; then
		account=AmLhv6rvLEMoL5MLws1tNTyiYYyzTx4JGjaPAugzqabpjxxWyP34
	elif [ "$account" == "ac3" ]; then
		account=AmPfKfrbSjbHD77JjpEvdktgH3w6sbWRhhwFmFAyi1ZSsiePs7XP
	elif [ "$account" == "ac4" ]; then
		account=AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw
	fi

	txhash=$(../bin/aergocli --keystore . --password bmttest \
	  contract multicall $account "$script" | jq .hash | sed 's/"//g')

	get_receipt $txhash

	status=$(cat receipt.json | jq .status | sed 's/"//g')
	ret=$(cat receipt.json | jq -c .ret)
	gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

	if [ "$ret" == '""' ]; then
	  ret=""
	fi

	#expected_result=$(echo "$expected_result" | jq -c .)
	#echo "ret            =$ret"
	#echo "expected_result=$expected_result"

	if [ -n "$expected_error" ]; then
		assert_equals "$status"   "ERROR"
		#assert_contains "$ret"    "$expected_error"
	else
		assert_equals "$status"   "SUCCESS"
		assert_equals "$ret"      "$expected_result"
	fi
}

multicall "ac1" '[
 ["call","'$address'","get_dict"],
 ["store result as","dict"],
 ["set","%dict%","two",22],
 ["set","%dict%","four",4],
 ["set","%dict%","one",null],
 ["get","%dict%","two"],
 ["set","%dict%","copy","%last result%"],
 ["return","%dict%"]
]' '' '{"copy":22,"four":4,"three":3,"two":22}'

multicall "ac1" '[
 ["call","'$address'","get_list"],
 ["store result as","array"],
 ["set","%array%",2,"2nd"],
 ["insert","%array%",1,"zero"],
 ["insert","%array%","last"],
 ["return","%array%"]
]' '' '["zero","first","2nd","third",123,12.5,true,"last"]'

multicall "ac1" '[
 ["call","'$address'","get_list"],
 ["store result as","array"],
 ["remove","%array%",3],
 ["return","%array%","%last result%"]
]' '' '[["first","second",123,12.5,true],"third"]'


# create new dict or array using fromjson

multicall "ac1" '[
 ["from json","{\"one\":1,\"two\":2}"],
 ["set","%last result%","three",3],
 ["return","%last result%"]
]' '' '{"one":1,"three":3,"two":2}'


# define dict or list using let

multicall "ac1" '[
 ["let","obj",{"one":1,"two":2}],
 ["set","%obj%","three",3],
 ["return","%obj%"]
]' '' '{"one":1,"three":3,"two":2}'

multicall "ac1" '[
 ["let","list",["one",1,"two",2,2.5,true,false]],
 ["set","%list%",4,"three"],
 ["insert","%list%",1,"first"],
 ["insert","%list%","last"],
 ["return","%list%"]
]' '' '["first","one",1,"two","three",2.5,true,false,"last"]'

multicall "ac1" '[
 ["let","list",["one",22,3.3,true,false]],
 ["get","%list%",1],
 ["assert","%last result%","=","one"],
 ["get","%list%",2],
 ["assert","%last result%","=",22],
 ["get","%list%",3],
 ["assert","%last result%","=",3.3],
 ["get","%list%",4],
 ["assert","%last result%","=",true],
 ["get","%list%",5],
 ["assert","%last result%","=",false],
 ["return","%list%"]
]' '' '["one",22,3.3,true,false]'


# get size

multicall "ac1" '[
 ["let","str","this is a string"],
 ["get size","%str%"],
 ["return","%last result%"]
]' '' '16'

multicall "ac1" '[
 ["let","list",["one",1,"two",2,2.5,true,false]],
 ["get size","%list%"],
 ["return","%last result%"]
]' '' '7'

multicall "ac1" '[
 ["let","obj",{"one":1,"two":2,"three":3}],
 ["get size","%obj%"],
 ["return","%last result%"]
]' '' '0'




# BIGNUM

multicall "ac1" '[
 ["to big number",123],
 ["store result as","a"],
 ["to big number",123],
 ["store result as","b"],
 ["multiply","%a%","%b%"],
 ["return","%last result%"]
]' '' '{"_bignum":"15129"}'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","100000"],
 ["store result as","b"],
 ["divide","%a%","%b%"],
 ["return","%last result%"]
]' '' '{"_bignum":"5000000000000000"}'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","100000"],
 ["store result as","b"],
 ["divide","%a%","%b%"],
 ["to string","%last result%"],
 ["return","%last result%"]
]' '' '"5000000000000000"'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],

 ["to big number","100000"],
 ["divide","%a%","%last result%"],
 ["store result as","a"],

 ["to big number","1000000000000000"],
 ["subtract","%a%","%last result%"],
 ["store result as","a"],

 ["to big number","1234"],
 ["add","%a%","%last result%"],
 ["store result as","a"],

 ["to big number","2"],
 ["remainder","%a%","10000"],

 ["return","%last result%"]
]' '' '{"_bignum":"1234"}'



# STRINGS

multicall "ac1" '[
 ["format","%s%s%s","hello"," ","world"],
 ["return","%last result%"]
]' '' '"hello world"'

multicall "ac1" '[
 ["let","s","hello world"],
 ["extract","%s%",1,4],
 ["return","%last result%"]
]' '' '"hell"'

multicall "ac1" '[
 ["let","s","hello world"],
 ["extract","%s%",-2,-1],
 ["return","%last result%"]
]' '' '"ld"'

multicall "ac1" '[
 ["let","s","the amount is 12345"],
 ["find","%s%","%d+"],
 ["to number"],
 ["return","%last result%"]
]' '' '12345'

multicall "ac1" '[
 ["let","s","rate: 55 10%"],
 ["find","%s%","(%d+)%%"],
 ["to number"],
 ["return","%last result%"]
]' '' '10'

multicall "ac1" '[
 ["let","s","rate: 12%"],
 ["find","%s%","%s*(%d+)%%"],
 ["to number"],
 ["return","%last result%"]
]' '' '12'

multicall "ac1" '[
 ["let","s","hello world"],
 ["replace","%s%","hello","good bye"],
 ["return","%last result%"]
]' '' '"good bye world"'

multicall "ac1" '[
 ["from json","{\"name\":\"ticket\",\"value\":12.5,\"amount\":10}"],
 ["replace","name = $name, value = $value, amount = $amount","%$(%w+)","%last result%"],
 ["return","%last result%"]
]' '' '"name = ticket, value = 12.5, amount = 10"'


# IF THEN ELSE

multicall "ac1" '[
 ["let","s",20],
 ["if","%s%",">=",20],
 ["let","b","big"],
 ["else if","%s%",">=",10],
 ["let","b","medium"],
 ["else"],
 ["let","b","low"],
 ["end if"],
 ["let","c","after"],
 ["return","%b%","%c%"]
]' '' '["big","after"]'

multicall "ac1" '[
 ["let","s",10],
 ["if","%s%",">=",20],
 ["let","b","big"],
 ["else if","%s%",">=",10],
 ["let","b","medium"],
 ["else"],
 ["let","b","low"],
 ["end if"],
 ["let","c","after"],
 ["return","%b%","%c%"]
]' '' '["medium","after"]'

multicall "ac1" '[
 ["let","s",5],
 ["if","%s%",">=",20],
 ["let","b","big"],
 ["else if","%s%",">=",10],
 ["let","b","medium"],
 ["else"],
 ["let","b","low"],
 ["end if"],
 ["let","c","after"],
 ["return","%b%","%c%"]
]' '' '["low","after"]'

multicall "ac1" '[
 ["let","s",20],
 ["if","%s%",">=",20],
 ["return","big"],
 ["else if","%s%",">=",10],
 ["return","medium"],
 ["else"],
 ["return","low"],
 ["end if"],
 ["return","after"]
]' '' '"big"'

multicall "ac1" '[
 ["let","s",10],
 ["if","%s%",">=",20],
 ["return","big"],
 ["else if","%s%",">=",10],
 ["return","medium"],
 ["else"],
 ["return","low"],
 ["end if"],
 ["return","after"]
]' '' '"medium"'

multicall "ac1" '[
 ["let","s",5],
 ["if","%s%",">=",20],
 ["return","big"],
 ["else if","%s%",">=",10],
 ["return","medium"],
 ["else"],
 ["return","low"],
 ["end if"],
 ["return","after"]
]' '' '"low"'


multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","500000000000000000000"],
 ["store result as","b"],
 ["if","%a%","=","%b%"],
 ["let","b","equal"],
 ["else"],
 ["let","b","diff"],
 ["end if"],
 ["return","%b%"]
]' '' '"equal"'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","500000000000000000001"],
 ["store result as","b"],
 ["if","%a%","=","%b%"],
 ["let","b","equal"],
 ["else"],
 ["let","b","diff"],
 ["end if"],
 ["return","%b%"]
]' '' '"diff"'

multicall "ac1" '[
 ["to big number","500000000000000000001"],
 ["store result as","a"],
 ["to big number","500000000000000000000"],
 ["store result as","b"],
 ["if","%a%",">","%b%"],
 ["let","b","bigger"],
 ["else"],
 ["let","b","lower"],
 ["end if"],
 ["return","%b%"]
]' '' '"bigger"'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","500000000000000000001"],
 ["store result as","b"],
 ["if","%a%",">","%b%"],
 ["let","b","bigger"],
 ["else"],
 ["let","b","lower"],
 ["end if"],
 ["return","%b%"]
]' '' '"lower"'


multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","500000000000000000001"],
 ["store result as","b"],
 ["if","%a%","<","%b%","and","1","=","0"],
 ["let","b","wrong 1"],
 ["else if","%a%","<","%b%","and","1","=","1"],
 ["let","b","correct"],
 ["else"],
 ["let","b","wrong 2"],
 ["end if"],
 ["return","%b%"]
]' '' '"correct"'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","500000000000000000001"],
 ["store result as","b"],
 ["if","%a%","<","%b%","and",1,"=",0],
 ["let","b","wrong 1"],
 ["else if","%a%","<","%b%","and",1,"=",1],
 ["let","b","correct"],
 ["else"],
 ["let","b","wrong 2"],
 ["end if"],
 ["return","%b%"]
]' '' '"correct"'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","500000000000000000001"],
 ["store result as","b"],
 ["to big number","400000000000000000000"],
 ["store result as","c"],
 ["if","%a%","<","%b%","and","%a%","<","%c%"],
 ["let","b","wrong 1"],
 ["else if","%a%",">","%b%","and","%a%",">","%c%"],
 ["let","b","wrong 2"],
 ["else if","%a%","<","%b%","and","%a%",">","%c%"],
 ["let","b","correct"],
 ["else"],
 ["let","b","wrong 3"],
 ["end if"],
 ["return","%b%"]
]' '' '"correct"'

multicall "ac1" '[
 ["to big number","500000000000000000000"],
 ["store result as","a"],
 ["to big number","500000000000000000001"],
 ["store result as","b"],
 ["to big number","400000000000000000000"],
 ["store result as","c"],
 ["to string",0],

 ["if","%a%",">","%b%","or","%a%","<","%c%"],
 ["format","%s%s","%last result%","1"],
 ["end if"],

 ["if","%a%","=","%b%","or","%a%","=","%c%"],
 ["format","%s%s","%last result%","2"],
 ["end if"],

 ["if","%a%","<","%b%","or","%a%","<","%c%"],
 ["format","%s%s","%last result%","3"],
 ["end if"],

 ["if","%a%",">","%b%","or","%a%",">","%c%"],
 ["format","%s%s","%last result%","4"],
 ["end if"],

 ["if","%a%","<","%b%","or","%a%",">","%c%"],
 ["format","%s%s","%last result%","5"],
 ["end if"],

 ["if","%a%","!=","%b%","or","%a%","=","%c%"],
 ["format","%s%s","%last result%","6"],
 ["end if"],

 ["if","%a%","=","%b%","or","%a%","!=","%c%"],
 ["format","%s%s","%last result%","7"],
 ["end if"],


 ["if","%a%",">=","%b%","and","%a%","<=","%c%"],
 ["format","%s%s","%last result%","8"],
 ["end if"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%"],
 ["format","%s%s","%last result%","9"],
 ["end if"],

 ["if","%b%",">=","%a%","and","%b%","<=","%c%"],
 ["format","%s%s","%last result%","A"],
 ["end if"],

 ["if","%b%",">=","%c%","and","%b%","<=","%a%"],
 ["format","%s%s","%last result%","B"],
 ["end if"],

 ["if","%c%",">=","%a%","and","%c%","<=","%b%"],
 ["format","%s%s","%last result%","C"],
 ["end if"],

 ["if","%c%",">=","%b%","and","%c%","<=","%a%"],
 ["format","%s%s","%last result%","D"],
 ["end if"],


 ["if","%a%",">=","%b%","and","%a%","<=","%c%","or",1,"=",0],
 ["format","%s%s","%last result%","E"],
 ["end if"],

 ["if","%a%",">=","%b%","and","%a%","<=","%c%","or",1,"=",1],
 ["format","%s%s","%last result%","F"],
 ["end if"],

 ["if","%a%",">=","%b%","and","%a%","<=","%c%","and",1,"=",0],
 ["format","%s%s","%last result%","G"],
 ["end if"],

 ["if","%a%",">=","%b%","and","%a%","<=","%c%","and",1,"=",1],
 ["format","%s%s","%last result%","H"],
 ["end if"],


 ["if","%a%",">=","%c%","and","%a%","<=","%b%","or",1,"=",0],
 ["format","%s%s","%last result%","I"],
 ["end if"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%","or",1,"=",1],
 ["format","%s%s","%last result%","J"],
 ["end if"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%","and",1,"=",0],
 ["format","%s%s","%last result%","K"],
 ["end if"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%","and",1,"=",1],
 ["format","%s%s","%last result%","L"],
 ["end if"],


 ["if",1,"=",0,"or","%a%",">=","%b%","and","%a%","<=","%c%"],
 ["format","%s%s","%last result%","M"],
 ["end if"],

 ["if",1,"=",1,"or","%a%",">=","%b%","and","%a%","<=","%c%"],
 ["format","%s%s","%last result%","N"],
 ["end if"],

 ["if",1,"=",0,"or","%a%",">=","%c%","and","%a%","<=","%b%"],
 ["format","%s%s","%last result%","O"],
 ["end if"],

 ["if",1,"=",1,"or","%a%",">=","%c%","and","%a%","<=","%b%"],
 ["format","%s%s","%last result%","P"],
 ["end if"],


 ["return","%last result%"]
]' '' '"0345679IJLNOP"'




# FOR

multicall "ac1" '[
 ["for","n",1,5],
 ["loop"],
 ["return","%n%"]
]' '' '6'

multicall "ac1" '[
 ["to number","0"],
 ["for","n",1,5],
 ["add","%last result%",1],
 ["loop"],
 ["return","%last result%"]
]' '' '5'

multicall "ac1" '[
 ["to big number","10000000000000000001"],
 ["store result as","to_add"],
 ["to big number","100000000000000000000"],

 ["for","n",1,3],
 ["add","%last result%","%to_add%"],
 ["loop"],

 ["to string","%last result%"],
 ["return","%last result%"]
]' '' '"130000000000000000003"'


multicall "ac1" '[
 ["to number","0"],
 ["for","n",500,10,-5],
 ["add","%last result%",1],
 ["loop"],
 ["return","%last result%"]
]' '' '99'


multicall "ac1" '[
 ["to number","0"],
 ["for","n",5,1,-1],
 ["add","%last result%",1],
 ["loop"],
 ["return","%last result%"]
]' '' '5'

multicall "ac1" '[
 ["to number","0"],
 ["for","n",5,1],
 ["add","%last result%",1],
 ["loop"],
 ["return","%last result%"]
]' '' '0'

multicall "ac1" '[
 ["to number","0"],
 ["for","n",1,5],
 ["add","%last result%",1],
 ["loop"],
 ["return","%last result%"]
]' '' '5'

multicall "ac1" '[
 ["to number","0"],
 ["for","n",1,5,-1],
 ["add","%last result%",1],
 ["loop"],
 ["return","%last result%"]
]' '' '0'



# FOREACH

multicall "ac1" '[
 ["let","list",[11,22,33]],
 ["let","r",0],
 ["for each","item","in","%list%"],
 ["add","%r%","%item%"],
 ["store result as","r"],
 ["loop"],
 ["return","%r%"]
]' '' '66'

multicall "ac1" '[
 ["let","list",[11,22,33]],
 ["let","counter",0],
 ["for each","item","in","%list%"],
 ["add","%counter%",1],
 ["store result as","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '3'

multicall "ac1" '[
 ["let","list",[]],
 ["let","counter",0],
 ["for each","item","in","%list%"],
 ["add","%counter%",1],
 ["store result as","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '0'

multicall "ac1" '[
 ["let","list",["one",1,"two",2,2.5,true,false]],
 ["let","counter",0],
 ["for each","item","in","%list%"],
 ["add","%counter%",1],
 ["store result as","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '7'

multicall "ac1" '[
 ["let","list",[10,21,32]],
 ["let","r",0],
 ["for each","item","in","%list%"],
 ["if","%item%","<",30],
 ["add","%r%","%item%"],
 ["store result as","r"],
 ["end if"],
 ["loop"],
 ["return","%r%"]
]' '' '31'



# FORPAIR

multicall "ac1" '[
 ["let","str",""],
 ["let","sum",0],
 ["let","obj",{"one":1,"two":2,"three":3}],
 ["for each","key","value","in","%obj%"],
 ["combine","%str%","%key%"],
 ["store result as","str"],
 ["add","%sum%","%value%"],
 ["store result as","sum"],
 ["loop"],
 ["return","%str%","%sum%"]
]' '' '["onethreetwo",6]'

multicall "ac1" '[
 ["let","str",""],
 ["let","sum",0],
 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
 ["for each","key","value","in","%obj%"],
 ["combine","%str%","%key%"],
 ["store result as","str"],
 ["add","%sum%","%value%"],
 ["store result as","sum"],
 ["loop"],
 ["return","%str%","%sum%"]
]' '' '["fouronethreetwo",12]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","values",[]],
 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
 ["for each","key","value","in","%obj%"],
 ["insert","%names%","%key%"],
 ["insert","%values%","%value%"],
 ["loop"],
 ["return","%names%","%values%"]
]' '' '[["four","one","three","two"],[4.5,1.5,3.5,2.5]]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","values",[]],
 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
 ["for each","key","value","in","%obj%"],
 ["insert","%names%","%key%"],
 ["insert","%values%","%value%"],
 ["loop"],
 ["call","'$address'","sort","%values%"],
 ["store result as","values"],
 ["return","%names%","%values%"]
]' '' '[["four","one","three","two"],[1.5,2.5,3.5,4.5]]'

multicall "ac1" '[
 ["let","obj",{}],
 ["let","counter",0],
 ["for each","key","value","in","%obj%"],
 ["add","%counter%",1],
 ["store result as","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '0'



# FOR "BREAK"

multicall "ac1" '[
 ["let","c",0],
 ["for","n",1,10],
 ["add","%c%",1],
 ["store result as","c"],
 ["if","%n%","=",5],
 ["let","n",500],
 ["end if"],
 ["loop"],
 ["return","%c%"]
]' '' '5'

multicall "ac1" '[
 ["to number","0"],
 ["for","n",500,10,-5],
 ["add","%last result%",1],
 ["if","%n%","=",475],
 ["let","n",2],
 ["end if"],
 ["loop"],
 ["return","%last result%"]
]' '' '6'

multicall "ac1" '[
 ["let","c",0],
 ["for","n",1,10],
 ["add","%c%",1],
 ["store result as","c"],
 ["if","%n%","=",5],
 ["break"],
 ["end if"],
 ["loop"],
 ["return","%c%"]
]' '' '5'

multicall "ac1" '[
 ["let","c",0],
 ["for","n",1,10],
 ["add","%c%",1],
 ["store result as","c"],
 ["break","if","%n%","=",5],
 ["loop"],
 ["return","%c%"]
]' '' '5'

multicall "ac1" '[
 ["to number","0"],
 ["for","n",500,10,-5],
 ["add","%last result%",1],
 ["if","%n%","=",475],
 ["break"],
 ["end if"],
 ["loop"],
 ["return","%last result%"]
]' '' '6'

multicall "ac1" '[
 ["to number","0"],
 ["for","n",500,10,-5],
 ["add","%last result%",1],
 ["break","if","%n%","=",475],
 ["loop"],
 ["return","%last result%"]
]' '' '6'

multicall "ac1" '[
 ["for","n",1,5],
 ["loop"],
 ["return","%n%"]
]' '' '6'

multicall "ac1" '[
 ["for","n",1,5],
 ["break"],
 ["loop"],
 ["return","%n%"]
]' '' '1'

multicall "ac1" '[
 ["let","names",[]],
 ["let","list",["one","two","three","four"]],
 ["for each","item","in","%list%"],
 ["if","%item%","=","three"],
 ["break"],
 ["end if"],
 ["insert","%names%","%item%"],
 ["loop"],
 ["return","%names%"]
]' '' '["one","two"]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","list",["one","two","three","four"]],
 ["for each","item","in","%list%"],
 ["break","if","%item%","=","three"],
 ["insert","%names%","%item%"],
 ["loop"],
 ["return","%names%"]
]' '' '["one","two"]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","obj",{"one":true,"two":false,"three":false,"four":true}],
 ["for each","key","value","in","%obj%"],
 ["if","%value%","=",false],
 ["break"],
 ["end if"],
 ["insert","%names%","%key%"],
 ["loop"],
 ["return","%names%"]
]' '' '["four","one"]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","obj",{"one":true,"two":false,"three":false,"four":true}],
 ["for each","key","value","in","%obj%"],
 ["break","if","%value%","=",false],
 ["insert","%names%","%key%"],
 ["loop"],
 ["return","%names%"]
]' '' '["four","one"]'



# RETURN before the end

multicall "ac1" '[
 ["let","v",123],
 ["if","%v%",">",100],
 ["return"],
 ["end if"],
 ["let","v",500],
 ["return","%v%"]
]' '' ''

multicall "ac1" '[
 ["let","v",123],
 ["if","%v%",">",200],
 ["return"],
 ["end if"],
 ["let","v",500],
 ["return","%v%"]
]' '' '500'



# FULL LOOPS

multicall "ac1" '[
 ["let","c","'$address'"],
 ["call","%c%","inc","n"],
 ["call","%c%","get","n"],
 ["if","%last result%",">=",5],
 ["return","%last result%"],
 ["end if"],
 ["loop"]
]' '' '5'



# CALLS

multicall "ac1" '[
 ["call","'$address'","works"],
 ["assert","%last result%","=",123],
 ["call","'$address'","works"],
 ["return","%last result%"]
]' '' '123'

multicall "ac1" '[
 ["call","'$address'","works"],
 ["call","'$address'","fails"]
]', 'this call should fail'


multicall "ac3" '[
 ["call","'$address'","set_name","test"],
 ["call","'$address'","get_name"],
 ["assert","%last result%","=","test"]
]'

multicall "ac3" '[
 ["call","'$address'","set_name","wrong"],
 ["call","'$address'","get_name"],
 ["assert","%last result%","=","wrong"],
 ["call","'$address'","set_name",123]
]', 'must be string'

multicall "ac3" '[
 ["call","'$address'","get_name"],
 ["assert","%last result%","=","test"],
 ["return","%last result%"]
]' '' '"test"'


multicall "ac3" '[
 ["let","c","'$address'"],
 ["call","%c%","set_name","test2"],
 ["call","%c%","get_name"],
 ["assert","%last result%","=","test2"]
]'

multicall "ac3" '[
 ["call","'$address'","sender"],
 ["assert","%last result%","=","%my account address%"],
 ["call","'$address'","is_contract","%my account address%"],
 ["assert","%last result%","=",false],
 ["call","'$address'","set","account","%my account address%"],
 ["call","'$address'","get","account"],
 ["assert","%last result%","=","%my account address%"]
]'

# CALL + SEND

multicall "ac3" '[
 ["get balance","%my account address%"],
 ["store result as","my balance before"],
 ["get balance","'$address'"],
 ["store result as","contract balance before"],
 ["call + send","0.25 aergo","'$address'","resend_to","%my account address%"],
 ["assert","%last result%","=","250000000000000000"],
 ["get balance","%my account address%"],
 ["assert","%last result%","=","%my balance before%"],
 ["get balance","'$address'"],
 ["assert","%last result%","=","%contract balance before%"]
]'

multicall "ac3" '[
 ["get balance","%my account address%"],
 ["store result as","my balance before"],
 ["get balance","'$address'"],
 ["store result as","contract balance before"],

 ["let","amount","1.5","aergo"],
 ["call + send","%amount%","'$address'","recv_aergo"],

 ["assert","%my aergo balance%","<","%my balance before%"],
 ["get balance","'$address'"],
 ["assert","%last result%",">","%contract balance before%"],

 ["call","'$address'","send_to","%my account address%","%amount%"],

 ["assert","%my aergo balance%","=","%my balance before%"],
 ["get balance","'$address'"],
 ["assert","%last result%","=","%contract balance before%"]
]'


# CALL LOOP

multicall "ac3" '[
 ["let","list",["first","second","third"]],
 ["for each","item","in","%list%"],
 ["call","'$address'","add","%item%"],
 ["loop"]
]'

multicall "ac1" '[
 ["call","'$address'","get","1"],
 ["assert","%last result%","=","first"],
 ["call","'$address'","get","2"],
 ["assert","%last result%","=","second"],
 ["call","'$address'","get","3"],
 ["assert","%last result%","=","third"]
]'

multicall "ac3" '[
 ["let","list",["1st","2nd","3rd"]],
 ["let","n",1],
 ["for each","item","in","%list%"],
 ["to string","%n%"],
 ["call","'$address'","set","%last result%","%item%"],
 ["add","%n%",1],
 ["store result as","n"],
 ["loop"]
]'

multicall "ac1" '[
 ["call","'$address'","get","1"],
 ["assert","%last result%","=","1st"],
 ["call","'$address'","get","2"],
 ["assert","%last result%","=","2nd"],
 ["call","'$address'","get","3"],
 ["assert","%last result%","=","3rd"]
]'



# TRY CALL

multicall "ac1" '[
 ["try call","'$address'","works"],
 ["assert","%call succeeded%","=",true],
 ["try call","'$address'","fails"],
 ["assert","%call succeeded%","=",false]
]'

multicall "ac3" '[
 ["try call","'$address'","set_name","1st"],
 ["assert","%call succeeded%"],

 ["try call","'$address'","get_name"],
 ["assert","%call succeeded%","=",true],
 ["assert","%last result%","=","1st"],

 ["try call","'$address'","set_name",22],
 ["assert","%call succeeded%","=",false],

 ["try call","'$address'","get_name"],
 ["assert","%call succeeded%","=",true],
 ["assert","%last result%","=","1st"],

 ["return","%last result%"]
]' '' '"1st"'


# TRY CALL + SEND

multicall "ac1" '[
 ["get balance","'$address'"],
 ["store result as","balance before"],
 ["try call + send","0.25 aergo","'$address'","recv_aergo"],
 ["assert","%call succeeded%"],
 ["try call + send","1 aergo","'$address'","recv_aergo"],
 ["assert","%call succeeded%","=",true],
 ["get balance","'$address'"],
 ["store result as","balance after"],
 ["subtract","%balance after%","%balance before%"],
 ["assert","%last result%","=","1.25 aergo"],
 ["try call + send","1 aergo","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRB","recv_aergo"],
 ["assert","%call succeeded%","=",false]
]'


# MULTICALL ON ACCOUNT ------------------------------------------


# deploy ac0 0 c1 test.lua
# deploy ac0 0 c2 test.lua
# deploy ac0 0 c3 test.lua

# c1: AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9
# c2: Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4
# c3: AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK

multicall "ac0" '[
 ["call","'$address1'","set_name","testing multicall"],
 ["call","'$address2'","set_name","contract 2"],
 ["call","'$address3'","set_name","third one"]
]'

multicall "ac0" '[
 ["call","'$address1'","get_name"],
 ["assert","%last result%","=","testing multicall"],
 ["store result as","r1"],
 ["call","'$address2'","get_name"],
 ["assert","%last result%","=","contract 2"],
 ["store result as","r2"],
 ["call","'$address3'","get_name"],
 ["assert","%last result%","=","third one"],
 ["store result as","r3"],
 ["return","%r1%","%r2%","%r3%"]
]' '' '["testing multicall","contract 2","third one"]'

multicall "ac0" '[
 ["from json","{}"],
 ["store result as","res"],
 ["call","'$address1'","get_name"],
 ["set","%res%","r1","%last result%"],
 ["call","'$address2'","get_name"],
 ["set","%res%","r2","%last result%"],
 ["call","'$address3'","get_name"],
 ["set","%res%","r3","%last result%"],
 ["return","%res%"]
]' '' '{"r1":"testing multicall","r2":"contract 2","r3":"third one"}'


multicall "ac0" '[
 ["call","'$address1'","set_name","wohooooooo"],
 ["call","'$address2'","set_name","it works!"],
 ["call","'$address3'","set_name","it really works!"]
]'

multicall "ac0" '[
 ["call","'$address1'","get_name"],
 ["assert","%last result%","=","wohooooooo"],
 ["store result as","r1"],
 ["call","'$address2'","get_name"],
 ["assert","%last result%","=","it works!"],
 ["store result as","r2"],
 ["call","'$address3'","get_name"],
 ["assert","%last result%","=","it really works!"],
 ["store result as","r3"],
 ["return","%r1%","%r2%","%r3%"]
]' '' '["wohooooooo","it works!","it really works!"]'

multicall "ac0" '[
 ["from json","{}"],
 ["store result as","res"],
 ["call","'$address1'","get_name"],
 ["set","%res%","r1","%last result%"],
 ["call","'$address2'","get_name"],
 ["set","%res%","r2","%last result%"],
 ["call","'$address3'","get_name"],
 ["set","%res%","r3","%last result%"],
 ["return","%res%"]
]' '' '{"r1":"wohooooooo","r2":"it works!","r3":"it really works!"}'



# aergo BALANCE and SEND

parse_balance() {
  local number=$1
  number=${number/ aergo/}
  local result=$(echo "scale=0; $number * 1000000000000000000 / 1" | bc)
  echo $result
}

# 0: AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R
# 1: AmNLhiVVdLxbPW5NxpLibqoLdobc2TaKVk8bwrPn5VXz6gUSLvke
# 2: AmLhv6rvLEMoL5MLws1tNTyiYYyzTx4JGjaPAugzqabpjxxWyP34
# 3: AmPfKfrbSjbHD77JjpEvdktgH3w6sbWRhhwFmFAyi1ZSsiePs7XP
# 4: AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw

account_state=$(../bin/aergocli getstate --address AmNLhiVVdLxbPW5NxpLibqoLdobc2TaKVk8bwrPn5VXz6gUSLvke)
echo $account_state
balance1=$(echo $account_state | jq .balance | sed 's/"//g')
balance1=$(parse_balance $balance1)
echo balance=$balance1

account_state=$(../bin/aergocli getstate --address AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw)
echo $account_state
balance4=$(echo $account_state | jq .balance | sed 's/"//g')
balance4=$(parse_balance $balance4)
echo balance=$balance4

multicall "ac4" '[
 ["get balance"],
 ["to string"],
 ["assert","%last result%","=","'$balance4'"],
 ["return","%last result%"]
]' '' '"'$balance4'"'

# retrieve it again because it decreased due to tx fee
account_state=$(../bin/aergocli getstate --address AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw)
echo $account_state
balance4=$(echo $account_state | jq .balance | sed 's/"//g')
balance4=$(parse_balance $balance4)
echo balance=$balance4

amount=1230000000000000000

balance4after=$(echo "$balance4 + $amount" | bc)
balance1after=$(echo "$balance1 - $amount" | bc)

multicall "ac1" '[
 ["get balance"],
 ["to string","%last result%"],
 ["assert","%last result%","=","'$balance1'"],

 ["get balance","AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw"],
 ["to string","%last result%"],
 ["assert","%last result%","=","'$balance4'"],

 ["send","AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw","'$amount'"],

 ["get balance","AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw"],
 ["to string","%last result%"],
 ["assert","%last result%","=","'$balance4after'"],

 ["get balance"],
 ["to string"],
 ["assert","%last result%","=","'$balance1after'"],

 ["return","%last result%"]
]' '' '"'$balance1after'"'
