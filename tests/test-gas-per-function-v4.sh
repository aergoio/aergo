set -e
source common.sh

fork_version=$1

# skip the test if the consensus type is raft
if [ "$consensus" = "raft" ]; then
  echo "  skipping on private chain..."
  exit 0
fi


echo "-- deploy --"

deploy ../contract/vm_dummy/test_files/gas_per_function.lua

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- transfer funds to the contract --"

from=AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R

txhash=$(../bin/aergocli --keystore . --password bmttest \
  sendtx --from $from --to $address --amount 5aergo \
  | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"
assert_equals "$ret"    ""


echo "-- get account's nonce --"

account_state=$(../bin/aergocli getstate --address $from)
nonce=$(echo $account_state | jq .nonce | sed 's/"//g')


# create an iterable map of function name -> expected gas
declare -a names
declare -A gas

add_test() {
  names+=("$1")
  gas["$1"]="$2"
}

add_test "comp_ops" 143204
add_test "unarytest_n_copy_ops" 143117
add_test "unary_ops" 143552
add_test "binary_ops" 145075
add_test "constant_ops" 143032
add_test "upvalue_n_func_ops" 144347
add_test "table_ops" 144482
add_test "call_n_vararg_ops" 145001
add_test "return_ops" 143037
add_test "loop_n_branche_ops" 146372
add_test "function_header_ops" 143016

add_test "assert" 143146
add_test "ipairs" 143039
add_test "pairs" 143039
add_test "next" 143087
add_test "select" 143166
add_test "tonumber" 143186
add_test "tostring" 143457
add_test "type" 143285
add_test "unpack" 150745
add_test "pcall" 147905
add_test "xpcall" 148177

add_test "string.byte" 157040
add_test "string.char" 160397
add_test "string.find" 147808
add_test "string.format" 143764
add_test "string.gmatch" 143799
add_test "string.gsub" 144943
add_test "string.len" 143097
add_test "string.lower" 148351
add_test "string.match" 143313
add_test "string.rep" 221928
add_test "string.reverse" 148351
add_test "string.sub" 145205
add_test "string.upper" 148351

add_test "table.concat" 163868
add_test "table.insert" 297254
add_test "table.remove" 156664
add_test "table.maxn" 147962
add_test "table.sort" 159866

add_test "math.abs" 143184
add_test "math.ceil" 143184
add_test "math.floor" 143184
add_test "math.max" 143556
add_test "math.min" 143556
add_test "math.pow" 143544

add_test "bit.tobit" 143079
add_test "bit.tohex" 143590
add_test "bit.bnot" 143056
add_test "bit.bor" 143130
add_test "bit.band" 143106
add_test "bit.xor" 143106
add_test "bit.lshift" 143079
add_test "bit.rshift" 143079
add_test "bit.ashift" 143079
add_test "bit.rol" 143079
add_test "bit.ror" 143079
add_test "bit.bswap" 143036

add_test "bignum.number" 144912
add_test "bignum.isneg" 145144
add_test "bignum.iszero" 145144
add_test "bignum.tonumber" 145464
add_test "bignum.tostring" 145755
add_test "bignum.neg" 147208
add_test "bignum.sqrt" 148084
add_test "bignum.compare" 145409
add_test "bignum.add" 146750
add_test "bignum.sub" 146695
add_test "bignum.mul" 149073
add_test "bignum.div" 148563
add_test "bignum.mod" 150498
add_test "bignum.pow" 149492
add_test "bignum.divmod" 154798
add_test "bignum.powmod" 154164
add_test "bignum.operators" 147416

add_test "json" 151357

add_test "crypto.sha256" 146183
add_test "crypto.ecverify" 148036

add_test "state.set" 145915
add_test "state.get" 145720
add_test "state.delete" 145727

add_test "system.getSender" 144261
add_test "system.getBlockheight" 143330
add_test "system.getTxhash" 143737
add_test "system.getTimestamp" 143330
add_test "system.getContractID" 144261
add_test "system.setItem" 144194
add_test "system.getItem" 144503
add_test "system.getAmount" 143408
add_test "system.getCreator" 143761
add_test "system.getOrigin" 144261

add_test "contract.send" 144321
#add_test "contract.balance" 144402
add_test "contract.deploy" 168092
add_test "contract.call" 159738
add_test "contract.pcall" 160659
add_test "contract.delegatecall" 153795
add_test "contract.event" 163452

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
