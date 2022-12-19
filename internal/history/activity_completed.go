package history

import "github.com/paveliak/go-workflows/internal/payload"

type ActivityCompletedAttributes struct {
	Result payload.Payload `json:"result,omitempty"`
}
