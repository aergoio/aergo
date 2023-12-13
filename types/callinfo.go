package types

type CallInfo struct {
	Name string        `json:"name,omitempty"`
	Args []interface{} `json:"args,omitempty"`
}
