package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/paveliak/go-workflows/backend"
	"github.com/paveliak/go-workflows/samples"
	scale "github.com/paveliak/go-workflows/samples/scale"
	"github.com/paveliak/go-workflows/worker"
)

var backendType = flag.String("backend", "redis", "backend to use: sqlite, mysql, redis")

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	b := samples.GetBackend("scale")

	// Run worker
	go RunWorker(ctx, b)

	c2 := make(chan os.Signal, 1)
	signal.Notify(c2, os.Interrupt)
	<-c2

	log.Println("Shutting down")
	cancel()
}

func RunWorker(ctx context.Context, mb backend.Backend) {
	w := worker.New(mb, &worker.Options{
		WorkflowPollers:          1,
		MaxParallelWorkflowTasks: 100,
		ActivityPollers:          1,
		MaxParallelActivityTasks: 100,
		HeartbeatWorkflowTasks:   false,
	})

	w.RegisterWorkflow(scale.Workflow1)

	w.RegisterActivity(scale.Activity1)
	w.RegisterActivity(scale.Activity2)

	if err := w.Start(ctx); err != nil {
		panic("could not start worker")
	}
}
