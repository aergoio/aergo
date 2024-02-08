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
 ["store","dict"],
 ["set","%dict%","two",22],
 ["set","%dict%","four",4],
 ["set","%dict%","one",null],
 ["get","%dict%","two"],
 ["set","%dict%","copy","%last_result%"],
 ["return","%dict%"]
]' '' '{"copy":22,"four":4,"three":3,"two":22}'

multicall "ac1" '[
 ["call","'$address'","get_list"],
 ["store","array"],
 ["set","%array%",2,"2nd"],
 ["insert","%array%",1,"zero"],
 ["insert","%array%","last"],
 ["return","%array%"]
]' '' '["zero","first","2nd","third",123,12.5,true,"last"]'

multicall "ac1" '[
 ["call","'$address'","get_list"],
 ["store","array"],
 ["remove","%array%",3],
 ["return","%array%","%last_result%"]
]' '' '[["first","second",123,12.5,true],"third"]'


# create new dict or array using fromjson

multicall "ac1" '[
 ["fromjson","{\"one\":1,\"two\":2}"],
 ["set","%last_result%","three",3],
 ["return","%last_result%"]
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
 ["assert","%last_result%","=","one"],
 ["get","%list%",2],
 ["assert","%last_result%","=",22],
 ["get","%list%",3],
 ["assert","%last_result%","=",3.3],
 ["get","%list%",4],
 ["assert","%last_result%","=",true],
 ["get","%list%",5],
 ["assert","%last_result%","=",false],
 ["return","%list%"]
]' '' '["one",22,3.3,true,false]'


# get_size

multicall "ac1" '[
 ["let","str","this is a string"],
 ["get_size","%str%"],
 ["return","%last_result%"]
]' '' '16'

multicall "ac1" '[
 ["let","list",["one",1,"two",2,2.5,true,false]],
 ["get_size","%list%"],
 ["return","%last_result%"]
]' '' '7'

multicall "ac1" '[
 ["let","obj",{"one":1,"two":2,"three":3}],
 ["get_size","%obj%"],
 ["return","%last_result%"]
]' '' '0'


# get_keys

multicall "ac1" '[
 ["let","obj",{"one":1,"two":2,"three":3}],
 ["get_keys","%obj%"],
 ["store","keys"],
 ["get_size","%keys%"],
 ["return","%last_result%","%keys%"]
]' '' '[3,["one","three","two"]]'




# BIGNUM

multicall "ac1" '[
 ["tobignum",123],
 ["store","a"],
 ["tobignum",123],
 ["store","b"],
 ["mul","%a%","%b%"],
 ["return","%last_result%"]
]' '' '{"_bignum":"15129"}'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","100000"],
 ["store","b"],
 ["div","%a%","%b%"],
 ["return","%last_result%"]
]' '' '{"_bignum":"5000000000000000"}'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","100000"],
 ["store","b"],
 ["div","%a%","%b%"],
 ["tostring","%last_result%"],
 ["return","%last_result%"]
]' '' '"5000000000000000"'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],

 ["tobignum","100000"],
 ["div","%a%","%last_result%"],
 ["store","a"],

 ["tobignum","1000000000000000"],
 ["sub","%a%","%last_result%"],
 ["store","a"],

 ["tobignum","1234"],
 ["add","%a%","%last_result%"],
 ["store","a"],

 ["tobignum","2"],
 ["pow","%a%","%last_result%"],
 ["sqrt","%last_result%"],
 ["store","a"],

 ["tobignum","2"],
 ["mod","%a%","10000"],

 ["return","%last_result%"]
]' '' '{"_bignum":"1234"}'

multicall "ac1" '[
 ["let","a",25],
 ["sqrt","%a%"],
 ["return","%last_result%"]
]' '' '{"_bignum":"5"}'

multicall "ac1" '[
 ["let","a",25],
 ["pow","%a%",0.5],
 ["return","%last_result%"]
]' '' '5'



# STRINGS

