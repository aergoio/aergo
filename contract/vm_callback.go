package contract

/*
#cgo CFLAGS: -I${SRCDIR}/../libtool/include/luajit-2.1
#cgo LDFLAGS: ${SRCDIR}/../libtool/lib/libluajit-5.1.a -lm

#include <stdlib.h>
#include <string.h>
#include "vm.h"
#include "lbc.h"
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
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/minio/sha256-simd"
)

var mulAergo, mulGaer, zeroBig *big.Int

func init() {
	mulAergo, _ = new(big.Int).SetString("1000000000000000000", 10)
	mulGaer, _ = new(big.Int).SetString("1000000000", 10)
	zeroBig = big.NewInt(0)
}

func luaPushStr(L *LState, str string) {
	cStr := C.CString(str)
	C.lua_pushstring(L, cStr)
	C.free(unsafe.Pointer(cStr))
}

func addUpdateSize(s *StateSet, updateSize int64) error {
	if s.dbUpdateTotalSize+updateSize > dbUpdateMaxLimit {
		return errors.New("exceeded size of updates in the state database")
	}
	s.dbUpdateTotalSize += updateSize
	return nil
}

//export LuaSetDB
func LuaSetDB(L *LState, service *C.int, key *C.char, value *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaSetDB] contract state not found")
		return -1
	}
	if stateSet.isQuery == true {
		luaPushStr(L, "[System.LuaSetDB] set not permitted in query")
		return -1
	}
	val := []byte(C.GoString(value))
	if err := stateSet.curContract.callState.ctrState.SetData([]byte(C.GoString(key)), val); err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	if err := addUpdateSize(stateSet, int64(3*32+len(val))); err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	return 0
}

//export LuaGetDB
func LuaGetDB(L *LState, service *C.int, key *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaGetDB] contract state not found")
		return -1
	}

	data, err := stateSet.curContract.callState.ctrState.GetData([]byte(C.GoString(key)))
	if err != nil {
		luaPushStr(L, err.Error())
		return -1
	}

	if data == nil {
		return 0
	}
	luaPushStr(L, string(data))
	return 1
}

//export LuaDelDB
func LuaDelDB(L *LState, service *C.int, key *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[System.LuaDelDB] contract state not found")
		return -1
	}
	if stateSet.isQuery {
		luaPushStr(L, "[System.LuaDelDB] delete not permitted in query")
		return -1
	}
	if err := stateSet.curContract.callState.ctrState.DeleteData([]byte(C.GoString(key))); err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	if err := addUpdateSize(stateSet, int64(32)); err != nil {
		luaPushStr(L, err.Error())
		return -1
	}
	return 0
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

//export LuaCallContract
func LuaCallContract(L *LState, service *C.int, contractId *C.char, fname *C.char, args *C.char,
	amount *C.char, gas uint64) C.int {
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.LuaCallContract] contract state not found")
		return -1
	}
	cid, err := getAddressNameResolved(C.GoString(contractId), stateSet.bs)
	if err != nil {
		luaPushStr(L, "[Contract.LuaCallContract] invalid contractId: "+err.Error())
		return -1
	}
	aid := types.ToAccountID(cid)
	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		luaPushStr(L, "[Contract.LuaCallContract] invalid amount: "+err.Error())
		return -1
	}

	callState, err := getCallState(stateSet, aid)
	if err != nil {
		luaPushStr(L, "[Contract.LuaCallContract] getAccount error: "+err.Error())
		return -1
	}
	if callState.ctrState == nil {
		callState.ctrState, err = stateSet.bs.OpenContractState(aid, callState.curState)
		if err != nil {
			luaPushStr(L, "[Contract.LuaCallContract] getAccount error: "+err.Error())
			return -1
		}
	}

	callee := getContract(callState.ctrState, nil)
	if callee == nil {
		luaPushStr(L, "[Contract.LuaCallContract] cannot find contract "+C.GoString(contractId))
		return -1
	}

	prevContractInfo := stateSet.curContract

	ce := newExecutor(callee, stateSet)
	defer ce.close()

	if ce.err != nil {
		luaPushStr(L, "[Contract.LuaCallContract] newExecutor error: "+ce.err.Error())
		return -1
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = getCallInfo(&ci.Args, []byte(argsStr), cid)
	if err != nil {
		luaPushStr(L, "[Contract.LuaCallContract] invalid arguments: "+err.Error())
		return -1
	}
	senderState := prevContractInfo.callState.curState
	if amountBig.Cmp(zeroBig) > 0 {
		if stateSet.isQuery == true {
			luaPushStr(L, "[Contract.LuaCallContract] send not permitted in query")
			return -1
		}
		if sendBalance(L, senderState, callState.curState, amountBig) == false {
			return -1
		}
	}
	if stateSet.lastRecoveryEntry != nil {
		err = setRecoveryPoint(aid, stateSet, senderState, callState, amountBig, false)
		if err != nil {
			stateSet.dbSystemError = true
			luaPushStr(L, "[System.LuaCallContract] database error: "+err.Error())
		}
	}
	stateSet.curContract = newContractInfo(callState, prevContractInfo.contractId, cid,
		callState.curState.SqlRecoveryPoint, amountBig)
	ret := ce.call(&ci, L)
	if ce.err != nil {
		stateSet.curContract = prevContractInfo
		luaPushStr(L, "[Contract.LuaCallContract] call err: "+ce.err.Error())
		return -1
	}
	stateSet.curContract = prevContractInfo
	return ret
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
	fname *C.char, args *C.char, gas uint64) C.int {
	contractIdStr := C.GoString(contractId)
	fnameStr := C.GoString(fname)
	argsStr := C.GoString(args)

	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.LuaDelegateCallContract] contract state not found")
		return -1
	}
	cid, err := getAddressNameResolved(contractIdStr, stateSet.bs)
	if err != nil {
		luaPushStr(L, "[Contract.LuaDelegateCallContract] invalid contractId: "+err.Error())
		return -1
	}
	aid := types.ToAccountID(cid)
	contractState, err := getOnlyContractState(stateSet, aid)
	if err != nil {
		luaPushStr(L, "[Contract.LuaDelegateCallContract]getContractState error"+err.Error())
		return -1
	}
	contract := getContract(contractState, nil)
	if contract == nil {
		luaPushStr(L, "[Contract.LuaDelegateCallContract] cannot find contract "+contractIdStr)
		return -1
	}
	ce := newExecutor(contract, stateSet)
	defer ce.close()

	if ce.err != nil {
		luaPushStr(L, "[Contract.LuaDelegateCallContract] newExecutor error: "+ce.err.Error())
		return -1
	}

	var ci types.CallInfo
	ci.Name = fnameStr
	err = getCallInfo(&ci.Args, []byte(argsStr), cid)
	if err != nil {
		luaPushStr(L, "[Contract.LuaDelegateCallContract] invalid arguments: "+err.Error())
		return -1
	}

	if stateSet.lastRecoveryEntry != nil {
		callState := stateSet.curContract.callState
		err = setRecoveryPoint(aid, stateSet, nil, callState, zeroBig, false)
		if err != nil {
			stateSet.dbSystemError = true
			luaPushStr(L, "[System.LuaDelegateCallContract] database error: "+err.Error())
			return -1
		}
	}
	ret := ce.call(&ci, L)
	if ce.err != nil {
		luaPushStr(L, "[Contract.LuaDelegateCallContract] call error: "+ce.err.Error())
		return -1
	}
	return ret
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
func LuaSendAmount(L *LState, service *C.int, contractId *C.char, amount *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.LuaSendAmount] contract state not found")
		return -1
	}
	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		luaPushStr(L, "[Contract.LuaSendAmount] invalid amount: "+err.Error())
		return -1
	}
	if stateSet.isQuery == true && amountBig.Cmp(zeroBig) > 0 {
		luaPushStr(L, "[Contract.LuaSendAmount] send not permitted in query")
		return -1
	}
	cid, err := getAddressNameResolved(C.GoString(contractId), stateSet.bs)
	if err != nil {
		luaPushStr(L, "[Contract.LuaSendAmount] invalid contractId: "+err.Error())
		return -1
	}

	aid := types.ToAccountID(cid)
	callState, err := getCallState(stateSet, aid)
	if err != nil {
		luaPushStr(L, "[Contract.LuaSendAmount] getAccount error: "+err.Error())
		return -1
	}
	senderState := stateSet.curContract.callState.curState
	if sendBalance(L, senderState, callState.curState, amountBig) == false {
		return -1
	}
	if stateSet.lastRecoveryEntry != nil {
		err := setRecoveryPoint(aid, stateSet, senderState, callState, amountBig, true)
		if err != nil {
			stateSet.dbSystemError = true
			luaPushStr(L, "[Contract.LuaSendAmount] database error: "+err.Error())
			return -1
		}
	}
	return 0
}

func sendBalance(L *LState, sender *types.State, receiver *types.State, amount *big.Int) bool {
	if sender == receiver {
		return true
	}
	if sender.GetBalanceBigInt().Cmp(amount) < 0 {
		luaPushStr(L, "[Contract.sendBalance] insufficient balance: "+
			sender.GetBalanceBigInt().String()+" : "+amount.String())
		return false
	} else {
		sender.Balance = new(big.Int).Sub(sender.GetBalanceBigInt(), amount).Bytes()
	}
	receiver.Balance = new(big.Int).Add(receiver.GetBalanceBigInt(), amount).Bytes()

	return true
}

//export LuaPrint
func LuaPrint(service *C.int, args *C.char) {
	stateSet := curStateSet[*service]
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
func LuaSetRecoveryPoint(L *LState, service *C.int) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.pcall] contract state not found")
		return -1
	}
	if stateSet.isQuery == true {
		return 0
	}
	curContract := stateSet.curContract
	err := setRecoveryPoint(types.ToAccountID(curContract.contractId), stateSet, nil,
		curContract.callState, zeroBig, false)
	if err != nil {
		luaPushStr(L, "[Contract.pcall] database error: "+err.Error())
		stateSet.dbSystemError = true
		return -1
	}
	return C.int(stateSet.lastRecoveryEntry.seq)
}

//export LuaClearRecovery
func LuaClearRecovery(L *LState, service *C.int, start int, error bool) C.int {
	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.pcall] contract state not found")
		return -1
	}
	item := stateSet.lastRecoveryEntry
	for {
		if error {
			if item.recovery() != nil {
				stateSet.dbSystemError = true
				luaPushStr(L, "[Contract.pcall] database error")
				return -1
			}
		}
		if item.seq == start {
			if error || item.prev == nil {
				stateSet.lastRecoveryEntry = item.prev
			}
			return 0
		}
		item = item.prev
		if item == nil {
			luaPushStr(L, "[Contract.pcall] internal error")
			return -1
		}
	}
}

//export LuaGetBalance
func LuaGetBalance(L *LState, service *C.int, contractId *C.char) C.int {
	stateSet := curStateSet[*service]
	if contractId == nil {
		luaPushStr(L, stateSet.curContract.callState.ctrState.GetBalanceBigInt().String())
		return 0
	}
	cid, err := getAddressNameResolved(C.GoString(contractId), stateSet.bs)
	if err != nil {
		luaPushStr(L, "[Contract.LuaGetBalance] invalid contractId: "+err.Error())
		return -1
	}
	aid := types.ToAccountID(cid)
	callState := stateSet.callState[aid]
	if callState == nil {
		bs := stateSet.bs

		as, err := bs.GetAccountState(aid)
		if err != nil {
			luaPushStr(L, "[Contract.LuaGetBalance] getAccount error: "+err.Error())
			return -1
		}
		luaPushStr(L, as.GetBalanceBigInt().String())
	} else {
		luaPushStr(L, callState.curState.GetBalanceBigInt().String())
	}

	return 0
}

//export LuaGetSender
func LuaGetSender(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, types.EncodeAddress(stateSet.curContract.sender))
}

//export LuaGetHash
func LuaGetHash(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, enc.ToString(stateSet.txHash))
}

//export LuaGetBlockNo
func LuaGetBlockNo(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	C.lua_pushinteger(L, C.lua_Integer(stateSet.blockHeight))
}

//export LuaGetTimeStamp
func LuaGetTimeStamp(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	C.lua_pushinteger(L, C.lua_Integer(stateSet.timestamp/1e9))
}

//export LuaGetContractId
func LuaGetContractId(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, types.EncodeAddress(stateSet.curContract.contractId))
}

//export LuaGetAmount
func LuaGetAmount(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, stateSet.curContract.amount.String())
}

//export LuaGetOrigin
func LuaGetOrigin(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, types.EncodeAddress(stateSet.origin))
}

//export LuaGetPrevBlockHash
func LuaGetPrevBlockHash(L *LState, service *C.int) {
	stateSet := curStateSet[*service]

	luaPushStr(L, enc.ToString(stateSet.prevBlockHash))
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
		stateSet.dbSystemError = true
		logger.Error().Err(err).Msg("Begin SQL Transaction")
		return nil
	}
	if stateSet.isQuery == false {
		err = tx.Savepoint()
		if err != nil {
			stateSet.dbSystemError = true
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
func LuaCryptoSha256(L *LState, arg unsafe.Pointer, argLen C.int) C.int {
	data := C.GoBytes(arg, argLen)
	if checkHexString(string(data)) {
		dataStr := data[2:]
		var err error
		data, err = hex.DecodeString(string(dataStr))
		if err != nil {
			luaPushStr(L, "[Contract.LuaCryptoSha256] hex decoding error: "+err.Error())
			return -1
		}
	}
	h := sha256.New()
	h.Write(data)
	resultHash := h.Sum(nil)

	luaPushStr(L, "0x"+hex.EncodeToString(resultHash))
	return 0
}

func decodeHex(hexStr string) ([]byte, error) {
	if checkHexString(hexStr) {
		hexStr = hexStr[2:]
	}
	return hex.DecodeString(hexStr)
}

//export LuaECVerify
func LuaECVerify(L *LState, msg *C.char, sig *C.char, addr *C.char) C.int {
	bMsg, err := decodeHex(C.GoString(msg))
	if err != nil {
		luaPushStr(L, "[Contract.LuaEcVerify] invalid message format: "+err.Error())
		return -1
	}
	bSig, err := decodeHex(C.GoString(sig))
	if err != nil {
		luaPushStr(L, "[Contract.LuaEcVerify] invalid signature format: "+err.Error())
		return -1
	}
	address := C.GoString(addr)

	var pubKey *btcec.PublicKey
	var verifyResult bool
	isAergo := len(address) == types.EncodedAddressLength

	/*Aergo Address*/
	if isAergo {
		bAddress, err := types.DecodeAddress(address)
		if err != nil {
			luaPushStr(L, "[Contract.LuaEcVerify] invalid aergo address: "+err.Error())
			return -1
		}
		pubKey, err = btcec.ParsePubKey(bAddress, btcec.S256())
		if err != nil {
			luaPushStr(L, "[Contract.LuaEcVerify] error parsing pubKey: "+err.Error())
			return -1
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
			luaPushStr(L, "[Contract.LuaEcVerify] error recoverCompact: "+err.Error())
			return -1
		}
		if pubKey != nil {
			verifyResult = pubKey.IsEqual(pub)
		} else {
			bAddress, err := decodeHex(address)
			if err != nil {
				luaPushStr(L, "[Contract.LuaEcVerify] invalid Ethereum address: "+err.Error())
				return -1
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
			luaPushStr(L, "[Contract.LuaEcVerify] error parsing signature: "+err.Error())
			return -1
		}
		if pubKey == nil {
			luaPushStr(L, "[Contract.LuaEcVerify] error recovering pubKey")
			return -1
		}
		verifyResult = sign.Verify(bMsg, pubKey)
	}
	if verifyResult {
		C.lua_pushboolean(L, 1)
	} else {
		C.lua_pushboolean(L, C.int(0))
	}
	return 0
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
		return ret, nil
	}
	num := strings.TrimSpace(amountStr[prev:])
	if len(num) == 0 {
		return ret, nil
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
func LuaDeployContract(L *LState, service *C.int, contract *C.char, args *C.char, amount *C.char) C.int {
	argsStr := C.GoString(args)
	contractStr := C.GoString(contract)

	stateSet := curStateSet[*service]
	if stateSet == nil {
		luaPushStr(L, "[Contract.LuaCallContract]not found contract state")
		return -1
	}
	if stateSet.isQuery == true {
		luaPushStr(L, "[Contract.LuaCallContract]send not permitted in query")
		return -1
	}
	bs := stateSet.bs

	// get code
	var code []byte

	cid, err := getAddressNameResolved(contractStr, bs)
	if err == nil {
		aid := types.ToAccountID(cid)
		contractState, err := getOnlyContractState(stateSet, aid)
		if err != nil {
			luaPushStr(L, "[Contract.LuaDeployContract]"+err.Error())
			return -1
		}
		code, err = contractState.GetCode()
		if err != nil {
			luaPushStr(L, "[Contract.LuaDeployContract]"+err.Error())
			return -1
		} else if len(code) == 0 {
			luaPushStr(L, "[Contract.LuaDeployContract]: not found code")
			return -1
		}
	}

	if len(code) == 0 {
		l := luacUtil.NewLState()
		if l == nil {
			luaPushStr(L, "[Contract.LuaDeployContract]compile error:"+err.Error())
			return -1
		}
		defer luacUtil.CloseLState(l)
		code, err = luacUtil.Compile(l, contractStr)
		if err != nil {
			luaPushStr(L, "[Contract.LuaDeployContract]compile error:"+err.Error())
			return -1
		}
	}

	// create account
	prevContractInfo := stateSet.curContract
	creator := prevContractInfo.callState.curState
	newContract, err := bs.CreateAccountStateV(CreateContractID(prevContractInfo.contractId, creator.GetNonce()))
	if err != nil {
		luaPushStr(L, "[Contract.LuaDeployContract]:"+err.Error())
	}
	contractState, err := bs.OpenContractState(newContract.AccountID(), newContract.State())
	if err != nil {
		luaPushStr(L, "[Contract.LuaDeployContract]:"+err.Error())
	}

	callState := &CallState{ctrState: contractState, prevState: &types.State{}, curState: newContract.State()}
	stateSet.callState[newContract.AccountID()] = callState

	amountBig, err := transformAmount(C.GoString(amount))
	if err != nil {
		luaPushStr(L, "[Contract.LuaCallContract]value not proper format:"+err.Error())
		return -1
	}
	runCode := getContract(contractState, code)
	ce := newExecutor(runCode, stateSet)
	defer ce.close()
	if ce.err != nil {
		luaPushStr(L, "[Contract.LuaCallContract]newExecutor Error :"+ce.err.Error())
		return -1
	}
	var ci types.CallInfo
	err = getCallInfo(&ci.Args, []byte(argsStr), newContract.ID())
	if err != nil {
		luaPushStr(L, "[Contract.LuaCallContract] invalid args:"+err.Error())
		return -1
	}
	senderState := prevContractInfo.callState.curState
	if amountBig.Cmp(zeroBig) > 0 {
		if sendBalance(L, senderState, callState.curState, amountBig) == false {
			return -1
		}
	}

	if stateSet.lastRecoveryEntry != nil {
		err = setRecoveryPoint(newContract.AccountID(), stateSet, senderState, callState, amountBig, false)
		if err != nil {
			stateSet.dbSystemError = true
			luaPushStr(L, "[System.LuaCallContract] DB err:"+err.Error())
		}
	}
	stateSet.curContract = newContractInfo(callState, prevContractInfo.contractId, newContract.ID(),
		callState.curState.SqlRecoveryPoint, amountBig)

	err = contractState.SetCode(code)
	if err != nil {
		luaPushStr(L, "[Contract.LuaDeployContract]:"+err.Error())
	}
	err = contractState.SetData([]byte("Creator"), []byte(types.EncodeAddress(prevContractInfo.contractId)))
	if err != nil {
		luaPushStr(L, "[Contract.LuaDeployContract]:"+err.Error())
	}
	// create a sql database for the contract
	db := LuaGetDbHandle(&stateSet.service)
	if db == nil {
		stateSet.dbSystemError = true
		luaPushStr(L, "[System.LuaCallContract] DB err:"+err.Error())
		return -1
	}
	senderState.Nonce += 1

	luaPushStr(L, types.EncodeAddress(newContract.ID()))
	ret := ce.constructCall(&ci, L)
	if ce.err != nil {
		stateSet.curContract = prevContractInfo
		luaPushStr(L, "[Contract.LuaCallContract] call err:"+ce.err.Error())
		return -1
	}
	stateSet.curContract = prevContractInfo
	return ret + 1
}

//export IsPublic
func IsPublic() C.int {
	if PubNet {
		return C.int(1)
	} else {
		return C.int(0)
	}
}

//export LuaRandom
func LuaRandom(L *LState, service C.int) C.int {
	stateSet := curStateSet[service]
	switch C.lua_gettop(L) {
	case 0:
		C.lua_pushnumber(L, C.double(stateSet.seed.Float64()))
	case 1:
		n := C.luaL_checkinteger(L, 1)
		if n < 1 {
			luaPushStr(L, "system.random: the maximum value must be greater than zero")
			return -1
		}
		C.lua_pushinteger(L, C.lua_Integer(stateSet.seed.Intn(int(n)))+C.lua_Integer(1))
	default:
		min := C.luaL_checkinteger(L, 1)
		max := C.luaL_checkinteger(L, 2)
		if min < 1 {
			luaPushStr(L, "system.random: the minimum value must be greater than zero")
			return -1
		}
		if min > max {
			luaPushStr(L, "system.random: the maximum value must be greater than the minimum value")
			return -1
		}
		C.lua_pushinteger(L, C.lua_Integer(stateSet.seed.Intn(int(max+C.lua_Integer(1)-min)))+min)
	}
	return 1
}

//export LuaEvent
func LuaEvent(L *LState, service *C.int, eventName *C.char, args *C.char) C.int {
	stateSet := curStateSet[*service]
	if stateSet.isQuery == true {
		luaPushStr(L, "[Contract.Event] event not permitted in query")
		return -1
	}
	stateSet.events = append(stateSet.events,
		&types.Event{
			ContractAddress: stateSet.curContract.contractId,
			EventIdx:        stateSet.eventCount,
			EventName:       C.GoString(eventName),
			JsonArgs:        C.GoString(args),
		})
	stateSet.eventCount++

	return 0
}
