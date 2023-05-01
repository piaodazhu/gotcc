package gooc


type DependencyExpression func() bool

func DefaultTrueExpr() bool {
	return true
}

func DefaultFalseExpr() bool {
	return false
}

func newDependencyExpr(valMap map[uint32]bool, key uint32) DependencyExpression {
	return func() bool {
		return valMap[key]
	}
}

func MakeNotExpr(Expr DependencyExpression) DependencyExpression {
	return func() bool {
		return !Expr()
	}
}

func MakeAndExpr(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return func() bool {
		return Expr1() && Expr2()
	}
}

func MakeOrExpr(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return func() bool {
		return Expr1() || Expr2()
	}
}

func MakeXorExpr(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return func() bool {
		return (Expr1() && !Expr2()) || (!Expr1() && Expr2())
	}
}