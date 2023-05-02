# gotcc

🤖 `gotcc` is a Golang package for Task Concurrency Control. It allows you to define tasks, their dependencies, and the controller will run the tasks concurrently while respecting the dependencies.

Features of `gotcc`
- Automatic task concurrency control based on dependency declarations.
- Support dependency logic expressions: `not`, `and`, `or`, `xor` and any combination of them.
- Many-to-many result delivery between tasks.
- Support tasks rollback in case of any error.
- Support multiple error collection.

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
	//   TaskA: bind arguments with ExampleFunc1
	taskA := controller.AddTask("taskA", ExampleFunc1, "BindArg-A")
	//   TaskB: like TaskA, but set undoFunction
	taskB := controller.AddTask("taskB", ExampleFunc1, "BindArg-B").SetUndoFunc(ExampleUndo, true)
	//   TaskC: bind arguments with ExampleFunc2
	taskC := controller.AddTask("taskC", ExampleFunc2, "BindArg-C")
	
	//   TaskD: bind arguments with ExampleFunc2
	taskD := controller.AddTask("taskD", ExampleFunc2, "BindArg-D")

	// 3. Define dependencies
	//   B depend on A
	taskB.SetDependency(taskB.NewDependencyExpr(taskA))
	//   C depend on A
	taskC.SetDependency(taskC.NewDependencyExpr(taskA))
	//   D depend on B and C
	taskD.SetDependency(gotcc.MakeAndExpr(taskD.NewDependencyExpr(taskB), taskD.NewDependencyExpr(taskC)))

	// 4. Define termination (Important)
	//   set TaskD's finish as termination
	controller.SetTermination(controller.NewTerminationExpr(taskD))
	
	// 5. Run the tasks
	result, err := controller.RunTasks()
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
In summary, a single execution of the TCController contains multiple tasks. There may be some dependencies between tasks, and the termination of the execution depends on the completion of some of these tasks. **Therefore, `controller.SetTermination` must be called before calling `controller.RunTasks`.**

### Task Function
The task function must have this form：
```go
func (args map[string]interface{}) (interface{}, error)
```
There are some built-in keys when running the task function:
- `BIND`: the value is the third arguments when `controller.AddTask()` was called.
- `CANCEL`: the value is a context.Context, with cancel.

Other keys are the **names** of its dependent tasks, and the corresponding values are the return value of these tasks.

### Undo Function
The undo function must have this form：
```go
func (args map[string]interface{}) error
```

The undo functions will be run in the reverse order of the task function completion. And the second arguments of `SetUndoFunc` means whether to skip this error if the undo function errors out.

The undo function will be executed when:
1. Some task return `err!=nil` when the controller execute the tasks.
2. The corresponding task has been completed.
3. The predecessor undo functions have been completed or skipped.

When the undo function run, the arguments `args` is exactly the same as its corresponding task.

### Errors

During the execution of TCController, multiple tasks may fail and after failure, multiple tasks may be cancelled. During rollback, multiple rollback functions may also encounter errors. Therefore, the error definitions in the return value of `RunTasks` are as follows:
```go
type ErrAborted struct {
	TaskErrors *ErrorList
	UndoErrors *ErrorList
}

type ErrorList struct {
	Lock  sync.Mutex
	Items []*ErrorMessage
}

type ErrorMessage struct {
	TaskName string
	Error    error
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

## License
`gotcc` is released under the MIT license. See [LICENSE](https://github.com/piaodazhu/gotcc/blob/master/LICENSE) for details.
