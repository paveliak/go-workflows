package workflow

import "github.com/paveliak/go-workflows/internal/sync"

type Channel[T any] interface {
	Send(ctx Context, v T)

	SendNonblocking(v T) (ok bool)

	Receive(ctx Context) (v T, ok bool)

	ReceiveNonBlocking() (v T, ok bool)

	Close()
}

func NewChannel[T any]() Channel[T] {
	return sync.NewChannel[T]()
}

func NewBufferedChannel[T any](size int) Channel[T] {
	return sync.NewBufferedChannel[T](size)
}