multicall "ac1" '[
 ["format","%s%s%s","hello"," ","world"],
 ["return","%last_result%"]
]' '' '"hello world"'

multicall "ac1" '[
 ["let","s","hello world"],
 ["substr","%s%",1,4],
 ["return","%last_result%"]
]' '' '"hell"'

multicall "ac1" '[
 ["let","s","hello world"],
 ["substr","%s%",-2,-1],
 ["return","%last_result%"]
]' '' '"ld"'

multicall "ac1" '[
 ["let","s","the amount is 12345"],
 ["find","%s%","%d+"],
 ["tonumber","%last_result%"],
 ["return","%last_result%"]
]' '' '12345'

multicall "ac1" '[
 ["let","s","rate: 55 10%"],
 ["find","%s%","(%d+)%%"],
 ["tonumber","%last_result%"],
 ["return","%last_result%"]
]' '' '10'

multicall "ac1" '[
 ["let","s","rate: 12%"],
 ["find","%s%","%s*(%d+)%%"],
 ["tonumber","%last_result%"],
 ["return","%last_result%"]
]' '' '12'

multicall "ac1" '[
 ["let","s","hello world"],
 ["replace","%s%","hello","good bye"],
 ["return","%last_result%"]
]' '' '"good bye world"'

multicall "ac1" '[
 ["fromjson","{\"name\":\"ticket\",\"value\":12.5,\"amount\":10}"],
 ["replace","name = $name, value = $value, amount = $amount","%$(%w+)","%last_result%"],
 ["return","%last_result%"]
]' '' '"name = ticket, value = 12.5, amount = 10"'


# IF THEN ELSE

multicall "ac1" '[
 ["let","s",20],
 ["if","%s%",">=",20],
 ["let","b","big"],
 ["elif","%s%",">=",10],
 ["let","b","medium"],
 ["else"],
 ["let","b","low"],
 ["end"],
 ["let","c","after"],
 ["return","%b%","%c%"]
]' '' '["big","after"]'

multicall "ac1" '[
 ["let","s",10],
 ["if","%s%",">=",20],
 ["let","b","big"],
 ["elif","%s%",">=",10],
 ["let","b","medium"],
 ["else"],
 ["let","b","low"],
 ["end"],
 ["let","c","after"],
 ["return","%b%","%c%"]
]' '' '["medium","after"]'

multicall "ac1" '[
 ["let","s",5],
 ["if","%s%",">=",20],
 ["let","b","big"],
 ["elif","%s%",">=",10],
 ["let","b","medium"],
 ["else"],
 ["let","b","low"],
 ["end"],
 ["let","c","after"],
 ["return","%b%","%c%"]
]' '' '["low","after"]'

multicall "ac1" '[
 ["let","s",20],
 ["if","%s%",">=",20],
 ["return","big"],
 ["elif","%s%",">=",10],
 ["return","medium"],
 ["else"],
 ["return","low"],
 ["end"],
 ["return","after"]
]' '' '"big"'

multicall "ac1" '[
 ["let","s",10],
 ["if","%s%",">=",20],
 ["return","big"],
 ["elif","%s%",">=",10],
 ["return","medium"],
 ["else"],
 ["return","low"],
 ["end"],
 ["return","after"]
]' '' '"medium"'

multicall "ac1" '[
 ["let","s",5],
 ["if","%s%",">=",20],
 ["return","big"],
 ["elif","%s%",">=",10],
 ["return","medium"],
 ["else"],
 ["return","low"],
 ["end"],
 ["return","after"]
]' '' '"low"'


multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","500000000000000000000"],
 ["store","b"],
 ["if","%a%","=","%b%"],
 ["let","b","equal"],
 ["else"],
 ["let","b","diff"],
 ["end"],
 ["return","%b%"]
]' '' '"equal"'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","500000000000000000001"],
 ["store","b"],
 ["if","%a%","=","%b%"],
 ["let","b","equal"],
 ["else"],
 ["let","b","diff"],
 ["end"],
 ["return","%b%"]
]' '' '"diff"'

