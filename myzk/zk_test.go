package myzk

import (
	"testing"
	"time"
	"fmt"
	"dronv1/task"
)

func TestRegister_ListTask(t *testing.T) {
	conn,_ := Connect([]string{"47.93.233.6:2181"},time.Second * 10)
	register,err := NewRegister(conn)
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
	register.RegisterTask("",c)
	list,err := register.ListTask("")
	if err != nil{
		fmt.Println(err)
		return
	}
	for k,v := range(list){
		fmt.Println(k)
		fmt.Println(v)
	}

}

func TestRegister_WatchNode(t *testing.T) {
	conn,_ := Connect([]string{"localhost:2181"},time.Second * 10)
	register,_ := NewRegister(conn)
	prefix := "/LCSCRON"
	event := make(chan bool,1)
	go register.GetDirWatchEvent(prefix,event)
	for{
		select {
			case e:=<-event:
			if e{
				list,_ := register.ListTask(prefix)
				for _,item := range(list){
					fmt.Println(item.TaskName)
				}
			}

		}
	}



}
func TestRegister_GetTask(t *testing.T) {
	conn,_ := Connect([]string{"localhost:2181"},time.Second * 10)
	register,_ := NewRegister(conn)
	list,_ := register.ListTask("/LCSCRON")
	for _,task := range(list){
		fmt.Println(task.TaskName)
	}
}

