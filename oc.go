package main

import (
	"context"
	"fmt"
	"time"
)

type Manager struct {
	Executors []Executor
	// MailBoxes map[uint32]chan interface{}
	// StartSet map[uint32]*Executor
	CancelCtx  context.Context
	CancelFunc context.CancelFunc
}

func NewOCManager() *Manager {
	ctx, cf := context.WithCancel(context.Background())
	return &Manager{
		Executors:  []Executor{},
		CancelCtx:  ctx,
		CancelFunc: cf,
	}
}

func (m *Manager) AddTask(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	e := Executor{
		Id:             uint32(time.Now().UnixNano() & 0xffffffff),
		Name:           name,
		Dependency:     map[uint32]bool{},
		DependencyExpr: NoDependency,
		MessageBuffer:  make(chan Message),
		Subscribers:    make([]chan Message, 0),

		Args:    args,
		Task:    f,
		Manager: m,
	}
	m.Executors = append(m.Executors, e)
	// m.StartSet[e.Id] = &e
	return &e
}

func (m *Manager) RunTask(DefaultArg interface{}) {
	for i := range m.Executors {
		e := m.Executors[i]
		go func() {
			args := map[string]interface{}{"DEFAULT": DefaultArg}

			nDep := len(e.Dependency)
			nRcv := 0

			for nRcv < nDep && !e.DependencyExpr(e.Dependency) {
				// wait until dep ok
				var msg Message
				select {
				case msg = <-e.MessageBuffer:
				case <-m.CancelCtx.Done():
					// cancelled
				}

				nRcv++
				if !msg.Finish {
					// todo
					e.Dependency[msg.Sender] = false
				} else {
					e.Dependency[msg.Sender] = true
				}
				args[msg.SenderName] = msg.Value
			}

			// a. all received but should not run this

			// b. not all received but should run this

			// how about pass? global stop?
			output := Message{Sender: e.Id, SenderName: e.Name}
			result, err := e.Task(args)
			if err != nil {
				fmt.Println(err)
				output.Finish = false
				output.Value = nil
			} else {
				output.Finish = true
				output.Value = result

				// add to finished stack...
			}

			for _, subscriber := range e.Subscribers {
				select {
				case <-m.CancelCtx.Done():
				default:
					subscriber <- output
				}
			}
		}()
	}
	// wait something to stop

}
