package jsonrpc

import (
	"encoding/json"

	"github.com/aergoio/aergo/v2/types"
)

func ConvNodeState(msg *types.SingleBytes) interface{} {
	var ns interface{}
	_ = json.Unmarshal(msg.GetValue(), &ns)
	
	return ns
}

type InOutNodeState struct {
	AccountsSvc 	map[string]interface{} `json:"AccountsSvc,omitempty"`
	ChainSvc 	map[string]interface{} `json:"ChainSvc,omitempty"`
	MemPoolSvc 	map[string]interface{} `json:"MemPoolSvc,omitempty"`
	RPCSvc		map[string]interface{} `json:"RPCSvc,omitempty"`
	SyncerSvc	map[string]interface{} `json:"SyncerSvc,omitempty"`
	MapSvc		map[string]interface{} `json:"mapSvc,omitempty"`
	P2pSvc		map[string]interface{} `json:"p2pSvc,omitempty"`
}


func ConvConfigItem(msg *types.ConfigItem) *InOutConfigItem {
	ci := &InOutConfigItem{}
	
	ci.Props = make(map[string]string)
	for key, value := range msg.Props {
		ci.Props[key] = value
    }
	return ci
}

type InOutConfigItem struct {
	Props map[string]string `json:"props,omitempty"`
}

func ConvServerInfo(msg *types.ServerInfo) *InOutServerInfo {
	si := &InOutServerInfo{}	
	
	si.Status = make(map[string]string)
	for key, status := range msg.Status {
		si.Status[key] = status
    }

	si.Config = make(map[string]*InOutConfigItem)
	for key, value := range msg.Config {
		si.Config[key] = ConvConfigItem(value)
    }
	
	return si
}

type InOutServerInfo struct {
	Status map[string]string      		`json:"status,omitempty"`
	Config map[string]*InOutConfigItem 	`json:"config,omitempty"`
}




