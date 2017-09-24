package worker

import (
	"testing"
	"dronv1/task"
	"fmt"
	"time"
)

func TestWorker_Run(t *testing.T) {
	w := NewWorker()
	task := &task.Task{TaskName:"hello",Command:"php /Users/huanghailin/goproject/src/dronv1/index.php",MaxRunTime:10}
	w.Run(task)
	for k,v := range(w.Runing){
		fmt.Println(k)
		fmt.Println(v.TaskName)
	}

	<-time.After(10*time.Second)

	for k,v := range(w.Runing){
		fmt.Println(k)
		fmt.Println(v.TaskName)
	}

}