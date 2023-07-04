set -e
source common.sh

fork_version=$1


echo "-- deploy --"

../bin/aergoluac --payload ../contract/vm_dummy/test_files/gas_per_function.lua > payload.out

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    --payload `cat payload.out` | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- get account's nonce --"

account_state=$(../bin/aergocli getstate --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R)
nonce=$(echo $account_state | jq .nonce | sed 's/"//g')


# create an iterable map of function name -> expected gas
declare -a names
declare -A gas

add_test() {
  names+=("$1")
  gas["$1"]="$2"
}

add_test "comp_ops" 134740
add_test "unarytest_n_copy_ops" 134653
add_test "unary_ops" 139660
add_test "binary_ops" 136575
add_test "constant_ops" 134568
add_test "upvalue_n_func_ops" 135847
add_test "table_ops" 139894
add_test "call_n_vararg_ops" 136501
add_test "return_ops" 134573
add_test "loop_n_branche_ops" 137908
add_test "function_header_ops" 134552

add_test "assert" 134682
add_test "getfenv" 134577
add_test "metatable" 135488
add_test "ipairs" 134575
add_test "pairs" 134575
add_test "next" 134623
add_test "rawequal" 134752
add_test "rawget" 134623
add_test "rawset" 135441
add_test "select" 134702
add_test "setfenv" 134612
add_test "tonumber" 134686
add_test "tostring" 134957
add_test "type" 134785
add_test "unpack" 142245
add_test "pcall" 137665
add_test "xpcall" 137937

add_test "string.byte" 148540
add_test "string.char" 151897
add_test "string.dump" 138366
add_test "string.find" 139344
add_test "string.format" 135264
add_test "string.gmatch" 135299
add_test "string.gsub" 136443
add_test "string.len" 134633
add_test "string.lower" 139851
add_test "string.match" 134813
add_test "string.rep" 213428
add_test "string.reverse" 139851
add_test "string.sub" 136705
add_test "string.upper" 139851

add_test "table.concat" 155368
add_test "table.insert" 288754
add_test "table.remove" 148164
add_test "table.maxn" 139462
add_test "table.sort" 151366

add_test "math.abs" 134720
add_test "math.ceil" 134720
add_test "math.floor" 134720
add_test "math.max" 135092
add_test "math.min" 135092
add_test "math.pow" 135080

add_test "bit.tobit" 134615
add_test "bit.tohex" 135090
add_test "bit.bnot" 134592
add_test "bit.bor" 134666
add_test "bit.band" 134642
add_test "bit.xor" 134642
add_test "bit.lshift" 134615
add_test "bit.rshift" 134615
add_test "bit.ashift" 134615
add_test "bit.rol" 134615
add_test "bit.ror" 134615
add_test "bit.bswap" 134572

add_test "bignum.number" 136412
add_test "bignum.isneg" 136644
add_test "bignum.iszero" 136644
add_test "bignum.tonumber" 136964
add_test "bignum.tostring" 137255
add_test "bignum.neg" 138603
add_test "bignum.sqrt" 139479
add_test "bignum.compare" 136804
add_test "bignum.add" 138145
add_test "bignum.sub" 138090
add_test "bignum.mul" 140468
add_test "bignum.div" 139958
add_test "bignum.mod" 141893
add_test "bignum.pow" 140887
add_test "bignum.divmod" 146193
add_test "bignum.powmod" 145559
add_test "bignum.operators" 138811

add_test "json" 142320

add_test "crypto.sha256" 137578
add_test "crypto.ecverify" 139467

add_test "state.set" 137059
add_test "state.get" 137115
add_test "state.delete" 137122

add_test "system.getSender" 135656
add_test "system.getBlockheight" 134761
add_test "system.getTxhash" 135132
add_test "system.getTimestamp" 134761
add_test "system.getContractID" 135656
add_test "system.setItem" 135589
add_test "system.getItem" 135898
add_test "system.getAmount" 134803
add_test "system.getCreator" 135156
add_test "system.getOrigin" 135656

add_test "contract.send" 135716
add_test "contract.balance" 135797
add_test "contract.deploy" 158752
add_test "contract.call" 149642
add_test "contract.pcall" 150563
add_test "contract.delegatecall" 144902
add_test "contract.event" 153263

# as the returned value differs in length (43 or 44)
# due to base58, the computed gas is different.
#add_test "system.getPrevBlockHash" 135132

# contract.balance() also use diff gas
# according to the returned string size

declare -A txhashes
i=0

echo "-- send the transactions --"

for function_name in "${names[@]}"; do
  echo -n "."
  i=$(($i+1))
  txhash=$(../bin/aergocli --keystore . --password bmttest --nonce $(($nonce+$i)) \
    contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    $address run_test "[\"$function_name\"]" | jq .hash | sed 's/"//g')
  txhashes[$function_name]=$txhash
done

echo ""
echo "-- check the results --"

for function_name in "${names[@]}"; do
  echo $function_name
  txhash=${txhashes[$function_name]}
  expected_gas=${gas[$function_name]}

  get_receipt $txhash

  status=$(cat receipt.json | jq .status | sed 's/"//g')
  used_gas=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

  assert_equals "$status"   "SUCCESS"
  assert_equals "$used_gas" "$expected_gas"
done


# it can have 2 variations of this file
# 1. one test per txn
# 2. multiple tests per txn - to use the pre-loader
