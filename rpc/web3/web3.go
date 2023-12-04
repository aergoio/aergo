package web3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	cfg "github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/rpc"
	"github.com/aergoio/aergo/v2/types/message"
	"github.com/didip/tollbooth"
	"github.com/rs/cors"
)

type Status string

const (
	CLOSE Status = "CLOSE"
	OPEN  Status = "OPEN"
)

type Web3 struct {
	*component.BaseComponent
	cfg *cfg.Config

	web3svc *Web3APIv1
	mux     *http.ServeMux

	port   int
	status Status
}

var (
	prefixV1 = "/v1"
)

var (
	logger = log.NewLogger("web3")
)

func NewWeb3(cfg *config.Config, rpc *rpc.AergoRPCService) *Web3 {
	mux := http.NewServeMux()

	// set limit per second
	maxLimit := float64(1)
	if cfg.Web3.MaxLimit != 0 {
		maxLimit = float64(cfg.Web3.MaxLimit)
	}

	limiter := tollbooth.NewLimiter(maxLimit, nil)
	limiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})

	// swagger
	mux.HandleFunc("/swagger.yaml", serveSwaggerYAML)
	mux.HandleFunc("/swagger", serveSwaggerUI)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// API v1
	web3svc := &Web3APIv1{rpc: rpc}
	web3svc.NewHandler()
	mux.Handle("/v1/", tollbooth.LimitHandler(limiter, c.Handler(http.HandlerFunc(web3svc.handler))))

	web3svr := &Web3{
		cfg:     cfg,
		web3svc: web3svc,
		mux:     mux,
		status:  CLOSE,
	}
	web3svr.BaseComponent = component.NewBaseComponent(message.Web3Svc, web3svr, logger)

	return web3svr
}

func (web3 *Web3) Start() {
	go web3.run()
}

func (web3 *Web3) run() {
	port := getPortFromConfig(web3.cfg)
	err := http.ListenAndServe(":"+strconv.Itoa(port), web3.mux)

	if err != nil {
		fmt.Println("Web3 Server running fail:", err)
	} else {
		fmt.Println("Web3 Server is listening on port " + strconv.Itoa(port) + "...")
		web3.port = port
		web3.status = OPEN
	}
}

func (web3 *Web3) Statistics() *map[string]interface{} {
	ret := map[string]interface{}{
		"status": web3.status,
		"port":   web3.port,
	}

	return &ret
}

func (web3 *Web3) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started, *actor.Stopping, *actor.Stopped, *component.CompStatReq:
	default:
		web3.Warn().Msgf("unknown msg received in web3 %s", reflect.TypeOf(msg).String())
	}
}

func getPortFromConfig(cfg *config.Config) int {
	if cfg == nil || cfg.Web3 == nil || cfg.Web3.NetServicePort == 0 {
		return 80
	}

	return cfg.Web3.NetServicePort
}

func serveSwaggerYAML(w http.ResponseWriter, r *http.Request) {
	yamlContent, err := os.ReadFile("./rpc/swagger/swagger.yaml")
	if err != nil {
		http.Error(w, "Failed to read YAML file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write(yamlContent)
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile("./rpc/swagger/swagger-ui.html")
	if err != nil {
		http.Error(w, "Failed to read HTML file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(htmlContent)
}

func commonResponseHandler(response interface{}, err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	})
}

func stringResponseHandler(response string, err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// jsonResponse, err := json.Marshal(response)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	})
}
