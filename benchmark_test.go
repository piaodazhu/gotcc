package gotcc

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkBatchRunSerialized(b *testing.B) {

}

func BenchmarkBatchRunManyToOne(b *testing.B) {

}

func BenchmarkBatchRunManyToMany(b *testing.B) {

}

func BenchmarkPoolRunSerialized(b *testing.B) {

}

func BenchmarkPoolRunManyToOne(b *testing.B) {

}

func BenchmarkPoolRunManyToMany(b *testing.B) {

}

// From tasktree.go by @lesismal :)
type Node struct {
	F     func() error
	Nodes []*Node

	selfCnt    int32
	parentCnt  *int32
	parentWg   *sync.WaitGroup
	parentFunc func() error
	rollback   func(error)
}

func (node *Node) doNodes(parentWg *sync.WaitGroup, parentCnt *int32, parentFunc func() error, rollback func(error)) {
	node.selfCnt = int32(len(node.Nodes))
	node.parentWg = parentWg
	node.parentCnt = parentCnt
	node.parentFunc = parentFunc
	node.rollback = rollback
	if node.selfCnt > 0 {
		for _, v := range node.Nodes {
			v.doNodes(nil, &node.selfCnt, node.doSelf, rollback)
		}
		return
	}
	go node.doSelf()
}

func (node *Node) doSelf() error {
	defer func() {
		if node.parentWg != nil {
			node.parentWg.Done()
		} else if atomic.AddInt32(node.parentCnt, -1) == 0 && node.parentFunc != nil {
			if err := node.parentFunc(); err != nil && node.rollback != nil {
				node.rollback(err)
			}
		}
	}()
	if err := node.F(); err != nil && node.rollback != nil {
		node.rollback(err)
	}
	return nil
}

type TaskTree struct{}

func (t *TaskTree) Go(root *Node, rollback func(error)) func() <-chan error {
	wg := &sync.WaitGroup{}
	chErr := make(chan error, 1)
	wg.Add(1)
	waitFunc := func() <-chan error {
		wg.Wait()
		if len(chErr) == 0 {
			chErr <- nil
		}
		return chErr
	}
	var n int32
	var rollbackCalled int32
	root.doNodes(wg, &n, nil, func(err error) {
		if atomic.AddInt32(&rollbackCalled, 1) == 1 {
			rollback(err)
			chErr <- err
		}
	})
	return waitFunc
}

func (t *TaskTree) GoN(nodes []*Node, rollback func(error)) func() <-chan error {
	wg := &sync.WaitGroup{}
	chErr := make(chan error, 1)
	waitFunc := func() <-chan error {
		wg.Wait()
		if len(chErr) == 0 {
			chErr <- nil
		}
		return chErr
	}
	if len(nodes) == 0 {
		return waitFunc
	}
	wg.Add(len(nodes))
	var n int32
	var rollbackCalled int32
	for _, v := range nodes {
		v.doNodes(wg, &n, nil, func(err error) {
			if atomic.AddInt32(&rollbackCalled, 1) == 1 {
				rollback(err)
				chErr <- err
			}
		})
	}
	return waitFunc
}

func BenchmarkTaskTreeSerialized(b *testing.B) {

}

func BenchmarkTaskTreeManyToOne(b *testing.B) {

}

func BenchmarkTaskTreeManyToMany(b *testing.B) {

}
