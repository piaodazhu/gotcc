package gotcc

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	ITERNUM = 0
	SLEEPTIME = 0
)

func TaskForBench(args map[string]interface{}) (interface{}, error) {
	time.Sleep(time.Microsecond*SLEEPTIME)
	i := 0
	for i < ITERNUM {
		i++
	}
	return nil, nil
}

func BenchmarkBatchRunSerialized10(b *testing.B) {
	controller := NewTCController()
	last := controller.AddTask("first", TaskForBench, 1)
	for i := 0; i < 9; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		task.SetDependency(task.NewDependencyExpr(last))
		last = task
	}
	controller.SetTermination(controller.NewTerminationExpr(last))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}
func BenchmarkBatchRunSerialized100(b *testing.B) {
	controller := NewTCController()
	last := controller.AddTask("first", TaskForBench, 1)
	for i := 0; i < 99; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		task.SetDependency(task.NewDependencyExpr(last))
		last = task
	}
	controller.SetTermination(controller.NewTerminationExpr(last))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}
func BenchmarkBatchRunSerialized1000(b *testing.B) {
	controller := NewTCController()
	last := controller.AddTask("first", TaskForBench, 1)
	for i := 0; i < 999; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		task.SetDependency(task.NewDependencyExpr(last))
		last = task
	}
	controller.SetTermination(controller.NewTerminationExpr(last))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkBatchRunManyToOne10(b *testing.B) {
	controller := NewTCController()
	end := controller.AddTask("last", TaskForBench, 1)
	for i := 0; i < 9; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkBatchRunManyToOne100(b *testing.B) {
	controller := NewTCController()
	end := controller.AddTask("last", TaskForBench, 1)
	for i := 0; i < 99; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkBatchRunManyToOne1000(b *testing.B) {
	controller := NewTCController()
	end := controller.AddTask("last", TaskForBench, 1)
	for i := 0; i < 999; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}
func BenchmarkBatchRunManyToMany10(b *testing.B) {
	controller := NewTCController()
	for j := 0; j < 5; j++ {
		l2node := controller.AddTask("l2node", TaskForBench, 1)
		for i := 0; i < 5; i++ {
			task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
			l2node.SetDependency(MakeAndExpr(l2node.DependencyExpr(), l2node.NewDependencyExpr(task)))
		}
		controller.SetTermination(MakeAndExpr(controller.TerminationExpr(), controller.NewTerminationExpr(l2node)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkBatchRunManyToMany100(b *testing.B) {
	controller := NewTCController()
	for j := 0; j < 50; j++ {
		l2node := controller.AddTask("l2node", TaskForBench, 1)
		for i := 0; i < 50; i++ {
			task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
			l2node.SetDependency(MakeAndExpr(l2node.DependencyExpr(), l2node.NewDependencyExpr(task)))
		}
		controller.SetTermination(MakeAndExpr(controller.TerminationExpr(), controller.NewTerminationExpr(l2node)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkBatchRunBinaryTree10(b *testing.B) {
	controller := NewTCController()
	root := controller.AddTask("root", TaskForBench, 1)
	controller.SetTermination(controller.NewTerminationExpr(root))
	queue := []*Executor{root}
	for i := 0; i < 3; i++ {
		tmp := []*Executor{}
		for _, e := range queue {
			left := controller.AddTask("left", TaskForBench, 1)
			right := controller.AddTask("right", TaskForBench, 1)
			e.SetDependency(MakeAndExpr(e.NewDependencyExpr(left), e.NewDependencyExpr(right)))
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkBatchRunBinaryTree100(b *testing.B) {
	controller := NewTCController()
	root := controller.AddTask("root", TaskForBench, 1)
	controller.SetTermination(controller.NewTerminationExpr(root))
	queue := []*Executor{root}
	for i := 0; i < 6; i++ {
		tmp := []*Executor{}
		for _, e := range queue {
			left := controller.AddTask("left", TaskForBench, 1)
			right := controller.AddTask("right", TaskForBench, 1)
			e.SetDependency(MakeAndExpr(e.NewDependencyExpr(left), e.NewDependencyExpr(right)))
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkBatchRunBinaryTree1000(b *testing.B) {
	controller := NewTCController()
	root := controller.AddTask("root", TaskForBench, 1)
	controller.SetTermination(controller.NewTerminationExpr(root))
	queue := []*Executor{root}
	for i := 0; i < 9; i++ {
		tmp := []*Executor{}
		for _, e := range queue {
			left := controller.AddTask("left", TaskForBench, 1)
			right := controller.AddTask("right", TaskForBench, 1)
			e.SetDependency(MakeAndExpr(e.NewDependencyExpr(left), e.NewDependencyExpr(right)))
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.BatchRun()
	}
}

func BenchmarkPoolRunSerialized10(b *testing.B) {
	controller := NewTCController()
	last := controller.AddTask("first", TaskForBench, 1)
	for i := 0; i < 9; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		task.SetDependency(task.NewDependencyExpr(last))
		last = task
	}
	controller.SetTermination(controller.NewTerminationExpr(last))

	pool := NewDefaultPool(9)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunSerialized100(b *testing.B) {
	controller := NewTCController()
	last := controller.AddTask("first", TaskForBench, 1)
	for i := 0; i < 99; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		task.SetDependency(task.NewDependencyExpr(last))
		last = task
	}
	controller.SetTermination(controller.NewTerminationExpr(last))

	pool := NewDefaultPool(99)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunSerialized1000(b *testing.B) {
	controller := NewTCController()
	last := controller.AddTask("first", TaskForBench, 1)
	for i := 0; i < 999; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		task.SetDependency(task.NewDependencyExpr(last))
		last = task
	}
	controller.SetTermination(controller.NewTerminationExpr(last))

	pool := NewDefaultPool(999)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunManyToOne10(b *testing.B) {
	controller := NewTCController()
	end := controller.AddTask("last", TaskForBench, 1)
	for i := 0; i < 9; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))

	pool := NewDefaultPool(9)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunManyToOne100(b *testing.B) {
	controller := NewTCController()
	end := controller.AddTask("last", TaskForBench, 1)
	for i := 0; i < 99; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))

	pool := NewDefaultPool(99)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunManyToOne1000(b *testing.B) {
	controller := NewTCController()
	end := controller.AddTask("last", TaskForBench, 1)
	for i := 0; i < 999; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))

	pool := NewDefaultPool(999)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunManyToMany10(b *testing.B) {
	controller := NewTCController()
	for j := 0; j < 5; j++ {
		l2node := controller.AddTask("l2node", TaskForBench, 1)
		for i := 0; i < 5; i++ {
			task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
			l2node.SetDependency(MakeAndExpr(l2node.DependencyExpr(), l2node.NewDependencyExpr(task)))
		}
		controller.SetTermination(MakeAndExpr(controller.TerminationExpr(), controller.NewTerminationExpr(l2node)))
	}

	pool := NewDefaultPool(5)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunManyToMany100(b *testing.B) {
	controller := NewTCController()
	for j := 0; j < 50; j++ {
		l2node := controller.AddTask("l2node", TaskForBench, 1)
		for i := 0; i < 50; i++ {
			task := controller.AddTask(strconv.Itoa(i), TaskForBench, 1)
			l2node.SetDependency(MakeAndExpr(l2node.DependencyExpr(), l2node.NewDependencyExpr(task)))
		}
		controller.SetTermination(MakeAndExpr(controller.TerminationExpr(), controller.NewTerminationExpr(l2node)))
	}

	pool := NewDefaultPool(50)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunBinaryTree10(b *testing.B) {
	controller := NewTCController()
	root := controller.AddTask("root", TaskForBench, 1)
	controller.SetTermination(controller.NewTerminationExpr(root))
	queue := []*Executor{root}
	for i := 0; i < 3; i++ {
		tmp := []*Executor{}
		for _, e := range queue {
			left := controller.AddTask("left", TaskForBench, 1)
			right := controller.AddTask("right", TaskForBench, 1)
			e.SetDependency(MakeAndExpr(e.NewDependencyExpr(left), e.NewDependencyExpr(right)))
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	pool := NewDefaultPool(8)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunBinaryTree100(b *testing.B) {
	controller := NewTCController()
	root := controller.AddTask("root", TaskForBench, 1)
	controller.SetTermination(controller.NewTerminationExpr(root))
	queue := []*Executor{root}
	for i := 0; i < 6; i++ {
		tmp := []*Executor{}
		for _, e := range queue {
			left := controller.AddTask("left", TaskForBench, 1)
			right := controller.AddTask("right", TaskForBench, 1)
			e.SetDependency(MakeAndExpr(e.NewDependencyExpr(left), e.NewDependencyExpr(right)))
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	pool := NewDefaultPool(64)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
}

func BenchmarkPoolRunBinaryTree1000(b *testing.B) {
	controller := NewTCController()
	root := controller.AddTask("root", TaskForBench, 1)
	controller.SetTermination(controller.NewTerminationExpr(root))
	queue := []*Executor{root}
	for i := 0; i < 9; i++ {
		tmp := []*Executor{}
		for _, e := range queue {
			left := controller.AddTask("left", TaskForBench, 1)
			right := controller.AddTask("right", TaskForBench, 1)
			e.SetDependency(MakeAndExpr(e.NewDependencyExpr(left), e.NewDependencyExpr(right)))
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	pool := NewDefaultPool(512)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		controller.PoolRun(pool)
	}
	b.StopTimer()
	pool.Close()
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

func TaskForComparison() error {
	time.Sleep(time.Microsecond*SLEEPTIME)
	i := 0
	for i < ITERNUM {
		i++
	}
	return nil
}

func BenchmarkTaskTreeSerialized10(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	last := root
	for i := 0; i < 9; i++ {
		node := &Node{F: TaskForComparison}
		last.Nodes = append(last.Nodes, node)
		last = node
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}
func BenchmarkTaskTreeSerialized100(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	last := root
	for i := 0; i < 99; i++ {
		node := &Node{F: TaskForComparison}
		last.Nodes = append(last.Nodes, node)
		last = node
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}
func BenchmarkTaskTreeSerialized1000(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	last := root
	for i := 0; i < 999; i++ {
		node := &Node{F: TaskForComparison}
		last.Nodes = append(last.Nodes, node)
		last = node
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

func BenchmarkTaskTreeManyToOne10(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	for i := 0; i < 9; i++ {
		node := &Node{F: TaskForComparison}
		root.Nodes = append(root.Nodes, node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

func BenchmarkTaskTreeManyToOne100(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	for i := 0; i < 99; i++ {
		node := &Node{F: TaskForComparison}
		root.Nodes = append(root.Nodes, node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

func BenchmarkTaskTreeManyToOne1000(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	for i := 0; i < 999; i++ {
		node := &Node{F: TaskForComparison}
		root.Nodes = append(root.Nodes, node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

func BenchmarkTaskTreeBinaryTree10(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	queue := []*Node{root}
	for i := 0; i < 3; i++ {
		tmp := []*Node{}
		for _, node := range queue {
			left := &Node{F: TaskForComparison}
			right := &Node{F: TaskForComparison}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

func BenchmarkTaskTreeBinaryTree100(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	queue := []*Node{root}
	for i := 0; i < 6; i++ {
		tmp := []*Node{}
		for _, node := range queue {
			left := &Node{F: TaskForComparison}
			right := &Node{F: TaskForComparison}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

func BenchmarkTaskTreeBinaryTree1000(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	queue := []*Node{root}
	for i := 0; i < 9; i++ {
		tmp := []*Node{}
		for _, node := range queue {
			left := &Node{F: TaskForComparison}
			right := &Node{F: TaskForComparison}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

func BenchmarkTaskTreeBinaryTree10000(b *testing.B) {
	t := &TaskTree{}
	root := &Node{F: TaskForComparison}
	queue := []*Node{root}
	for i := 0; i < 12; i++ {
		tmp := []*Node{}
		for _, node := range queue {
			left := &Node{F: TaskForComparison}
			right := &Node{F: TaskForComparison}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go(root, nil)
		<-chWait()
	}
}

// ---------------------------------------
// After updating the code of lesismal
// ---------------------------------------
type Node2 struct {
	F     func(context.Context) error
	Nodes []*Node2

	selfCnt    int32
	parentCnt  *int32
	parentDone chan struct{}
	parentFunc func(context.Context) error
	failFunc   func(error)
}

func (node *Node2) doNodes2(ctx context.Context, parentCnt *int32, parentDone chan struct{}, parentFunc func(context.Context) error, failFunc func(error)) {
	node.selfCnt = int32(len(node.Nodes))
	node.parentDone = parentDone
	node.parentCnt = parentCnt
	node.parentFunc = parentFunc
	node.failFunc = failFunc
	if node.selfCnt > 0 {
		for _, v := range node.Nodes {
			v.doNodes2(ctx, &node.selfCnt, nil, node.doSelf2, failFunc)
		}
		return
	}
	go node.doSelf2(ctx)
}

func (node *Node2) doSelf2(ctx context.Context) error {
	var err error
	defer func() {
		if node.parentDone != nil {
			select {
			case node.parentDone <- struct{}{}:
			default:
			}
		} else if err == nil && atomic.AddInt32(node.parentCnt, -1) == 0 && node.parentFunc != nil {
			if err := node.parentFunc(ctx); err != nil && node.failFunc != nil {
				node.failFunc(err)
			}
		}
	}()
	err = node.F(ctx)
	if err != nil && node.failFunc != nil {
		node.failFunc(err)
	}
	return nil
}

type TaskTree2 struct{}

func (t *TaskTree2) Go2(ctx context.Context, root *Node2) func() <-chan error {
	chErr := make(chan error, 1)
	chDone := make(chan struct{}, 1)
	waitFunc := func() <-chan error {
		var err error
		select {
		case <-chDone:
		case err = <-chErr:
		}
		select {
		case chErr <- err:
		default:
		}
		return chErr
	}
	var n int32
	var failFuncCalled int32
	root.doNodes2(ctx, &n, chDone, nil, func(err error) {
		if atomic.AddInt32(&failFuncCalled, 1) == 1 {
			chErr <- err
		}
	})
	return waitFunc
}

func TaskForComparison2(ctx context.Context) error {
	time.Sleep(time.Microsecond*SLEEPTIME)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	i := 0
	for i < ITERNUM {
		i++
	}
	return nil
}

func BenchmarkTaskTree2Serialized10(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	last := root
	for i := 0; i < 9; i++ {
		node := &Node2{F: TaskForComparison2}
		last.Nodes = append(last.Nodes, node)
		last = node
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}
func BenchmarkTaskTree2Serialized100(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	last := root
	for i := 0; i < 99; i++ {
		node := &Node2{F: TaskForComparison2}
		last.Nodes = append(last.Nodes, node)
		last = node
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}
func BenchmarkTaskTree2Serialized1000(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	last := root
	for i := 0; i < 999; i++ {
		node := &Node2{F: TaskForComparison2}
		last.Nodes = append(last.Nodes, node)
		last = node
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}

func BenchmarkTaskTree2ManyToOne10(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	for i := 0; i < 9; i++ {
		node := &Node2{F: TaskForComparison2}
		root.Nodes = append(root.Nodes, node)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}

func BenchmarkTaskTree2ManyToOne100(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	for i := 0; i < 99; i++ {
		node := &Node2{F: TaskForComparison2}
		root.Nodes = append(root.Nodes, node)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}

func BenchmarkTaskTree2ManyToOne1000(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	for i := 0; i < 999; i++ {
		node := &Node2{F: TaskForComparison2}
		root.Nodes = append(root.Nodes, node)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}

func BenchmarkTaskTree2BinaryTree10(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	queue := []*Node2{root}
	for i := 0; i < 3; i++ {
		tmp := []*Node2{}
		for _, node := range queue {
			left := &Node2{F: TaskForComparison2}
			right := &Node2{F: TaskForComparison2}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}

func BenchmarkTaskTree2BinaryTree100(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	queue := []*Node2{root}
	for i := 0; i < 6; i++ {
		tmp := []*Node2{}
		for _, node := range queue {
			left := &Node2{F: TaskForComparison2}
			right := &Node2{F: TaskForComparison2}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}

func BenchmarkTaskTree2BinaryTree1000(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	queue := []*Node2{root}
	for i := 0; i < 9; i++ {
		tmp := []*Node2{}
		for _, node := range queue {
			left := &Node2{F: TaskForComparison2}
			right := &Node2{F: TaskForComparison2}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}

func BenchmarkTaskTree2BinaryTree10000(b *testing.B) {
	t := &TaskTree2{}
	root := &Node2{F: TaskForComparison2}
	queue := []*Node2{root}
	for i := 0; i < 12; i++ {
		tmp := []*Node2{}
		for _, node := range queue {
			left := &Node2{F: TaskForComparison2}
			right := &Node2{F: TaskForComparison2}
			node.Nodes = append(node.Nodes, left)
			node.Nodes = append(node.Nodes, right)
			tmp = append(tmp, left)
			tmp = append(tmp, right)
		}
		queue = tmp
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chWait := t.Go2(ctx, root)
		<-chWait()
	}
}
