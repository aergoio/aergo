package contract

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.1
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
#include "lgmp.h"
*/
import "C"
import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"index/suffixarray"
	"math/big"
	"regexp"
	"strings"
	"unsafe"

	luacUtil "github.com/aergoio/aergo/cmd/aergoluac/util"
	"github.com/aergoio/aergo/contract/name"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/minio/sha256-simd"
)

var mulAergo, mulGaer, zeroBig *big.Int

const maxEventCnt = 50
const maxEventNameSize = 64
const maxEventArgSize = 4096
const luaCallCountDeduc = 1000

func init() {
	mulAergo, _ = new(big.Int).SetString("1000000000000000000", 10)
	mulGaer, _ = new(big.Int).SetString("1000000000", 10)
	zeroBig = big.NewInt(0)
}

func addUpdateSize(s *StateSet, updateSize int64) error {
	if s.dbUpdateTotalSize+updateSize > dbUpdateMaxLimit {
		return errors.New("exceeded size of updates in the state database")
	}
	s.dbUpdateTotalSize += updateSize
	return nil
}

//export LuaSetDB
func LuaSetDB(L *LState, service *C.int, key *C.char, value *C.char) *C.char {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return C.CString("[System.LuaSetDB] contract state not found")
	}
	if stateSet.isQuery == true {
		return C.CString("[System.LuaSetDB] set not permitted in query")
	}
	val := []byte(C.GoString(value))
	if err := stateSet.curContract.callState.ctrState.SetData([]byte(C.GoString(key)), val); err != nil {
		return C.CString(err.Error())
	}
	if err := addUpdateSize(stateSet, int64(types.HashIDLength+len(val))); err != nil {
		C.luaL_setuncatchablerror(L)
		return C.CString(err.Error())
	}
	return nil
}

//export LuaGetDB
func LuaGetDB(L *LState, service *C.int, key *C.char) (*C.char, *C.char) {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return nil, C.CString("[System.LuaGetDB] contract state not found")
	}
	data, err := stateSet.curContract.callState.ctrState.GetData([]byte(C.GoString(key)))
	if err != nil {
		return nil, C.CString(err.Error())
	}
	if data == nil {
		return nil, nil
	}
	return C.CString(string(data)), nil
}

//export LuaDelDB
func LuaDelDB(L *LState, service *C.int, key *C.char) *C.char {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return C.CString("[System.LuaDelDB] contract state not found")
	}
	if stateSet.isQuery {
		return C.CString("[System.LuaDelDB] delete not permitted in query")
	}
	if err := stateSet.curContract.callState.ctrState.DeleteData([]byte(C.GoString(key))); err != nil {
		return C.CString(err.Error())
	}
	if err := addUpdateSize(stateSet, int64(32)); err != nil {
		C.luaL_setuncatchablerror(L)
		return C.CString(err.Error())
	}
	return nil
}

func getCallState(stateSet *StateSet, aid types.AccountID) (*CallState, error) {
	callState := stateSet.callState[aid]
	if callState == nil {
		bs := stateSet.bs

		prevState, err := bs.GetAccountState(aid)
		if err != nil {
			return nil, err
		}

		curState := types.Clone(*prevState).(types.State)
		callState =
			&CallState{prevState: prevState, curState: &curState}
		stateSet.callState[aid] = callState
	}
	return callState, nil
}

func getCtrState(stateSet *StateSet, aid types.AccountID) (*CallState, error) {
	callState, err := getCallState(stateSet, aid)
	if err != nil {
		return nil, err
	}
	if callState.ctrState == nil {
		callState.ctrState, err = stateSet.bs.OpenContractState(aid, callState.curState)
	}
	return callState, err
}

func setInstCount(parent *LState, child *LState) {
	C.luaL_setinstcount(parent, C.luaL_instcount(child))
}

func setInstMinusCount(L *LState, deduc C.int) {
	C.luaL_setinstcount(L, minusCallCount(C.luaL_instcount(L), deduc))
}

func minusCallCount(curCount C.int, deduc C.int) C.int {
	remain := curCount - deduc
	if remain <= 0 {
		remain = 1
	}
	return remain
}

