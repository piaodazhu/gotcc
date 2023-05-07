package gotcc

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TaskForBench(args map[string]interface{}) (interface{}, error) {
	// time.Sleep(time.Microsecond)
	// i := 0
	// for i < 100000 {
	// 	i++
	// }
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
	time.Sleep(time.Microsecond)
	i := 0
	for i < 100000 {
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
