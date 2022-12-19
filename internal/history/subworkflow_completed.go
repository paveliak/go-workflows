package history

import "github.com/paveliak/go-workflows/internal/payload"

type SubWorkflowCompletedAttributes struct {
	Result payload.Payload `json:"result,omitempty"`
}