//export LuaCallContract
func LuaCallContract(L *LState, service *C.int, contractId *C.char, fname *C.char, args *C.char,
	amount *C.char, gas uint64) (C.int, *C.char) {
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	stateSet := curStateSet[*service]
	if stateSet == nil {
		return -1, C.CString("[Contract.LuaCallContract] contract state not found")
	}
	contractAddress := C.GoString(contractId)
	cid, err := getAddressNameResolved(contractAddress, stateSet.bs)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)
	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid amount: " + err.Error())
	}

	callState, err := getCtrState(stateSet, aid)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] getAccount error: " + err.Error())
	}

	callee := getContract(callState.ctrState, nil)
	if callee == nil {
		return -1, C.CString("[Contract.LuaCallContract] cannot find contract " + C.GoString(contractId))
	}

	prevContractInfo := stateSet.curContract

	var ci types.CallInfo
	ci.Name = fnameStr
	err = getCallInfo(&ci.Args, []byte(argsStr), cid)
	if err != nil {
		return -1, C.CString("[Contract.LuaCallContract] invalid arguments: " + err.Error())
	}

	ce := newExecutor(callee, cid, stateSet, &ci, amountBig, false)
	defer ce.close()

	if ce.err != nil {
		return -1, C.CString("[Contract.LuaCallContract] newExecutor error: " + ce.err.Error())
	}

	senderState := prevContractInfo.callState.curState
	if amountBig.Cmp(zeroBig) > 0 {
		if stateSet.isQuery == true {
			return -1, C.CString("[Contract.LuaCallContract] send not permitted in query")
		}
		if r := sendBalance(L, senderState, callState.curState, amountBig); r != nil {
			return -1, r
		}
	}
	if stateSet.lastRecoveryEntry != nil {
		err = setRecoveryPoint(aid, stateSet, senderState, callState, amountBig, false)
		if err != nil {
			C.luaL_setsyserror(L)
			return -1, C.CString("[System.LuaCallContract] database error: " + err.Error())
		}
	}
	stateSet.curContract = newContractInfo(callState, prevContractInfo.contractId, cid,
		callState.curState.SqlRecoveryPoint, amountBig)

	ce.setCountHook(minusCallCount(C.luaL_instcount(L), luaCallCountDeduc))
	defer setInstCount(L, ce.L)

	ret := ce.call(L)
	if ce.err != nil {
		stateSet.curContract = prevContractInfo
		return -1, C.CString("[Contract.LuaCallContract] call err: " + ce.err.Error())
	}
	stateSet.curContract = prevContractInfo
	return ret, nil
}

func getOnlyContractState(stateSet *StateSet, aid types.AccountID) (*state.ContractState, error) {
	callState := stateSet.callState[aid]
	if callState == nil || callState.ctrState == nil {
		return stateSet.bs.OpenContractStateAccount(aid)
	}
	return callState.ctrState, nil
}

//export LuaDelegateCallContract
func LuaDelegateCallContract(L *LState, service *C.int, contractId *C.char,
	fname *C.char, args *C.char, gas uint64) (C.int, *C.char) {
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	stateSet := curStateSet[*service]
	if stateSet == nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] contract state not found")
	}
	cid, err := getAddressNameResolved(contractIdStr, stateSet.bs)
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)
	contractState, err := getOnlyContractState(stateSet, aid)
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract]getContractState error" + err.Error())
	}
	contract := getContract(contractState, nil)
	if contract == nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] cannot find contract " + contractIdStr)
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = getCallInfo(&ci.Args, []byte(argsStr), cid)
	if err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] invalid arguments: " + err.Error())
	}

	ce := newExecutor(contract, cid, stateSet, &ci, zeroBig, false)
	defer ce.close()

	if ce.err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] newExecutor error: " + ce.err.Error())
	}

	if stateSet.lastRecoveryEntry != nil {
		callState := stateSet.curContract.callState
		err = setRecoveryPoint(aid, stateSet, nil, callState, zeroBig, false)
		if err != nil {
			C.luaL_setsyserror(L)
			return -1, C.CString("[System.LuaDelegateCallContract] database error: " + err.Error())
		}
	}

	ce.setCountHook(minusCallCount(C.luaL_instcount(L), luaCallCountDeduc))
	defer setInstCount(L, ce.L)

	ret := ce.call(L)
	if ce.err != nil {
		return -1, C.CString("[Contract.LuaDelegateCallContract] call error: " + ce.err.Error())
	}
	return ret, nil
}

