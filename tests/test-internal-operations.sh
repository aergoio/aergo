set -e
source common.sh

fork_version=$1


cat > test-args.lua << EOF
function do_call(...)
  local args = {...}
  return contract.call(unpack(args))
end

function do_delegate_call(...)
  local args = {...}
  return contract.delegatecall(unpack(args))
end

function do_multicall(script)
  return contract.delegatecall("multicall", script)
end

function do_event(...)
  local args = {...}
  return contract.event(unpack(args))
end

function do_deploy(...)
  local args = {...}
  return contract.deploy(unpack(args))
end

function do_send(...)
  local args = {...}
  return contract.send(unpack(args))
end

function constructor(...)
  local args = {...}
  contract.event("constructor", unpack(args))
end

function default()
  contract.event("aergo received", system.getAmount())
  return system.getAmount()
end

abi.register(do_call, do_delegate_call, do_multicall, do_event, do_deploy, do_send, constructor)
abi.payable(default)
EOF


echo "-- deploy test contract --"

deploy test-args.lua
rm test-args.lua

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
test_args_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"

get_internal_operations $txhash
internal_operations=$(cat internal_operations.json)

assert_equals "$internal_operations" '{
  "call": {
    "contract": "'$test_args_address'",
    "function": "constructor",
    "operations": [
      {
        "args": [
          "constructor",
          "[]"
        ],
        "op": "event"
      }
    ]
  }
}'


echo "-- call test contract --"

#txhash=$(../bin/aergocli --keystore . --password bmttest \
#  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
#  $test_args_address do_event '["test",123,4.56,"test",true,{"_bignum":"1234567890"}]' | jq .hash | sed 's/"//g')

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  $test_args_address do_call '["'$test_args_address'","do_event","test",123,4.56,"test",true,{"_bignum":"1234567890"}]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')
gasUsed=$(cat receipt.json | jq .gasUsed | sed 's/"//g')

assert_equals "$status" "SUCCESS"

get_internal_operations $txhash
internal_operations=$(cat internal_operations.json)

assert_equals "$internal_operations" '{
  "call": {
    "args": [
      "'$test_args_address'",
      "do_event",
      "test",
      123,
      4.56,
      "test",
      true,
      {
        "_bignum": "1234567890"
      }
    ],
    "contract": "'$test_args_address'",
    "function": "do_call",
    "operations": [
      {
        "args": [
          "'$test_args_address'",
          "do_event",
          "[\"test\",123,4.56,\"test\",true,{\"_bignum\":\"1234567890\"}]"
        ],
        "call": {
          "args": [
            "test",
            123,
            4.56,
            "test",
            true,
            {
              "_bignum": "1234567890"
            }
          ],
          "contract": "'$test_args_address'",
          "function": "do_event",
          "operations": [
            {
              "args": [
                "test",
                "[123,4.56,\"test\",true,{\"_bignum\":\"1234567890\"}]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "call"
      }
    ]
  }
}'



echo "-- deploy ARC1 factory --"

deploy ../contract/vm_dummy/test_files/arc1-factory.lua

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
arc1_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- deploy ARC2 factory --"

deploy ../contract/vm_dummy/test_files/arc2-factory.lua

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
arc2_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"


echo "-- deploy caller contract --"

get_deploy_args ../contract/vm_dummy/test_files/token-deployer.lua

txhash=$(../bin/aergocli --keystore . --password bmttest \
    contract deploy AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
    $deploy_args "[\"$arc1_address\",\"$arc2_address\"]" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
deployer_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"

get_internal_operations $txhash

internal_operations=$(cat internal_operations.json)

assert_equals "$internal_operations" '{
  "call": {
    "args": [
      "'$arc1_address'",
      "'$arc2_address'"
    ],
    "contract": "'$deployer_address'",
    "function": "constructor"
  }
}'



echo "-- transfer 1 aergo to the contract --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  sendtx --from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R --to $test_args_address --amount 1aergo \
  | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status"   "SUCCESS"
assert_equals "$ret"      "1000000000000000000"


get_internal_operations $txhash

internal_operations=$(cat internal_operations.json)

assert_equals "$internal_operations" '{
  "call": {
    "contract": "'$test_args_address'",
    "function": "default",
    "operations": [
      {
        "args": [
          "aergo received",
          "[\"1000000000000000000\"]"
        ],
        "op": "event"
      }
    ]
  }
}'


if [ "$fork_version" -lt "4" ]; then
  exit 0  # composable transactions are only available from hard fork 4
fi

echo "-- multicall --"

script='[
 ["send","'$test_args_address'","0.125 aergo"],
 ["get aergo balance","'$test_args_address'"],
 ["to string"],
 ["store result as","amount"],
 ["call","'$test_args_address'","do_send","%my account address%","0.125 aergo"],
 ["return","%amount%"]
]'

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract multicall AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R "$script" | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status"   "SUCCESS"

if [ "$consensus" == "sbp" ]; then
  assert_equals "$ret"      "20001125000000000000000"
else
  assert_equals "$ret"      "1125000000000000000"
fi

get_internal_operations $txhash

internal_operations=$(cat internal_operations.json)

assert_equals "$internal_operations" '{
  "call": {
    "args": [
      [
        [
          "send",
          "'$test_args_address'",
          "0.125 aergo"
        ],
        [
          "get aergo balance",
          "'$test_args_address'"
        ],
        [
          "to string"
        ],
        [
          "store result as",
          "amount"
        ],
        [
          "call",
          "'$test_args_address'",
          "do_send",
          "%my account address%",
          "0.125 aergo"
        ],
        [
          "return",
          "%amount%"
        ]
      ]
    ],
    "contract": "AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R",
    "function": "execute",
    "operations": [
      {
        "amount": "0.125 aergo",
        "args": [
          "'$test_args_address'"
        ],
        "call": {
          "contract": "'$test_args_address'",
          "function": "default",
          "operations": [
            {
              "args": [
                "aergo received",
                "[\"125000000000000000\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "send"
      },
      {
        "args": [
          "'$test_args_address'",
          "do_send",
          "[\"AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R\",\"0.125 aergo\"]"
        ],
        "call": {
          "args": [
            "AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R",
            "0.125 aergo"
          ],
          "contract": "'$test_args_address'",
          "function": "do_send",
          "operations": [
            {
              "amount": "0.125 aergo",
              "args": [
                "AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R"
              ],
              "op": "send"
            }
          ]
        },
        "op": "call"
      }
    ]
  }
}'
