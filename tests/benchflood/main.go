package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/account/key"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"google.golang.org/grpc"
)

const maxRPCMsg = 64 * 1024 * 1024

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "wait-chain":
		runWaitChain(os.Args[2:])
	case "accounts":
		runAccounts(os.Args[2:])
	case "fund":
		runFund(os.Args[2:])
	case "flood":
		runFlood(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `benchflood - fast tx flood for hashtabledb benchmark

Usage:
  benchflood wait-chain [--rpc ADDR]
  benchflood accounts [flags]   create keystore accounts (no chain txs)
  benchflood fund [flags]       create accounts + transfer funds (non-testmode)
  benchflood flood [flags]

Fund flags:
  --creator ADDR          funder account (must be in keystore)
  --count N               create and fund N accounts
  --amount AMT            transfer amount per account (default 0.1aergo)
  --accounts-out FILE     write new account addresses (one per line)

Flood flags:
  1. Pre-sign all txs (not timed for throughput)
  2. Timer starts, all CommitTX batches fire in parallel
  3. Separate timer for confirmation (first..last receipt)
  4. Block span and per-block tx counts measured on-chain

Flood flags:
  --rpc ADDR              gRPC address (default 127.0.0.1:7845)
  --keystore DIR          keystore directory
  --password PASS         keystore password
  --contract ADDR         contract address
  --accounts-file FILE    one account address per line
  --txs-per-account N     nonces 1..N per account (default 40)
  --batch N               CommitTX batch size (default 500)
  --sign-workers N        parallel signers (default NumCPU)
  --commit-workers N      parallel commit batches (default 20)
  --query-count N         read queries after writes (default 1000)
  --confirm-timeout SEC   max wait for confirmation (default 120)
  --confirm-workers N     parallel block scanners (default 8)
  --stuck-timeout SEC     fail if chain height unchanged (default 30)
`)
}

type accountKey struct {
	addr []byte
	key  *btcec.PrivateKey
}

func dialRPC(addr string) (types.AergoRPCServiceClient, types.AdminRPCServiceClient, func()) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxRPCMsg),
			grpc.MaxCallSendMsgSize(maxRPCMsg),
		),
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		fatal("dial %s: %v", addr, err)
	}
	return types.NewAergoRPCServiceClient(conn), types.NewAdminRPCServiceClient(conn), func() { _ = conn.Close() }
}

type mempoolAccountStat struct {
	Pooled   struct{ Count int `json:"count"` } `json:"pooled"`
	Orphaned struct{ Count int `json:"count"` } `json:"orphaned"`
}

func mempoolPending(admin types.AdminRPCServiceClient) (int, bool) {
	if admin == nil {
		return 0, false
	}
	r, err := admin.MempoolTxStat(context.Background(), &types.Empty{})
	if err != nil || r == nil {
		return 0, false
	}
	var stats []mempoolAccountStat
	if err := json.Unmarshal(r.GetValue(), &stats); err != nil {
		return 0, false
	}
	total := 0
	for _, s := range stats {
		total += s.Pooled.Count + s.Orphaned.Count
	}
	return total, true
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FAIL: "+format+"\n", args...)
	os.Exit(1)
}