func getAddressNameResolved(account string, bs *state.BlockState) ([]byte, error) {
	accountLen := len(account)
	if accountLen == types.EncodedAddressLength {
		return types.DecodeAddress(account)
	} else if accountLen == types.NameLength {
		cid := name.Resolve(bs, []byte(account))
		if cid == nil {
			return nil, errors.New("name not founded :" + account)
		}
		return cid, nil
	}
	return nil, errors.New("invalid account length:" + account)
}

//export LuaSendAmount
func LuaSendAmount(L *LState, service *C.int, contractId *C.char, amount *C.char) *C.char {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return C.CString("[Contract.LuaSendAmount] contract state not found")
	}
	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] invalid amount: " + err.Error())
	}
	if stateSet.isQuery == true && amountBig.Cmp(zeroBig) > 0 {
		return C.CString("[Contract.LuaSendAmount] send not permitted in query")
	}
	cid, err := getAddressNameResolved(C.GoString(contractId), stateSet.bs)
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] invalid contractId: " + err.Error())
	}

	aid := types.ToAccountID(cid)
	callState, err := getCallState(stateSet, aid)
	if err != nil {
		return C.CString("[Contract.LuaSendAmount] getAccount error: " + err.Error())
	}

	senderState := stateSet.curContract.callState.curState
	if len(callState.curState.GetCodeHash()) > 0 {
		if callState.ctrState == nil {
			callState.ctrState, err = stateSet.bs.OpenContractState(aid, callState.curState)
			if err != nil {
				return C.CString("[Contract.LuaSendAmount] getContractState error: " + err.Error())
			}
		}
		var ci types.CallInfo
		ci.Name = "default"
		code := getContract(callState.ctrState, nil)
		if code == nil {
			return C.CString("[Contract.LuaSendAmount] cannot find contract:" + C.GoString(contractId))
		}

		ce := newExecutor(code, cid, stateSet, &ci, amountBig, false)
		defer ce.close()
		if ce.err != nil {
			return C.CString("[Contract.LuaSendAmount] newExecutor error: " + ce.err.Error())
		}

		if amountBig.Cmp(zeroBig) > 0 {
			if r := sendBalance(L, senderState, callState.curState, amountBig); r != nil {
				return r
			}
		}
		if stateSet.lastRecoveryEntry != nil {
			err = setRecoveryPoint(aid, stateSet, senderState, callState, amountBig, false)
			if err != nil {
				C.luaL_setsyserror(L)
				return C.CString("[System.LuaSendAmount] database error: " + err.Error())
			}
		}
		prevContractInfo := stateSet.curContract
		stateSet.curContract = newContractInfo(callState, prevContractInfo.contractId, cid,
			callState.curState.SqlRecoveryPoint, amountBig)

		ce.setCountHook(minusCallCount(C.luaL_instcount(L), luaCallCountDeduc))
		defer setInstCount(L, ce.L)

		ce.call(L)
		if ce.err != nil {
			stateSet.curContract = prevContractInfo
			return C.CString("[Contract.LuaSendAmount] call err: " + ce.err.Error())
		}
		stateSet.curContract = prevContractInfo
		return nil
	}
	if amountBig.Cmp(zeroBig) == 0 {
		return nil
	}

	if r := sendBalance(L, senderState, callState.curState, amountBig); r != nil {
		return r
	}
	if stateSet.lastRecoveryEntry != nil {
		_ = setRecoveryPoint(aid, stateSet, senderState, callState, amountBig, true)
	}
	return nil
}

func sendBalance(L *LState, sender *types.State, receiver *types.State, amount *big.Int) *C.char {
	if sender == receiver {
		return nil
	}
	if sender.GetBalanceBigInt().Cmp(amount) < 0 {
		return C.CString("[Contract.sendBalance] insufficient balance: " +
			sender.GetBalanceBigInt().String() + " : " + amount.String())
	} else {
		sender.Balance = new(big.Int).Sub(sender.GetBalanceBigInt(), amount).Bytes()
	}
	receiver.Balance = new(big.Int).Add(receiver.GetBalanceBigInt(), amount).Bytes()

	return nil
}