multicall "ac1" '[
 ["tobignum","500000000000000000001"],
 ["store","a"],
 ["tobignum","500000000000000000000"],
 ["store","b"],
 ["if","%a%",">","%b%"],
 ["let","b","bigger"],
 ["else"],
 ["let","b","lower"],
 ["end"],
 ["return","%b%"]
]' '' '"bigger"'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","500000000000000000001"],
 ["store","b"],
 ["if","%a%",">","%b%"],
 ["let","b","bigger"],
 ["else"],
 ["let","b","lower"],
 ["end"],
 ["return","%b%"]
]' '' '"lower"'


multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","500000000000000000001"],
 ["store","b"],
 ["if","%a%","<","%b%","and","1","=","0"],
 ["let","b","wrong 1"],
 ["elif","%a%","<","%b%","and","1","=","1"],
 ["let","b","correct"],
 ["else"],
 ["let","b","wrong 2"],
 ["end"],
 ["return","%b%"]
]' '' '"correct"'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","500000000000000000001"],
 ["store","b"],
 ["if","%a%","<","%b%","and",1,"=",0],
 ["let","b","wrong 1"],
 ["elif","%a%","<","%b%","and",1,"=",1],
 ["let","b","correct"],
 ["else"],
 ["let","b","wrong 2"],
 ["end"],
 ["return","%b%"]
]' '' '"correct"'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","500000000000000000001"],
 ["store","b"],
 ["tobignum","400000000000000000000"],
 ["store","c"],
 ["if","%a%","<","%b%","and","%a%","<","%c%"],
 ["let","b","wrong 1"],
 ["elif","%a%",">","%b%","and","%a%",">","%c%"],
 ["let","b","wrong 2"],
 ["elif","%a%","<","%b%","and","%a%",">","%c%"],
 ["let","b","correct"],
 ["else"],
 ["let","b","wrong 3"],
 ["end"],
 ["return","%b%"]
]' '' '"correct"'

