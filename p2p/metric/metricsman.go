/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

type MetricsManager interface {
	Start()
	Stop()

	NewMetric(pid types.PeerID, manNum uint32) *PeerMetric
	Remove(pid types.PeerID, manNum uint32) *PeerMetric

	Metric(pid types.PeerID) (*PeerMetric, bool)
	Metrics() []*PeerMetric

	Summary() map[string]interface{}
	PrintMetrics() string
}

//go:generate mockgen -source=metricsman.go -package=p2pmock -destination=../p2pmock/mock_metricsman.go

type metricsManager struct {
	logger    *log.Logger
	startTime time.Time

	metricsMap map[types.PeerID]*PeerMetric

	interval int
	ticker   *time.Ticker
	mutex    sync.RWMutex

	deadTotalIn  int64
	deadTotalOut int64
}

var _ MetricsManager = (*metricsManager)(nil)

func NewMetricManager(interval int) *metricsManager {
	mm := &metricsManager{logger: log.NewLogger("p2p"), metricsMap: make(map[types.PeerID]*PeerMetric), interval: interval, startTime: time.Now()}

	return mm
}

func (mm *metricsManager) Start() {
	go func() {
		mm.logger.Info().Msg("Starting p2p metrics manager ")
		mm.ticker = time.NewTicker(time.Second * time.Duration(mm.interval))
		for range mm.ticker.C {
			mm.mutex.RLock()
			//mm.logger.Debug().Int("peer_cnt", len(mm.metricsMap)).Msg("Calculating peer metrics")
			for _, peerMetric := range mm.metricsMap {
				peerMetric.InMetric.Calculate()
				peerMetric.OutMetric.Calculate()
			}
			mm.mutex.RUnlock()
		}
	}()
}

func (mm *metricsManager) Stop() {
	mm.logger.Info().Msg("Finishing p2p metrics manager")
	mm.ticker.Stop()
}

func (mm *metricsManager) NewMetric(pid types.PeerID, manNum uint32) *PeerMetric {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	if old, found := mm.metricsMap[pid]; found {
		mm.logger.Warn().Str("peer_id", p2putil.ShortForm(pid)).Uint32("oldNum", old.seq).Uint32("newNum", manNum).Msg("metric for to add peer is already exist. replacing new metric")
	}
	peerMetric := &PeerMetric{mm: mm, PeerID: pid, seq: manNum, InMetric: NewExponentMetric5(mm.interval), OutMetric: NewExponentMetric5(mm.interval), Since: time.Now()}
	mm.metricsMap[pid] = peerMetric
	return peerMetric
}

func (mm *metricsManager) Remove(pid types.PeerID, manNum uint32) *PeerMetric {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	if metric, found := mm.metricsMap[pid]; !found {
		mm.logger.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(pid)).Msg("metric for to remove peer is not exist.")
		return nil
	} else {
		if metric.seq != manNum {
			mm.logger.Warn().Uint32("exist_num", metric.seq).Uint32("man_num", manNum).Str(p2putil.LogPeerID, p2putil.ShortForm(pid)).Msg("ignore remove. different manage number")
		}
		atomic.AddInt64(&mm.deadTotalIn, metric.totalIn)
		atomic.AddInt64(&mm.deadTotalOut, metric.totalOut)
		delete(mm.metricsMap, pid)
		return metric
	}
}

func (mm *metricsManager) Metric(pid types.PeerID) (*PeerMetric, bool) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	pm, found := mm.metricsMap[pid]
	return pm, found
}

func (mm *metricsManager) Metrics() []*PeerMetric {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	view := make([]*PeerMetric, 0, len(mm.metricsMap))
	for _, pm := range mm.metricsMap {
		view = append(view, pm)
	}
	return view
}

func (mm *metricsManager) Summary() map[string]interface{} {
	// There can be a little error
	sum := make(map[string]interface{})
	sum["since"] = mm.startTime
	var totalIn, totalOut int64
	if len(mm.Metrics()) > 0 {
		var cnt = 0
		//var inAps, inLoad, outAps, outLoad int64
		for _, met := range mm.Metrics() {
			cnt++
			totalIn += met.totalIn
			totalOut += met.totalOut
		}
	}
	totalIn += atomic.LoadInt64(&mm.deadTotalIn)
	totalOut += atomic.LoadInt64(&mm.deadTotalOut)
	sum["in"] = totalIn
	sum["out"] = totalOut
	return sum
}

func (mm *metricsManager) PrintMetrics() string {
	sb := bytes.Buffer{}
	sb.WriteString("p2p metric summary \n")
	if len(mm.Metrics()) > 0 {
		sb.WriteString("PeerID      :  IN_TOTAL,    IN_AVR,   IN_LOAD  :   OUT_TOTAL,   OUT_AVR,  OUT_LOAD\n")
		for _, met := range mm.Metrics() {
			sb.WriteString(p2putil.ShortForm(met.PeerID))
			sb.WriteString(fmt.Sprintf("  :  %10d,%10d,%10d", met.totalIn, met.InMetric.APS(), met.InMetric.LoadScore()))
			sb.WriteString(fmt.Sprintf("  :  %10d,%10d,%10d", met.totalOut, met.OutMetric.APS(), met.OutMetric.LoadScore()))
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
