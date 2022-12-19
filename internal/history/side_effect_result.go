package history

import "github.com/paveliak/go-workflows/internal/payload"

type SideEffectResultAttributes struct {
	Result payload.Payload `json:"result,omitempty"`
}
