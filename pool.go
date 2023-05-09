package gotcc

import (
	"sync"

	"github.com/panjf2000/ants/v2"
)

type GoroutinePool interface {
	Go(task func()) error
}

func (m *TCController) PoolRun(pool GoroutinePool) (map[string]interface{}, error) {
	if len(m.termination.dependency) == 0 {
		return nil, ErrNoTermination{}
	}
	taskorder, noloop := m.analyzeDependency()
	if !noloop {
		return nil, ErrLoopDependency{}
	}
	sortedId, canSort := m.sortExecutor(taskorder)
	if !canSort {
		return nil, ErrPoolUnsupport{}
	}

	defer m.reset()

	wg := sync.WaitGroup{}
lauchLoop:
	for _, taskid := range sortedId {
		e := m.executors[taskid]
		wg.Add(1)
		err := pool.Go(func() {
			defer wg.Done()
			args := map[string]interface{}{"BIND": e.bindArgs, "CANCEL": m.cancelCtx, "NAME": e.name}

			for !e.dependencyExpr.f() {
				// wait until dep ok
				select {
				case <-m.cancelCtx.Done():
					return
				case msg := <-e.messageBuffer:
					e.MarkDependency(msg.senderId, true)
					args[msg.senderName] = msg.value
				}
			}

			outMsg := message{senderId: e.id, senderName: e.name}
			result, err := e.task(args)
			if err != nil {
				switch err := err.(type) {
				// case ErrSilentFail:
				// 	m.errorMsgs.Append(NewErrorMessage(e.Name, err))
				case ErrCancelled:
					m.cancelled.append(NewStateMessage(e.name, err.State))
				default:
					m.errorMsgs.append(NewErrorMessage(e.name, err))
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
		})
		if err != nil {
			return nil, err
		}
		select {
		case <-m.cancelCtx.Done():
			break lauchLoop
		default:
		}
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
			t.MarkDependency(msg.senderId, true)
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

// Default coroutine pool: actually not a coroutine pool but only launch new goroutines.
type DefaultNoPool struct{}

func (DefaultNoPool) Go(task func()) error {
	go task()
	return nil
}

// Default coroutine pool: base on ants pool.
type DefaultPool struct {
	pool *ants.Pool
}

func NewDefaultPool(size int) DefaultPool {
	if size <= 0 {
		pool, _ := ants.NewPool(size)
		return DefaultPool{pool}
	} else {
		pool, _ := ants.NewPool(size, ants.WithPreAlloc(true))
		return DefaultPool{pool}
	}
}

func (p DefaultPool) Go(task func()) error {
	return p.pool.Submit(task)
}

func (p DefaultPool) Close() {
	p.pool.Release()
}
