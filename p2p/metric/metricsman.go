/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

import (
	"bytes"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/libp2p/go-libp2p-peer"
	"sync"
	"time"
)

type MetricsManager interface {
	Start()
	Stop()

	Add(pid peer.ID, reader *MetricReader, writer *MetricWriter) *PeerMetric
	Remove(pid peer.ID) *PeerMetric

	Metric(pid peer.ID) (*PeerMetric, bool)
	Metrics() []*PeerMetric

	Summary() string
}

type metricsManager struct {
	logger *log.Logger

	metricsMap map[peer.ID]*PeerMetric

	interval int
	ticker *time.Ticker
	mutex sync.RWMutex
}

func NewMetricManager(interval int) *metricsManager {
	mm := &metricsManager{logger:log.NewLogger("p2p"), metricsMap:make(map[peer.ID]*PeerMetric), interval:interval}

	return mm
}

func (mm *metricsManager) Start() {
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
}

func (mm *metricsManager) Stop() {
	mm.logger.Info().Msg("Finishing p2p metrics manager")
	mm.ticker.Stop()
}

func (mm *metricsManager) Add(pid peer.ID, reader *MetricReader, writer *MetricWriter) *PeerMetric {
	mm.mutex.Lock()
	defer  mm.mutex.Unlock()
	if _, found := mm.metricsMap[pid] ; found {
		mm.logger.Warn().Str("peer_id", pid.Pretty()).Msg("metric for to add peer is already exist. replacing new metric")
	}
	peerMetric := &PeerMetric{PeerID:pid, InMetric:NewExponentMetric5(mm.interval), OutMetric:NewExponentMetric5(mm.interval), Since:time.Now()}
	reader.AddListener(peerMetric.InMetric.AddBytes)
	reader.AddListener(peerMetric.InputAdded)
	writer.AddListener(peerMetric.OutMetric.AddBytes)
	writer.AddListener(peerMetric.OutputAdded)
	mm.metricsMap[pid] = peerMetric
	return peerMetric
}

func (mm *metricsManager) Remove(pid peer.ID) *PeerMetric {
	mm.mutex.Lock()
	defer  mm.mutex.Unlock()
	if metric, found := mm.metricsMap[pid] ; !found {
		mm.logger.Warn().Str("peer_id", pid.Pretty()).Msg("metric for to remove peer is not exist.")
		return nil
	} else {
		delete(mm.metricsMap,pid)
		return metric
	}
}


func (mm *metricsManager) Metric(pid peer.ID) (*PeerMetric, bool) {
	mm.mutex.RLock()
	defer  mm.mutex.RUnlock()

	pm, found := mm.metricsMap[pid]
	return pm, found
}

func (mm *metricsManager) Metrics() []*PeerMetric {
	mm.mutex.RLock()
	defer  mm.mutex.RUnlock()
	view := make([]*PeerMetric, 0, len(mm.metricsMap))
	for _, pm := range mm.metricsMap {
		view = append(view, pm)
	}
	return view
}

func (mm *metricsManager) Summary() string {
	sb := bytes.Buffer{}
	sb.WriteString("p2p metric summary \n")
	if len(mm.Metrics()) > 0 {
		sb.WriteString("PeerID                                                   :  IN_TOTAL,    IN_AVR,   IN_LOAD  :   OUT_TOTAL,   OUT_AVR,  OUT_LOAD\n")
		for _, met := range mm.Metrics() {
			sb.WriteString(met.PeerID.Pretty())
			sb.WriteString(fmt.Sprintf("  :  %10d,%10d,%10d", met.totalIn, met.InMetric.APS(), met.InMetric.LoadScore()))
			sb.WriteString(fmt.Sprintf("  :  %10d,%10d,%10d", met.totalOut, met.OutMetric.APS(), met.OutMetric.LoadScore()))
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
