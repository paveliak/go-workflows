package history

import (
	"github.com/paveliak/go-workflows/internal/core"
	"github.com/paveliak/go-workflows/internal/payload"
)

type ExecutionStartedAttributes struct {
	Name string `json:"name,omitempty"`

	Metadata *core.WorkflowMetadata `json:"metadata,omitempty"`

	Inputs []payload.Payload `json:"inputs,omitempty"`
}
