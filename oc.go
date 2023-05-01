package gooc

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Manager struct {
	Executors []*Executor

	CancelCtx  context.Context
	CancelFunc context.CancelFunc

	Termination *Executor

	ErrorMsgs []ErrorMessage
}

func NewOCManager() *Manager {
	ctx, cf := context.WithCancel(context.Background())
	m := &Manager{
		Executors:   []*Executor{},
		CancelCtx:   ctx,
		CancelFunc:  cf,
		Termination: newExecutor("TERMINATION", nil, nil),
		ErrorMsgs:   []ErrorMessage{},
	}
	m.Termination.Manager = m
	return m
}

func (m *Manager) AddTask(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	e := newExecutor(name, f, args)
	e.Manager = m

	m.Executors = append(m.Executors, e)
	return e
}

func (m *Manager) SetTermination(Expr DependencyExpression) {
	m.Termination.SetDependency(Expr)
}

func (m *Manager) NewTerminationExpr(d *Executor) DependencyExpression {
	if _, exists := m.Termination.Dependency[d.Id]; !exists {
		m.Termination.Dependency[d.Id] = false
		m.Termination.MessageBuffer = make(chan Message, cap(m.Termination.MessageBuffer)+1)
		d.Subscribers = append(d.Subscribers, m.Termination.MessageBuffer)
	}
	return newDependencyExpr(m.Termination.Dependency, d.Id)
}

type ErrNoTermination struct{}

func (ErrNoTermination) Error() string { return "Error: No termination condition has been set!" }

type ErrAborted struct {
	Reasons []ErrorMessage
}

func (m ErrAborted) Error() string {
	var res strings.Builder
	res.WriteString("Error: Tasks aborted! Reasons: ")
	for i := range m.Reasons {
		res.WriteString(m.Reasons[i].SenderName)
		res.WriteString(": ")
		res.WriteString(m.Reasons[i].Value.Error())
		res.WriteString("; ")
	}
	return res.String()
}

func (m *Manager) RunTask() (map[string]interface{}, error) {
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

				// abort all task...
				m.CancelFunc()
				return
			} else {
				outMsg.Value = result

				// add to finished stack...
			}

			for _, subscriber := range e.Subscribers {
				select {
				case <-m.CancelCtx.Done():
				case subscriber <- outMsg:
				}
			}
		}()
	}
	// wait something to stop

	t := m.Termination
	Results := map[string]interface{}{}
	Aborted := false

waitLoop:
	for !t.CalcDependency() {
		// wait until dep ok
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
		// some thing wrong
		// rollback
		wg.Wait()
		returnErr := ErrAborted{m.ErrorMsgs}

		return nil, returnErr
	}
}