multicall "ac1" '[
 ["tobignum","500000000000000000000"],
 ["store","a"],
 ["tobignum","500000000000000000001"],
 ["store","b"],
 ["tobignum","400000000000000000000"],
 ["store","c"],
 ["tostring",0],

 ["if","%a%",">","%b%","or","%a%","<","%c%"],
 ["format","%s%s","%last_result%","1"],
 ["end"],

 ["if","%a%","=","%b%","or","%a%","=","%c%"],
 ["format","%s%s","%last_result%","2"],
 ["end"],

 ["if","%a%","<","%b%","or","%a%","<","%c%"],
 ["format","%s%s","%last_result%","3"],
 ["end"],

 ["if","%a%",">","%b%","or","%a%",">","%c%"],
 ["format","%s%s","%last_result%","4"],
 ["end"],

 ["if","%a%","<","%b%","or","%a%",">","%c%"],
 ["format","%s%s","%last_result%","5"],
 ["end"],

 ["if","%a%","!=","%b%","or","%a%","=","%c%"],
 ["format","%s%s","%last_result%","6"],
 ["end"],

 ["if","%a%","=","%b%","or","%a%","!=","%c%"],
 ["format","%s%s","%last_result%","7"],
 ["end"],


 ["if","%a%",">=","%b%","and","%a%","<=","%c%"],
 ["format","%s%s","%last_result%","8"],
 ["end"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%"],
 ["format","%s%s","%last_result%","9"],
 ["end"],

 ["if","%b%",">=","%a%","and","%b%","<=","%c%"],
 ["format","%s%s","%last_result%","A"],
 ["end"],

 ["if","%b%",">=","%c%","and","%b%","<=","%a%"],
 ["format","%s%s","%last_result%","B"],
 ["end"],

 ["if","%c%",">=","%a%","and","%c%","<=","%b%"],
 ["format","%s%s","%last_result%","C"],
 ["end"],

 ["if","%c%",">=","%b%","and","%c%","<=","%a%"],
 ["format","%s%s","%last_result%","D"],
 ["end"],


 ["if","%a%",">=","%b%","and","%a%","<=","%c%","or",1,"=",0],
 ["format","%s%s","%last_result%","E"],
 ["end"],

 ["if","%a%",">=","%b%","and","%a%","<=","%c%","or",1,"=",1],
 ["format","%s%s","%last_result%","F"],
 ["end"],

 ["if","%a%",">=","%b%","and","%a%","<=","%c%","and",1,"=",0],
 ["format","%s%s","%last_result%","G"],
 ["end"],

 ["if","%a%",">=","%b%","and","%a%","<=","%c%","and",1,"=",1],
 ["format","%s%s","%last_result%","H"],
 ["end"],


 ["if","%a%",">=","%c%","and","%a%","<=","%b%","or",1,"=",0],
 ["format","%s%s","%last_result%","I"],
 ["end"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%","or",1,"=",1],
 ["format","%s%s","%last_result%","J"],
 ["end"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%","and",1,"=",0],
 ["format","%s%s","%last_result%","K"],
 ["end"],

 ["if","%a%",">=","%c%","and","%a%","<=","%b%","and",1,"=",1],
 ["format","%s%s","%last_result%","L"],
 ["end"],


 ["if",1,"=",0,"or","%a%",">=","%b%","and","%a%","<=","%c%"],
 ["format","%s%s","%last_result%","M"],
 ["end"],

 ["if",1,"=",1,"or","%a%",">=","%b%","and","%a%","<=","%c%"],
 ["format","%s%s","%last_result%","N"],
 ["end"],

 ["if",1,"=",0,"or","%a%",">=","%c%","and","%a%","<=","%b%"],
 ["format","%s%s","%last_result%","O"],
 ["end"],

 ["if",1,"=",1,"or","%a%",">=","%c%","and","%a%","<=","%b%"],
 ["format","%s%s","%last_result%","P"],
 ["end"],


 ["return","%last_result%"]
]' '' '"0345679IJLNOP"'




# FOR

multicall "ac1" '[
 ["for","n",1,5],
 ["loop"],
 ["return","%n%"]
]' '' '6'

multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",1,5],
 ["add","%last_result%",1],
 ["loop"],
 ["return","%last_result%"]
]' '' '5'

multicall "ac1" '[
 ["tobignum","10000000000000000001"],
 ["store","to_add"],
 ["tobignum","100000000000000000000"],

 ["for","n",1,3],
 ["add","%last_result%","%to_add%"],
 ["loop"],

 ["tostring","%last_result%"],
 ["return","%last_result%"]
]' '' '"130000000000000000003"'


multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",500,10,-5],
 ["add","%last_result%",1],
 ["loop"],
 ["return","%last_result%"]
]' '' '99'


multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",5,1,-1],
 ["add","%last_result%",1],
 ["loop"],
 ["return","%last_result%"]
]' '' '5'

multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",5,1],
 ["add","%last_result%",1],
 ["loop"],
 ["return","%last_result%"]
]' '' '0'

multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",1,5],
 ["add","%last_result%",1],
 ["loop"],
 ["return","%last_result%"]
]' '' '5'

multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",1,5,-1],
 ["add","%last_result%",1],
 ["loop"],
 ["return","%last_result%"]
]' '' '0'



# FOREACH

multicall "ac1" '[
 ["let","list",[11,22,33]],
 ["let","r",0],
 ["foreach","item","%list%"],
 ["add","%r%","%item%"],
 ["store","r"],
 ["loop"],
 ["return","%r%"]
]' '' '66'

