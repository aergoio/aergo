set -e
source common.sh

fork_version=$1


echo "-- deploy --"

deploy ../contract/vm_dummy/test_files/gas_per_function.lua $fork_version

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

add_test "comp_ops" 134635
add_test "unarytest_n_copy_ops" 134548
add_test "unary_ops" 134947
add_test "binary_ops" 136470
add_test "constant_ops" 134463
add_test "upvalue_n_func_ops" 135742
add_test "table_ops" 135733
add_test "call_n_vararg_ops" 136396
add_test "return_ops" 134468
add_test "loop_n_branche_ops" 137803
add_test "function_header_ops" 134447

add_test "assert" 134577
add_test "getfenv" 134472
add_test "metatable" 135383
add_test "ipairs" 134470
add_test "pairs" 134470
add_test "next" 134518
add_test "rawequal" 134647
add_test "rawget" 134518
add_test "rawset" 135336
add_test "select" 134597
add_test "setfenv" 134507
add_test "tonumber" 134581
add_test "tostring" 134852
add_test "type" 134680
add_test "unpack" 142140
add_test "pcall" 138169
add_test "xpcall" 138441

add_test "string.byte" 148435
add_test "string.char" 151792
add_test "string.dump" 138300
add_test "string.find" 139239
add_test "string.format" 135159
add_test "string.gmatch" 135194
add_test "string.gsub" 136338
add_test "string.len" 134528
add_test "string.lower" 139746
add_test "string.match" 134708
add_test "string.rep" 213323
add_test "string.reverse" 139746
add_test "string.sub" 136600
add_test "string.upper" 139746

add_test "table.concat" 155263
add_test "table.insert" 288649
add_test "table.remove" 148059
add_test "table.maxn" 139357
add_test "table.sort" 151261

add_test "math.abs" 134615
add_test "math.ceil" 134615
add_test "math.floor" 134615
add_test "math.max" 134987
add_test "math.min" 134987
add_test "math.pow" 134975

add_test "bit.tobit" 134510
add_test "bit.tohex" 134985
add_test "bit.bnot" 134487
add_test "bit.bor" 134561
add_test "bit.band" 134537
add_test "bit.xor" 134537
add_test "bit.lshift" 134510
add_test "bit.rshift" 134510
add_test "bit.ashift" 134510
add_test "bit.rol" 134510
add_test "bit.ror" 134510
add_test "bit.bswap" 134467

add_test "bignum.number" 136307
add_test "bignum.isneg" 136539
add_test "bignum.iszero" 136539
add_test "bignum.tonumber" 136859
add_test "bignum.tostring" 137150
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

add_test "state.set" 137310
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
