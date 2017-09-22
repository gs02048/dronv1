package worker

import (
	"testing"
	"dronv1/task"
	"fmt"
	"time"
)

func TestWorker_Run(t *testing.T) {
	w := NewWorker()
	task := &task.Task{TaskName:"hello"}
	w.Run(task)
	for k,v := range(w.Runing){
		fmt.Println(k)
		fmt.Println(v.TaskName)
	}

	<-time.After(2*time.Second)

	for k,v := range(w.Runing){
		fmt.Println(k)
		fmt.Println(v.TaskName)
	}

}