multicall "ac1" '[
 ["let","list",[11,22,33]],
 ["let","counter",0],
 ["foreach","item","%list%"],
 ["add","%counter%",1],
 ["store","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '3'

multicall "ac1" '[
 ["let","list",[]],
 ["let","counter",0],
 ["foreach","item","%list%"],
 ["add","%counter%",1],
 ["store","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '0'

multicall "ac1" '[
 ["let","list",["one",1,"two",2,2.5,true,false]],
 ["let","counter",0],
 ["foreach","item","%list%"],
 ["add","%counter%",1],
 ["store","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '7'

multicall "ac1" '[
 ["let","list",[10,21,32]],
 ["let","r",0],
 ["foreach","item","%list%"],
 ["if","%item%","<",30],
 ["add","%r%","%item%"],
 ["store","r"],
 ["end"],
 ["loop"],
 ["return","%r%"]
]' '' '31'


multicall "ac1" '[
 ["let","str",""],
 ["let","obj",{"one":1,"two":2,"three":3}],
 ["get_keys","%obj%"],
 ["foreach","key","%last_result%"],
 ["concat","%str%","%key%"],
 ["store","str"],
 ["loop"],
 ["return","%str%"]
]' '' '"onethreetwo"'



# FORPAIR

multicall "ac1" '[
 ["let","str",""],
 ["let","sum",0],
 ["let","obj",{"one":1,"two":2,"three":3}],
 ["forpair","key","value","%obj%"],
 ["concat","%str%","%key%"],
 ["store","str"],
 ["add","%sum%","%value%"],
 ["store","sum"],
 ["loop"],
 ["return","%str%","%sum%"]
]' '' '["onethreetwo",6]'

multicall "ac1" '[
 ["let","str",""],
 ["let","sum",0],
 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
 ["forpair","key","value","%obj%"],
 ["concat","%str%","%key%"],
 ["store","str"],
 ["add","%sum%","%value%"],
 ["store","sum"],
 ["loop"],
 ["return","%str%","%sum%"]
]' '' '["fouronethreetwo",12]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","values",[]],
 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
 ["forpair","key","value","%obj%"],
 ["insert","%names%","%key%"],
 ["insert","%values%","%value%"],
 ["loop"],
 ["return","%names%","%values%"]
]' '' '[["four","one","three","two"],[4.5,1.5,3.5,2.5]]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","values",[]],
 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
 ["forpair","key","value","%obj%"],
 ["insert","%names%","%key%"],
 ["insert","%values%","%value%"],
 ["loop"],
 ["call","'$address'","sort","%values%"],
 ["store","values"],
 ["return","%names%","%values%"]
]' '' '[["four","one","three","two"],[1.5,2.5,3.5,4.5]]'

multicall "ac1" '[
 ["let","obj",{}],
 ["let","counter",0],
 ["forpair","key","value","%obj%"],
 ["add","%counter%",1],
 ["store","counter"],
 ["loop"],
 ["return","%counter%"]
]' '' '0'



# FOR "BREAK"

multicall "ac1" '[
 ["let","c",0],
 ["for","n",1,10],
 ["add","%c%",1],
 ["store","c"],
 ["if","%n%","=",5],
 ["let","n",500],
 ["end"],
 ["loop"],
 ["return","%c%"]
]' '' '5'

multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",500,10,-5],
 ["add","%last_result%",1],
 ["if","%n%","=",475],
 ["let","n",2],
 ["end"],
 ["loop"],
 ["return","%last_result%"]
]' '' '6'

multicall "ac1" '[
 ["let","c",0],
 ["for","n",1,10],
 ["add","%c%",1],
 ["store","c"],
 ["if","%n%","=",5],
 ["break"],
 ["end"],
 ["loop"],
 ["return","%c%"]
]' '' '5'

multicall "ac1" '[
 ["let","c",0],
 ["for","n",1,10],
 ["add","%c%",1],
 ["store","c"],
 ["break","if","%n%","=",5],
 ["loop"],
 ["return","%c%"]
]' '' '5'

multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",500,10,-5],
 ["add","%last_result%",1],
 ["if","%n%","=",475],
 ["break"],
 ["end"],
 ["loop"],
 ["return","%last_result%"]
]' '' '6'

