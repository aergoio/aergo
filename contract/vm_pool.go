package contract

import (
	"sync"
	"net"
	"fmt"
	"strconv"
	"math/rand"
	"os"
	"os/exec"
	"time"
	"path/filepath"

	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/contract/msg"
)

var maxInstances int
var getCh chan *VmInstance
var freeCh chan *VmInstance
var once sync.Once
var VmPoolStarted bool

func StartVMPool(numInstances int) {
	once.Do(func() {
		maxInstances = numInstances
		// create channels for getting and freeing vm instances
		getCh = make(chan *VmInstance, numInstances)
		freeCh = make(chan *VmInstance, numInstances)
		// start a goroutine to manage the vm instances
		go vmPoolRoutine()
		VmPoolStarted = true
	})
}

func vmPoolRoutine() {

	// create vm instances
	spawnVmInstances(maxInstances)

	// wait for instances to be released
	for {
		select {
		case vmInstance := <-freeCh:
			// close the vm instance
			vmInstance.close()
			// replenish the pool
			repopulatePool()
		}
	}

}

//--------------------------------------------------------------------//
// exported functions

func GetVmInstance() *VmInstance {
	vmInstance := <-getCh
	ctrLgr.Trace().Msg("VmInstance acquired")
	return vmInstance
}

func FreeVmInstance(vmInstance *VmInstance) {
	if vmInstance != nil {
		freeCh <- vmInstance
		ctrLgr.Trace().Msg("VmInstance released")
	}
}

// flush and renew all vm instances
func FlushVmInstances() {
	// first retrieve all vm instances, so when releasing the first one
	// the pool is empty and then it will spawn many at once
	list := []*VmInstance{}
	num := len(getCh)
	for i := 0; i < num; i++ {
		vmInstance := GetVmInstance()
		list = append(list, vmInstance)
	}
	for _, vmInstance := range list {
		FreeVmInstance(vmInstance)
	}
}

//--------------------------------------------------------------------//
// VmInstance type

type VmInstance struct {
	id         uint64
	socketName string
	secretKey  [32]byte
	listener   *net.UnixListener
	conn       *net.UnixConn
	pid        int
	used       bool
}

// pool of vm instances
var pool []*VmInstance

// repopulate the pool with new vm instances
func repopulatePool() {

	for {
		// check how many instances are available on the getCh
		numAvailable := len(getCh)
		// if there are less than maxInstances, create new ones
		if numAvailable < maxInstances {
			spawnVmInstances(maxInstances - numAvailable)
		} else {
			break
		}
	}

}

