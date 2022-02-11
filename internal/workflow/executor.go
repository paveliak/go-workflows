package workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/cschleiden/go-dt/internal/command"
	"github.com/cschleiden/go-dt/internal/payload"
	"github.com/cschleiden/go-dt/internal/sync"
	"github.com/cschleiden/go-dt/pkg/core"
	"github.com/cschleiden/go-dt/pkg/core/task"
	"github.com/cschleiden/go-dt/pkg/history"
	"github.com/google/uuid"
	errs "github.com/pkg/errors"
)

type WorkflowExecutor interface {
	ExecuteTask(ctx context.Context, t *task.Workflow) ([]history.Event, []core.WorkflowEvent, error)

	Close()
}

type executor struct {
	registry          *Registry
	workflow          *workflow
	workflowState     *workflowState
	workflowCtx       sync.Context
	workflowCtxCancel sync.CancelFunc
	logger            *log.Logger
	lastEventID       string // TODO: Not the same as the sequence number Event ID
}

func NewExecutor(registry *Registry, instance core.WorkflowInstance) (WorkflowExecutor, error) {
	state := newWorkflowState(instance)
	wfCtx, cancel := sync.WithCancel(withWfState(sync.Background(), state))

	return &executor{
		registry:          registry,
		workflowState:     state,
		workflowCtx:       wfCtx,
		workflowCtxCancel: cancel,
		logger:            log.New(io.Discard, "", log.LstdFlags),
		//logger: log.Default(),
	}, nil
}

func (e *executor) ExecuteTask(ctx context.Context, t *task.Workflow) ([]history.Event, []core.WorkflowEvent, error) {
	if t.Kind == task.Continuation {
		// Check if the current state matches the backend's history state
		newestHistoryEvent := t.History[len(t.History)-1]
		if newestHistoryEvent.ID != e.lastEventID {
			return nil, nil, fmt.Errorf("mismatch in execution, last event %v not found in history, last there is %v (%v)", e.lastEventID, newestHistoryEvent.ID, newestHistoryEvent.Type)
		}

		// Clear commands from previous executions
		e.workflowState.clearCommands()
	} else {
		// Replay history
		e.workflowState.setReplaying(true)
		for _, event := range t.History {
			if err := e.executeEvent(event); err != nil {
				return nil, nil, errs.Wrap(err, "error while replaying event")
			}
		}
	}

	// Always pad the received events with WorkflowTaskStarted/Finished events to indicate the execution
	events := []history.Event{history.NewHistoryEvent(history.EventType_WorkflowTaskStarted, -1, &history.WorkflowTaskStartedAttributes{})}
	events = append(events, t.NewEvents...)

	// Execute new events received from the backend
	if err := e.executeNewEvents(events); err != nil {
		return nil, nil, errs.Wrap(err, "error while executing new events")
	}

	newCommandEvents, workflowEvents, err := e.processCommands(ctx, t)
	if err != nil {
		return nil, nil, errs.Wrap(err, "could not process commands")
	}
	events = append(events, newCommandEvents...)

	// Execution of this task is finished, add event to history
	events = append(events, history.NewHistoryEvent(history.EventType_WorkflowTaskFinished, -1, &history.WorkflowTaskFinishedAttributes{}))

	// TODO: The finished event isn't actually executed, does this make sense?
	e.lastEventID = events[len(events)-1].ID

	return events, workflowEvents, nil
}

func (e *executor) executeNewEvents(newEvents []history.Event) error {
	e.workflowState.setReplaying(false)

	for _, event := range newEvents {
		if err := e.executeEvent(event); err != nil {
			return errs.Wrap(err, "error while executing event")
		}

		// Remember that we executed this event last
		e.lastEventID = event.ID
	}

	if e.workflow.Completed() {
		if err := e.workflowCompleted(e.workflow.Result(), e.workflow.Error()); err != nil {
			return err
		}
	}

	return nil
}

func (e *executor) Close() {
	if e.workflow != nil {
		// End workflow if running to prevent leaking goroutines
		e.workflow.Close(withWfState(sync.Background(), e.workflowState)) // TODO: Cache this context?
	}
}

