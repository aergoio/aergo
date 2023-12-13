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

	"github.com/rs/cors"
	"golang.org/x/time/rate"
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

	// swagger
	mux.HandleFunc("/swagger.yaml", serveSwaggerYAML(cfg))
	mux.HandleFunc("/swagger", serveSwaggerUI(cfg))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// API v1
	web3svc := &Web3APIv1{rpc: rpc}

	var liminter *rate.Limiter
	if cfg.Web3.MaxLimit > 0 {
		liminter = rate.NewLimiter(rate.Limit(cfg.Web3.MaxLimit), 1)
	} else {
		liminter = rate.NewLimiter(rate.Inf, 0)
	}

	limitedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !liminter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		web3svc.handler(w, r)
	})
	mux.Handle("/v1/", c.Handler(limitedHandler))

	web3svr := &Web3{
		cfg:     cfg,
		web3svc: web3svc,
		mux:     mux,
		status:  CLOSE,
	}
	web3svr.BaseComponent = component.NewBaseComponent(message.Web3Svc, web3svr, logger)

	return web3svr
}

func (web3 *Web3) run() {
	port := getPortFromConfig(web3.cfg)
	web3.status = OPEN

	err := http.ListenAndServe(":"+strconv.Itoa(port), web3.mux)

	if err != nil {
		fmt.Println("Web3 Server running fail:", err)
		web3.status = CLOSE
	} else {
		fmt.Println("Web3 Server is listening on port " + strconv.Itoa(port) + "...")

	}
}

func (web3 *Web3) Statistics() *map[string]interface{} {
	ret := map[string]interface{}{
		"config": web3.cfg.Web3,
		"status": web3.status,
	}

	return &ret
}

func (web3 *Web3) BeforeStart() {
}

func (web3 *Web3) AfterStart() {
	fmt.Println("Web3 Server Start")
	web3.web3svc.NewHandler()
	go web3.run()
}

func (web3 *Web3) BeforeStop() {
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

func serveSwaggerYAML(cfg *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		yamlContent, err := os.ReadFile(cfg.Web3.SwaggerPath + "swagger.yaml")
		if err != nil {
			http.Error(w, "Failed to read YAML file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-yaml")
		w.Write(yamlContent)
	}
}

func serveSwaggerUI(cfg *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := os.ReadFile(cfg.Web3.SwaggerPath + "swagger-ui.html")
		if err != nil {
			http.Error(w, "Failed to read HTML file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlContent)
	}
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

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	})
}
