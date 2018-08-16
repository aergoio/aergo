package main

import (
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompile1(t *testing.T) {
	cmd := exec.Command("../../bin/aergoluac", "good.lua", "good.bc")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finishied with error: %v", err)
	assert.Nil(t, err)
	fi, e := os.Stat("good.bc")
	if e != nil {
		log.Fatal(e)
	}
	assert.Equal(t, int64(273), fi.Size())
}

func TestCompile2(t *testing.T) {
	cmd := exec.Command("../../bin/aergoluac", "bad.lua", "bad.bc")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finishied with error: %v", err)
	assert.NotNil(t, err)
	_, e := os.Stat("bad.bc")
	assert.NotNil(t, e)
}