//export LuaPrint
func LuaPrint(L *LState, service *C.int, args *C.char) {
	stateSet := curStateSet[*service]
	setInstMinusCount(L, 1000)
	logger.Info().Str("Contract SystemPrint", types.EncodeAddress(stateSet.curContract.contractId)).Msg(C.GoString(args))
}

func setRecoveryPoint(aid types.AccountID, stateSet *StateSet, senderState *types.State,
	callState *CallState, amount *big.Int, isSend bool) error {
	var seq int
	prev := stateSet.lastRecoveryEntry
	if prev != nil {
		seq = prev.seq + 1
	} else {
		seq = 1
	}
	recoveryEntry := &recoveryEntry{
		seq,
		amount,
		senderState,
		senderState.GetNonce(),
		callState,
		isSend,
		nil,
		-1,
		prev,
	}
	stateSet.lastRecoveryEntry = recoveryEntry
	if isSend {
		return nil
	}
	recoveryEntry.stateRevision = callState.ctrState.Snapshot()
	tx := callState.tx
	if tx != nil {
		saveName := fmt.Sprintf("%s_%p", aid.String(), &recoveryEntry)
		err := tx.SubSavepoint(saveName)
		if err != nil {
			return err
		}
		recoveryEntry.sqlSaveName = &saveName
	}
	return nil
}

//export LuaSetRecoveryPoint
func LuaSetRecoveryPoint(L *LState, service *C.int) (C.int, *C.char) {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return -1, C.CString("[Contract.pcall] contract state not found")
	}
	if stateSet.isQuery == true {
		return 0, nil
	}
	curContract := stateSet.curContract
	err := setRecoveryPoint(types.ToAccountID(curContract.contractId), stateSet, nil,
		curContract.callState, zeroBig, false)
	if err != nil {
		C.luaL_setsyserror(L)
		return -1, C.CString("[Contract.pcall] database error: " + err.Error())
	}
	return C.int(stateSet.lastRecoveryEntry.seq), nil
}

//export LuaClearRecovery
func LuaClearRecovery(L *LState, service *C.int, start int, error bool) *C.char {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return C.CString("[Contract.pcall] contract state not found")
	}
	item := stateSet.lastRecoveryEntry
	for {
		if error {
			if item.recovery() != nil {
				C.luaL_setsyserror(L)
				return C.CString("[Contract.pcall] database error")
			}
		}
		if item.seq == start {
			if error || item.prev == nil {
				stateSet.lastRecoveryEntry = item.prev
			}
			return nil
		}
		item = item.prev
		if item == nil {
			return C.CString("[Contract.pcall] internal error")
		}
	}
}

//export LuaGetBalance
func LuaGetBalance(L *LState, service *C.int, contractId *C.char) (*C.char, *C.char) {
	stateSet := curStateSet[*service]
	if contractId == nil {
		return C.CString(stateSet.curContract.callState.ctrState.GetBalanceBigInt().String()), nil
	}
	cid, err := getAddressNameResolved(C.GoString(contractId), stateSet.bs)
	if err != nil {
		return nil, C.CString("[Contract.LuaGetBalance] invalid contractId: " + err.Error())
	}
	aid := types.ToAccountID(cid)
	callState := stateSet.callState[aid]
	if callState == nil {
		bs := stateSet.bs

		as, err := bs.GetAccountState(aid)
		if err != nil {
			return nil, C.CString("[Contract.LuaGetBalance] getAccount error: " + err.Error())
		}
		return C.CString(as.GetBalanceBigInt().String()), nil
	}
	return C.CString(callState.curState.GetBalanceBigInt().String()), nil
}

//export LuaGetSender
func LuaGetSender(L *LState, service *C.int) *C.char {
	stateSet := curStateSet[*service]
	setInstMinusCount(L, 1000)
	return C.CString(types.EncodeAddress(stateSet.curContract.sender))
}

//export LuaGetHash
func LuaGetHash(L *LState, service *C.int) *C.char {
	stateSet := curStateSet[*service]
	return C.CString(enc.ToString(stateSet.txHash))
}

