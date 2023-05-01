package gooc

import (
	"testing"
)

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
	return 2, nil
}

func helloworld(args map[string]interface{}) (interface{}, error) {
	globalTest.Log("helloworld")
	for k, v := range args {
		globalTest.Logf("arg %s=%d\n", k, v.(int))
	}
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

func TestManager(t *testing.T) {
	globalTest = t
	manager := NewOCManager()

	hello := manager.AddTask("hello", hello, 0)
	world := manager.AddTask("world", world, 0)
	helloworld := manager.AddTask("helloworld", helloworld, 0)
	foo := manager.AddTask("foo", foo, 0)
	bar := manager.AddTask("bar", bar, 0)
	foobar := manager.AddTask("foobar", foobar, 0)

	helloworld.SetDependency(MakeAndExpr(helloworld.NewDependencyExpr(hello), helloworld.NewDependencyExpr(world)))

	foobar.SetDependency(MakeOrExpr(foobar.NewDependencyExpr(foo), foobar.NewDependencyExpr(bar)))

	manager.SetTermination(MakeAndExpr(manager.NewTerminationExpr(foobar), manager.NewTerminationExpr(helloworld)))
	
	manager.RunTask()
}
