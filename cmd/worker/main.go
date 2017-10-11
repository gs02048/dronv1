package main

import (
	"dronv1/worker"
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

	worker.Main(cfg)
}
