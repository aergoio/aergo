/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package esindexer

import (
	"reflect"
	"time"
	"context"
	"errors"
	"sync/atomic"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/olivere/elastic"
	"github.com/aergoio/aergo-lib/log"
	"golang.org/x/sync/errgroup"
)

var (
	logger = log.NewLogger("esindexer")
)

type EsIndexer struct {
	*component.BaseComponent
	conf *config.Config
	hub *component.ComponentHub
	ca types.ChainAccessor

	client *elastic.Client
	indexNamePrefix string
}

type EsBlock struct {
	Timestamp time.Time `json:"ts"`
	BlockNo   uint64    `json:"no"`
}

type EsTx struct {
	id         string
	Timestamp  time.Time `json:"ts"`
	BlockNo    uint64    `json:"blockno"`
	Account    string    `json:"from"`
	Recipient  string    `json:"to"`
}

var mappings = map[string]string{
	"tx": `{
		"mappings":{
			"tx":{
				"properties":{
					"ts": {
						"type": "date"
					},
					"blockno": {
						"type": "long"
					},
					"from": {
						"type": "keyword"
					},
					"to": {
						"type": "keyword"
					}
				}
			}
		}
	}`,
	"block": `{
		"mappings":{
			"block":{
				"properties": {
					"ts": {
						"type": "date"
					},
					"no": {
						"type": "long"
					}
				}
			}
		}
	}`,
}

func NewEsIndexer(hub *component.ComponentHub, cfg *config.Config, chainAccessor types.ChainAccessor) *EsIndexer {
	svc := &EsIndexer{
		conf: cfg,
		hub:  hub,
		ca:   chainAccessor,
		indexNamePrefix: "blockchain_",
	}
	svc.BaseComponent = component.NewBaseComponent("EsIndexer", svc, logger)
	return svc
}

func (ns *EsIndexer) CreateIndexIfNotExists(documentType string) {
	ctx := context.Background()
	exists, err := ns.client.IndexExists(ns.indexNamePrefix + documentType).Do(ctx)
	if err != nil {
		panic(err)
	}
	if !exists {
		createIndex, err := ns.client.CreateIndex(ns.indexNamePrefix + documentType).BodyString(mappings[documentType]).Do(ctx)
		if err != nil {
			panic(err)
		}
		if !createIndex.Acknowledged {
		}
		ns.Info().Str("name", documentType).Msg("Created index")
	}
}

func (ns *EsIndexer) BeforeStart() {
}

func (ns *EsIndexer) AfterStart() {
	client, err := elastic.NewClient(elastic.SetURL("http://127.0.0.1:9200"))
	if err != nil {
		ns.Warn().Err(err).Msg("Could not start elasticsearch indexer")
		ns.Stop()
		ns.hub.Unregister(ns)
		return
	}
	ns.client = client
	ns.CreateIndexIfNotExists("tx")
	ns.CreateIndexIfNotExists("block")
	ns.Info().Msg("Started Elasticsearch Indexer")
}

func (ns *EsIndexer) BeforeStop() {
}

// Index one block
func (ns *EsIndexer) IndexBlock(block *types.Block) {
	ctx := context.Background()
	esBlock := EsBlock{
		Timestamp: time.Unix(0, block.Header.Timestamp),
		BlockNo:   block.Header.BlockNo,
	}
	blockEncoded := util.ConvBlock(block)
	put, err := ns.client.Index().Index(ns.indexNamePrefix + "block").Type("block").Id(blockEncoded.Hash).BodyJson(esBlock).Do(ctx)
	if err != nil {
		ns.Error().Err(err).Msg("Failed to index block")
		return
	}
	ns.Info().Uint64("blockno", block.Header.BlockNo).Str("id", put.Id).Msg("Indexed block")

	if len(blockEncoded.Body.Txs) > 0 {
		chunkSize := 5000
		ns.IndexTxs(block, blockEncoded.Body.Txs, chunkSize)
	}
}

// Bulk index a list of transactions
func (ns *EsIndexer) IndexTxs(block *types.Block, txs []*util.InOutTx, chunkSize int) {
	// Setup a group of goroutines
	// The first goroutine will emit documents and send it to the second goroutine via the docsc channel.
	// The second goroutine will simply bulk insert the documents.
	g, ctx := errgroup.WithContext(context.TODO())
	docsc := make(chan EsTx)

	begin := time.Now()
	
	// Goroutine to create documents
	blockTs := time.Unix(0, block.Header.Timestamp)
	g.Go(func() error {
		defer close(docsc)

		for _, tx := range txs {
			d := EsTx{
				id:        tx.Hash,
				Timestamp: blockTs,
				BlockNo:   block.Header.BlockNo,
				Account:   tx.Body.Account,
				Recipient: tx.Body.Recipient,
			}

			// Send over to 2nd goroutine, or cancel
			select {
			case docsc <- d:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	// Second goroutine consumes the documents sent from the first and bulk insert into ES
	var total uint64
	g.Go(func() error {
		bulk := ns.client.Bulk().Index(ns.indexNamePrefix + "tx").Type("tx")
		for d := range docsc {
			atomic.AddUint64(&total, 1)
			bulk.Add(elastic.NewBulkIndexRequest().Id(d.id).Doc(d))
			if bulk.NumberOfActions() >= chunkSize {
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
					return errors.New("bulk commit failed")
				}
			}

			select {
			default:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Commit the final batch before exiting
		if bulk.NumberOfActions() > 0 {
			_, err := bulk.Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Wait until all goroutines are finished
	if err := g.Wait(); err != nil {
		ns.Fatal().Err(err)
	}

	// Final results
	dur := time.Since(begin).Seconds()
	sec := int(dur)
	pps := int64(float64(total) / dur)
	ns.Info().Uint64("total", total).Int("seconds", sec).Int64("tx/s", pps).Msg("Indexed transactions")
}


func (ns *EsIndexer) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *message.NotifyNewBlock:
		ns.IndexBlock(msg.Block)
	case *actor.Started:
	case *actor.Stopping:
	case *actor.Stopped:
		// Ignore actor lifecycle messages
	default:
		ns.Warn().Msgf("unknown msg received in esindexer %s", reflect.TypeOf(msg).String())
	}
}

const defaultTTL = time.Second * 4

// TellRequest implement interface method of ActorService
func (ns *EsIndexer) TellRequest(actor string, msg interface{}) {
	ns.TellTo(actor, msg)
}

// SendRequest implement interface method of ActorService
func (ns *EsIndexer) SendRequest(actor string, msg interface{}) {
	ns.RequestTo(actor, msg)
}

// FutureRequest implement interface method of ActorService
func (ns *EsIndexer) FutureRequest(actor string, msg interface{}, timeout time.Duration) *actor.Future {
	return ns.RequestToFuture(actor, msg, timeout)
}

// FutureRequestDefaultTimeout implement interface method of ActorService
func (ns *EsIndexer) FutureRequestDefaultTimeout(actor string, msg interface{}) *actor.Future {
	return ns.RequestToFuture(actor, msg, defaultTTL)
}

// CallRequest implement interface method of ActorService
func (ns *EsIndexer) CallRequest(actor string, msg interface{}, timeout time.Duration) (interface{}, error) {
	future := ns.RequestToFuture(actor, msg, timeout)
	return future.Result()
}

// CallRequest implement interface method of ActorService
func (ns *EsIndexer) CallRequestDefaultTimeout(actor string, msg interface{}) (interface{}, error) {
	future := ns.RequestToFuture(actor, msg, defaultTTL)
	return future.Result()
}
