package gotcc

import "sync"

type UndoStack struct {
	Lock  sync.Mutex
	Items []*UndoFunc
}

type UndoFunc struct {
	Name      string
	SkipError bool
	Func      func() error
}

func EmptyUndoFunc(args map[string]interface{}) error {
	return nil
}

func NewUndoFunc(name string, skipError bool, undo func(args map[string]interface{}) error, args map[string]interface{}) *UndoFunc {
	return &UndoFunc{
		Name:      name,
		SkipError: skipError,
		Func: func() error {
			return undo(args)
		},
	}
}

func (u *UndoStack) Push(uf *UndoFunc) {
	u.Lock.Lock()
	u.Items = append(u.Items, uf)
	u.Lock.Unlock()
}

func (u *UndoStack) UndoAll() *ErrorList {
	errors := &ErrorList{}
	for i := len(u.Items) - 1; i >= 0; i-- {
		err := u.Items[i].Func()
		if err != nil {
			errors.Append(NewErrorMessage(u.Items[i].Name, err))
			if !u.Items[i].SkipError {
				return errors
			}
		}
	}
	return errors
}
