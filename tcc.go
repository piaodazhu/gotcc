package gotcc

import (
	"context"
	"strings"
	"sync"
)

type TCController struct {
	executors []*Executor

	cancelCtx  context.Context
	cancelFunc context.CancelFunc

	termination *Executor

	errorMsgs ErrorList
	undoStack UndoStack
}

func NewTCController() *TCController {
	ctx, cf := context.WithCancel(context.Background())
	return &TCController{
		executors:   []*Executor{},
		cancelCtx:   ctx,
		cancelFunc:  cf,
		termination: newExecutor("TERMINATION", nil, nil),
		errorMsgs:   ErrorList{},
		undoStack:   UndoStack{},
	}
}

func (m *TCController) AddTask(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	e := newExecutor(name, f, args)
	m.executors = append(m.executors, e)
	return e
}

func (m *TCController) SetTermination(Expr DependencyExpression) {
	m.termination.SetDependency(Expr)
}

func (m *TCController) NewTerminationExpr(d *Executor) DependencyExpression {
	if _, exists := m.termination.dependency[d.Id]; !exists {
		m.termination.dependency[d.Id] = false
		m.termination.messageBuffer = make(chan Message, cap(m.termination.messageBuffer)+1)
		d.subscribers = append(d.subscribers, &m.termination.messageBuffer)
	}
	return newDependencyExpr(m.termination.dependency, d.Id)
}

type ErrNoTermination struct{}

func (ErrNoTermination) Error() string { return "Error: No termination condition has been set!" }

type ErrAborted struct {
	TaskErrors *ErrorList
	UndoErrors *ErrorList
}

func (m ErrAborted) Error() string {
	var sb strings.Builder
	sb.WriteString("\n[x] TaskErrors:\n")
	sb.WriteString(m.TaskErrors.String())
	sb.WriteString("\n[-] UndoErrors:\n")
	sb.WriteString(m.UndoErrors.String())
	return sb.String()
}

func (m *TCController) RunTask() (map[string]interface{}, error) {
	if len(m.termination.dependency) == 0 {
		return nil, ErrNoTermination{}
	}
	wg := sync.WaitGroup{}
	wg.Add(len(m.executors))
	for i := range m.executors {
		e := m.executors[i]
		go func() {
			defer wg.Done()
			args := map[string]interface{}{"BIND": e.BindArgs, "CANCEL": m.cancelCtx}

			for !e.dependencyExpr() {
				// wait until dep ok
				select {
				case <-m.cancelCtx.Done():
					return
				case msg := <-e.messageBuffer:
					e.MarkDependency(msg.Sender, true)
					args[msg.SenderName] = msg.Value
				}
			}

			outMsg := Message{Sender: e.Id, SenderName: e.Name}
			result, err := e.Task(args)
			if err != nil {
				m.errorMsgs.Append(NewErrorMessage(e.Name, err))

				// abort all task...
				m.cancelFunc()
				return
			} else {
				outMsg.Value = result

				// add to finished stack...
				m.undoStack.Push(NewUndoFunc(e.Name, e.UndoSkipError, e.Undo, args))
			}

			for _, subscriber := range e.subscribers {
				*subscriber <- outMsg
			}
		}()
	}

	// wait termination
	t := m.termination
	Results := map[string]interface{}{}
	Aborted := false

waitLoop:
	for !t.dependencyExpr() {
		select {
		case <-m.cancelCtx.Done():
			// aborted
			Aborted = true
			break waitLoop
		case msg := <-t.messageBuffer:
			t.MarkDependency(msg.Sender, true)
			Results[msg.SenderName] = msg.Value
		}
	}
	if !Aborted {
		// all done!
		m.cancelFunc()
		wg.Wait()
		return Results, nil
	} else {
		// aborted because of some error
		wg.Wait()
		returnErr := ErrAborted{
			TaskErrors: &m.errorMsgs,
		}

		// do the rollback
		returnErr.UndoErrors = m.undoStack.UndoAll(&m.errorMsgs)

		return nil, returnErr
	}
}
