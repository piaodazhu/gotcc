package gotcc

import "strings"

// ---------- Executor-Level Errors -----------

type ErrCancelled struct {
	State State
}

func (e ErrCancelled) Error() string {
	return "Error: Task is canncelled due to other errors."
}

type ErrSilentFail struct{}

func (e ErrSilentFail) Error() string {
	return "Error: Task failed in silence."
}

// ---------- Controller-Level Errors -----------

type ErrNoTermination struct{}

func (ErrNoTermination) Error() string {
	return "Error: No termination condition has been set!"
}

type ErrLoopDependency struct {
	State State
}

func (e ErrLoopDependency) Error() string {
	return "Error: Tasks has loop dependency."
}

type ErrAborted struct {
	TaskErrors []*ErrorMessage
	UndoErrors []*ErrorMessage
	Cancelled  []*StateMessage
}

func (e ErrAborted) Error() string {
	var sb strings.Builder
	sb.WriteString("\n[x] TaskErrors:\n")
	sb.WriteString((&ErrorList{Items: e.TaskErrors}).String())
	sb.WriteString("[-] UndoErrors:\n")
	sb.WriteString((&ErrorList{Items: e.UndoErrors}).String())
	sb.WriteString("[/] Cancelled:\n")
	sb.WriteString((&CancelList{Items: e.Cancelled}).String())
	return sb.String()
}

type ErrPoolUnsupport struct{}

func (ErrPoolUnsupport) Error() string {
	return "Error: PoolRun is not support when any dependency is not AND!"
}
