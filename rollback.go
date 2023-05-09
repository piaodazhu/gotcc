package gotcc

import "sync"

type undoStack struct {
	lock  sync.Mutex
	items []*undoFunc
}

type undoFunc struct {
	name string

	skipError bool

	args map[string]interface{}
	f    func(map[string]interface{}) error
}

func newUndoFunc(name string, skipError bool, undo func(args map[string]interface{}) error, args map[string]interface{}) *undoFunc {
	return &undoFunc{
		name:      name,
		skipError: skipError,
		args:      args,
		f:         undo,
	}
}

func (u *undoStack) push(uf *undoFunc) {
	u.lock.Lock()
	u.items = append(u.items, uf)
	u.lock.Unlock()
}

func (u *undoStack) reset() {
	u.lock.Lock()
	u.items = []*undoFunc{}
	u.lock.Unlock()
}

func (u *undoStack) undoAll(taskErrors *errorLisk, cancelled *cancelList) *errorLisk {
	undoErrors := &errorLisk{}
	for i := len(u.items) - 1; i >= 0; i-- {
		u.items[i].args["TASKERR"] = taskErrors.items
		u.items[i].args["UNDOERR"] = undoErrors.items
		u.items[i].args["CANCELLED"] = cancelled.items

		err := u.items[i].f(u.items[i].args)
		if err != nil {
			undoErrors.append(newErrorMessage(u.items[i].name, err))
			if !u.items[i].skipError {
				return undoErrors
			}
		}
	}
	return undoErrors
}

// Default undo function
var EmptyUndoFunc = func(args map[string]interface{}) error {
	return nil
}
