package history

import "github.com/paveliak/go-workflows/internal/payload"

type ExecutionCompletedAttributes struct {
	Result payload.Payload `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}
