package gotcc

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type TCController struct {
	Executors []*Executor

	CancelCtx  context.Context
	CancelFunc context.CancelFunc

	Termination *Executor

	ErrorMsgs ErrorList

	UndoStack UndoStack
}

func NewTCController() *TCController {
	ctx, cf := context.WithCancel(context.Background())
	m := &TCController{
		Executors:   []*Executor{},
		CancelCtx:   ctx,
		CancelFunc:  cf,
		Termination: newExecutor("TERMINATION", nil, nil),
		ErrorMsgs:   ErrorList{},
		UndoStack:   UndoStack{},
	}
	m.Termination.TCController = m
	return m
}

func (m *TCController) AddTask(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	e := newExecutor(name, f, args)
	e.TCController = m
	m.Executors = append(m.Executors, e)
	return e
}

func (m *TCController) SetTermination(Expr DependencyExpression) {
	m.Termination.SetDependency(Expr)
}

func (m *TCController) NewTerminationExpr(d *Executor) DependencyExpression {
	if _, exists := m.Termination.Dependency[d.Id]; !exists {
		m.Termination.Dependency[d.Id] = false
		m.Termination.MessageBuffer = make(chan Message, cap(m.Termination.MessageBuffer)+1)
		d.Subscribers = append(d.Subscribers, &m.Termination.MessageBuffer)
	}
	return newDependencyExpr(m.Termination.Dependency, d.Id)
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
	if len(m.Termination.Dependency) == 0 {
		return nil, ErrNoTermination{}
	}
	wg := sync.WaitGroup{}
	wg.Add(len(m.Executors))
	for i := range m.Executors {
		e := m.Executors[i]
		go func() {
			defer wg.Done()
			args := map[string]interface{}{"BIND": e.BindArgs}

			for !e.CalcDependency() {
				// wait until dep ok
				select {
				case <-m.CancelCtx.Done():
					return
				case msg := <-e.MessageBuffer:
					e.Dependency[msg.Sender] = true
					args[msg.SenderName] = msg.Value
				}
			}

			outMsg := Message{Sender: e.Id, SenderName: e.Name}
			result, err := e.Task(args)
			if err != nil {
				fmt.Println(err)

				m.ErrorMsgs.Append(NewErrorMessage(e.Name, err))

				// abort all task...
				m.CancelFunc()
				return
			} else {
				outMsg.Value = result

				// add to finished stack...
				m.UndoStack.Push(NewUndoFunc(e.Name, e.UndoSkipError, e.Undo, args))
			}

			for _, subscriber := range e.Subscribers {
				select {
				case <-m.CancelCtx.Done():
					return
				case *subscriber <- outMsg:
				}
			}
		}()
	}

	// wait termination
	t := m.Termination
	Results := map[string]interface{}{}
	Aborted := false

waitLoop:
	for !t.CalcDependency() {
		select {
		case <-m.CancelCtx.Done():
			// aborted
			Aborted = true
			break waitLoop
		case msg := <-t.MessageBuffer:
			t.Dependency[msg.Sender] = true
			Results[msg.SenderName] = msg.Value
		}
	}
	if !Aborted {
		// all done!
		m.CancelFunc()
		wg.Wait()
		return Results, nil
	} else {
		// aborted because of some error
		wg.Wait()
		returnErr := ErrAborted{
			TaskErrors: &m.ErrorMsgs,
		}

		// do the rollback
		returnErr.UndoErrors = m.UndoStack.UndoAll()

		return nil, returnErr
	}
}
