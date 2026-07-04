#!/bin/bash
# Benchmark db engines: hashtabledb (mmap / no-mmap) vs badgerdb.

set -euo pipefail
cd "$(dirname "$0")"

MMAP_BIN_DIR="${1:-../bench-bins/mmap}"
NOMMAP_BIN_DIR="${2:-../bench-bins/nommap}"

NUM_ACCOUNTS="${NUM_ACCOUNTS:-100}"
TXS_PER_ACCOUNT="${TXS_PER_ACCOUNT:-1000}"
QUERY_COUNT="${QUERY_COUNT:-10000}"
QUERY_WORKERS="${QUERY_WORKERS:-80}"
BATCH_SIZE="${BATCH_SIZE:-500}"
COMMIT_WORKERS="${COMMIT_WORKERS:-40}"
SIGN_WORKERS="${SIGN_WORKERS:-$(nproc)}"
CONFIRM_WORKERS="${CONFIRM_WORKERS:-8}"
CONFIRM_TIMEOUT="${CONFIRM_TIMEOUT:-120}"
STUCK_TIMEOUT="${STUCK_TIMEOUT:-30}"

WRITE_TX_COUNT=$((NUM_ACCOUNTS * TXS_PER_ACCOUNT))
RESULTS_DIR="./benchmark-results"
CREATOR="AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R"
BENCHFLOOD="./benchflood-bin"

cleanup() { pkill -f "aergosvr --testmode --home ./benchmark-" 2>/dev/null || true; sleep 1; }

reset_workdirs() {
  rm -rf ./benchmark-mmap ./benchmark-nommap ./benchmark-badger \
         ./benchmark-results
  mkdir -p "$RESULTS_DIR"
}

build_benchflood() {
  if [ ! -x "$BENCHFLOOD" ] || [ benchflood/main.go -nt "$BENCHFLOOD" ]; then
    echo "building benchflood..."
    (cd .. && go build -o tests/benchflood-bin ./tests/benchflood)
  fi
}

wait_chain() {
  "$BENCHFLOOD" wait-chain --rpc 127.0.0.1:7845
}

wait_receipt() {
  local b=$1 h=$2 max=${3:-600} n=0 last_h=0 stuck=0
  while [ $n -lt "$max" ]; do
    if "$b/aergocli" receipt get --port 7845 "$h" >receipt.json 2>err.txt && jq -e . receipt.json >/dev/null; then
      [ "$(jq -r .status receipt.json)" != "ERROR" ] && return 0
      echo "FAIL: tx error"; cat receipt.json; exit 1
    fi
    grep -q "tx not found" err.txt 2>/dev/null || { cat err.txt; exit 1; }
    cur=$("$b/aergocli" blockchain --port 7845 2>/dev/null | jq -r '.Height//0')
    [ "$cur" = "$last_h" ] && stuck=$((stuck+1)) || { stuck=0; last_h=$cur; }
    [ $stuck -ge 40 ] && { echo "FAIL: chain stuck at $cur"; exit 1; }
    sleep 0.5; n=$((n+1))
  done
  echo "FAIL: timeout $h"; exit 1
}

setup_node() {
  local bin=$1 wd=$2 dbtype=$3
  rm -rf "$wd"; mkdir -p "$wd"
  cat >"$wd/config.toml" <<EOF
datadir = "$wd/data"
dbtype = "$dbtype"
personal = true
authdir = "$wd/auth"
[rpc]
netserviceaddr = "127.0.0.1"
netserviceport = 7845
[p2p]
netprotocolport = 7846
npbindport = -1
npusepolaris = false
[blockchain]
maxblocksize = 1048576
numworkers = "16"
numclosers = "8"
[mempool]
verifiers = 16
[consensus]
enablebp = true
blockinterval = 1
[hardfork]
v2 = "0"
v3 = "10000"
v4 = "10000"
v5 = "10000"
EOF
  "$bin/aergosvr" --testmode --home "$wd" >>"$wd/logs" 2>&1 &
  echo $! >"$wd/pid"
  wait_chain
}

stop_node() {
  local wd=$1
  [ -f "$wd/pid" ] && kill "$(cat "$wd/pid")" 2>/dev/null && wait "$(cat "$wd/pid")" 2>/dev/null || true
  rm -f "$wd/pid"
}

FAILED_RUNS=()

