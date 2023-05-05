package gotcc

import (
	"context"
	"strings"
	"sync"
)

type TCController struct {
	executors map[uint32]*Executor

	cancelCtx  context.Context
	cancelFunc context.CancelFunc

	termination *Executor

	cancelled CancelList
	errorMsgs ErrorList
	undoStack UndoStack
}

func NewTCController() *TCController {
	ctx, cf := context.WithCancel(context.Background())
	return &TCController{
		executors:   map[uint32]*Executor{},
		cancelCtx:   ctx,
		cancelFunc:  cf,
		termination: newExecutor("TERMINATION", nil, nil),
		errorMsgs:   ErrorList{},
		undoStack:   UndoStack{},
	}
}

func (m *TCController) AddTask(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	e := newExecutor(name, f, args)
	m.executors[e.Id] = e
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

func (m *TCController) loopDependency() bool {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[uint32]int{}
	var dfs func(curr uint32) bool
	dfs = func(curr uint32) bool {
		color[curr] = gray
		for neighbor := range m.executors[curr].dependency {
			switch color[neighbor] {
			case white:
				if dfs(neighbor) {
					return true
				}
			case gray:
				return true
			case black:
			}
		}
		color[curr] = black
		return false
	}

	for taskid := range m.executors {
		if dfs(taskid) {
			return true
		}
	}
	return false
}

type ErrNoTermination struct{}

func (ErrNoTermination) Error() string {
	return "Error: No termination condition has been set!"
}

type ErrAborted struct {
	TaskErrors *ErrorList
	UndoErrors *ErrorList
	Cancelled  *CancelList
}

func (e ErrAborted) Error() string {
	var sb strings.Builder
	sb.WriteString("\n[x] TaskErrors:\n")
	sb.WriteString(e.TaskErrors.String())
	sb.WriteString("\n[-] UndoErrors:\n")
	sb.WriteString(e.UndoErrors.String())
	sb.WriteString("\n[/] Cancelled:\n")
	sb.WriteString(e.Cancelled.String())
	return sb.String()
}

type ErrCancelled struct {
	State State
}

func (e ErrCancelled) Error() string {
	return "Error: Task is canncelled due to other errors."
}

type ErrLoopDependency struct {
	State State
}

func (e ErrLoopDependency) Error() string {
	return "Error: Tasks has loop dependency."
}

func (m *TCController) RunTask() (map[string]interface{}, error) {
	if len(m.termination.dependency) == 0 {
		return nil, ErrNoTermination{}
	}
	if m.loopDependency() {
		return nil, ErrLoopDependency{}
	}
	wg := sync.WaitGroup{}
	wg.Add(len(m.executors))
	for taskid := range m.executors {
		e := m.executors[taskid]
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
				if ec, isCancelled := err.(ErrCancelled); isCancelled {
					m.cancelled.Append(NewStateMessage(e.Name, ec.State))
				} else {
					m.errorMsgs.Append(NewErrorMessage(e.Name, err))
				}
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
			Cancelled:  &m.cancelled,
		}

		// do the rollback
		returnErr.UndoErrors = m.undoStack.UndoAll(&m.errorMsgs, &m.cancelled)

		return nil, returnErr
	}
}
