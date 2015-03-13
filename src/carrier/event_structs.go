package carrier

import "encoding/json"

type Event interface {
	To_JSON() string
}

type callEvent struct {
	Type           string
	CallId         int
	CallType       string
	CallStopReason string
	Source         int
	Destination    int
}

func (ce *callEvent) to_JSON() string {
	result, _ := json.Marshal(ce)
	return string(result)
}
