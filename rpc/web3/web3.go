package web3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/rpc"
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

func NewWeb3(rpc *rpc.AergoRPCService) {
	// swagger setting
	http.HandleFunc("/swagger.yaml", serveSwaggerYAML)
	http.HandleFunc("/swagger", serveSwaggerUI)

	// v1
	web3svcV1 := &Web3APIv1{rpc: rpc}
	http.HandleFunc("/v1/", web3svcV1.handler)
	
	go func() {
		fmt.Println("Web3 Server is listening on port 80...")
    	http.ListenAndServe(":80", nil)
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

