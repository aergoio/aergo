package exec

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
)

// ExecutePack reads a contract from inputFile and packs it, optionally writing to outputFile
func ExecutePack(inputFile, outputFile string) int {

	// Read the contract from the input file
	packedCode, err := util.ReadContract(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading contract: %s\n", err)
		return 1
	}

	// If output file is not specified, print an error
	if outputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: output file is required\n")
		return 1
	}

	// Write the packed code to the output file
	err = ioutil.WriteFile(outputFile, []byte(packedCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to output file: %s\n", err)
		return 1
	}

	fmt.Printf("Packed contract written to %s\n", outputFile)
	return 0
}
