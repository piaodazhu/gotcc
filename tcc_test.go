package gotcc

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
)

func TaskDefault(args map[string]interface{}) (interface{}, error) {
	sum := 0
	for _, v := range args {
		if num, ok := v.(int); ok {
			sum += num
		}
	}
	return sum, nil
}

func TestNoTermination(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)
	E := controller.AddTask("E", TaskDefault, 5)
	F := controller.AddTask("F", TaskDefault, 6)

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B))) // 3 + 1 + 2 = 6
	D.SetDependency(D.NewDependencyExpr(C))                                      // 4 + 6 = 10
	E.SetDependency(MakeAndExpr(E.NewDependencyExpr(B), E.NewDependencyExpr(C))) // 5 + 2 + 6 = 13
	F.SetDependency(MakeAndExpr(F.NewDependencyExpr(D), F.NewDependencyExpr(E))) // 6 + 10 + 13 = 29

	// controller.SetTermination(controller.NewTerminationExpr(F))

	_, err := controller.BatchRun()
	if _, ok := err.(ErrNoTermination); !ok {
		t.Fatal(err)
	}
	t.Log(err.Error())

	pool := NewDefaultPool(20)
	defer pool.Close()
	_, err = controller.PoolRun(pool)
	if _, ok := err.(ErrNoTermination); !ok {
		t.Fatal(err)
	}
	t.Log(err.Error())
}

func TestLoopDependencyCond1(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)

	// C <- A and B   A <- C

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B)))

	// make loop dependency
	A.SetDependency(A.NewDependencyExpr(C))

	controller.SetTermination(controller.NewTerminationExpr(C))

	_, err := controller.BatchRun()
	if _, ok := err.(ErrLoopDependency); !ok {
		t.Fatal(err)
	}
	t.Log(err.Error())

	pool := NewDefaultPool(2)
	defer pool.Close()
	_, err = controller.PoolRun(pool)
	if _, ok := err.(ErrLoopDependency); !ok {
		t.Fatal(err)
	}
	t.Log(err.Error())
}

func TestLoopDependencyCond2(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)

	// C <- A  A <- B  B <- C

	// make loop dependency
	C.SetDependency(C.NewDependencyExpr(A))
	A.SetDependency(A.NewDependencyExpr(B))
	B.SetDependency(B.NewDependencyExpr(C))

	controller.SetTermination(controller.NewTerminationExpr(C))

	_, err := controller.BatchRun()
	if _, ok := err.(ErrLoopDependency); !ok {
		t.Fatal(err)
	}
	t.Log(err.Error())
}

func TestLoopDependencyCond3(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)

	// B <- A  D <- A  C <- A and B and D

	C.SetDependency(MakeAndExpr(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B)), C.NewDependencyExpr(D)))

	B.SetDependency(B.NewDependencyExpr(A))
	D.SetDependency(D.NewDependencyExpr(A))

	controller.SetTermination(controller.NewTerminationExpr(C))

	res, err := controller.BatchRun()
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res["C"]; !ok {
		t.Fatal(err)
	} else if sum != 12 {
		t.Fatal("Sum Error", sum)
	}
}

func TestLoopDependencyCond4(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)

	// B <- A  D <- A  C <- B and D  A <- C

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(B), C.NewDependencyExpr(D)))
	B.SetDependency(B.NewDependencyExpr(A))
	D.SetDependency(D.NewDependencyExpr(A))

	// make loop dependency
	A.SetDependency(A.NewDependencyExpr(C))

	controller.SetTermination(controller.NewTerminationExpr(C))

	_, err := controller.BatchRun()
	if _, ok := err.(ErrLoopDependency); !ok {
		t.Fatal(err)
	}
	t.Log(err.Error())
}

func TestPoolUnsupport(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)
	E := controller.AddTask("E", TaskDefault, 5)
	F := controller.AddTask("F", TaskDefault, 6)

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B)))
	D.SetDependency(D.NewDependencyExpr(C))
	// make an 'OR' expr
	E.SetDependency(MakeOrExpr(E.NewDependencyExpr(B), E.NewDependencyExpr(C)))
	F.SetDependency(MakeAndExpr(F.NewDependencyExpr(D), F.NewDependencyExpr(E)))

	controller.SetTermination(controller.NewTerminationExpr(F))

	pool := NewDefaultPool(2)
	defer pool.Close()
	_, err := controller.PoolRun(pool)
	if _, ok := err.(ErrPoolUnsupport); !ok {
		t.Fatal(err)
	}
	t.Log(err.Error())
}

