package api

type CallRequest struct {
	CapabilityID string
	Operation    string
	Args         map[string]any
}

type CallResult struct {
	Value map[string]any
}

type PollResult struct {
	Done  bool
	Value map[string]any
	Error string
}
