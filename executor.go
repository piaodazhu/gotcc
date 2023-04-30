package main


type Executor struct {
	Id   uint32
	Name string

	Dependency     map[uint32]bool
	DependencyExpr func(dependency map[uint32]bool) bool

	MessageBuffer chan Message
	Subscribers   []chan Message

	Args         interface{}
	Task         func(args map[string]interface{}) (interface{}, error)
	RollbackTask func(args map[string]interface{}) (interface{}, error)

	Manager *Manager
}

type Message struct {
	Sender     uint32
	SenderName string
	Finish     bool
	Value      interface{}
}
