package gotcc

import "strings"

// ---------- Executor-Level Errors -----------

// It means the task is cancelled by the controller because the some other task return a fatal error.
// If a task is aborted with args["CANCEL"].(context.Context).done(), it should return ErrCancelled.
// And its running state can be set into the error.
// The task and its state will be put into Cancelled list.
type ErrCancelled struct {
	State State
}

func (e ErrCancelled) Error() string {
	return "Error: Task is canncelled due to other errors."
}

// It means the task failed, but you don't want the tasks execution aborted.
// The failed task will be put into TaskErrors if the tasks execution finally failed.
type ErrSilentFail struct{}

func (e ErrSilentFail) Error() string {
	return "Error: Task failed in silence."
}

// ---------- Controller-Level Errors -----------

// It means the controller's termination condition haven't been set.
type ErrNoTermination struct{}

func (ErrNoTermination) Error() string {
	return "Error: No termination condition has been set!"
}

// It means there is loop dependency among the tasks.
type ErrLoopDependency struct {
	State State
}

func (e ErrLoopDependency) Error() string {
	return "Error: Tasks has loop dependency."
}

// It means the controller doesn't support PoolRun() because not all dependency expressions are `AND`.
type ErrPoolUnsupport struct{}

func (ErrPoolUnsupport) Error() string {
	return "Error: PoolRun is not support when any dependency is not AND!"
}

// It means some fatal errors occur so the execution failed.
// It consists of multiple errors. TaskErrors: errors from task running.
// UndoErrors: errors from undo function running. Cancelled: running but cancelled tasks.
type ErrAborted struct {
	TaskErrors []*ErrorMessage
	UndoErrors []*ErrorMessage
	Cancelled  []*StateMessage
}

func (e ErrAborted) Error() string {
	var sb strings.Builder
	sb.WriteString("\n[x] TaskErrors:\n")
	sb.WriteString((&errorLisk{items: e.TaskErrors}).String())
	sb.WriteString("[-] UndoErrors:\n")
	sb.WriteString((&errorLisk{items: e.UndoErrors}).String())
	sb.WriteString("[/] Cancelled:\n")
	sb.WriteString((&cancelList{items: e.Cancelled}).String())
	return sb.String()
}
