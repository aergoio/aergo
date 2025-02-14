package main

import (
	"encoding/binary"
	"bytes"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
	"github.com/aergoio/aergo/v2/contract/msg"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

var contractCode = `
state.var {
	kv = state.map()
}
function add(a, b)
	return a + b
end
function set(key, value)
	kv[key] = value
end
function get(key)
	return kv[key]
end
function send(to, amount)
	return contract.send(to, amount)
end
function call(...)
	return contract.call(...)
end
function call_with_send(amount, ...)
	return contract.call.value(amount)(...)
end
function delegatecall(...)
	return contract.delegatecall(...)
end
function deploy(...)
	return contract.deploy(...)
end
function deploy_with_send(amount, ...)
	return contract.deploy.value(amount)(...)
end
function get_info()
	return system.getContractID(), contract.balance(), system.getAmount(), system.getSender(), system.getOrigin(), system.isFeeDelegation()
end
function events()
	contract.event('first', 123, 'abc')
	contract.event('second', '456', 7.89)
end
abi.register(add, set, get, send, call, call_with_send, delegatecall, deploy, deploy_with_send, get_info, events)
`

var initDone = false

func compileVmExecutable(t *testing.T) {

	// Get the root folder based on the current file's directory
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "../..")

	// Change the current working directory to the root directory
	err := os.Chdir(rootDir)
	require.NoError(t, err, "Failed to change working directory")

	// Check if the VM binary exists
	if _, err := os.Stat("bin/aergovm"); os.IsNotExist(err) {
		t.Log("VM binary does not exist, compiling it")
		cmd := exec.Command("make", "aergovm")
		err := cmd.Run()
		require.NoError(t, err, "Failed to compile the VM executable")
	}

}

func NewVmInstance(t *testing.T) (*exec.Cmd, net.Conn, chan struct{}) {

	if !initDone {
		compileVmExecutable(t)
		initDone = true
	}

	// Set up the Unix domain socket
	t.Log("Creating Unix domain socket")
	socketName := "\x00test_socket"
	rawListener, err := net.Listen("unix", socketName)
	require.NoError(t, err, "Failed to create Unix domain socket")
	defer rawListener.Close()
	listener, ok := rawListener.(*net.UnixListener)
	require.True(t, ok, "Failed to assign listener to *net.UnixListener")

	// Start the VM process
	t.Log("Starting VM process")
	vmCmd := exec.Command("bin/aergovm", "3", "false", socketName[1:], "test_secret_key")

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	vmCmd.Stdout = &stdout
	vmCmd.Stderr = &stderr
	err = vmCmd.Start()
	require.NoError(t, err, "Failed to start VM process")

	done := make(chan struct{})

	go func() {
		err := vmCmd.Wait()
		select {
		case <-done:
			// Test completed successfully, do nothing
		default:
			// VM process exited before test completion
			if err != nil {
				t.Errorf("VM process exited unexpectedly: %v", err)
				t.Logf("Stderr: %s", stderr.String())
				t.Logf("Stdout: %s", stdout.String())
				t.FailNow()
			}
		}
	}()

	// Wait for and accept the connection from VM with a timeout
	t.Log("Waiting for connection from VM")
	err = listener.SetDeadline(time.Now().Add(3 * time.Second))
	require.NoError(t, err, "Failed to set listener deadline")
	conn, err := listener.AcceptUnix()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Fatal("Timed out waiting for VM to connect")
		}
		require.NoError(t, err, "Failed to accept connection from VM")
	}

	t.Log("Connection accepted from VM")

	// Reset the deadline
	err = listener.SetDeadline(time.Time{})
	require.NoError(t, err, "Failed to reset listener deadline")

	t.Log("Waiting for ready message")
	readyMsg, err := msg.WaitForMessage(conn, time.Now().Add(2*time.Second))
	require.NoError(t, err, "Failed to receive ready message")
	require.Equal(t, "ready", string(readyMsg), "Unexpected ready message")

	return vmCmd, conn, done
}

func TestVMExecutionBasicPlainCode(t *testing.T) {

	vmCmd, conn, done := NewVmInstance(t)

	// Test the execute command
	t.Log("Sending execute command")
	executeCmd := []string{"execute", "contract_address", contractCode, "add", `[123,456]`, "\x00\x00\x00\x01\x00\x00\x00\x00", "test_caller", "false", "false", ""}
	serializedCmd := msg.SerializeMessage(executeCmd...)
	err := msg.SendMessage(conn, serializedCmd)
	require.NoError(t, err, "Failed to send execute command")

	t.Log("Waiting for execute response")
	response, err := msg.WaitForMessage(conn, time.Now().Add(250*time.Millisecond))
	require.NoError(t, err, "Failed to receive execute response")

	// Deserialize the response
	args, err := msg.DeserializeMessage(response)
	require.NoError(t, err)
	require.Len(t, args, 4, "Unexpected number of response arguments")
	command := args[0]
	result := args[1]
	errStr := args[2]
	inView := args[3]

	// Extract used gas and result
	require.Greater(t, len(result), 8, "expected to contain encoded gas")
	usedGas := binary.LittleEndian.Uint64([]byte(result[:8]))
	result = result[8:]

	assert.Equal(t, "return", command)
	assert.Equal(t, "579", result)
	assert.Equal(t, "", errStr)
	assert.Equal(t, uint64(3648), usedGas)
	assert.Equal(t, "0", inView)

	vmCmd.Process.Kill()
	conn.Close()
	// Signal that the test is done
	done <- struct{}{}
	close(done)
}