func TestBatchRun(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)
	E := controller.AddTask("E", TaskDefault, 5)
	F := controller.AddTask("F", TaskDefault, 6)

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B))) // 3 + 1 + 2 = 6
	D.SetDependency(D.NewDependencyExpr(C))                                      // 4 + 6 = 10
	E.SetDependency(MakeAndExpr(E.NewDependencyExpr(B), E.NewDependencyExpr(C))) // 5 + 2 + 6 = 13
	F.SetDependency(MakeAndExpr(F.NewDependencyExpr(D), F.NewDependencyExpr(E))) // 6 + 10 + 13 = 29

	controller.SetTermination(controller.NewTerminationExpr(F))

	res, err := controller.BatchRun()
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res["F"]; !ok {
		t.Fatal(err)
	} else if sum != 29 {
		t.Fatal("Sum Error", sum)
	}
}

func TestPoolRun(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)
	E := controller.AddTask("E", TaskDefault, 5)
	F := controller.AddTask("F", TaskDefault, 6)

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B))) // 3 + 1 + 2 = 6
	D.SetDependency(D.NewDependencyExpr(C))                                      // 4 + 6 = 10
	E.SetDependency(MakeAndExpr(E.NewDependencyExpr(B), E.NewDependencyExpr(C))) // 5 + 2 + 6 = 13
	F.SetDependency(MakeAndExpr(F.NewDependencyExpr(D), F.NewDependencyExpr(E))) // 6 + 10 + 13 = 29

	controller.SetTermination(controller.NewTerminationExpr(F))

	// auto-sized pool
	pool := NewDefaultPool(0)
	res, err := controller.PoolRun(pool)
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res["F"]; !ok {
		t.Fatal(err)
	} else if sum != 29 {
		t.Fatal("Sum Error", sum)
	}
	pool.Close()

	// size=2 pool
	pool = NewDefaultPool(2)
	res, err = controller.PoolRun(pool)
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res["F"]; !ok {
		t.Fatal(err)
	} else if sum != 29 {
		t.Fatal("Sum Error", sum)
	}
	pool.Close()

	// just goroutine
	res, err = controller.PoolRun(DefaultNoPool{})
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res["F"]; !ok {
		t.Fatal(err)
	} else if sum != 29 {
		t.Fatal("Sum Error", sum)
	}
	pool.Close()
}

func TestPoolRunError(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)
	E := controller.AddTask("E", TaskDefault, 5)
	F := controller.AddTask("F", TaskDefault, 6)

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B))) // 3 + 1 + 2 = 6
	D.SetDependency(D.NewDependencyExpr(C))                                      // 4 + 6 = 10
	E.SetDependency(MakeAndExpr(E.NewDependencyExpr(B), E.NewDependencyExpr(C))) // 5 + 2 + 6 = 13
	F.SetDependency(MakeAndExpr(F.NewDependencyExpr(D), F.NewDependencyExpr(E))) // 6 + 10 + 13 = 29

	controller.SetTermination(controller.NewTerminationExpr(F))

	// should error!
	pool, err := ants.NewPool(2, ants.WithNonblocking(true))
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	_, err = controller.PoolRun(DefaultPool{pool})
	if err == nil {
		t.Fatal("should error")
	}
	t.Log(err.Error())
}

func TestRunMultiple(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskDefault, 2)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)
	E := controller.AddTask("E", TaskDefault, 5)
	F := controller.AddTask("F", TaskDefault, 6)

	C.SetDependency(MakeAndExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B))) // 3 + 1 + 2 = 6
	D.SetDependency(D.NewDependencyExpr(C))                                      // 4 + 6 = 10
	E.SetDependency(MakeAndExpr(E.NewDependencyExpr(B), E.NewDependencyExpr(C))) // 5 + 2 + 6 = 13
	F.SetDependency(MakeAndExpr(F.NewDependencyExpr(D), F.NewDependencyExpr(E))) // 6 + 10 + 13 = 29

	controller.SetTermination(controller.NewTerminationExpr(F))

	innerState := controller.String()
	for i := 0; i < 4; i++ {
		res, err := controller.BatchRun()
		if err != nil {
			t.Fatal(err)
		}
		if sum, ok := res["F"]; !ok {
			t.Fatal(err)
		} else if sum != 29 {
			t.Fatal("Sum Error", sum)
		}
		if controller.String() != innerState {
			t.Fatal("Reset Error", controller.String(), innerState)
		}
	}

	pool := NewDefaultPool(2)
	defer pool.Close()

	for i := 0; i < 4; i++ {
		res, err := controller.PoolRun(pool)
		if err != nil {
			t.Fatal(err)
		}
		if sum, ok := res["F"]; !ok {
			t.Fatal(err)
		} else if sum != 29 {
			t.Fatal("Sum Error", sum)
		}
		if controller.String() != innerState {
			t.Fatal("Reset Error", controller.String(), innerState)
		}
	}
}