func runWaitChain(args []string) {
	fs := flag.NewFlagSet("wait-chain", flag.ExitOnError)
	rpc := fs.String("rpc", "127.0.0.1:7845", "gRPC address")
	_ = fs.Parse(args)

	client, _, closeConn := dialRPC(*rpc)
	defer closeConn()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		st, err := client.Blockchain(context.Background(), &types.Empty{})
		if err == nil && st.GetBestHeight() >= 1 {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	fatal("node not ready within 60s")
}

func readAccounts(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		fatal("open accounts: %v", err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		fatal("no accounts in %s", path)
	}
	return out
}

func loadKeys(keystore, pass string, addrs []string) []accountKey {
	ks := key.NewStore(keystore, 0)
	defer ks.CloseStore()
	out := make([]accountKey, len(addrs))
	for i, a := range addrs {
		addr, err := types.DecodeAddress(a)
		if err != nil {
			fatal("decode %s: %v", a, err)
		}
		pk, err := ks.GetKey(addr, pass)
		if err != nil {
			fatal("load key %s: %v", a, err)
		}
		out[i] = accountKey{addr: addr, key: pk}
	}
	return out
}

func callPayload(idx int) []byte {
	ci := types.CallInfo{
		Name: "set_value",
		Args: []interface{}{
			fmt.Sprintf("key_%d", idx),
			fmt.Sprintf("value_%d", idx),
		},
	}
	b, err := json.Marshal(ci)
	if err != nil {
		fatal("marshal callinfo: %v", err)
	}
	return b
}

func signCall(chainID, contract []byte, ak accountKey, nonce uint64, payload []byte) *types.Tx {
	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:       nonce,
			Account:     ak.addr,
			Recipient:   contract,
			Payload:     payload,
			Amount:      big.NewInt(0).Bytes(),
			Type:        types.TxType_CALL,
			ChainIdHash: chainID,
		},
	}
	if err := key.SignTx(tx, ak.key); err != nil {
		fatal("sign: %v", err)
	}
	return tx
}

func heightBytes(h uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, h)
	return b
}

