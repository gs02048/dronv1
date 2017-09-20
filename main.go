package main

import (
	log"github.com/alecthomas/log4go"
	"flag"
	"os"
)


func main(){
	cmd := flag.String("cmd","","cmd")
	adminport := flag.String("port","9999","admin port")
	flag.Parse()
	if *cmd == "cron"{
		InitZk()
		go StartCron()
	}else if *cmd == "run" {
		InitZk()
		InitHttp()
	}else if *cmd == "admin"{
		InitZk()
		InitAdminHttp(*adminport)
	} else{
		log.Info("cmd err")
		os.Exit(1)
	}
	signalCH := InitSignal()
	HandleSignal(signalCH)
	// exit
	log.Info("comet stop")
}
