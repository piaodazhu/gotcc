package gotcc

import (
	"context"
	"fmt"
	"sort"
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

func (m *TCController) Run() (map[string]interface{}, error) {
	if len(m.termination.dependency) == 0 {
		return nil, ErrNoTermination{}
	}
	if m.loopDependency() {
		return nil, ErrLoopDependency{}
	}

	defer m.Reset()

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
				switch err := err.(type) {
				case ErrSilentFail:
					m.errorMsgs.Append(NewErrorMessage(e.Name, err))
				case ErrCancelled:
					m.cancelled.Append(NewStateMessage(e.Name, err.State))
				default:
					m.errorMsgs.Append(NewErrorMessage(e.Name, err))
					m.cancelFunc()
				}
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
			TaskErrors: m.errorMsgs.Items,
			Cancelled:  m.cancelled.Items,
		}

		// do the rollback
		returnErr.UndoErrors = m.undoStack.UndoAll(&m.errorMsgs, &m.cancelled).Items
		// fmt.Println(returnErr.Error())
		return nil, returnErr
	}
}

func (m *TCController) Reset() {
	m.cancelCtx, m.cancelFunc = context.WithCancel(context.Background())
	m.cancelled.Reset()
	m.errorMsgs.Reset()
	m.undoStack.Reset()
	for _, e := range m.executors {
		e.messageBuffer = make(chan Message, cap(e.messageBuffer))
		for dep := range e.dependency {
			e.dependency[dep] = false
		}
	}
	for term := range m.termination.dependency {
		m.termination.dependency[term] = false
	}
}

func (m *TCController) String() string {
	var sb strings.Builder
	sb.WriteString("\ncanncelled list:\n")
	sb.WriteString(m.cancelled.String())
	sb.WriteString("errmessage list:\n")
	sb.WriteString(m.errorMsgs.String())
	sb.WriteString("tasks:\n")
	ids := make([]uint32, 0, len(m.executors))
	for id := range m.executors {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		e := m.executors[id]
		sb.WriteString(e.Name)
		sb.WriteString(fmt.Sprintf("[msgBuffer cap=%d]: (", cap(e.messageBuffer)))

		depids := make([]uint32, 0, len(e.dependency))
		for depid := range e.dependency {
			depids = append(depids, depid)
		}
		sort.Slice(depids, func(i, j int) bool { return depids[i] < depids[j] })
		for _, depid := range depids {
			sb.WriteString(m.executors[depid].Name)
			sb.WriteString(", ")
		}
		sb.WriteString(")\n")
	}
	sb.WriteString(fmt.Sprintf("@termination[msgBuffer cap=%d]: (", cap(m.termination.messageBuffer)))
	termids := make([]uint32, 0, len(m.termination.dependency))
	for termid := range m.termination.dependency {
		termids = append(termids, termid)
	}
	for _, termid := range termids {
		sb.WriteString(m.executors[termid].Name)
		sb.WriteString(", ")
	}
	sb.WriteString(")\n")

	return sb.String()
}

type ErrNoTermination struct{}

func (ErrNoTermination) Error() string {
	return "Error: No termination condition has been set!"
}

type ErrLoopDependency struct {
	State State
}

func (e ErrLoopDependency) Error() string {
	return "Error: Tasks has loop dependency."
}

type ErrAborted struct {
	TaskErrors []*ErrorMessage
	UndoErrors []*ErrorMessage
	Cancelled  []*StateMessage
}

func (e ErrAborted) Error() string {
	var sb strings.Builder
	sb.WriteString("\n[x] TaskErrors:\n")
	sb.WriteString((&ErrorList{Items: e.TaskErrors}).String())
	sb.WriteString("[-] UndoErrors:\n")
	sb.WriteString((&ErrorList{Items: e.UndoErrors}).String())
	sb.WriteString("[/] Cancelled:\n")
	sb.WriteString((&CancelList{Items: e.Cancelled}).String())
	return sb.String()
}
