/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"sync"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/polaris/common"
)

type HealthCheckManager interface {
	Start()
	Stop()
}

type healthCheckManager struct {
	logger *log.Logger
	ms     mapService
	nt     p2pcommon.NetworkTransport
	finish chan interface{}

	workerCnt int
}

var _ HealthCheckManager = (*healthCheckManager)(nil)

func (hcm *healthCheckManager) Start() {
	go hcm.runHCM()
}

func (hcm *healthCheckManager) Stop() {
	hcm.finish <- struct{}{}
}

func NewHCM(mapService *PeerMapService, nt p2pcommon.NetworkTransport) *healthCheckManager {
	hcm := &healthCheckManager{ms: mapService, nt: nt, logger: mapService.Logger, workerCnt: ConcurrentHealthCheckCount,
		finish: make(chan interface{}, 1)}

	return hcm
}

func (hcm *healthCheckManager) runHCM() {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			hcm.checkPeers()

		case <-hcm.finish:
			break
		}
	}

}

func (hcm *healthCheckManager) checkPeers() {
	checkers := hcm.ms.getPeerCheckers()
	thresholdTime := time.Now().Add(PeerHealthcheckInterval)
	toCheck := make([]peerChecker, 0, len(checkers)>>2)
	for _, ps := range checkers {
		if ps.lastCheck().Before(thresholdTime) {
			toCheck = append(toCheck, ps)
		}
	}

	hcm.logger.Debug().Int("all_peers", len(checkers)).Int("check_peers", len(toCheck)).Msg("Starting peers health check")
	wg := &sync.WaitGroup{}
	wg.Add(len(toCheck))
	for _, ps := range toCheck {
		// TODO make a pool and limit count of concurrent pings
		go ps.check(wg, common.PolarisConnectionTTL)
	}
	wg.Wait()
	hcm.logger.Debug().Msg("Finished checks")
}
