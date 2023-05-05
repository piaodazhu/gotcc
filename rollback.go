package gotcc

import "sync"

type UndoStack struct {
	Lock  sync.Mutex
	Items []*UndoFunc
}

type UndoFunc struct {
	Name string

	SkipError bool

	Args map[string]interface{}
	Func func(map[string]interface{}) error
}

func EmptyUndoFunc(args map[string]interface{}) error {
	return nil
}

func NewUndoFunc(name string, skipError bool, undo func(args map[string]interface{}) error, args map[string]interface{}) *UndoFunc {
	return &UndoFunc{
		Name:      name,
		SkipError: skipError,
		Args:      args,
		Func:      undo,
	}
}

func (u *UndoStack) Push(uf *UndoFunc) {
	u.Lock.Lock()
	u.Items = append(u.Items, uf)
	u.Lock.Unlock()
}

func (u *UndoStack) Reset() {
	u.Lock.Lock()
	u.Items = []*UndoFunc{}
	u.Lock.Unlock()
}

func (u *UndoStack) UndoAll(taskErrors *ErrorList, cancelled *CancelList) *ErrorList {
	undoErrors := &ErrorList{}
	for i := len(u.Items) - 1; i >= 0; i-- {
		u.Items[i].Args["TASKERR"] = taskErrors.Items
		u.Items[i].Args["UNDOERR"] = undoErrors.Items
		u.Items[i].Args["CANCELLED"] = cancelled.Items

		err := u.Items[i].Func(u.Items[i].Args)
		if err != nil {
			undoErrors.Append(NewErrorMessage(u.Items[i].Name, err))
			if !u.Items[i].SkipError {
				return undoErrors
			}
		}
	}
	return undoErrors
}
