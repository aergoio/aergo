package contract
/*
#include "db_module.h"
*/
import "C"
import (
	"encoding/json"
	"encoding/binary"
	"errors"
	"strconv"
	"unsafe"
	"time"
	"github.com/aergoio/aergo/v2/contract/msg"
	"github.com/aergoio/aergo/v2/types"
)

// convert the arguments to a single string containing the JSON array
func (ce *executor) convertArgsToJSON() (string, error) {
	if ce.ci == nil || ce.ci.Args == nil {
		return "[]", nil
	}
	args, err := json.Marshal(ce.ci.Args)
	if err != nil {
		return "", err
	}
	return string(args), nil
}

func (ce *executor) call(extractUsedGas bool) {

	if ce.err != nil {
		return
	}
	if ce.preErr != nil {
		ce.err = ce.preErr
		return
	}

	/*
	defer func() {
		if PubNet == false {
			C.db_release_resource()
			// test: A -> B -> C, what happens if A and C use the same db? (same contract)
			// maybe contract C should release only its own resources
			// the struct could store the instance id, and release only its own resources
		}
	}()
	*/

	if ce.isView == true {
		ce.ctx.nestedView++
		defer func() {
			ce.ctx.nestedView--
		}()
	}

	// what to send:
	// - address: types.EncodeAddress(ce.ctx.curContract.contractId) string
	// - bytecode: ce.code []byte
	// - function name: ce.fname string
	// - args: ce.ci.Args []interface{}
	// - gas: ce.contractGasLimit uint64
	// - sender: types.EncodeAddress(ce.ctx.curContract.sender) string
	// - hasParent: ce.ctx.callDepth > 1
	// - isFeeDelegation: ce.ctx.isFeeDelegation bool

	address := types.EncodeAddress(ce.ctx.curContract.contractId)
	bytecode := string(ce.code)
	fname := ce.fname
	if ce.isAutoload == true {
		fname = "autoload:" + ce.fname
	}
	// convert the parameters to strings
	args, err := ce.convertArgsToJSON()
	if err != nil {
		ce.err = err
		return
	}
	//gas := strconv.FormatUint(ce.contractGasLimit, 10)
	gas := string((*[8]byte)(unsafe.Pointer(&ce.contractGasLimit))[:])
	sender := types.EncodeAddress(ce.ctx.curContract.sender)
	hasParent := strconv.FormatBool(ce.ctx.callDepth > 1)
	isFeeDelegation := strconv.FormatBool(ce.ctx.isFeeDelegation)

	// build the message
	message := msg.SerializeMessage("execute", address, bytecode, fname, args, gas, sender, hasParent, isFeeDelegation)

	// send the execution request to the VM instance
	err = ce.SendMessage(message)
	if err != nil {
		ce.err = err
		return
	}

	// if this is the first call, wait messages in a loop
	//if ce.ctx.callDepth == 1 {
	//	return MessageLoop()
	//}

	// wait for and process messages in a loop
	result, err := ce.MessageLoop()

	if extractUsedGas && len(result) >= 8 {
		// extract the used gas from the result
		ce.usedGas = binary.LittleEndian.Uint64([]byte(result[:8]))
		result = result[8:]
	}

	// return the result from the VM instance
	ce.jsonRet = result
	ce.err = err

	// when a message arrives, process it
	// when the first VM finishes (or timeout occurs) return from this function

}

// only messages from the last/top contract can be processed
// (a hacker could send as if it is from another contract)
// also use encryption with diff key for each instance

// incoming messages processed:
// 1. only from the last/top contract
// 2. only for 'vm_callback' functions, and responses


func (ce *executor) MessageLoop() (result string, err error) {

	// wait for messages in a loop
	for {
		message, err := ce.WaitForMessage()
		if err != nil {
			return "", err
		}
		// deserialize the message
		args, err := msg.DeserializeMessage(message)
		if err != nil {
			return "", err
		}
		// extract the command, arguments and whether it is within a view function
		if len(args) < 2 {
			return "", errors.New("[MessageLoop] invalid arguments from VM")
		}
		command := args[0]
		inView := args[len(args)-1] == "1"
		args = args[1:len(args)-1]
		// process the request
		if inView {
			ce.ctx.nestedView++
		}
		result, err = ce.ProcessCommand(command, args)
		if inView {
			ce.ctx.nestedView--
		}
		// if the VM finished, return the result
		if command == "return" {
			return result, err
		}
		// serialize the response
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		response := msg.SerializeMessage(result, errMsg)
		// send the response
		err = ce.SendMessage(response)
		if err != nil {
			return "", err  // different type of error
		}
	}

}

// sends a message to the VM instance
func (ce *executor) SendMessage(message []byte) (err error) {
	return msg.SendMessage(ce.vmInstance.conn, message)
}

// waits for a message from the VM instance
func (ce *executor) WaitForMessage() ([]byte, error) {

	if ce.ctx.callDepth == 1 && ce.ctx.deadline.IsZero() {
		// define a global deadline for contract execution
		ce.ctx.deadline = time.Now().Add(250 * time.Millisecond)
	}

	return msg.WaitForMessage(ce.vmInstance.conn, ce.ctx.deadline)
}