func (e *executor) executeEvent(event history.Event) error {
	e.logger.Println("Handling:", event.Type)

	var err error

	switch event.Type {
	case history.EventType_WorkflowExecutionStarted:
		err = e.handleWorkflowExecutionStarted(event.Attributes.(*history.ExecutionStartedAttributes))

	case history.EventType_WorkflowExecutionFinished:
	// Ignore

	case history.EventType_WorkflowExecutionCanceled:
		err = e.handleWorkflowCanceled()

	case history.EventType_WorkflowTaskStarted:
		err = e.handleWorkflowTaskStarted(event, event.Attributes.(*history.WorkflowTaskStartedAttributes))

	case history.EventType_WorkflowTaskFinished:
		// Ignore

	case history.EventType_ActivityScheduled:
		err = e.handleActivityScheduled(event, event.Attributes.(*history.ActivityScheduledAttributes))

	case history.EventType_ActivityFailed:
		err = e.handleActivityFailed(event, event.Attributes.(*history.ActivityFailedAttributes))

	case history.EventType_ActivityCompleted:
		err = e.handleActivityCompleted(event, event.Attributes.(*history.ActivityCompletedAttributes))

	case history.EventType_TimerScheduled:
		err = e.handleTimerScheduled(event, event.Attributes.(*history.TimerScheduledAttributes))

	case history.EventType_TimerFired:
		err = e.handleTimerFired(event, event.Attributes.(*history.TimerFiredAttributes))

	case history.EventType_SignalReceived:
		err = e.handleSignalReceived(event, event.Attributes.(*history.SignalReceivedAttributes))

	case history.EventType_SideEffectResult:
		err = e.handleSideEffectResult(event, event.Attributes.(*history.SideEffectResultAttributes))

	case history.EventType_SubWorkflowScheduled:
		err = e.handleSubWorkflowScheduled(event, event.Attributes.(*history.SubWorkflowScheduledAttributes))

	case history.EventType_SubWorkflowCompleted:
		err = e.handleSubWorkflowCompleted(event, event.Attributes.(*history.SubWorkflowCompletedAttributes))

	default:
		return fmt.Errorf("unknown event type: %v", event.Type)
	}

	return err
}

func (e *executor) handleWorkflowExecutionStarted(a *history.ExecutionStartedAttributes) error {
	wfFn, err := e.registry.GetWorkflow(a.Name)
	if err != nil {
		return fmt.Errorf("workflow %s not found", a.Name)
	}

	e.workflow = NewWorkflow(reflect.ValueOf(wfFn))

	return e.workflow.Execute(e.workflowCtx, a.Inputs)
}

func (e *executor) handleWorkflowCanceled() error {
	e.workflowCtxCancel()

	return e.workflow.Continue(e.workflowCtx)
}

func (e *executor) handleWorkflowTaskStarted(event history.Event, a *history.WorkflowTaskStartedAttributes) error {
	e.workflowState.setTime(event.Timestamp)

	return nil
}

func (e *executor) handleActivityScheduled(event history.Event, a *history.ActivityScheduledAttributes) error {
	c := e.workflowState.removeCommandByEventID(event.EventID)
	if c != nil {
		// Ensure the same activity is scheduled again
		ca := c.Attr.(*command.ScheduleActivityTaskCommandAttr)
		if a.Name != ca.Name {
			return fmt.Errorf("previous workflow execution scheduled different type of activity: %s, %s", a.Name, ca.Name)
		}
	}

	return nil
}

func (e *executor) handleActivityCompleted(event history.Event, a *history.ActivityCompletedAttributes) error {
	f, ok := e.workflowState.pendingFutures[event.EventID]
	if !ok {
		return nil
	}

	e.workflowState.removeCommandByEventID(event.EventID)
	f.Set(a.Result, nil)

	return e.workflow.Continue(e.workflowCtx)
}

func (e *executor) handleActivityFailed(event history.Event, a *history.ActivityFailedAttributes) error {
	f, ok := e.workflowState.pendingFutures[event.EventID]
	if !ok {
		return errors.New("no pending future found for activity failed event")
	}

	e.workflowState.removeCommandByEventID(event.EventID)

	f.Set(nil, errors.New(a.Reason))

	return e.workflow.Continue(e.workflowCtx)
}

func (e *executor) handleTimerScheduled(event history.Event, a *history.TimerScheduledAttributes) error {
	e.workflowState.removeCommandByEventID(event.EventID)

	return nil
}

func (e *executor) handleTimerFired(event history.Event, a *history.TimerFiredAttributes) error {
	f, ok := e.workflowState.pendingFutures[event.EventID]
	if !ok {
		// Timer already canceled ignore
		return nil
	}

	e.workflowState.removeCommandByEventID(event.EventID)

	f.Set(nil, nil)

	return e.workflow.Continue(e.workflowCtx)
}

func (e *executor) handleSubWorkflowScheduled(event history.Event, a *history.SubWorkflowScheduledAttributes) error {
	c := e.workflowState.removeCommandByEventID(event.EventID)
	if c != nil {
		ca := c.Attr.(*command.ScheduleSubWorkflowCommandAttr)
		if a.Name != ca.Name {
			return errors.New("previous workflow execution scheduled a different sub workflow")
		}
	}

	return nil
}

func (e *executor) handleSubWorkflowCompleted(event history.Event, a *history.SubWorkflowCompletedAttributes) error {
	f, ok := e.workflowState.pendingFutures[event.EventID]
	if !ok {
		return errors.New("no pending future found for sub workflow completed event")
	}

	e.workflowState.removeCommandByEventID(event.EventID)

	if a.Error != "" {
		f.Set(nil, errors.New(a.Error))
	} else {
		f.Set(a.Result, nil)
	}

	return e.workflow.Continue(e.workflowCtx)
}