func TestVMCompileAndExecutionBasic(t *testing.T) {

	vmCmd, conn, done := NewVmInstance(t)

	// Compile the contract
	t.Log("Sending compile command")
	compileCmd := []string{"compile", contractCode, "false"}
	serializedCmd := msg.SerializeMessage(compileCmd...)
	err := msg.SendMessage(conn, serializedCmd)
	require.NoError(t, err, "Failed to send compile command")

	t.Log("Waiting for compile response")
	response, err := msg.WaitForMessage(conn, time.Now().Add(250*time.Millisecond))
	require.NoError(t, err, "Failed to receive compile response")

	args, err := msg.DeserializeMessage(response)
	require.NoError(t, err)
	require.Len(t, args, 2, "Unexpected number of response arguments")
	bytecodeAbi := args[0]
	errMsg := args[1]
	require.Equal(t, "", errMsg)

	vmCmd.Process.Kill()
	conn.Close()
	// Signal that the test is done
	done <- struct{}{}
	close(done)


	bytecode := util.LuaCode(bytecodeAbi).ByteCode()


	vmCmd, conn, done = NewVmInstance(t)

	// Test the execute command
	t.Log("Sending execute command")
	executeCmd := []string{"execute", "contract_address", string(bytecode), "add", `[123,456]`, "\x00\x00\x00\x01\x00\x00\x00\x00", "test_caller", "false", "false", ""}
	serializedCmd = msg.SerializeMessage(executeCmd...)
	err = msg.SendMessage(conn, serializedCmd)
	require.NoError(t, err, "Failed to send execute command")

	t.Log("Waiting for execute response")
	response, err = msg.WaitForMessage(conn, time.Now().Add(250*time.Millisecond))
	require.NoError(t, err, "Failed to receive execute response")

	// Deserialize the response
	args, err = msg.DeserializeMessage(response)
	require.NoError(t, err)
	require.Len(t, args, 4, "Unexpected number of response arguments")
	command := args[0]
	result := args[1]
	errStr := args[2]
	inView := args[3]

	// Extract used gas and result
	require.GreaterOrEqual(t, len(result), 8, "expected to contain encoded gas")
	usedGas := binary.LittleEndian.Uint64([]byte(result[:8]))
	result = result[8:]

	assert.Equal(t, "return", command)
	assert.Equal(t, "579", result)
	assert.Equal(t, "", errStr)
	assert.Equal(t, uint64(5856), usedGas)
	assert.Equal(t, "0", inView)

	vmCmd.Process.Kill()
	conn.Close()
	// Signal that the test is done
	done <- struct{}{}
	close(done)
}

func TestVMExecutionWithCallback(t *testing.T) {

	vmCmd, conn, done := NewVmInstance(t)

	// Test the execute command
	t.Log("Sending execute command")
	executeCmd := []string{"execute", "contract_address", contractCode, "send", `["test_to","9876543210"]`, "\x00\x00\x00\x01\x00\x00\x00\x00", "test_caller", "false", "false", ""}
	serializedCmd := msg.SerializeMessage(executeCmd...)
	err := msg.SendMessage(conn, serializedCmd)
	require.NoError(t, err, "Failed to send execute command")

	t.Log("Waiting for execute response")
	response, err := msg.WaitForMessage(conn, time.Now().Add(250*time.Millisecond))
	require.NoError(t, err, "Failed to receive execute response")

	// Deserialize the response
	args, err := msg.DeserializeMessage(response)
	require.NoError(t, err)
	require.Len(t, args, 5)
	require.Equal(t, "send", args[0])
	require.Equal(t, "test_to", args[1])
	require.Equal(t, "9876543210", args[2])
	require.Equal(t, "\x84\xf0\xff\x00\x00\x00\x00\x00", args[3])
	require.Equal(t, "0", args[4])

	// Send response back to the VM instance
	args = []string{"\x09\x00\x01\x00\x00\x00\x00\x00", ""}
	message := msg.SerializeMessage(args...)
	err = msg.SendMessage(conn, message)
	require.NoError(t, err, "Failed to send response back to VM")

	t.Log("Waiting for execute response")
	response, err = msg.WaitForMessage(conn, time.Now().Add(250*time.Millisecond))
	require.NoError(t, err, "Failed to receive execute response")

	// Deserialize the response
	args, err = msg.DeserializeMessage(response)
	require.NoError(t, err)
	require.Len(t, args, 4, "Unexpected number of response arguments")
	command := args[0]
	result := args[1]
	errStr := args[2]
	inView := args[3]

	// Extract used gas and result
	require.GreaterOrEqual(t, len(result), 8, "expected to contain encoded gas")
	usedGas := binary.LittleEndian.Uint64([]byte(result[:8]))
	result = result[8:]

	assert.Equal(t, "return", command)
	assert.Equal(t, "", result)
	assert.Equal(t, "", errStr)
	assert.Equal(t, uint64(69509), usedGas)
	assert.Equal(t, "0", inView)

	vmCmd.Process.Kill()
	conn.Close()
	// Signal that the test is done
	done <- struct{}{}
	close(done)
}