func TestRunOneByOne(t *testing.T) {
	controller := NewTCController()
	last := controller.AddTask("first", TaskDefault, 1)
	for i := 0; i < 999; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskDefault, 1)
		task.SetDependency(task.NewDependencyExpr(last))
		last = task
	}
	controller.SetTermination(controller.NewTerminationExpr(last))
	res, err := controller.BatchRun()
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res[last.Name]; !ok {
		t.Fatal(err)
	} else if sum != 1000 {
		t.Fatal("Sum Error", sum)
	}
}

func TestRunAll(t *testing.T) {
	controller := NewTCController()
	end := controller.AddTask("last", TaskDefault, 1)
	for i := 0; i < 999; i++ {
		task := controller.AddTask(strconv.Itoa(i), TaskDefault, 1)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))
	res, err := controller.BatchRun()
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res[end.Name]; !ok {
		t.Fatal(err)
	} else if sum != 1000 {
		t.Fatal("Sum Error", sum)
	}
}

type ErrTaskFailed struct{ id int }

func (e ErrTaskFailed) Error() string { return "ErrTaskFailed: " + strconv.Itoa(e.id) }

type ErrTaskCanceled struct{ id int }

func (e ErrTaskCanceled) Error() string { return "ErrTaskCanceled: " + strconv.Itoa(e.id) }

type ErrUndoFailed struct{ id int }

func (e ErrUndoFailed) Error() string { return "ErrUndoFailed: " + strconv.Itoa(e.id) }

type argsForTest struct {
	id               int
	taskShouldFailed bool
	undoShouldFailed bool
	infoShouldPrint  bool
	sleepTime        int
}

func TaskMayFailed(args map[string]interface{}) (interface{}, error) {
	select {
	case <-time.After(time.Duration(args["BIND"].(argsForTest).sleepTime) * time.Millisecond * 3):
		if args["BIND"].(argsForTest).taskShouldFailed {
			return nil, ErrTaskFailed{args["BIND"].(argsForTest).id}
		} else {
			return "DONE", nil
		}
	case <-args["CANCEL"].(context.Context).Done():
		return nil, ErrTaskCanceled{args["BIND"].(argsForTest).id}
	}
}

func UndoMayFailed(args map[string]interface{}) error {
	if args["BIND"].(argsForTest).undoShouldFailed == false {
		return nil
	} else {
		return ErrUndoFailed{args["BIND"].(argsForTest).id}
	}
}

func TestCollectTaskErrors(t *testing.T) {
	controller := NewTCController()
	endarg := argsForTest{
		id:               100,
		taskShouldFailed: false,
		sleepTime:        100,
	}
	end := controller.AddTask("mid", TaskMayFailed, endarg)

	for i := 0; i < 99; i++ {
		arg := argsForTest{
			id:               i,
			taskShouldFailed: i == 96, // task 96 will fail
			undoShouldFailed: i == 5,
			sleepTime:        i,
		}
		task := controller.AddTask(strconv.Itoa(i), TaskMayFailed, arg)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))
	_, err := controller.BatchRun()
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}
}

func TestCollectRollbackErrors(t *testing.T) {
	controller := NewTCController()
	endarg := argsForTest{
		id:               100,
		taskShouldFailed: false,
		sleepTime:        100,
	}
	end := controller.AddTask("mid", TaskMayFailed, endarg).SetUndoFunc(UndoMayFailed, false)

	for i := 0; i < 99; i++ {
		arg := argsForTest{
			id:               i,
			taskShouldFailed: i == 96, // task 96 will fail
			undoShouldFailed: i == 5,
			sleepTime:        i,
		}
		task := controller.AddTask(strconv.Itoa(i), TaskMayFailed, arg).SetUndoFunc(UndoMayFailed, false)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))
	_, err := controller.BatchRun()
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}

	pool := NewDefaultPool(20)
	defer pool.Close()
	_, err = controller.PoolRun(pool)
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}
}

func TestCollectRollbackErrorsNoSkip(t *testing.T) {
	controller := NewTCController()
	endarg := argsForTest{
		id:               100,
		taskShouldFailed: false,
		sleepTime:        100,
	}
	end := controller.AddTask("mid", TaskMayFailed, endarg).SetUndoFunc(UndoMayFailed, true)

	for i := 0; i < 99; i++ {
		arg := argsForTest{
			id:               i,
			taskShouldFailed: i == 96,                    // task 96 will fail
			undoShouldFailed: i == 5 || i == 2 || i == 1, // undo 1 2 5 will fail
			sleepTime:        i,
		}
		task := controller.AddTask(strconv.Itoa(i), TaskMayFailed, arg).SetUndoFunc(UndoMayFailed, true)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))
	_, err := controller.BatchRun()
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}

	pool := NewDefaultPool(20)
	defer pool.Close()
	_, err = controller.PoolRun(pool)
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}
}

type TaskState struct {
	TS string
}

func (t TaskState) String() string {
	return t.TS
}