multicall "ac1" '[
 ["tonumber","0"],
 ["for","n",500,10,-5],
 ["add","%last_result%",1],
 ["break","if","%n%","=",475],
 ["loop"],
 ["return","%last_result%"]
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
 ["foreach","item","%list%"],
 ["if","%item%","=","three"],
 ["break"],
 ["end"],
 ["insert","%names%","%item%"],
 ["loop"],
 ["return","%names%"]
]' '' '["one","two"]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","list",["one","two","three","four"]],
 ["foreach","item","%list%"],
 ["break","if","%item%","=","three"],
 ["insert","%names%","%item%"],
 ["loop"],
 ["return","%names%"]
]' '' '["one","two"]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","obj",{"one":true,"two":false,"three":false,"four":true}],
 ["forpair","key","value","%obj%"],
 ["if","%value%","=",false],
 ["break"],
 ["end"],
 ["insert","%names%","%key%"],
 ["loop"],
 ["return","%names%"]
]' '' '["four","one"]'

multicall "ac1" '[
 ["let","names",[]],
 ["let","obj",{"one":true,"two":false,"three":false,"four":true}],
 ["forpair","key","value","%obj%"],
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
 ["end"],
 ["let","v",500],
 ["return","%v%"]
]' '' ''

multicall "ac1" '[
 ["let","v",123],
 ["if","%v%",">",200],
 ["return"],
 ["end"],
 ["let","v",500],
 ["return","%v%"]
]' '' '500'



# FULL LOOPS

multicall "ac1" '[
 ["let","c","'$address'"],
 ["call","%c%","inc","n"],
 ["call","%c%","get","n"],
 ["if","%last_result%",">=",5],
 ["return","%last_result%"],
 ["end"],
 ["loop"]
]' '' '5'



# CALLS

multicall "ac1" '[
 ["call","'$address'","works"],
 ["assert","%last_result%","=",123],
 ["call","'$address'","works"],
 ["return","%last_result%"]
]' '' '123'

multicall "ac1" '[
 ["call","'$address'","works"],
 ["call","'$address'","fails"]
]', 'this call should fail'


multicall "ac3" '[
 ["call","'$address'","set_name","test"],
 ["call","'$address'","get_name"],
 ["assert","%last_result%","=","test"]
]'

multicall "ac3" '[
 ["call","'$address'","set_name","wrong"],
 ["call","'$address'","get_name"],
 ["assert","%last_result%","=","wrong"],
 ["call","'$address'","set_name",123]
]', 'must be string'

multicall "ac3" '[
 ["call","'$address'","get_name"],
 ["assert","%last_result%","=","test"],
 ["return","%last_result%"]
]' '' '"test"'


multicall "ac3" '[
 ["let","c","'$address'"],
 ["call","%c%","set_name","test2"],
 ["call","%c%","get_name"],
 ["assert","%last_result%","=","test2"]
]'


# CALL LOOP

multicall "ac3" '[
 ["let","list",["first","second","third"]],
 ["foreach","item","%list%"],
 ["call","'$address'","add","%item%"],
 ["loop"]
]'

multicall "ac1" '[
 ["call","'$address'","get","1"],
 ["assert","%last_result%","=","first"],
 ["call","'$address'","get","2"],
 ["assert","%last_result%","=","second"],
 ["call","'$address'","get","3"],
 ["assert","%last_result%","=","third"]
]'

multicall "ac3" '[
 ["let","list",["1st","2nd","3rd"]],
 ["let","n",1],
 ["foreach","item","%list%"],
 ["tostring","%n%"],
 ["call","'$address'","set","%last_result%","%item%"],
 ["add","%n%",1],
 ["store","n"],
 ["loop"]
]'

multicall "ac1" '[
 ["call","'$address'","get","1"],
 ["assert","%last_result%","=","1st"],
 ["call","'$address'","get","2"],
 ["assert","%last_result%","=","2nd"],
 ["call","'$address'","get","3"],
 ["assert","%last_result%","=","3rd"]
]'



