package gotcc

import (
	"github.com/google/uuid"
)

type Executor struct {
	id   uint32
	name string

	bindArgs      interface{}
	task          func(args map[string]interface{}) (interface{}, error)
	undo          func(args map[string]interface{}) error
	undoSkipError bool

	dependency     map[uint32]bool
	dependencyExpr DependencyExpression

	messageBuffer chan message
	subscribers   []*chan message
}

func newExecutor(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	return &Executor{
		id:   uuid.New().ID(),
		name: name,

		dependency:     map[uint32]bool{},
		dependencyExpr: DefaultTrueExpr,

		messageBuffer: make(chan message),
		subscribers:   []*chan message{},

		bindArgs: args,
		task:     f,
		undo:     EmptyUndoFunc,
	}
}

// Create a dependency expression for the executor.
// It means the task launching may depend on executor `d`.
func (e *Executor) NewDependencyExpr(d *Executor) DependencyExpression {
	if _, exists := e.dependency[d.id]; !exists {
		e.dependency[d.id] = false
		e.messageBuffer = make(chan message, cap(e.messageBuffer)+1)
		d.subscribers = append(d.subscribers, &e.messageBuffer)
	}
	return newDependencyExpr(e.dependency, d.id)
}

// Get dependency expression of the executor.
func (e *Executor) DependencyExpr() DependencyExpression {
	return e.dependencyExpr
}

// Set dependency expression for the executor. `Expr` is a dependency expression.
func (e *Executor) SetDependency(Expr DependencyExpression) *Executor {
	e.dependencyExpr = Expr
	return e
}

func (e *Executor) calcDependency() bool {
	return e.dependencyExpr.f()
}

func (e *Executor) markDependency(id uint32, finished bool) {
	e.dependency[id] = finished
}

// Set undo function the task executor. The undo function will get all arguments of the task function.
func (e *Executor) SetUndoFunc(undo func(args map[string]interface{}) error, skipError bool) *Executor {
	e.undo = undo
	e.undoSkipError = skipError
	return e
}

// Get task name of the executor.
func (e *Executor) Name() string {
	return e.name
}
