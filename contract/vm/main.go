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
	if len(args) != 4 {
		fmt.Println("Usage: vm <hardforkVersion> <isPubNet> <socketName> <secretKey>")
		return
	}

  // get the hardfork version from command line
	hardforkVersion, err = strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Error: Invalid hardfork version")
		return
	}

  // get PubNet from command line
	isPubNet, err = strconv.ParseBool(args[1])
	if err != nil {
		fmt.Println("Error: Invalid PubNet")
		return
	}

  // get socket name from command line
	socketName = args[2]
	if socketName == "" {
		fmt.Println("Error: Invalid socket name")
		return
	}

  // get secret key from command line
	secretKey = args[3]
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

	// wait for commands from the server
	MessageLoop()

	// close the connection
	conn.Close()
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
		address := args[0]
		code := args[1]
		fname := args[2]
		fargs := args[3]
		gasStr := args[4]
		caller := args[5]
		isFeeDelegation, err := strconv.ParseBool(args[6])
		if err != nil {
			return "", err
		}

		var gas uint64
		gasBytes := []byte(gasStr)
		if len(gasBytes) != 8 {
			return "", fmt.Errorf("invalid gas string length")
		}
		gas = binary.LittleEndian.Uint64(gasBytes)

		res, err := Execute(address, code, fname, fargs, gas, caller, isFeeDelegation)
		return res, err

	case "compile":
		code := args[0]
		hasParent, err := strconv.ParseBool(args[1])
		if err != nil {
			return "", err
		}

		res, err := Compile(code, hasParent)
		return string(res), err

	// if the contract is executing, this can only be received if using another thread
	case "timeout":
		timedout = true
		return "", nil

	case "exit":
		os.Exit(0)

	}

	return "", fmt.Errorf("unknown command: %s", command)
}
