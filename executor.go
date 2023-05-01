package gooc

import (
	"github.com/google/uuid"
)

type Executor struct {
	Id   uint32
	Name string

	Dependency     map[uint32]bool
	DependencyExpr DependencyExpression

	MessageBuffer chan Message
	Subscribers   []chan Message

	BindArgs     interface{}
	Task         func(args map[string]interface{}) (interface{}, error)
	RollbackTask func(args map[string]interface{}) (interface{}, error)

	Manager *Manager
}

func newExecutor(name string, f func(args map[string]interface{}) (interface{}, error), args interface{}) *Executor {
	return &Executor{
		Id:   uuid.New().ID(),
		Name: name,

		Dependency:     map[uint32]bool{},
		DependencyExpr: DefaultTrueExpr,

		MessageBuffer: make(chan Message),
		Subscribers:   []chan Message{},

		BindArgs:     args,
		Task:         f,
		RollbackTask: f,

		Manager: nil,
	}
}

func (e *Executor) NewDependencyExpr(d *Executor) DependencyExpression {
	if _, exists := e.Dependency[d.Id]; !exists {
		e.Dependency[d.Id] = false
		// need buffered chan?
		e.MessageBuffer = make(chan Message, cap(e.MessageBuffer)+1)
		d.Subscribers = append(d.Subscribers, e.MessageBuffer)
	}
	return newDependencyExpr(e.Dependency, d.Id)
}

func (e *Executor) SetDependency(Expr DependencyExpression) *Executor {
	// delete(e.Manager.StartSet, e.Id)
	e.DependencyExpr = Expr
	return e
}

func (e *Executor) CalcDependency() bool {
	return e.DependencyExpr()
}
