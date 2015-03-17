package carrier

import "encoding/json"

type Event interface {
	To_JSON() string
}

type callResultEvent struct {
	Type     string
	CallId   int
	Decision bool
}

type callEvent struct {
	Type           string
	CallId         int
	CallType       string
	CallStopReason string
	Source         int
	Destination    int
}

func (cr *callResultEvent) to_JSON() string {
	result, _ := json.Marshal(cr)
	return string(result)
}

func (ce *callEvent) to_JSON() string {
	result, _ := json.Marshal(ce)
	return string(result)
}
