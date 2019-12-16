package main

import (
	"fmt"
	"net"

	"github.com/aergoio/aergo/config"
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
	r.GET("/debug/voting-power/rankers", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"cmd": "rankers",
		})
	})

	r.GET("/debug/voting-power/rankers/:topn", func(c *gin.Context) {
		topN := c.Params.ByName("topn")
		c.JSON(200, gin.H{
			"cmd":  topN,
			"topn": topN,
		})
	})

	if err := r.Run(hostPort(cfg.DumpPort)); err != nil {
		svrlog.Fatal().Err(err).Msg("failed to start dumper")
	}
}
