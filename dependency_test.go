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

	C.MarkDependency(A.Id, A.CalcDependency())
	C.MarkDependency(B.Id, B.CalcDependency())

	D.MarkDependency(A.Id, A.CalcDependency())
	D.MarkDependency(B.Id, B.CalcDependency())

	E.MarkDependency(D.Id, D.CalcDependency())

	F.MarkDependency(B.Id, B.CalcDependency())
	F.MarkDependency(D.Id, D.CalcDependency())

	if A.CalcDependency() != false || B.CalcDependency() != false || C.CalcDependency() != false || D.CalcDependency() != false || E.CalcDependency() != true || F.CalcDependency() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 1 0   0  1  0  1
	A.SetDependency(DefaultTrueExpr)
	B.SetDependency(DefaultFalseExpr)

	C.MarkDependency(A.Id, A.CalcDependency())
	C.MarkDependency(B.Id, B.CalcDependency())

	D.MarkDependency(A.Id, A.CalcDependency())
	D.MarkDependency(B.Id, B.CalcDependency())

	E.MarkDependency(D.Id, D.CalcDependency())

	F.MarkDependency(B.Id, B.CalcDependency())
	F.MarkDependency(D.Id, D.CalcDependency())

	if A.CalcDependency() != true || B.CalcDependency() != false || C.CalcDependency() != false || D.CalcDependency() != true || E.CalcDependency() != false || F.CalcDependency() != true {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 0 1   0  1  0  0
	A.SetDependency(DefaultFalseExpr)
	B.SetDependency(DefaultTrueExpr)

	C.MarkDependency(A.Id, A.CalcDependency())
	C.MarkDependency(B.Id, B.CalcDependency())

	D.MarkDependency(A.Id, A.CalcDependency())
	D.MarkDependency(B.Id, B.CalcDependency())

	E.MarkDependency(D.Id, D.CalcDependency())

	F.MarkDependency(B.Id, B.CalcDependency())
	F.MarkDependency(D.Id, D.CalcDependency())

	if A.CalcDependency() != false || B.CalcDependency() != true || C.CalcDependency() != false || D.CalcDependency() != true || E.CalcDependency() != false || F.CalcDependency() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 1 1   1  1  0  0
	A.SetDependency(DefaultTrueExpr)
	B.SetDependency(DefaultTrueExpr)

	C.MarkDependency(A.Id, A.CalcDependency())
	C.MarkDependency(B.Id, B.CalcDependency())

	D.MarkDependency(A.Id, A.CalcDependency())
	D.MarkDependency(B.Id, B.CalcDependency())

	E.MarkDependency(D.Id, D.CalcDependency())

	F.MarkDependency(B.Id, B.CalcDependency())
	F.MarkDependency(D.Id, D.CalcDependency())

	if A.CalcDependency() != true || B.CalcDependency() != true || C.CalcDependency() != true || D.CalcDependency() != true || E.CalcDependency() != false || F.CalcDependency() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}
}
