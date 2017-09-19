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
	for{
		select{
		case m := <-ismaster:
			crontab = cron.New()
			if m{
				cronlist,_ := zk.ListCron(ZkConn)
				for _,item := range(cronlist){
					crontab.AddFunc(item.Spece,func(){
						log.Debug(item.Command)
					})
				}
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