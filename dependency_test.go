package gotcc

import "testing"

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
	checkResult := func(valA bool, valB bool, valC bool, valD bool, valE bool, valF bool) bool {
		if valA {
			A.SetDependency(DefaultTrueExpr)
		} else {
			A.SetDependency(DefaultFalseExpr)
		}
		if valB {
			B.SetDependency(DefaultTrueExpr)
		} else {
			B.SetDependency(DefaultFalseExpr)
		}

		C.MarkDependency(A.id, A.CalcDependency())
		C.MarkDependency(B.id, B.CalcDependency())

		D.MarkDependency(A.id, A.CalcDependency())
		D.MarkDependency(B.id, B.CalcDependency())

		E.MarkDependency(D.id, D.CalcDependency())

		F.MarkDependency(B.id, B.CalcDependency())
		F.MarkDependency(D.id, D.CalcDependency())

		if valC != C.CalcDependency() {
			return false
		}
		if valD != D.CalcDependency() {
			return false
		}
		if valE != E.CalcDependency() {
			return false
		}
		if valF != F.CalcDependency() {
			return false
		}
		return true
	}

	// 0 0   0  0  1  0
	if !checkResult(false, false, false, false, true, false) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 1 0   0  1  0  1
	if !checkResult(true, false, false, true, false, true) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 0 1   0  1  0  0
	if !checkResult(false, true, false, true, false, false) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 1 1   1  1  0  0
	if !checkResult(true, true, true, true, false, false) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}
}
