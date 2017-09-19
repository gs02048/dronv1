package main

import (
	"github.com/samuel/go-zookeeper/zk"
	myzk "dronv1/zk"
	"time"
	log "github.com/alecthomas/log4go"
)

var ZkConn *zk.Conn

func InitZk()  {
	var err error
	ZkConn, err = myzk.Connect([]string{"47.93.233.6:2181"}, time.Second*30,true)
	if err != nil {
		log.Error(err)
		return
	}
}