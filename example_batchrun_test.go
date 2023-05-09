package gotcc

import (
	"fmt"
	"time"
)

// This example demonstrates how to use controller.BatchRun().
// in this example:
// hello -+
//        +-(&&)-> helloworld +
// world -+                   +
//  foo  -+                   +-(&&)-> [termination]
//        +-(||)->   foobar   +
//  bar  -+

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
