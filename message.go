package gotcc

import (
	"strings"
	"sync"
)

type message struct {
	senderId   uint32
	senderName string
	value      interface{}
}

type ErrorMessage struct {
	TaskName string
	Error    error
}

func newErrorMessage(taskName string, err error) *ErrorMessage {
	return &ErrorMessage{
		TaskName: taskName,
		Error:    err,
	}
}

type errorLisk struct {
	lock  sync.Mutex
	items []*ErrorMessage
}

func (el *errorLisk) append(em *ErrorMessage) {
	el.lock.Lock()
	el.items = append(el.items, em)
	el.lock.Unlock()
}

func (el *errorLisk) reset() {
	el.lock.Lock()
	el.items = []*ErrorMessage{}
	el.lock.Unlock()
}

func (el *errorLisk) String() string {
	var sb strings.Builder
	el.lock.Lock()
	for i := range el.items {
		sb.WriteString(el.items[i].TaskName)
		sb.WriteString(": ")
		sb.WriteString(el.items[i].Error.Error())
		sb.WriteString("\n")
	}
	el.lock.Unlock()
	return sb.String()
}

type StateMessage struct {
	TaskName string
	State    State
}

type State interface {
	String() string
}

func newStateMessage(taskName string, state State) *StateMessage {
	return &StateMessage{
		TaskName: taskName,
		State:    state,
	}
}

type cancelList struct {
	lock  sync.Mutex
	items []*StateMessage
}

func (cl *cancelList) append(sm *StateMessage) {
	cl.lock.Lock()
	cl.items = append(cl.items, sm)
	cl.lock.Unlock()
}

func (cl *cancelList) reset() {
	cl.lock.Lock()
	cl.items = []*StateMessage{}
	cl.lock.Unlock()
}

func (cl *cancelList) String() string {
	var sb strings.Builder
	cl.lock.Lock()
	for i := range cl.items {
		sb.WriteString(cl.items[i].TaskName)
		sb.WriteString(": ")
		sb.WriteString(cl.items[i].State.String())
		sb.WriteString("\n")
	}
	cl.lock.Unlock()
	return sb.String()
}
