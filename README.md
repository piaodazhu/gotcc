[![Go Reference](https://pkg.go.dev/badge/github.com/piaodazhu/gotcc.svg)](https://pkg.go.dev/github.com/piaodazhu/gotcc)
[![Go Report Card](https://goreportcard.com/badge/github.com/piaodazhu/gotcc)](https://goreportcard.com/report/github.com/piaodazhu/gotcc)
[![codecov](https://codecov.io/gh/piaodazhu/gotcc/branch/master/graph/badge.svg?token=KMOWEKDPN5)](https://codecov.io/gh/piaodazhu/gotcc)

# gotcc

🤖 `gotcc` is a Golang package for Task Concurrency Control. It allows you to define tasks and their dependencies, then the controller will run the tasks concurrently while respecting the dependencies.

Features of `gotcc`
- Automatic task concurrency control based on dependency declarations.
- Support dependency logic expressions: `not`, `and`, `or`, `xor` and any combination of them.
- Many-to-many result delivery between tasks.
- Support tasks rollback in case of any error.
- Support multiple errors collection.
- Support coroutine pool: default([panjf2000/ants](https://github.com/panjf2000/ants)) or user-defined coroutine pool.

## Installation

```bash
go get github.com/piaodazhu/gotcc
```

## Usage

A simple usage:

```go
import "github.com/piaodazhu/gotcc"

// User-defined task function
func ExampleFunc1(args map[string]interface{}) (interface{}, error) {
	fmt.Println(args["BIND"].(string))
	return "DONE", nil
}
func ExampleFunc2(args map[string]interface{}) (interface{}, error) {
	return args["BIND"].(string), nil
}

// User-defined undo function
func ExampleUndo(args map[string]interface{}) error {
	fmt.Println("Undo > ", args["BIND"].(string))
	return nil
}

func main() {
	// 1. Create a new controller
	controller := gotcc.NewTCController()

	// 2. Add tasks to the controller

	// TaskA: bind arguments with ExampleFunc1
	taskA := controller.AddTask("taskA", ExampleFunc1, "BindArg-A")

	// TaskB: like TaskA, but set undoFunction
	taskB := controller.AddTask("taskB", ExampleFunc1, "BindArg-B").SetUndoFunc(ExampleUndo, true)

	// TaskC: bind arguments with ExampleFunc2
	taskC := controller.AddTask("taskC", ExampleFunc2, "BindArg-C")
	
	// TaskD: bind arguments with ExampleFunc2
	taskD := controller.AddTask("taskD", ExampleFunc2, "BindArg-D")

	// 3. Define dependencies

	// B depend on A
	taskB.SetDependency(taskB.NewDependencyExpr(taskA))

	// C depend on A
	taskC.SetDependency(taskC.NewDependencyExpr(taskA))

	// D depend on B and C
	taskD.SetDependency(gotcc.MakeAndExpr(taskD.NewDependencyExpr(taskB), taskD.NewDependencyExpr(taskC)))

	// 4. Define termination (Important)

	// set TaskD's finish as termination
	controller.SetTermination(controller.NewTerminationExpr(taskD))
	
	// 5. Run the tasks
	result, err := controller.BatchRun()
	if err != nil {
		// get taskErrors: err.(ErrAborted).TaskErrors
		// get undoErrors: err.(ErrAborted).UndoErrors
	}

	// 6. Will Print "BindArg-D"
	fmt.Println(result["taskD"].(string))
}
```
Tasks will run concurrently, but taskB and taskC will not start until taskA completes, and taskD will not start until both taskB and taskC complete. But if taskD failed (return err!=nil), `ExampleUndo("BindArg-B")` will be executed.

More detailed usage information can be found in test files, you can refer to `example_test.go` for a more complex dependency topology, `dependency_test.go` for the advanced usage of dependency logic expressions, and `tcc_test.go` for tasks rollback and error message collection.

## Specifications

### Execution
In summary, a single execution of the TCController contains multiple tasks. There may be some dependencies between tasks, and the termination of the execution depends on the completion of some of these tasks. **Therefore, `controller.SetTermination` must be called before calling `controller.BatchRun` or `controller.PoolRun`.**

There are 2 mode for the TCController to execute the tasks: `BatchRun` and `PoolRun`. `BatchRun` will create NumOf(tasks) goroutines at most. If we need to control the max number of running goroutines, `PoolRun` is recommand. A default coroutine pool is provided, based on [panjf2000/ants](https://github.com/panjf2000/ants). User-defined coroutine pool should implement this interface, where Go() is a task submission method and should **block** when workers are busy:
```go
type GoroutinePool interface {
	Go(task func()) error
}
```

> Node that `PoolRun` mode only avalible when all dependency expressions are `AND`.

### Task Function
The task function must have this form：
```go
func (args map[string]interface{}) (interface{}, error)
```
There are some built-in keys when running the task function:
- `NAME`: the value is the name of this task.
- `BIND`: the value is the third arguments when `controller.AddTask()` was called.
- `CANCEL`: the value is a context.Context, with cancel.

Other keys are the **names** of its dependent tasks, and the corresponding values are the return value of these tasks.

**IMPORTANT**: Inside task functions, if the task is cancelled by receiving signal from `args["CANCEL"].(context.Context).done()`, it should return `gotcc.ErrCancelled` (with state if necessary). if the task failed but you don't want abort the execution, it should return `gotcc.ErrSilentFail`.

### Undo Function
The undo function must have this form：
```go
func (args map[string]interface{}) error
```
There are some built-in keys when running the undo function:
- `NAME`: the value is the name of this task.
- `BIND`: the value is the third arguments when `controller.AddTask()` was called.
- `TASKERR`: the value type is `[]*gotcc.ErrorMessage`, recording the errors of tasks execution.
- `UNDOERR`: the value type is `[]*gotcc.ErrorMessage`, recording the previous errors of undo execution.
- `CANCELLED`: the value type is `[]*gotcc.StateMessage`, recording the state of canncelld task. (For example, what process in that task has been done before cancelled.)

The undo functions will be run in the reverse order of the task function completion. And the second arguments of `SetUndoFunc` means whether to skip this error if the undo function errors out.

The undo function will be executed when:
1. Some task return `err!=nil` when the controller execute the tasks.
2. The corresponding task has been completed.
3. The predecessor undo functions have been completed or skipped.

When the undo function run, the arguments `args` is exactly the same as its corresponding task.

### Errors

During the execution of TCController, multiple tasks may fail and after failure, multiple tasks may be cancelled. During rollback, multiple rollback functions may also encounter errors. Therefore, the error definitions in the return value of `Run` are as follows:
```go
type ErrAborted struct {
	TaskErrors []*ErrorMessage
	UndoErrors []*ErrorMessage
	Cancelled  []*StateMessage
}

type ErrorMessage struct {
	TaskName string
	Error    error
}

type StateMessage struct {
	TaskName string
	State    State
}

```

### Dependency Expression

Supported dependency logic expressions are `not`, `and`, `or`, `xor` and any combination of them.

For taskB, create a dependency expression about taskA:
```go
ExprAB := taskB.NewDependencyExpr(taskA)
```

Combine existing dependency expressions to generate dependency expressions:
```go
Expr3 := gotcc.MakeOrExpr(Expr1, Expr2)
```

Get the current dependency expression of taskA.
```go
Expr := taskA.DependencyExpr()
```

Set the dependency expression for taskA.
```go
taskA.SetDependencyExpr(Expr)
```

And termination setup has the same logic as above.

## Performance
```bash
goos: linux
goarch: amd64
pkg: github.com/piaodazhu/gotcc
cpu: 11th Gen Intel(R) Core(TM) i7-11700 @ 2.50GHz
BenchmarkBatchRunSerialized10-4            54928             19533 ns/op            7106 B/op         84 allocs/op
BenchmarkBatchRunSerialized100-4            6452            172314 ns/op           68873 B/op        750 allocs/op
BenchmarkBatchRunSerialized1000-4            507           2301349 ns/op          772775 B/op       7493 allocs/op
BenchmarkBatchRunManyToOne10-4             59264             20095 ns/op            7673 B/op         77 allocs/op
BenchmarkBatchRunManyToOne100-4             5623            201600 ns/op           78210 B/op        659 allocs/op
BenchmarkBatchRunManyToOne1000-4             100          11388471 ns/op          899267 B/op       6178 allocs/op
BenchmarkBatchRunManyToMany10-4            23252             50629 ns/op           21549 B/op        212 allocs/op
BenchmarkBatchRunManyToMany100-4             410           2814498 ns/op         2270278 B/op      15945 allocs/op
BenchmarkBatchRunBinaryTree10-4            34041             37472 ns/op           10857 B/op        116 allocs/op
BenchmarkBatchRunBinaryTree100-4            6380            204777 ns/op           91623 B/op        880 allocs/op
BenchmarkBatchRunBinaryTree1000-4            804           1506047 ns/op          749709 B/op       6848 allocs/op
BenchmarkPoolRunSerialized10-4             49352             24033 ns/op            7622 B/op         91 allocs/op
BenchmarkPoolRunSerialized100-4             5956            180609 ns/op           72491 B/op        754 allocs/op
BenchmarkPoolRunSerialized1000-4             710           1617208 ns/op          783005 B/op       7222 allocs/op
BenchmarkPoolRunManyToOne10-4              38798             31265 ns/op            8193 B/op         84 allocs/op
BenchmarkPoolRunManyToOne100-4              5371            215742 ns/op           81924 B/op        664 allocs/op
BenchmarkPoolRunManyToOne1000-4              100          11651428 ns/op          937995 B/op       6226 allocs/op
BenchmarkPoolRunManyToMany10-4             22332             52499 ns/op           22833 B/op        218 allocs/op
BenchmarkPoolRunManyToMany100-4              336           3698301 ns/op         2360297 B/op      15935 allocs/op
BenchmarkPoolRunBinaryTree10-4             29043             42600 ns/op           11543 B/op        122 allocs/op
BenchmarkPoolRunBinaryTree100-4             5572            216561 ns/op           96321 B/op        885 allocs/op
BenchmarkPoolRunBinaryTree1000-4             826           1394863 ns/op          782612 B/op       6812 allocs/op
```

## License
`gotcc` is released under the MIT license. See [LICENSE](https://github.com/piaodazhu/gotcc/blob/master/LICENSE) for details.
