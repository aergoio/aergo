/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package restservice

import (
	//"html"
	//	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	//"sync"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	bc "github.com/aergoio/aergo/blockchain"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
)

type RestService struct {
	*component.BaseComponent

	cfg *cfg.Config
	bc  *bc.ChainService
}

var _ component.IComponent = (*RestService)(nil)

//var wait sync.WaitGroup

var (
	logger = log.NewLogger("rest")
)

func NewRestService(cfg *cfg.Config, bc *bc.ChainService) *RestService {
	cs := &RestService{
		BaseComponent: component.NewBaseComponent(message.RestSvc, logger),
		cfg:           cfg,
		bc:            bc,
	}

	return cs
}

func (cs *RestService) Start() {
	cs.BaseComponent.Start(cs)
	//wait.Add(1)
	go func() {
		http.HandleFunc("/chaintree", func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logger.Error().Err(err).Msg("Error reading body")
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}
			logger.Debug().Str("body", string(body)).Msg("Recieved")
			// Sorry, Just for ChainTree lookup now
			i, _ := cs.bc.GetChainTree()
			w.Write(i)
		})
		logger.Info().Int("port", cs.cfg.REST.RestPort).Msg("Rest Service Started")
		portNo := fmt.Sprintf(":%v", cs.cfg.REST.RestPort)
		err := http.ListenAndServe(portNo, nil)
		logger.Info().Err(err).Msg("Start rest server")
	}()
}

func (cs *RestService) Stop() {
	//wait.Wait()
	cs.BaseComponent.Stop()
}

func (cs *RestService) Receive(context actor.Context) {
	cs.BaseComponent.Receive(context)

	switch msg := context.Message().(type) {
	case *component.CompStatReq:
		context.Respond(
			&component.CompStatRsp{
				"component": cs.BaseComponent.Statics(msg),
			})
	}
}
