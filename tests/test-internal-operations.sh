set -e
source common.sh

fork_version=$1


cat > test-args.lua << EOF
function do_call(...)
  return contract.call(...)
end

function do_delegate_call(...)
  return contract.delegatecall(...)
end

function do_multicall(script)
  return contract.delegatecall("multicall", script)
end

function do_event(...)
  return contract.event(...)
end

function do_deploy(...)
  return contract.deploy(...)
end

function do_send(...)
  return contract.send(...)
end

function do_gov(commands)
  for i, command in ipairs(commands) do
    local cmd = command[1]
    local arg = command[2]
    if cmd == "stake" then
      contract.stake(arg)
    elseif cmd == "vote" then
      contract.vote(arg)
    elseif cmd == "unstake" then
      contract.unstake(arg)
    end
  end
end

function constructor(...)
  contract.event("constructor", ...)
end

function default()
  contract.event("aergo received", system.getAmount())
  return system.getAmount()
end

abi.register(do_call, do_delegate_call, do_multicall, do_event, do_deploy, do_send, do_gov, constructor)
abi.payable(default)
EOF


cat > test-reverted-operations.lua << EOF
function error_case(to_call, to_send)
  --contract.stake("10000 aergo")
  --contract.vote("16Uiu2HAm2gtByd6DQu95jXURJXnS59Dyb9zTe16rDrcwKQaxma4p")
  contract.call(to_call, "do_event", "ping", "called - within")
  contract.send(to_send, "15 aergo")
  contract.event("ping", "within")
  assert(false) -- revert all operations above
end

function test_pcall(to_call, to_send)
  contract.call(to_call, "do_event", "ping", "called - before")
  contract.send(to_send, "15 aergo")
  contract.event("ping", "before")

  pcall(error_case, to_call, to_send)

  contract.send(to_send, "15 aergo")
  contract.event("ping", "after")
  contract.call(to_call, "do_event", "ping", "called - after")
end

abi.payable(test_pcall, error_case)
EOF


echo "-- deploy test contract 1 --"

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


echo "-- call test contract 1 --"

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
    "amount": "1000000000000000000",
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
  # composable transactions are only available from hard fork 4
  # the tracking of reverted operations is also only available from hard fork 4
  rm test-reverted-operations.lua
  exit 0
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
          "amount": "125000000000000000",
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


echo "-- deploy test contract 2 --"

deploy test-reverted-operations.lua
rm test-reverted-operations.lua

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
test_reverted_address=$(cat receipt.json | jq .contractAddress | sed 's/"//g')

assert_equals "$status" "CREATED"

#get_internal_operations $txhash
#internal_operations=$(cat internal_operations.json)

#assert_equals "$internal_operations" ''


echo "-- call test contract 2 - reverted operations --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  --amount 50000000000000000000 \
  $test_reverted_address test_pcall '["'$test_args_address'","'$test_args_address'"]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "SUCCESS"

get_internal_operations $txhash
internal_operations=$(cat internal_operations.json)

assert_equals "$internal_operations" '{
  "call": {
    "amount": "50000000000000000000",
    "args": [
      "'$test_args_address'",
      "'$test_args_address'"
    ],
    "contract": "'$test_reverted_address'",
    "function": "test_pcall",
    "operations": [
      {
        "args": [
          "'$test_args_address'",
          "do_event",
          "[\"ping\",\"called - before\"]"
        ],
        "call": {
          "args": [
            "ping",
            "called - before"
          ],
          "contract": "'$test_args_address'",
          "function": "do_event",
          "operations": [
            {
              "args": [
                "ping",
                "[\"called - before\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "call"
      },
      {
        "amount": "15 aergo",
        "args": [
          "'$test_args_address'"
        ],
        "call": {
          "amount": "15000000000000000000",
          "contract": "'$test_args_address'",
          "function": "default",
          "operations": [
            {
              "args": [
                "aergo received",
                "[\"15000000000000000000\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "send"
      },
      {
        "args": [
          "ping",
          "[\"before\"]"
        ],
        "op": "event"
      },
      {
        "args": [
          "'$test_args_address'",
          "do_event",
          "[\"ping\",\"called - within\"]"
        ],
        "call": {
          "args": [
            "ping",
            "called - within"
          ],
          "contract": "'$test_args_address'",
          "function": "do_event",
          "operations": [
            {
              "args": [
                "ping",
                "[\"called - within\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "call",
        "reverted": true
      },
      {
        "amount": "15 aergo",
        "args": [
          "'$test_args_address'"
        ],
        "call": {
          "amount": "15000000000000000000",
          "contract": "'$test_args_address'",
          "function": "default",
          "operations": [
            {
              "args": [
                "aergo received",
                "[\"15000000000000000000\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "send",
        "reverted": true
      },
      {
        "args": [
          "ping",
          "[\"within\"]"
        ],
        "op": "event",
        "reverted": true
      },
      {
        "amount": "15 aergo",
        "args": [
          "'$test_args_address'"
        ],
        "call": {
          "amount": "15000000000000000000",
          "contract": "'$test_args_address'",
          "function": "default",
          "operations": [
            {
              "args": [
                "aergo received",
                "[\"15000000000000000000\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "send"
      },
      {
        "args": [
          "ping",
          "[\"after\"]"
        ],
        "op": "event"
      },
      {
        "args": [
          "'$test_args_address'",
          "do_event",
          "[\"ping\",\"called - after\"]"
        ],
        "call": {
          "args": [
            "ping",
            "called - after"
          ],
          "contract": "'$test_args_address'",
          "function": "do_event",
          "operations": [
            {
              "args": [
                "ping",
                "[\"called - after\"]"
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


echo "-- call test contract 2 - error case --"

txhash=$(../bin/aergocli --keystore . --password bmttest \
  contract call AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R \
  --amount 50000000000000000000 \
  $test_reverted_address error_case '["'$test_args_address'","'$test_args_address'"]' | jq .hash | sed 's/"//g')

get_receipt $txhash

status=$(cat receipt.json | jq .status | sed 's/"//g')
ret=$(cat receipt.json | jq .ret | sed 's/"//g')

assert_equals "$status" "ERROR"
assert_contains "$ret" "assertion failed!"

get_internal_operations $txhash
internal_operations=$(cat internal_operations.json)

assert_equals "$internal_operations" '{
  "call": {
    "amount": "50000000000000000000",
    "args": [
      "'$test_args_address'",
      "'$test_args_address'"
    ],
    "contract": "'$test_reverted_address'",
    "function": "error_case",
    "operations": [
      {
        "args": [
          "'$test_args_address'",
          "do_event",
          "[\"ping\",\"called - within\"]"
        ],
        "call": {
          "args": [
            "ping",
            "called - within"
          ],
          "contract": "'$test_args_address'",
          "function": "do_event",
          "operations": [
            {
              "args": [
                "ping",
                "[\"called - within\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "call"
      },
      {
        "amount": "15 aergo",
        "args": [
          "'$test_args_address'"
        ],
        "call": {
          "amount": "15000000000000000000",
          "contract": "'$test_args_address'",
          "function": "default",
          "operations": [
            {
              "args": [
                "aergo received",
                "[\"15000000000000000000\"]"
              ],
              "op": "event"
            }
          ]
        },
        "op": "send"
      },
      {
        "args": [
          "ping",
          "[\"within\"]"
        ],
        "op": "event"
      }
    ]
  }
}'
