package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/consensus/chain"
	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/pkg/component"
)

type dumper struct {
	*component.ComponentHub
	cfg *config.Config
}

// NewDumper returns a new dumer object.
func NewDumper(cfg *config.Config, hub *component.ComponentHub) *dumper {
	return &dumper{
		ComponentHub: hub,
		cfg:          cfg,
	}
}

func (dmp *dumper) Start() {
	go dmp.run()
}

func (dmp *dumper) run() {
	hostPort := func(port int) string {
		// Allow debug dump to access only from the local machine.
		host := "127.0.0.1"
		if port <= 0 {
			port = config.GetDefaultDumpPort()
		}
		return net.JoinHostPort(host, fmt.Sprintf("%d", port))
	}

	///////////////////////////////////////////////////////////////////////////
	// Dump Voting Power Rankers
	///////////////////////////////////////////////////////////////////////////

	// Dump Handler Generator
	dumpFn := func(topN int) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var buf bytes.Buffer

			dumpRankers := func() error {
				chain.Lock()
				defer chain.Unlock()

				return system.DumpVotingPowerRankers(&buf, topN)
			}

			if err := dumpRankers(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write(buf.Bytes())
		}
	}
	// Dump all rankers.
	http.HandleFunc("/debug/voting-power/rankers", func(w http.ResponseWriter, r *http.Request) {
		dumpFn(0)(w, r)
	})

	// Dump the top n rankers.
	http.HandleFunc("/debug/voting-power/rankers/", func(w http.ResponseWriter, r *http.Request) {
		topN := 0
		if n, err := strconv.Atoi(r.URL.Path[len("/debug/voting-power/rankers/"):]); err == nil && n > 0 {
			topN = n
		}
		dumpFn(topN)(w, r)
	})

	// Start the HTTP server
	server := &http.Server{
		Addr: hostPort(dmp.cfg.DumpPort),
	}

	if err := server.ListenAndServe(); err != nil {
		svrlog.Fatal().Err(err).Msg("failed to start dumper")
	}
}
