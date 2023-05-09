package gotcc

import (
	"sort"
)

// A dependency expression is a filter to describe the tasks' dependency
// A task will be launched only if the expression is true.
type DependencyExpression struct {
	f      func() bool
	allAnd bool
}

func MakeNotExpr(Expr DependencyExpression) DependencyExpression {
	return DependencyExpression{
		f: func() bool {
			return !Expr.f()
		},
		allAnd: false,
	}
}

func MakeAndExpr(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return DependencyExpression{
		f: func() bool {
			return Expr1.f() && Expr2.f()
		},
		allAnd: Expr1.allAnd && Expr2.allAnd,
	}
}

func MakeOrExpr(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return DependencyExpression{
		f: func() bool {
			return Expr1.f() || Expr2.f()
		},
		allAnd: false,
	}
}

func MakeXorExpr(Expr1 DependencyExpression, Expr2 DependencyExpression) DependencyExpression {
	return DependencyExpression{
		f: func() bool {
			return (Expr1.f() && !Expr2.f()) || (!Expr1.f() && Expr2.f())
		},
		allAnd: false,
	}
}

func newDependencyExpr(valMap map[uint32]bool, key uint32) DependencyExpression {
	return DependencyExpression{
		f: func() bool {
			return valMap[key]
		},
		allAnd: true,
	}
}

func (m *TCController) analyzeDependency() (map[uint32]int, bool) {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[uint32]int{}
	order := map[uint32]int{}
	var dfs func(curr uint32) bool
	dfs = func(curr uint32) bool {
		if len(m.executors[curr].dependency) == 0 {
			order[curr] = 0
			color[curr] = black
			return true
		}

		color[curr] = gray
		maxorder := 0
		for neighbor := range m.executors[curr].dependency {
			switch color[neighbor] {
			case white:
				if !dfs(neighbor) {
					return false
				}
				maxorder = max(maxorder, order[neighbor])
			case gray:
				return false
			case black:
				maxorder = max(maxorder, order[neighbor])
			}
		}
		color[curr] = black
		order[curr] = maxorder + 1
		return true
	}

	for taskid := range m.executors {
		if !dfs(taskid) {
			return nil, false
		}
	}

	return order, true
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func (m *TCController) sortExecutor(taskorder map[uint32]int) ([]uint32, bool) {
	type item struct {
		taskid   uint32
		taskname string
		order    int
	}
	itemlist := make([]item, 0, len(taskorder))
	res := make([]uint32, 0, len(taskorder))
	for taskid, order := range taskorder {
		e := m.executors[taskid]
		if !e.dependencyExpr.allAnd {
			return nil, false
		}
		itemlist = append(itemlist, item{
			taskid:   taskid,
			taskname: e.name,
			order:    order,
		})
	}
	sort.Slice(itemlist, func(i, j int) bool {
		if itemlist[i].order == itemlist[j].order {
			return itemlist[i].taskname < itemlist[j].taskname
		}
		return itemlist[i].order < itemlist[j].order
	})
	for i := range itemlist {
		res = append(res, itemlist[i].taskid)
	}
	return res, true
}

// default dependency expression: always return true
var DefaultTrueExpr = DependencyExpression{
	f: func() bool {
		return true
	},
	allAnd: true,
}

// default dependency expression: always return false
var DefaultFalseExpr = DependencyExpression{
	f: func() bool {
		return false
	},
	allAnd: true,
}
