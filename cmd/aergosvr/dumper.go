package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/gin-gonic/gin"
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

	r := gin.Default()

	///////////////////////////////////////////////////////////////////////////
	// Dump Voting Power Rankers
	///////////////////////////////////////////////////////////////////////////

	// Dump Handler Generator
	dumpFn := func(topN int) func(c *gin.Context) {
		return func(c *gin.Context) {
			var buf bytes.Buffer

			if err := system.DumpVotingPowerRankers(&buf, topN); err != nil {
				c.JSON(400, gin.H{
					"message": err.Error(),
				})
				return
			}

			c.Header("Content-Type", "application/json; charset=utf-8")
			c.String(200, string(buf.Bytes()))
		}
	}

	// Dump all rankers.
	r.GET("/debug/voting-power/rankers", dumpFn(0))

	// Dump the top n rankers.
	r.GET("/debug/voting-power/rankers/:topn", func(c *gin.Context) {
		topN := 0
		if n, err := strconv.Atoi(c.Params.ByName("topn")); err == nil && n > 0 {
			topN = n
		}
		dumpFn(topN)(c)
	})

	if err := r.Run(hostPort(cfg.DumpPort)); err != nil {
		svrlog.Fatal().Err(err).Msg("failed to start dumper")
	}
}
