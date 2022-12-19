package worker

import (
	"time"

	"github.com/paveliak/go-workflows/internal/workflow"
)

type Options struct {
	// WorkflowsPollers is the number of pollers to start. Defaults to 2.
	WorkflowPollers int

	// MaxParallelWorkflowTasks determines the maximum number of concurrent workflow tasks processed
	// by the worker. The default is 0 which is no limit.
	MaxParallelWorkflowTasks int

	// ActivityPollers is the number of pollers to start. Defaults to 2.
	ActivityPollers int

	// MaxParallelActivityTasks determines the maximum number of concurrent activity tasks processed
	// by the worker. The default is 0 which is no limit.
	MaxParallelActivityTasks int

	// ActivityHeartbeatInterval is the interval between heartbeat attempts for activity tasks. Defaults
	// to 25 seconds
	ActivityHeartbeatInterval time.Duration

	// HeartbeatWorkflowTasks determines if the lock on workflow tasks should be periodically
	// extended while they are being processed. Given that workflow executions should be
	// very quick, this is usually not necessary.
	HeartbeatWorkflowTasks bool

	// WorkflowHeartbeatInterval is the interval between heartbeat attempts on workflow tasks, when enabled.
	WorkflowHeartbeatInterval time.Duration

	// WorkflowExecutorCache is the max size of the workflow executor cache. Defaults to 128
	WorkflowExecutorCacheSize int

	// WorkflowExecutorCache is the max TTL of the workflow executor cache. Defaults to 10 seconds
	WorkflowExecutorCacheTTL time.Duration

	// WorkflowExecutorCache is the cache to use for workflow executors. If nil, a default cache implementation
	// will be used.
	WorkflowExecutorCache workflow.ExecutorCache
}

var DefaultOptions = Options{
	WorkflowPollers:           2,
	ActivityPollers:           2,
	MaxParallelWorkflowTasks:  0,
	MaxParallelActivityTasks:  0,
	ActivityHeartbeatInterval: 25 * time.Second,
	WorkflowHeartbeatInterval: 25 * time.Second,

	WorkflowExecutorCacheSize: 128,
	WorkflowExecutorCacheTTL:  time.Second * 10,
	WorkflowExecutorCache:     nil,
}
