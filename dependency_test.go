package gotcc

import (
	"testing"
)

func TestDependency(t *testing.T) {
	A := newExecutor("A", nil, "A")
	B := newExecutor("B", nil, "B")
	C := newExecutor("C", nil, "C")
	D := newExecutor("D", nil, "D")
	E := newExecutor("E", nil, "E")
	F := newExecutor("F", nil, "F")

	// C <- A && B
	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B)))

	// D <- A || B
	D.SetDependency(MakeOrExpr(D.NewDependencyExpr(A), D.NewDependencyExpr(B)))

	// E <- !D
	E.SetDependency(MakeNotExpr(E.NewDependencyExpr(D)))

	// F <- B ^ D
	F.SetDependency(MakeXorExpr(F.NewDependencyExpr(B), F.NewDependencyExpr(D)))

	// value table:
	// A B   C  D  E  F
	// 0 0   0  0  1  0
	// 1 0   0  1  0  1
	// 0 1   0  1  0  0
	// 1 1   1  1  0  0

	// 0 0   0  0  1  0
	A.SetDependency(DefaultFalseExpr)
	B.SetDependency(DefaultFalseExpr)

	C.MarkDependency(A.Id, A.DependencyExpr()())
	C.MarkDependency(B.Id, B.DependencyExpr()())

	D.MarkDependency(A.Id, A.DependencyExpr()())
	D.MarkDependency(B.Id, B.DependencyExpr()())

	E.MarkDependency(D.Id, D.DependencyExpr()())

	F.MarkDependency(B.Id, B.DependencyExpr()())
	F.MarkDependency(D.Id, D.DependencyExpr()())

	if A.DependencyExpr()() != false || B.DependencyExpr()() != false || C.DependencyExpr()() != false || D.DependencyExpr()() != false || E.DependencyExpr()() != true || F.DependencyExpr()() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.DependencyExpr(), B.DependencyExpr(), C.DependencyExpr(), D.DependencyExpr(), E.DependencyExpr(), F.DependencyExpr())
	}

	// 1 0   0  1  0  1
	A.SetDependency(DefaultTrueExpr)
	B.SetDependency(DefaultFalseExpr)

	C.MarkDependency(A.Id, A.DependencyExpr()())
	C.MarkDependency(B.Id, B.DependencyExpr()())

	D.MarkDependency(A.Id, A.DependencyExpr()())
	D.MarkDependency(B.Id, B.DependencyExpr()())

	E.MarkDependency(D.Id, D.DependencyExpr()())

	F.MarkDependency(B.Id, B.DependencyExpr()())
	F.MarkDependency(D.Id, D.DependencyExpr()())

	if A.DependencyExpr()() != true || B.DependencyExpr()() != false || C.DependencyExpr()() != false || D.DependencyExpr()() != true || E.DependencyExpr()() != false || F.DependencyExpr()() != true {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.DependencyExpr(), B.DependencyExpr(), C.DependencyExpr(), D.DependencyExpr(), E.DependencyExpr(), F.DependencyExpr())
	}

	// 0 1   0  1  0  0
	A.SetDependency(DefaultFalseExpr)
	B.SetDependency(DefaultTrueExpr)

	C.MarkDependency(A.Id, A.DependencyExpr()())
	C.MarkDependency(B.Id, B.DependencyExpr()())

	D.MarkDependency(A.Id, A.DependencyExpr()())
	D.MarkDependency(B.Id, B.DependencyExpr()())

	E.MarkDependency(D.Id, D.DependencyExpr()())

	F.MarkDependency(B.Id, B.DependencyExpr()())
	F.MarkDependency(D.Id, D.DependencyExpr()())

	if A.DependencyExpr()() != false || B.DependencyExpr()() != true || C.DependencyExpr()() != false || D.DependencyExpr()() != true || E.DependencyExpr()() != false || F.DependencyExpr()() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.DependencyExpr(), B.DependencyExpr(), C.DependencyExpr(), D.DependencyExpr(), E.DependencyExpr(), F.DependencyExpr())
	}

	// 1 1   1  1  0  0
	A.SetDependency(DefaultTrueExpr)
	B.SetDependency(DefaultTrueExpr)

	C.MarkDependency(A.Id, A.DependencyExpr()())
	C.MarkDependency(B.Id, B.DependencyExpr()())

	D.MarkDependency(A.Id, A.DependencyExpr()())
	D.MarkDependency(B.Id, B.DependencyExpr()())

	E.MarkDependency(D.Id, D.DependencyExpr()())

	F.MarkDependency(B.Id, B.DependencyExpr()())
	F.MarkDependency(D.Id, D.DependencyExpr()())

	if A.DependencyExpr()() != true || B.DependencyExpr()() != true || C.DependencyExpr()() != true || D.DependencyExpr()() != true || E.DependencyExpr()() != false || F.DependencyExpr()() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.DependencyExpr(), B.DependencyExpr(), C.DependencyExpr(), D.DependencyExpr(), E.DependencyExpr(), F.DependencyExpr())
	}
}
