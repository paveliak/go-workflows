package workflowstate

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/paveliak/go-workflows/internal/converter"
	"github.com/paveliak/go-workflows/internal/core"
	"github.com/paveliak/go-workflows/internal/logger"
	"github.com/paveliak/go-workflows/internal/payload"
	"github.com/paveliak/go-workflows/internal/sync"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_PendingFutures(t *testing.T) {
	i := core.NewWorkflowInstance(uuid.NewString(), "")

	wfState := NewWorkflowState(i, logger.NewDefaultLogger(), clock.New())

	require.False(t, wfState.HasPendingFutures())

	f := sync.NewFuture[int]()
	wfState.TrackFuture(1, func(v payload.Payload, err error) error {
		var r int
		require.NoError(t, converter.DefaultConverter.From(v, &r))
		f.Set(r, nil)
		return nil
	})

	require.True(t, wfState.HasPendingFutures())

	wfState.RemoveFuture(1)

	require.False(t, wfState.HasPendingFutures())
}
