package gotcc

import (
	"github.com/google/uuid"
)

type Executor struct {
	Id   uint32
	Name string

	BindArgs      interface{}
	Task          func(args map[string]interface{}) (interface{}, error)
	Undo          func(args map[string]interface{}) error
	UndoSkipError bool

	dependency     map[uint32]bool
	dependencyExpr DependencyExpression

	messageBuffer chan Message
	subscribers   []*chan Message
}

func newExecutor(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	return &Executor{
		Id:   uuid.New().ID(),
		Name: name,

		dependency:     map[uint32]bool{},
		dependencyExpr: DefaultTrueExpr,

		messageBuffer: make(chan Message),
		subscribers:   []*chan Message{},

		BindArgs: args,
		Task:     f,
		Undo:     EmptyUndoFunc,
	}
}

func (e *Executor) NewDependencyExpr(d *Executor) DependencyExpression {
	if _, exists := e.dependency[d.Id]; !exists {
		e.dependency[d.Id] = false
		e.messageBuffer = make(chan Message, cap(e.messageBuffer)+1)
		d.subscribers = append(d.subscribers, &e.messageBuffer)
	}
	return newDependencyExpr(e.dependency, d.Id)
}

func (e *Executor) DependencyExpr() DependencyExpression {
	return e.dependencyExpr
}

func (e *Executor) SetDependency(Expr DependencyExpression) *Executor {
	e.dependencyExpr = Expr
	return e
}

func (e *Executor) MarkDependency(id uint32, finished bool) {
	e.dependency[id] = finished
}

func (e *Executor) SetUndoFunc(undo func(args map[string]interface{}) error, SkipError bool) *Executor {
	e.Undo = undo
	e.UndoSkipError = SkipError
	return e
}

type ErrCancelled struct {
	State State
}

func (e ErrCancelled) Error() string {
	return "Error: Task is canncelled due to other errors."
}
