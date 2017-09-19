package zk

import (
	"testing"
	"time"
	"fmt"

)

func TestZK(t *testing.T) {
	conn, err := Connect([]string{"47.93.233.6:2181"}, time.Second*30,true)
	if err != nil {
		t.Error(err)
	}else {
		fmt.Println("ab")
	}
	defer conn.Close()
}

func TestCron(t *testing.T){
	conn, err := Connect([]string{"47.93.233.6:2181"}, time.Second*30,true)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	list,_ := ListCron(conn)
	for _,v := range(list){
		fmt.Println(v.Name)
		fmt.Println(v.Spece)
		fmt.Println(v.Command)
	}
}