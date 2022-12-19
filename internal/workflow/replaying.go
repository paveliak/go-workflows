package workflow

import (
	"github.com/paveliak/go-workflows/internal/sync"
	"github.com/paveliak/go-workflows/internal/workflowstate"
)

func Replaying(ctx sync.Context) bool {
	wfState := workflowstate.WorkflowState(ctx)
	return wfState.Replaying()
}
