package main

import (
	"fmt"
)

func hello(args map[string]interface{}) (interface{}, error) {
	fmt.Println("hello")
	return 1, nil
}

func world(args map[string]interface{}) (interface{}, error) {
	fmt.Println("world")
	return 2, nil
}

func helloworld(args map[string]interface{}) (interface{}, error) {
	fmt.Println("--------")
	for k, v := range args {
		fmt.Println(k, v)
	}
	fmt.Println("--------")
	fmt.Println("helloworld")
	return 3, nil
}

func foo(args map[string]interface{}) (interface{}, error) {
	fmt.Println("foo")
	return 4, nil
}

func bar(args map[string]interface{}) (interface{}, error) {
	fmt.Println("bar")
	return 5, nil
}

func main() {
	manager := NewOCManager()
	
	manager.AddTask("hello", hello, nil)
	manager.AddTask("world", world, nil)
	manager.AddTask("helloworld", helloworld, nil)
	manager.AddTask("foo", foo, nil)
	manager.AddTask("bar", bar, nil)

	manager.RunTask(1)
}