func (e *executor) handleSignalReceived(event history.Event, a *history.SignalReceivedAttributes) error {
	// Send signal to workflow channel
	sc := e.workflowState.getSignalChannel(a.Name)
	sc.SendNonblocking(e.workflowCtx, a.Arg)

	e.workflowState.removeCommandByEventID(event.EventID)

	return e.workflow.Continue(e.workflowCtx)
}

func (e *executor) handleSideEffectResult(event history.Event, a *history.SideEffectResultAttributes) error {
	f, ok := e.workflowState.pendingFutures[event.EventID]
	if !ok {
		return errors.New("no pending future found for side effect result event")
	}

	f.Set(a.Result, nil)

	return e.workflow.Continue(e.workflowCtx)
}

func (e *executor) workflowCompleted(result payload.Payload, err error) error {
	eventId := e.workflowState.eventID
	e.workflowState.eventID++

	cmd := command.NewCompleteWorkflowCommand(eventId, result, err)
	e.workflowState.addCommand(&cmd)

	return nil
}

func (e *executor) processCommands(ctx context.Context, t *task.Workflow) ([]history.Event, []core.WorkflowEvent, error) {
	instance := t.WorkflowInstance
	commands := e.workflowState.commands

	newEvents := make([]history.Event, 0)
	workflowEvents := make([]core.WorkflowEvent, 0)

	for _, c := range commands {
		// TODO: Move to state machine?
		// Mark this command as committed.
		c.State = command.CommandState_Committed

		switch c.Type {
		case command.CommandType_ScheduleActivityTask:
			a := c.Attr.(*command.ScheduleActivityTaskCommandAttr)

			newEvents = append(newEvents, history.NewHistoryEvent(
				history.EventType_ActivityScheduled,
				c.ID,
				&history.ActivityScheduledAttributes{
					Name:   a.Name,
					Inputs: a.Inputs,
				},
			))

		case command.CommandType_ScheduleSubWorkflow:
			a := c.Attr.(*command.ScheduleSubWorkflowCommandAttr)

			subWorkflowInstance := core.NewSubWorkflowInstance(a.InstanceID, uuid.NewString(), instance, c.ID)

			newEvents = append(newEvents, history.NewHistoryEvent(
				history.EventType_SubWorkflowScheduled,
				c.ID,
				&history.SubWorkflowScheduledAttributes{
					InstanceID: subWorkflowInstance.GetInstanceID(),
					Name:       a.Name,
					Inputs:     a.Inputs,
				},
			))

			// Send message to new workflow instance
			workflowEvents = append(workflowEvents, core.WorkflowEvent{
				WorkflowInstance: subWorkflowInstance,
				HistoryEvent: history.NewHistoryEvent(
					history.EventType_WorkflowExecutionStarted,
					c.ID,
					&history.ExecutionStartedAttributes{
						Name:   a.Name,
						Inputs: a.Inputs,
					},
				),
			})

		case command.CommandType_SideEffect:
			a := c.Attr.(*command.SideEffectCommandAttr)
			newEvents = append(newEvents, history.NewHistoryEvent(
				history.EventType_SideEffectResult,
				c.ID,
				&history.SideEffectResultAttributes{
					Result: a.Result,
				},
			))

		case command.CommandType_ScheduleTimer:
			a := c.Attr.(*command.ScheduleTimerCommandAttr)

			newEvents = append(newEvents, history.NewHistoryEvent(
				history.EventType_TimerScheduled,
				c.ID,
				&history.TimerScheduledAttributes{
					At: a.At,
				},
			))

			// Create timer_fired event which will become visible in the future
			workflowEvents = append(workflowEvents, core.WorkflowEvent{
				WorkflowInstance: instance,
				HistoryEvent: history.NewFutureHistoryEvent(
					history.EventType_TimerFired,
					c.ID,
					&history.TimerFiredAttributes{
						At: a.At,
					},
					a.At,
				)},
			)

		case command.CommandType_CompleteWorkflow:
			a := c.Attr.(*command.CompleteWorkflowCommandAttr)

			newEvents = append(newEvents, history.NewHistoryEvent(
				history.EventType_WorkflowExecutionFinished,
				c.ID,
				&history.ExecutionCompletedAttributes{
					Result: a.Result,
					Error:  a.Error,
				},
			))

			if instance.SubWorkflow() {
				// Send completion message back to parent workflow instance
				workflowEvents = append(workflowEvents, core.WorkflowEvent{
					WorkflowInstance: instance.ParentInstance(),
					HistoryEvent: history.NewHistoryEvent(
						history.EventType_SubWorkflowCompleted,
						instance.ParentEventID(), // Ensure the message gets sent back to the parent workflow with the right eventID
						&history.SubWorkflowCompletedAttributes{
							Result: a.Result,
							Error:  a.Error,
						},
					),
				})
			}

		default:
			return nil, nil, fmt.Errorf("unknown command type: %v", c.Type)
		}
	}

	return newEvents, workflowEvents, nil
}
