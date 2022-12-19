package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/paveliak/go-workflows/backend"
	"github.com/paveliak/go-workflows/client"
	"github.com/paveliak/go-workflows/samples"
	"github.com/paveliak/go-workflows/worker"
	"github.com/paveliak/go-workflows/workflow"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	b := samples.GetBackend("concurrent")

	// Run worker
	go RunWorker(ctx, b)

	// Start workflow via client
	c := client.New(b)

	startWorkflow(ctx, c)

	c2 := make(chan os.Signal, 1)
	signal.Notify(c2, os.Interrupt)
	<-c2
}

func startWorkflow(ctx context.Context, c client.Client) {
	wf, err := c.CreateWorkflowInstance(ctx, client.WorkflowInstanceOptions{
		InstanceID: uuid.NewString(),
	}, Workflow1, "Hello world")
	if err != nil {
		panic("could not start workflow")
	}

	log.Println("Started workflow", wf.InstanceID)
}

func RunWorker(ctx context.Context, mb backend.Backend) {
	w := worker.New(mb, nil)

	w.RegisterWorkflow(Workflow1)

	w.RegisterActivity(Activity1)
	w.RegisterActivity(Activity2)

	if err := w.Start(ctx); err != nil {
		panic("could not start worker")
	}
}

func Workflow1(ctx workflow.Context, msg string) (string, error) {
	logger := workflow.Logger(ctx)
	logger.Debug("Entering Workflow1")
	logger.Debug("\tWorkflow instance input:", msg)

	defer func() {
		logger.Debug("Leaving Workflow1")
	}()

	a1 := workflow.ExecuteActivity[int](ctx, workflow.DefaultActivityOptions, Activity1, 35, 12)
	a2 := workflow.ExecuteActivity[int](ctx, workflow.DefaultActivityOptions, Activity2)

	results := 0

	for results < 2 {
		logger.Debug("Selecting...")

		workflow.Select(
			ctx,
			workflow.Await(a2, func(ctx workflow.Context, f2 workflow.Future[int]) {
				r, err := f2.Get(ctx)
				if err != nil {
					panic(err)
				}

				logger.Debug("A2 result", "r", r)
				results++
			}),
			workflow.Await(a1, func(ctx workflow.Context, f1 workflow.Future[int]) {
				r, err := f1.Get(ctx)
				if err != nil {
					panic(err)
				}

				logger.Debug("A1 result", "r", r)

				results++
			}),
		)

		logger.Debug("Selected")
	}

	return "result", nil
}

func Activity1(ctx context.Context, a, b int) (int, error) {
	log.Println("Entering Activity1")

	defer func() {
		log.Println("Leaving Activity1")
	}()

	return a + b, nil
}

func Activity2(ctx context.Context) (int, error) {
	log.Println("Entering Activity2")

	time.Sleep(5 * time.Second)

	defer func() {
		log.Println("Leaving Activity2")
	}()

	return 12, nil
}