//export LuaGetBlockNo
func LuaGetBlockNo(L *LState, service *C.int) C.lua_Integer {
	stateSet := curStateSet[*service]
	return C.lua_Integer(stateSet.blockHeight)
}

//export LuaGetTimeStamp
func LuaGetTimeStamp(L *LState, service *C.int) C.lua_Integer {
	stateSet := curStateSet[*service]
	return C.lua_Integer(stateSet.timestamp / 1e9)
}

//export LuaGetContractId
func LuaGetContractId(L *LState, service *C.int) *C.char {
	stateSet := curStateSet[*service]
	setInstMinusCount(L, 1000)
	return C.CString(types.EncodeAddress(stateSet.curContract.contractId))
}

//export LuaGetAmount
func LuaGetAmount(L *LState, service *C.int) *C.char {
	stateSet := curStateSet[*service]
	return C.CString(stateSet.curContract.amount.String())
}

//export LuaGetOrigin
func LuaGetOrigin(L *LState, service *C.int) *C.char {
	stateSet := curStateSet[*service]
	setInstMinusCount(L, 1000)
	return C.CString(types.EncodeAddress(stateSet.origin))
}

//export LuaGetPrevBlockHash
func LuaGetPrevBlockHash(L *LState, service *C.int) *C.char {
	stateSet := curStateSet[*service]
	return C.CString(enc.ToString(stateSet.prevBlockHash))
}

//export LuaGetDbHandle
func LuaGetDbHandle(service *C.int) *C.sqlite3 {
	stateSet := curStateSet[*service]
	curContract := stateSet.curContract
	callState := curContract.callState
	if callState.tx != nil {
		return callState.tx.GetHandle()
	}
	var tx Tx
	var err error

	aid := types.ToAccountID(curContract.contractId)
	if stateSet.isQuery == true {
		tx, err = BeginReadOnly(aid.String(), curContract.rp)
	} else {
		tx, err = BeginTx(aid.String(), curContract.rp)
	}
	if err != nil {
		logger.Error().Err(err).Msg("Begin SQL Transaction")
		return nil
	}
	if stateSet.isQuery == false {
		err = tx.Savepoint()
		if err != nil {
			logger.Error().Err(err).Msg("Begin SQL Transaction")
			return nil
		}
	}
	callState.tx = tx
	return callState.tx.GetHandle()
}

func checkHexString(data string) bool {
	if len(data) >= 2 && data[0] == '0' && (data[1] == 'x' || data[1] == 'X') {
		return true
	}
	return false
}

//export LuaCryptoSha256
func LuaCryptoSha256(L *LState, arg unsafe.Pointer, argLen C.int) (*C.char, *C.char) {
	data := C.GoBytes(arg, argLen)
	if checkHexString(string(data)) {
		dataStr := data[2:]
		var err error
		data, err = hex.DecodeString(string(dataStr))
		if err != nil {
			return nil, C.CString("[Contract.LuaCryptoSha256] hex decoding error: " + err.Error())
		}
	}
	h := sha256.New()
	h.Write(data)
	resultHash := h.Sum(nil)

	return C.CString("0x" + hex.EncodeToString(resultHash)), nil
}

func decodeHex(hexStr string) ([]byte, error) {
	if checkHexString(hexStr) {
		hexStr = hexStr[2:]
	}
	return hex.DecodeString(hexStr)
}