// process the command from the VM instance
func (ce *executor) ProcessCommand(command string, args []string) (result string, err error) {

	ctx := ce.ctx

	switch command {

	// return from call

	case "return":
		return ce.handleReturnFromCall(args)

	// state variables

	case "set":
		return ctx.handleSetVariable(args)
	case "get":
		return ctx.handleGetVariable(args)
	case "del":
		return ctx.handleDelVariable(args)

	// contract

	case "deploy":
		return ctx.handleDeploy(args)
	case "call":
		return ctx.handleCall(args)
	case "delegate-call":
		return ctx.handleDelegateCall(args)
	case "send":
		return ctx.handleSend(args)
	case "balance":
		return ctx.handleGetBalance(args)
	case "event":
		return ctx.handleEvent(args)

	// system

	case "toPubkey":
		return ctx.handleToPubkey(args)
	case "toAddress":
		return ctx.handleToAddress(args)
	case "isContract":
		return ctx.handleIsContract(args)
	case "getContractId":
		return ctx.handleGetContractId()
	case "getAmount":
		return ctx.handleGetAmount()
	case "getBlockNo":
		return ctx.handleGetBlockNo()
	case "getTimeStamp":
		return ctx.handleGetTimeStamp()
	case "getPrevBlockHash":
		return ctx.handleGetPrevBlockHash()
	case "getTxHash":
		return ctx.handleGetTxHash()
	case "getOrigin":
		return ctx.handleGetOrigin()
	case "randomInt":
		return ctx.handleRandomInt(args)
	case "print":
		return ctx.handlePrint(args)

	// name service

	case "nameResolve":
		return ctx.handleNameResolve(args)

	// governance

	case "governance":
		return ctx.handleGovernance(args)
	case "getStaking":
		return ctx.handleGetStaking(args)

	// crypto

	case "sha256":
		return ctx.handleCryptoSha256(args)
	case "keccak256":
		return ctx.handleCryptoKeccak256(args)
	case "ecVerify":
		return ctx.handleECVerify(args)
	case "verifyEthStorageProof":
		return ctx.handleCryptoVerifyEthStorageProof(args)

	// db

	case "dbExec":
		return ctx.handleDbExec(args)
	case "dbQuery":
		return ctx.handleDbQuery(args)
	case "dbPrepare":
		return ctx.handleDbPrepare(args)
	case "stmtExec":
		return ctx.handleStmtExec(args)
	case "stmtQuery":
		return ctx.handleStmtQuery(args)
	case "stmtColumnInfo":
		return ctx.handleStmtColumnInfo(args)
	case "rsNext":
		return ctx.handleRsNext(args)
	case "rsGet":
		return ctx.handleRsGet(args)
	//case "rsClose":
	//	return ctx.handleRsClose(args)
	case "lastInsertRowid":
		return ctx.handleLastInsertRowid(args)
	case "dbOpenWithSnapshot":
		return ctx.handleDbOpenWithSnapshot(args)
	case "dbGetSnapshot":
		return ctx.handleDbGetSnapshot(args)

	// internal

	case "setRecoveryPoint":
		return ctx.handleSetRecoveryPoint()
	case "clearRecovery":
		return ctx.handleClearRecovery(args)

	}

	return "", errors.New("invalid command: " + command)

}

// handle the return from a call
func (ce *executor) handleReturnFromCall(args []string) (result string, err error) {

	if len(args) != 2 {
		return "", errors.New("[ReturnFromVM] invalid return value from contract")
	}
	result = args[0]   // JSON
	errStr := args[1]  // error message

	/*
	// add the used gas and check if the execution ran out of gas
	err = ce.processUsedGas(result)

	if errStr != "" {
		if err != nil {
			err = errors.New("[ReturnFromVM] 1: " + err.Error() + ", 2: " + errStr)
		} else {
			err = errors.New(errStr)
		}
	}
	*/

	if errStr != "" {
		err = errors.New(errStr)
	}

	return result, err
}

/*
// add the used gas and check if the execution ran out of gas
func (ce *executor) processUsedGas(result string) (err error) {

	// check if the used gas is a valid uint64 value
	if len(result) < 8 {
		return errors.New("[ReturnFromVM] invalid used gas value from contract")
	}
	// convert the used gas to a uint64 value
	usedGas := binary.LittleEndian.Uint64([]byte(result[:8]))

	// add the gas used by this contract to the total gas
	ce.ctx.accumulatedGas += usedGas

	// check if the contract ran out of the transaction gas limit
	if ce.ctx.accumulatedGas >= ce.ctx.gasLimit {
		return errors.New("[ReturnFromVM] contract ran out of the transaction gas limit")
	}

	// check if the contract ran out of the contract gas limit
	if usedGas >= ce.contractGasLimit {
		return errors.New("[ReturnFromVM] contract ran out of the contract gas limit")
	}

	return nil
}
*/




// sent when a VM is created:
// hardfork version + IsPublic + abstract domain socket name + secret key

// sent when a contract is called:
// sender + IsFeeDelegation


// in the case of timeout:
// the VM pool should:
// - close all connections to the used VMs
// - kill all the processes linked with the current execution
