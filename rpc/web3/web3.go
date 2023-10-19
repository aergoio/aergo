package web3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/rpc"

	"github.com/aergoio/aergo/config"
	"github.com/didip/tollbooth"
	"github.com/rs/cors"
)

type RestAPI struct {
	rpc *rpc.AergoRPCService
	request *http.Request
}

var (
	prefixV1 = "/v1"
)

var (
	logger = log.NewLogger("web3")
)

func NewWeb3(cfg *config.Config, rpc *rpc.AergoRPCService) {
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
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"*"},
		AllowCredentials: true,
	})
	
	// API v1
	web3svcV1 := &Web3APIv1{rpc: rpc}
	mux.Handle("/v1/", tollbooth.LimitHandler(limiter, c.Handler(http.HandlerFunc(web3svcV1.handler))))

	go func() {		
		fmt.Println("Web3 Server is listening on port "+ strconv.Itoa(cfg.Web3.NetServicePort)+"...")
		http.ListenAndServe(":"+strconv.Itoa(cfg.Web3.NetServicePort), mux)
	}()
}

func serveSwaggerYAML(w http.ResponseWriter, r *http.Request) {
	yamlContent, err := ioutil.ReadFile("./swagger/swagger.yaml")
	if err != nil {
		http.Error(w, "Failed to read YAML file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write(yamlContent)
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := ioutil.ReadFile("./swagger/swagger-ui.html")
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

