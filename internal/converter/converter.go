package converter

import (
	"github.com/paveliak/go-workflows/internal/payload"
)

type Converter interface {
	To(v interface{}) (payload.Payload, error)
	From(data payload.Payload, v interface{}) error
}

var DefaultConverter Converter = &jsonConverter{}
