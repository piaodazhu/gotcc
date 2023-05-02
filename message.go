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
