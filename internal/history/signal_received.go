package history

import "github.com/paveliak/go-workflows/internal/payload"

type SignalReceivedAttributes struct {
	Name string          `json:"name,omitempty"`
	Arg  payload.Payload `json:"arg,omitempty"`
}