run_benchmark_leg() {
  local label=$1 bin=$2 wd=$3 dbtype=$4 results=$5
  local rc=0
  echo ""
  echo "======== $label ($WRITE_TX_COUNT writes) ========"
  setup_node "$bin" "$wd" "$dbtype" || {
    echo "FAIL: could not start node for $label"
    FAILED_RUNS+=("$label")
    cleanup
    return 0
  }
  run_test "$label" "$bin" "$wd" "$results" || rc=$?
  stop_node "$wd"
  cleanup
  if [ "$rc" -ne 0 ]; then
    echo "FAIL: $label (exit $rc)"
    FAILED_RUNS+=("$label")
  else
    echo "OK: $label"
  fi
  return 0
}
run_test() {
  local label=$1 bin=$2 wd=$3 out=$4
  local CLI="$bin/aergocli" LUAC="$bin/aergoluac"

  "$CLI" account import --keystore "$wd" \
    --if 47zh1byk8MqWkQo5y8dvbrex99ZMdgZqfydar7w2QQgQqc7YrmFsBuMeF1uHWa5TwA1ZwQ7V6 \
    --password bmttest >/dev/null

  local nonce payload txhash contract
  nonce=$("$CLI" getstate --address "$CREATOR" | jq -r .nonce)
  "$LUAC" --payload test-performance.lua >"$wd/payload.out"
  payload=$(cat "$wd/payload.out")
  txhash=$("$CLI" --keystore "$wd" --password bmttest \
    contract deploy --payload "$payload" "$CREATOR" | jq -r .hash)
  wait_receipt "$bin" "$txhash" 120
  contract=$(jq -r .contractAddress receipt.json)

  echo "-- create $NUM_ACCOUNTS accounts (testmode: virtual balance) --"
  "$BENCHFLOOD" accounts \
    --keystore "$wd" \
    --password bmttest \
    --count "$NUM_ACCOUNTS" \
    --accounts-out "$wd/accounts.txt"

  echo "-- flood via benchflood --"
  "$BENCHFLOOD" flood \
    --rpc 127.0.0.1:7845 \
    --keystore "$wd" \
    --password bmttest \
    --contract "$contract" \
    --accounts-file "$wd/accounts.txt" \
    --txs-per-account "$TXS_PER_ACCOUNT" \
    --batch "$BATCH_SIZE" \
    --commit-workers "$COMMIT_WORKERS" \
    --sign-workers "$SIGN_WORKERS" \
    --query-count "$QUERY_COUNT" \
    --query-workers "$QUERY_WORKERS" \
    --confirm-timeout "$CONFIRM_TIMEOUT" \
    --confirm-workers "$CONFIRM_WORKERS" \
    --stuck-timeout "$STUCK_TIMEOUT" \
    --results "$out"
}

for d in "$MMAP_BIN_DIR" "$NOMMAP_BIN_DIR"; do
  for b in aergosvr aergocli aergoluac; do
    [ -x "$d/$b" ] || { echo "missing $d/$b"; exit 1; }
  done
done

build_benchflood
cleanup
reset_workdirs
echo "DB benchmark: $WRITE_TX_COUNT writes, $QUERY_COUNT reads, batch=$BATCH_SIZE commit_workers=$COMMIT_WORKERS"

set +e
run_benchmark_leg badgerdb "$NOMMAP_BIN_DIR" ./benchmark-badger badgerdb "$RESULTS_DIR/badgerdb.txt"
run_benchmark_leg "hashtabledb no-mmap" "$NOMMAP_BIN_DIR" ./benchmark-nommap hashtabledb "$RESULTS_DIR/nommap.txt"
run_benchmark_leg "hashtabledb mmap" "$MMAP_BIN_DIR" ./benchmark-mmap hashtabledb "$RESULTS_DIR/mmap.txt"
set -e

if [ ${#FAILED_RUNS[@]} -gt 0 ]; then
  echo ""
  echo "Failed runs: ${FAILED_RUNS[*]}"
fi

echo ""
echo "======== COMPARISON (badgerdb | no-mmap | mmap) ========"
printf "%-26s %12s %12s %12s %8s %8s\n" "Metric" "badgerdb" "no-mmap" "mmap" "mmap/nm" "mmap/bg"
load() {
  local f=$1 k=$2
  [ -f "$f" ] || { echo "N/A"; return; }
  grep "^${k}=" "$f" 2>/dev/null | cut -d= -f2- || echo "N/A"
}
ratio() {
  local a=$1 b=$2 hi=$3
  if [ "$a" = "N/A" ] || [ "$b" = "N/A" ]; then echo "N/A"; return; fi
  if [ "$a" = "0" ] && [ "$b" = "0" ]; then echo "1.00"; return; fi
  if [ "$hi" = 1 ]; then echo "scale=2; $b / $a" | bc -l
  else echo "scale=2; $a / $b" | bc -l; fi
}
cmp3() {
  local n=$1 k=$2 hi=$3
  local bg nm mm s1 s2
  bg=$(load "$RESULTS_DIR/badgerdb.txt" "$k")
  nm=$(load "$RESULTS_DIR/nommap.txt" "$k")
  mm=$(load "$RESULTS_DIR/mmap.txt" "$k")
  s1=$(ratio "$nm" "$mm" "$hi")
  s2=$(ratio "$bg" "$mm" "$hi")
  printf "%-26s %12s %12s %12s %7sx %7sx\n" "$n" "$bg" "$nm" "$mm" "$s1" "$s2"
}
cmp3 "Confirmed txs" confirmed_count 1
cmp3 "Mined txs" mined_tx_count 1
cmp3 "Send TPS" send_tps 1
cmp3 "Confirm TPS" confirm_tps 1
cmp3 "Total TPS" total_tps 1
cmp3 "Peak tx/block" peak_tx_per_block 1
cmp3 "Blocks used" blocks_used 0
cmp3 "Send time (s)" send_seconds 0
cmp3 "Confirm time (s)" confirm_seconds 0
cmp3 "Total time (s)" total_seconds 0
cmp3 "Query QPS" query_qps 1
cmp3 "Query errors" query_errors 0

[ ${#FAILED_RUNS[@]} -eq 0 ] || exit 1
