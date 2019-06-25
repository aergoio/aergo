package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateConfChangeID(t *testing.T) {
	hash := GenerateConfChangeIDFromTxID(DecodeB58("AGDyRUQDB7rFuHFG1HngZ2skGyVXyFSC4P79Y5CN67P5"))
	assert.Equal(t, hash, uint64(17098409255004287055))

	hash = GenerateConfChangeIDFromTxID(DecodeB58("BnPTRc8DYQMq8HvW7Vw98Bo8oLNJgxfA4RA59BUY3GhZ"))
	assert.Equal(t, hash, uint64(3014046605232202636))
}
