package myzk

import (
	"testing"
	"time"
	"fmt"
	"dronv1/task"
)

func TestRegister_ListTask(t *testing.T) {
	conn,_ := Connect([]string{"47.93.233.6:2181"},time.Second * 10)
	register,err := NewRegister(conn,"/LCSCRON")
	if err != nil{
		fmt.Println(err)
		return
	}
	c := &task.Task{
		TaskName:"test1",
		Desc:"test",
		Command:"php index.php",
		Spec:"*/1 * * * * *",
		TaskType:1,
		MaxRunTime:60,
	}
	register.RegisterTask(c)
	list,err := register.ListTask()
	if err != nil{
		fmt.Println(err)
		return
	}
	for k,v := range(list){
		fmt.Println(k)
		fmt.Println(v)
	}

}
