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
	"reflect"
	//"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
	bc "github.com/aergoio/aergo/blockchain"
	cfg "github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/pkg/log"
)

type RestService struct {
	*component.BaseComponent

	cfg *cfg.Config
	bc  *bc.ChainService
}

var _ component.IComponent = (*RestService)(nil)

//var wait sync.WaitGroup

var (
	logger = log.NewLogger(log.Rest)
)

func NewRestService(cfg *cfg.Config, bc *bc.ChainService) *RestService {
	cs := &RestService{
		BaseComponent: component.NewBaseComponent(message.RestSvc, logger, cfg.EnableDebugMsg),
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
				logger.Errorf("Error reading body: %v", err)
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}
			logger.Debugf("Recieved: ", string(body))
			// Sorry, Just for ChainTree lookup now
			i, _ := cs.bc.GetChainTree()
			w.Write(i)
		})
		logger.Infof("Rest Service Started using port(%v)", cs.cfg.REST.RestPort)
		portNo := fmt.Sprintf(":%v", cs.cfg.REST.RestPort)
		logger.Info(http.ListenAndServe(portNo, nil))
	}()
}

func (cs *RestService) Stop() {
	//wait.Wait()
	cs.BaseComponent.Stop()
}

func (cs *RestService) Receive(context actor.Context) {
	cs.BaseComponent.Receive(context)

	switch msg := context.Message().(type) {
	default:
		logger.Debugf("Missed message. (%v) %s", reflect.TypeOf(msg), msg)
	}
}
