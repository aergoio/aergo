
get_receipt() {
  txhash=$1
  counter=0
  set +e

  while true; do
    output=$(../bin/aergocli receipt get $txhash 2>&1 > receipt.json)

    #echo "output: $output"

    if [[ $output == *"tx not found"* ]]; then
      sleep 0.5
      counter=$((counter+1))
      if [ $counter -gt 10 ]; then
        echo "Error: tx not found: $txhash"
        exit 1
      fi
      continue
    elif [[ -n $output ]]; then
      echo "Error: $output"
      exit 1
    else
      break
    fi
  done

  set -e
}

assert_equals() {
  local var="$1"
  local expected="$2"

  if [[ ! "$var" == "$expected" ]]; then
    echo "Assertion failed: $var != $expected"
    exit 1
  fi
}

assert_contains() {
  local var="$1"
  local substring="$2"

  if [[ ! "$var" == *"$substring"* ]]; then
    echo "Assertion failed: $var does not contain $substring"
    exit 1
  fi
}