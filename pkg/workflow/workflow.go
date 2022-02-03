package workflow

import (
	"time"

	"github.com/cschleiden/go-dt/internal/sync"
	"github.com/cschleiden/go-dt/internal/workflow"
	internal "github.com/cschleiden/go-dt/internal/workflow"
	"github.com/cschleiden/go-dt/pkg/core"
)

type (
	Workflow           = internal.Workflow
	SubWorkflowOptions = internal.SubWorkflowOptions
	Activity           = internal.Activity
	ActivityOptions    = internal.ActivityOptions
	RetryOptions       = internal.RetryOptions
)

var DefaultRetryOptions = workflow.DefaultRetryOptions

func Replaying(ctx Context) bool {
	return internal.Replaying(ctx)
}

func CreateSubWorkflowInstance(ctx Context, options SubWorkflowOptions, workflow Workflow, args ...interface{}) Future {
	return internal.CreateSubWorkflowInstance(ctx, options, workflow, args...)
}

var DefaultActivityOptions = internal.DefaultActivityOptions

// ExecuteActivity schedules the given activity to be executed
func ExecuteActivity(ctx Context, options ActivityOptions, activity Activity, args ...interface{}) Future {
	return internal.ExecuteActivity(ctx, options, activity, args...)
}

func ScheduleTimer(ctx Context, delay time.Duration) Future {
	return internal.ScheduleTimer(ctx, delay)
}

func NewSignalChannel(ctx Context, name string) Channel {
	return internal.NewSignalChannel(ctx, name)
}

func WorkflowInstance(ctx Context) core.WorkflowInstance {
	return internal.WorkflowInstance(ctx)
}

func Now(ctx Context) time.Time {
	return internal.Now(ctx)
}

func Sleep(ctx Context, d time.Duration) error {
	return internal.Sleep(ctx, d)
}

func Go(ctx Context, f func(ctx Context)) {
	sync.Go(ctx, f)
}

func NewSelector() sync.Selector {
	return sync.NewSelector()
}

func NewChannel() Channel {
	return sync.NewChannel()
}

func NewBufferedChannel(size int) Channel {
	return sync.NewBufferedChannel(size)
}
