package gotcc

import (
	"fmt"
	"testing"
)

func PrintTask(args map[string]interface{}) (interface{}, error) {
	for name := range args {
		fmt.Println("args: ", name)
	}
	str := args["BIND"].(string)
	fmt.Println(str)
	return nil, nil
}

func TestDependency(t *testing.T) {
	A := newExecutor("A", PrintTask, "A")
	B := newExecutor("B", PrintTask, "B")
	C := newExecutor("C", PrintTask, "C")
	D := newExecutor("D", PrintTask, "D")
	E := newExecutor("E", PrintTask, "E")
	F := newExecutor("F", PrintTask, "F")

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

	C.Dependency[A.Id] = A.CalcDependency()
	C.Dependency[B.Id] = B.CalcDependency()
	D.Dependency[A.Id] = A.CalcDependency()
	D.Dependency[B.Id] = B.CalcDependency()
	E.Dependency[D.Id] = D.CalcDependency()
	F.Dependency[B.Id] = B.CalcDependency()
	F.Dependency[D.Id] = D.CalcDependency()

	if A.CalcDependency() != false || B.CalcDependency() != false || C.CalcDependency() != false || D.CalcDependency() != false || E.CalcDependency() != true || F.CalcDependency() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 1 0   0  1  0  1
	A.SetDependency(DefaultTrueExpr)
	B.SetDependency(DefaultFalseExpr)

	C.Dependency[A.Id] = A.CalcDependency()
	C.Dependency[B.Id] = B.CalcDependency()
	D.Dependency[A.Id] = A.CalcDependency()
	D.Dependency[B.Id] = B.CalcDependency()
	E.Dependency[D.Id] = D.CalcDependency()
	F.Dependency[B.Id] = B.CalcDependency()
	F.Dependency[D.Id] = D.CalcDependency()

	if A.CalcDependency() != true || B.CalcDependency() != false || C.CalcDependency() != false || D.CalcDependency() != true || E.CalcDependency() != false || F.CalcDependency() != true {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 0 1   0  1  0  0
	A.SetDependency(DefaultFalseExpr)
	B.SetDependency(DefaultTrueExpr)

	C.Dependency[A.Id] = A.CalcDependency()
	C.Dependency[B.Id] = B.CalcDependency()
	D.Dependency[A.Id] = A.CalcDependency()
	D.Dependency[B.Id] = B.CalcDependency()
	E.Dependency[D.Id] = D.CalcDependency()
	F.Dependency[B.Id] = B.CalcDependency()
	F.Dependency[D.Id] = D.CalcDependency()

	if A.CalcDependency() != false || B.CalcDependency() != true || C.CalcDependency() != false || D.CalcDependency() != true || E.CalcDependency() != false || F.CalcDependency() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}

	// 1 1   1  1  0  0
	A.SetDependency(DefaultTrueExpr)
	B.SetDependency(DefaultTrueExpr)

	C.Dependency[A.Id] = A.CalcDependency()
	C.Dependency[B.Id] = B.CalcDependency()
	D.Dependency[A.Id] = A.CalcDependency()
	D.Dependency[B.Id] = B.CalcDependency()
	E.Dependency[D.Id] = D.CalcDependency()
	F.Dependency[B.Id] = B.CalcDependency()
	F.Dependency[D.Id] = D.CalcDependency()

	if A.CalcDependency() != true || B.CalcDependency() != true || C.CalcDependency() != true || D.CalcDependency() != true || E.CalcDependency() != false || F.CalcDependency() != false {
		t.Errorf("Error: A=%v, B=%v, C=%v, D=%v, E=%v, F=%v\n", A.CalcDependency(), B.CalcDependency(), C.CalcDependency(), D.CalcDependency(), E.CalcDependency(), F.CalcDependency())
	}
}
