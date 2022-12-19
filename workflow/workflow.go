package workflow

import (
	"github.com/paveliak/go-workflows/internal/core"
)

type (
	Instance = core.WorkflowInstance
	Metadata = core.WorkflowMetadata
	Workflow = interface{}
)
