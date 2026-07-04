# Hashtabledb mmap benchmark

Compare **badgerdb**, **hashtabledb (no mmap)**, and **hashtabledb (mmap)** on a single-node SBP chain using a contract write flood and read queries.

## Prerequisites

- Linux build host with **Go 1.23+**, **cmake**, **make**, **jq**, **bc**
- Repository branch with:
  - mmap-enabled `aergo-lib` in `go.mod` (current HEAD on `hashtabledb-with-mmap`)
  - trie `parseBatch` mmap fix (`pkg/trie/trie.go`)
  - SBP backpressure / timeout fixes (`consensus/impl/sbp/sbp.go`)

## Build binaries

From `tests/`:

```bash
chmod +x build-bench-bins.sh benchmark-hashtabledb-mmap.sh
./build-bench-bins.sh
```

This writes:

| Path | Contents |
|------|----------|
| `../bench-bins/mmap/` | `aergosvr`, `aergocli`, `aergoluac` built with current `go.mod` (`UseMmap: true` in aergo-lib) |
| `../bench-bins/nommap/` | Same Aergo sources, `aergo-lib` pinned to pre-mmap revision (no `UseMmap`) |
| `./benchflood-bin` | Go flood tool (sign, commit, confirm, query) |

The mmap vs no-mmap difference is **only** the `aergo-lib` hashtabledb backend version. Badger and no-mmap hashtabledb legs both use the nommap `aergosvr` binary with different `dbtype` in config.

Rebuild after changing `benchflood/main.go`, trie/SBP code, or `go.mod`.

## Run benchmark

From `tests/`:

```bash
./benchmark-hashtabledb-mmap.sh
```

Optional custom binary directories:

```bash
./benchmark-hashtabledb-mmap.sh /path/to/mmap-bins /path/to/nommap-bins
```

### Default workload

| Variable | Default | Meaning |
|----------|---------|---------|
| `NUM_ACCOUNTS` | 100 | Flood accounts (testmode virtual balance) |
| `TXS_PER_ACCOUNT` | 1000 | Nonces 1..N per account → **100k writes** |
| `QUERY_COUNT` | 10000 | Contract `get_value` reads after writes |
| `BATCH_SIZE` | 500 | `CommitTX` batch size |
| `COMMIT_WORKERS` | 40 | Parallel commit batches |
| `SIGN_WORKERS` | `nproc` | Parallel signers (not timed) |
| `QUERY_WORKERS` | 80 | Parallel query RPCs |
| `CONFIRM_WORKERS` | 8 | Parallel block scanners |
| `CONFIRM_TIMEOUT` | 120 | Max seconds to confirm all txs |
| `STUCK_TIMEOUT` | 30 | Fail if confirm progress stalls |

Example larger run:

```bash
NUM_ACCOUNTS=200 TXS_PER_ACCOUNT=1000 CONFIRM_TIMEOUT=300 \
  ./benchmark-hashtabledb-mmap.sh
```

### Background run (nohup)

```bash
cd tests
nohup ./benchmark-hashtabledb-mmap.sh > benchmark-run.log 2>&1 &
tail -f benchmark-run.log
```

Per-leg logs: `benchmark-badger/logs`, `benchmark-nommap/logs`, `benchmark-mmap/logs`.  
Numeric results: `benchmark-results/{badgerdb,nommap,mmap}.txt`.  
The script prints a comparison table at the end (badgerdb | no-mmap | mmap).

### Stop a run

```bash
pkill -f benchmark-hashtabledb-mmap.sh
pkill -f benchflood-bin
pkill -f 'aergosvr --testmode --home ./benchmark-'
```

## What gets measured

Each leg deploys `test-performance.lua`, creates accounts, then `benchflood flood`:

1. Pre-sign all txs (not included in send TPS)
2. Time parallel `CommitTX` batches → **send TPS**
3. Scan blocks from height before commit until all tx hashes confirmed → **confirm TPS**, **peak tx/block**, **blocks used**
4. Run contract queries → **query QPS**

The orchestrator runs three legs in order: **badgerdb** → **hashtabledb no-mmap** → **hashtabledb mmap**. Any failed leg is reported; the script exits non-zero if any leg failed.

## Transfer to another machine

**Commit / copy (source):**

- `tests/benchmark-hashtabledb-mmap.sh`
- `tests/build-bench-bins.sh`
- `tests/benchflood/main.go`
- `tests/test-performance.lua`
- `tests/BENCHMARK.md`
- Branch commits: trie mmap fix, SBP fixes, mmap `go.mod` bump

**Do not commit:**

- `bench-bins/` (rebuild with `build-bench-bins.sh`)
- `tests/benchflood-bin`
- `tests/benchmark-*` workdirs, `benchmark-results/`, `benchmark-run.log`

On the target machine: clone branch → `cd tests && ./build-bench-bins.sh` → `./benchmark-hashtabledb-mmap.sh`.

## Files

| File | Role |
|------|------|
| `benchmark-hashtabledb-mmap.sh` | Orchestrator: 3 legs + comparison table |
| `build-bench-bins.sh` | Build mmap/nommap server CLIs + benchflood |
| `benchflood/main.go` | Flood, confirm, query tool |
| `test-performance.lua` | Contract: `set_value` / `get_value` |
