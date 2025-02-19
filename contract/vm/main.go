package main

import (
	"fmt"
	"strconv"
	"encoding/binary"
	"os"
	"net"
	"time"
	"github.com/aergoio/aergo/v2/contract/msg"
)

/*
#include "vm.h"
#include "_cgo_export.h"
*/

// global variables

var hardforkVersion int
var isPubNet bool

var secretKey string
var conn *net.UnixConn
var timedout bool


func main(){
	var socketName string
	var err error

	args := os.Args

	// check if args are empty
	if len(args) != 5 {
		fmt.Println("Usage: aergovm <hardforkVersion> <isPubNet> <socketName> <secretKey>")
		return
	}

	// get the hardfork version from command line
	hardforkVersion, err = strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Error: Invalid hardfork version")
		return
	}

	// get PubNet from command line
	isPubNet, err = strconv.ParseBool(args[2])
	if err != nil {
		fmt.Println("Error: Invalid PubNet")
		return
	}

	// get socket name from command line
	socketName = args[3]
	if socketName == "" {
		fmt.Println("Error: Invalid socket name")
		return
	}

	// get secret key from command line
	secretKey = args[4]
	if secretKey == "" {
		fmt.Println("Error: Invalid secret key")
		return
	}

	// initialize Lua modules
	InitializeVM()

	// connect to the server
	err = connectToServer(socketName)
	if err != nil {
		fmt.Println("Error: Could not connect to server")
		return
	}

	// send ready message
	sendReadyMessage()

	// wait for commands from the server
	MessageLoop()

	// exit
	closeApp(0)
}

// connect to the server using an abstract unix domain socket (they start with a null byte)
func connectToServer(socketName string) (err error) {
	rawConn, err := net.Dial("unix", "\x00"+socketName)
	if err != nil {
		return err
	}
	conn = rawConn.(*net.UnixConn)
	return nil
}

func sendReadyMessage() {
	message := []byte("ready")
	/*/ encrypt the message
	message, err = msg.Encrypt(message, secretKey)
	if err != nil {
		fmt.Printf("Error: failed to encrypt message: %v\n", err)
		return
	} */
	msg.SendMessage(conn, message)
}

func MessageLoop() {

	for {
		// wait for command to execute, with null deadline
		message, err := msg.WaitForMessage(conn, time.Time{})
		if err != nil {
			fmt.Printf("Error: failed to receive message: %v\n", err)
			return
		}
		/*/ decrypt the message
		message, err = msg.Decrypt(message, secretKey)
		if err != nil {
			fmt.Printf("Error: failed to decrypt message: %v\n", err)
			return
		} */
		// deserialize the message
		args, err := msg.DeserializeMessage(message)
		if err != nil {
			fmt.Printf("Error: failed to deserialize message: %v\n", err)
			return
		}
		command := args[0]
		args = args[1:]
		fmt.Println("Received message: ", command, args)
		// process the request
		result, err := processCommand(command, args)
		//if err != nil {
		//	return "", err
		//}
		// serialize the result and error
		response := msg.SerializeMessage(result, err.Error())
		/*/ encrypt the response
		response, err = msg.Encrypt(response, secretKey)
		if err != nil {
			fmt.Printf("Error: failed to encrypt message: %v\n", err)
			return
		} */
		// send the response
		err = msg.SendMessage(conn, response)
		if err != nil {
			fmt.Printf("Error: failed to send message: %v\n", err)
			return
		}
	}

}

func processCommand(command string, args []string) (string, error) {

	switch command {
	case "execute":
		if len(args) != 10 {
			fmt.Println("execute: invalid number of arguments")
			sendMessage([]string{"", "execute: invalid number of arguments"})
			closeApp(1)
		}
		address := args[0]
		code := args[1]
		fname := args[2]
		fargs := args[3]
		useGas, err := strconv.ParseBool(args[4])
		if err != nil {
			fmt.Println("execute: invalid useGas argument")
			sendMessage([]string{"", "execute: invalid useGas argument"})
			closeApp(1)
		}
		limitStr := args[5]
		caller := args[6]
		hasParent, err := strconv.ParseBool(args[7])
		if err != nil {
			fmt.Println("execute: invalid hasParent argument")
			sendMessage([]string{"", "execute: invalid hasParent argument"})
			closeApp(1)
		}
		isFeeDelegation, err := strconv.ParseBool(args[8])
		if err != nil {
			fmt.Println("execute: invalid isFeeDelegation argument")
			sendMessage([]string{"", "execute: invalid isFeeDelegation argument"})
			closeApp(1)
		}
		abiError := args[9]

		var limit uint64
		limitBytes := []byte(limitStr)
		if len(limitBytes) != 8 {
			fmt.Println("execute: invalid limit string length")
			sendMessage([]string{"", "execute: invalid limit string length"})
			closeApp(1)
		}
		limit = binary.LittleEndian.Uint64(limitBytes)

		if useGas {
			contractUseGas = true
			contractGasLimit = limit
		} else {
			contractUseGas = false
			contractInstructionLimit = limit
		}

		res, err, usedResources := Execute(address, code, fname, fargs, caller, hasParent, isFeeDelegation, abiError)

		// encode the used resources together with the result
		resourcesBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(resourcesBytes, usedResources)
		res = string(resourcesBytes) + res

		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		err = sendApiMessage("return", []string{res, errStr})
		if err != nil {
			fmt.Printf("execute: failed to send message: %v\n", err)
			closeApp(1)
		}
		closeApp(0)

	case "compile":
		if len(args) != 2 {
			fmt.Println("compile: invalid number of arguments")
			sendMessage([]string{"", "compile: invalid number of arguments"})
			closeApp(1)
		}
		code := args[0]
		hasParent, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Println("compile: invalid hasParent argument")
			sendMessage([]string{"", "compile: invalid hasParent argument"})
			closeApp(1)
		}

		bytecodeAbi, err := Compile(code, hasParent)

		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		sendMessage([]string{string(bytecodeAbi), errStr})
		closeApp(0)

	// if the contract is executing, this can only be received if using another thread
	// or if checking for incoming messages in regular intervals (expensive operation)
	case "timeout":
		timedout = true
		return "", nil

	case "exit":
		closeApp(0)

	}

	fmt.Println("aergovm: unknown command: ", command)
	sendMessage([]string{"", "aergovm: unknown command: " + command})
	closeApp(1)
	return "", nil
}

func closeApp(ret int) {

	if conn != nil {
		// ensure the data is sent before closing the connection
		err := conn.CloseWrite()
		if err != nil {
			fmt.Printf("aergovm: failed to close write end of connection: %v\n", err)
			os.Exit(1)
		}
		// close the connection
		conn.Close()
	}

	os.Exit(ret)
}
