package cache

import (
	"context"
	"time"

	"github.com/paveliak/go-workflows/internal/core"
	"github.com/paveliak/go-workflows/internal/metrickeys"
	"github.com/paveliak/go-workflows/internal/workflow"
	"github.com/paveliak/go-workflows/metrics"
	"github.com/jellydator/ttlcache/v3"
)

type LruCache struct {
	mc metrics.Client
	c  *ttlcache.Cache[string, workflow.WorkflowExecutor]
}

func NewWorkflowExecutorLRUCache(mc metrics.Client, size int, expiration time.Duration) workflow.ExecutorCache {
	c := ttlcache.New(
		ttlcache.WithCapacity[string, workflow.WorkflowExecutor](uint64(size)),
		ttlcache.WithTTL[string, workflow.WorkflowExecutor](expiration),
	)

	c.OnEviction(func(ctx context.Context, er ttlcache.EvictionReason, i *ttlcache.Item[string, workflow.WorkflowExecutor]) {
		// Close the executor to allow it to clean up resources.
		i.Value().Close()

		reason := ""
		switch er {
		case ttlcache.EvictionReasonExpired:
			reason = "expired"
		case ttlcache.EvictionReasonCapacityReached:
			reason = "capacity"
		}

		mc.Counter(metrickeys.WorkflowInstanceCacheEviction, metrics.Tags{metrickeys.EvictionReason: reason}, 1)
	})

	return &LruCache{
		mc: mc,
		c:  c,
	}
}

func (lc *LruCache) Get(ctx context.Context, instance *core.WorkflowInstance) (workflow.WorkflowExecutor, bool, error) {
	e := lc.c.Get(getKey(instance))
	if e != nil {
		return e.Value(), true, nil
	}

	return nil, false, nil
}

func (lc *LruCache) Store(ctx context.Context, instance *core.WorkflowInstance, executor workflow.WorkflowExecutor) error {
	lc.c.Set(getKey(instance), executor, ttlcache.DefaultTTL)

	lc.mc.Gauge(metrickeys.WorkflowInstanceCacheSize, metrics.Tags{}, int64(lc.c.Len()))

	return nil
}

func (lc *LruCache) StartEviction(ctx context.Context) {
	go lc.c.Start()

	<-ctx.Done()

	lc.c.Stop()
}
