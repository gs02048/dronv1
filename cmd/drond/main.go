package main

import (
	"dronv1/drond"
	"time"
)

func main(){
	cfg := &drond.Config{
		ZkTaskPrefix:"/LCSCRON",
		ZkTimeout:time.Second * 10,
		ZkServicePrefix:"/CRONSERVICE",
		ZkAddrs:[]string{"localhost:2181"},

		Electpre:"/CRONELECT",
		Electpath:"MASTER",
	}

	d := drond.New(cfg)
	d.Main()

	select {

	}
}
