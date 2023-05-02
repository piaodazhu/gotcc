package gotcc

import (
	"testing"
)

type ErrForTest struct{}

func (ErrForTest) Error() string { return "ErrForTest" }

func hello(args map[string]interface{}) (interface{}, error) {
	globalTest.Log("hello")
	for k, v := range args {
		globalTest.Logf("arg %s=%d\n", k, v.(int))
	}
	return 1, nil
}

func world(args map[string]interface{}) (interface{}, error) {
	globalTest.Log("world")
	for k, v := range args {
		globalTest.Logf("arg %s=%d\n", k, v.(int))
	}
	// time.Sleep(1 * time.Second)
	// return 0, ErrForTest{}
	return 2, nil
}

func helloworld(args map[string]interface{}) (interface{}, error) {
	globalTest.Log("helloworld")
	for k, v := range args {
		globalTest.Logf("arg %s=%d\n", k, v.(int))
	}
	// time.Sleep(1 * time.Second)
	// return 0, ErrForTest{}
	return 3, nil
}

func foo(args map[string]interface{}) (interface{}, error) {
	globalTest.Log("foo")
	for k, v := range args {
		globalTest.Logf("arg %s=%d\n", k, v.(int))
	}
	return 4, nil
}

func bar(args map[string]interface{}) (interface{}, error) {
	globalTest.Log("bar")
	for k, v := range args {
		globalTest.Logf("arg %s=%d\n", k, v.(int))
	}
	return 5, nil
}

func foobar(args map[string]interface{}) (interface{}, error) {
	globalTest.Log("foobar")
	for k, v := range args {
		globalTest.Logf("arg %s=%d\n", k, v.(int))
	}
	return 5, nil
}

var globalTest *testing.T

func TestTCController(t *testing.T) {
	globalTest = t
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

	ret, err := controller.RunTask()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}