# PCALL

multicall "ac1" '[
 ["pcall","'$address'","works"],
 ["get","%last_result%",1],
 ["assert","%last_result%","=",true],
 ["pcall","'$address'","fails"],
 ["get","%last_result%",1],
 ["assert","%last_result%","=",false]
]'

multicall "ac3" '[
 ["pcall","'$address'","set_name","1st"],
 ["get","%last_result%",1],
 ["assert","%last_result%","=",true],

 ["pcall","'$address'","get_name"],
 ["store","ret"],
 ["get","%ret%",1],
 ["assert","%last_result%","=",true],
 ["get","%ret%",2],
 ["assert","%last_result%","=","1st"],

 ["pcall","'$address'","set_name",22],
 ["get","%last_result%",1],
 ["assert","%last_result%","=",false],

 ["pcall","'$address'","get_name"],
 ["store","ret"],
 ["get","%ret%",1],
 ["assert","%last_result%","=",true],
 ["get","%ret%",2],
 ["assert","%last_result%","=","1st"],

 ["return","%last_result%"]
]' '' '"1st"'



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
 ["assert","%last_result%","=","testing multicall"],
 ["store","r1"],
 ["call","'$address2'","get_name"],
 ["assert","%last_result%","=","contract 2"],
 ["store","r2"],
 ["call","'$address3'","get_name"],
 ["assert","%last_result%","=","third one"],
 ["store","r3"],
 ["return","%r1%","%r2%","%r3%"]
]' '' '["testing multicall","contract 2","third one"]'

multicall "ac0" '[
 ["fromjson","{}"],
 ["store","res"],
 ["call","'$address1'","get_name"],
 ["set","%res%","r1","%last_result%"],
 ["call","'$address2'","get_name"],
 ["set","%res%","r2","%last_result%"],
 ["call","'$address3'","get_name"],
 ["set","%res%","r3","%last_result%"],
 ["return","%res%"]
]' '' '{"r1":"testing multicall","r2":"contract 2","r3":"third one"}'


multicall "ac0" '[
 ["call","'$address1'","set_name","wohooooooo"],
 ["call","'$address2'","set_name","it works!"],
 ["call","'$address3'","set_name","it really works!"]
]'

multicall "ac0" '[
 ["call","'$address1'","get_name"],
 ["assert","%last_result%","=","wohooooooo"],
 ["store","r1"],
 ["call","'$address2'","get_name"],
 ["assert","%last_result%","=","it works!"],
 ["store","r2"],
 ["call","'$address3'","get_name"],
 ["assert","%last_result%","=","it really works!"],
 ["store","r3"],
 ["return","%r1%","%r2%","%r3%"]
]' '' '["wohooooooo","it works!","it really works!"]'

multicall "ac0" '[
 ["fromjson","{}"],
 ["store","res"],
 ["call","'$address1'","get_name"],
 ["set","%res%","r1","%last_result%"],
 ["call","'$address2'","get_name"],
 ["set","%res%","r2","%last_result%"],
 ["call","'$address3'","get_name"],
 ["set","%res%","r3","%last_result%"],
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
 ["balance"],
 ["tostring","%last_result%"],
 ["assert","%last_result%","=","'$balance4'"],
 ["return","%last_result%"]
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
 ["balance"],
 ["tostring","%last_result%"],
 ["assert","%last_result%","=","'$balance1'"],

 ["balance","AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw"],
 ["tostring","%last_result%"],
 ["assert","%last_result%","=","'$balance4'"],

 ["send","AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw","'$amount'"],

 ["balance","AmLaPgDNg3tsebXSU19bftkr1XxvmySWGusEti9SaHoKDJEZNjSw"],
 ["tostring","%last_result%"],
 ["assert","%last_result%","=","'$balance4after'"],

 ["balance"],
 ["tostring","%last_result%"],
 ["assert","%last_result%","=","'$balance1after'"],

 ["return","%last_result%"]
]' '' '"'$balance1after'"'
