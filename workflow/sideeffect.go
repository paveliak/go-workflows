package workflow

import (
	"github.com/paveliak/go-workflows/internal/command"
	"github.com/paveliak/go-workflows/internal/converter"
	"github.com/paveliak/go-workflows/internal/sync"
	"github.com/paveliak/go-workflows/internal/workflowstate"
	"github.com/paveliak/go-workflows/internal/workflowtracer"
)

func SideEffect[TResult any](ctx Context, f func(ctx Context) TResult) Future[TResult] {
	ctx, span := workflowtracer.Tracer(ctx).Start(ctx, "SideEffect")
	defer span.End()

	future := sync.NewFuture[TResult]()

	if ctx.Err() != nil {
		future.Set(*new(TResult), ctx.Err())
		return future
	}

	wfState := workflowstate.WorkflowState(ctx)
	scheduleEventID := wfState.GetNextScheduleEventID()

	wfState.TrackFuture(scheduleEventID, workflowstate.AsDecodingSettable(future))

	cmd := command.NewSideEffectCommand(scheduleEventID)
	wfState.AddCommand(cmd)

	if !Replaying(ctx) {
		// Execute side effect
		r := f(ctx)

		payload, err := converter.DefaultConverter.To(r)
		if err != nil {
			future.Set(*new(TResult), err)
		}

		cmd.SetResult(payload)
		future.Set(r, nil)
		wfState.RemoveFuture(scheduleEventID)
	}

	return future
}
