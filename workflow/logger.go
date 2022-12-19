package workflow

import (
	"github.com/paveliak/go-workflows/internal/workflowstate"
	"github.com/paveliak/go-workflows/log"
)

func Logger(ctx Context) log.Logger {
	wfState := workflowstate.WorkflowState(ctx)
	return wfState.Logger()
}