// spawn a number of vm instances
func spawnVmInstances(num int) {
	var num_repeats int

	for i := 0; i < num; i++ {
		// get a random id
		var id uint64
	outer:
		for {
			id = rand.Uint64()
			// check if it is already used
			for _, vmInstance := range pool {
				if vmInstance.id == id {
					continue outer
				}
			}
			break
		}

		// get a random secret key
		secretKey := [32]byte{}
		rand.Read(secretKey[:])

		// get a random name for the abstract unix domain socket
		socketName := fmt.Sprintf("aergo-vm-%x", id)

		// create an abstract unix domain socket listener
		listener, err := net.Listen("unix", "\x00"+socketName)
		if err != nil {
			ctrLgr.Error().Msg("Failed to create unix domain socket listener")
			// try again
			num_repeats++
			if num_repeats > 10 {
				os.Exit(1)
			}
			i--
			continue
		}
		unixListener, ok := listener.(*net.UnixListener)
		if !ok {
			ctrLgr.Error().Msg("Failed to assign listener to *net.UnixListener")
			listener.Close()
			// try again
			num_repeats++
			if num_repeats > 10 {
				os.Exit(1)
			}
			i--
			continue
		}

		// get the directory path of the current executable
		var execDir string
		execPath, err := os.Executable()
		if err != nil {
			ctrLgr.Error().Err(err).Msg("Failed to get executable path")
		} else {
			execDir = filepath.Dir(execPath)
		}

		// try different paths for the external VM executable
		execPath = os.Getenv("AERGOVM_PATH")
		if execPath == "" {
			execPath = filepath.Join(execDir, "aergovm")
			// check if the file exists
			if _, err := os.Stat(execPath); err != nil {
				execPath = "./aergovm"
				if _, err := os.Stat(execPath); err != nil {
					execPath = "aergovm"
				}
			}
		}

		// spawn the exernal VM executable process
		cmd := exec.Command(execPath, strconv.Itoa(int(CurrentForkVersion)), map[bool]string{true: "1", false: "0"}[PubNet], socketName, hex.Encode(secretKey[:]))
		err = cmd.Start()
		if err != nil {
			ctrLgr.Error().Err(err).Msg("Failed to spawn external VM process")
			listener.Close()
			// try again
			num_repeats++
			if num_repeats > 10 {
				os.Exit(1)
			}
			i--
			continue
		}
		// get the process id
		pid := cmd.Process.Pid
		ctrLgr.Trace().Msgf("Spawned external VM process with pid: %d", pid)

		// create a vm instance object
		vmInstance := &VmInstance{
			id:         id,
			socketName: socketName,
			secretKey:  secretKey,
			listener:   unixListener,
			conn:       nil,
			pid:        pid,
			used:       false,
		}

		// add the vm instance to the pool
		pool = append(pool, vmInstance)

	}

	// keep track of the instances that should be removed
	instancesToRemove := []*VmInstance{}

	// keep track of the new instances that are connected
	instancesToRead := []*VmInstance{}

	// the timeout is 100 milliseconds for each vm instance
	timeout := time.Millisecond * time.Duration(100 * num)
	if timeout < time.Second {
		timeout = time.Second
	}
	// set a common deadline for the accept calls
	deadline := time.Now().Add(timeout)

	// iterate over all instances
	for _, vmInstance := range pool {
		// if this VM instance is not yet connected
		if vmInstance.conn == nil {
			// set a deadline for the accept call
			vmInstance.listener.SetDeadline(deadline)
			// wait for the incoming connection
			var err error
			vmInstance.conn, err = vmInstance.listener.AcceptUnix()
			if err == nil {
				// connection accepted
				instancesToRead = append(instancesToRead, vmInstance)
			} else {
				ctrLgr.Error().Msgf("Failed to accept incoming connection: %v", err)
				instancesToRemove = append(instancesToRemove, vmInstance)
			}
		}
	}

	// remove the instances that are not connected
	for _, vmInstance := range instancesToRemove {
		vmInstance.close()
	}

	// iterate over the instances that are connected
	for _, vmInstance := range instancesToRead {
		// wait for a message from the vm instance
		message, err := msg.WaitForMessage(vmInstance.conn, deadline)
		if err != nil {
			ctrLgr.Error().Msgf("Failed to read incoming message: %v", err)
			vmInstance.close()
			continue
		}
		// check if the data is valid
		if !isValidMessage(vmInstance, message) {
			ctrLgr.Error().Msg("Invalid message received")
			vmInstance.close()
			continue
		}
		// send the instance to the getCh
		getCh <- vmInstance
	}

}

func isValidMessage(vmInstance *VmInstance, message []byte) bool {
	if string(message) == "ready" {
		return true
	}
	return false
}

// this should ONLY be called by the vmPoolRoutine. use FreeVmInstance() to release a vm instance
func (vmInstance *VmInstance) close() {
	if vmInstance != nil {
		// close the connections
		vmInstance.listener.Close()
		vmInstance.conn.Close()
		// terminate the process
		process, err := os.FindProcess(vmInstance.pid)
		if err == nil {
			process.Kill()
		}
		// remove the vm instance from the pool
		for i, v := range pool {
			if v == vmInstance {
				pool = append(pool[:i], pool[i+1:]...)
				break
			}
		}
	}
}
