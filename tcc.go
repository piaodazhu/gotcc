package gotcc

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Task Concurrency Controller
type TCController struct {
	executors map[uint32]*Executor

	cancelCtx  context.Context
	cancelFunc context.CancelFunc

	termination *Executor

	cancelled cancelList
	errorMsgs errorLisk
	undoStack undoStack
}

// Create an empty task concurrency controller
func NewTCController() *TCController {
	ctx, cf := context.WithCancel(context.Background())
	return &TCController{
		executors:   map[uint32]*Executor{},
		cancelCtx:   ctx,
		cancelFunc:  cf,
		termination: newExecutor("TERMINATION", nil, nil),
		cancelled:   cancelList{},
		errorMsgs:   errorLisk{},
		undoStack:   undoStack{},
	}
}

// Add a task to the controller. `name` is a user-defined string identifier of the task.
// `f` is the task function. `args` is arguments bind with the task, which can be obtained
// inside the task function from args["BIND"].
func (m *TCController) AddTask(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	e := newExecutor(name, f, args)
	m.executors[e.id] = e
	return e
}

// Set termination condition for the controller. `Expr` is a dependency expression.
func (m *TCController) SetTermination(Expr DependencyExpression) {
	m.termination.SetDependency(Expr)
}

// Get termination condition of the controller.
func (m *TCController) TerminationExpr() DependencyExpression {
	return m.termination.dependencyExpr
}

// Create a termination dependency expression for the controller.
// It means the execution termination may depend on task `d`.
func (m *TCController) NewTerminationExpr(d *Executor) DependencyExpression {
	if _, exists := m.termination.dependency[d.id]; !exists {
		m.termination.dependency[d.id] = false
		m.termination.messageBuffer = make(chan message, cap(m.termination.messageBuffer)+1)
		d.subscribers = append(d.subscribers, &m.termination.messageBuffer)
	}
	return newDependencyExpr(m.termination.dependency, d.id)
}

// Run the execution. If success, return a map[name]value, where names are task
// of termination dependent tasks and values are their return value.
// If failed, return ErrNoTermination, ErrLoopDependency or ErrAborted
func (m *TCController) BatchRun() (map[string]interface{}, error) {
	if len(m.termination.dependency) == 0 {
		return nil, ErrNoTermination{}
	}
	if _, noloop := m.analyzeDependency(); !noloop {
		return nil, ErrLoopDependency{}
	}

	defer m.reset()

	wg := sync.WaitGroup{}
	wg.Add(len(m.executors))
	for taskid := range m.executors {
		e := m.executors[taskid]
		go func() {
			defer wg.Done()
			args := map[string]interface{}{"BIND": e.bindArgs, "CANCEL": m.cancelCtx, "NAME": e.name}

			for !e.dependencyExpr.f() {
				// wait until dep ok
				select {
				case <-m.cancelCtx.Done():
					return
				case msg := <-e.messageBuffer:
					e.markDependency(msg.senderId, true)
					args[msg.senderName] = msg.value
				}
			}

			outMsg := message{senderId: e.id, senderName: e.name}
			result, err := e.task(args)
			if err != nil {
				switch err := err.(type) {
				case ErrSilentFail:
					m.errorMsgs.append(newErrorMessage(e.name, err))
				case ErrCancelled:
					m.cancelled.append(newStateMessage(e.name, err.State))
				default:
					m.errorMsgs.append(newErrorMessage(e.name, err))
					m.cancelFunc()
				}
				return
			} else {
				outMsg.value = result

				// add to finished stack...
				m.undoStack.push(newUndoFunc(e.name, e.undoSkipError, e.undo, args))
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
	for !t.dependencyExpr.f() {
		select {
		case <-m.cancelCtx.Done():
			// aborted
			Aborted = true
			break waitLoop
		case msg := <-t.messageBuffer:
			t.markDependency(msg.senderId, true)
			Results[msg.senderName] = msg.value
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
			TaskErrors: m.errorMsgs.items,
			Cancelled:  m.cancelled.items,
		}

		// do the rollback
		returnErr.UndoErrors = m.undoStack.undoAll(&m.errorMsgs, &m.cancelled).items
		// fmt.Println(returnErr.Error())
		return nil, returnErr
	}
}

func (m *TCController) reset() {
	m.cancelCtx, m.cancelFunc = context.WithCancel(context.Background())
	m.cancelled.reset()
	m.errorMsgs.reset()
	m.undoStack.reset()
	for _, e := range m.executors {
		e.messageBuffer = make(chan message, cap(e.messageBuffer))
		for dep := range e.dependency {
			e.dependency[dep] = false
		}
	}
	for term := range m.termination.dependency {
		m.termination.dependency[term] = false
	}
}

// The inner state of the controller
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
		sb.WriteString(e.name)
		sb.WriteString(fmt.Sprintf("[msgBuffer cap=%d]: (", cap(e.messageBuffer)))

		depids := make([]uint32, 0, len(e.dependency))
		for depid := range e.dependency {
			depids = append(depids, depid)
		}
		sort.Slice(depids, func(i, j int) bool { return depids[i] < depids[j] })
		for _, depid := range depids {
			sb.WriteString(m.executors[depid].name)
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
		sb.WriteString(m.executors[termid].name)
		sb.WriteString(", ")
	}
	sb.WriteString(")\n")

	return sb.String()
}
