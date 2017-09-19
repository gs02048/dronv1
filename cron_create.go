package main

import (
	"dronv1/zk"
	log"github.com/alecthomas/log4go"
	"github.com/robfig/cron"
)
var (
	ismaster chan bool
	crontab *cron.Cron
)

func StartCron(){
	ismaster = make(chan bool,1)
	go elect()
	crontab = cron.New()
	for{
		select{
		case m := <-ismaster:
			if m{
				crontab.AddFunc("",func(){})
				log.Info("run cron")
				crontab.Run()
			}else{
				log.Info("stop cron")
				crontab.Stop()
			}
		}
	}

}

func elect(){
	err := zk.ElectMaster(ZkConn,ismaster)
	if err != nil{
		log.Info("elect master err")
	}
	zk.WatchMaster(ZkConn,ismaster)
}