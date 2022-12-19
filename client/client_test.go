package client

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/paveliak/go-workflows/backend"
	"github.com/paveliak/go-workflows/internal/converter"
	"github.com/paveliak/go-workflows/internal/core"
	"github.com/paveliak/go-workflows/internal/history"
	"github.com/paveliak/go-workflows/internal/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Client_GetWorkflowResultTimeout(t *testing.T) {
	instance := core.NewWorkflowInstance(uuid.NewString(), "test")

	ctx := context.Background()

	b := &backend.MockBackend{}
	b.On("GetWorkflowInstanceState", mock.Anything, instance).Return(core.WorkflowInstanceStateActive, nil)

	c := &client{
		backend: b,
		clock:   clock.New(),
	}

	result, err := GetWorkflowResult[int](ctx, c, instance, time.Microsecond*1)
	require.Zero(t, result)
	require.EqualError(t, err, "workflow did not finish in time: workflow did not finish in specified timeout")
	b.AssertExpectations(t)
}

func Test_Client_GetWorkflowResultSuccess(t *testing.T) {
	instance := core.NewWorkflowInstance(uuid.NewString(), "test")

	ctx := context.Background()

	mockClock := clock.NewMock()

	r, _ := converter.DefaultConverter.To(42)

	b := &backend.MockBackend{}
	b.On("GetWorkflowInstanceState", mock.Anything, instance).Return(core.WorkflowInstanceStateActive, nil).Once().Run(func(args mock.Arguments) {
		// After the first call, advance the clock to immediately go to the second call below
		mockClock.Add(time.Second)
	})
	b.On("GetWorkflowInstanceState", mock.Anything, instance).Return(core.WorkflowInstanceStateFinished, nil)
	b.On("GetWorkflowInstanceHistory", mock.Anything, instance, (*int64)(nil)).Return([]history.Event{
		history.NewHistoryEvent(1, time.Now(), history.EventType_WorkflowExecutionStarted, &history.ExecutionStartedAttributes{}),
		history.NewHistoryEvent(2, time.Now(), history.EventType_WorkflowExecutionFinished, &history.ExecutionCompletedAttributes{
			Result: r,
			Error:  "",
		}),
	}, nil)

	c := &client{
		backend: b,
		clock:   mockClock,
	}

	result, err := GetWorkflowResult[int](ctx, c, instance, 0)
	require.Equal(t, 42, result)
	require.NoError(t, err)
	b.AssertExpectations(t)
}

func Test_Client_SignalWorkflow(t *testing.T) {
	instanceID := uuid.NewString()

	ctx := context.Background()

	b := &backend.MockBackend{}
	b.On("Logger").Return(logger.NewDefaultLogger())
	b.On("SignalWorkflow", ctx, instanceID, mock.MatchedBy(func(event history.Event) bool {
		return event.Type == history.EventType_SignalReceived &&
			event.Attributes.(*history.SignalReceivedAttributes).Name == "test"
	})).Return(nil)

	c := &client{
		backend: b,
		clock:   clock.New(),
	}

	err := c.SignalWorkflow(ctx, instanceID, "test", "signal")

	require.Nil(t, err)
	b.AssertExpectations(t)
}

func Test_Client_SignalWorkflow_WithArgs(t *testing.T) {
	instanceID := uuid.NewString()

	ctx := context.Background()

	arg := 42

	input, _ := converter.DefaultConverter.To(arg)

	b := &backend.MockBackend{}
	b.On("Logger").Return(logger.NewDefaultLogger())
	b.On("SignalWorkflow", ctx, instanceID, mock.MatchedBy(func(event history.Event) bool {
		return event.Type == history.EventType_SignalReceived &&
			event.Attributes.(*history.SignalReceivedAttributes).Name == "test" &&
			bytes.Equal(event.Attributes.(*history.SignalReceivedAttributes).Arg, input)
	})).Return(nil)

	c := &client{
		backend: b,
		clock:   clock.New(),
	}

	err := c.SignalWorkflow(ctx, instanceID, "test", arg)

	require.Nil(t, err)
	b.AssertExpectations(t)
}