func waitReceipt(client types.AergoRPCServiceClient, hash []byte, timeout, stuckTimeout time.Duration) *types.Receipt {
	deadline := time.Now().Add(timeout)
	var lastHeight uint64
	stuckSince := time.Now()
	for time.Now().Before(deadline) {
		rcpt, err := client.GetReceipt(context.Background(), &types.SingleBytes{Value: hash})
		if err == nil && rcpt != nil {
			if rcpt.GetStatus() == "ERROR" {
				fatal("tx failed: %s", rcpt.GetRet())
			}
			return rcpt
		}
		st, err := client.Blockchain(context.Background(), &types.Empty{})
		if err == nil {
			h := st.GetBestHeight()
			if h != lastHeight {
				lastHeight = h
				stuckSince = time.Now()
			} else if time.Since(stuckSince) >= stuckTimeout {
				fatal("chain stuck at height %d for %s", h, stuckTimeout)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	fatal("receipt timeout after %s", timeout)
	return nil
}

type floodConfirm struct {
	confirmedCount int
	minedTxCount     int
	firstBlock       uint64
	lastBlock        uint64
	scanFrom         uint64
	scanThrough      uint64
	peak             int
	blockTxCounts    map[uint64]int
}

func hashKey(h []byte) string {
	return string(h)
}

// verifyFloodConfirm checks hash matches and that on-chain tx counts match what we sent.
func verifyFloodConfirm(fc floodConfirm, sent int) {
	if sent == 0 {
		return
	}
	if fc.confirmedCount != sent {
		fatal("confirmed %d/%d flood txs by hash", fc.confirmedCount, sent)
	}
	if fc.minedTxCount < sent {
		fatal("mined %d txs on chain since height %d but sent %d",
			fc.minedTxCount, fc.scanFrom, sent)
	}
	if fc.minedTxCount != fc.confirmedCount {
		fatal("mined %d txs in scanned blocks but only %d matched flood hashes",
			fc.minedTxCount, fc.confirmedCount)
	}
}

func isBlockNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") || strings.Contains(msg, "NotFound")
}

// scanBlockMetadata reads metadata for [from..to] one block at a time.
// Stops at the first missing block — BestHeight can run ahead of connected blocks on slow DBs.
func scanBlockMetadata(client types.AergoRPCServiceClient, from, to uint64) ([]*types.BlockMetadata, uint64) {
	if to < from {
		return nil, from - 1
	}
	out := make([]*types.BlockMetadata, 0, to-from+1)
	lastOK := from - 1
	for h := from; h <= to; h++ {
		meta, err := client.GetBlockMetadata(context.Background(), &types.SingleBytes{Value: heightBytes(h)})
		if err != nil {
			if isBlockNotFound(err) {
				break
			}
			fatal("get block metadata %d: %v", h, err)
		}
		out = append(out, meta)
		lastOK = h
	}
	return out, lastOK
}

func matchBlockTxs(client types.AergoRPCServiceClient, blockNo uint64, pending map[string]struct{}, fc *floodConfirm, mu *sync.Mutex) bool {
	blk, err := client.GetBlock(context.Background(), &types.SingleBytes{Value: heightBytes(blockNo)})
	if err != nil || blk == nil || blk.Body == nil {
		return false
	}
	var matched []string
	for _, tx := range blk.Body.Txs {
		h := tx.GetHash()
		if len(h) == 0 {
			h = tx.CalculateTxHash()
		}
		if len(h) == 0 {
			continue
		}
		key := hashKey(h)
		mu.Lock()
		_, ok := pending[key]
		mu.Unlock()
		if ok {
			matched = append(matched, key)
		}
	}
	if len(matched) == 0 {
		return true
	}
	mu.Lock()
	defer mu.Unlock()
	for _, key := range matched {
		if _, ok := pending[key]; !ok {
			continue
		}
		delete(pending, key)
		fc.confirmedCount++
		if fc.firstBlock == 0 || blockNo < fc.firstBlock {
			fc.firstBlock = blockNo
		}
		if blockNo > fc.lastBlock {
			fc.lastBlock = blockNo
		}
	}
	return true
}

// waitFloodConfirm scans blocks via GetBlockMetadata and matches flood tx hashes in block bodies.
// scanFrom is the chain height before commits started; scanning begins at scanFrom+1 so txs
// included during the commit window are not missed.
func waitFloodConfirm(client types.AergoRPCServiceClient, admin types.AdminRPCServiceClient, hashes [][]byte, scanFrom uint64, timeout, stuckTimeout time.Duration, workers int) floodConfirm {
	fc := floodConfirm{
		blockTxCounts: make(map[uint64]int),
		scanFrom:      scanFrom,
	}
	if len(hashes) == 0 {
		return fc
	}
	if workers < 1 {
		workers = 1
	}

	pending := make(map[string]struct{}, len(hashes))
	for _, h := range hashes {
		pending[hashKey(h)] = struct{}{}
	}

	deadline := time.Now().Add(timeout)
	lastScanned := scanFrom
	var lastHeight uint64
	heightStuckSince := time.Now()
	confirmStuckSince := time.Now()
	lastConfirmed := 0
	lastReport := time.Now()
	retryBodies := make(map[uint64]struct{})

	fmt.Fprintf(os.Stderr, "  scanning blocks from height %d (pre-commit height %d)\n", scanFrom+1, scanFrom)

	for len(pending) > 0 && time.Now().Before(deadline) {
		best := chainHeight(client)
		now := time.Now()
		if best != lastHeight {
			lastHeight = best
			heightStuckSince = now
		} else if time.Since(heightStuckSince) >= stuckTimeout {
			fatal("chain stuck at height %d for %s while confirming (%d/%d txs)",
				best, stuckTimeout, fc.confirmedCount, len(hashes))
		}
		if fc.confirmedCount > lastConfirmed {
			lastConfirmed = fc.confirmedCount
			confirmStuckSince = now
		} else if time.Since(confirmStuckSince) >= stuckTimeout {
			mpNote := ""
			if mp, ok := mempoolPending(admin); ok {
				mpNote = fmt.Sprintf(", mempool %d", mp)
			}
			fatal("no confirmation progress for %s: confirmed %d/%d txs at height %d, scanned through %d (%d pending%s)",
				stuckTimeout, fc.confirmedCount, len(hashes), best, lastScanned, len(pending), mpNote)
		}

		var bodyBlocks []uint64
		for b := range retryBodies {
			bodyBlocks = append(bodyBlocks, b)
		}

		if best > lastScanned {
			metas, through := scanBlockMetadata(client, lastScanned+1, best)
			for _, meta := range metas {
				if meta == nil || meta.Header == nil {
					continue
				}
				blockNo := meta.Header.BlockNo
				ntx := int(meta.Txcount)
				if _, seen := fc.blockTxCounts[blockNo]; !seen {
					fc.minedTxCount += ntx
				}
				fc.blockTxCounts[blockNo] = ntx
				if ntx > fc.peak {
					fc.peak = ntx
				}
				if ntx > 0 && len(pending) > 0 {
					bodyBlocks = append(bodyBlocks, blockNo)
				}
			}
			if through > lastScanned {
				lastScanned = through
				fc.scanThrough = through
			}
		}

		if len(bodyBlocks) > 0 {
			work := make(chan uint64, len(bodyBlocks))
			seen := make(map[uint64]struct{}, len(bodyBlocks))
			for _, b := range bodyBlocks {
				if _, dup := seen[b]; dup {
					continue
				}
				seen[b] = struct{}{}
				work <- b
			}
			close(work)
			var wg sync.WaitGroup
			var mu sync.Mutex
			nextRetry := make(map[uint64]struct{})
			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for blockNo := range work {
						mu.Lock()
						empty := len(pending) == 0
						mu.Unlock()
						if empty {
							return
						}
						ok := matchBlockTxs(client, blockNo, pending, &fc, &mu)
						if !ok {
							mu.Lock()
							nextRetry[blockNo] = struct{}{}
							mu.Unlock()
						}
					}
				}()
			}
			wg.Wait()
			retryBodies = nextRetry
		}

		if time.Since(lastReport) >= 2*time.Second {
			mpNote := ""
			if mp, ok := mempoolPending(admin); ok {
				mpNote = fmt.Sprintf(", mempool %d", mp)
			}
			fmt.Fprintf(os.Stderr, "  confirmed %d/%d txs, height %d (scanned through %d), pending %d%s\n",
				fc.confirmedCount, len(hashes), best, lastScanned, len(pending), mpNote)
			lastReport = time.Now()
		}

		if len(pending) > 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if len(pending) > 0 {
		mpNote := ""
		if mp, ok := mempoolPending(admin); ok {
			mpNote = fmt.Sprintf(", mempool %d", mp)
		}
		fatal("receipt timeout after %s: confirmed %d/%d txs, scanned through %d%s",
			timeout, fc.confirmedCount, len(hashes), lastScanned, mpNote)
	}
	verifyFloodConfirm(fc, len(hashes))
	return fc
}

func chainHeight(client types.AergoRPCServiceClient) uint64 {
	st, err := client.Blockchain(context.Background(), &types.Empty{})
	if err != nil {
		fatal("blockchain: %v", err)
	}
	return st.GetBestHeight()
}

func commitBatch(client types.AergoRPCServiceClient, batch []*types.Tx) [][]byte {
	res, err := client.CommitTX(context.Background(), &types.TxList{Txs: batch})
	if err != nil {
		fatal("committx: %v", err)
	}
	hashes := make([][]byte, 0, len(batch))
	for _, r := range res.Results {
		if r.GetError() != types.CommitStatus_TX_OK {
			fatal("commit error: %s", r.GetError())
		}
		hashes = append(hashes, r.Hash)
	}
	return hashes
}

func floodCommits(client types.AergoRPCServiceClient, txs []*types.Tx, batchSize, commitWorkers int) (hashes [][]byte, sendDur float64) {
	type txBatch struct{ txs []*types.Tx }
	var batches []txBatch
	for off := 0; off < len(txs); off += batchSize {
		end := off + batchSize
		if end > len(txs) {
			end = len(txs)
		}
		batches = append(batches, txBatch{txs: txs[off:end]})
	}
	fmt.Fprintf(os.Stderr, "-- commit %d txs in %d batches (%d workers) --\n",
		len(txs), len(batches), commitWorkers)
	sendStart := time.Now()
	batchCh := make(chan txBatch, len(batches))
	for _, b := range batches {
		batchCh <- b
	}
	close(batchCh)
	var mu sync.Mutex
	hashes = make([][]byte, 0, len(txs))
	var wg sync.WaitGroup
	for w := 0; w < commitWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for b := range batchCh {
				h := commitBatch(client, b.txs)
				mu.Lock()
				hashes = append(hashes, h...)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	sendDur = time.Since(sendStart).Seconds()
	fmt.Fprintf(os.Stderr, "committed %d txs in %.3fs (%.0f commit/s)\n",
		len(hashes), sendDur, float64(len(hashes))/sendDur)
	return hashes, sendDur
}

func signTransfer(chainID []byte, creator accountKey, recipient []byte, nonce uint64, amount *big.Int) *types.Tx {
	tx := &types.Tx{
		Body: &types.TxBody{
			Nonce:       nonce,
			Account:     creator.addr,
			Recipient:   recipient,
			Amount:      amount.Bytes(),
			Type:        types.TxType_TRANSFER,
			ChainIdHash: chainID,
		},
	}
	if err := key.SignTx(tx, creator.key); err != nil {
		fatal("sign transfer: %v", err)
	}
	return tx
}

func runAccounts(args []string) {
	fs := flag.NewFlagSet("accounts", flag.ExitOnError)
	keystore := fs.String("keystore", "", "keystore directory")
	password := fs.String("password", "", "keystore password")
	count := fs.Int("count", 0, "accounts to create")
	accountsOut := fs.String("accounts-out", "", "output accounts file")
	_ = fs.Parse(args)

	if *keystore == "" || *password == "" || *count <= 0 || *accountsOut == "" {
		fatal("need --keystore --password --count --accounts-out")
	}

	ks := key.NewStore(*keystore, 0)
	defer ks.CloseStore()

	fmt.Fprintf(os.Stderr, "-- create %d accounts (testmode: no funding needed) --\n", *count)
	start := time.Now()
	addresses := make([]string, *count)
	for i := 0; i < *count; i++ {
		addr, err := ks.CreateKey(*password)
		if err != nil {
			fatal("create key: %v", err)
		}
		addresses[i] = types.EncodeAddress(addr)
	}
	fmt.Fprintf(os.Stderr, "created in %.3fs\n", time.Since(start).Seconds())

	var sb strings.Builder
	for _, a := range addresses {
		sb.WriteString(a)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(*accountsOut, []byte(sb.String()), 0644); err != nil {
		fatal("write accounts: %v", err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", *accountsOut)
}

func runFund(args []string) {
	fs := flag.NewFlagSet("fund", flag.ExitOnError)
	rpc := fs.String("rpc", "127.0.0.1:7845", "gRPC address")
	keystore := fs.String("keystore", "", "keystore directory")
	password := fs.String("password", "", "keystore password")
	creatorStr := fs.String("creator", "", "creator address")
	count := fs.Int("count", 0, "accounts to create and fund")
	amountStr := fs.String("amount", "0.1aergo", "amount per account")
	accountsOut := fs.String("accounts-out", "", "output accounts file")
	batchSize := fs.Int("batch", 500, "commit batch size")
	commitWorkers := fs.Int("commit-workers", 20, "parallel commit batches")
	confirmTimeout := fs.Int("confirm-timeout", 120, "confirmation timeout in seconds")
	stuckTimeout := fs.Int("stuck-timeout", 30, "fail if chain height unchanged for this many seconds")
	_ = fs.Parse(args)

	if *keystore == "" || *password == "" || *creatorStr == "" || *count <= 0 || *accountsOut == "" {
		fatal("need --keystore --password --creator --count --accounts-out")
	}
	amount, err := util.ParseUnit(*amountStr)
	if err != nil {
		fatal("amount: %v", err)
	}

	client, _, closeConn := dialRPC(*rpc)
	defer closeConn()

	st, err := client.Blockchain(context.Background(), &types.Empty{})
	if err != nil {
		fatal("blockchain: %v", err)
	}
	chainID := st.GetBestChainIdHash()

	creatorAddr, err := types.DecodeAddress(*creatorStr)
	if err != nil {
		fatal("creator address: %v", err)
	}
	creatorState, err := client.GetState(context.Background(), &types.SingleBytes{Value: creatorAddr})
	if err != nil {
		fatal("creator state: %v", err)
	}
	baseNonce := creatorState.GetNonce()

	ks := key.NewStore(*keystore, 0)
	defer ks.CloseStore()
	creatorPK, err := ks.GetKey(creatorAddr, *password)
	if err != nil {
		fatal("creator key: %v", err)
	}
	creator := accountKey{addr: creatorAddr, key: creatorPK}

	fmt.Fprintf(os.Stderr, "-- create %d accounts --\n", *count)
	createStart := time.Now()
	recipients := make([][]byte, *count)
	addresses := make([]string, *count)
	for i := 0; i < *count; i++ {
		addr, err := ks.CreateKey(*password)
		if err != nil {
			fatal("create key: %v", err)
		}
		recipients[i] = addr
		addresses[i] = types.EncodeAddress(addr)
	}
	fmt.Fprintf(os.Stderr, "created in %.3fs\n", time.Since(createStart).Seconds())

	fmt.Fprintf(os.Stderr, "-- sign %d fund txs --\n", *count)
	signStart := time.Now()
	txs := make([]*types.Tx, *count)
	for i := 0; i < *count; i++ {
		txs[i] = signTransfer(chainID, creator, recipients[i], baseNonce+1+uint64(i), amount)
	}
	fmt.Fprintf(os.Stderr, "signed in %.3fs\n", time.Since(signStart).Seconds())

	hashes, _ := floodCommits(client, txs, *batchSize, *commitWorkers)

	fmt.Fprintf(os.Stderr, "-- wait fund confirmation --\n")
	confirmStart := time.Now()
	confirmLimit := time.Duration(*confirmTimeout) * time.Second
	stuckLimit := time.Duration(*stuckTimeout) * time.Second
	waitReceipt(client, hashes[len(hashes)-1], confirmLimit, stuckLimit)
	fmt.Fprintf(os.Stderr, "funded in %.3fs\n", time.Since(confirmStart).Seconds())

	var sb strings.Builder
	for _, a := range addresses {
		sb.WriteString(a)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(*accountsOut, []byte(sb.String()), 0644); err != nil {
		fatal("write accounts: %v", err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", *accountsOut)
}

func runFlood(args []string) {
	fs := flag.NewFlagSet("flood", flag.ExitOnError)
	rpc := fs.String("rpc", "127.0.0.1:7845", "gRPC address")
	keystore := fs.String("keystore", "", "keystore directory")
	password := fs.String("password", "", "keystore password")
	contractStr := fs.String("contract", "", "contract address")
	accountsFile := fs.String("accounts-file", "", "accounts file")
	txsPerAccount := fs.Int("txs-per-account", 40, "txs per account")
	batchSize := fs.Int("batch", 500, "commit batch size")
	signWorkers := fs.Int("sign-workers", runtime.NumCPU(), "sign workers")
	commitWorkers := fs.Int("commit-workers", 20, "parallel commit batches")
	queryCount := fs.Int("query-count", 1000, "query count")
	queryWorkers := fs.Int("query-workers", 40, "query workers")
	confirmTimeout := fs.Int("confirm-timeout", 120, "confirmation timeout in seconds")
	confirmWorkers := fs.Int("confirm-workers", 8, "parallel block scanners")
	stuckTimeout := fs.Int("stuck-timeout", 30, "fail if chain height unchanged for this many seconds")
	resultsFile := fs.String("results", "", "results output file")
	_ = fs.Parse(args)

	if *keystore == "" || *password == "" || *contractStr == "" || *accountsFile == "" {
		fatal("need --keystore --password --contract --accounts-file")
	}

	accounts := readAccounts(*accountsFile)
	keys := loadKeys(*keystore, *password, accounts)
	contract, err := types.DecodeAddress(*contractStr)
	if err != nil {
		fatal("contract address: %v", err)
	}

	client, admin, closeConn := dialRPC(*rpc)
	defer closeConn()

	st, err := client.Blockchain(context.Background(), &types.Empty{})
	if err != nil {
		fatal("blockchain: %v", err)
	}
	chainID := st.GetBestChainIdHash()

	numAccounts := len(accounts)
	writeCount := numAccounts * (*txsPerAccount)

	type job struct {
		idx   int
		nonce uint64
		acct  int
	}
	jobs := make([]job, 0, writeCount)
	for j := 1; j <= *txsPerAccount; j++ {
		for i := 0; i < numAccounts; i++ {
			idx := (j-1)*numAccounts + (i + 1)
			jobs = append(jobs, job{idx: idx, nonce: uint64(j), acct: i})
		}
	}

	// Phase 1: pre-sign (not part of send/confirm throughput)
	fmt.Fprintf(os.Stderr, "-- pre-sign %d txs (%d workers) --\n", writeCount, *signWorkers)
	signStart := time.Now()
	signed := make([]*types.Tx, writeCount)
	work := make(chan int, *signWorkers*2)
	var signWG sync.WaitGroup
	for w := 0; w < *signWorkers; w++ {
		signWG.Add(1)
		go func() {
			defer signWG.Done()
			for ji := range work {
				j := jobs[ji]
				payload := callPayload(j.idx)
				signed[ji] = signCall(chainID, contract, keys[j.acct], j.nonce, payload)
			}
		}()
	}
	for i := range jobs {
		work <- i
	}
	close(work)
	signWG.Wait()
	signDur := time.Since(signStart).Seconds()
	fmt.Fprintf(os.Stderr, "pre-signed in %.3fs (%.0f sign/s)\n", signDur, float64(writeCount)/signDur)

	heightBefore := chainHeight(client)

	allHashes, sendDur := floodCommits(client, signed, *batchSize, *commitWorkers)
	scanFrom := heightBefore // txs land during commit too; scan from next block

	fmt.Fprintf(os.Stderr, "-- wait confirmation (%d txs, scan from block %d, %d workers) --\n",
		len(allHashes), scanFrom+1, *confirmWorkers)
	confirmStart := time.Now()
	confirmLimit := time.Duration(*confirmTimeout) * time.Second
	stuckLimit := time.Duration(*stuckTimeout) * time.Second
	confirm := waitFloodConfirm(client, admin, allHashes, scanFrom, confirmLimit, stuckLimit, *confirmWorkers)
	confirmDur := time.Since(confirmStart).Seconds()
	totalDur := sendDur + confirmDur
	confirmedCount := confirm.confirmedCount

	firstBlock := confirm.firstBlock
	lastBlock := confirm.lastBlock
	blocks := int(lastBlock - firstBlock + 1)
	if blocks < 1 {
		blocks = 1
	}

	peak := confirm.peak
	totalInBlocks := 0
	for b := firstBlock; b <= lastBlock; b++ {
		ntx := confirm.blockTxCounts[b]
		totalInBlocks += ntx
		if b <= firstBlock+8 || b >= lastBlock-2 {
			fmt.Fprintf(os.Stderr, "  block %d: %d txs\n", b, ntx)
		} else if b == firstBlock+9 {
			fmt.Fprintf(os.Stderr, "  ...\n")
		}
	}
	heightAfter := chainHeight(client)
	mpEnd := -1
	if mp, ok := mempoolPending(admin); ok {
		mpEnd = mp
	}
	fmt.Fprintf(os.Stderr, "  confirmed %d/%d txs, blocks %d..%d (%d blocks, chain +%d), peak %d tx/block",
		confirmedCount, writeCount, firstBlock, lastBlock, blocks, heightAfter-heightBefore, peak)
	if mpEnd >= 0 {
		fmt.Fprintf(os.Stderr, ", mempool %d", mpEnd)
	}
	fmt.Fprintln(os.Stderr)

	// Let state settle before read benchmark
	time.Sleep(2 * time.Second)

	// Phase 4: queries (separate timer)
	fmt.Fprintf(os.Stderr, "-- %d queries (%d workers) --\n", *queryCount, *queryWorkers)
	qStart := time.Now()
	qErrors := int64(0)
	qWork := make(chan int, *queryWorkers*2)
	var qwg sync.WaitGroup
	for w := 0; w < *queryWorkers; w++ {
		qwg.Add(1)
		go func() {
			defer qwg.Done()
			for k := range qWork {
				ci := types.CallInfo{
					Name: "get_value",
					Args: []interface{}{fmt.Sprintf("key_%d", k)},
				}
				payload, _ := json.Marshal(ci)
				ret, err := client.QueryContract(context.Background(), &types.Query{
					ContractAddress: contract,
					Queryinfo:       payload,
				})
				if err != nil {
					atomic.AddInt64(&qErrors, 1)
					continue
				}
				want := fmt.Sprintf("value_%d", k)
				got := strings.TrimSpace(string(ret.GetValue()))
				got = strings.TrimPrefix(got, "value:")
				got = strings.ReplaceAll(got, `"`, "")
				got = strings.ReplaceAll(got, `\`, "")
				got = strings.ReplaceAll(got, " ", "")
				if got != want {
					atomic.AddInt64(&qErrors, 1)
				}
			}
		}()
	}
	for k := 1; k <= *queryCount; k++ {
		qWork <- k
	}
	close(qWork)
	qwg.Wait()
	qDur := time.Since(qStart).Seconds()

	out := []string{
		fmt.Sprintf("write_tx_count=%d", writeCount),
		fmt.Sprintf("committed_count=%d", len(allHashes)),
		fmt.Sprintf("confirmed_count=%d", confirmedCount),
		fmt.Sprintf("mined_tx_count=%d", confirm.minedTxCount),
		fmt.Sprintf("scan_from_block=%d", scanFrom+1),
		fmt.Sprintf("scan_through_block=%d", confirm.scanThrough),
		fmt.Sprintf("sign_seconds=%.3f", signDur),
		fmt.Sprintf("send_seconds=%.3f", sendDur),
		fmt.Sprintf("confirm_seconds=%.3f", confirmDur),
		fmt.Sprintf("total_seconds=%.3f", totalDur),
		fmt.Sprintf("send_tps=%.2f", float64(len(allHashes))/sendDur),
		fmt.Sprintf("confirm_tps=%.2f", float64(writeCount)/confirmDur),
		fmt.Sprintf("total_tps=%.2f", float64(writeCount)/totalDur),
		fmt.Sprintf("first_block=%d", firstBlock),
		fmt.Sprintf("last_block=%d", lastBlock),
		fmt.Sprintf("blocks_used=%d", blocks),
		fmt.Sprintf("chain_blocks_delta=%d", heightAfter-heightBefore),
		fmt.Sprintf("peak_tx_per_block=%d", peak),
		fmt.Sprintf("tx_per_block=%.2f", float64(writeCount)/float64(blocks)),
		fmt.Sprintf("tx_in_block_span=%d", totalInBlocks),
		fmt.Sprintf("query_count=%d", *queryCount),
		fmt.Sprintf("query_seconds=%.3f", qDur),
		fmt.Sprintf("query_qps=%.2f", float64(*queryCount)/qDur),
		fmt.Sprintf("query_errors=%d", atomic.LoadInt64(&qErrors)),
		// legacy keys for comparison script
		fmt.Sprintf("write_send_seconds=%.3f", sendDur),
		fmt.Sprintf("write_total_seconds=%.3f", totalDur),
		fmt.Sprintf("write_tps_send=%.2f", float64(len(allHashes))/sendDur),
		fmt.Sprintf("write_tps_total=%.2f", float64(writeCount)/totalDur),
	}
	if mpEnd >= 0 {
		out = append(out, fmt.Sprintf("mempool_remaining=%d", mpEnd))
	}
	text := strings.Join(out, "\n") + "\n"
	if *resultsFile != "" {
		if err := os.WriteFile(*resultsFile, []byte(text), 0644); err != nil {
			fatal("write results: %v", err)
		}
	} else {
		fmt.Print(text)
	}
	fmt.Fprintf(os.Stderr, "send: %.3fs | confirm: %.3fs | total: %.3fs | confirmed %d/%d | %d blocks | peak: %d tx/block | queries: %.3fs errors=%d\n",
		sendDur, confirmDur, totalDur, confirmedCount, writeCount, blocks, peak, qDur, atomic.LoadInt64(&qErrors))
}
