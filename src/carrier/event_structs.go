package carrier

import "encoding/json"

type Event interface {
	To_JSON() string
}

type callResultEvent struct {
	Type     string `json:"type"`
	CallId   int    `json:"call_id"`
	Decision bool   `json:"decision"`
}

type callEvent struct {
	Type           string `json:"type"`
	CallId         int    `json:"call_id"`
	CallType       string `json:"call_type"`
	CallStopReason string `json:"call_stop_reason,omitempty"`
	Source         int    `json:"source"`
	Destination    int    `json:"destination"`
}

type callAftermathEvent struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	CallId int    `json:"call_id"`
}

func (cr *callResultEvent) to_JSON() string {
	result, _ := json.Marshal(cr)
	return string(result)
}

func (ce *callEvent) to_JSON() string {
	result, _ := json.Marshal(ce)
	return string(result)
}

func (ce *callAftermathEvent) to_JSON() string {
	result, _ := json.Marshal(ce)
	return string(result)
}
