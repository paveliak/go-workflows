package workflow

import (
	"time"

	"github.com/paveliak/go-workflows/internal/sync"
	"github.com/paveliak/go-workflows/internal/workflowstate"
)

func Now(ctx sync.Context) time.Time {
	wfState := workflowstate.WorkflowState(ctx)
	return wfState.Time()
}