//export LuaECVerify
func LuaECVerify(L *LState, msg *C.char, sig *C.char, addr *C.char) (C.int, *C.char) {
	bMsg, err := decodeHex(C.GoString(msg))
	if err != nil {
		return -1, C.CString("[Contract.LuaEcVerify] invalid message format: " + err.Error())
	}
	bSig, err := decodeHex(C.GoString(sig))
	if err != nil {
		return -1, C.CString("[Contract.LuaEcVerify] invalid signature format: " + err.Error())
	}
	address := C.GoString(addr)
	setInstMinusCount(L, 10000)

	var pubKey *btcec.PublicKey
	var verifyResult bool
	isAergo := len(address) == types.EncodedAddressLength

	/*Aergo Address*/
	if isAergo {
		bAddress, err := types.DecodeAddress(address)
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] invalid aergo address: " + err.Error())
		}
		pubKey, err = btcec.ParsePubKey(bAddress, btcec.S256())
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] error parsing pubKey: " + err.Error())
		}
	}

	// CompactSign
	if len(bSig) == 65 {
		// ethereum
		if !isAergo {
			btcsig := make([]byte, 65)
			btcsig[0] = bSig[64] + 27
			copy(btcsig[1:], bSig)
			bSig = btcsig
		}
		pub, _, err := btcec.RecoverCompact(btcec.S256(), bSig, bMsg)
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] error recoverCompact: " + err.Error())
		}
		if pubKey != nil {
			verifyResult = pubKey.IsEqual(pub)
		} else {
			bAddress, err := decodeHex(address)
			if err != nil {
				return -1, C.CString("[Contract.LuaEcVerify] invalid Ethereum address: " + err.Error())
			}
			bPub := pub.SerializeUncompressed()
			h := sha256.New()
			h.Write(bPub[1:])
			signAddress := h.Sum(nil)[12:]
			verifyResult = bytes.Equal(bAddress, signAddress)
		}
	} else {
		sign, err := btcec.ParseSignature(bSig, btcec.S256())
		if err != nil {
			return -1, C.CString("[Contract.LuaEcVerify] error parsing signature: " + err.Error())
		}
		if pubKey == nil {
			return -1, C.CString("[Contract.LuaEcVerify] error recovering pubKey")
		}
		verifyResult = sign.Verify(bMsg, pubKey)
	}
	if verifyResult {
		return C.int(1), nil
	}
	return C.int(0), nil
}

func transformAmount(amountStr string) (*big.Int, error) {
	var ret *big.Int
	var prev int
	if len(amountStr) == 0 {
		return zeroBig, nil
	}
	index := suffixarray.New([]byte(amountStr))
	r := regexp.MustCompile("(?i)aergo|gaer|aer")

	res := index.FindAllIndex(r, -1)
	for _, pair := range res {
		amountBig, _ := new(big.Int).SetString(strings.TrimSpace(amountStr[prev:pair[0]]), 10)
		if amountBig == nil {
			return nil, errors.New("converting error for BigNum: " + amountStr[prev:])
		}
		cmp := amountBig.Cmp(zeroBig)
		if cmp < 0 {
			return nil, errors.New("negative amount not allowed")
		} else if cmp == 0 {
			prev = pair[1]
			continue
		}
		switch pair[1] - pair[0] {
		case 3:
		case 4:
			amountBig = new(big.Int).Mul(amountBig, mulGaer)
		case 5:
			amountBig = new(big.Int).Mul(amountBig, mulAergo)
		}
		if ret != nil {
			ret = new(big.Int).Add(ret, amountBig)
		} else {
			ret = amountBig
		}
		prev = pair[1]
	}

	if prev >= len(amountStr) {
		if ret != nil {
			return ret, nil
		} else {
			return zeroBig, nil
		}
	}
	num := strings.TrimSpace(amountStr[prev:])
	if len(num) == 0 {
		if ret != nil {
			return ret, nil
		} else {
			return zeroBig, nil
		}
	}

	amountBig, _ := new(big.Int).SetString(num, 10)

	if amountBig == nil {
		return nil, errors.New("converting error for Integer: " + amountStr[prev:])
	}
	if amountBig.Cmp(zeroBig) < 0 {
		return nil, errors.New("negative amount not allowed")
	}
	if ret != nil {
		ret = new(big.Int).Add(ret, amountBig)
	} else {
		ret = amountBig
	}
	return ret, nil
}

