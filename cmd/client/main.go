package main

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
	"fmt"
)

func main(){
	conn,_,err := zk.Connect([]string{"localhost:2181"},10*time.Second)
	if err != nil{
		fmt.Println(err)
		return
	}

	for{
		children,stat,event,err := conn.ChildrenW("/LCSCRON")
		if err != nil{
			fmt.Println(err)
			continue
		}
		for _,v := range children{
			fmt.Println(v)
		}
		fmt.Println("children num:",stat.NumChildren)


		select {
		case e:=<-event:
			fmt.Println(e)

		}
	}


}
