package main

type DependencyExpression func(dependency map[uint32]bool) bool

func NoDependency(dependency map[uint32]bool) bool {
	return true
}

func NewDependencyExpr(e *Executor, dependency map[uint32]bool) DependencyExpression {
	return func (dependency map[uint32]bool) bool {
		return dependency[e.Id]
	}
}

func not(Expr DependencyExpression) DependencyExpression {
	return func(dependency map[uint32]bool) bool {
		return !Expr(dependency)
	}
}

func and(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return func(dependency map[uint32]bool) bool {
		return Expr1(dependency) && Expr2(dependency)
	}
}

func or(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return func(dependency map[uint32]bool) bool {
		return Expr1(dependency) || Expr2(dependency)
	}
}

func (e *Executor) MakeNotDependency(d *Executor) DependencyExpression {
	if _, exists := e.Dependency[d.Id]; !exists {
		e.Dependency[d.Id] = false
		// need buffered chan?
		e.MessageBuffer = make(chan Message, cap(e.MessageBuffer) + 1)
		d.Subscribers = append(d.Subscribers, e.MessageBuffer)
	}
	return not(NewDependencyExpr(d, e.Dependency))
}

func (e *Executor) MakeAndDependency(d *Executor, Expr DependencyExpression) DependencyExpression {
	if _, exists := e.Dependency[d.Id]; !exists {
		e.Dependency[d.Id] = false
		// need buffered chan?
		e.MessageBuffer = make(chan Message, cap(e.MessageBuffer) + 1)
		d.Subscribers = append(d.Subscribers, e.MessageBuffer)
	}
	return and(NewDependencyExpr(d, e.Dependency), Expr)
}

func (e *Executor) MakeOrDependency(d *Executor, Expr DependencyExpression) DependencyExpression {
	if _, exists := e.Dependency[d.Id]; !exists {
		e.Dependency[d.Id] = false
		// need buffered chan?
		e.MessageBuffer = make(chan Message, cap(e.MessageBuffer) + 1)
		d.Subscribers = append(d.Subscribers, e.MessageBuffer)
	}
	return or(NewDependencyExpr(d, e.Dependency), Expr)
}

func (e *Executor) SetDependency(Expr DependencyExpression) *Executor {
	// delete(e.Manager.StartSet, e.Id)
	e.DependencyExpr = Expr 
	return e
}