//export LuaDeployContract
func LuaDeployContract(
	L *LState,
	service *C.int,
	contract *C.char,
	args *C.char,
	amount *C.char,
) (C.int, *C.char) {

	argsStr := C.GoString(args)
	contractStr := C.GoString(contract)

	stateSet := curStateSet[*service]
	if stateSet == nil {
		return -1, C.CString("[Contract.LuaDeployContract]not found contract state")
	}
	if stateSet.isQuery == true {
		return -1, C.CString("[Contract.LuaDeployContract]send not permitted in query")
	}
	bs := stateSet.bs

	// get code
	var code []byte

	cid, err := getAddressNameResolved(contractStr, bs)
	if err == nil {
		aid := types.ToAccountID(cid)
		contractState, err := getOnlyContractState(stateSet, aid)
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]" + err.Error())
		}
		code, err = contractState.GetCode()
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]" + err.Error())
		} else if len(code) == 0 {
			return -1, C.CString("[Contract.LuaDeployContract]: not found code")
		}
	}

	if len(code) == 0 {
		l := luacUtil.NewLState()
		if l == nil {
			return -1, C.CString("[Contract.LuaDeployContract] get luaState error")
		}
		defer luacUtil.CloseLState(l)
		code, err = luacUtil.Compile(l, contractStr)
		if err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]compile error:" + err.Error())
		}
	}

	err = addUpdateSize(stateSet, int64(len(code)))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	// create account
	prevContractInfo := stateSet.curContract
	creator := prevContractInfo.callState.curState
	newContract, err := bs.CreateAccountStateV(CreateContractID(prevContractInfo.contractId, creator.GetNonce()))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}
	contractState, err := bs.OpenContractState(newContract.AccountID(), newContract.State())
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	callState := &CallState{ctrState: contractState, prevState: &types.State{}, curState: newContract.State()}
	stateSet.callState[newContract.AccountID()] = callState

	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]value not proper format:" + err.Error())
	}
	var ci types.CallInfo
	err = getCallInfo(&ci.Args, []byte(argsStr), newContract.ID())
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract] invalid args:" + err.Error())
	}
	runCode := getContract(contractState, code)

	senderState := prevContractInfo.callState.curState
	if amountBig.Cmp(zeroBig) > 0 {
		if rv := sendBalance(L, senderState, callState.curState, amountBig); rv != nil {
			return -1, rv
		}
	}

	if stateSet.lastRecoveryEntry != nil {
		err = setRecoveryPoint(newContract.AccountID(), stateSet, senderState, callState, amountBig, false)
		if err != nil {
			C.luaL_setsyserror(L)
			return -1, C.CString("[System.LuaDeployContract] DB err:" + err.Error())
		}
	}
	stateSet.curContract = newContractInfo(callState, prevContractInfo.contractId, newContract.ID(),
		callState.curState.SqlRecoveryPoint, amountBig)

	err = contractState.SetCode(code)
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}
	err = contractState.SetData([]byte("Creator"), []byte(types.EncodeAddress(prevContractInfo.contractId)))
	if err != nil {
		return -1, C.CString("[Contract.LuaDeployContract]:" + err.Error())
	}

	ce := newExecutor(runCode, newContract.ID(), stateSet, &ci, amountBig, true)
	if ce != nil {
		defer ce.close()
		if ce.err != nil {
			return -1, C.CString("[Contract.LuaDeployContract]newExecutor Error :" + ce.err.Error())
		}
	}

	// create a sql database for the contract
	db := LuaGetDbHandle(&stateSet.service)
	if db == nil {
		C.luaL_setsyserror(L)
		return -1, C.CString("[System.LuaDeployContract] DB err: cannot open a database")
	}
	senderState.Nonce += 1

	addr := C.CString(types.EncodeAddress(newContract.ID()))
	ret := C.int(1)
	if ce != nil {
		ce.setCountHook(minusCallCount(C.luaL_instcount(L), luaCallCountDeduc))
		defer setInstCount(L, ce.L)

		ret += ce.call(L)
		if ce.err != nil {
			stateSet.curContract = prevContractInfo
			return -1, C.CString("[Contract.LuaDeployContract] call err:" + ce.err.Error())
		}
	}
	stateSet.curContract = prevContractInfo
	return ret, addr
}

//export IsPublic
func IsPublic() C.int {
	if PubNet {
		return C.int(1)
	} else {
		return C.int(0)
	}
}

//export LuaRandomInt
func LuaRandomInt(min, max, service C.int) C.int {
	stateSet := curStateSet[service]
	if stateSet.seed == nil {
		setRandomSeed(stateSet)
	}
	return C.int(stateSet.seed.Intn(int(max+C.int(1)-min)) + int(min))
}

