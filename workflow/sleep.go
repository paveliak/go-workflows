package workflow

import (
	"time"

	"github.com/paveliak/go-workflows/internal/sync"
	"github.com/paveliak/go-workflows/internal/workflowtracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Sleep(ctx sync.Context, d time.Duration) error {
	ctx, span := workflowtracer.Tracer(ctx).Start(ctx, "Sleep",
		trace.WithAttributes(attribute.Int64("duration_ms", int64(d/time.Millisecond))))
	defer span.End()

	_, err := ScheduleTimer(ctx, d).Get(ctx)

	return err
}
