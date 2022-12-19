package test

import (
	"context"

	"github.com/paveliak/go-workflows/backend"
	"github.com/paveliak/go-workflows/internal/history"
)

type TestBackend interface {
	backend.Backend

	GetFutureEvents(ctx context.Context) ([]history.Event, error)
}
