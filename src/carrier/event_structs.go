package carrier

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

type userMessage struct {
	Type        string `json:"type"`
	CallId      int    `json:"call_id,omitempty"`
	Action      string `json:"action"`
	Source      int    `json:"source,omitempty"`
	Destination int    `json:"destination, omitempty"`
	CreatedAt   string `json:"created_at"`
	Text        string `json:"text"`
}
