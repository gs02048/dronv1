package main

import (
	"dronv1/admin"

	//"fmt"
	"time"
)

func main(){
	cfg := &admin.Config{
		HttpPort:"5890",
		ZkTaskPrefix:"/LCSCRON",
		ZkTimeout:time.Second * 10,
		ZkServicePrefix:"/CRONSERVICE",
		ZkAddrs:[]string{"47.93.233.6:2181"},
	}

	ad := admin.NewAdmin(cfg)
	admin.InitAdminHttp(ad)
	select {

	}
}
