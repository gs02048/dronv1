package main

import (
	log"github.com/alecthomas/log4go"
	"flag"
	"os"
)


func main(){
	cmd := flag.String("cmd","","cmd")
	flag.Parse()
	if *cmd == "cron"{
		InitZk()
		StartCron()
	}else if *cmd == "run" {
		InitZk()
		InitHttp()
	}else{
		log.Info("cmd err")
		os.Exit(1)
	}
	signalCH := InitSignal()
	HandleSignal(signalCH)
	// exit
	log.Info("comet stop")
}
