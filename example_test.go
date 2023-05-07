package gotcc

import (
	"fmt"
	"time"
)

func hello(args map[string]interface{}) (interface{}, error) {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("hello")
	return 1, nil
}

func world(args map[string]interface{}) (interface{}, error) {
	time.Sleep(20 * time.Millisecond)
	fmt.Println("world")
	return 2, nil
}

func helloworld(args map[string]interface{}) (interface{}, error) {
	fmt.Println("helloworld")
	return 3, nil
}

func foo(args map[string]interface{}) (interface{}, error) {
	time.Sleep(30 * time.Millisecond)
	fmt.Println("foo")
	return 4, nil
}

func bar(args map[string]interface{}) (interface{}, error) {
	time.Sleep(40 * time.Millisecond)
	fmt.Println("bar")
	return 5, nil
}

func foobar(args map[string]interface{}) (interface{}, error) {
	fmt.Println("foobar")
	return 5, nil
}

func ExampleTCController_BatchRun() {
	// in this example:
	// hello -+
	//        +-(&&)-> helloworld +
	// world -+                   +
	//  foo  -+                   +-(&&)-> [termination]
	//        +-(||)->   foobar   +
	//  bar  -+

	controller := NewTCController()

	hello := controller.AddTask("hello", hello, 0)
	world := controller.AddTask("world", world, 1)
	helloworld := controller.AddTask("helloworld", helloworld, 2)
	foo := controller.AddTask("foo", foo, 3)
	bar := controller.AddTask("bar", bar, 4)
	foobar := controller.AddTask("foobar", foobar, 5)

	helloworld.SetDependency(MakeAndExpr(helloworld.NewDependencyExpr(hello), helloworld.NewDependencyExpr(world)))

	foobar.SetDependency(MakeOrExpr(foobar.NewDependencyExpr(foo), foobar.NewDependencyExpr(bar)))

	controller.SetTermination(MakeAndExpr(controller.NewTerminationExpr(foobar), controller.NewTerminationExpr(helloworld)))

	_, err := controller.BatchRun()
	if err != nil {
		panic(err)
	}

	// Output:
	// hello
	// world
	// helloworld
	// foo
	// foobar
	// bar
}

func sleeptask(args map[string]interface{}) (interface{}, error) {
	time.Sleep(10 * time.Millisecond * time.Duration(args["BIND"].(int)))
	fmt.Println(args["NAME"].(string))
	return nil, nil
}

func ExampleTCController_PoolRun() {
	// in this example:
	// we use ants pool (panjf2000/ants) with cap=2
	// A(1) -+
	//       +-(&&)-> E(2) +
	// B(3) -+             |
	// C(1) -+             +-(&&)-> H(1) --> [termination]
	//       +-(&&)-> F(2) +
	// D(2) -+             |
	// G(2) ---------------+
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
