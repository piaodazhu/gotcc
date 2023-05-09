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

func (e *Executor) NewDependencyExpr(d *Executor) DependencyExpression {
	if _, exists := e.dependency[d.id]; !exists {
		e.dependency[d.id] = false
		e.messageBuffer = make(chan message, cap(e.messageBuffer)+1)
		d.subscribers = append(d.subscribers, &e.messageBuffer)
	}
	return newDependencyExpr(e.dependency, d.id)
}

func (e *Executor) DependencyExpr() DependencyExpression {
	return e.dependencyExpr
}

func (e *Executor) SetDependency(Expr DependencyExpression) *Executor {
	e.dependencyExpr = Expr
	return e
}

func (e *Executor) CalcDependency() bool {
	return e.dependencyExpr.f()
}

func (e *Executor) MarkDependency(id uint32, finished bool) {
	e.dependency[id] = finished
}

func (e *Executor) SetUndoFunc(undo func(args map[string]interface{}) error, skipError bool) *Executor {
	e.undo = undo
	e.undoSkipError = skipError
	return e
}

func (e *Executor) Name() string {
	return e.name
}
