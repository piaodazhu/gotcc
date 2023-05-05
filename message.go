package gotcc

import (
	"strings"
	"sync"
)

type Message struct {
	Sender     uint32
	SenderName string
	Value      interface{}
}

type ErrorMessage struct {
	TaskName string
	Error    error
}

func NewErrorMessage(taskName string, err error) *ErrorMessage {
	return &ErrorMessage{
		TaskName: taskName,
		Error:    err,
	}
}

type ErrorList struct {
	Lock  sync.Mutex
	Items []*ErrorMessage
}

func (el *ErrorList) Append(em *ErrorMessage) {
	el.Lock.Lock()
	el.Items = append(el.Items, em)
	el.Lock.Unlock()
}

func (el *ErrorList) Reset() {
	el.Lock.Lock()
	el.Items = []*ErrorMessage{}
	el.Lock.Unlock()
}

func (el *ErrorList) String() string {
	var sb strings.Builder
	el.Lock.Lock()
	for i := range el.Items {
		sb.WriteString(el.Items[i].TaskName)
		sb.WriteString(": ")
		sb.WriteString(el.Items[i].Error.Error())
		sb.WriteString("\n")
	}
	el.Lock.Unlock()
	return sb.String()
}

type StateMessage struct {
	TaskName string
	State    State
}

type State interface {
	String() string
}

func NewStateMessage(taskName string, state State) *StateMessage {
	return &StateMessage{
		TaskName: taskName,
		State:    state,
	}
}

type CancelList struct {
	Lock  sync.Mutex
	Items []*StateMessage
}

func (cl *CancelList) Append(sm *StateMessage) {
	cl.Lock.Lock()
	cl.Items = append(cl.Items, sm)
	cl.Lock.Unlock()
}

func (cl *CancelList) Reset() {
	cl.Lock.Lock()
	cl.Items = []*StateMessage{}
	cl.Lock.Unlock()
}

func (cl *CancelList) String() string {
	var sb strings.Builder
	cl.Lock.Lock()
	for i := range cl.Items {
		sb.WriteString(cl.Items[i].TaskName)
		sb.WriteString(": ")
		sb.WriteString(cl.Items[i].State.String())
		sb.WriteString("\n")
	}
	cl.Lock.Unlock()
	return sb.String()
}
