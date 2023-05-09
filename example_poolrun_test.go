package gotcc

import (
	"fmt"
	"time"
)

// This example demonstrates how to use controller.PoolRun().
// in this example:
// we use ants pool (panjf2000/ants) with cap=2
// A(1) -+
//       +-(&&)-> E(2) +
// B(3) -+             |
// C(1) -+             +-(&&)-> H(1) --> [termination]
//       +-(&&)-> F(2) +
// D(2) -+             |
// G(2) ---------------+

func sleeptask(args map[string]interface{}) (interface{}, error) {
	time.Sleep(10 * time.Millisecond * time.Duration(args["BIND"].(int)))
	fmt.Println(args["NAME"].(string))
	return nil, nil
}

func ExampleTCController_PoolRun() {
	controller := NewTCController()

	A := controller.AddTask("A", sleeptask, 1)
	B := controller.AddTask("B", sleeptask, 3)
	C := controller.AddTask("C", sleeptask, 1)
	D := controller.AddTask("D", sleeptask, 2)
	E := controller.AddTask("E", sleeptask, 2)
	F := controller.AddTask("F", sleeptask, 2)
	G := controller.AddTask("G", sleeptask, 2)
	H := controller.AddTask("H", sleeptask, 1)

	E.SetDependency(MakeAndExpr(E.NewDependencyExpr(A), E.NewDependencyExpr(B)))
	F.SetDependency(MakeAndExpr(F.NewDependencyExpr(C), F.NewDependencyExpr(D)))
	H.SetDependency(MakeAndExpr(
		MakeAndExpr(H.NewDependencyExpr(E), H.NewDependencyExpr(F)),
		H.NewDependencyExpr(G),
	))

	controller.SetTermination(controller.NewTerminationExpr(H))

	pool := NewDefaultPool(2)
	defer pool.Close()

	_, err := controller.PoolRun(pool)
	if err != nil {
		panic(err)
	}

	// Output:
	// A
	// C
	// B
	// D
	// G
	// E
	// F
	// H
}
