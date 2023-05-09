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

		C.markDependency(A.id, A.calcDependency())
		C.markDependency(B.id, B.calcDependency())

		D.markDependency(A.id, A.calcDependency())
		D.markDependency(B.id, B.calcDependency())

		E.markDependency(D.id, D.calcDependency())

		F.markDependency(B.id, B.calcDependency())
		F.markDependency(D.id, D.calcDependency())

		if valC != C.calcDependency() {
			return false
		}
		if valD != D.calcDependency() {
			return false
		}
		if valE != E.calcDependency() {
			return false
		}
		if valF != F.calcDependency() {
			return false
		}
		return true
	}

	// 0 0   0  0  1  0
	if !checkResult(false, false, false, false, true, false) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.calcDependency(), B.calcDependency(), C.calcDependency(), D.calcDependency(), E.calcDependency(), F.calcDependency())
	}

	// 1 0   0  1  0  1
	if !checkResult(true, false, false, true, false, true) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.calcDependency(), B.calcDependency(), C.calcDependency(), D.calcDependency(), E.calcDependency(), F.calcDependency())
	}

	// 0 1   0  1  0  0
	if !checkResult(false, true, false, true, false, false) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.calcDependency(), B.calcDependency(), C.calcDependency(), D.calcDependency(), E.calcDependency(), F.calcDependency())
	}

	// 1 1   1  1  0  0
	if !checkResult(true, true, true, true, false, false) {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.calcDependency(), B.calcDependency(), C.calcDependency(), D.calcDependency(), E.calcDependency(), F.calcDependency())
	}
}