//export LuaEvent
func LuaEvent(L *LState, service *C.int, eventName *C.char, args *C.char) *C.char {
	stateSet := curStateSet[*service]
	if stateSet.isQuery == true {
		return C.CString("[Contract.Event] event not permitted in query")
	}
	if stateSet.eventCount >= maxEventCnt {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum number of events(%d)", maxEventCnt))
	}
	if len(C.GoString(eventName)) > maxEventNameSize {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event name(%d)", maxEventNameSize))
	}
	if len(C.GoString(args)) > maxEventArgSize {
		return C.CString(fmt.Sprintf("[Contract.Event] exceeded the maximum length of event args(%d)", maxEventArgSize))
	}
	stateSet.events = append(
		stateSet.events,
		&types.Event{
			ContractAddress: stateSet.curContract.contractId,
			EventIdx:        stateSet.eventCount,
			EventName:       C.GoString(eventName),
			JsonArgs:        C.GoString(args),
		},
	)
	stateSet.eventCount++
	return nil
}

//export LuaIsContract
func LuaIsContract(L *LState, service *C.int, contractId *C.char) (C.int, *C.char) {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return -1, C.CString("[Contract.LuaIsContract] contract state not found")
	}
	cid, err := getAddressNameResolved(C.GoString(contractId), stateSet.bs)
	if err != nil {
		return -1, C.CString("[Contract.LuaIsContract] invalid contractId: " + err.Error())
	}

	aid := types.ToAccountID(cid)
	callState, err := getCallState(stateSet, aid)
	if err != nil {
		return -1, C.CString("[Contract.LuaIsContract] getAccount error: " + err.Error())
	}
	return C.int(len(callState.curState.GetCodeHash())), nil
}

//export LuaGovernance
func LuaGovernance(L *LState, service *C.int, gType C.char, arg *C.char) *C.char {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		return C.CString("[Contract.LuaGovernance] contract state not found")
	}
	var amountBig *big.Int
	var payload []byte

	if gType != 'V' {
		var err error
		amountBig, err = transformAmount(C.GoString(arg))
		if err != nil {
			return C.CString("[Contract.LuaGovernance] invalid amount: " + err.Error())
		}
		if stateSet.isQuery == true && amountBig.Cmp(zeroBig) > 0 {
			return C.CString("[Contract.LuaGovernance] governance not permitted in query")
		}
		if gType == 'S' {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Stake))
		} else {
			payload = []byte(fmt.Sprintf(`{"Name":"%s"}`, types.Unstake))
		}
	} else {
		amountBig = zeroBig
		payload = []byte(fmt.Sprintf(`{"Name":"%s","Args":%s}`, types.VoteBP, C.GoString(arg)))
	}
	aid := types.ToAccountID([]byte(types.AergoSystem))
	scsState, err := getCtrState(stateSet, aid)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] getAccount error: " + err.Error())
	}
	curContract := stateSet.curContract

	senderState := stateSet.curContract.callState.curState
	sender := stateSet.bs.InitAccountStateV(curContract.contractId,
		curContract.callState.prevState, curContract.callState.curState)
	receiver := stateSet.bs.InitAccountStateV([]byte(types.AergoSystem), scsState.prevState, scsState.curState)
	txBody := types.TxBody{
		Amount:  amountBig.Bytes(),
		Payload: payload,
	}
	err = types.ValidateSystemTx(&txBody)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] error: " + err.Error())
	}
	if stateSet.lastRecoveryEntry != nil {
		err = setRecoveryPoint(aid, stateSet, senderState, scsState, zeroBig, false)
		if err != nil {
			C.luaL_setsyserror(L)
			return C.CString("[Contract.LuaGovernance] database error: " + err.Error())
		}
	}
	evs, err := system.ExecuteSystemTx(scsState.ctrState, &txBody, sender, receiver, stateSet.blockHeight)
	if err != nil {
		return C.CString("[Contract.LuaGovernance] error: " + err.Error())
	}
	stateSet.eventCount += int32(len(evs))
	stateSet.events = append(stateSet.events, evs...)

	if stateSet.lastRecoveryEntry != nil {
		if gType == 'S' {
			_ = setRecoveryPoint(aid, stateSet, senderState, scsState, amountBig, true)
		} else if gType == 'U' {
			_ = setRecoveryPoint(aid, stateSet, scsState.curState, stateSet.curContract.callState, amountBig, true)
		}
	}
	return nil
}
