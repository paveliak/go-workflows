package workflow

import (
	"github.com/paveliak/go-workflows/internal/core"
	"github.com/paveliak/go-workflows/internal/sync"
	"github.com/paveliak/go-workflows/internal/workflowstate"
)

func WorkflowInstance(ctx sync.Context) *core.WorkflowInstance {
	wfState := workflowstate.WorkflowState(ctx)
	return wfState.Instance()
}
