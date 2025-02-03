package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"unsafe"
)


// convert the ce.ci.Args to []string
func (ce *executor) convertArgs() ([]string, error) {
	args := make([]string, len(ce.ci.Args))
	for i, arg := range ce.ci.Args {
		args[i] = fmt.Sprintf("%v", arg)
	}
	return args, nil
}

// convert the ce.ci.Args to a single string (JSON array)
func (ce *executor) convertArgsToJSON() (string, error) {
	args, err := ce.convertArgs()
	if err != nil {
		return "", err
	}
	return json.Marshal(args)
}

// convert the ce.ci.Args to a single string (JSON array)
func (ce *executor) convertArgsToJSON() (string, error) {
	args, err := json.Marshal(ce.ci.Args)
	if err != nil {
		return "", err
	}
	return string(args), nil
}


func (ce *executor) call() (result string, err error) {

	if ce.isView == true {
		ce.ctx.nestedView++
		defer func() {
			ce.ctx.nestedView--
		}()
	}

	// what to send:
	// - bytecode: ce.code []byte
	// - function name: ce.fname string
	// - args: ce.ci.Args []interface{}
	// - gas: ce.remainingGas uint64
	// - sender: types.EncodeAddress(ce.ctx.curContract.sender) string
	// - amount: ce.ctx.curContract.amount.String() string
	// - isFeeDelegation: ce.ctx.isFeeDelegation bool

// max uint64: 18446744073709551615

	// convert the parameters to strings
	args, err := ce.convertArgsToJSON()
	if err != nil {
		return "", err
	}
	//gas := strconv.FormatUint(ce.remainingGas, 10)
	gas := string((*[8]byte)(unsafe.Pointer(&ce.remainingGas))[:])
	sender := types.EncodeAddress(ce.ctx.curContract.sender)
	amount := ce.ctx.curContract.amount.String()
	isFeeDelegation := strconv.FormatBool(ce.ctx.isFeeDelegation)

	// build the message
	msg := SerializeMessage("execute", string(ce.code), ce.fname, args, gas, sender, amount, isFeeDelegation)

	// send the execution request to the VM instance
	err := ce.SendMessage(msg)
	if err != nil {
		return "", err
	}

	// if this is the first call, wait messages in a loop
	//if ce.ctx.callDepth == 1 {
	//	return MessageLoop()
	//}

	// wait for and process messages in a loop
	result, err = ce.MessageLoop()
	if err != nil {
		return "", err
	}

	// return the result from the VM instance
	return result, nil

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
		msg, err := ce.WaitForMessage()
		if err != nil {
			return "", err
		}
		// deserialize the message
		args, err := DeserializeMessage(msg)
		if err != nil {
			return "", err
		}
		command := args[0]
		args = args[1:]
		// process the request
		result, err = ce.ProcessCommand(command, args)
		if command == "return" {
			return result, err
		} else if err != nil {
			return "", err
		}
		// send the response
		err = ce.SendMessage(result)
		if err != nil {
			return "", err  // different type of error
		}
	}

}

// sends a message to the VM instance
func (ce *executor) SendMessage(message string) (err error) {
	return msg.SendMessage(ce.vmInstance.conn, message)
}

// waits for a message from the VM instance
func (ce *executor) WaitForMessage() (string, error) {

	if ce.ctx.callDepth == 1 {
		// define a global deadline for contract execution
		ce.ctx.deadline = time.Now().Add(ce.ctx.timeout)
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
		return ctx.handleGetContractId(args)
	case "getBlockNo":
		return ctx.handleGetBlockNo(args)
	case "getTimeStamp":
		return ctx.handleGetTimeStamp(args)
	case "getPrevBlockHash":
		return ctx.handleGetPrevBlockHash(args)
	case "getTxHash":
		return ctx.handleGetTxHash(args)
	case "getOrigin":
		return ctx.handleGetOrigin(args)
	case "random":
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
	case "rsClose":
		return ctx.handleRsClose(args)
	case "lastInsertRowid":
		return ctx.handleLastInsertRowid(args)
	case "dbOpenWithSnapshot":
		return ctx.handleDbOpenWithSnapshot(args)
	case "dbGetSnapshot":
		return ctx.handleDbGetSnapshot(args)

	// internal

	case "setRecoveryPoint":
		return ctx.handleSetRecoveryPoint(args)
	case "clearRecovery":
		return ctx.handleClearRecovery(args)


	default:
		return errors.New("invalid command: " + command)
	}


}

// handle the return from a call
func (ce *executor) handleReturnFromCall(args []string) (result string, err error) {

	if len(args) != 3 {
		return "", errors.New("[ReturnFromVM] invalid return value from contract")
	}
	result = args[0]   // JSON
	errStr = args[1]   // error message
	usedGas = args[2]  // used gas (from just 1 contract)

	// add the used gas and check if the execution ran out of gas
	err = ce.processUsedGas(usedGas)

	if errStr != "" {
		if err != nil {
			err = errors.New("[ReturnFromVM] 1: " + err.Error() + ", 2: " + errStr)
		} else {
			err = errors.New(errStr)
		}
	}

	return result, err
}

// add the used gas and check if the execution ran out of gas
func (ce *executor) processUsedGas(usedGasStr string) (err error) {

	/*
	usedGas, err := strconv.ParseUint(usedGasStr, 10, 64)
	if err != nil {
		return errors.New("[ReturnFromVM] invalid used gas value from contract")
	}
	*/

	// check if the used gas is a valid uint64 value
	if len(usedGasStr) != 8 {
		return errors.New("[ReturnFromVM] invalid used gas value from contract")
	}
	// convert the used gas to a uint64 value
	usedGas := binary.LittleEndian.Uint64([]byte(usedGasStr))

	// add the gas used by this contract to the total gas
	ce.ctx.usedGas += usedGas

	// check if the contract ran out of the transaction gas limit
	if ce.ctx.usedGas >= ce.ctx.gasLimit {
		return errors.New("[ReturnFromVM] contract ran out of the transaction gas limit")
	}

	// check if the contract ran out of the contract gas limit
	if usedGas >= ce.remainingGas {
		return errors.New("[ReturnFromVM] contract ran out of the contract gas limit")
	}

	return nil
}





// sent when a VM is created:
// hardfork version + IsPublic + abstract domain socket name + secret key

// sent when a contract is called:
// sender + amount + timestamp + IsFeeDelegation

luaGetSender(ctx) *C.char
luaGetAmount(ctx) *C.char
luaIsFeeDelegation(ctx) (C.int, *C.char)




// in the case of timeout:
// the VM pool should:
// - close all connections to the used VMs
// - kill all the processes linked with the current execution