func TaskMayFailedOrCancelled(args map[string]interface{}) (interface{}, error) {
	select {
	case <-time.After(time.Duration(args["BIND"].(argsForTest).sleepTime) * time.Millisecond * 3):
		if args["BIND"].(argsForTest).taskShouldFailed {
			return nil, ErrTaskFailed{args["BIND"].(argsForTest).id}
		} else {
			return "DONE", nil
		}
	case <-args["CANCEL"].(context.Context).Done():
		err := ErrCancelled{}
		return nil, ErrCancelled{TaskState{"ID-" + strconv.Itoa(args["BIND"].(argsForTest).id) + err.Error()}}
	}
}

type ErrUndoFailedWithInfo struct{ info string }

func (e ErrUndoFailedWithInfo) Error() string { return e.info }

func UndoMayFailedWithInfo(args map[string]interface{}) error {
	if args["BIND"].(argsForTest).undoShouldFailed == false {
		return nil
	} else {
		if args["BIND"].(argsForTest).infoShouldPrint == false {
			return ErrUndoFailed{args["BIND"].(argsForTest).id}
		}
		var info strings.Builder
		info.WriteString("\n=====INFO======\n")
		info.WriteString("Task errors: \n")
		ems := args["TASKERR"].([]*ErrorMessage)
		for _, em := range ems {
			info.WriteString(em.TaskName)
			info.WriteByte(':')
			info.WriteString(em.Error.Error())
			info.WriteByte('\n')
		}
		info.WriteString("Undo errors: \n")
		ums := args["UNDOERR"].([]*ErrorMessage)
		for _, um := range ums {
			info.WriteString(um.TaskName)
			info.WriteByte(':')
			info.WriteString(um.Error.Error())
			info.WriteByte('\n')
		}
		info.WriteString("Cancelled: \n")
		sms := args["CANCELLED"].([]*StateMessage)
		for _, sm := range sms {
			info.WriteString(sm.TaskName)
			info.WriteByte(':')
			info.WriteString(sm.State.String())
			info.WriteByte('\n')
		}
		return ErrUndoFailedWithInfo{info.String()}
	}
}

func TestCollectAllErrorsNoSkip(t *testing.T) {
	controller := NewTCController()
	endarg := argsForTest{
		id:               100,
		taskShouldFailed: false,
		sleepTime:        100,
	}
	end := controller.AddTask("mid", TaskMayFailedOrCancelled, endarg).SetUndoFunc(UndoMayFailedWithInfo, true)

	for i := 0; i < 99; i++ {
		arg := argsForTest{
			id:               i,
			taskShouldFailed: i == 96,                    // task 96 will fail
			undoShouldFailed: i == 5 || i == 2 || i == 1, // undo 1 2 5 will fail
			infoShouldPrint:  i == 1,
			sleepTime:        i,
		}
		task := controller.AddTask(strconv.Itoa(i), TaskMayFailedOrCancelled, arg).SetUndoFunc(UndoMayFailedWithInfo, true)
		end.SetDependency(MakeAndExpr(end.DependencyExpr(), end.NewDependencyExpr(task)))
	}
	controller.SetTermination(controller.NewTerminationExpr(end))
	_, err := controller.BatchRun()
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}

	pool := NewDefaultPool(2)
	defer pool.Close()
	_, err = controller.PoolRun(pool)
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}
}

func TaskSilentFail(args map[string]interface{}) (interface{}, error) {
	return nil, ErrSilentFail{}
}

func TestSuccessWithSilentFail(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskSilentFail, nil)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskDefault, 4)

	C.SetDependency(MakeOrExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B)))
	D.SetDependency(D.NewDependencyExpr(C))

	controller.SetTermination(controller.NewTerminationExpr(D))

	res, err := controller.BatchRun()
	if err != nil {
		t.Fatal(err)
	}
	if sum, ok := res["D"]; !ok {
		t.Fatal(err)
	} else if sum != 8 {
		t.Fatal("Sum Error", sum)
	}
}

func TaskMustFail(args map[string]interface{}) (interface{}, error) {
	return nil, ErrTaskFailed{args["BIND"].(int)}
}

func TestFailedWithSilentFail(t *testing.T) {
	controller := NewTCController()
	A := controller.AddTask("A", TaskDefault, 1)
	B := controller.AddTask("B", TaskSilentFail, nil)
	C := controller.AddTask("C", TaskDefault, 3)
	D := controller.AddTask("D", TaskMustFail, 4)

	C.SetDependency(MakeOrExpr(C.NewDependencyExpr(A), C.NewDependencyExpr(B)))
	D.SetDependency(D.NewDependencyExpr(C))

	controller.SetTermination(controller.NewTerminationExpr(D))

	_, err := controller.BatchRun()
	if err == nil {
		t.Fatal("Task not fail!")
	} else {
		t.Log(err)
	}
}
