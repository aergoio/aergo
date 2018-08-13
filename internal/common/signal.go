package common

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/aergoio/aergo/pkg/log"
)

// HandleKillSig gets killing signals (interrupt, quit and terminate) and calls
// a registered handler function for cleanup. Finally, this will exit program
func HandleKillSig(handler func(), logger *log.Logger) {
	sigChannel := make(chan os.Signal, 1)

	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for signal := range sigChannel {
			logger.Info().Msgf("Receive signal %s, Shutting down...", signal)
			handler()
			os.Exit(1)
		}
	}()
}
